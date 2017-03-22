// Copyright 2013 The Go Authors. All rights reserved.
//
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file or at
// https://developers.google.com/open-source/licenses/bsd.

package main

import (
	"fmt"
	"io/ioutil"
	"log"
	"os"
	"strings"

	"github.com/garyburd/redigo/redis"
	"github.com/golang/gddo/database"
)

var crawlCommand = &command{
	name:  "crawl",
	run:   crawl,
	usage: "crawl [new]",
}

func crawl(c *command) {
	if len(c.flag.Args()) > 1 {
		c.printUsage()
		os.Exit(1)
	}
	db, err := database.New()
	if err != nil {
		log.Fatal(err)
	}

	if len(c.flag.Args()) == 1 {
		p, err := ioutil.ReadFile(c.flag.Args()[0])
		if err != nil {
			log.Fatal(err)
		}
		for _, p := range strings.Fields(string(p)) {
			db.AddNewCrawl(p)
		}
	}

	conn := db.Pool.Get()
	defer conn.Close()
	paths, err := redis.Strings(conn.Do("SMEMBERS", "newCrawl"))
	if err != nil {
		log.Fatal(err)
	}
	fmt.Println("NEW")
	for _, path := range paths {
		fmt.Println(path)
	}

	paths, err = redis.Strings(conn.Do("SMEMBERS", "badCrawl"))
	if err != nil {
		log.Fatal(err)
	}
	fmt.Println("BAD")
	for _, path := range paths {
		fmt.Println(path)
	}
}
