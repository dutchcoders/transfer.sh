// Copyright 2013 The Go Authors. All rights reserved.
//
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file or at
// https://developers.google.com/open-source/licenses/bsd.

package main

import (
	"net/url"
	"path"
	"regexp"
	"strings"
)

func importPathFromGoogleBrowse(m []string) string {
	project := m[1]
	dir := m[2]
	if dir == "" {
		dir = "/"
	} else if dir[len(dir)-1] == '/' {
		dir = dir[:len(dir)-1]
	}
	subrepo := ""
	if len(m[3]) > 0 {
		v, _ := url.ParseQuery(m[3][1:])
		subrepo = v.Get("repo")
		if len(subrepo) > 0 {
			subrepo = "." + subrepo
		}
	}
	if strings.HasPrefix(m[4], "#hg%2F") {
		d, _ := url.QueryUnescape(m[4][len("#hg%2f"):])
		if i := strings.IndexRune(d, '%'); i >= 0 {
			d = d[:i]
		}
		dir = dir + "/" + d
	}
	return "code.google.com/p/" + project + subrepo + dir
}

var browsePatterns = []struct {
	pat *regexp.Regexp
	fn  func([]string) string
}{
	{
		// GitHub tree  browser.
		regexp.MustCompile(`^https?://(github\.com/[^/]+/[^/]+)(?:/tree/[^/]+(/.*))?$`),
		func(m []string) string { return m[1] + m[2] },
	},
	{
		// GitHub file browser.
		regexp.MustCompile(`^https?://(github\.com/[^/]+/[^/]+)/blob/[^/]+/(.*)$`),
		func(m []string) string {
			d := path.Dir(m[2])
			if d == "." {
				return m[1]
			}
			return m[1] + "/" + d
		},
	},
	{
		// GitHub issues, pulls, etc.
		regexp.MustCompile(`^https?://(github\.com/[^/]+/[^/]+)(.*)$`),
		func(m []string) string { return m[1] },
	},
	{
		// Bitbucket source borwser.
		regexp.MustCompile(`^https?://(bitbucket\.org/[^/]+/[^/]+)(?:/src/[^/]+(/[^?]+)?)?`),
		func(m []string) string { return m[1] + m[2] },
	},
	{
		// Google Project Hosting source browser.
		regexp.MustCompile(`^http:/+code\.google\.com/p/([^/]+)/source/browse(/[^?#]*)?(\?[^#]*)?(#.*)?$`),
		importPathFromGoogleBrowse,
	},
	{
		// Launchpad source browser.
		regexp.MustCompile(`^https?:/+bazaar\.(launchpad\.net/.*)/files$`),
		func(m []string) string { return m[1] },
	},
	{
		regexp.MustCompile(`^https?://(.+)$`),
		func(m []string) string { return strings.Trim(m[1], "/") },
	},
}

// isBrowserURL returns importPath and true if URL looks like a URL for a VCS
// source browser.
func isBrowseURL(s string) (importPath string, ok bool) {
	for _, c := range browsePatterns {
		if m := c.pat.FindStringSubmatch(s); m != nil {
			return c.fn(m), true
		}
	}
	return "", false
}
