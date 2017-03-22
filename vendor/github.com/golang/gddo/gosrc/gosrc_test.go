// Copyright 2014 The Go Authors. All rights reserved.
//
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file or at
// https://developers.google.com/open-source/licenses/bsd.

package gosrc

import (
	"fmt"
	"io/ioutil"
	"net/http"
	"path"
	"reflect"
	"regexp"
	"strings"
	"testing"
)

var testWeb = map[string]string{
	// Package at root of a GitHub repo.
	"https://alice.org/pkg": `<head> <meta name="go-import" content="alice.org/pkg git https://github.com/alice/pkg"></head>`,
	// Package in sub-diretory.
	"https://alice.org/pkg/sub": `<head> <meta name="go-import" content="alice.org/pkg git https://github.com/alice/pkg"><body>`,
	// Fallback to http.
	"http://alice.org/pkg/http": `<head> <meta name="go-import" content="alice.org/pkg git https://github.com/alice/pkg">`,
	// Meta tag in sub-directory does not match meta tag at root.
	"https://alice.org/pkg/mismatch": `<head> <meta name="go-import" content="alice.org/pkg hg https://github.com/alice/pkg">`,
	// More than one matching meta tag.
	"http://alice.org/pkg/multiple": `<head> ` +
		`<meta name="go-import" content="alice.org/pkg git https://github.com/alice/pkg">` +
		`<meta name="go-import" content="alice.org/pkg git https://github.com/alice/pkg">`,
	// Package with go-source meta tag.
	"https://alice.org/pkg/source": `<head>` +
		`<meta name="go-import" content="alice.org/pkg git https://github.com/alice/pkg">` +
		`<meta name="go-source" content="alice.org/pkg http://alice.org/pkg http://alice.org/pkg{/dir} http://alice.org/pkg{/dir}?f={file}#Line{line}">`,
	"https://alice.org/pkg/ignore": `<head>` +
		`<title>Hello</title>` +
		// Unknown meta name
		`<meta name="go-junk" content="alice.org/pkg http://alice.org/pkg http://alice.org/pkg{/dir} http://alice.org/pkg{/dir}?f={file}#Line{line}">` +
		// go-source before go-meta
		`<meta name="go-source" content="alice.org/pkg http://alice.org/pkg http://alice.org/pkg{/dir} http://alice.org/pkg{/dir}?f={file}#Line{line}">` +
		// go-import tag for the package
		`<meta name="go-import" content="alice.org/pkg git https://github.com/alice/pkg">` +
		// go-import with wrong number of fields
		`<meta name="go-import" content="alice.org/pkg https://github.com/alice/pkg">` +
		// go-import with no fields
		`<meta name="go-import" content="">` +
		// go-source with wrong number of fields
		`<meta name="go-source" content="alice.org/pkg blah">` +
		// meta tag for a different package
		`<meta name="go-import" content="alice.org/other git https://github.com/alice/other">` +
		// meta tag for a different package
		`<meta name="go-import" content="alice.org/other git https://github.com/alice/other">` +
		`</head>` +
		// go-import outside of head
		`<meta name="go-import" content="alice.org/pkg git https://github.com/alice/pkg">`,

	// Package at root of a Git repo.
	"https://bob.com/pkg": `<head> <meta name="go-import" content="bob.com/pkg git https://vcs.net/bob/pkg.git">`,
	// Package at in sub-directory of a Git repo.
	"https://bob.com/pkg/sub": `<head> <meta name="go-import" content="bob.com/pkg git https://vcs.net/bob/pkg.git">`,
	// Package with go-source meta tag.
	"https://bob.com/pkg/source": `<head>` +
		`<meta name="go-import" content="bob.com/pkg git https://vcs.net/bob/pkg.git">` +
		`<meta name="go-source" content="bob.com/pkg http://bob.com/pkg http://bob.com/pkg{/dir}/ http://bob.com/pkg{/dir}/?f={file}#Line{line}">`,
	// Meta refresh to godoc.org
	"http://rsc.io/benchstat": `<!DOCTYPE html><html><head>` +
		`<meta http-equiv="Content-Type" content="text/html; charset=utf-8"/>` +
		`<meta name="go-import" content="rsc.io/benchstat git https://github.com/rsc/benchstat">` +
		`<meta http-equiv="refresh" content="0; url=https://godoc.org/rsc.io/benchstat">` +
		`</head>`,

	// Package with go-source meta tag, where {file} appears on the right of '#' in the file field URL template.
	"https://azul3d.org/examples": `<!DOCTYPE html><html><head>` +
		`<meta http-equiv="Content-Type" content="text/html; charset=utf-8"/>` +
		`<meta name="go-import" content="azul3d.org/examples git https://github.com/azul3d/examples">` +
		`<meta name="go-source" content="azul3d.org/examples https://github.com/azul3d/examples https://gotools.org/azul3d.org/examples{/dir} https://gotools.org/azul3d.org/examples{/dir}#{file}-L{line}">` +
		`<meta http-equiv="refresh" content="0; url=https://godoc.org/azul3d.org/examples">` +
		`</head>`,
	"https://azul3d.org/examples/abs": `<!DOCTYPE html><html><head>` +
		`<meta http-equiv="Content-Type" content="text/html; charset=utf-8"/>` +
		`<meta name="go-import" content="azul3d.org/examples git https://github.com/azul3d/examples">` +
		`<meta name="go-source" content="azul3d.org/examples https://github.com/azul3d/examples https://gotools.org/azul3d.org/examples{/dir} https://gotools.org/azul3d.org/examples{/dir}#{file}-L{line}">` +
		`<meta http-equiv="refresh" content="0; url=https://godoc.org/azul3d.org/examples/abs">` +
		`</head>`,
}

var getDynamicTests = []struct {
	importPath string
	dir        *Directory
}{
	{"alice.org/pkg", &Directory{
		BrowseURL:    "https://github.com/alice/pkg",
		ImportPath:   "alice.org/pkg",
		LineFmt:      "%s#L%d",
		ProjectName:  "pkg",
		ProjectRoot:  "alice.org/pkg",
		ProjectURL:   "https://alice.org/pkg",
		ResolvedPath: "github.com/alice/pkg",
		VCS:          "git",
		Files:        []*File{{Name: "main.go", BrowseURL: "https://github.com/alice/pkg/blob/master/main.go"}},
	}},
	{"alice.org/pkg/sub", &Directory{
		BrowseURL:    "https://github.com/alice/pkg/tree/master/sub",
		ImportPath:   "alice.org/pkg/sub",
		LineFmt:      "%s#L%d",
		ProjectName:  "pkg",
		ProjectRoot:  "alice.org/pkg",
		ProjectURL:   "https://alice.org/pkg",
		ResolvedPath: "github.com/alice/pkg/sub",
		VCS:          "git",
		Files:        []*File{{Name: "main.go", BrowseURL: "https://github.com/alice/pkg/blob/master/sub/main.go"}},
	}},
	{"alice.org/pkg/http", &Directory{
		BrowseURL:    "https://github.com/alice/pkg/tree/master/http",
		ImportPath:   "alice.org/pkg/http",
		LineFmt:      "%s#L%d",
		ProjectName:  "pkg",
		ProjectRoot:  "alice.org/pkg",
		ProjectURL:   "https://alice.org/pkg",
		ResolvedPath: "github.com/alice/pkg/http",
		VCS:          "git",
		Files:        []*File{{Name: "main.go", BrowseURL: "https://github.com/alice/pkg/blob/master/http/main.go"}},
	}},
	{"alice.org/pkg/source", &Directory{
		BrowseURL:    "http://alice.org/pkg/source",
		ImportPath:   "alice.org/pkg/source",
		LineFmt:      "%s#Line%d",
		ProjectName:  "pkg",
		ProjectRoot:  "alice.org/pkg",
		ProjectURL:   "http://alice.org/pkg",
		ResolvedPath: "github.com/alice/pkg/source",
		VCS:          "git",
		Files:        []*File{{Name: "main.go", BrowseURL: "http://alice.org/pkg/source?f=main.go"}},
	}},
	{"alice.org/pkg/ignore", &Directory{
		BrowseURL:    "http://alice.org/pkg/ignore",
		ImportPath:   "alice.org/pkg/ignore",
		LineFmt:      "%s#Line%d",
		ProjectName:  "pkg",
		ProjectRoot:  "alice.org/pkg",
		ProjectURL:   "http://alice.org/pkg",
		ResolvedPath: "github.com/alice/pkg/ignore",
		VCS:          "git",
		Files:        []*File{{Name: "main.go", BrowseURL: "http://alice.org/pkg/ignore?f=main.go"}},
	}},
	{"alice.org/pkg/mismatch", nil},
	{"alice.org/pkg/multiple", nil},
	{"alice.org/pkg/notfound", nil},

	{"bob.com/pkg", &Directory{
		ImportPath:   "bob.com/pkg",
		ProjectName:  "pkg",
		ProjectRoot:  "bob.com/pkg",
		ProjectURL:   "https://bob.com/pkg",
		ResolvedPath: "vcs.net/bob/pkg.git",
		VCS:          "git",
		Files:        []*File{{Name: "main.go"}},
	}},
	{"bob.com/pkg/sub", &Directory{
		ImportPath:   "bob.com/pkg/sub",
		ProjectName:  "pkg",
		ProjectRoot:  "bob.com/pkg",
		ProjectURL:   "https://bob.com/pkg",
		ResolvedPath: "vcs.net/bob/pkg.git/sub",
		VCS:          "git",
		Files:        []*File{{Name: "main.go"}},
	}},
	{"bob.com/pkg/source", &Directory{
		BrowseURL:    "http://bob.com/pkg/source/",
		ImportPath:   "bob.com/pkg/source",
		LineFmt:      "%s#Line%d",
		ProjectName:  "pkg",
		ProjectRoot:  "bob.com/pkg",
		ProjectURL:   "http://bob.com/pkg",
		ResolvedPath: "vcs.net/bob/pkg.git/source",
		VCS:          "git",
		Files:        []*File{{Name: "main.go", BrowseURL: "http://bob.com/pkg/source/?f=main.go"}},
	}},
	{"rsc.io/benchstat", &Directory{
		BrowseURL:    "https://github.com/rsc/benchstat",
		ImportPath:   "rsc.io/benchstat",
		LineFmt:      "%s#L%d",
		ProjectName:  "benchstat",
		ProjectRoot:  "rsc.io/benchstat",
		ProjectURL:   "https://github.com/rsc/benchstat",
		ResolvedPath: "github.com/rsc/benchstat",
		VCS:          "git",
		Files:        []*File{{Name: "main.go", BrowseURL: "https://github.com/rsc/benchstat/blob/master/main.go"}},
	}},
	{"azul3d.org/examples/abs", &Directory{
		BrowseURL:    "https://gotools.org/azul3d.org/examples/abs",
		ImportPath:   "azul3d.org/examples/abs",
		LineFmt:      "%s-L%d",
		ProjectName:  "examples",
		ProjectRoot:  "azul3d.org/examples",
		ProjectURL:   "https://github.com/azul3d/examples",
		ResolvedPath: "github.com/azul3d/examples/abs",
		VCS:          "git",
		Files:        []*File{{Name: "main.go", BrowseURL: "https://gotools.org/azul3d.org/examples/abs#main.go"}},
	}},
}

type testTransport map[string]string

func (t testTransport) RoundTrip(req *http.Request) (*http.Response, error) {
	statusCode := http.StatusOK
	req.URL.RawQuery = ""
	body, ok := t[req.URL.String()]
	if !ok {
		statusCode = http.StatusNotFound
	}
	resp := &http.Response{
		StatusCode: statusCode,
		Body:       ioutil.NopCloser(strings.NewReader(body)),
	}
	return resp, nil
}

var githubPattern = regexp.MustCompile(`^github\.com/(?P<owner>[a-z0-9A-Z_.\-]+)/(?P<repo>[a-z0-9A-Z_.\-]+)(?P<dir>/[a-z0-9A-Z_.\-/]*)?$`)

func testGet(client *http.Client, match map[string]string, etag string) (*Directory, error) {
	importPath := match["importPath"]

	if m := githubPattern.FindStringSubmatch(importPath); m != nil {
		browseURL := fmt.Sprintf("https://github.com/%s/%s", m[1], m[2])
		if m[3] != "" {
			browseURL = fmt.Sprintf("%s/tree/master%s", browseURL, m[3])
		}
		return &Directory{
			BrowseURL:   browseURL,
			ImportPath:  importPath,
			LineFmt:     "%s#L%d",
			ProjectName: m[2],
			ProjectRoot: fmt.Sprintf("github.com/%s/%s", m[1], m[2]),
			ProjectURL:  fmt.Sprintf("https://github.com/%s/%s", m[1], m[2]),
			VCS:         "git",
			Files: []*File{{
				Name:      "main.go",
				BrowseURL: fmt.Sprintf("https://github.com/%s/%s/blob/master%s/main.go", m[1], m[2], m[3]),
			}},
		}, nil
	}

	if strings.HasPrefix(match["repo"], "vcs.net") {
		return &Directory{
			ImportPath:  importPath,
			ProjectName: path.Base(match["repo"]),
			ProjectRoot: fmt.Sprintf("%s.%s", match["repo"], match["vcs"]),
			VCS:         match["vcs"],
			Files:       []*File{{Name: "main.go"}},
		}, nil
	}

	return nil, errNoMatch
}

func TestGetDynamic(t *testing.T) {
	savedServices := services
	savedGetVCSDirFn := getVCSDirFn
	defer func() {
		services = savedServices
		getVCSDirFn = savedGetVCSDirFn
	}()
	services = []*service{{pattern: regexp.MustCompile(".*"), get: testGet}}
	getVCSDirFn = testGet
	client := &http.Client{Transport: testTransport(testWeb)}

	for _, tt := range getDynamicTests {
		dir, err := getDynamic(client, tt.importPath, "")

		if tt.dir == nil {
			if err == nil {
				t.Errorf("getDynamic(client, %q, etag) did not return expected error", tt.importPath)
			}
			continue
		}

		if err != nil {
			t.Errorf("getDynamic(client, %q, etag) return unexpected error: %v", tt.importPath, err)
			continue
		}

		if !reflect.DeepEqual(dir, tt.dir) {
			t.Errorf("getDynamic(client, %q, etag) =\n     %+v,\nwant %+v", tt.importPath, dir, tt.dir)
			for i, f := range dir.Files {
				var want *File
				if i < len(tt.dir.Files) {
					want = tt.dir.Files[i]
				}
				t.Errorf("file %d = %+v, want %+v", i, f, want)
			}
		}
	}
}
