// Copyright 2013 The Go Authors. All rights reserved.
//
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file or at
// https://developers.google.com/open-source/licenses/bsd.

package gosrc

import (
	"errors"
	"net/http"
	"net/url"
	"regexp"
	"strings"
)

func init() {
	addService(&service{
		pattern:         regexp.MustCompile(`^code\.google\.com/(?P<pr>[pr])/(?P<repo>[a-z0-9\-]+)(:?\.(?P<subrepo>[a-z0-9\-]+))?(?P<dir>/[a-z0-9A-Z_.\-/]+)?$`),
		prefix:          "code.google.com/",
		get:             getGoogleDir,
		getPresentation: getGooglePresentation,
	})
}

var (
	googleRepoRe     = regexp.MustCompile(`id="checkoutcmd">(hg|git|svn)`)
	googleRevisionRe = regexp.MustCompile(`<h2>(?:[^ ]+ - )?Revision *([^:]+):`)
	googleEtagRe     = regexp.MustCompile(`^(hg|git|svn)-`)
	googleFileRe     = regexp.MustCompile(`<li><a href="([^"]+)"`)
)

func checkGoogleRedir(c *httpClient, match map[string]string) error {
	resp, err := c.getNoFollow(expand("https://code.google.com/{pr}/{repo}/", match))
	if err != nil {
		return err
	}
	defer resp.Body.Close()
	if resp.StatusCode == http.StatusOK {
		return nil
	}
	if resp.StatusCode == http.StatusMovedPermanently {
		if u, err := url.Parse(resp.Header.Get("Location")); err == nil {
			p := u.Host + u.Path + match["dir"]
			return NotFoundError{Message: "Project moved", Redirect: p}
		}
	}
	return c.err(resp)
}

func getGoogleDir(client *http.Client, match map[string]string, savedEtag string) (*Directory, error) {
	setupGoogleMatch(match)
	c := &httpClient{client: client}

	if err := checkGoogleRedir(c, match); err != nil {
		return nil, err
	}

	if m := googleEtagRe.FindStringSubmatch(savedEtag); m != nil {
		match["vcs"] = m[1]
	} else if err := getGoogleVCS(c, match); err != nil {
		return nil, err
	}

	// Scrape the repo browser to find the project revision and individual Go files.
	p, err := c.getBytes(expand("http://{subrepo}{dot}{repo}.googlecode.com/{vcs}{dir}/", match))
	if err != nil {
		return nil, err
	}

	var etag string
	m := googleRevisionRe.FindSubmatch(p)
	if m == nil {
		return nil, errors.New("Could not find revision for " + match["importPath"])
	}
	etag = expand("{vcs}-{0}", match, string(m[1]))
	if etag == savedEtag {
		return nil, NotModifiedError{}
	}

	var subdirs []string
	var files []*File
	var dataURLs []string
	for _, m := range googleFileRe.FindAllSubmatch(p, -1) {
		fname := string(m[1])
		switch {
		case strings.HasSuffix(fname, "/"):
			fname = fname[:len(fname)-1]
			if isValidPathElement(fname) {
				subdirs = append(subdirs, fname)
			}
		case isDocFile(fname):
			files = append(files, &File{Name: fname, BrowseURL: expand("http://code.google.com/{pr}/{repo}/source/browse{dir}/{0}{query}", match, fname)})
			dataURLs = append(dataURLs, expand("http://{subrepo}{dot}{repo}.googlecode.com/{vcs}{dir}/{0}", match, fname))
		}
	}

	if err := c.getFiles(dataURLs, files); err != nil {
		return nil, err
	}

	var projectURL string
	if match["subrepo"] == "" {
		projectURL = expand("https://code.google.com/{pr}/{repo}/", match)
	} else {
		projectURL = expand("https://code.google.com/{pr}/{repo}/source/browse?repo={subrepo}", match)
	}

	return &Directory{
		BrowseURL:   expand("http://code.google.com/{pr}/{repo}/source/browse{dir}/{query}", match),
		Etag:        etag,
		Files:       files,
		LineFmt:     "%s#%d",
		ProjectName: expand("{repo}{dot}{subrepo}", match),
		ProjectRoot: expand("code.google.com/{pr}/{repo}{dot}{subrepo}", match),
		ProjectURL:  projectURL,
		VCS:         match["vcs"],
	}, nil
}

func setupGoogleMatch(match map[string]string) {
	if s := match["subrepo"]; s != "" {
		match["dot"] = "."
		match["query"] = "?repo=" + s
	} else {
		match["dot"] = ""
		match["query"] = ""
	}
}

func getGoogleVCS(c *httpClient, match map[string]string) error {
	// Scrape the HTML project page to find the VCS.
	p, err := c.getBytes(expand("http://code.google.com/{pr}/{repo}/source/checkout", match))
	if err != nil {
		return err
	}
	m := googleRepoRe.FindSubmatch(p)
	if m == nil {
		return NotFoundError{Message: "Could not find VCS on Google Code project page."}
	}
	match["vcs"] = string(m[1])
	return nil
}

func getGooglePresentation(client *http.Client, match map[string]string) (*Presentation, error) {
	c := &httpClient{client: client}

	setupGoogleMatch(match)
	if err := getGoogleVCS(c, match); err != nil {
		return nil, err
	}

	rawBase, err := url.Parse(expand("http://{subrepo}{dot}{repo}.googlecode.com/{vcs}{dir}/", match))
	if err != nil {
		return nil, err
	}

	p, err := c.getBytes(expand("http://{subrepo}{dot}{repo}.googlecode.com/{vcs}{dir}/{file}", match))
	if err != nil {
		return nil, err
	}

	b := &presBuilder{
		data:     p,
		filename: match["file"],
		fetch: func(fnames []string) ([]*File, error) {
			var files []*File
			var dataURLs []string
			for _, fname := range fnames {
				u, err := rawBase.Parse(fname)
				if err != nil {
					return nil, err
				}
				files = append(files, &File{Name: fname})
				dataURLs = append(dataURLs, u.String())
			}
			err := c.getFiles(dataURLs, files)
			return files, err
		},
		resolveURL: func(fname string) string {
			u, err := rawBase.Parse(fname)
			if err != nil {
				return "/notfound"
			}
			return u.String()
		},
	}

	return b.build()
}
