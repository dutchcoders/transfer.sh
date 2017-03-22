// Copyright 2013 The Go Authors. All rights reserved.
//
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file or at
// https://developers.google.com/open-source/licenses/bsd.

package gosrc

import (
	"log"
	"net/http"
	"path"
	"regexp"
	"time"
)

func init() {
	addService(&service{
		pattern: regexp.MustCompile(`^bitbucket\.org/(?P<owner>[a-z0-9A-Z_.\-]+)/(?P<repo>[a-z0-9A-Z_.\-]+)(?P<dir>/[a-z0-9A-Z_.\-/]*)?$`),
		prefix:  "bitbucket.org/",
		get:     getBitbucketDir,
	})
}

var bitbucketEtagRe = regexp.MustCompile(`^(hg|git)-`)

type bitbucketRepo struct {
	Scm         string
	CreatedOn   string `json:"created_on"`
	LastUpdated string `json:"last_updated"`
	ForkOf      struct {
		Scm string
	} `json:"fork_of"`
	Followers int  `json:"followers_count"`
	IsFork    bool `json:"is_fork"`
}

type bitbucketNode struct {
	Node      string `json:"node"`
	Timestamp string `json:"utctimestamp"`
}

func getBitbucketDir(client *http.Client, match map[string]string, savedEtag string) (*Directory, error) {
	var repo *bitbucketRepo
	c := &httpClient{client: client}

	if m := bitbucketEtagRe.FindStringSubmatch(savedEtag); m != nil {
		match["vcs"] = m[1]
	} else {
		repo, err := getBitbucketRepo(c, match)
		if err != nil {
			return nil, err
		}

		match["vcs"] = repo.Scm
	}

	tags := make(map[string]string)
	timestamps := make(map[string]time.Time)

	for _, nodeType := range []string{"branches", "tags"} {
		var nodes map[string]bitbucketNode
		if _, err := c.getJSON(expand("https://api.bitbucket.org/1.0/repositories/{owner}/{repo}/{0}", match, nodeType), &nodes); err != nil {
			return nil, err
		}
		for t, n := range nodes {
			tags[t] = n.Node
			const timeFormat = "2006-01-02 15:04:05Z07:00"
			committed, err := time.Parse(timeFormat, n.Timestamp)
			if err != nil {
				log.Println("error parsing timestamp:", n.Timestamp)
				continue
			}
			timestamps[t] = committed
		}
	}

	var err error
	tag, commit, err := bestTag(tags, defaultTags[match["vcs"]])
	if err != nil {
		return nil, err
	}
	match["tag"] = tag
	match["commit"] = commit
	etag := expand("{vcs}-{commit}", match)
	if etag == savedEtag {
		return nil, NotModifiedError{Since: timestamps[tag]}
	}

	if repo == nil {
		repo, err = getBitbucketRepo(c, match)
		if err != nil {
			return nil, err
		}
	}

	var contents struct {
		Directories []string
		Files       []struct {
			Path string
		}
	}

	if _, err := c.getJSON(expand("https://api.bitbucket.org/1.0/repositories/{owner}/{repo}/src/{tag}{dir}/", match), &contents); err != nil {
		return nil, err
	}

	var files []*File
	var dataURLs []string

	for _, f := range contents.Files {
		_, name := path.Split(f.Path)
		if isDocFile(name) {
			files = append(files, &File{Name: name, BrowseURL: expand("https://bitbucket.org/{owner}/{repo}/src/{tag}/{0}", match, f.Path)})
			dataURLs = append(dataURLs, expand("https://api.bitbucket.org/1.0/repositories/{owner}/{repo}/raw/{tag}/{0}", match, f.Path))
		}
	}

	if err := c.getFiles(dataURLs, files); err != nil {
		return nil, err
	}

	status := Active
	if isBitbucketDeadEndFork(repo) {
		status = DeadEndFork
	}

	return &Directory{
		BrowseURL:      expand("https://bitbucket.org/{owner}/{repo}/src/{tag}{dir}", match),
		Etag:           etag,
		Files:          files,
		LineFmt:        "%s#cl-%d",
		ProjectName:    match["repo"],
		ProjectRoot:    expand("bitbucket.org/{owner}/{repo}", match),
		ProjectURL:     expand("https://bitbucket.org/{owner}/{repo}/", match),
		Subdirectories: contents.Directories,
		VCS:            match["vcs"],
		Status:         status,
		Fork:           repo.IsFork,
		Stars:          repo.Followers,
	}, nil
}

func getBitbucketRepo(c *httpClient, match map[string]string) (*bitbucketRepo, error) {
	var repo bitbucketRepo
	if _, err := c.getJSON(expand("https://api.bitbucket.org/1.0/repositories/{owner}/{repo}", match), &repo); err != nil {
		return nil, err
	}

	return &repo, nil
}

func isBitbucketDeadEndFork(repo *bitbucketRepo) bool {
	l := "2006-01-02T15:04:05.999999999"
	created, err := time.Parse(l, repo.CreatedOn)
	if err != nil {
		return false
	}

	updated, err := time.Parse(l, repo.LastUpdated)
	if err != nil {
		return false
	}

	isDeadEndFork := false
	if repo.ForkOf.Scm != "" && created.Unix() >= updated.Unix() {
		isDeadEndFork = true
	}

	return isDeadEndFork
}
