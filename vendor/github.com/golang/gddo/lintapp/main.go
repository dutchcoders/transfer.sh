// Copyright 2013 The Go Authors. All rights reserved.
//
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file or at
// https://developers.google.com/open-source/licenses/bsd.

// Package lintapp implements the go-lint.appspot.com server.
package lintapp

import (
	"bytes"
	"encoding/gob"
	"fmt"
	"html/template"
	"net/http"
	"os"
	"path/filepath"
	"strconv"
	"strings"
	"time"

	"golang.org/x/net/context"
	"google.golang.org/appengine"
	"google.golang.org/appengine/datastore"
	"google.golang.org/appengine/log"
	"google.golang.org/appengine/urlfetch"

	"github.com/golang/gddo/gosrc"
	"github.com/golang/gddo/httputil"

	"github.com/golang/lint"
)

func init() {
	http.Handle("/", handlerFunc(serveRoot))
	http.Handle("/-/bot", handlerFunc(serveBot))
	http.Handle("/-/refresh", handlerFunc(serveRefresh))
	if s := os.Getenv("CONTACT_EMAIL"); s != "" {
		contactEmail = s
	}
}

var (
	contactEmail    = "golang-dev@googlegroups.com"
	homeTemplate    = parseTemplate("common.html", "index.html")
	packageTemplate = parseTemplate("common.html", "package.html")
	errorTemplate   = parseTemplate("common.html", "error.html")
	templateFuncs   = template.FuncMap{
		"timeago":      timeagoFn,
		"contactEmail": contactEmailFn,
	}
	github = httputil.NewAuthTransportFromEnvironment(nil)
)

func parseTemplate(fnames ...string) *template.Template {
	paths := make([]string, len(fnames))
	for i := range fnames {
		paths[i] = filepath.Join("assets/templates", fnames[i])
	}
	t, err := template.New("").Funcs(templateFuncs).ParseFiles(paths...)
	if err != nil {
		panic(err)
	}
	t = t.Lookup("ROOT")
	if t == nil {
		panic(fmt.Sprintf("ROOT template not found in %v", fnames))
	}
	return t
}

func contactEmailFn() string {
	return contactEmail
}

func timeagoFn(t time.Time) string {
	d := time.Since(t)
	switch {
	case d < time.Second:
		return "just now"
	case d < 2*time.Second:
		return "one second ago"
	case d < time.Minute:
		return fmt.Sprintf("%d seconds ago", d/time.Second)
	case d < 2*time.Minute:
		return "one minute ago"
	case d < time.Hour:
		return fmt.Sprintf("%d minutes ago", d/time.Minute)
	case d < 2*time.Hour:
		return "one hour ago"
	case d < 48*time.Hour:
		return fmt.Sprintf("%d hours ago", d/time.Hour)
	default:
		return fmt.Sprintf("%d days ago", d/(time.Hour*24))
	}
}

func writeResponse(w http.ResponseWriter, status int, t *template.Template, v interface{}) error {
	var buf bytes.Buffer
	if err := t.Execute(&buf, v); err != nil {
		return err
	}
	w.Header().Set("Content-Type", "text/html; charset=utf-8")
	w.Header().Set("Content-Length", strconv.Itoa(buf.Len()))
	w.WriteHeader(status)
	_, err := w.Write(buf.Bytes())
	return err
}

func writeErrorResponse(w http.ResponseWriter, status int) error {
	return writeResponse(w, status, errorTemplate, http.StatusText(status))
}

func httpClient(r *http.Request) *http.Client {
	c := appengine.NewContext(r)
	return &http.Client{
		Transport: &httputil.Transport{
			Token:        github.Token,
			ClientID:     github.ClientID,
			ClientSecret: github.ClientSecret,
			Base:         &urlfetch.Transport{Context: c, Deadline: 10 * time.Second},
			UserAgent:    fmt.Sprintf("%s (+http://%s/-/bot)", appengine.AppID(c), r.Host),
		},
	}
}

const version = 1

type storePackage struct {
	Data    []byte
	Version int
}

type lintPackage struct {
	Files   []*lintFile
	Path    string
	Updated time.Time
	LineFmt string
	URL     string
}

type lintFile struct {
	Name     string
	Problems []*lintProblem
	URL      string
}

type lintProblem struct {
	Line       int
	Text       string
	LineText   string
	Confidence float64
	Link       string
}

func putPackage(c context.Context, importPath string, pkg *lintPackage) error {
	var buf bytes.Buffer
	if err := gob.NewEncoder(&buf).Encode(pkg); err != nil {
		return err
	}
	_, err := datastore.Put(c,
		datastore.NewKey(c, "Package", importPath, 0, nil),
		&storePackage{Data: buf.Bytes(), Version: version})
	return err
}

func getPackage(c context.Context, importPath string) (*lintPackage, error) {
	var spkg storePackage
	if err := datastore.Get(c, datastore.NewKey(c, "Package", importPath, 0, nil), &spkg); err != nil {
		if err == datastore.ErrNoSuchEntity {
			err = nil
		}
		return nil, err
	}
	if spkg.Version != version {
		return nil, nil
	}
	var pkg lintPackage
	if err := gob.NewDecoder(bytes.NewReader(spkg.Data)).Decode(&pkg); err != nil {
		return nil, err
	}
	return &pkg, nil
}

func runLint(r *http.Request, importPath string) (*lintPackage, error) {
	dir, err := gosrc.Get(httpClient(r), importPath, "")
	if err != nil {
		return nil, err
	}

	pkg := lintPackage{
		Path:    importPath,
		Updated: time.Now(),
		LineFmt: dir.LineFmt,
		URL:     dir.BrowseURL,
	}
	linter := lint.Linter{}
	for _, f := range dir.Files {
		if !strings.HasSuffix(f.Name, ".go") {
			continue
		}
		problems, err := linter.Lint(f.Name, f.Data)
		if err == nil && len(problems) == 0 {
			continue
		}
		file := lintFile{Name: f.Name, URL: f.BrowseURL}
		if err != nil {
			file.Problems = []*lintProblem{{Text: err.Error()}}
		} else {
			for _, p := range problems {
				file.Problems = append(file.Problems, &lintProblem{
					Line:       p.Position.Line,
					Text:       p.Text,
					LineText:   p.LineText,
					Confidence: p.Confidence,
					Link:       p.Link,
				})
			}
		}
		if len(file.Problems) > 0 {
			pkg.Files = append(pkg.Files, &file)
		}
	}

	if err := putPackage(appengine.NewContext(r), importPath, &pkg); err != nil {
		return nil, err
	}

	return &pkg, nil
}

func filterByConfidence(r *http.Request, pkg *lintPackage) {
	minConfidence, err := strconv.ParseFloat(r.FormValue("minConfidence"), 64)
	if err != nil {
		minConfidence = 0.8
	}
	for _, f := range pkg.Files {
		j := 0
		for i := range f.Problems {
			if f.Problems[i].Confidence >= minConfidence {
				f.Problems[j] = f.Problems[i]
				j++
			}
		}
		f.Problems = f.Problems[:j]
	}
}

type handlerFunc func(http.ResponseWriter, *http.Request) error

func (f handlerFunc) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	c := appengine.NewContext(r)
	err := f(w, r)
	if err == nil {
		return
	} else if gosrc.IsNotFound(err) {
		writeErrorResponse(w, 404)
	} else if e, ok := err.(*gosrc.RemoteError); ok {
		log.Infof(c, "Remote error %s: %v", e.Host, e)
		writeResponse(w, 500, errorTemplate, fmt.Sprintf("Error accessing %s.", e.Host))
	} else if err != nil {
		log.Errorf(c, "Internal error %v", err)
		writeErrorResponse(w, 500)
	}
}

func serveRoot(w http.ResponseWriter, r *http.Request) error {
	switch {
	case r.Method != "GET" && r.Method != "HEAD":
		return writeErrorResponse(w, 405)
	case r.URL.Path == "/":
		return writeResponse(w, 200, homeTemplate, nil)
	default:
		importPath := r.URL.Path[1:]
		if !gosrc.IsValidPath(importPath) {
			return gosrc.NotFoundError{Message: "bad path"}
		}
		c := appengine.NewContext(r)
		pkg, err := getPackage(c, importPath)
		if pkg == nil && err == nil {
			pkg, err = runLint(r, importPath)
		}
		if err != nil {
			return err
		}
		filterByConfidence(r, pkg)
		return writeResponse(w, 200, packageTemplate, pkg)
	}
}

func serveRefresh(w http.ResponseWriter, r *http.Request) error {
	if r.Method != "POST" {
		return writeErrorResponse(w, 405)
	}
	importPath := r.FormValue("importPath")
	pkg, err := runLint(r, importPath)
	if err != nil {
		return err
	}
	http.Redirect(w, r, "/"+pkg.Path, 301)
	return nil
}

func serveBot(w http.ResponseWriter, r *http.Request) error {
	c := appengine.NewContext(r)
	_, err := fmt.Fprintf(w, "Contact %s for help with the %s bot.", contactEmail, appengine.AppID(c))
	return err
}
