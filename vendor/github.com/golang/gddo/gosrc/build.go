// Copyright 2013 The Go Authors. All rights reserved.
//
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file or at
// https://developers.google.com/open-source/licenses/bsd.

package gosrc

import (
	"bytes"
	"go/build"
	"io"
	"io/ioutil"
	"os"
	"path"
	"strings"
	"time"
)

// Import returns details about the package in the directory.
func (dir *Directory) Import(ctx *build.Context, mode build.ImportMode) (*build.Package, error) {
	safeCopy := *ctx
	ctx = &safeCopy
	ctx.JoinPath = path.Join
	ctx.IsAbsPath = path.IsAbs
	ctx.SplitPathList = func(list string) []string { return strings.Split(list, ":") }
	ctx.IsDir = func(path string) bool { return false }
	ctx.HasSubdir = func(root, dir string) (rel string, ok bool) { return "", false }
	ctx.ReadDir = dir.readDir
	ctx.OpenFile = dir.openFile
	return ctx.ImportDir(".", mode)
}

type fileInfo struct{ f *File }

func (fi fileInfo) Name() string       { return fi.f.Name }
func (fi fileInfo) Size() int64        { return int64(len(fi.f.Data)) }
func (fi fileInfo) Mode() os.FileMode  { return 0 }
func (fi fileInfo) ModTime() time.Time { return time.Time{} }
func (fi fileInfo) IsDir() bool        { return false }
func (fi fileInfo) Sys() interface{}   { return nil }

func (dir *Directory) readDir(name string) ([]os.FileInfo, error) {
	if name != "." {
		return nil, os.ErrNotExist
	}
	fis := make([]os.FileInfo, len(dir.Files))
	for i, f := range dir.Files {
		fis[i] = fileInfo{f}
	}
	return fis, nil
}

func (dir *Directory) openFile(path string) (io.ReadCloser, error) {
	name := strings.TrimPrefix(path, "./")
	for _, f := range dir.Files {
		if f.Name == name {
			return ioutil.NopCloser(bytes.NewReader(f.Data)), nil
		}
	}
	return nil, os.ErrNotExist
}
