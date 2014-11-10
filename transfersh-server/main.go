/*
The MIT License (MIT)

Copyright (c) 2014 DutchCoders [https://github.com/dutchcoders/]

Permission is hereby granted, free of charge, to any person obtaining a copy
of this software and associated documentation files (the "Software"), to deal
in the Software without restriction, including without limitation the rights
to use, copy, modify, merge, publish, distribute, sublicense, and/or sell
copies of the Software, and to permit persons to whom the Software is
furnished to do so, subject to the following conditions:

The above copyright notice and this permission notice shall be included in
all copies or substantial portions of the Software.

THE SOFTWARE IS PROVIDED "AS IS", WITHOUT WARRANTY OF ANY KIND, EXPRESS OR
IMPLIED, INCLUDING BUT NOT LIMITED TO THE WARRANTIES OF MERCHANTABILITY,
FITNESS FOR A PARTICULAR PURPOSE AND NONINFRINGEMENT. IN NO EVENT SHALL THE
AUTHORS OR COPYRIGHT HOLDERS BE LIABLE FOR ANY CLAIM, DAMAGES OR OTHER
LIABILITY, WHETHER IN AN ACTION OF CONTRACT, TORT OR OTHERWISE, ARISING FROM,
OUT OF OR IN CONNECTION WITH THE SOFTWARE OR THE USE OR OTHER DEALINGS IN
THE SOFTWARE.
*/

package main

import (
	// _ "transfer.sh/app/handlers"
	// _ "transfer.sh/app/utils"
	"flag"
	"fmt"
	"github.com/PuerkitoBio/ghost/handlers"
	"github.com/gorilla/mux"
	"log"
	"math/rand"
	"mime"
	"net/http"
	"net/url"
	"os"
	"time"
)

const SERVER_INFO = "transfer.sh"

// parse request with maximum memory of _24Kilobits
const _24K = (1 << 20) * 24

var config struct {
	AWS_ACCESS_KEY string
	AWS_SECRET_KEY string
	BUCKET         string
	VIRUSTOTAL_KEY string
	Temp           string
}

var storage Storage

func init() {
	config.AWS_ACCESS_KEY = os.Getenv("AWS_ACCESS_KEY")
	config.AWS_SECRET_KEY = os.Getenv("AWS_SECRET_KEY")
	config.BUCKET = os.Getenv("BUCKET")
	config.VIRUSTOTAL_KEY = os.Getenv("VIRUSTOTAL_KEY")
	config.Temp = ""
}

func main() {
	rand.Seed(time.Now().UTC().UnixNano())

	r := mux.NewRouter()

	r.PathPrefix("/scripts/").Handler(http.FileServer(http.Dir("./static/")))
	r.PathPrefix("/styles/").Handler(http.FileServer(http.Dir("./static/")))
	r.PathPrefix("/images/").Handler(http.FileServer(http.Dir("./static/")))
	r.PathPrefix("/fonts/").Handler(http.FileServer(http.Dir("./static/")))
	r.PathPrefix("/ico/").Handler(http.FileServer(http.Dir("./static/")))
	r.PathPrefix("/favicon.ico").Handler(http.FileServer(http.Dir("./static/")))
	r.PathPrefix("/robots.txt").Handler(http.FileServer(http.Dir("./static/")))

	r.HandleFunc("/({files:.*}).zip", zipHandler).Methods("GET")
	r.HandleFunc("/({files:.*}).tar", tarHandler).Methods("GET")
	r.HandleFunc("/({files:.*}).tar.gz", tarGzHandler).Methods("GET")
	r.HandleFunc("/download/{token}/{filename}", getHandler).Methods("GET")

	r.HandleFunc("/{token}/{filename}", previewHandler).MatcherFunc(func(r *http.Request, rm *mux.RouteMatch) bool {
		if !acceptsHtml(r.Header) {
			return false
		}

		match := (r.Referer() == "")

		u, err := url.Parse(r.Referer())
		if err != nil {
			log.Fatal(err)
			return false
		}

		match = match || (u.Host == "transfer.sh")

		match = match || (u.Host == "127.0.0.1")

		return match
	}).Methods("GET")

	r.HandleFunc("/{token}/{filename}", getHandler).Methods("GET")
	r.HandleFunc("/get/{token}/{filename}", getHandler).Methods("GET")
	r.HandleFunc("/{filename}/virustotal", virusTotalHandler).Methods("PUT")
	r.HandleFunc("/{filename}/scan", scanHandler).Methods("PUT")
	r.HandleFunc("/put/{filename}", putHandler).Methods("PUT")
	r.HandleFunc("/upload/{filename}", putHandler).Methods("PUT")
	r.HandleFunc("/{filename}", putHandler).Methods("PUT")
	r.HandleFunc("/health.html", healthHandler).Methods("GET")
	r.HandleFunc("/", postHandler).Methods("POST")
	// r.HandleFunc("/{page}", viewHandler).Methods("GET")
	r.HandleFunc("/", viewHandler).Methods("GET")

	r.NotFoundHandler = http.HandlerFunc(notFoundHandler)

	port := flag.String("port", "8080", "port number, default: 8080")
	temp := flag.String("temp", "", "")
	basedir := flag.String("basedir", "", "")
	logpath := flag.String("log", "", "")
	provider := flag.String("provider", "s3", "")

	flag.Parse()

	if *logpath != "" {
		f, err := os.OpenFile(*logpath, os.O_RDWR|os.O_CREATE|os.O_APPEND, 0666)
		if err != nil {
			log.Fatalf("error opening file: %v", err)
		}

		defer f.Close()

		log.SetOutput(f)
	}

	config.Temp = *temp

	var err error

	switch *provider {
	case "s3":
		storage, err = NewS3Storage()
	case "local":
		if *basedir == "" {
			log.Panic("basedir not set")
		}

		storage, err = NewLocalStorage(*basedir)
	}

	if err != nil {
		log.Panic("Error while creating storage.")
	}

	mime.AddExtensionType(".md", "text/x-markdown")

	log.Printf("Transfer.sh server started. :%v using temp folder: %s", *port, config.Temp)
	log.Printf("---------------------------")

	s := &http.Server{
		Addr:    fmt.Sprintf(":%s", *port),
		Handler: handlers.PanicHandler(LoveHandler(RedirectHandler(handlers.LogHandler(r, handlers.NewLogOptions(log.Printf, "_default_")))), nil),
	}

	log.Panic(s.ListenAndServe())
	log.Printf("Server stopped.")
}
