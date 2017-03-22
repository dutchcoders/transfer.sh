// Copyright 2013 The Go Authors. All rights reserved.
//
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file or at
// https://developers.google.com/open-source/licenses/bsd.

package gosrc

import (
	"encoding/json"
	"fmt"
	"io"
	"io/ioutil"
	"net/http"
)

type httpClient struct {
	errFn  func(*http.Response) error
	header http.Header
	client *http.Client
}

func (c *httpClient) err(resp *http.Response) error {
	if resp.StatusCode == 404 {
		return NotFoundError{Message: "Resource not found: " + resp.Request.URL.String()}
	}
	if c.errFn != nil {
		return c.errFn(resp)
	}
	return &RemoteError{resp.Request.URL.Host, fmt.Errorf("%d: (%s)", resp.StatusCode, resp.Request.URL.String())}
}

// get issues a GET to the specified URL.
func (c *httpClient) get(url string) (*http.Response, error) {
	req, err := http.NewRequest("GET", url, nil)
	if err != nil {
		return nil, err
	}
	for k, vs := range c.header {
		req.Header[k] = vs
	}
	resp, err := c.client.Do(req)
	if err != nil {
		return nil, &RemoteError{req.URL.Host, err}
	}
	return resp, err
}

// getNoFollow issues a GET to the specified URL without following redirects.
func (c *httpClient) getNoFollow(url string) (*http.Response, error) {
	req, err := http.NewRequest("GET", url, nil)
	if err != nil {
		return nil, err
	}
	for k, vs := range c.header {
		req.Header[k] = vs
	}
	t := c.client.Transport
	if t == nil {
		t = http.DefaultTransport
	}
	resp, err := t.RoundTrip(req)
	if err != nil {
		return nil, &RemoteError{req.URL.Host, err}
	}
	return resp, err
}

func (c *httpClient) getBytes(url string) ([]byte, error) {
	resp, err := c.get(url)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()
	if resp.StatusCode != 200 {
		return nil, c.err(resp)
	}
	p, err := ioutil.ReadAll(resp.Body)
	return p, err
}

func (c *httpClient) getReader(url string) (io.ReadCloser, error) {
	resp, err := c.get(url)
	if err != nil {
		return nil, err
	}
	if resp.StatusCode != 200 {
		err = c.err(resp)
		resp.Body.Close()
		return nil, err
	}
	return resp.Body, nil
}

func (c *httpClient) getJSON(url string, v interface{}) (*http.Response, error) {
	resp, err := c.get(url)
	if err != nil {
		return resp, err
	}
	defer resp.Body.Close()
	if resp.StatusCode != 200 {
		return resp, c.err(resp)
	}
	err = json.NewDecoder(resp.Body).Decode(v)
	if _, ok := err.(*json.SyntaxError); ok {
		err = NotFoundError{Message: "JSON syntax error at " + url}
	}
	return resp, err
}

func (c *httpClient) getFiles(urls []string, files []*File) error {
	ch := make(chan error, len(files))
	for i := range files {
		go func(i int) {
			resp, err := c.get(urls[i])
			if err != nil {
				ch <- err
				return
			}
			defer resp.Body.Close()
			if resp.StatusCode != 200 {
				var err error
				if c.errFn != nil {
					err = c.errFn(resp)
				} else {
					err = &RemoteError{resp.Request.URL.Host, fmt.Errorf("get %s -> %d", urls[i], resp.StatusCode)}
				}
				ch <- err
				return
			}
			files[i].Data, err = ioutil.ReadAll(resp.Body)
			if err != nil {
				ch <- &RemoteError{resp.Request.URL.Host, err}
				return
			}
			ch <- nil
		}(i)
	}
	for range files {
		if err := <-ch; err != nil {
			return err
		}
	}
	return nil
}
