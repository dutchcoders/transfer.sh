// Copyright 2013 The Go Authors. All rights reserved.
//
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file or at
// https://developers.google.com/open-source/licenses/bsd.

package gosrc

import (
	"go/build"
	"io/ioutil"
	"path/filepath"
	"strconv"
	"time"
)

var localPath string

// SetLocalDevMode sets the package to local development mode. In this mode,
// the GOPATH specified by path is used to find directories instead of version
// control services.
func SetLocalDevMode(path string) {
	localPath = path
}

func getLocal(importPath string) (*Directory, error) {
	ctx := build.Default
	if localPath != "" {
		ctx.GOPATH = localPath
	}
	bpkg, err := ctx.Import(importPath, ".", build.FindOnly)
	if err != nil {
		return nil, err
	}
	dir := filepath.Join(bpkg.SrcRoot, filepath.FromSlash(importPath))
	fis, err := ioutil.ReadDir(dir)
	if err != nil {
		return nil, err
	}
	var modTime time.Time
	var files []*File
	for _, fi := range fis {
		if fi.IsDir() || !isDocFile(fi.Name()) {
			continue
		}
		if fi.ModTime().After(modTime) {
			modTime = fi.ModTime()
		}
		b, err := ioutil.ReadFile(filepath.Join(dir, fi.Name()))
		if err != nil {
			return nil, err
		}
		files = append(files, &File{
			Name: fi.Name(),
			Data: b,
		})
	}
	return &Directory{
		ImportPath: importPath,
		Etag:       strconv.FormatInt(modTime.Unix(), 16),
		Files:      files,
	}, nil
}
