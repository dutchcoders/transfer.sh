/*
The MIT License (MIT)

Copyright (c) 2014-2017 DutchCoders [https://github.com/dutchcoders/]

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

package server

import (
	// _ "transfer.sh/app/handlers"
	// _ "transfer.sh/app/utils"

	"archive/tar"
	"archive/zip"
	"bytes"
	"compress/gzip"
	"encoding/base64"
	"errors"
	"fmt"
	"html"
	html_template "html/template"
	"io"
	"io/ioutil"
	"log"
	"math/rand"
	"mime"
	"net"
	"net/http"
	"net/url"
	"os"
	"path"
	"path/filepath"
	"strconv"
	"strings"
	"sync"
	text_template "text/template"
	"time"

	web "github.com/dutchcoders/transfer.sh-web"
	"github.com/dutchcoders/transfer.sh/server/storage"
	"github.com/gorilla/mux"
	"github.com/microcosm-cc/bluemonday"
	blackfriday "github.com/russross/blackfriday/v2"
	"github.com/skip2/go-qrcode"
)

const getPathPart = "get"

var (
	htmlTemplates = initHTMLTemplates()
	textTemplates = initTextTemplates()
)

func stripPrefix(path string) string {
	return strings.Replace(path, web.Prefix+"/", "", -1)
}

func initTextTemplates() *text_template.Template {
	templateMap := text_template.FuncMap{"format": formatNumber}

	// Templates with functions available to them
	var templates = text_template.New("").Funcs(templateMap)
	return templates
}

func initHTMLTemplates() *html_template.Template {
	templateMap := html_template.FuncMap{"format": formatNumber}

	// Templates with functions available to them
	var templates = html_template.New("").Funcs(templateMap)

	return templates
}

func healthHandler(w http.ResponseWriter, r *http.Request) {
	_, _ = w.Write([]byte("Approaching Neutral Zone, all systems normal and functioning."))
}

/* The preview handler will show a preview of the content for browsers (accept type text/html), and referer is not transfer.sh */
func (s *Server) previewHandler(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)

	token := vars["token"]
	filename := vars["filename"]

	metadata, err := s.CheckMetadata(token, filename, false)

	if err != nil {
		log.Printf("Error metadata: %s", err.Error())
		http.Error(w, http.StatusText(http.StatusNotFound), http.StatusNotFound)
		return
	}

	contentType := metadata.ContentType

	var templatePath string
	var content html_template.HTML

	switch {
	case strings.HasPrefix(contentType, "image/"):
		templatePath = "download.image.html"
	case strings.HasPrefix(contentType, "video/"):
		templatePath = "download.video.html"
	case strings.HasPrefix(contentType, "audio/"):
		templatePath = "download.audio.html"
	case strings.HasPrefix(contentType, "text/"):
		templatePath = "download.markdown.html"

		var reader io.ReadCloser
		if reader, _, err = s.storage.Get(token, filename); err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}

		var data []byte
		data = make([]byte, _5M)
		if _, err = reader.Read(data); err != io.EOF && err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}

		if strings.HasPrefix(contentType, "text/x-markdown") || strings.HasPrefix(contentType, "text/markdown") {
			unsafe := blackfriday.Run(data)
			output := bluemonday.UGCPolicy().SanitizeBytes(unsafe)
			content = html_template.HTML(output)
		} else if strings.HasPrefix(contentType, "text/plain") {
			content = html_template.HTML(fmt.Sprintf("<pre>%s</pre>", html.EscapeString(string(data))))
		} else {
			templatePath = "download.sandbox.html"
		}

	default:
		templatePath = "download.html"
	}

	relativeURL, _ := url.Parse(path.Join(s.proxyPath, token, filename))
	resolvedURL := resolveURL(r, relativeURL)
	relativeURLGet, _ := url.Parse(path.Join(s.proxyPath, getPathPart, token, filename))
	resolvedURLGet := resolveURL(r, relativeURLGet)
	var png []byte
	png, err = qrcode.Encode(resolvedURL, qrcode.High, 150)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	qrCode := base64.StdEncoding.EncodeToString(png)

	hostname := getURL(r).Host
	webAddress := resolveWebAddress(r, s.proxyPath)

	data := struct {
		ContentType   string
		Content       html_template.HTML
		Filename      string
		Url           string
		UrlGet        string
		Hostname      string
		WebAddress    string
		ContentLength int64
		GAKey         string
		UserVoiceKey  string
		QRCode        string
	}{
		contentType,
		content,
		filename,
		resolvedURL,
		resolvedURLGet,
		hostname,
		webAddress,
		metadata.ContentLength,
		s.gaKey,
		s.userVoiceKey,
		qrCode,
	}

	if err := htmlTemplates.ExecuteTemplate(w, templatePath, data); err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

}

// this handler will output html or text, depending on the
// support of the client (Accept header).

func (s *Server) viewHandler(w http.ResponseWriter, r *http.Request) {
	// vars := mux.Vars(r)

	hostname := getURL(r).Host
	webAddress := resolveWebAddress(r, s.proxyPath)

	data := struct {
		Hostname     string
		WebAddress   string
		GAKey        string
		UserVoiceKey string
	}{
		hostname,
		webAddress,
		s.gaKey,
		s.userVoiceKey,
	}

	if acceptsHTML(r.Header) {
		if err := htmlTemplates.ExecuteTemplate(w, "index.html", data); err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}
	} else {
		if err := textTemplates.ExecuteTemplate(w, "index.txt", data); err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}
	}
}

func (s *Server) notFoundHandler(w http.ResponseWriter, r *http.Request) {
	http.Error(w, http.StatusText(404), 404)
}

func sanitize(fileName string) string {
	return path.Clean(path.Base(fileName))
}

func (s *Server) postHandler(w http.ResponseWriter, r *http.Request) {
	if err := r.ParseMultipartForm(_24K); nil != err {
		log.Printf("%s", err.Error())
		http.Error(w, "Error occurred copying to output stream", 500)
		return
	}

	token := Encode(10000000 + int64(rand.Intn(1000000000)))

	w.Header().Set("Content-Type", "text/plain")

	for _, fheaders := range r.MultipartForm.File {
		for _, fheader := range fheaders {
			filename := sanitize(fheader.Filename)
			contentType := fheader.Header.Get("Content-Type")

			if contentType == "" {
				contentType = mime.TypeByExtension(filepath.Ext(fheader.Filename))
			}

			var f io.Reader
			var err error

			if f, err = fheader.Open(); err != nil {
				log.Printf("%s", err.Error())
				http.Error(w, err.Error(), 500)
				return
			}

			var b bytes.Buffer

			n, err := io.CopyN(&b, f, _24K+1)
			if err != nil && err != io.EOF {
				log.Printf("%s", err.Error())
				http.Error(w, err.Error(), 500)
				return
			}

			var file *os.File
			var reader io.Reader

			if n > _24K {
				file, err = ioutil.TempFile(s.tempPath, "transfer-")
				if err != nil {
					log.Fatal(err)
				}

				defer cleanTmpFile(file)

				n, err = io.Copy(file, io.MultiReader(&b, f))
				if err != nil {
					log.Printf("%s", err.Error())
					http.Error(w, err.Error(), 500)
					return
				}

				reader, err = os.Open(file.Name())
			} else {
				reader = bytes.NewReader(b.Bytes())
			}

			metadata := s.metadataForRequest(contentType, n, r)

			log.Printf("Uploading %s %s %d %s", token, filename, metadata.ContentLength, metadata.ContentType)

			if err = s.storage.Put(token, filename, reader, metadata); err != nil {
				log.Printf("Backend storage error: %s", err.Error())
				http.Error(w, err.Error(), 500)
				return

			}

			filename = url.PathEscape(filename)
			relativeURL, _ := url.Parse(path.Join(s.proxyPath, token, filename))
			_, _ = fmt.Fprintln(w, getURL(r).ResolveReference(relativeURL).String())
		}
	}
}

func cleanTmpFile(f *os.File) {
	if f != nil {
		err := f.Close()
		if err != nil {
			log.Printf("Error closing tmpfile: %s (%s)", err, f.Name())
		}

		err = os.Remove(f.Name())
		if err != nil {
			log.Printf("Error removing tmpfile: %s (%s)", err, f.Name())
		}
	}
}

func (s *Server) metadataForRequest(contentType string, contentLength int64, r *http.Request) storage.Metadata {
	metadata := storage.Metadata{
		ContentType:   contentType,
		ContentLength: contentLength,
		MaxDate:       time.Now().Add(s.lifetime),
		Downloads:     0,
		MaxDownloads:  -1,
		DeletionToken: Encode(10000000+int64(rand.Intn(1000000000))) + Encode(10000000+int64(rand.Intn(1000000000))),
	}

	if v := r.Header.Get("Max-Downloads"); v == "" {
	} else if v, err := strconv.Atoi(v); err != nil {
	} else {
		metadata.MaxDownloads = v
	}

	if maxDays := r.Header.Get("Max-Days"); maxDays != "" {
		v, err := strconv.Atoi(maxDays)
		if err != nil {
			return metadata
		}
		metadata.MaxDate = time.Now().Add(time.Hour * 24 * time.Duration(v))
	}
	return metadata
}

func (s *Server) putHandler(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)

	filename := sanitize(vars["filename"])

	contentLength := r.ContentLength

	var reader io.Reader

	reader = r.Body

	defer r.Body.Close()

	if contentLength == -1 {
		// queue file to disk, because s3 needs content length
		var err error
		var f io.Reader

		f = reader

		var b bytes.Buffer

		n, err := io.CopyN(&b, f, _24K+1)
		if err != nil && err != io.EOF {
			log.Printf("Error putting new file: %s", err.Error())
			http.Error(w, err.Error(), 500)
			return
		}

		var file *os.File

		if n > _24K {
			file, err = ioutil.TempFile(s.tempPath, "transfer-")
			if err != nil {
				log.Printf("%s", err.Error())
				http.Error(w, err.Error(), 500)
				return
			}

			defer cleanTmpFile(file)

			n, err = io.Copy(file, io.MultiReader(&b, f))
			if err != nil {
				log.Printf("%s", err.Error())
				http.Error(w, err.Error(), 500)
				return
			}

			reader, err = os.Open(file.Name())
		} else {
			reader = bytes.NewReader(b.Bytes())
		}

		contentLength = n
	}

	if contentLength == 0 {
		log.Print("Empty content-length")
		http.Error(w, errors.New("could not upload empty file").Error(), 400)
		return
	}

	contentType := r.Header.Get("Content-Type")

	if contentType == "" {
		contentType = mime.TypeByExtension(filepath.Ext(vars["filename"]))
	}

	token := Encode(10000000 + int64(rand.Intn(1000000000)))

	metadata := s.metadataForRequest(contentType, contentLength, r)

	log.Printf("Uploading %s %s %d %s", token, filename, contentLength, contentType)

	var err error

	if err = s.storage.Put(token, filename, reader, metadata); err != nil {
		log.Printf("Error putting new file: %s", err.Error())
		http.Error(w, errors.New("could not save file").Error(), 500)
		return
	}

	// w.Statuscode = 200

	w.Header().Set("Content-Type", "text/plain")

	filename = url.PathEscape(filename)
	relativeURL, _ := url.Parse(path.Join(s.proxyPath, token, filename))
	deleteURL, _ := url.Parse(path.Join(s.proxyPath, token, filename, metadata.DeletionToken))

	w.Header().Set("X-Url-Delete", resolveURL(r, deleteURL))

	_, _ = fmt.Fprint(w, resolveURL(r, relativeURL))
}

func resolveURL(r *http.Request, u *url.URL) string {
	r.URL.Path = ""

	return getURL(r).ResolveReference(u).String()
}

func resolveKey(key, proxyPath string) string {
	if strings.HasPrefix(key, "/") {
		key = key[1:]
	}

	if strings.HasPrefix(key, proxyPath) {
		key = key[len(proxyPath):]
	}

	key = strings.Replace(key, "\\", "/", -1)

	return key
}

func resolveWebAddress(r *http.Request, proxyPath string) string {
	rUrl := getURL(r)

	var webAddress string

	if len(proxyPath) == 0 {
		webAddress = fmt.Sprintf("%s://%s/",
			rUrl.ResolveReference(rUrl).Scheme,
			rUrl.ResolveReference(rUrl).Host)
	} else {
		webAddress = fmt.Sprintf("%s://%s/%s",
			rUrl.ResolveReference(rUrl).Scheme,
			rUrl.ResolveReference(rUrl).Host,
			proxyPath)
	}

	return webAddress
}

func getURL(r *http.Request) *url.URL {
	u, _ := url.Parse(r.URL.String())

	if r.TLS != nil {
		u.Scheme = "https"
	} else if proto := r.Header.Get("X-Forwarded-Proto"); proto != "" {
		u.Scheme = proto
	} else {
		u.Scheme = "http"
	}

	if u.Host != "" {
	} else if host, port, err := net.SplitHostPort(r.Host); err != nil {
		u.Host = r.Host
	} else {
		if port == "80" && u.Scheme == "http" {
			u.Host = host
		} else if port == "443" && u.Scheme == "https" {
			u.Host = host
		} else {
			u.Host = net.JoinHostPort(host, port)
		}
	}

	return u
}

func (s *Server) Lock(token, filename string) {
	key := path.Join(token, filename)

	if _, ok := s.locks[key]; !ok {
		s.locks[key] = &sync.Mutex{}
	}

	s.locks[key].Lock()
}

func (s *Server) Unlock(token, filename string) {
	key := path.Join(token, filename)
	s.locks[key].Unlock()
}

func (s *Server) CheckMetadata(token, filename string, increaseDownload bool) (metadata storage.Metadata, err error) {
	s.Lock(token, filename)
	defer s.Unlock(token, filename)

	metadata, err = s.storage.Head(token, filename)
	if s.storage.IsNotExist(err) {
		return metadata, nil
	} else if err != nil {
		return metadata, err
	}

	if metadata.MaxDownloads != -1 && metadata.Downloads >= metadata.MaxDownloads {
		return metadata, errors.New("max downloads exceeded")
	} else if !metadata.MaxDate.IsZero() && time.Now().After(metadata.MaxDate) {
		return metadata, errors.New("file access expired")
	} else {

		// update number of downloads
		if increaseDownload {
			metadata.Downloads++
		}

		if err := s.storage.Meta(token, filename, metadata); err != nil {
			return metadata, errors.New("could not save metadata")
		}
	}

	return metadata, nil
}

func (s *Server) CheckDeletionToken(deletionToken, token, filename string) error {
	s.Lock(token, filename)
	defer s.Unlock(token, filename)

	metadata, err := s.storage.Head(token, filename)
	if s.storage.IsNotExist(err) {
		return nil
	}
	if err != nil {
		return err
	}

	if metadata.DeletionToken != deletionToken {
		return errors.New("deletion token does not match")
	}

	return nil
}

func (s *Server) deleteHandler(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)

	token := vars["token"]
	filename := vars["filename"]
	deletionToken := vars["deletionToken"]

	if err := s.CheckDeletionToken(deletionToken, token, filename); err != nil {
		log.Printf("Error metadata: %s", err.Error())
		http.Error(w, http.StatusText(http.StatusNotFound), http.StatusNotFound)
		return
	}

	err := s.storage.Delete(token, filename)
	if s.storage.IsNotExist(err) {
		http.Error(w, http.StatusText(http.StatusNotFound), http.StatusNotFound)
		return
	} else if err != nil {
		log.Printf("%s", err.Error())
		http.Error(w, "Could not delete file.", 500)
		return
	}
}

func (s *Server) zipHandler(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)

	files := vars["files"]

	zipfilename := fmt.Sprintf("transfersh-%d.zip", uint16(time.Now().UnixNano()))

	w.Header().Set("Content-Type", "application/zip")
	w.Header().Set("Content-Disposition", fmt.Sprintf("attachment; filename=\"%s\"", zipfilename))
	w.Header().Set("Connection", "close")

	zw := zip.NewWriter(w)

	for _, key := range strings.Split(files, ",") {
		key = resolveKey(key, s.proxyPath)

		token := strings.Split(key, "/")[0]
		filename := sanitize(strings.Split(key, "/")[1])

		if _, err := s.CheckMetadata(token, filename, true); err != nil {
			log.Printf("Error metadata: %s", err.Error())
			continue
		}

		reader, _, err := s.storage.Get(token, filename)

		if err != nil {
			if s.storage.IsNotExist(err) {
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
			Name:     strings.Split(key, "/")[1],
			Method:   zip.Store,
			Modified: time.Now(),
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

func (s *Server) tarGzHandler(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)

	files := vars["files"]

	tarfilename := fmt.Sprintf("transfersh-%d.tar.gz", uint16(time.Now().UnixNano()))

	w.Header().Set("Content-Type", "application/x-gzip")
	w.Header().Set("Content-Disposition", fmt.Sprintf("attachment; filename=\"%s\"", tarfilename))
	w.Header().Set("Connection", "close")

	writer := gzip.NewWriter(w)
	defer writer.Close()

	zw := tar.NewWriter(writer)
	defer zw.Close()

	for _, key := range strings.Split(files, ",") {
		key = resolveKey(key, s.proxyPath)

		token := strings.Split(key, "/")[0]
		filename := sanitize(strings.Split(key, "/")[1])

		if _, err := s.CheckMetadata(token, filename, true); err != nil {
			log.Printf("Error metadata: %s", err.Error())
			continue
		}

		reader, metadata, err := s.storage.Get(token, filename)
		if err != nil {
			if s.storage.IsNotExist(err) {
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
			Size: metadata.ContentLength,
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

func (s *Server) tarHandler(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)

	files := vars["files"]

	tarfilename := fmt.Sprintf("transfersh-%d.tar", uint16(time.Now().UnixNano()))

	w.Header().Set("Content-Type", "application/x-tar")
	w.Header().Set("Content-Disposition", fmt.Sprintf("attachment; filename=\"%s\"", tarfilename))
	w.Header().Set("Connection", "close")

	zw := tar.NewWriter(w)
	defer zw.Close()

	for _, key := range strings.Split(files, ",") {
		key = resolveKey(key, s.proxyPath)

		token := strings.Split(key, "/")[0]
		filename := strings.Split(key, "/")[1]

		if _, err := s.CheckMetadata(token, filename, true); err != nil {
			log.Printf("Error metadata: %s", err.Error())
			continue
		}

		reader, metadata, err := s.storage.Get(token, filename)
		if err != nil {
			if s.storage.IsNotExist(err) {
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
			Size: metadata.ContentLength,
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

func (s *Server) headHandler(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)

	token := vars["token"]
	filename := vars["filename"]

	metadata, err := s.CheckMetadata(token, filename, false)

	if err != nil {
		log.Printf("Error metadata: %s", err.Error())
		http.Error(w, http.StatusText(http.StatusNotFound), http.StatusNotFound)
		return
	}

	if s.storage.IsNotExist(err) {
		http.Error(w, http.StatusText(http.StatusNotFound), http.StatusNotFound)
		return
	} else if err != nil {
		log.Printf("%s", err.Error())
		http.Error(w, "Could not retrieve file.", 500)
		return
	}

	remainingDownloads, remainingDays := metadata.RemainingLimitHeaderValues()

	w.Header().Set("Content-Type", metadata.ContentType)
	w.Header().Set("Content-Length", strconv.FormatInt(metadata.ContentLength, 10))
	w.Header().Set("Connection", "close")
	w.Header().Set("X-Remaining-Downloads", remainingDownloads)
	w.Header().Set("X-Remaining-Days", remainingDays)
}

func (s *Server) getHandler(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)

	action := vars["action"]
	token := vars["token"]
	filename := vars["filename"]

	metadata, err := s.CheckMetadata(token, filename, true)

	if err != nil {
		log.Printf("Error metadata: %s", err.Error())
		http.Error(w, http.StatusText(http.StatusNotFound), http.StatusNotFound)
		return
	}

	reader, _, err := s.storage.Get(token, filename)
	if s.storage.IsNotExist(err) {
		http.Error(w, http.StatusText(http.StatusNotFound), http.StatusNotFound)
		return
	} else if err != nil {
		log.Printf("%s", err.Error())
		http.Error(w, "Could not retrieve file.", 500)
		return
	}

	defer reader.Close()

	var disposition string

	if action == "inline" {
		disposition = "inline"
	} else {
		disposition = "attachment"
	}

	remainingDownloads, remainingDays := metadata.RemainingLimitHeaderValues()

	w.Header().Set("Content-Type", metadata.ContentType)
	w.Header().Set("Content-Length", strconv.FormatInt(metadata.ContentLength, 10))
	w.Header().Set("Content-Disposition", fmt.Sprintf("%s; filename=\"%s\"", disposition, filename))
	w.Header().Set("Connection", "keep-alive")
	w.Header().Set("X-Remaining-Downloads", remainingDownloads)
	w.Header().Set("X-Remaining-Days", remainingDays)

	if w.Header().Get("Range") == "" {
		if _, err = io.Copy(w, reader); err != nil {
			log.Printf("%s", err.Error())
			http.Error(w, "Error occurred copying to output stream", 500)
			return
		}

		return
	}

	file, err := ioutil.TempFile(s.tempPath, "range-")
	if err != nil {
		log.Printf("%s", err.Error())
		http.Error(w, "Error occurred copying to output stream", 500)
		return
	}

	defer cleanTmpFile(file)

	tee := io.TeeReader(reader, file)
	for {
		b := make([]byte, _5M)
		_, err = tee.Read(b)
		if err == io.EOF {
			break
		}

		if err != nil {
			log.Printf("%s", err.Error())
			http.Error(w, "Error occurred copying to output stream", 500)
			return
		}
	}

	http.ServeContent(w, r, filename, time.Now(), file)
}

func (s *Server) RedirectHandler(h http.Handler) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		if !s.forceHTTPs {
			// we don't want to enforce https
		} else if r.URL.Path == "/health.html" {
			// health check url won't redirect
		} else if strings.HasSuffix(ipAddrFromRemoteAddr(r.Host), ".onion") {
			// .onion addresses cannot get a valid certificate, so don't redirect
		} else if r.Header.Get("X-Forwarded-Proto") == "https" {
		} else if r.URL.Scheme == "https" {
		} else {
			u := getURL(r)
			u.Scheme = "https"

			http.Redirect(w, r, u.String(), http.StatusPermanentRedirect)
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

func IPFilterHandler(h http.Handler, ipFilterOptions *IPFilterOptions) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		if ipFilterOptions == nil {
			h.ServeHTTP(w, r)
		} else {
			WrapIPFilter(h, *ipFilterOptions).ServeHTTP(w, r)
		}
		return
	}
}

func (s *Server) BasicAuthHandler(h http.Handler) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		if s.AuthUser == "" || s.AuthPass == "" {
			h.ServeHTTP(w, r)
			return
		}

		w.Header().Set("WWW-Authenticate", "Basic realm=\"Restricted\"")

		username, password, authOK := r.BasicAuth()
		if authOK == false {
			http.Error(w, "Not authorized", 401)
			return
		}

		if username != s.AuthUser || password != s.AuthPass {
			http.Error(w, "Not authorized", 401)
			return
		}

		h.ServeHTTP(w, r)
	}
}
