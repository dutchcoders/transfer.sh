// Copyright 2014 The Go Authors. All rights reserved.
//
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file or at
// https://developers.google.com/open-source/licenses/bsd.

package gosrc

import (
	"errors"
	"net/http"
	"regexp"
	"strings"
)

var (
	golangBuildVersionRe = regexp.MustCompile(`Build version ([-+:. 0-9A-Za-z]+)`)
	golangFileRe         = regexp.MustCompile(`<a href="([^"]+)"`)
)

func getStandardDir(client *http.Client, importPath string, savedEtag string) (*Directory, error) {
	c := &httpClient{client: client}

	browseURL := "https://golang.org/src/" + importPath + "/"
	p, err := c.getBytes(browseURL)
	if err != nil {
		return nil, err
	}

	var etag string
	m := golangBuildVersionRe.FindSubmatch(p)
	if m == nil {
		return nil, errors.New("Could not find revision for " + importPath)
	}
	etag = strings.Trim(string(m[1]), ". ")
	if etag == savedEtag {
		return nil, NotModifiedError{}
	}

	var files []*File
	var dataURLs []string
	for _, m := range golangFileRe.FindAllSubmatch(p, -1) {
		fname := string(m[1])
		if isDocFile(fname) {
			files = append(files, &File{Name: fname, BrowseURL: browseURL + fname})
			dataURLs = append(dataURLs, browseURL+fname+"?m=text")
		}
	}

	if err := c.getFiles(dataURLs, files); err != nil {
		return nil, err
	}

	return &Directory{
		BrowseURL:    browseURL,
		Etag:         etag,
		Files:        files,
		ImportPath:   importPath,
		LineFmt:      "%s#L%d",
		ProjectName:  "Go",
		ProjectRoot:  "",
		ProjectURL:   "https://golang.org/",
		ResolvedPath: importPath,
	}, nil
}
