// Copyright 2013 The Go Authors. All rights reserved.
//
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file or at
// https://developers.google.com/open-source/licenses/bsd.

package database

import (
	"path"
	"regexp"
	"strings"
	"unicode"

	"github.com/golang/gddo/doc"
	"github.com/golang/gddo/gosrc"
)

func isStandardPackage(path string) bool {
	return strings.Index(path, ".") < 0
}

func isTermSep(r rune) bool {
	return unicode.IsSpace(r) ||
		r != '.' && unicode.IsPunct(r) ||
		unicode.IsSymbol(r)
}

func normalizeProjectRoot(projectRoot string) string {
	if projectRoot == "" {
		return "go"
	}
	return projectRoot
}

var synonyms = map[string]string{
	"redis":    "redisdb", // append db to avoid stemming to 'red'
	"rand":     "random",
	"postgres": "postgresql",
	"mongo":    "mongodb",
}

func term(s string) string {
	s = strings.ToLower(s)
	if x, ok := synonyms[s]; ok {
		s = x
	}

	// Trim the trailing period at the end of any sentence.
	return stem(strings.TrimSuffix(s, "."))
}

var httpPat = regexp.MustCompile(`https?://\S+`)

func collectSynopsisTerms(terms map[string]bool, synopsis string) {

	synopsis = httpPat.ReplaceAllLiteralString(synopsis, "")

	fields := strings.FieldsFunc(synopsis, isTermSep)
	for i := range fields {
		fields[i] = strings.ToLower(fields[i])
	}

	// Ignore boilerplate in the following common patterns:
	//  Package foo ...
	//  Command foo ...
	//  Package foo implements ... (and provides, contains)
	//  The foo package ...
	//  The foo package implements ...
	//  The foo command ...

	checkPackageVerb := false
	switch {
	case len(fields) >= 1 && fields[0] == "package":
		fields = fields[1:]
		checkPackageVerb = true
	case len(fields) >= 1 && fields[0] == "command":
		fields = fields[1:]
	case len(fields) >= 3 && fields[0] == "the" && fields[2] == "package":
		fields[2] = fields[1]
		fields = fields[2:]
		checkPackageVerb = true
	case len(fields) >= 3 && fields[0] == "the" && fields[2] == "command":
		fields[2] = fields[1]
		fields = fields[2:]
	}

	if checkPackageVerb && len(fields) >= 2 &&
		(fields[1] == "implements" || fields[1] == "provides" || fields[1] == "contains") {
		fields[1] = fields[0]
		fields = fields[1:]
	}

	for _, s := range fields {
		if !stopWord[s] {
			terms[term(s)] = true
		}
	}
}

func termSlice(terms map[string]bool) []string {
	result := make([]string, 0, len(terms))
	for term := range terms {
		result = append(result, term)
	}
	return result
}

func documentTerms(pdoc *doc.Package, score float64) []string {

	terms := make(map[string]bool)

	// Project root

	projectRoot := normalizeProjectRoot(pdoc.ProjectRoot)
	terms["project:"+projectRoot] = true

	if strings.HasPrefix(pdoc.ImportPath, "golang.org/x/") {
		terms["project:subrepo"] = true
	}

	// Imports

	for _, path := range pdoc.Imports {
		if gosrc.IsValidPath(path) {
			terms["import:"+path] = true
		}
	}

	if score > 0 {

		for _, term := range parseQuery(pdoc.ImportPath) {
			terms[term] = true
		}
		if !isStandardPackage(pdoc.ImportPath) {
			terms["all:"] = true
			for _, term := range parseQuery(pdoc.ProjectName) {
				terms[term] = true
			}
			for _, term := range parseQuery(pdoc.Name) {
				terms[term] = true
			}
		}

		// Synopsis

		collectSynopsisTerms(terms, pdoc.Synopsis)

	}

	return termSlice(terms)
}

// vendorPat matches the path of a vendored package.
var vendorPat = regexp.MustCompile(
	// match directories used by tools to vendor packages.
	`/(?:_?third_party|vendors|Godeps/_workspace/src)/` +
		// match a domain name.
		`[^./]+\.[^/]+`)

func documentScore(pdoc *doc.Package) float64 {
	if pdoc.Name == "" ||
		pdoc.Status != gosrc.Active ||
		len(pdoc.Errors) > 0 ||
		strings.HasSuffix(pdoc.ImportPath, ".go") ||
		strings.HasPrefix(pdoc.ImportPath, "gist.github.com/") ||
		strings.HasSuffix(pdoc.ImportPath, "/internal") ||
		strings.Contains(pdoc.ImportPath, "/internal/") ||
		vendorPat.MatchString(pdoc.ImportPath) {
		return 0
	}

	for _, p := range pdoc.Imports {
		if strings.HasSuffix(p, ".go") {
			return 0
		}
	}

	r := 1.0
	if pdoc.IsCmd {
		if pdoc.Doc == "" {
			// Do not include command in index if it does not have documentation.
			return 0
		}
		if !importsGoPackages(pdoc) {
			// Penalize commands that don't use the "go/*" packages.
			r *= 0.9
		}
	} else {
		if !pdoc.Truncated &&
			len(pdoc.Consts) == 0 &&
			len(pdoc.Vars) == 0 &&
			len(pdoc.Funcs) == 0 &&
			len(pdoc.Types) == 0 &&
			len(pdoc.Examples) == 0 {
			// Do not include package in index if it does not have exports.
			return 0
		}
		if pdoc.Doc == "" {
			// Penalty for no documentation.
			r *= 0.95
		}
		if path.Base(pdoc.ImportPath) != pdoc.Name {
			// Penalty for last element of path != package name.
			r *= 0.9
		}
		for i := 0; i < strings.Count(pdoc.ImportPath[len(pdoc.ProjectRoot):], "/"); i++ {
			// Penalty for deeply nested packages.
			r *= 0.99
		}
		if strings.Index(pdoc.ImportPath[len(pdoc.ProjectRoot):], "/src/") > 0 {
			r *= 0.95
		}
		for _, p := range pdoc.Imports {
			if vendorPat.MatchString(p) {
				// Penalize packages that import vendored packages.
				r *= 0.1
				break
			}
		}
	}
	return r
}

func parseQuery(q string) []string {
	var terms []string
	q = strings.ToLower(q)
	for _, s := range strings.FieldsFunc(q, isTermSep) {
		if !stopWord[s] {
			terms = append(terms, term(s))
		}
	}
	return terms
}

func importsGoPackages(pdoc *doc.Package) bool {
	for _, m := range pdoc.Imports {
		if strings.HasPrefix(m, "go/") {
			return true
		}
	}
	return false
}
