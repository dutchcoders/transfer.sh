// Copyright 2014 The Go Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

// +build ignore

package main

import (
	"bytes"
	"flag"
	"go/format"
	"io/ioutil"
	"log"
	"net/http"
	"os"
	"os/exec"
	"strings"
	"text/template"
)

const (
	goRepoPath = 1 << iota
	packagePath
)

var tmpl = template.Must(template.New("").Parse(`// Created by go generate; DO NOT EDIT
// Copyright 2014 The Go Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package gosrc

const (
    goRepoPath = {{.goRepoPath}}
    packagePath = {{.packagePath}}
)

var pathFlags = map[string]int{
{{range $k, $v := .pathFlags}}{{printf "%q" $k}}: {{$v}},
{{end}} }

var validTLDs = map[string]bool{
{{range  $v := .validTLDs}}{{printf "%q" $v}}: true,
{{end}} }
`))

var output = flag.String("output", "data.go", "file name to write")

func main() {
	log.SetFlags(0)
	log.SetPrefix("gen: ")
	flag.Parse()
	if flag.NArg() != 0 {
		log.Fatal("usage: decgen [--output filename]")
	}

	// Build map of standard repository path flags.

	cmd := exec.Command("go", "list", "std", "cmd")
	p, err := cmd.Output()
	if err != nil {
		log.Fatal(err)
	}
	pathFlags := map[string]int{
		"builtin": packagePath | goRepoPath,
		"C":       packagePath,
	}
	for _, path := range strings.Fields(string(p)) {
		pathFlags[path] |= packagePath | goRepoPath
		for {
			i := strings.LastIndex(path, "/")
			if i < 0 {
				break
			}
			path = path[:i]
			pathFlags[path] |= goRepoPath
		}
	}

	// Get list of valid TLDs.

	resp, err := http.Get("https://data.iana.org/TLD/tlds-alpha-by-domain.txt")
	if err != nil {
		log.Fatal(err)
	}
	defer resp.Body.Close()
	p, err = ioutil.ReadAll(resp.Body)
	if err != nil {
		log.Fatal(err)
	}

	var validTLDs []string
	for _, line := range strings.Split(string(p), "\n") {
		line = strings.TrimSpace(line)
		if len(line) == 0 || line[0] == '#' {
			continue
		}
		validTLDs = append(validTLDs, "."+strings.ToLower(line))
	}

	// Generate output.

	var buf bytes.Buffer
	err = tmpl.Execute(&buf, map[string]interface{}{
		"output":      *output,
		"goRepoPath":  goRepoPath,
		"packagePath": packagePath,
		"pathFlags":   pathFlags,
		"validTLDs":   validTLDs,
	})
	if err != nil {
		log.Fatal("template error:", err)
	}
	source, err := format.Source(buf.Bytes())
	if err != nil {
		log.Fatal("source format error:", err)
	}
	fd, err := os.Create(*output)
	_, err = fd.Write(source)
	if err != nil {
		log.Fatal(err)
	}
}
