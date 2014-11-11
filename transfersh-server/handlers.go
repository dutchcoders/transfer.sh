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

	"archive/tar"
	"archive/zip"
	"bytes"
	"compress/gzip"
	"errors"
	"fmt"
	"github.com/dutchcoders/go-clamd"
	"github.com/golang/gddo/httputil/header"
	"github.com/gorilla/mux"
	"github.com/kennygrant/sanitize"
	html_template "html/template"
	"io"
	"io/ioutil"
	"log"
	"math/rand"
	"mime"
	"net/http"
	"os"
	"path/filepath"
	"strconv"
	"strings"
	text_template "text/template"
	"time"
)

func healthHandler(w http.ResponseWriter, r *http.Request) {
	fmt.Fprintf(w, "Approaching Neutral Zone, all systems normal and functioning.")
}

// this handler will output html or text, depending on the
// support of the client (Accept header).

func viewHandler(w http.ResponseWriter, r *http.Request) {
	// vars := mux.Vars(r)

	actual := header.ParseAccept(r.Header, "Accept")

	html := false

	for _, s := range actual {
		if s.Value == "text/html" {
			html = true
		}
	}

	if html {
		tmpl, err := html_template.ParseFiles("static/index.html")

		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}

		if err := tmpl.Execute(w, nil); err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}
	} else {
		tmpl, err := text_template.ParseFiles("static/index.txt")

		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}

		if err := tmpl.Execute(w, nil); err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}
	}
}

func notFoundHandler(w http.ResponseWriter, r *http.Request) {
	http.Error(w, http.StatusText(404), 404)
}

func postHandler(w http.ResponseWriter, r *http.Request) {
	if err := r.ParseMultipartForm(_24K); nil != err {
		log.Println(err)
		http.Error(w, "Error occured copying to output stream", 500)
		return
	}

	token := Encode(10000000 + int64(rand.Intn(1000000000)))

	w.Header().Set("Content-Type", "text/plain")

	for _, fheaders := range r.MultipartForm.File {
		for _, fheader := range fheaders {
			filename := sanitize.Path(filepath.Base(fheader.Filename))
			contentType := fheader.Header.Get("Content-Type")

			if contentType == "" {
				contentType = mime.TypeByExtension(filepath.Ext(fheader.Filename))
			}

			var f io.Reader
			var err error

			if f, err = fheader.Open(); err != nil {
				log.Print(err)
				http.Error(w, err.Error(), 500)
				return
			}

			var b bytes.Buffer

			n, err := io.CopyN(&b, f, _24K+1)
			if err != nil && err != io.EOF {
				log.Print(err)
				http.Error(w, err.Error(), 500)
				return
			}

			var reader io.Reader

			if n > _24K {
				file, err := ioutil.TempFile(config.Temp, "transfer-")
				if err != nil {
					log.Fatal(err)
				}
				defer file.Close()

				n, err = io.Copy(file, io.MultiReader(&b, f))
				if err != nil {
					os.Remove(file.Name())

					log.Print(err)
					http.Error(w, err.Error(), 500)
					return
				}

				reader, err = os.Open(file.Name())
			} else {
				reader = bytes.NewReader(b.Bytes())
			}

			contentLength := n

			log.Printf("Uploading %s %s %d %s", token, filename, contentLength, contentType)

			if err = storage.Put(token, filename, reader, contentType, uint64(contentLength)); err != nil {
				log.Print(err)
				http.Error(w, err.Error(), 500)
				return

			}

			fmt.Fprintf(w, "https://%s/%s/%s\n", ipAddrFromRemoteAddr(r.Host), token, filename)
		}
	}
}

func scanHandler(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)

	filename := sanitize.Path(filepath.Base(vars["filename"]))

	contentLength := r.ContentLength
	contentType := r.Header.Get("Content-Type")

	log.Printf("Scanning %s %d %s", filename, contentLength, contentType)

	var reader io.Reader

	reader = r.Body

	c := clamd.NewClamd("/tmp/clamd.socket")

	response, err := c.ScanStream(reader)
	if err != nil {
		http.Error(w, err.Error(), 500)
	}

	var b string

	for s := range response {
		b += s

		if !strings.HasPrefix(s, "stream: ") {
			continue
		}

		w.Write([]byte(fmt.Sprintf("%v\n", s[8:])))
	}
}

func putHandler(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)

	filename := sanitize.Path(filepath.Base(vars["filename"]))

	contentLength := r.ContentLength

	var reader io.Reader

	reader = r.Body

	if contentLength == -1 {
		// queue file to disk, because s3 needs content length
		var err error
		var f io.Reader

		f = reader

		var b bytes.Buffer

		n, err := io.CopyN(&b, f, _24K+1)
		if err != nil && err != io.EOF {
			log.Print(err)
			http.Error(w, err.Error(), 500)
			return
		}

		if n > _24K {
			file, err := ioutil.TempFile(config.Temp, "transfer-")
			if err != nil {
				log.Print(err)
				http.Error(w, err.Error(), 500)
				return
			}

			defer file.Close()

			n, err = io.Copy(file, io.MultiReader(&b, f))
			if err != nil {
				os.Remove(file.Name())

				log.Print(err)
				http.Error(w, err.Error(), 500)
				return
			}

			reader, err = os.Open(file.Name())
		} else {
			reader = bytes.NewReader(b.Bytes())
		}

		contentLength = n
	}

	contentType := r.Header.Get("Content-Type")

	if contentType == "" {
		contentType = mime.TypeByExtension(filepath.Ext(vars["filename"]))
	}

	token := Encode(10000000 + int64(rand.Intn(1000000000)))

	log.Printf("Uploading %s %d %s", token, filename, contentLength, contentType)

	var err error

	if err = storage.Put(token, filename, reader, contentType, uint64(contentLength)); err != nil {
		http.Error(w, errors.New("Could not save file").Error(), 500)
		return
	}

	// w.Statuscode = 200

	w.Header().Set("Content-Type", "text/plain")

	fmt.Fprintf(w, "https://%s/%s/%s\n", ipAddrFromRemoteAddr(r.Host), token, filename)
}

func zipHandler(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)

	files := vars["files"]

	zipfilename := fmt.Sprintf("transfersh-%d.zip", uint16(time.Now().UnixNano()))

	w.Header().Set("Content-Type", "application/zip")
	w.Header().Set("Content-Disposition", fmt.Sprintf("attachment; filename=\"%s\"", zipfilename))
	w.Header().Set("Connection", "close")

	zw := zip.NewWriter(w)

	for _, key := range strings.Split(files, ",") {
		token := sanitize.Path(strings.Split(key, "/")[0])
		filename := sanitize.Path(strings.Split(key, "/")[1])

		reader, _, _, err := storage.Get(token, filename)
		if err != nil {
			if err.Error() == "The specified key does not exist." {
				http.Error(w, "File not found", 404)
				return
			} else {
				log.Printf("%s", err.Error())
				http.Error(w, "Could not retrieve file.", 500)
				return
			}
		}

		defer reader.Close()

		header := &zip.FileHeader{
			Name:         strings.Split(key, "/")[1],
			Method:       zip.Store,
			ModifiedTime: uint16(time.Now().UnixNano()),
			ModifiedDate: uint16(time.Now().UnixNano()),
		}

		fw, err := zw.CreateHeader(header)

		if err != nil {
			log.Printf("%s", err.Error())
			http.Error(w, "Internal server error.", 500)
			return
		}

		if _, err = io.Copy(fw, reader); err != nil {
			log.Printf("%s", err.Error())
			http.Error(w, "Internal server error.", 500)
			return
		}
	}

	if err := zw.Close(); err != nil {
		log.Printf("%s", err.Error())
		http.Error(w, "Internal server error.", 500)
		return
	}
}

func tarGzHandler(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)

	files := vars["files"]

	tarfilename := fmt.Sprintf("transfersh-%d.tar.gz", uint16(time.Now().UnixNano()))

	w.Header().Set("Content-Type", "application/x-gzip")
	w.Header().Set("Content-Disposition", fmt.Sprintf("attachment; filename=\"%s\"", tarfilename))
	w.Header().Set("Connection", "close")

	os := gzip.NewWriter(w)
	defer os.Close()

	zw := tar.NewWriter(os)
	defer zw.Close()

	for _, key := range strings.Split(files, ",") {
		token := strings.Split(key, "/")[0]
		filename := strings.Split(key, "/")[1]

		reader, _, contentLength, err := storage.Get(token, filename)
		if err != nil {
			if err.Error() == "The specified key does not exist." {
				http.Error(w, "File not found", 404)
				return
			} else {
				log.Printf("%s", err.Error())
				http.Error(w, "Could not retrieve file.", 500)
				return
			}
		}

		defer reader.Close()

		header := &tar.Header{
			Name: strings.Split(key, "/")[1],
			Size: int64(contentLength),
		}

		err = zw.WriteHeader(header)
		if err != nil {
			log.Printf("%s", err.Error())
			http.Error(w, "Internal server error.", 500)
			return
		}

		if _, err = io.Copy(zw, reader); err != nil {
			log.Printf("%s", err.Error())
			http.Error(w, "Internal server error.", 500)
			return
		}
	}
}

func tarHandler(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)

	files := vars["files"]

	tarfilename := fmt.Sprintf("transfersh-%d.tar", uint16(time.Now().UnixNano()))

	w.Header().Set("Content-Type", "application/x-tar")
	w.Header().Set("Content-Disposition", fmt.Sprintf("attachment; filename=\"%s\"", tarfilename))
	w.Header().Set("Connection", "close")

	zw := tar.NewWriter(w)
	defer zw.Close()

	for _, key := range strings.Split(files, ",") {
		token := strings.Split(key, "/")[0]
		filename := strings.Split(key, "/")[1]

		reader, _, contentLength, err := storage.Get(token, filename)
		if err != nil {
			if err.Error() == "The specified key does not exist." {
				http.Error(w, "File not found", 404)
				return
			} else {
				log.Printf("%s", err.Error())
				http.Error(w, "Could not retrieve file.", 500)
				return
			}
		}

		defer reader.Close()

		header := &tar.Header{
			Name: strings.Split(key, "/")[1],
			Size: int64(contentLength),
		}

		err = zw.WriteHeader(header)
		if err != nil {
			log.Printf("%s", err.Error())
			http.Error(w, "Internal server error.", 500)
			return
		}

		if _, err = io.Copy(zw, reader); err != nil {
			log.Printf("%s", err.Error())
			http.Error(w, "Internal server error.", 500)
			return
		}
	}
}

func getHandler(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)

	token := vars["token"]
	filename := vars["filename"]

	reader, contentType, contentLength, err := storage.Get(token, filename)
	if err != nil {
		if err.Error() == "The specified key does not exist." {
			http.Error(w, "File not found", 404)
			return
		} else {
			log.Printf("%s", err.Error())
			http.Error(w, "Could not retrieve file.", 500)
			return
		}
	}

	defer reader.Close()

	w.Header().Set("Content-Type", contentType)
	w.Header().Set("Content-Length", strconv.FormatUint(contentLength, 10))

	mediaType, _, _ := mime.ParseMediaType(contentType)

	switch {
	case mediaType == "text/html":
		w.Header().Set("Content-Disposition", fmt.Sprintf("attachment; filename=\"%s\"", filename))
		break
	case strings.HasPrefix(mediaType, "text"):
	case mediaType == "":
		break
	default:
		w.Header().Set("Content-Disposition", fmt.Sprintf("attachment; filename=\"%s\"", filename))
	}

	w.Header().Set("Connection", "close")

	if _, err = io.Copy(w, reader); err != nil {
		http.Error(w, "Error occured copying to output stream", 500)
		return
	}
}

func RedirectHandler(h http.Handler) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		if ipAddrFromRemoteAddr(r.Host) != "transfer.sh" && ipAddrFromRemoteAddr(r.Host) != "127.0.0.1" && r.URL.Path != "/health.html" {
			http.Redirect(w, r, "https://transfer.sh/", 301)
			return
		}

		if ipAddrFromRemoteAddr(r.Host) == "transfer.sh" && r.Header.Get("X-Forwarded-Proto") != "https" && r.Method == "GET" {
			http.Redirect(w, r, "https://transfer.sh/", 301)
			return
		}

		h.ServeHTTP(w, r)
	}
}

// Create a log handler for every request it receives.
func LoveHandler(h http.Handler) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("x-made-with", "<3 by DutchCoders")
		w.Header().Set("x-served-by", "Proudly served by DutchCoders")
		w.Header().Set("Server", "Transfer.sh HTTP Server 1.0")
		h.ServeHTTP(w, r)
	}
}
