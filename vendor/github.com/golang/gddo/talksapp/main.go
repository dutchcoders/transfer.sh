// Copyright 2013 The Go Authors. All rights reserved.
//
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file or at
// https://developers.google.com/open-source/licenses/bsd.

// Package talksapp implements the go-talks.appspot.com server.
package talksapp

import (
	"bytes"
	"errors"
	"fmt"
	"html/template"
	"io"
	"net/http"
	"os"
	"path"
	"time"

	"google.golang.org/appengine"
	"google.golang.org/appengine/log"
	"google.golang.org/appengine/memcache"
	"google.golang.org/appengine/urlfetch"

	"github.com/golang/gddo/gosrc"
	"github.com/golang/gddo/httputil"

	"golang.org/x/net/context"
	"golang.org/x/tools/present"
)

var (
	presentTemplates = map[string]*template.Template{
		".article": parsePresentTemplate("article.tmpl"),
		".slide":   parsePresentTemplate("slides.tmpl"),
	}
	homeArticle  = loadHomeArticle()
	contactEmail = "golang-dev@googlegroups.com"

	// used for mocking in tests
	getPresentation = gosrc.GetPresentation
	playCompileURL  = "https://play.golang.org/compile"
)

func init() {
	http.Handle("/", handlerFunc(serveRoot))
	http.Handle("/compile", handlerFunc(serveCompile))
	http.Handle("/bot.html", handlerFunc(serveBot))
	present.PlayEnabled = true
	if s := os.Getenv("CONTACT_EMAIL"); s != "" {
		contactEmail = s
	}

	if appengine.IsDevAppServer() {
		return
	}
	github := httputil.NewAuthTransportFromEnvironment(nil)
	if github.Token == "" || github.ClientID == "" || github.ClientSecret == "" {
		panic("missing GitHub metadata, follow the instructions on README.md")
	}
}

func playable(c present.Code) bool {
	return present.PlayEnabled && c.Play && c.Ext == ".go"
}

func parsePresentTemplate(name string) *template.Template {
	t := present.Template()
	t = t.Funcs(template.FuncMap{"playable": playable})
	if _, err := t.ParseFiles("present/templates/"+name, "present/templates/action.tmpl"); err != nil {
		panic(err)
	}
	t = t.Lookup("root")
	if t == nil {
		panic("root template not found for " + name)
	}
	return t
}

func loadHomeArticle() []byte {
	const fname = "assets/home.article"
	f, err := os.Open(fname)
	if err != nil {
		panic(err)
	}
	defer f.Close()
	doc, err := present.Parse(f, fname, 0)
	if err != nil {
		panic(err)
	}
	var buf bytes.Buffer
	if err := renderPresentation(&buf, fname, doc); err != nil {
		panic(err)
	}
	return buf.Bytes()
}

func renderPresentation(w io.Writer, fname string, doc *present.Doc) error {
	t := presentTemplates[path.Ext(fname)]
	if t == nil {
		return errors.New("unknown template extension")
	}
	data := struct {
		*present.Doc
		Template     *template.Template
		PlayEnabled  bool
		NotesEnabled bool
	}{doc, t, true, true}
	return t.Execute(w, &data)
}

type presFileNotFoundError string

func (s presFileNotFoundError) Error() string { return fmt.Sprintf("File %s not found.", string(s)) }

func writeHTMLHeader(w http.ResponseWriter, status int) {
	w.Header().Set("Content-Type", "text/html; charset=utf-8")
	w.WriteHeader(status)
}

func writeTextHeader(w http.ResponseWriter, status int) {
	w.Header().Set("Content-Type", "text/plain; charset=utf-8")
	w.WriteHeader(status)
}

func httpClient(r *http.Request) *http.Client {
	ctx, _ := context.WithTimeout(appengine.NewContext(r), 10*time.Second)
	github := httputil.NewAuthTransportFromEnvironment(nil)

	return &http.Client{
		Transport: &httputil.AuthTransport{
			Token:        github.Token,
			ClientID:     github.ClientID,
			ClientSecret: github.ClientSecret,
			Base:         &urlfetch.Transport{Context: ctx},
			UserAgent:    fmt.Sprintf("%s (+http://%s/-/bot)", appengine.AppID(ctx), r.Host),
		},
	}
}

type handlerFunc func(http.ResponseWriter, *http.Request) error

func (f handlerFunc) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	ctx := appengine.NewContext(r)

	err := f(w, r)
	if err == nil {
		return
	} else if gosrc.IsNotFound(err) {
		writeTextHeader(w, 400)
		io.WriteString(w, "Not Found.")
	} else if e, ok := err.(*gosrc.RemoteError); ok {
		writeTextHeader(w, 500)
		fmt.Fprintf(w, "Error accessing %s.\n%v", e.Host, e)
		log.Infof(ctx, "Remote error %s: %v", e.Host, e)
	} else if e, ok := err.(presFileNotFoundError); ok {
		writeTextHeader(w, 200)
		io.WriteString(w, e.Error())
	} else if err != nil {
		writeTextHeader(w, 500)
		io.WriteString(w, "Internal server error.")
		log.Errorf(ctx, "Internal error %v", err)
	}
}

func serveRoot(w http.ResponseWriter, r *http.Request) error {
	switch {
	case r.Method != "GET" && r.Method != "HEAD":
		writeTextHeader(w, 405)
		_, err := io.WriteString(w, "Method not supported.")
		return err
	case r.URL.Path == "/":
		writeHTMLHeader(w, 200)
		_, err := w.Write(homeArticle)
		return err
	default:
		return servePresentation(w, r)
	}
}

func servePresentation(w http.ResponseWriter, r *http.Request) error {
	ctx := appengine.NewContext(r)
	importPath := r.URL.Path[1:]

	item, err := memcache.Get(ctx, importPath)
	if err == nil {
		writeHTMLHeader(w, 200)
		w.Write(item.Value)
		return nil
	} else if err != memcache.ErrCacheMiss {
		log.Errorf(ctx, "Could not get item from Memcache: %v", err)
	}

	log.Infof(ctx, "Fetching presentation %s.", importPath)
	pres, err := getPresentation(httpClient(r), importPath)
	if err != nil {
		return err
	}
	parser := &present.Context{
		ReadFile: func(name string) ([]byte, error) {
			if p, ok := pres.Files[name]; ok {
				return p, nil
			}
			return nil, presFileNotFoundError(name)
		},
	}
	doc, err := parser.Parse(bytes.NewReader(pres.Files[pres.Filename]), pres.Filename, 0)
	if err != nil {
		return err
	}

	var buf bytes.Buffer
	if err := renderPresentation(&buf, importPath, doc); err != nil {
		return err
	}

	if err := memcache.Add(ctx, &memcache.Item{
		Key:        importPath,
		Value:      buf.Bytes(),
		Expiration: time.Hour,
	}); err != nil {
		log.Errorf(ctx, "Could not cache presentation %s: %v", importPath, err)
		return nil
	}

	writeHTMLHeader(w, 200)
	_, err = w.Write(buf.Bytes())
	return err
}

func serveCompile(w http.ResponseWriter, r *http.Request) error {
	client := urlfetch.Client(appengine.NewContext(r))
	if err := r.ParseForm(); err != nil {
		return err
	}
	resp, err := client.PostForm(playCompileURL, r.Form)
	if err != nil {
		return err
	}
	defer resp.Body.Close()
	w.Header().Set("Content-Type", resp.Header.Get("Content-Type"))
	_, err = io.Copy(w, resp.Body)
	return err
}

func serveBot(w http.ResponseWriter, r *http.Request) error {
	ctx := appengine.NewContext(r)
	writeTextHeader(w, 200)
	_, err := fmt.Fprintf(w, "Contact %s for help with the %s bot.", contactEmail, appengine.AppID(ctx))
	return err
}
