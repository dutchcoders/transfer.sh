// Copyright 2013 The Go Authors. All rights reserved.
//
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file or at
// https://developers.google.com/open-source/licenses/bsd.

package gosrc

import (
	"archive/tar"
	"bytes"
	"compress/gzip"
	"crypto/md5"
	"encoding/hex"
	"io"
	"net/http"
	"path"
	"regexp"
	"sort"
	"strings"
)

func init() {
	addService(&service{
		pattern: regexp.MustCompile(`^launchpad\.net/(?P<repo>(?P<project>[a-z0-9A-Z_.\-]+)(?P<series>/[a-z0-9A-Z_.\-]+)?|~[a-z0-9A-Z_.\-]+/(\+junk|[a-z0-9A-Z_.\-]+)/[a-z0-9A-Z_.\-]+)(?P<dir>/[a-z0-9A-Z_.\-/]+)*$`),
		prefix:  "launchpad.net/",
		get:     getLaunchpadDir,
	})
}

type byHash []byte

func (p byHash) Len() int { return len(p) / md5.Size }
func (p byHash) Less(i, j int) bool {
	return -1 == bytes.Compare(p[i*md5.Size:(i+1)*md5.Size], p[j*md5.Size:(j+1)*md5.Size])
}
func (p byHash) Swap(i, j int) {
	var temp [md5.Size]byte
	copy(temp[:], p[i*md5.Size:])
	copy(p[i*md5.Size:(i+1)*md5.Size], p[j*md5.Size:])
	copy(p[j*md5.Size:], temp[:])
}

func getLaunchpadDir(client *http.Client, match map[string]string, savedEtag string) (*Directory, error) {
	c := &httpClient{client: client}

	if match["project"] != "" && match["series"] != "" {
		rc, err := c.getReader(expand("https://code.launchpad.net/{project}{series}/.bzr/branch-format", match))
		switch {
		case err == nil:
			rc.Close()
			// The structure of the import path is launchpad.net/{root}/{dir}.
		case IsNotFound(err):
			// The structure of the import path is is launchpad.net/{project}/{dir}.
			match["repo"] = match["project"]
			match["dir"] = expand("{series}{dir}", match)
		default:
			return nil, err
		}
	}

	p, err := c.getBytes(expand("https://bazaar.launchpad.net/+branch/{repo}/tarball", match))
	if err != nil {
		return nil, err
	}

	gzr, err := gzip.NewReader(bytes.NewReader(p))
	if err != nil {
		return nil, err
	}
	defer gzr.Close()

	tr := tar.NewReader(gzr)

	var hash []byte
	inTree := false
	dirPrefix := expand("+branch/{repo}{dir}/", match)
	var files []*File
	for {
		h, err := tr.Next()
		if err == io.EOF {
			break
		}
		if err != nil {
			return nil, err
		}
		d, f := path.Split(h.Name)
		if !isDocFile(f) {
			continue
		}
		b := make([]byte, h.Size)
		if _, err := io.ReadFull(tr, b); err != nil {
			return nil, err
		}

		m := md5.New()
		m.Write(b)
		hash = m.Sum(hash)

		if !strings.HasPrefix(h.Name, dirPrefix) {
			continue
		}
		inTree = true
		if d == dirPrefix {
			files = append(files, &File{
				Name:      f,
				BrowseURL: expand("http://bazaar.launchpad.net/+branch/{repo}/view/head:{dir}/{0}", match, f),
				Data:      b})
		}
	}

	if !inTree {
		return nil, NotFoundError{Message: "Directory tree does not contain Go files."}
	}

	sort.Sort(byHash(hash))
	m := md5.New()
	m.Write(hash)
	hash = m.Sum(hash[:0])
	etag := hex.EncodeToString(hash)
	if etag == savedEtag {
		return nil, NotModifiedError{}
	}

	return &Directory{
		BrowseURL:   expand("http://bazaar.launchpad.net/+branch/{repo}/view/head:{dir}/", match),
		Etag:        etag,
		Files:       files,
		LineFmt:     "%s#L%d",
		ProjectName: match["repo"],
		ProjectRoot: expand("launchpad.net/{repo}", match),
		ProjectURL:  expand("https://launchpad.net/{repo}/", match),
		VCS:         "bzr",
	}, nil
}
