// Copyright 2016 The Go Authors. All rights reserved.
//
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file or at
// https://developers.google.com/open-source/licenses/bsd.

package database

import (
	"math"
	"strconv"
	"testing"

	"google.golang.org/appengine/aetest"
	"google.golang.org/appengine/search"

	"github.com/golang/gddo/doc"
)

var pdoc = &doc.Package{
	ImportPath: "github.com/golang/test",
	Name:       "test",
	Synopsis:   "This is a test package.",
	Fork:       true,
	Stars:      10,
}

func TestPutIndexWithEmptyId(t *testing.T) {
	c, done, err := aetest.NewContext()
	if err != nil {
		t.Fatal(err)
	}
	defer done()

	if err := PutIndex(c, nil, "", 0, 0); err == nil {
		t.Errorf("PutIndex succeeded unexpectedly")
	}
}

func TestPutIndexCreateNilDoc(t *testing.T) {
	c, done, err := aetest.NewContext()
	if err != nil {
		t.Fatal(err)
	}
	defer done()

	if err := PutIndex(c, nil, "12345", -1, 2); err == nil {
		t.Errorf("PutIndex succeeded unexpectedly")
	}
}

func TestPutIndexNewPackageAndUpdate(t *testing.T) {
	c, done, err := aetest.NewContext()
	if err != nil {
		t.Fatal(err)
	}
	defer done()

	// Put a new package into search index.
	if err := PutIndex(c, pdoc, "12345", 0.99, 1); err != nil {
		t.Fatal(err)
	}

	// Verify the package was put in is as expected.
	idx, err := search.Open("packages")
	if err != nil {
		t.Fatal(err)
	}
	var got Package
	if err = idx.Get(c, "12345", &got); err != nil && err != search.ErrNoSuchDocument {
		t.Fatal(err)
	}
	wanted := Package{
		Name:        pdoc.Name,
		Path:        pdoc.ImportPath,
		Synopsis:    pdoc.Synopsis,
		ImportCount: 1,
		Fork:        true,
		Stars:       10,
		Score:       0.99,
	}
	if got != wanted {
		t.Errorf("PutIndex got %v, want %v", got, wanted)
	}

	// Update the import count of the package.
	if err := PutIndex(c, nil, "12345", -1, 2); err != nil {
		t.Fatal(err)
	}
	if err := idx.Get(c, "12345", &got); err != nil && err != search.ErrNoSuchDocument {
		t.Fatal(err)
	}
	wanted.ImportCount = 2
	if got != wanted {
		t.Errorf("PutIndex got %v, want %v", got, wanted)
	}
}

func TestSearchResultSorted(t *testing.T) {
	c, done, err := aetest.NewContext()
	if err != nil {
		t.Fatal(err)
	}
	defer done()

	// Put multiple packages into the search index and the search result
	// should be sorted properly.
	id := "1"
	for i := 2; i < 6; i++ {
		id += strconv.Itoa(i)
		pdoc.Synopsis = id
		if err := PutIndex(c, pdoc, id, math.Pow(0.9, float64(i)), 10*i); err != nil {
			t.Fatal(err)
		}
	}
	got, err := Search(c, "test")
	if err != nil {
		t.Fatal(err)
	}
	wanted := []string{"123", "12", "1234", "12345"}
	for i, p := range got {
		if p.Synopsis != wanted[i] {
			t.Errorf("Search got %v, want %v", p.Synopsis, wanted[i])
		}
	}
}
