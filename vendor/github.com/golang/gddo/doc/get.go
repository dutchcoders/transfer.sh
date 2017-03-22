// Copyright 2013 The Go Authors. All rights reserved.
//
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file or at
// https://developers.google.com/open-source/licenses/bsd.

// Package doc fetches Go package documentation from version control services.
package doc

import (
	"github.com/golang/gddo/gosrc"
	"go/doc"
	"net/http"
	"strings"
)

func Get(client *http.Client, importPath string, etag string) (*Package, error) {

	const versionPrefix = PackageVersion + "-"

	if strings.HasPrefix(etag, versionPrefix) {
		etag = etag[len(versionPrefix):]
	} else {
		etag = ""
	}

	dir, err := gosrc.Get(client, importPath, etag)
	if err != nil {
		return nil, err
	}

	pdoc, err := newPackage(dir)
	if err != nil {
		return pdoc, err
	}

	if pdoc.Synopsis == "" &&
		pdoc.Doc == "" &&
		!pdoc.IsCmd &&
		pdoc.Name != "" &&
		dir.ImportPath == dir.ProjectRoot &&
		len(pdoc.Errors) == 0 {
		project, err := gosrc.GetProject(client, dir.ResolvedPath)
		switch {
		case err == nil:
			pdoc.Synopsis = doc.Synopsis(project.Description)
		case gosrc.IsNotFound(err):
			// ok
		default:
			return nil, err
		}
	}

	return pdoc, nil
}
