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

	"github.com/golang/gddo/database"
)

var (
	popularCommand = &command{
		name:  "popular",
		usage: "popular",
	}
)

func init() {
	popularCommand.run = popular
}

func popular(c *command) {
	if len(c.flag.Args()) != 0 {
		c.printUsage()
		os.Exit(1)
	}
	db, err := database.New()
	if err != nil {
		log.Fatal(err)
	}
	pkgs, err := db.PopularWithScores()
	if err != nil {
		log.Fatal(err)
	}
	for _, pkg := range pkgs {
		fmt.Println(pkg.Path, pkg.Synopsis)
	}
}
