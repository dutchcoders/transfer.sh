// Copyright 2013 The Go Authors. All rights reserved.
//
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file or at
// https://developers.google.com/open-source/licenses/bsd.

package main

import (
	"fmt"
	"log"
	"os"
	"sort"

	"github.com/golang/gddo/database"
)

var statsCommand = &command{
	name:  "stats",
	run:   stats,
	usage: "stats",
}

type itemSize struct {
	path string
	size int
}

type bySizeDesc []itemSize

func (p bySizeDesc) Len() int           { return len(p) }
func (p bySizeDesc) Less(i, j int) bool { return p[i].size > p[j].size }
func (p bySizeDesc) Swap(i, j int)      { p[i], p[j] = p[j], p[i] }

func stats(c *command) {
	if len(c.flag.Args()) != 0 {
		c.printUsage()
		os.Exit(1)
	}
	db, err := database.New()
	if err != nil {
		log.Fatal(err)
	}

	var packageSizes []itemSize
	var truncatedPackages []string
	projectSizes := make(map[string]int)
	err = db.Do(func(pi *database.PackageInfo) error {
		packageSizes = append(packageSizes, itemSize{pi.PDoc.ImportPath, pi.Size})
		projectSizes[pi.PDoc.ProjectRoot] += pi.Size
		if pi.PDoc.Truncated {
			truncatedPackages = append(truncatedPackages, pi.PDoc.ImportPath)
		}
		return nil
	})

	var sizes []itemSize
	for path, size := range projectSizes {
		sizes = append(sizes, itemSize{path, size})
	}
	sort.Sort(bySizeDesc(sizes))
	fmt.Println("PROJECT SIZES")
	for _, size := range sizes {
		fmt.Printf("%6d %s\n", size.size, size.path)
	}

	sort.Sort(bySizeDesc(packageSizes))
	fmt.Println("PACKAGE SIZES")
	for _, size := range packageSizes {
		fmt.Printf("%6d %s\n", size.size, size.path)
	}

	sort.Sort(sort.StringSlice(truncatedPackages))
	fmt.Println("TRUNCATED PACKAGES")
	for _, p := range truncatedPackages {
		fmt.Printf("%s\n", p)
	}
}
