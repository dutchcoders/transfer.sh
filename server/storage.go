package server

import (
	"fmt"
	"io"
	"log"
	"mime"
	"os"
	"path/filepath"

	"github.com/minio/minio-go"
)

type Storage interface {
	Get(token string, filename string) (reader io.ReadCloser, contentType string, contentLength uint64, err error)
	Head(token string, filename string) (contentType string, contentLength uint64, err error)
	Put(token string, filename string, reader io.Reader, contentType string, contentLength uint64) error
	IsNotExist(err error) bool

	Type() string
}

type LocalStorage struct {
	Storage
	basedir string
}

func NewLocalStorage(basedir string) (*LocalStorage, error) {
	return &LocalStorage{basedir: basedir}, nil
}

func (s *LocalStorage) Type() string {
	return "local"
}

func (s *LocalStorage) Head(token string, filename string) (contentType string, contentLength uint64, err error) {
	path := filepath.Join(s.basedir, token, filename)

	var fi os.FileInfo
	if fi, err = os.Lstat(path); err != nil {
		return
	}

	contentLength = uint64(fi.Size())

	contentType = mime.TypeByExtension(filepath.Ext(filename))

	return
}

func (s *LocalStorage) Get(token string, filename string) (reader io.ReadCloser, contentType string, contentLength uint64, err error) {
	path := filepath.Join(s.basedir, token, filename)

	// content type , content length
	if reader, err = os.Open(path); err != nil {
		return
	}

	var fi os.FileInfo
	if fi, err = os.Lstat(path); err != nil {
		return
	}

	contentLength = uint64(fi.Size())

	contentType = mime.TypeByExtension(filepath.Ext(filename))

	return
}

func (s *LocalStorage) IsNotExist(err error) bool {
	if err == nil {
		return false
	}

	return os.IsNotExist(err)
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
	bucket string
	client *minio.Client
}

func NewS3Storage(endpoint, accessKey, secretKey, bucket string) (*S3Storage, error) {
	s3Client, err := minio.NewV4(endpoint, accessKey, secretKey, true)
	if err != nil {
		return nil, err
	}

	return &S3Storage{
		client: s3Client,
		bucket: bucket,
	}, nil
}

func (s *S3Storage) Type() string {
	return "s3"
}

func (s *S3Storage) Head(token string, filename string) (contentType string, contentLength uint64, err error) {
	key := fmt.Sprintf("%s/%s", token, filename)

	// content type , content length
	objInfo, err := s.client.StatObject(s.bucket, key)
	if err != nil {
		return
	}

	contentType = objInfo.ContentType
	contentLength = uint64(objInfo.Size)
	return
}

func (s *S3Storage) IsNotExist(err error) bool {
	if err == nil {
		return false
	}

	log.Printf("IsNotExist: %s, %#v", err.Error(), err)

	b := (err.Error() == "The specified key does not exist.")
	b = b || (err.Error() == "Access Denied")
	return b
}

func (s *S3Storage) Get(token string, filename string) (reader io.ReadCloser, contentType string, contentLength uint64, err error) {
	key := fmt.Sprintf("%s/%s", token, filename)

	// content type , content length
	obj, err := s.client.GetObject(s.bucket, key)
	if err != nil {
		return
	}
	// obj is *minio.Object - implements io.ReadCloser.
	reader = obj

	objInfo, err := obj.Stat()
	if err != nil {
		return
	}

	contentType = objInfo.ContentType
	contentLength = uint64(objInfo.Size)
	return
}

func (s *S3Storage) Put(token string, filename string, reader io.Reader, contentType string, contentLength uint64) (err error) {
	key := fmt.Sprintf("%s/%s", token, filename)

	n, err := s.client.PutObject(s.bucket, key, reader, contentType)
	if err != nil {
		return
	}
	if uint64(n) != contentLength {
		err = fmt.Errorf("Uploaded content %d is not equal to requested length %d", n, contentLength)
		return
	}
	log.Printf("Completed uploading %s", key)
	return
}
