// Copyright 2013 The Go Authors. All rights reserved.
//
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file or at
// https://developers.google.com/open-source/licenses/bsd.

// Package gosrc fetches Go package source code from version control services.
package gosrc

import (
	"encoding/xml"
	"errors"
	"fmt"
	"io"
	"net/http"
	"path"
	"regexp"
	"strings"
	"time"
)

const ExpiresAfter = 2 * 365 * 24 * time.Hour // Package with no commits and imports expires.

// File represents a file.
type File struct {
	// File name with no directory.
	Name string

	// Contents of the file.
	Data []byte

	// Location of file on version control service website.
	BrowseURL string
}

type DirectoryStatus int

const (
	Active          DirectoryStatus = iota
	DeadEndFork                     // Forks with no commits
	QuickFork                       // Forks with less than 3 commits, all within a week from creation
	NoRecentCommits                 // No commits for ExpiresAfter

	// No commits for ExpiresAfter and no imports.
	// This is a status derived from NoRecentCommits and the imports count information in the db.
	Inactive
)

// Directory describes a directory on a version control service.
type Directory struct {
	// The import path for this package.
	ImportPath string

	// Import path of package after resolving go-import meta tags, if any.
	ResolvedPath string

	// Import path prefix for all packages in the project.
	ProjectRoot string

	// Name of the project.
	ProjectName string

	// Project home page.
	ProjectURL string

	// Version control system: git, hg, bzr, ...
	VCS string

	// Version control: active or should be suppressed.
	Status DirectoryStatus

	// Cache validation tag. This tag is not necessarily an HTTP entity tag.
	// The tag is "" if there is no meaningful cache validation for the VCS.
	Etag string

	// Files.
	Files []*File

	// Subdirectories, not guaranteed to contain Go code.
	Subdirectories []string

	// Location of directory on version control service website.
	BrowseURL string

	// Format specifier for link to source line. It must contain one %s (file URL)
	// followed by one %d (source line number), or be empty string if not available.
	// Example: "%s#L%d".
	LineFmt string

	// Whether the repository of this directory is a fork of another one.
	Fork bool

	// How many stars (for a GitHub project) or followers (for a BitBucket
	// project) the repository of this directory has.
	Stars int
}

// Project represents a repository.
type Project struct {
	Description string
}

// NotFoundError indicates that the directory or presentation was not found.
type NotFoundError struct {
	// Diagnostic message describing why the directory was not found.
	Message string

	// Redirect specifies the path where package can be found.
	Redirect string
}

func (e NotFoundError) Error() string {
	return e.Message
}

// IsNotFound returns true if err is of type NotFoundError.
func IsNotFound(err error) bool {
	_, ok := err.(NotFoundError)
	return ok
}

type RemoteError struct {
	Host string
	err  error
}

func (e *RemoteError) Error() string {
	return e.err.Error()
}

type NotModifiedError struct {
	Since  time.Time
	Status DirectoryStatus
}

func (e NotModifiedError) Error() string {
	msg := "package not modified"
	if !e.Since.IsZero() {
		msg += fmt.Sprintf(" since %s", e.Since.Format(time.RFC1123))
	}
	if e.Status == QuickFork {
		msg += " (package is a quick fork)"
	}
	return msg
}

var errNoMatch = errors.New("no match")

// service represents a source code control service.
type service struct {
	pattern         *regexp.Regexp
	prefix          string
	get             func(*http.Client, map[string]string, string) (*Directory, error)
	getPresentation func(*http.Client, map[string]string) (*Presentation, error)
	getProject      func(*http.Client, map[string]string) (*Project, error)
}

var services []*service

func addService(s *service) {
	if s.prefix == "" {
		services = append(services, s)
	} else {
		services = append([]*service{s}, services...)
	}
}

func (s *service) match(importPath string) (map[string]string, error) {
	if !strings.HasPrefix(importPath, s.prefix) {
		return nil, nil
	}
	m := s.pattern.FindStringSubmatch(importPath)
	if m == nil {
		if s.prefix != "" {
			return nil, NotFoundError{Message: "Import path prefix matches known service, but regexp does not."}
		}
		return nil, nil
	}
	match := map[string]string{"importPath": importPath}
	for i, n := range s.pattern.SubexpNames() {
		if n != "" {
			match[n] = m[i]
		}
	}
	return match, nil
}

// importMeta represents the values in a go-import meta tag.
type importMeta struct {
	projectRoot string
	vcs         string
	repo        string
}

// sourceMeta represents the values in a go-source meta tag.
type sourceMeta struct {
	projectRoot  string
	projectURL   string
	dirTemplate  string
	fileTemplate string
}

func isHTTPURL(s string) bool {
	return strings.HasPrefix(s, "https://") || strings.HasPrefix(s, "http://")
}

func replaceDir(s string, dir string) string {
	slashDir := ""
	dir = strings.Trim(dir, "/")
	if dir != "" {
		slashDir = "/" + dir
	}
	s = strings.Replace(s, "{dir}", dir, -1)
	s = strings.Replace(s, "{/dir}", slashDir, -1)
	return s
}

func attrValue(attrs []xml.Attr, name string) string {
	for _, a := range attrs {
		if strings.EqualFold(a.Name.Local, name) {
			return a.Value
		}
	}
	return ""
}

func fetchMeta(client *http.Client, importPath string) (scheme string, im *importMeta, sm *sourceMeta, redir bool, err error) {
	uri := importPath
	if !strings.Contains(uri, "/") {
		// Add slash for root of domain.
		uri = uri + "/"
	}
	uri = uri + "?go-get=1"

	c := httpClient{client: client}
	scheme = "https"
	resp, err := c.get(scheme + "://" + uri)
	if err != nil || resp.StatusCode != 200 {
		if err == nil {
			resp.Body.Close()
		}
		scheme = "http"
		resp, err = c.get(scheme + "://" + uri)
		if err != nil {
			return scheme, nil, nil, false, err
		}
	}
	defer resp.Body.Close()
	im, sm, redir, err = parseMeta(scheme, importPath, resp.Body)
	return scheme, im, sm, redir, err
}

var refreshToGodocPat = regexp.MustCompile(`(?i)^\d+; url=https?://godoc\.org/`)

func parseMeta(scheme, importPath string, r io.Reader) (im *importMeta, sm *sourceMeta, redir bool, err error) {
	errorMessage := "go-import meta tag not found"

	d := xml.NewDecoder(r)
	d.Strict = false
metaScan:
	for {
		t, tokenErr := d.Token()
		if tokenErr != nil {
			break metaScan
		}
		switch t := t.(type) {
		case xml.EndElement:
			if strings.EqualFold(t.Name.Local, "head") {
				break metaScan
			}
		case xml.StartElement:
			if strings.EqualFold(t.Name.Local, "body") {
				break metaScan
			}
			if !strings.EqualFold(t.Name.Local, "meta") {
				continue metaScan
			}
			if strings.EqualFold(attrValue(t.Attr, "http-equiv"), "refresh") {
				// Check for http-equiv refresh back to godoc.org.
				redir = refreshToGodocPat.MatchString(attrValue(t.Attr, "content"))
				continue metaScan
			}
			nameAttr := attrValue(t.Attr, "name")
			if nameAttr != "go-import" && nameAttr != "go-source" {
				continue metaScan
			}
			fields := strings.Fields(attrValue(t.Attr, "content"))
			if len(fields) < 1 {
				continue metaScan
			}
			projectRoot := fields[0]
			if !strings.HasPrefix(importPath, projectRoot) ||
				!(len(importPath) == len(projectRoot) || importPath[len(projectRoot)] == '/') {
				// Ignore if root is not a prefix of the  path. This allows a
				// site to use a single error page for multiple repositories.
				continue metaScan
			}
			switch nameAttr {
			case "go-import":
				if len(fields) != 3 {
					errorMessage = "go-import meta tag content attribute does not have three fields"
					continue metaScan
				}
				if im != nil {
					im = nil
					errorMessage = "more than one go-import meta tag found"
					break metaScan
				}
				im = &importMeta{
					projectRoot: projectRoot,
					vcs:         fields[1],
					repo:        fields[2],
				}
			case "go-source":
				if sm != nil {
					// Ignore extra go-source meta tags.
					continue metaScan
				}
				if len(fields) != 4 {
					continue metaScan
				}
				sm = &sourceMeta{
					projectRoot:  projectRoot,
					projectURL:   fields[1],
					dirTemplate:  fields[2],
					fileTemplate: fields[3],
				}
			}
		}
	}
	if im == nil {
		return nil, nil, redir, NotFoundError{Message: fmt.Sprintf("%s at %s://%s", errorMessage, scheme, importPath)}
	}
	if sm != nil && sm.projectRoot != im.projectRoot {
		sm = nil
	}
	return im, sm, redir, nil
}

// getVCSDirFn is called by getDynamic to fetch source using VCS commands. The
// default value here does nothing. If the code is not built for App Engine,
// then getvCSDirFn is set getVCSDir, the function that actually does the work.
var getVCSDirFn = func(client *http.Client, m map[string]string, etag string) (*Directory, error) {
	return nil, errNoMatch
}

// getDynamic gets a directory from a service that is not statically known.
func getDynamic(client *http.Client, importPath, etag string) (*Directory, error) {
	metaProto, im, sm, redir, err := fetchMeta(client, importPath)
	if err != nil {
		return nil, err
	}

	if im.projectRoot != importPath {
		var imRoot *importMeta
		metaProto, imRoot, _, redir, err = fetchMeta(client, im.projectRoot)
		if err != nil {
			return nil, err
		}
		if *imRoot != *im {
			return nil, NotFoundError{Message: "project root mismatch."}
		}
	}

	// clonePath is the repo URL from import meta tag, with the "scheme://" prefix removed.
	// It should be used for cloning repositories.
	// repo is the repo URL from import meta tag, with the "scheme://" prefix removed, and
	// a possible ".vcs" suffix trimmed.
	i := strings.Index(im.repo, "://")
	if i < 0 {
		return nil, NotFoundError{Message: "bad repo URL: " + im.repo}
	}
	proto := im.repo[:i]
	clonePath := im.repo[i+len("://"):]
	repo := strings.TrimSuffix(clonePath, "."+im.vcs)
	dirName := importPath[len(im.projectRoot):]

	resolvedPath := repo + dirName
	dir, err := getStatic(client, resolvedPath, etag)
	if err == errNoMatch {
		resolvedPath = repo + "." + im.vcs + dirName
		match := map[string]string{
			"dir":        dirName,
			"importPath": importPath,
			"clonePath":  clonePath,
			"repo":       repo,
			"scheme":     proto,
			"vcs":        im.vcs,
		}
		dir, err = getVCSDirFn(client, match, etag)
	}
	if err != nil || dir == nil {
		return nil, err
	}

	dir.ImportPath = importPath
	dir.ProjectRoot = im.projectRoot
	dir.ResolvedPath = resolvedPath
	dir.ProjectName = path.Base(im.projectRoot)
	if !redir {
		dir.ProjectURL = metaProto + "://" + im.projectRoot
	}

	if sm == nil {
		return dir, nil
	}

	if isHTTPURL(sm.projectURL) {
		dir.ProjectURL = sm.projectURL
	}

	if isHTTPURL(sm.dirTemplate) {
		dir.BrowseURL = replaceDir(sm.dirTemplate, dirName)
	}

	// TODO: Refactor this to be simpler, implement the go-source meta tag spec fully.
	if isHTTPURL(sm.fileTemplate) {
		fileTemplate := replaceDir(sm.fileTemplate, dirName)
		if strings.Contains(fileTemplate, "{file}") {
			cut := strings.LastIndex(fileTemplate, "{file}") + len("{file}") // Cut point is right after last {file} section.
			switch hash := strings.Index(fileTemplate, "#"); {
			case hash == -1: // If there's no '#', place cut at the end.
				cut = len(fileTemplate)
			case hash > cut: // If a '#' comes after last {file}, use it as cut point.
				cut = hash
			}
			head, tail := fileTemplate[:cut], fileTemplate[cut:]
			for _, f := range dir.Files {
				f.BrowseURL = strings.Replace(head, "{file}", f.Name, -1)
			}

			if strings.Contains(tail, "{line}") {
				s := strings.Replace(tail, "%", "%%", -1)
				s = strings.Replace(s, "{line}", "%d", 1)
				dir.LineFmt = "%s" + s
			}
		}
	}

	return dir, nil
}

// getStatic gets a diretory from a statically known service. getStatic
// returns errNoMatch if the import path is not recognized.
func getStatic(client *http.Client, importPath, etag string) (*Directory, error) {
	for _, s := range services {
		if s.get == nil {
			continue
		}
		match, err := s.match(importPath)
		if err != nil {
			return nil, err
		}
		if match != nil {
			dir, err := s.get(client, match, etag)
			if dir != nil {
				dir.ImportPath = importPath
				dir.ResolvedPath = importPath
			}
			return dir, err
		}
	}
	return nil, errNoMatch
}

func Get(client *http.Client, importPath string, etag string) (dir *Directory, err error) {
	switch {
	case localPath != "":
		dir, err = getLocal(importPath)
	case IsGoRepoPath(importPath):
		dir, err = getStandardDir(client, importPath, etag)
	case IsValidRemotePath(importPath):
		dir, err = getStatic(client, importPath, etag)
		if err == errNoMatch {
			dir, err = getDynamic(client, importPath, etag)
		}
	default:
		err = errNoMatch
	}

	if err == errNoMatch {
		err = NotFoundError{Message: "Import path not valid:"}
	}

	return dir, err
}

// GetPresentation gets a presentation from the the given path.
func GetPresentation(client *http.Client, importPath string) (*Presentation, error) {
	ext := path.Ext(importPath)
	if ext != ".slide" && ext != ".article" {
		return nil, NotFoundError{Message: "unknown file extension."}
	}

	importPath, file := path.Split(importPath)
	importPath = strings.TrimSuffix(importPath, "/")
	for _, s := range services {
		if s.getPresentation == nil {
			continue
		}
		match, err := s.match(importPath)
		if err != nil {
			return nil, err
		}
		if match != nil {
			match["file"] = file
			return s.getPresentation(client, match)
		}
	}
	return nil, NotFoundError{Message: "path does not match registered service"}
}

// GetProject gets information about a repository.
func GetProject(client *http.Client, importPath string) (*Project, error) {
	for _, s := range services {
		if s.getProject == nil {
			continue
		}
		match, err := s.match(importPath)
		if err != nil {
			return nil, err
		}
		if match != nil {
			return s.getProject(client, match)
		}
	}
	return nil, NotFoundError{Message: "path does not match registered service"}
}
