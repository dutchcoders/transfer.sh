// Copyright 2013 The Go Authors. All rights reserved.
//
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file or at
// https://developers.google.com/open-source/licenses/bsd.

package main

import (
	"log"
	"os"
	"time"

	"github.com/golang/gddo/database"
	"github.com/golang/gddo/doc"
)

var reindexCommand = &command{
	name:  "reindex",
	run:   reindex,
	usage: "reindex",
}

func fix(pdoc *doc.Package) {
	/*
	   	for _, v := range pdoc.Consts {
	   	}
	   	for _, v := range pdoc.Vars {
	   	}
	   	for _, v := range pdoc.Funcs {
	   	}
	   	for _, t := range pdoc.Types {
	   		for _, v := range t.Consts {
	   		}
	   		for _, v := range t.Vars {
	   		}
	   		for _, v := range t.Funcs {
	   		}
	   		for _, v := range t.Methods {
	   		}
	   	}
	       for _, notes := range pdoc.Notes {
	           for _, v := range notes {
	           }
	       }
	*/
}

func reindex(c *command) {
	if len(c.flag.Args()) != 0 {
		c.printUsage()
		os.Exit(1)
	}
	db, err := database.New()
	if err != nil {
		log.Fatal(err)
	}
	var n int
	err = db.Do(func(pi *database.PackageInfo) error {
		n++
		fix(pi.PDoc)
		return db.Put(pi.PDoc, time.Time{}, false)
	})
	if err != nil {
		log.Fatal(err)
	}
	log.Printf("Updated %d documents", n)
}
