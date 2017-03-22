// Copyright 2013 The Go Authors. All rights reserved.
//
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file or at
// https://developers.google.com/open-source/licenses/bsd.

package gosrc

import (
	"path/filepath"
	"regexp"
	"time"
)

type Presentation struct {
	Filename string
	Files    map[string][]byte
	Updated  time.Time
}

type presBuilder struct {
	filename   string
	data       []byte
	resolveURL func(fname string) string
	fetch      func(fnames []string) ([]*File, error)
}

var assetPat = regexp.MustCompile(`(?m)^\.(play|code|image|iframe|html)\s+(?:-\S+\s+)*(\S+)`)

func (b *presBuilder) build() (*Presentation, error) {
	var data []byte
	var fnames []string
	i := 0
	for _, m := range assetPat.FindAllSubmatchIndex(b.data, -1) {
		name := filepath.Clean(string(b.data[m[4]:m[5]]))
		switch string(b.data[m[2]:m[3]]) {
		case "iframe", "image":
			data = append(data, b.data[i:m[4]]...)
			data = append(data, b.resolveURL(name)...)
		case "html":
			// TODO: sanitize and fix relative URLs in HTML.
			data = append(data, "\nERROR: .html not supported\n"...)
		case "play", "code":
			data = append(data, b.data[i:m[5]]...)
			found := false
			for _, n := range fnames {
				if n == name {
					found = true
					break
				}
			}
			if !found {
				fnames = append(fnames, name)
			}
		default:
			data = append(data, "\nERROR: unknown command\n"...)
		}
		i = m[5]
	}
	data = append(data, b.data[i:]...)
	files, err := b.fetch(fnames)
	if err != nil {
		return nil, err
	}
	pres := &Presentation{
		Updated:  time.Now().UTC(),
		Filename: b.filename,
		Files:    map[string][]byte{b.filename: data},
	}
	for _, f := range files {
		pres.Files[f.Name] = f.Data
	}
	return pres, nil
}
