// Copyright 2013 The Go Authors. All rights reserved.
//
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file or at
// https://developers.google.com/open-source/licenses/bsd.

package main

import (
	"fmt"
	"io/ioutil"
	"net/http"
	"regexp"
	"strings"

	"github.com/golang/gddo/doc"
)

func findExamples(pdoc *doc.Package, export, method string) []*doc.Example {
	if "package" == export {
		return pdoc.Examples
	}
	for _, f := range pdoc.Funcs {
		if f.Name == export {
			return f.Examples
		}
	}
	for _, t := range pdoc.Types {
		for _, f := range t.Funcs {
			if f.Name == export {
				return f.Examples
			}
		}
		if t.Name == export {
			if method == "" {
				return t.Examples
			}
			for _, m := range t.Methods {
				if method == m.Name {
					return m.Examples
				}
			}
			return nil
		}
	}
	return nil
}

func findExample(pdoc *doc.Package, export, method, name string) *doc.Example {
	for _, e := range findExamples(pdoc, export, method) {
		if name == e.Name {
			return e
		}
	}
	return nil
}

var exampleIDPat = regexp.MustCompile(`([^-]+)(?:-([^-]*)(?:-(.*))?)?`)

func playURL(pdoc *doc.Package, id, countryHeader string) (string, error) {
	if m := exampleIDPat.FindStringSubmatch(id); m != nil {
		if e := findExample(pdoc, m[1], m[2], m[3]); e != nil && e.Play != "" {
			req, err := http.NewRequest("POST", "https://play.golang.org/share", strings.NewReader(e.Play))
			if err != nil {
				return "", err
			}
			req.Header.Set("Content-Type", "text/plain")
			if countryHeader != "" {
				// Forward the App Engine country header.
				req.Header.Set("X-AppEngine-Country", countryHeader)
			}
			resp, err := httpClient.Do(req)
			if err != nil {
				return "", err
			}
			defer resp.Body.Close()
			p, err := ioutil.ReadAll(resp.Body)
			if err != nil {
				return "", err
			}
			if resp.StatusCode > 399 {
				return "", &httpError{
					status: resp.StatusCode,
					err:    fmt.Errorf("Error from play.golang.org: %s", p),
				}
			}
			return fmt.Sprintf("http://play.golang.org/p/%s", p), nil
		}
	}
	return "", &httpError{status: http.StatusNotFound}
}
