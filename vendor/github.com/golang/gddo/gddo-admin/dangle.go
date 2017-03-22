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
	"github.com/golang/gddo/gosrc"
)

var dangleCommand = &command{
	name:  "dangle",
	run:   dangle,
	usage: "dangle",
}

func dangle(c *command) {
	if len(c.flag.Args()) != 0 {
		c.printUsage()
		os.Exit(1)
	}
	db, err := database.New()
	if err != nil {
		log.Fatal(err)
	}
	m := make(map[string]int)
	err = db.Do(func(pi *database.PackageInfo) error {
		m[pi.PDoc.ImportPath] |= 1
		for _, p := range pi.PDoc.Imports {
			if gosrc.IsValidPath(p) {
				m[p] |= 2
			}
		}
		for _, p := range pi.PDoc.TestImports {
			if gosrc.IsValidPath(p) {
				m[p] |= 2
			}
		}
		for _, p := range pi.PDoc.XTestImports {
			if gosrc.IsValidPath(p) {
				m[p] |= 2
			}
		}
		return nil
	})

	for p, v := range m {
		if v == 2 {
			fmt.Println(p)
		}
	}
}
