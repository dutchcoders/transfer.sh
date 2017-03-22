// Copyright 2017 The Go Authors. All rights reserved.
//
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file or at
// https://developers.google.com/open-source/licenses/bsd.

// Command gddo-server is the GoPkgDoc server.
package main

import (
	"bytes"
	"crypto/md5"
	"encoding/json"
	"errors"
	"fmt"
	"go/build"
	"html/template"
	"io"
	"log"
	"net/http"
	"os"
	"path"
	"regexp"
	"runtime/debug"
	"sort"
	"strconv"
	"strings"
	"time"

	"cloud.google.com/go/compute/metadata"
	"cloud.google.com/go/logging"
	"github.com/spf13/viper"
	"golang.org/x/net/context"
	"google.golang.org/appengine"

	"github.com/golang/gddo/database"
	"github.com/golang/gddo/doc"
	"github.com/golang/gddo/gosrc"
	"github.com/golang/gddo/httputil"
)

const (
	jsonMIMEType = "application/json; charset=utf-8"
	textMIMEType = "text/plain; charset=utf-8"
	htmlMIMEType = "text/html; charset=utf-8"
)

var errUpdateTimeout = errors.New("refresh timeout")

type httpError struct {
	status int   // HTTP status code.
	err    error // Optional reason for the HTTP error.
}

func (err *httpError) Error() string {
	if err.err != nil {
		return fmt.Sprintf("status %d, reason %s", err.status, err.err.Error())
	}
	return fmt.Sprintf("Status %d", err.status)
}

const (
	humanRequest = iota
	robotRequest
	queryRequest
	refreshRequest
	apiRequest
)

type crawlResult struct {
	pdoc *doc.Package
	err  error
}

// getDoc gets the package documentation from the database or from the version
// control system as needed.
func getDoc(path string, requestType int) (*doc.Package, []database.Package, error) {
	if path == "-" {
		// A hack in the database package uses the path "-" to represent the
		// next document to crawl. Block "-" here so that requests to /- always
		// return not found.
		return nil, nil, &httpError{status: http.StatusNotFound}
	}

	pdoc, pkgs, nextCrawl, err := db.Get(path)
	if err != nil {
		return nil, nil, err
	}

	needsCrawl := false
	switch requestType {
	case queryRequest, apiRequest:
		needsCrawl = nextCrawl.IsZero() && len(pkgs) == 0
	case humanRequest:
		needsCrawl = nextCrawl.Before(time.Now())
	case robotRequest:
		needsCrawl = nextCrawl.IsZero() && len(pkgs) > 0
	}

	if !needsCrawl {
		return pdoc, pkgs, nil
	}

	c := make(chan crawlResult, 1)
	go func() {
		pdoc, err := crawlDoc("web  ", path, pdoc, len(pkgs) > 0, nextCrawl)
		c <- crawlResult{pdoc, err}
	}()

	timeout := viper.GetDuration(ConfigGetTimeout)
	if pdoc == nil {
		timeout = viper.GetDuration(ConfigFirstGetTimeout)
	}

	select {
	case cr := <-c:
		err = cr.err
		if err == nil {
			pdoc = cr.pdoc
		}
	case <-time.After(timeout):
		err = errUpdateTimeout
	}

	switch {
	case err == nil:
		return pdoc, pkgs, nil
	case gosrc.IsNotFound(err):
		return nil, nil, err
	case pdoc != nil:
		log.Printf("Serving %q from database after error getting doc: %v", path, err)
		return pdoc, pkgs, nil
	case err == errUpdateTimeout:
		log.Printf("Serving %q as not found after timeout getting doc", path)
		return nil, nil, &httpError{status: http.StatusNotFound}
	default:
		return nil, nil, err
	}
}

func templateExt(req *http.Request) string {
	if httputil.NegotiateContentType(req, []string{"text/html", "text/plain"}, "text/html") == "text/plain" {
		return ".txt"
	}
	return ".html"
}

var (
	robotPat = regexp.MustCompile(`(:?\+https?://)|(?:\Wbot\W)|(?:^Python-urllib)|(?:^Go )|(?:^Java/)`)
)

func isRobot(req *http.Request) bool {
	if robotPat.MatchString(req.Header.Get("User-Agent")) {
		return true
	}
	host := httputil.StripPort(req.RemoteAddr)
	n, err := db.IncrementCounter(host, 1)
	if err != nil {
		log.Printf("error incrementing counter for %s, %v", host, err)
		return false
	}
	if n > viper.GetFloat64(ConfigRobotThreshold) {
		log.Printf("robot %.2f %s %s", n, host, req.Header.Get("User-Agent"))
		return true
	}
	return false
}

func popularLinkReferral(req *http.Request) bool {
	return strings.HasSuffix(req.Header.Get("Referer"), "//"+req.Host+"/")
}

func isView(req *http.Request, key string) bool {
	rq := req.URL.RawQuery
	return strings.HasPrefix(rq, key) &&
		(len(rq) == len(key) || rq[len(key)] == '=' || rq[len(key)] == '&')
}

// httpEtag returns the package entity tag used in HTTP transactions.
func httpEtag(pdoc *doc.Package, pkgs []database.Package, importerCount int, flashMessages []flashMessage) string {
	b := make([]byte, 0, 128)
	b = strconv.AppendInt(b, pdoc.Updated.Unix(), 16)
	b = append(b, 0)
	b = append(b, pdoc.Etag...)
	if importerCount >= 8 {
		importerCount = 8
	}
	b = append(b, 0)
	b = strconv.AppendInt(b, int64(importerCount), 16)
	for _, pkg := range pkgs {
		b = append(b, 0)
		b = append(b, pkg.Path...)
		b = append(b, 0)
		b = append(b, pkg.Synopsis...)
	}
	if viper.GetBool(ConfigSidebar) {
		b = append(b, "\000xsb"...)
	}
	for _, m := range flashMessages {
		b = append(b, 0)
		b = append(b, m.ID...)
		for _, a := range m.Args {
			b = append(b, 1)
			b = append(b, a...)
		}
	}
	h := md5.New()
	h.Write(b)
	b = h.Sum(b[:0])
	return fmt.Sprintf("\"%x\"", b)
}

func servePackage(resp http.ResponseWriter, req *http.Request) error {
	p := path.Clean(req.URL.Path)
	if strings.HasPrefix(p, "/pkg/") {
		p = p[len("/pkg"):]
	}
	if p != req.URL.Path {
		http.Redirect(resp, req, p, http.StatusMovedPermanently)
		return nil
	}

	if isView(req, "status.svg") {
		statusImageHandlerSVG.ServeHTTP(resp, req)
		return nil
	}

	if isView(req, "status.png") {
		statusImageHandlerPNG.ServeHTTP(resp, req)
		return nil
	}

	requestType := humanRequest
	if isRobot(req) {
		requestType = robotRequest
	}

	importPath := strings.TrimPrefix(req.URL.Path, "/")
	pdoc, pkgs, err := getDoc(importPath, requestType)

	if e, ok := err.(gosrc.NotFoundError); ok && e.Redirect != "" {
		// To prevent dumb clients from following redirect loops, respond with
		// status 404 if the target document is not found.
		if _, _, err := getDoc(e.Redirect, requestType); gosrc.IsNotFound(err) {
			return &httpError{status: http.StatusNotFound}
		}
		u := "/" + e.Redirect
		if req.URL.RawQuery != "" {
			u += "?" + req.URL.RawQuery
		}
		setFlashMessages(resp, []flashMessage{{ID: "redir", Args: []string{importPath}}})
		http.Redirect(resp, req, u, http.StatusFound)
		return nil
	}
	if err != nil {
		return err
	}

	flashMessages := getFlashMessages(resp, req)

	if pdoc == nil {
		if len(pkgs) == 0 {
			return &httpError{status: http.StatusNotFound}
		}
		pdocChild, _, _, err := db.Get(pkgs[0].Path)
		if err != nil {
			return err
		}
		pdoc = &doc.Package{
			ProjectName: pdocChild.ProjectName,
			ProjectRoot: pdocChild.ProjectRoot,
			ProjectURL:  pdocChild.ProjectURL,
			ImportPath:  importPath,
		}
	}

	switch {
	case len(req.Form) == 0:
		importerCount := 0
		if pdoc.Name != "" {
			importerCount, err = db.ImporterCount(importPath)
			if err != nil {
				return err
			}
		}

		etag := httpEtag(pdoc, pkgs, importerCount, flashMessages)
		status := http.StatusOK
		if req.Header.Get("If-None-Match") == etag {
			status = http.StatusNotModified
		}

		if requestType == humanRequest &&
			pdoc.Name != "" && // not a directory
			pdoc.ProjectRoot != "" && // not a standard package
			!pdoc.IsCmd &&
			len(pdoc.Errors) == 0 &&
			!popularLinkReferral(req) {
			if err := db.IncrementPopularScore(pdoc.ImportPath); err != nil {
				log.Printf("ERROR db.IncrementPopularScore(%s): %v", pdoc.ImportPath, err)
			}
		}
		if gceLogger != nil {
			gceLogger.LogEvent(resp, req, nil)
		}

		template := "dir"
		switch {
		case pdoc.IsCmd:
			template = "cmd"
		case pdoc.Name != "":
			template = "pkg"
		}
		template += templateExt(req)

		return executeTemplate(resp, template, status, http.Header{"Etag": {etag}}, map[string]interface{}{
			"flashMessages": flashMessages,
			"pkgs":          pkgs,
			"pdoc":          newTDoc(pdoc),
			"importerCount": importerCount,
		})
	case isView(req, "imports"):
		if pdoc.Name == "" {
			break
		}
		pkgs, err = db.Packages(pdoc.Imports)
		if err != nil {
			return err
		}
		return executeTemplate(resp, "imports.html", http.StatusOK, nil, map[string]interface{}{
			"flashMessages": flashMessages,
			"pkgs":          pkgs,
			"pdoc":          newTDoc(pdoc),
		})
	case isView(req, "tools"):
		proto := "http"
		if req.Host == "godoc.org" {
			proto = "https"
		}
		return executeTemplate(resp, "tools.html", http.StatusOK, nil, map[string]interface{}{
			"flashMessages": flashMessages,
			"uri":           fmt.Sprintf("%s://%s/%s", proto, req.Host, importPath),
			"pdoc":          newTDoc(pdoc),
		})
	case isView(req, "importers"):
		if pdoc.Name == "" {
			break
		}
		pkgs, err = db.Importers(importPath)
		if err != nil {
			return err
		}
		template := "importers.html"
		if requestType == robotRequest {
			// Hide back links from robots.
			template = "importers_robot.html"
		}
		return executeTemplate(resp, template, http.StatusOK, nil, map[string]interface{}{
			"flashMessages": flashMessages,
			"pkgs":          pkgs,
			"pdoc":          newTDoc(pdoc),
		})
	case isView(req, "import-graph"):
		if requestType == robotRequest {
			return &httpError{status: http.StatusForbidden}
		}
		if pdoc.Name == "" {
			break
		}
		hide := database.ShowAllDeps
		switch req.Form.Get("hide") {
		case "1":
			hide = database.HideStandardDeps
		case "2":
			hide = database.HideStandardAll
		}
		pkgs, edges, err := db.ImportGraph(pdoc, hide)
		if err != nil {
			return err
		}
		b, err := renderGraph(pdoc, pkgs, edges)
		if err != nil {
			return err
		}
		return executeTemplate(resp, "graph.html", http.StatusOK, nil, map[string]interface{}{
			"flashMessages": flashMessages,
			"svg":           template.HTML(b),
			"pdoc":          newTDoc(pdoc),
			"hide":          hide,
		})
	case isView(req, "play"):
		u, err := playURL(pdoc, req.Form.Get("play"), req.Header.Get("X-AppEngine-Country"))
		if err != nil {
			return err
		}
		http.Redirect(resp, req, u, http.StatusMovedPermanently)
		return nil
	case req.Form.Get("view") != "":
		// Redirect deprecated view= queries.
		var q string
		switch view := req.Form.Get("view"); view {
		case "imports", "importers":
			q = view
		case "import-graph":
			if req.Form.Get("hide") == "1" {
				q = "import-graph&hide=1"
			} else {
				q = "import-graph"
			}
		}
		if q != "" {
			u := *req.URL
			u.RawQuery = q
			http.Redirect(resp, req, u.String(), http.StatusMovedPermanently)
			return nil
		}
	}
	return &httpError{status: http.StatusNotFound}
}

func serveRefresh(resp http.ResponseWriter, req *http.Request) error {
	importPath := req.Form.Get("path")
	_, pkgs, _, err := db.Get(importPath)
	if err != nil {
		return err
	}
	c := make(chan error, 1)
	go func() {
		_, err := crawlDoc("rfrsh", importPath, nil, len(pkgs) > 0, time.Time{})
		c <- err
	}()
	select {
	case err = <-c:
	case <-time.After(viper.GetDuration(ConfigGetTimeout)):
		err = errUpdateTimeout
	}
	if e, ok := err.(gosrc.NotFoundError); ok && e.Redirect != "" {
		setFlashMessages(resp, []flashMessage{{ID: "redir", Args: []string{importPath}}})
		importPath = e.Redirect
		err = nil
	} else if err != nil {
		setFlashMessages(resp, []flashMessage{{ID: "refresh", Args: []string{errorText(err)}}})
	}
	http.Redirect(resp, req, "/"+importPath, http.StatusFound)
	return nil
}

func serveGoIndex(resp http.ResponseWriter, req *http.Request) error {
	pkgs, err := db.GoIndex()
	if err != nil {
		return err
	}
	return executeTemplate(resp, "std.html", http.StatusOK, nil, map[string]interface{}{
		"pkgs": pkgs,
	})
}

func serveGoSubrepoIndex(resp http.ResponseWriter, req *http.Request) error {
	pkgs, err := db.GoSubrepoIndex()
	if err != nil {
		return err
	}
	return executeTemplate(resp, "subrepo.html", http.StatusOK, nil, map[string]interface{}{
		"pkgs": pkgs,
	})
}

func runReindex(resp http.ResponseWriter, req *http.Request) {
	fmt.Fprintln(resp, "Reindexing...")
	go reindex()
}

func runPurgeIndex(resp http.ResponseWriter, req *http.Request) {
	fmt.Fprintln(resp, "Purging the search index...")
	go purgeIndex()
}

type byPath struct {
	pkgs []database.Package
	rank []int
}

func (bp *byPath) Len() int           { return len(bp.pkgs) }
func (bp *byPath) Less(i, j int) bool { return bp.pkgs[i].Path < bp.pkgs[j].Path }
func (bp *byPath) Swap(i, j int) {
	bp.pkgs[i], bp.pkgs[j] = bp.pkgs[j], bp.pkgs[i]
	bp.rank[i], bp.rank[j] = bp.rank[j], bp.rank[i]
}

type byRank struct {
	pkgs []database.Package
	rank []int
}

func (br *byRank) Len() int           { return len(br.pkgs) }
func (br *byRank) Less(i, j int) bool { return br.rank[i] < br.rank[j] }
func (br *byRank) Swap(i, j int) {
	br.pkgs[i], br.pkgs[j] = br.pkgs[j], br.pkgs[i]
	br.rank[i], br.rank[j] = br.rank[j], br.rank[i]
}

func popular() ([]database.Package, error) {
	const n = 25

	pkgs, err := db.Popular(2 * n)
	if err != nil {
		return nil, err
	}

	rank := make([]int, len(pkgs))
	for i := range pkgs {
		rank[i] = i
	}

	sort.Sort(&byPath{pkgs, rank})

	j := 0
	prev := "."
	for i, pkg := range pkgs {
		if strings.HasPrefix(pkg.Path, prev) {
			if rank[j-1] < rank[i] {
				rank[j-1] = rank[i]
			}
			continue
		}
		prev = pkg.Path + "/"
		pkgs[j] = pkg
		rank[j] = rank[i]
		j++
	}
	pkgs = pkgs[:j]

	sort.Sort(&byRank{pkgs, rank})

	if len(pkgs) > n {
		pkgs = pkgs[:n]
	}

	sort.Sort(&byPath{pkgs, rank})

	return pkgs, nil
}

func serveHome(resp http.ResponseWriter, req *http.Request) error {
	if req.URL.Path != "/" {
		return servePackage(resp, req)
	}

	q := strings.TrimSpace(req.Form.Get("q"))
	if q == "" {
		pkgs, err := popular()
		if err != nil {
			return err
		}

		return executeTemplate(resp, "home"+templateExt(req), http.StatusOK, nil,
			map[string]interface{}{"Popular": pkgs})
	}

	if path, ok := isBrowseURL(q); ok {
		q = path
	}

	if gosrc.IsValidRemotePath(q) || (strings.Contains(q, "/") && gosrc.IsGoRepoPath(q)) {
		pdoc, pkgs, err := getDoc(q, queryRequest)
		if e, ok := err.(gosrc.NotFoundError); ok && e.Redirect != "" {
			http.Redirect(resp, req, "/"+e.Redirect, http.StatusFound)
			return nil
		}
		if err == nil && (pdoc != nil || len(pkgs) > 0) {
			http.Redirect(resp, req, "/"+q, http.StatusFound)
			return nil
		}
	}

	ctx := appengine.NewContext(req)
	pkgs, err := database.Search(ctx, q)
	if err != nil {
		return err
	}
	if gceLogger != nil {
		// Log up to top 10 packages we served upon a search.
		logPkgs := pkgs
		if len(pkgs) > 10 {
			logPkgs = pkgs[:10]
		}
		gceLogger.LogEvent(resp, req, logPkgs)
	}

	return executeTemplate(resp, "results"+templateExt(req), http.StatusOK, nil,
		map[string]interface{}{"q": q, "pkgs": pkgs})
}

func serveAbout(resp http.ResponseWriter, req *http.Request) error {
	return executeTemplate(resp, "about.html", http.StatusOK, nil,
		map[string]interface{}{"Host": req.Host})
}

func serveBot(resp http.ResponseWriter, req *http.Request) error {
	return executeTemplate(resp, "bot.html", http.StatusOK, nil, nil)
}

func serveHealthCheck(resp http.ResponseWriter, req *http.Request) {
	resp.Write([]byte("Health check: ok\n"))
}

func logError(req *http.Request, err error, rv interface{}) {
	if err != nil {
		var buf bytes.Buffer
		fmt.Fprintf(&buf, "Error serving %s: %v\n", req.URL, err)
		if rv != nil {
			fmt.Fprintln(&buf, rv)
			buf.Write(debug.Stack())
		}
		log.Print(buf.String())
	}
}

func serveAPISearch(resp http.ResponseWriter, req *http.Request) error {
	q := strings.TrimSpace(req.Form.Get("q"))

	var pkgs []database.Package

	if gosrc.IsValidRemotePath(q) || (strings.Contains(q, "/") && gosrc.IsGoRepoPath(q)) {
		pdoc, _, err := getDoc(q, apiRequest)
		if e, ok := err.(gosrc.NotFoundError); ok && e.Redirect != "" {
			pdoc, _, err = getDoc(e.Redirect, robotRequest)
		}
		if err == nil && pdoc != nil {
			pkgs = []database.Package{{Path: pdoc.ImportPath, Synopsis: pdoc.Synopsis}}
		}
	}

	if pkgs == nil {
		var err error
		ctx := appengine.NewContext(req)
		pkgs, err = database.Search(ctx, q)
		if err != nil {
			return err
		}
	}

	var data = struct {
		Results []database.Package `json:"results"`
	}{
		pkgs,
	}
	resp.Header().Set("Content-Type", jsonMIMEType)
	return json.NewEncoder(resp).Encode(&data)
}

func serveAPIPackages(resp http.ResponseWriter, req *http.Request) error {
	pkgs, err := db.AllPackages()
	if err != nil {
		return err
	}
	data := struct {
		Results []database.Package `json:"results"`
	}{
		pkgs,
	}
	resp.Header().Set("Content-Type", jsonMIMEType)
	return json.NewEncoder(resp).Encode(&data)
}

func serveAPIImporters(resp http.ResponseWriter, req *http.Request) error {
	importPath := strings.TrimPrefix(req.URL.Path, "/importers/")
	pkgs, err := db.Importers(importPath)
	if err != nil {
		return err
	}
	data := struct {
		Results []database.Package `json:"results"`
	}{
		pkgs,
	}
	resp.Header().Set("Content-Type", jsonMIMEType)
	return json.NewEncoder(resp).Encode(&data)
}

func serveAPIImports(resp http.ResponseWriter, req *http.Request) error {
	importPath := strings.TrimPrefix(req.URL.Path, "/imports/")
	pdoc, _, err := getDoc(importPath, robotRequest)
	if err != nil {
		return err
	}
	if pdoc == nil || pdoc.Name == "" {
		return &httpError{status: http.StatusNotFound}
	}
	imports, err := db.Packages(pdoc.Imports)
	if err != nil {
		return err
	}
	testImports, err := db.Packages(pdoc.TestImports)
	if err != nil {
		return err
	}
	data := struct {
		Imports     []database.Package `json:"imports"`
		TestImports []database.Package `json:"testImports"`
	}{
		imports,
		testImports,
	}
	resp.Header().Set("Content-Type", jsonMIMEType)
	return json.NewEncoder(resp).Encode(&data)
}

func serveAPIHome(resp http.ResponseWriter, req *http.Request) error {
	return &httpError{status: http.StatusNotFound}
}

func runHandler(resp http.ResponseWriter, req *http.Request,
	fn func(resp http.ResponseWriter, req *http.Request) error, errfn httputil.Error) {
	defer func() {
		if rv := recover(); rv != nil {
			err := errors.New("handler panic")
			logError(req, err, rv)
			errfn(resp, req, http.StatusInternalServerError, err)
		}
	}()

	// TODO(stephenmw): choose headers based on if we are on App Engine
	if viper.GetBool(ConfigTrustProxyHeaders) {
		// If running on GAE, use X-Appengine-User-Ip to identify real ip of requests.
		if s := req.Header.Get("X-Appengine-User-Ip"); s != "" {
			req.RemoteAddr = s
		} else if s := req.Header.Get("X-Real-Ip"); s != "" {
			req.RemoteAddr = s
		}
	}

	req.Body = http.MaxBytesReader(resp, req.Body, 2048)
	req.ParseForm()
	var rb httputil.ResponseBuffer
	err := fn(&rb, req)
	if err == nil {
		rb.WriteTo(resp)
	} else if e, ok := err.(*httpError); ok {
		if e.status >= 500 {
			logError(req, err, nil)
		}
		errfn(resp, req, e.status, e.err)
	} else if gosrc.IsNotFound(err) {
		errfn(resp, req, http.StatusNotFound, nil)
	} else {
		logError(req, err, nil)
		errfn(resp, req, http.StatusInternalServerError, err)
	}
}

type handler func(resp http.ResponseWriter, req *http.Request) error

func (h handler) ServeHTTP(resp http.ResponseWriter, req *http.Request) {
	runHandler(resp, req, h, handleError)
}

type apiHandler func(resp http.ResponseWriter, req *http.Request) error

func (h apiHandler) ServeHTTP(resp http.ResponseWriter, req *http.Request) {
	runHandler(resp, req, h, handleAPIError)
}

func errorText(err error) string {
	if err == errUpdateTimeout {
		return "Timeout getting package files from the version control system."
	}
	if e, ok := err.(*gosrc.RemoteError); ok {
		return "Error getting package files from " + e.Host + "."
	}
	return "Internal server error."
}

func handleError(resp http.ResponseWriter, req *http.Request, status int, err error) {
	switch status {
	case http.StatusNotFound:
		executeTemplate(resp, "notfound"+templateExt(req), status, nil, map[string]interface{}{
			"flashMessages": getFlashMessages(resp, req),
		})
	default:
		resp.Header().Set("Content-Type", textMIMEType)
		resp.WriteHeader(http.StatusInternalServerError)
		io.WriteString(resp, errorText(err))
	}
}

func handleAPIError(resp http.ResponseWriter, req *http.Request, status int, err error) {
	var data struct {
		Error struct {
			Message string `json:"message"`
		} `json:"error"`
	}
	data.Error.Message = http.StatusText(status)
	resp.Header().Set("Content-Type", jsonMIMEType)
	resp.WriteHeader(status)
	json.NewEncoder(resp).Encode(&data)
}

type rootHandler []struct {
	prefix string
	h      http.Handler
}

func (m rootHandler) ServeHTTP(resp http.ResponseWriter, req *http.Request) {
	var h http.Handler
	for _, ph := range m {
		if strings.HasPrefix(req.Host, ph.prefix) {
			h = ph.h
			break
		}
	}

	h.ServeHTTP(resp, req)
}

// otherDomainHandler redirects to another domain keeping the rest of the URL.
type otherDomainHandler struct {
	scheme       string
	targetDomain string
}

func (h otherDomainHandler) ServeHTTP(w http.ResponseWriter, req *http.Request) {
	u := *req.URL
	u.Scheme = h.scheme
	u.Host = h.targetDomain
	http.Redirect(w, req, u.String(), http.StatusFound)
}

func defaultBase(path string) string {
	p, err := build.Default.Import(path, "", build.FindOnly)
	if err != nil {
		return "."
	}
	return p.Dir
}

var (
	db                    *database.Database
	httpClient            *http.Client
	statusImageHandlerPNG http.Handler
	statusImageHandlerSVG http.Handler
	gceLogger             *GCELogger
)

func main() {
	doc.SetDefaultGOOS(viper.GetString(ConfigDefaultGOOS))
	httpClient = newHTTPClient()

	var (
		gceLogName string
		projID     string
	)

	// TODO(stephenmw): merge into viper config infrastructure.
	if metadata.OnGCE() {
		acct, err := metadata.ProjectAttributeValue("ga-account")
		if err != nil {
			log.Printf("querying metadata for ga-account: %v", err)
		} else {
			gaAccount = acct
		}

		// Get the log name on GCE and setup context for creating a GCE log client.
		if name, err := metadata.ProjectAttributeValue("gce-log-name"); err != nil {
			log.Printf("querying metadata for gce-log-name: %v", err)
		} else {
			gceLogName = name
			if id, err := metadata.ProjectID(); err != nil {
				log.Printf("querying metadata for project ID: %v", err)
			} else {
				projID = id
			}
		}
	} else {
		gaAccount = os.Getenv("GA_ACCOUNT")
	}

	if err := parseHTMLTemplates([][]string{
		{"about.html", "common.html", "layout.html"},
		{"bot.html", "common.html", "layout.html"},
		{"cmd.html", "common.html", "layout.html"},
		{"dir.html", "common.html", "layout.html"},
		{"home.html", "common.html", "layout.html"},
		{"importers.html", "common.html", "layout.html"},
		{"importers_robot.html", "common.html", "layout.html"},
		{"imports.html", "common.html", "layout.html"},
		{"notfound.html", "common.html", "layout.html"},
		{"pkg.html", "common.html", "layout.html"},
		{"results.html", "common.html", "layout.html"},
		{"tools.html", "common.html", "layout.html"},
		{"std.html", "common.html", "layout.html"},
		{"subrepo.html", "common.html", "layout.html"},
		{"graph.html", "common.html"},
	}); err != nil {
		log.Fatal(err)
	}

	if err := parseTextTemplates([][]string{
		{"cmd.txt", "common.txt"},
		{"dir.txt", "common.txt"},
		{"home.txt", "common.txt"},
		{"notfound.txt", "common.txt"},
		{"pkg.txt", "common.txt"},
		{"results.txt", "common.txt"},
	}); err != nil {
		log.Fatal(err)
	}

	var err error
	db, err = database.New()
	if err != nil {
		log.Fatalf("Error opening database: %v", err)
	}

	go runBackgroundTasks()

	staticServer := httputil.StaticServer{
		Dir:    viper.GetString(ConfigAssetsDir),
		MaxAge: time.Hour,
		MIMETypes: map[string]string{
			".css": "text/css; charset=utf-8",
			".js":  "text/javascript; charset=utf-8",
		},
	}
	statusImageHandlerPNG = staticServer.FileHandler("status.png")
	statusImageHandlerSVG = staticServer.FileHandler("status.svg")

	apiMux := http.NewServeMux()
	apiMux.Handle("/favicon.ico", staticServer.FileHandler("favicon.ico"))
	apiMux.Handle("/google3d2f3cd4cc2bb44b.html", staticServer.FileHandler("google3d2f3cd4cc2bb44b.html"))
	apiMux.Handle("/humans.txt", staticServer.FileHandler("humans.txt"))
	apiMux.Handle("/robots.txt", staticServer.FileHandler("apiRobots.txt"))
	apiMux.Handle("/search", apiHandler(serveAPISearch))
	apiMux.Handle("/packages", apiHandler(serveAPIPackages))
	apiMux.Handle("/importers/", apiHandler(serveAPIImporters))
	apiMux.Handle("/imports/", apiHandler(serveAPIImports))
	apiMux.Handle("/", apiHandler(serveAPIHome))

	mux := http.NewServeMux()
	mux.Handle("/-/site.js", staticServer.FilesHandler(
		"third_party/jquery.timeago.js",
		"site.js"))
	mux.Handle("/-/site.css", staticServer.FilesHandler("site.css"))
	mux.Handle("/-/bootstrap.min.css", staticServer.FilesHandler("bootstrap.min.css"))
	mux.Handle("/-/bootstrap.min.js", staticServer.FilesHandler("bootstrap.min.js"))
	mux.Handle("/-/jquery-2.0.3.min.js", staticServer.FilesHandler("jquery-2.0.3.min.js"))
	if viper.GetBool(ConfigSidebar) {
		mux.Handle("/-/sidebar.css", staticServer.FilesHandler("sidebar.css"))
	}
	mux.Handle("/-/", http.NotFoundHandler())

	mux.Handle("/-/about", handler(serveAbout))
	mux.Handle("/-/bot", handler(serveBot))
	mux.Handle("/-/go", handler(serveGoIndex))
	mux.Handle("/-/subrepo", handler(serveGoSubrepoIndex))
	mux.Handle("/-/refresh", handler(serveRefresh))
	mux.Handle("/-/admin/reindex", http.HandlerFunc(runReindex))
	mux.Handle("/-/admin/purgeindex", http.HandlerFunc(runPurgeIndex))
	mux.Handle("/about", http.RedirectHandler("/-/about", http.StatusMovedPermanently))
	mux.Handle("/favicon.ico", staticServer.FileHandler("favicon.ico"))
	mux.Handle("/google3d2f3cd4cc2bb44b.html", staticServer.FileHandler("google3d2f3cd4cc2bb44b.html"))
	mux.Handle("/humans.txt", staticServer.FileHandler("humans.txt"))
	mux.Handle("/robots.txt", staticServer.FileHandler("robots.txt"))
	mux.Handle("/BingSiteAuth.xml", staticServer.FileHandler("BingSiteAuth.xml"))
	mux.Handle("/C", http.RedirectHandler("http://golang.org/doc/articles/c_go_cgo.html", http.StatusMovedPermanently))
	mux.Handle("/code.jquery.com/", http.NotFoundHandler())
	mux.Handle("/_ah/health", http.HandlerFunc(serveHealthCheck))
	mux.Handle("/_ah/", http.NotFoundHandler())
	mux.Handle("/", handler(serveHome))

	cacheBusters.Handler = mux

	var root http.Handler = rootHandler{
		{"api.", apiMux},
		{"talks.godoc.org", otherDomainHandler{"https", "go-talks.appspot.com"}},
		{"www.godoc.org", otherDomainHandler{"https", "godoc.org"}},
		{"", mux},
	}
	if gceLogName != "" {
		ctx := context.Background()

		logc, err := logging.NewClient(ctx, projID)
		if err != nil {
			log.Fatalf("Failed to create cloud logging client: %v", err)
		}
		logger := logc.Logger(gceLogName)

		if err := logc.Ping(ctx); err != nil {
			log.Fatalf("Failed to ping Google Cloud Logging: %v", err)
		}

		gceLogger = newGCELogger(logger)
	}

	http.Handle("/", root)
	appengine.Main()
}
