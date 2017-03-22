// Copyright 2013 The Go Authors. All rights reserved.
//
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file or at
// https://developers.google.com/open-source/licenses/bsd.

package database

import (
	"math"
	"reflect"
	"strconv"
	"testing"
	"time"

	"github.com/garyburd/redigo/redis"
	"golang.org/x/net/context"
	"google.golang.org/appengine/aetest"

	"github.com/golang/gddo/doc"
)

func newDB(t *testing.T) *Database {
	p := redis.NewPool(func() (redis.Conn, error) {
		c, err := redis.DialTimeout("tcp", ":6379", 0, 1*time.Second, 1*time.Second)
		if err != nil {
			return nil, err
		}
		_, err = c.Do("SELECT", "9")
		if err != nil {
			c.Close()
			return nil, err
		}
		return c, nil
	}, 1)

	c := p.Get()
	defer c.Close()
	n, err := redis.Int(c.Do("DBSIZE"))
	if n != 0 || err != nil {
		t.Errorf("DBSIZE returned %d, %v", n, err)
	}
	return &Database{Pool: p}
}

func closeDB(db *Database) {
	c := db.Pool.Get()
	c.Do("FLUSHDB")
	c.Close()
}

func TestPutGet(t *testing.T) {
	var nextCrawl = time.Unix(time.Now().Add(time.Hour).Unix(), 0).UTC()
	ctx, done, err := aetest.NewContext()
	if err != nil {
		t.Fatal(err)
	}
	defer done()
	bgCtx = func() context.Context {
		return ctx
	}

	db := newDB(t)
	defer closeDB(db)
	pdoc := &doc.Package{
		ImportPath:  "github.com/user/repo/foo/bar",
		Name:        "bar",
		Synopsis:    "hello",
		ProjectRoot: "github.com/user/repo",
		ProjectName: "foo",
		Updated:     time.Now().Add(-time.Hour),
		Imports:     []string{"C", "errors", "github.com/user/repo/foo/bar"}, // self import for testing convenience.
	}
	if err := db.Put(pdoc, nextCrawl, false); err != nil {
		t.Errorf("db.Put() returned error %v", err)
	}
	if err := db.Put(pdoc, time.Time{}, false); err != nil {
		t.Errorf("second db.Put() returned error %v", err)
	}

	actualPdoc, actualSubdirs, actualCrawl, err := db.Get("github.com/user/repo/foo/bar")
	if err != nil {
		t.Fatalf("db.Get(.../foo/bar) returned %v", err)
	}
	if len(actualSubdirs) != 0 {
		t.Errorf("db.Get(.../foo/bar) returned subdirs %v, want none", actualSubdirs)
	}
	if !reflect.DeepEqual(actualPdoc, pdoc) {
		t.Errorf("db.Get(.../foo/bar) returned doc %v, want %v", actualPdoc, pdoc)
	}
	if !nextCrawl.Equal(actualCrawl) {
		t.Errorf("db.Get(.../foo/bar) returned crawl %v, want %v", actualCrawl, nextCrawl)
	}

	before := time.Now().Unix()
	if err := db.BumpCrawl(pdoc.ProjectRoot); err != nil {
		t.Errorf("db.BumpCrawl() returned %v", err)
	}
	after := time.Now().Unix()

	_, _, actualCrawl, _ = db.Get("github.com/user/repo/foo/bar")
	if actualCrawl.Unix() < before || after < actualCrawl.Unix() {
		t.Errorf("actualCrawl=%v, expect value between %v and %v", actualCrawl.Unix(), before, after)
	}

	// Popular

	if err := db.IncrementPopularScore(pdoc.ImportPath); err != nil {
		t.Errorf("db.IncrementPopularScore() returned %v", err)
	}

	// Get "-"

	actualPdoc, _, _, err = db.Get("-")
	if err != nil {
		t.Fatalf("db.Get(-) returned %v", err)
	}
	if !reflect.DeepEqual(actualPdoc, pdoc) {
		t.Errorf("db.Get(-) returned doc %v, want %v", actualPdoc, pdoc)
	}

	actualPdoc, actualSubdirs, _, err = db.Get("github.com/user/repo/foo")
	if err != nil {
		t.Fatalf("db.Get(.../foo) returned %v", err)
	}
	if actualPdoc != nil {
		t.Errorf("db.Get(.../foo) returned doc %v, want %v", actualPdoc, nil)
	}
	expectedSubdirs := []Package{{Path: "github.com/user/repo/foo/bar", Synopsis: "hello"}}
	if !reflect.DeepEqual(actualSubdirs, expectedSubdirs) {
		t.Errorf("db.Get(.../foo) returned subdirs %v, want %v", actualSubdirs, expectedSubdirs)
	}
	actualImporters, err := db.Importers("github.com/user/repo/foo/bar")
	if err != nil {
		t.Fatalf("db.Importers() returned error %v", err)
	}
	expectedImporters := []Package{{Path: "github.com/user/repo/foo/bar", Synopsis: "hello"}}
	if !reflect.DeepEqual(actualImporters, expectedImporters) {
		t.Errorf("db.Importers() = %v, want %v", actualImporters, expectedImporters)
	}
	actualImports, err := db.Packages(pdoc.Imports)
	if err != nil {
		t.Fatalf("db.Imports() returned error %v", err)
	}
	for i := range actualImports {
		if actualImports[i].Path == "C" {
			actualImports[i].Synopsis = ""
		}
	}
	expectedImports := []Package{
		{Path: "C", Synopsis: ""},
		{Path: "errors", Synopsis: ""},
		{Path: "github.com/user/repo/foo/bar", Synopsis: "hello"},
	}
	if !reflect.DeepEqual(actualImports, expectedImports) {
		t.Errorf("db.Imports() = %v, want %v", actualImports, expectedImports)
	}
	importerCount, _ := db.ImporterCount("github.com/user/repo/foo/bar")
	if importerCount != 1 {
		t.Errorf("db.ImporterCount() = %d, want %d", importerCount, 1)
	}
	if err := db.Delete("github.com/user/repo/foo/bar"); err != nil {
		t.Errorf("db.Delete() returned error %v", err)
	}

	db.Query("bar")

	if err := db.Put(pdoc, time.Time{}, false); err != nil {
		t.Errorf("db.Put() returned error %v", err)
	}

	if err := db.Block("github.com/user/repo"); err != nil {
		t.Errorf("db.Block() returned error %v", err)
	}

	blocked, err := db.IsBlocked("github.com/user/repo/foo/bar")
	if !blocked || err != nil {
		t.Errorf("db.IsBlocked(github.com/user/repo/foo/bar) returned %v, %v, want true, nil", blocked, err)
	}

	blocked, err = db.IsBlocked("github.com/foo/bar")
	if blocked || err != nil {
		t.Errorf("db.IsBlocked(github.com/foo/bar) returned %v, %v, want false, nil", blocked, err)
	}

	c := db.Pool.Get()
	defer c.Close()
	c.Send("DEL", "maxQueryId")
	c.Send("DEL", "maxPackageId")
	c.Send("DEL", "block")
	c.Send("DEL", "popular:0")
	c.Send("DEL", "newCrawl")
	keys, err := redis.Values(c.Do("HKEYS", "ids"))
	for _, key := range keys {
		t.Errorf("unexpected id %s", key)
	}
	keys, err = redis.Values(c.Do("KEYS", "*"))
	for _, key := range keys {
		t.Errorf("unexpected key %s", key)
	}
}

const epsilon = 0.000001

func TestPopular(t *testing.T) {
	db := newDB(t)
	defer closeDB(db)
	c := db.Pool.Get()
	defer c.Close()

	// Add scores for packages. On each iteration, add half-life to time and
	// divide the score by two. All packages should have the same score.

	now := time.Now()
	score := float64(4048)
	for id := 12; id >= 0; id-- {
		path := "github.com/user/repo/p" + strconv.Itoa(id)
		c.Do("HSET", "ids", path, id)
		err := db.incrementPopularScoreInternal(path, score, now)
		if err != nil {
			t.Fatal(err)
		}
		now = now.Add(popularHalfLife)
		score /= 2
	}

	values, _ := redis.Values(c.Do("ZRANGE", "popular", "0", "100000", "WITHSCORES"))
	if len(values) != 26 {
		t.Fatalf("Expected 26 values, got %d", len(values))
	}

	// Check for equal scores.
	score, err := redis.Float64(values[1], nil)
	if err != nil {
		t.Fatal(err)
	}
	for i := 3; i < len(values); i += 2 {
		s, _ := redis.Float64(values[i], nil)
		if math.Abs(score-s)/score > epsilon {
			t.Errorf("Bad score, score[1]=%g, score[%d]=%g", score, i, s)
		}
	}
}

func TestCounter(t *testing.T) {
	db := newDB(t)
	defer closeDB(db)

	const key = "127.0.0.1"

	now := time.Now()
	n, err := db.incrementCounterInternal(key, 1, now)
	if err != nil {
		t.Fatal(err)
	}
	if math.Abs(n-1.0) > epsilon {
		t.Errorf("1: got n=%g, want 1", n)
	}
	n, err = db.incrementCounterInternal(key, 1, now)
	if err != nil {
		t.Fatal(err)
	}
	if math.Abs(n-2.0)/2.0 > epsilon {
		t.Errorf("2: got n=%g, want 2", n)
	}
	now = now.Add(counterHalflife)
	n, err = db.incrementCounterInternal(key, 1, now)
	if err != nil {
		t.Fatal(err)
	}
	if math.Abs(n-2.0)/2.0 > epsilon {
		t.Errorf("3: got n=%g, want 2", n)
	}
}
