// Copyright 2013 The Go Authors. All rights reserved.
//
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file or at
// https://developers.google.com/open-source/licenses/bsd.

// +build !appengine

package gosrc

import (
	"bytes"
	"errors"
	"io/ioutil"
	"log"
	"net/http"
	"os"
	"os/exec"
	"path"
	"path/filepath"
	"regexp"
	"strings"
	"time"
)

func init() {
	addService(&service{
		pattern: regexp.MustCompile(`^(?P<repo>(?:[a-z0-9.\-]+\.)+[a-z0-9.\-]+(?::[0-9]+)?/[A-Za-z0-9_.\-/]*?)\.(?P<vcs>bzr|git|hg|svn)(?P<dir>/[A-Za-z0-9_.\-/]*)?$`),
		prefix:  "",
		get:     getVCSDir,
	})
	getVCSDirFn = getVCSDir
}

const (
	lsRemoteTimeout = 5 * time.Minute
	cloneTimeout    = 10 * time.Minute
	fetchTimeout    = 5 * time.Minute
	checkoutTimeout = 1 * time.Minute
)

// Store temporary data in this directory.
var TempDir = filepath.Join(os.TempDir(), "gddo")

type urlTemplates struct {
	re         *regexp.Regexp
	fileBrowse string
	project    string
	line       string
}

var vcsServices = []*urlTemplates{
	{
		regexp.MustCompile(`^git\.gitorious\.org/(?P<repo>[^/]+/[^/]+)$`),
		"https://gitorious.org/{repo}/blobs/{tag}/{dir}{0}",
		"https://gitorious.org/{repo}",
		"%s#line%d",
	},
	{
		regexp.MustCompile(`^git\.oschina\.net/(?P<repo>[^/]+/[^/]+)$`),
		"http://git.oschina.net/{repo}/blob/{tag}/{dir}{0}",
		"http://git.oschina.net/{repo}",
		"%s#L%d",
	},
	{
		regexp.MustCompile(`^(?P<r1>[^.]+)\.googlesource.com/(?P<r2>[^./]+)$`),
		"https://{r1}.googlesource.com/{r2}/+/{tag}/{dir}{0}",
		"https://{r1}.googlesource.com/{r2}/+/{tag}",
		"%s#%d",
	},
	{
		regexp.MustCompile(`^gitcafe.com/(?P<repo>[^/]+/.[^/]+)$`),
		"https://gitcafe.com/{repo}/tree/{tag}/{dir}{0}",
		"https://gitcafe.com/{repo}",
		"",
	},
}

// lookupURLTemplate finds an expand() template, match map and line number
// format for well known repositories.
func lookupURLTemplate(repo, dir, tag string) (*urlTemplates, map[string]string) {
	if strings.HasPrefix(dir, "/") {
		dir = dir[1:] + "/"
	}
	for _, t := range vcsServices {
		if m := t.re.FindStringSubmatch(repo); m != nil {
			match := map[string]string{
				"dir": dir,
				"tag": tag,
			}
			for i, name := range t.re.SubexpNames() {
				if name != "" {
					match[name] = m[i]
				}
			}
			return t, match
		}
	}
	return &urlTemplates{}, nil
}

type vcsCmd struct {
	schemes  []string
	download func(schemes []string, clonePath, repo, savedEtag string) (tag, etag string, err error)
}

var vcsCmds = map[string]*vcsCmd{
	"git": {
		schemes:  []string{"http", "https", "ssh", "git"},
		download: downloadGit,
	},
	"svn": {
		schemes:  []string{"http", "https", "svn"},
		download: downloadSVN,
	},
}

var lsremoteRe = regexp.MustCompile(`(?m)^([0-9a-f]{40})\s+refs/(?:tags|heads)/(.+)$`)

func downloadGit(schemes []string, clonePath, repo, savedEtag string) (string, string, error) {
	var p []byte
	var scheme string
	for i := range schemes {
		cmd := exec.Command("git", "ls-remote", "--heads", "--tags", schemes[i]+"://"+clonePath)
		log.Println(strings.Join(cmd.Args, " "))
		var err error
		p, err = outputWithTimeout(cmd, lsRemoteTimeout)
		if err == nil {
			scheme = schemes[i]
			break
		}
	}

	if scheme == "" {
		return "", "", NotFoundError{Message: "VCS not found"}
	}

	tags := make(map[string]string)
	for _, m := range lsremoteRe.FindAllSubmatch(p, -1) {
		tags[string(m[2])] = string(m[1])
	}

	tag, commit, err := bestTag(tags, "master")
	if err != nil {
		return "", "", err
	}

	etag := scheme + "-" + commit

	if etag == savedEtag {
		return "", "", NotModifiedError{}
	}

	dir := filepath.Join(TempDir, repo+".git")
	p, err = ioutil.ReadFile(filepath.Join(dir, ".git", "HEAD"))
	switch {
	case err != nil:
		if err := os.MkdirAll(dir, 0777); err != nil {
			return "", "", err
		}
		cmd := exec.Command("git", "clone", scheme+"://"+clonePath, dir)
		log.Println(strings.Join(cmd.Args, " "))
		if err := runWithTimeout(cmd, cloneTimeout); err != nil {
			return "", "", err
		}
	case string(bytes.TrimRight(p, "\n")) == commit:
		return tag, etag, nil
	default:
		cmd := exec.Command("git", "fetch")
		log.Println(strings.Join(cmd.Args, " "))
		cmd.Dir = dir
		if err := runWithTimeout(cmd, fetchTimeout); err != nil {
			return "", "", err
		}
	}

	cmd := exec.Command("git", "checkout", "--detach", "--force", commit)
	cmd.Dir = dir
	if err := runWithTimeout(cmd, checkoutTimeout); err != nil {
		return "", "", err
	}

	return tag, etag, nil
}

func downloadSVN(schemes []string, clonePath, repo, savedEtag string) (string, string, error) {
	var scheme string
	var revno string
	for i := range schemes {
		var err error
		revno, err = getSVNRevision(schemes[i] + "://" + clonePath)
		if err == nil {
			scheme = schemes[i]
			break
		}
	}

	if scheme == "" {
		return "", "", NotFoundError{Message: "VCS not found"}
	}

	etag := scheme + "-" + revno
	if etag == savedEtag {
		return "", "", NotModifiedError{}
	}

	dir := filepath.Join(TempDir, repo+".svn")
	localRevno, err := getSVNRevision(dir)
	switch {
	case err != nil:
		log.Printf("err: %v", err)
		if err := os.MkdirAll(dir, 0777); err != nil {
			return "", "", err
		}
		cmd := exec.Command("svn", "checkout", scheme+"://"+clonePath, "-r", revno, dir)
		log.Println(strings.Join(cmd.Args, " "))
		if err := runWithTimeout(cmd, cloneTimeout); err != nil {
			return "", "", err
		}
	case localRevno != revno:
		cmd := exec.Command("svn", "update", "-r", revno)
		log.Println(strings.Join(cmd.Args, " "))
		cmd.Dir = dir
		if err := runWithTimeout(cmd, fetchTimeout); err != nil {
			return "", "", err
		}
	}

	return "", etag, nil
}

var svnrevRe = regexp.MustCompile(`(?m)^Last Changed Rev: ([0-9]+)$`)

func getSVNRevision(target string) (string, error) {
	cmd := exec.Command("svn", "info", target)
	log.Println(strings.Join(cmd.Args, " "))
	out, err := outputWithTimeout(cmd, lsRemoteTimeout)
	if err != nil {
		return "", err
	}
	match := svnrevRe.FindStringSubmatch(string(out))
	if match != nil {
		return match[1], nil
	}
	return "", NotFoundError{Message: "Last changed revision not found"}
}

func getVCSDir(client *http.Client, match map[string]string, etagSaved string) (*Directory, error) {
	cmd := vcsCmds[match["vcs"]]
	if cmd == nil {
		return nil, NotFoundError{Message: expand("VCS not supported: {vcs}", match)}
	}

	scheme := match["scheme"]
	if scheme == "" {
		i := strings.Index(etagSaved, "-")
		if i > 0 {
			scheme = etagSaved[:i]
		}
	}

	schemes := cmd.schemes
	if scheme != "" {
		for i := range cmd.schemes {
			if cmd.schemes[i] == scheme {
				schemes = cmd.schemes[i : i+1]
				break
			}
		}
	}

	clonePath, ok := match["clonePath"]
	if !ok {
		// clonePath may be unset if we're being called via the generic repo.vcs/dir regexp matcher.
		// In that case, set it to the repo value.
		clonePath = match["repo"]
	}

	// Download and checkout.

	tag, etag, err := cmd.download(schemes, clonePath, match["repo"], etagSaved)
	if err != nil {
		return nil, err
	}

	// Find source location.

	template, urlMatch := lookupURLTemplate(match["repo"], match["dir"], tag)

	// Slurp source files.

	d := filepath.Join(TempDir, filepath.FromSlash(expand("{repo}.{vcs}", match)), filepath.FromSlash(match["dir"]))
	f, err := os.Open(d)
	if err != nil {
		if os.IsNotExist(err) {
			err = NotFoundError{Message: err.Error()}
		}
		return nil, err
	}
	fis, err := f.Readdir(-1)
	if err != nil {
		return nil, err
	}

	var files []*File
	var subdirs []string
	for _, fi := range fis {
		switch {
		case fi.IsDir():
			if isValidPathElement(fi.Name()) {
				subdirs = append(subdirs, fi.Name())
			}
		case isDocFile(fi.Name()):
			b, err := ioutil.ReadFile(filepath.Join(d, fi.Name()))
			if err != nil {
				return nil, err
			}
			files = append(files, &File{
				Name:      fi.Name(),
				BrowseURL: expand(template.fileBrowse, urlMatch, fi.Name()),
				Data:      b,
			})
		}
	}

	return &Directory{
		LineFmt:        template.line,
		ProjectRoot:    expand("{repo}.{vcs}", match),
		ProjectName:    path.Base(match["repo"]),
		ProjectURL:     expand(template.project, urlMatch),
		BrowseURL:      "",
		Etag:           etag,
		VCS:            match["vcs"],
		Subdirectories: subdirs,
		Files:          files,
	}, nil
}

func runWithTimeout(cmd *exec.Cmd, timeout time.Duration) error {
	if err := cmd.Start(); err != nil {
		return err
	}
	t := time.AfterFunc(timeout, func() { cmd.Process.Kill() })
	defer t.Stop()
	return cmd.Wait()
}

func outputWithTimeout(cmd *exec.Cmd, timeout time.Duration) ([]byte, error) {
	if cmd.Stdout != nil {
		return nil, errors.New("exec: Stdout already set")
	}
	var b bytes.Buffer
	cmd.Stdout = &b
	err := runWithTimeout(cmd, timeout)
	return b.Bytes(), err
}
