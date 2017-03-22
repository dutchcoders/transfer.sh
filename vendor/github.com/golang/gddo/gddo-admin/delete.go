// Copyright 2013 The Go Authors. All rights reserved.
//
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file or at
// https://developers.google.com/open-source/licenses/bsd.

package main

import (
	"log"
	"os"

	"github.com/golang/gddo/database"
)

var deleteCommand = &command{
	name:  "delete",
	run:   del,
	usage: "delete path",
}

func del(c *command) {
	if len(c.flag.Args()) != 1 {
		c.printUsage()
		os.Exit(1)
	}
	db, err := database.New()
	if err != nil {
		log.Fatal(err)
	}
	if err := db.Delete(c.flag.Args()[0]); err != nil {
		log.Fatal(err)
	}
}
