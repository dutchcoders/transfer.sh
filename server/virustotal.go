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
	"fmt"
	"io"
	"net/http"

	_ "github.com/PuerkitoBio/ghost/handlers"
	"github.com/gorilla/mux"

	virustotal "github.com/dutchcoders/go-virustotal"
)

func (s *Server) virusTotalHandler(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)

	filename := sanitize(vars["filename"])

	contentLength := r.ContentLength
	contentType := r.Header.Get("Content-Type")

	s.logger.Printf("Submitting to VirusTotal: %s %d %s", filename, contentLength, contentType)

	vt, err := virustotal.NewVirusTotal(s.VirusTotalKey)
	if err != nil {
		http.Error(w, err.Error(), 500)
	}

	var reader io.Reader

	reader = r.Body

	result, err := vt.Scan(filename, reader)
	if err != nil {
		http.Error(w, err.Error(), 500)
	}

	s.logger.Println(result)
	w.Write([]byte(fmt.Sprintf("%v\n", result.Permalink)))
}
