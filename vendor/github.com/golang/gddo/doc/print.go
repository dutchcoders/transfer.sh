// Copyright 2013 The Go Authors. All rights reserved.
//
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file or at
// https://developers.google.com/open-source/licenses/bsd.

// +build ignore

// Command print fetches and prints package documentation.
//
// Usage: go run print.go importPath
package main

import (
	"flag"
	"log"
	"net/http"
	"os"

	"github.com/davecgh/go-spew/spew"
	"github.com/golang/gddo/doc"
	"github.com/golang/gddo/gosrc"
)

var (
	etag  = flag.String("etag", "", "Etag")
	local = flag.Bool("local", false, "Get package from local directory.")
)

func main() {
	flag.Parse()
	if len(flag.Args()) != 1 {
		log.Fatal("Usage: go run print.go importPath")
	}
	path := flag.Args()[0]

	var (
		pdoc *doc.Package
		err  error
	)
	if *local {
		gosrc.SetLocalDevMode(os.Getenv("GOPATH"))
	}
	pdoc, err = doc.Get(http.DefaultClient, path, *etag)
	//}
	if err != nil {
		log.Fatal(err)
	}
	spew.Dump(pdoc)
}
