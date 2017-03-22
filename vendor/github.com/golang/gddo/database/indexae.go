// Copyright 2016 The Go Authors. All rights reserved.
//
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file or at
// https://developers.google.com/open-source/licenses/bsd.

package database

import (
	"bytes"
	"errors"
	"fmt"
	"log"
	"math"
	"strings"
	"unicode"

	"golang.org/x/net/context"
	"google.golang.org/appengine/search"

	"github.com/golang/gddo/doc"
)

func (p *Package) Load(fields []search.Field, meta *search.DocumentMetadata) error {
	for _, f := range fields {
		switch f.Name {
		case "Name":
			if v, ok := f.Value.(search.Atom); ok {
				p.Name = string(v)
			}
		case "Path":
			if v, ok := f.Value.(string); ok {
				p.Path = v
			}
		case "Synopsis":
			if v, ok := f.Value.(string); ok {
				p.Synopsis = v
			}
		case "ImportCount":
			if v, ok := f.Value.(float64); ok {
				p.ImportCount = int(v)
			}
		case "Stars":
			if v, ok := f.Value.(float64); ok {
				p.Stars = int(v)
			}
		case "Score":
			if v, ok := f.Value.(float64); ok {
				p.Score = v
			}
		}
	}
	if p.Path == "" {
		return errors.New("Invalid document: missing Path field")
	}
	for _, f := range meta.Facets {
		if f.Name == "Fork" {
			p.Fork = f.Value.(search.Atom) == "true"
		}
	}
	return nil
}

func (p *Package) Save() ([]search.Field, *search.DocumentMetadata, error) {
	fields := []search.Field{
		{Name: "Name", Value: search.Atom(p.Name)},
		{Name: "Path", Value: p.Path},
		{Name: "Synopsis", Value: p.Synopsis},
		{Name: "Score", Value: p.Score},
		{Name: "ImportCount", Value: float64(p.ImportCount)},
		{Name: "Stars", Value: float64(p.Stars)},
	}
	fork := fmt.Sprint(p.Fork) // "true" or "false"
	meta := &search.DocumentMetadata{
		// Customize the rank property by the product of the package score and
		// natural logarithm of the import count. Rank must be a positive integer.
		// Use 1 as minimum rank and keep 3 digits of precision to distinguish
		// close ranks.
		Rank: int(math.Max(1, 1000*p.Score*math.Log(math.E+float64(p.ImportCount)))),
		Facets: []search.Facet{
			{Name: "Fork", Value: search.Atom(fork)},
		},
	}
	return fields, meta, nil
}

// PutIndex creates or updates a package entry in the search index. id identifies the document in the index.
// If pdoc is non-nil, PutIndex will update the package's name, path and synopsis supplied by pdoc.
// pdoc must be non-nil for a package's first call to PutIndex.
// PutIndex updates the Score to score, if non-negative.
func PutIndex(c context.Context, pdoc *doc.Package, id string, score float64, importCount int) error {
	if id == "" {
		return errors.New("indexae: no id assigned")
	}
	idx, err := search.Open("packages")
	if err != nil {
		return err
	}

	var pkg Package
	if err := idx.Get(c, id, &pkg); err != nil {
		if err != search.ErrNoSuchDocument {
			return err
		} else if pdoc == nil {
			// Cannot update a non-existing document.
			return errors.New("indexae: cannot create new document with nil pdoc")
		}
		// No such document in the index, fall through.
	}

	// Update document information accordingly.
	if pdoc != nil {
		pkg.Name = pdoc.Name
		pkg.Path = pdoc.ImportPath
		pkg.Synopsis = pdoc.Synopsis
		pkg.Stars = pdoc.Stars
		pkg.Fork = pdoc.Fork
	}
	if score >= 0 {
		pkg.Score = score
	}
	pkg.ImportCount = importCount

	if _, err := idx.Put(c, id, &pkg); err != nil {
		return err
	}
	return nil
}

// Search searches the packages index for a given query. A path-like query string
// will be passed in unchanged, whereas single words will be stemmed.
func Search(c context.Context, q string) ([]Package, error) {
	index, err := search.Open("packages")
	if err != nil {
		return nil, err
	}
	var pkgs []Package
	opt := &search.SearchOptions{
		Limit: 100,
	}
	for it := index.Search(c, parseQuery2(q), opt); ; {
		var p Package
		_, err := it.Next(&p)
		if err == search.Done {
			break
		}
		if err != nil {
			return nil, err
		}
		pkgs = append(pkgs, p)
	}
	return pkgs, nil
}

func parseQuery2(q string) string {
	var buf bytes.Buffer
	for _, s := range strings.FieldsFunc(q, isTermSep2) {
		if strings.ContainsAny(s, "./") {
			// Quote terms with / or . for path like query.
			fmt.Fprintf(&buf, "%q ", s)
		} else {
			// Stem for single word terms.
			fmt.Fprintf(&buf, "~%v ", s)
		}
	}
	return buf.String()
}

func isTermSep2(r rune) bool {
	return unicode.IsSpace(r) ||
		r != '.' && r != '/' && unicode.IsPunct(r) ||
		unicode.IsSymbol(r)
}

func deleteIndex(c context.Context, id string) error {
	idx, err := search.Open("packages")
	if err != nil {
		return err
	}
	return idx.Delete(c, id)
}

// PurgeIndex deletes all the packages from the search index.
func PurgeIndex(c context.Context) error {
	idx, err := search.Open("packages")
	if err != nil {
		return err
	}
	n := 0

	for it := idx.List(c, &search.ListOptions{IDsOnly: true}); ; n++ {
		var pkg Package
		id, err := it.Next(&pkg)
		if err == search.Done {
			break
		}
		if err != nil {
			return err
		}
		if err := idx.Delete(c, id); err != nil {
			log.Printf("Failed to delete package %s: %v", id, err)
			continue
		}
	}
	log.Printf("Purged %d packages from the search index.", n)
	return nil
}
