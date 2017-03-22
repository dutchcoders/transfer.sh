// Copyright 2013 The Go Authors. All rights reserved.
//
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file or at
// https://developers.google.com/open-source/licenses/bsd.

// +build ignore

// Command print fetches and prints package.
//
// Usage: go run print.go importPath
package main

import (
	"flag"
	"fmt"
	"log"
	"net/http"
	"strings"

	"github.com/golang/gddo/gosrc"
)

var (
	etag    = flag.String("etag", "", "Etag")
	local   = flag.String("local", "", "Get package from local workspace.")
	present = flag.Bool("present", false, "Get presentation.")
)

func main() {
	flag.Parse()
	if len(flag.Args()) != 1 {
		log.Fatal("Usage: go run print.go importPath")
	}
	if *present {
		printPresentation(flag.Args()[0])
	} else {
		printDir(flag.Args()[0])
	}
}

func printDir(path string) {
	if *local != "" {
		gosrc.SetLocalDevMode(*local)
	}
	dir, err := gosrc.Get(http.DefaultClient, path, *etag)
	if e, ok := err.(gosrc.NotFoundError); ok && e.Redirect != "" {
		log.Fatalf("redirect to %s", e.Redirect)
	} else if err != nil {
		log.Fatalf("%+v", err)
	}

	fmt.Println("ImportPath    ", dir.ImportPath)
	fmt.Println("ResovledPath  ", dir.ResolvedPath)
	fmt.Println("ProjectRoot   ", dir.ProjectRoot)
	fmt.Println("ProjectName   ", dir.ProjectName)
	fmt.Println("ProjectURL    ", dir.ProjectURL)
	fmt.Println("VCS           ", dir.VCS)
	fmt.Println("Etag          ", dir.Etag)
	fmt.Println("BrowseURL     ", dir.BrowseURL)
	fmt.Println("Subdirectories", strings.Join(dir.Subdirectories, ", "))
	fmt.Println("LineFmt       ", dir.LineFmt)
	fmt.Println("Files:")
	for _, file := range dir.Files {
		fmt.Printf("%30s %5d %s\n", file.Name, len(file.Data), file.BrowseURL)
	}
}

func printPresentation(path string) {
	pres, err := gosrc.GetPresentation(http.DefaultClient, path)
	if err != nil {
		log.Fatal(err)
	}
	fmt.Printf("%s\n", pres.Files[pres.Filename])
	for name, data := range pres.Files {
		if name != pres.Filename {
			fmt.Printf("---------- %s ----------\n%s\n", name, data)
		}
	}
}
