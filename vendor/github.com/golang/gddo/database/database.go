// Copyright 2013 The Go Authors. All rights reserved.
//
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file or at
// https://developers.google.com/open-source/licenses/bsd.

// Redis keys and types:
//
// maxPackageId string: next id to assign
// ids hash maps import path to package id
// pkg:<id> hash
//      terms: space separated search terms
//      path: import path
//      synopsis: synopsis
//      gob: snappy compressed gob encoded doc.Package
//      score: document search score
//      etag:
//      kind: p=package, c=command, d=directory with no go files
// index:<term> set: package ids for given search term
// index:import:<path> set: packages with import path
// index:project:<root> set: packages in project with root
// block set: packages to block
// popular zset: package id, score
// popular:0 string: scaled base time for popular scores
// nextCrawl zset: package id, Unix time for next crawl
// newCrawl set: new paths to crawl
// badCrawl set: paths that returned error when crawling.

// Package database manages storage for GoPkgDoc.
package database

import (
	"bytes"
	"encoding/gob"
	"errors"
	"fmt"
	"log"
	"math"
	"net/url"
	"os"
	"path"
	"sort"
	"strconv"
	"strings"
	"time"
	"unicode/utf8"

	"github.com/garyburd/redigo/redis"
	"github.com/golang/snappy"
	"golang.org/x/net/context"
	"google.golang.org/appengine"
	"google.golang.org/appengine/search"

	"github.com/golang/gddo/doc"
	"github.com/golang/gddo/gosrc"
)

type Database struct {
	Pool interface {
		Get() redis.Conn
	}
}

// Package represents the content of a package both for the search index and
// for the HTML template. It implements the search.FieldLoadSaver interface
// to customize the Rank function in the search index.
type Package struct {
	Name        string  `json:"name,omitempty"`
	Path        string  `json:"path"`
	ImportCount int     `json:"import_count"`
	Synopsis    string  `json:"synopsis,omitempty"`
	Fork        bool    `json:"fork,omitempty"`
	Stars       int     `json:"stars,omitempty"`
	Score       float64 `json:"score,omitempty"`
}

type byPath []Package

func (p byPath) Len() int           { return len(p) }
func (p byPath) Less(i, j int) bool { return p[i].Path < p[j].Path }
func (p byPath) Swap(i, j int)      { p[i], p[j] = p[j], p[i] }

// Configuration variables (and default values of flags.
var (
	RedisServer      = "redis://127.0.0.1:6379" // URL of Redis server
	RedisIdleTimeout = 250 * time.Second        // Close Redis connections after remaining idle for this duration.
	RedisLog         = false                    // Log database commands
)

func dialDb() (c redis.Conn, err error) {
	u, err := url.Parse(RedisServer)
	if err != nil {
		return nil, err
	}

	defer func() {
		if err != nil && c != nil {
			c.Close()
		}
	}()

	c, err = redis.Dial("tcp", u.Host)
	if err != nil {
		return
	}

	if RedisLog {
		l := log.New(os.Stderr, "", log.LstdFlags)
		c = redis.NewLoggingConn(c, l, "")
	}

	if u.User != nil {
		if pw, ok := u.User.Password(); ok {
			if _, err = c.Do("AUTH", pw); err != nil {
				return
			}
		}
	}
	return
}

// New creates a database configured from command line flags.
func New() (*Database, error) {
	pool := &redis.Pool{
		Dial:        dialDb,
		MaxIdle:     10,
		IdleTimeout: RedisIdleTimeout,
	}

	c := pool.Get()
	if c.Err() != nil {
		return nil, c.Err()
	}
	c.Close()

	return &Database{Pool: pool}, nil
}

// Exists returns true if package with import path exists in the database.
func (db *Database) Exists(path string) (bool, error) {
	c := db.Pool.Get()
	defer c.Close()
	return redis.Bool(c.Do("HEXISTS", "ids", path))
}

var putScript = redis.NewScript(0, `
    local path = ARGV[1]
    local synopsis = ARGV[2]
    local score = ARGV[3]
    local gob = ARGV[4]
    local terms = ARGV[5]
    local etag = ARGV[6]
    local kind = ARGV[7]
    local nextCrawl = ARGV[8]

    local id = redis.call('HGET', 'ids', path)
    if not id then
        id = redis.call('INCR', 'maxPackageId')
        redis.call('HSET', 'ids', path, id)
    end

    if etag ~= '' and etag == redis.call('HGET', 'pkg:' .. id, 'clone') then
        terms = ''
        score = 0
    end

    local update = {}
    for term in string.gmatch(redis.call('HGET', 'pkg:' .. id, 'terms') or '', '([^ ]+)') do
        update[term] = 1
    end

    for term in string.gmatch(terms, '([^ ]+)') do
        update[term] = (update[term] or 0) + 2
    end

    for term, x in pairs(update) do
        if x == 1 then
            redis.call('SREM', 'index:' .. term, id)
        elseif x == 2 then 
            redis.call('SADD', 'index:' .. term, id)
        end
    end

    redis.call('SREM', 'badCrawl', path)
    redis.call('SREM', 'newCrawl', path)

    if nextCrawl ~= '0' then
        redis.call('ZADD', 'nextCrawl', nextCrawl, id)
        redis.call('HSET', 'pkg:' .. id, 'crawl', nextCrawl)
    end

    return redis.call('HMSET', 'pkg:' .. id, 'path', path, 'synopsis', synopsis, 'score', score, 'gob', gob, 'terms', terms, 'etag', etag, 'kind', kind)
`)

var addCrawlScript = redis.NewScript(0, `
    for i=1,#ARGV do
        local pkg = ARGV[i]
        if redis.call('HEXISTS', 'ids',  pkg) == 0  and redis.call('SISMEMBER', 'badCrawl', pkg) == 0 then
            redis.call('SADD', 'newCrawl', pkg)
        end
    end
`)

func (db *Database) AddNewCrawl(importPath string) error {
	if !gosrc.IsValidRemotePath(importPath) {
		return errors.New("bad path")
	}
	c := db.Pool.Get()
	defer c.Close()
	_, err := addCrawlScript.Do(c, importPath)
	return err
}

var bgCtx = appengine.BackgroundContext // replaced by tests

// Put adds the package documentation to the database.
func (db *Database) Put(pdoc *doc.Package, nextCrawl time.Time, hide bool) error {
	c := db.Pool.Get()
	defer c.Close()

	score := 0.0
	if !hide {
		score = documentScore(pdoc)
	}
	terms := documentTerms(pdoc, score)

	var gobBuf bytes.Buffer
	if err := gob.NewEncoder(&gobBuf).Encode(pdoc); err != nil {
		return err
	}

	gobBytes := snappy.Encode(nil, gobBuf.Bytes())

	// Truncate large documents.
	if len(gobBytes) > 400000 {
		pdocNew := *pdoc
		pdoc = &pdocNew
		pdoc.Truncated = true
		pdoc.Vars = nil
		pdoc.Funcs = nil
		pdoc.Types = nil
		pdoc.Consts = nil
		pdoc.Examples = nil
		gobBuf.Reset()
		if err := gob.NewEncoder(&gobBuf).Encode(pdoc); err != nil {
			return err
		}
		gobBytes = snappy.Encode(nil, gobBuf.Bytes())
	}

	kind := "p"
	switch {
	case pdoc.Name == "":
		kind = "d"
	case pdoc.IsCmd:
		kind = "c"
	}

	t := int64(0)
	if !nextCrawl.IsZero() {
		t = nextCrawl.Unix()
	}

	// Get old version of the package to extract its imports.
	// If the package does not exist, both oldDoc and err will be nil.
	old, _, err := db.getDoc(c, pdoc.ImportPath)
	if err != nil {
		return err
	}

	_, err = putScript.Do(c, pdoc.ImportPath, pdoc.Synopsis, score, gobBytes, strings.Join(terms, " "), pdoc.Etag, kind, t)
	if err != nil {
		return err
	}

	id, n, err := pkgIDAndImportCount(c, pdoc.ImportPath)
	if err != nil {
		return err
	}
	ctx := bgCtx()

	if score > 0 {
		if err := PutIndex(ctx, pdoc, id, score, n); err != nil {
			log.Printf("Cannot put %q in index: %v", pdoc.ImportPath, err)
		}

		if old != nil {
			if err := updateImportsIndex(c, ctx, old, pdoc); err != nil {
				return err
			}
		}
	} else {
		if err := deleteIndex(ctx, id); err != nil {
			return err
		}
	}

	if nextCrawl.IsZero() {
		// Skip crawling related packages if this is not a full save.
		return nil
	}

	paths := make(map[string]bool)
	for _, p := range pdoc.Imports {
		if gosrc.IsValidRemotePath(p) {
			paths[p] = true
		}
	}
	for _, p := range pdoc.TestImports {
		if gosrc.IsValidRemotePath(p) {
			paths[p] = true
		}
	}
	for _, p := range pdoc.XTestImports {
		if gosrc.IsValidRemotePath(p) {
			paths[p] = true
		}
	}
	if pdoc.ImportPath != pdoc.ProjectRoot && pdoc.ProjectRoot != "" {
		paths[pdoc.ProjectRoot] = true
	}
	for _, p := range pdoc.Subdirectories {
		paths[pdoc.ImportPath+"/"+p] = true
	}

	args := make([]interface{}, 0, len(paths))
	for p := range paths {
		args = append(args, p)
	}
	_, err = addCrawlScript.Do(c, args...)
	return err
}

// pkgIDAndImportCount returns the ID and import count of a specified package.
func pkgIDAndImportCount(c redis.Conn, path string) (id string, numImported int, err error) {
	numImported, err = redis.Int(c.Do("SCARD", "index:import:"+path))
	if err != nil {
		return
	}
	id, err = redis.String(c.Do("HGET", "ids", path))
	if err == redis.ErrNil {
		return "", 0, nil
	}
	return id, numImported, nil
}

func updateImportsIndex(c redis.Conn, ctx context.Context, oldDoc, newDoc *doc.Package) error {
	// Create a map to store any import change since last time we indexed the package.
	changes := make(map[string]bool)
	for _, p := range oldDoc.Imports {
		if gosrc.IsValidRemotePath(p) {
			changes[p] = true
		}
	}
	for _, p := range newDoc.Imports {
		if gosrc.IsValidRemotePath(p) {
			delete(changes, p)
		}
	}

	// For each import change, re-index that package with updated NumImported.
	// In practice this should not happen often and when it does, the changes are
	// likely to be a small amount.
	for p, _ := range changes {
		id, n, err := pkgIDAndImportCount(c, p)
		if err != nil {
			return err
		}
		if id != "" {
			PutIndex(ctx, nil, id, -1, n)
		}
	}
	return nil
}

var setNextCrawlScript = redis.NewScript(0, `
    local path = ARGV[1]
    local nextCrawl = ARGV[2]

    local id = redis.call('HGET', 'ids', path)
    if not id then
        return false
    end

    redis.call('ZADD', 'nextCrawl', nextCrawl, id)
    redis.call('HSET', 'pkg:' .. id, 'crawl', nextCrawl)
`)

// SetNextCrawl sets the next crawl time for a package.
func (db *Database) SetNextCrawl(path string, t time.Time) error {
	c := db.Pool.Get()
	defer c.Close()
	_, err := setNextCrawlScript.Do(c, path, t.Unix())
	return err
}

// bumpCrawlScript sets the crawl time to now. To avoid continuously crawling
// frequently updated repositories, the crawl is scheduled in the future.
var bumpCrawlScript = redis.NewScript(0, `
    local root = ARGV[1]
    local now = tonumber(ARGV[2])
    local nextCrawl = now + 7200
    local pkgs = redis.call('SORT', 'index:project:' .. root, 'GET', '#')

    for i=1,#pkgs do
        local v = redis.call('HMGET', 'pkg:' .. pkgs[i], 'crawl', 'kind')
        local t = tonumber(v[1] or 0)
        if t == 0 or now < t then
            redis.call('HSET', 'pkg:' .. pkgs[i], 'crawl', now)
        end
        local nextCrawl = now + 86400
        if v[2] == 'p' then
            nextCrawl = now + 7200
        end
        t = tonumber(redis.call('ZSCORE', 'nextCrawl', pkgs[i]) or 0)
        if t == 0 or nextCrawl < t then
            redis.call('ZADD', 'nextCrawl', nextCrawl, pkgs[i])
        end
    end
`)

func (db *Database) BumpCrawl(projectRoot string) error {
	c := db.Pool.Get()
	defer c.Close()
	_, err := bumpCrawlScript.Do(c, normalizeProjectRoot(projectRoot), time.Now().Unix())
	return err
}

// getDocScript gets the package documentation and update time for the
// specified path. If path is "-", then the oldest document is returned.
var getDocScript = redis.NewScript(0, `
    local path = ARGV[1]

    local id
    if path == '-' then
        local r = redis.call('ZRANGE', 'nextCrawl', 0, 0)
        if not r or #r == 0 then
            return false
        end
        id = r[1]
    else
        id = redis.call('HGET', 'ids', path)
        if not id then
            return false
        end
    end

    local gob = redis.call('HGET', 'pkg:' .. id, 'gob')
    if not gob then
        return false
    end

    local nextCrawl = redis.call('HGET', 'pkg:' .. id, 'crawl')
    if not nextCrawl then 
        nextCrawl = redis.call('ZSCORE', 'nextCrawl', id)
        if not nextCrawl then
            nextCrawl = 0
        end
    end
    
    return {gob, nextCrawl}
`)

func (db *Database) getDoc(c redis.Conn, path string) (*doc.Package, time.Time, error) {
	r, err := redis.Values(getDocScript.Do(c, path))
	if err == redis.ErrNil {
		return nil, time.Time{}, nil
	} else if err != nil {
		return nil, time.Time{}, err
	}

	var p []byte
	var t int64

	if _, err := redis.Scan(r, &p, &t); err != nil {
		return nil, time.Time{}, err
	}

	p, err = snappy.Decode(nil, p)
	if err != nil {
		return nil, time.Time{}, err
	}

	var pdoc doc.Package
	if err := gob.NewDecoder(bytes.NewReader(p)).Decode(&pdoc); err != nil {
		return nil, time.Time{}, err
	}

	nextCrawl := pdoc.Updated
	if t != 0 {
		nextCrawl = time.Unix(t, 0).UTC()
	}

	return &pdoc, nextCrawl, err
}

var getSubdirsScript = redis.NewScript(0, `
    local reply
    for i = 1,#ARGV do
        reply = redis.call('SORT', 'index:project:' .. ARGV[i], 'ALPHA', 'BY', 'pkg:*->path', 'GET', 'pkg:*->path', 'GET', 'pkg:*->synopsis', 'GET', 'pkg:*->kind')
        if #reply > 0 then
            break
        end
    end
    return reply
`)

func (db *Database) getSubdirs(c redis.Conn, path string, pdoc *doc.Package) ([]Package, error) {
	var reply interface{}
	var err error

	switch {
	case isStandardPackage(path):
		reply, err = getSubdirsScript.Do(c, "go")
	case pdoc != nil:
		reply, err = getSubdirsScript.Do(c, pdoc.ProjectRoot)
	default:
		var roots []interface{}
		projectRoot := path
		for i := 0; i < 5; i++ {
			roots = append(roots, projectRoot)
			if j := strings.LastIndex(projectRoot, "/"); j < 0 {
				break
			} else {
				projectRoot = projectRoot[:j]
			}
		}
		reply, err = getSubdirsScript.Do(c, roots...)
	}

	values, err := redis.Values(reply, err)
	if err != nil {
		return nil, err
	}

	var subdirs []Package
	prefix := path + "/"

	for len(values) > 0 {
		var pkg Package
		var kind string
		values, err = redis.Scan(values, &pkg.Path, &pkg.Synopsis, &kind)
		if err != nil {
			return nil, err
		}
		if (kind == "p" || kind == "c") && strings.HasPrefix(pkg.Path, prefix) {
			subdirs = append(subdirs, pkg)
		}
	}

	return subdirs, err
}

// Get gets the package documentation and sub-directories for the the given
// import path.
func (db *Database) Get(path string) (*doc.Package, []Package, time.Time, error) {
	c := db.Pool.Get()
	defer c.Close()

	pdoc, nextCrawl, err := db.getDoc(c, path)
	if err != nil {
		return nil, nil, time.Time{}, err
	}

	if pdoc != nil {
		// fixup for speclal "-" path.
		path = pdoc.ImportPath
	}

	subdirs, err := db.getSubdirs(c, path, pdoc)
	if err != nil {
		return nil, nil, time.Time{}, err
	}
	return pdoc, subdirs, nextCrawl, nil
}

func (db *Database) GetDoc(path string) (*doc.Package, time.Time, error) {
	c := db.Pool.Get()
	defer c.Close()
	return db.getDoc(c, path)
}

var deleteScript = redis.NewScript(0, `
    local path = ARGV[1]

    local id = redis.call('HGET', 'ids', path)
    if not id then
        return false
    end

    for term in string.gmatch(redis.call('HGET', 'pkg:' .. id, 'terms') or '', '([^ ]+)') do
        redis.call('SREM', 'index:' .. term, id)
    end

    redis.call('ZREM', 'nextCrawl', id)
    redis.call('SREM', 'newCrawl', path)
    redis.call('ZREM', 'popular', id)
    redis.call('DEL', 'pkg:' .. id)
    return redis.call('HDEL', 'ids', path)
`)

// Delete deletes the documentation for the given import path.
func (db *Database) Delete(path string) error {
	c := db.Pool.Get()
	defer c.Close()

	ctx := bgCtx()
	id, err := redis.String(c.Do("HGET", "ids", path))
	if err == redis.ErrNil {
		return nil
	}
	if err != nil {
		return err
	}
	if err := deleteIndex(ctx, id); err != nil {
		return err
	}

	_, err = deleteScript.Do(c, path)
	return err
}

func packages(reply interface{}, all bool) ([]Package, error) {
	values, err := redis.Values(reply, nil)
	if err != nil {
		return nil, err
	}
	result := make([]Package, 0, len(values)/3)
	for len(values) > 0 {
		var pkg Package
		var kind string
		values, err = redis.Scan(values, &pkg.Path, &pkg.Synopsis, &kind)
		if err != nil {
			return nil, err
		}
		if !all && kind == "d" {
			continue
		}
		if pkg.Path == "C" {
			pkg.Synopsis = "Package C is a \"pseudo-package\" used to access the C namespace from a cgo source file."
		}
		result = append(result, pkg)
	}
	return result, nil
}

func (db *Database) getPackages(key string, all bool) ([]Package, error) {
	c := db.Pool.Get()
	defer c.Close()
	reply, err := c.Do("SORT", key, "ALPHA", "BY", "pkg:*->path", "GET", "pkg:*->path", "GET", "pkg:*->synopsis", "GET", "pkg:*->kind")
	if err != nil {
		return nil, err
	}
	return packages(reply, all)
}

func (db *Database) GoIndex() ([]Package, error) {
	return db.getPackages("index:project:go", false)
}

func (db *Database) GoSubrepoIndex() ([]Package, error) {
	return db.getPackages("index:project:subrepo", false)
}

func (db *Database) Index() ([]Package, error) {
	return db.getPackages("index:all:", false)
}

func (db *Database) Project(projectRoot string) ([]Package, error) {
	return db.getPackages("index:project:"+normalizeProjectRoot(projectRoot), true)
}

func (db *Database) AllPackages() ([]Package, error) {
	c := db.Pool.Get()
	defer c.Close()
	values, err := redis.Values(c.Do("SORT", "nextCrawl", "DESC", "BY", "pkg:*->score", "GET", "pkg:*->path", "GET", "pkg:*->kind"))
	if err != nil {
		return nil, err
	}
	result := make([]Package, 0, len(values)/2)
	for len(values) > 0 {
		var pkg Package
		var kind string
		values, err = redis.Scan(values, &pkg.Path, &kind)
		if err != nil {
			return nil, err
		}
		if kind == "d" {
			continue
		}
		result = append(result, pkg)
	}
	return result, nil
}

var packagesScript = redis.NewScript(0, `
    local result = {}
    for i = 1,#ARGV do
        local path = ARGV[i]
        local synopsis = ''
        local kind = 'u'
        local id = redis.call('HGET', 'ids',  path)
        if id then
            synopsis = redis.call('HGET', 'pkg:' .. id, 'synopsis')
            kind = redis.call('HGET', 'pkg:' .. id, 'kind')
        end
        result[#result+1] = path
        result[#result+1] = synopsis
        result[#result+1] = kind
    end
    return result
`)

func (db *Database) Packages(paths []string) ([]Package, error) {
	var args []interface{}
	for _, p := range paths {
		args = append(args, p)
	}
	c := db.Pool.Get()
	defer c.Close()
	reply, err := packagesScript.Do(c, args...)
	if err != nil {
		return nil, err
	}
	pkgs, err := packages(reply, true)
	sort.Sort(byPath(pkgs))
	return pkgs, err
}

func (db *Database) ImporterCount(path string) (int, error) {
	c := db.Pool.Get()
	defer c.Close()
	return redis.Int(c.Do("SCARD", "index:import:"+path))
}

func (db *Database) Importers(path string) ([]Package, error) {
	return db.getPackages("index:import:"+path, false)
}

func (db *Database) Block(root string) error {
	c := db.Pool.Get()
	defer c.Close()
	if _, err := c.Do("SADD", "block", root); err != nil {
		return err
	}
	keys, err := redis.Strings(c.Do("HKEYS", "ids"))
	if err != nil {
		return err
	}
	for _, key := range keys {
		if key == root || strings.HasPrefix(key, root) && key[len(root)] == '/' {
			if _, err := deleteScript.Do(c, key); err != nil {
				return err
			}
		}
	}
	return nil
}

var isBlockedScript = redis.NewScript(0, `
    local path = ''
    for s in string.gmatch(ARGV[1], '[^/]+') do
        path = path .. s
        if redis.call('SISMEMBER', 'block', path) == 1 then
            return 1
        end
        path = path .. '/'
    end
    return  0
`)

func (db *Database) IsBlocked(path string) (bool, error) {
	c := db.Pool.Get()
	defer c.Close()
	return redis.Bool(isBlockedScript.Do(c, path))
}

type queryResult struct {
	Path     string
	Synopsis string
	Score    float64
}

type byScore []*queryResult

func (p byScore) Len() int           { return len(p) }
func (p byScore) Less(i, j int) bool { return p[j].Score < p[i].Score }
func (p byScore) Swap(i, j int)      { p[i], p[j] = p[j], p[i] }

func (db *Database) Query(q string) ([]Package, error) {
	terms := parseQuery(q)
	if len(terms) == 0 {
		return nil, nil
	}
	c := db.Pool.Get()
	defer c.Close()
	n, err := redis.Int(c.Do("INCR", "maxQueryId"))
	if err != nil {
		return nil, err
	}
	id := "tmp:query-" + strconv.Itoa(n)

	args := []interface{}{id}
	for _, term := range terms {
		args = append(args, "index:"+term)
	}
	c.Send("SINTERSTORE", args...)
	c.Send("SORT", id, "DESC", "BY", "nosort", "GET", "pkg:*->path", "GET", "pkg:*->synopsis", "GET", "pkg:*->score")
	c.Send("DEL", id)
	c.Flush()
	c.Receive()                              // SINTERSTORE
	values, err := redis.Values(c.Receive()) // SORT
	if err != nil {
		return nil, err
	}
	c.Receive() // DEL

	var queryResults []*queryResult
	if err := redis.ScanSlice(values, &queryResults, "Path", "Synopsis", "Score"); err != nil {
		return nil, err
	}

	for _, qr := range queryResults {
		c.Send("SCARD", "index:import:"+qr.Path)
	}
	c.Flush()

	for _, qr := range queryResults {
		importCount, err := redis.Int(c.Receive())
		if err != nil {
			return nil, err
		}

		qr.Score *= math.Log(float64(10 + importCount))

		if isStandardPackage(qr.Path) {
			if strings.HasSuffix(qr.Path, q) {
				// Big bump for exact match on standard package name.
				qr.Score *= 10000
			} else {
				qr.Score *= 1.2
			}
		}

		if q == path.Base(qr.Path) {
			qr.Score *= 1.1
		}
	}

	sort.Sort(byScore(queryResults))

	pkgs := make([]Package, len(queryResults))
	for i, qr := range queryResults {
		pkgs[i].Path = qr.Path
		pkgs[i].Synopsis = qr.Synopsis
	}

	return pkgs, nil
}

type PackageInfo struct {
	PDoc  *doc.Package
	Score float64
	Kind  string
	Size  int
}

// Do executes function f for each document in the database.
func (db *Database) Do(f func(*PackageInfo) error) error {
	c := db.Pool.Get()
	defer c.Close()
	cursor := 0
	c.Send("SCAN", cursor, "MATCH", "pkg:*")
	c.Flush()
	for {
		// Receive previous SCAN.
		values, err := redis.Values(c.Receive())
		if err != nil {
			return err
		}
		var keys [][]byte
		if _, err := redis.Scan(values, &cursor, &keys); err != nil {
			return err
		}
		if cursor == 0 {
			break
		}
		for _, key := range keys {
			c.Send("HMGET", key, "gob", "score", "kind", "path", "terms", "synopis")
		}
		c.Send("SCAN", cursor, "MATCH", "pkg:*")
		c.Flush()
		for _ = range keys {
			values, err := redis.Values(c.Receive())
			if err != nil {
				return err
			}

			var (
				pi       PackageInfo
				p        []byte
				path     string
				terms    string
				synopsis string
			)

			if _, err := redis.Scan(values, &p, &pi.Score, &pi.Kind, &path, &terms, &synopsis); err != nil {
				return err
			}

			if p == nil {
				continue
			}

			pi.Size = len(path) + len(p) + len(terms) + len(synopsis)

			p, err = snappy.Decode(nil, p)
			if err != nil {
				return fmt.Errorf("snappy decoding %s: %v", path, err)
			}

			if err := gob.NewDecoder(bytes.NewReader(p)).Decode(&pi.PDoc); err != nil {
				return fmt.Errorf("gob decoding %s: %v", path, err)
			}
			if err := f(&pi); err != nil {
				return fmt.Errorf("func %s: %v", path, err)
			}
		}
	}
	return nil
}

var importGraphScript = redis.NewScript(0, `
    local path = ARGV[1]

    local id = redis.call('HGET', 'ids', path)
    if not id then
        return false
    end

    return redis.call('HMGET', 'pkg:' .. id, 'synopsis', 'terms')
`)

// DepLevel specifies the level of depdenencies to show in an import graph.
type DepLevel int

const (
	ShowAllDeps      DepLevel = iota // show all dependencies
	HideStandardDeps                 // don't show dependencies of standard libraries
	HideStandardAll                  // don't show standard libraries at all
)

func (db *Database) ImportGraph(pdoc *doc.Package, level DepLevel) ([]Package, [][2]int, error) {

	// This breadth-first traversal of the package's dependencies uses the
	// Redis pipeline as queue. Links to packages with invalid import paths are
	// only included for the root package.

	c := db.Pool.Get()
	defer c.Close()
	if err := importGraphScript.Load(c); err != nil {
		return nil, nil, err
	}

	nodes := []Package{{Path: pdoc.ImportPath, Synopsis: pdoc.Synopsis}}
	edges := [][2]int{}
	index := map[string]int{pdoc.ImportPath: 0}

	for _, path := range pdoc.Imports {
		if level >= HideStandardAll && isStandardPackage(path) {
			continue
		}
		j := len(nodes)
		index[path] = j
		edges = append(edges, [2]int{0, j})
		nodes = append(nodes, Package{Path: path})
		importGraphScript.Send(c, path)
	}

	for i := 1; i < len(nodes); i++ {
		c.Flush()
		r, err := redis.Values(c.Receive())
		if err == redis.ErrNil {
			continue
		} else if err != nil {
			return nil, nil, err
		}
		var synopsis, terms string
		if _, err := redis.Scan(r, &synopsis, &terms); err != nil {
			return nil, nil, err
		}
		nodes[i].Synopsis = synopsis
		for _, term := range strings.Fields(terms) {
			if strings.HasPrefix(term, "import:") {
				path := term[len("import:"):]
				if level >= HideStandardDeps && isStandardPackage(path) {
					continue
				}
				j, ok := index[path]
				if !ok {
					j = len(nodes)
					index[path] = j
					nodes = append(nodes, Package{Path: path})
					importGraphScript.Send(c, path)
				}
				edges = append(edges, [2]int{i, j})
			}
		}
	}
	return nodes, edges, nil
}

func (db *Database) PutGob(key string, value interface{}) error {
	var buf bytes.Buffer
	if err := gob.NewEncoder(&buf).Encode(value); err != nil {
		return err
	}
	c := db.Pool.Get()
	defer c.Close()
	_, err := c.Do("SET", "gob:"+key, buf.Bytes())
	return err
}

func (db *Database) GetGob(key string, value interface{}) error {
	c := db.Pool.Get()
	defer c.Close()
	p, err := redis.Bytes(c.Do("GET", "gob:"+key))
	if err == redis.ErrNil {
		return nil
	} else if err != nil {
		return err
	}
	return gob.NewDecoder(bytes.NewReader(p)).Decode(value)
}

var incrementPopularScoreScript = redis.NewScript(0, `
    local path = ARGV[1]
    local n = ARGV[2]
    local t = ARGV[3]

    local id = redis.call('HGET', 'ids', path)
    if not id then
        return
    end

    local t0 = redis.call('GET', 'popular:0') or '0'
    local f = math.exp(tonumber(t) - tonumber(t0))
    redis.call('ZINCRBY', 'popular', tonumber(n) * f, id)
    if f > 10 then
        redis.call('SET', 'popular:0', t)
        redis.call('ZUNIONSTORE', 'popular', 1, 'popular', 'WEIGHTS', 1.0 / f)
        redis.call('ZREMRANGEBYSCORE', 'popular', '-inf', 0.05)
    end
`)

const popularHalfLife = time.Hour * 24 * 7

func (db *Database) incrementPopularScoreInternal(path string, delta float64, t time.Time) error {
	// nt = n0 * math.Exp(-lambda * t)
	// lambda = math.Ln2 / thalf
	c := db.Pool.Get()
	defer c.Close()
	const lambda = math.Ln2 / float64(popularHalfLife)
	scaledTime := lambda * float64(t.Sub(time.Unix(1257894000, 0)))
	_, err := incrementPopularScoreScript.Do(c, path, delta, scaledTime)
	return err
}

func (db *Database) IncrementPopularScore(path string) error {
	return db.incrementPopularScoreInternal(path, 1, time.Now())
}

var popularScript = redis.NewScript(0, `
    local stop = ARGV[1]
    local ids = redis.call('ZREVRANGE', 'popular', '0', stop)
    local result = {}
    for i=1,#ids do
        local values = redis.call('HMGET', 'pkg:' .. ids[i], 'path', 'synopsis', 'kind')
        result[#result+1] = values[1]
        result[#result+1] = values[2]
        result[#result+1] = values[3]
    end
    return result
`)

func (db *Database) Popular(count int) ([]Package, error) {
	c := db.Pool.Get()
	defer c.Close()
	reply, err := popularScript.Do(c, count-1)
	if err != nil {
		return nil, err
	}
	pkgs, err := packages(reply, false)
	return pkgs, err
}

var popularWithScoreScript = redis.NewScript(0, `
    local ids = redis.call('ZREVRANGE', 'popular', '0', -1, 'WITHSCORES')
    local result = {}
    for i=1,#ids,2 do
        result[#result+1] = redis.call('HGET', 'pkg:' .. ids[i], 'path')
        result[#result+1] = ids[i+1]
        result[#result+1] = 'p'
    end
    return result
`)

func (db *Database) PopularWithScores() ([]Package, error) {
	c := db.Pool.Get()
	defer c.Close()
	reply, err := popularWithScoreScript.Do(c)
	if err != nil {
		return nil, err
	}
	pkgs, err := packages(reply, false)
	return pkgs, err
}

func (db *Database) PopNewCrawl() (string, bool, error) {
	c := db.Pool.Get()
	defer c.Close()

	var subdirs []Package

	path, err := redis.String(c.Do("SPOP", "newCrawl"))
	switch {
	case err == redis.ErrNil:
		err = nil
		path = ""
	case err == nil:
		subdirs, err = db.getSubdirs(c, path, nil)
	}
	return path, len(subdirs) > 0, err
}

func (db *Database) AddBadCrawl(path string) error {
	c := db.Pool.Get()
	defer c.Close()
	_, err := c.Do("SADD", "badCrawl", path)
	return err
}

var incrementCounterScript = redis.NewScript(0, `
    local key = 'counter:' .. ARGV[1]
    local n = tonumber(ARGV[2])
    local t = tonumber(ARGV[3])
    local exp = tonumber(ARGV[4])

    local counter = redis.call('GET', key)
    if counter then
        counter = cjson.decode(counter)
        n = n + counter.n * math.exp(counter.t - t)
    end

    redis.call('SET', key, cjson.encode({n = n; t = t}))
    redis.call('EXPIRE', key, exp)
    return tostring(n)
`)

const counterHalflife = time.Hour

func (db *Database) incrementCounterInternal(key string, delta float64, t time.Time) (float64, error) {
	// nt = n0 * math.Exp(-lambda * t)
	// lambda = math.Ln2 / thalf
	c := db.Pool.Get()
	defer c.Close()
	const lambda = math.Ln2 / float64(counterHalflife)
	scaledTime := lambda * float64(t.Sub(time.Unix(1257894000, 0)))
	return redis.Float64(incrementCounterScript.Do(c, key, delta, scaledTime, (4*counterHalflife)/time.Second))
}

func (db *Database) IncrementCounter(key string, delta float64) (float64, error) {
	return db.incrementCounterInternal(key, delta, time.Now())
}

// Reindex gets all the packages in database and put them into the search index.
// This will update the search index with the path, synopsis, score, import counts
// of all the packages in the database.
func (db *Database) Reindex(ctx context.Context) error {
	c := db.Pool.Get()
	defer c.Close()

	idx, err := search.Open("packages")
	if err != nil {
		return fmt.Errorf("database: failed to open packages: %v", err)
	}
	npkgs := 0
	for {
		// Get 200 packages from the nextCrawl set each time. Use npkgs as a cursor
		// to store the current position we actually indexed. Retry from the cursor
		// position if we received a timeout error from app engine.
		values, err := redis.Values(c.Do(
			"SORT", "nextCrawl",
			"LIMIT", strconv.Itoa(npkgs), "200",
			"GET", "pkg:*->path",
			"GET", "pkg:*->synopsis",
			"GET", "pkg:*->score",
		))
		if err != nil {
			return err
		}
		if len(values) == 0 {
			break // all done
		}

		// The Search API should support put in batches of up to 200 documents,
		// the Go version of this API does not support this yet.
		// TODO(shantuo): Put packages in batch operations.
		for ; len(values) > 0; npkgs++ {
			var pdoc doc.Package
			var score float64
			values, err = redis.Scan(values, &pdoc.ImportPath, &pdoc.Synopsis, &score)
			if err != nil {
				return err
			}
			// There are some corrupted data in our current database
			// that causes an error when putting the package into the
			// search index which only supports UTF8 encoding.
			if !utf8.ValidString(pdoc.Synopsis) {
				pdoc.Synopsis = ""
			}
			id, n, err := pkgIDAndImportCount(c, pdoc.ImportPath)
			if err != nil {
				return err
			}
			if _, err := idx.Put(ctx, id, &Package{
				Path:        pdoc.ImportPath,
				Synopsis:    pdoc.Synopsis,
				Score:       score,
				ImportCount: n,
			}); err != nil {
				if appengine.IsTimeoutError(err) {
					log.Printf("App Engine timeout: %v. Continue...", err)
					break
				}
				return fmt.Errorf("Failed to put index %s: %v", id, err)
			}
		}
	}
	log.Printf("%d packages are reindexed", npkgs)
	return nil
}
