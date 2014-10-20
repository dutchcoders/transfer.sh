package main

import (
	"fmt"
	"github.com/goamz/goamz/s3"
	"io"
	"os"
	"path/filepath"
	"strconv"
)

type Storage interface {
	Get(token string, filename string) (reader io.ReadCloser, contentType string, contentLength uint64, err error)
	Put(token string, filename string, reader io.Reader, contentType string, contentLength uint64) error
}

type LocalStorage struct {
	Storage
	basedir string
}

func NewLocalStorage(basedir string) (*LocalStorage, error) {
	return &LocalStorage{basedir: basedir}, nil
}

func (s *LocalStorage) Get(token string, filename string) (reader io.ReadCloser, contentType string, contentLength uint64, err error) {
	path := filepath.Join(s.basedir, token, filename)

	// content type , content length
	if reader, err = os.Open(path); err != nil {
		return
	}

	var fi os.FileInfo
	if fi, err = os.Lstat(path); err != nil {
	}

	contentLength = uint64(fi.Size())

	contentType = ""

	return
}

func (s *LocalStorage) Put(token string, filename string, reader io.Reader, contentType string, contentLength uint64) error {
	var f io.WriteCloser
	var err error

	path := filepath.Join(s.basedir, token)

	if err = os.Mkdir(path, 0700); err != nil && !os.IsExist(err) {
		return err
	}

	if f, err = os.OpenFile(filepath.Join(path, filename), os.O_RDWR|os.O_CREATE|os.O_TRUNC, 0600); err != nil {
		fmt.Printf("%s", err)
		return err
	}

	defer f.Close()

	if _, err = io.Copy(f, reader); err != nil {
		return err
	}

	return nil
}

type S3Storage struct {
	Storage
	bucket *s3.Bucket
}

func NewS3Storage() (*S3Storage, error) {
	bucket, err := getBucket()
	if err != nil {
		return nil, err
	}

	return &S3Storage{bucket: bucket}, nil
}

func (s *S3Storage) Get(token string, filename string) (reader io.ReadCloser, contentType string, contentLength uint64, err error) {
	key := fmt.Sprintf("%s/%s", token, filename)

	// content type , content length
	response, err := s.bucket.GetResponse(key)
	contentType = ""
	contentLength, err = strconv.ParseUint(response.Header.Get("Content-Length"), 10, 0)

	reader = response.Body
	return
}

func (s *S3Storage) Put(token string, filename string, reader io.Reader, contentType string, contentLength uint64) error {
	key := fmt.Sprintf("%s/%s", token, filename)
	err := s.bucket.PutReader(key, reader, int64(contentLength), contentType, s3.Private, s3.Options{})
	return err
}
