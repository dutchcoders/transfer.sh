package server

import (
	"bytes"
	"fmt"
	"io"
	"log"
	"mime"
	"os"
	"path/filepath"
	"strconv"
	"sync"

	"github.com/goamz/goamz/s3"
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
	bucket *s3.Bucket
}

func NewS3Storage(accessKey, secretKey, bucketName, endpoint string) (*S3Storage, error) {
	bucket, err := getBucket(accessKey, secretKey, bucketName, endpoint)
	if err != nil {
		return nil, err
	}

	return &S3Storage{bucket: bucket}, nil
}

func (s *S3Storage) Type() string {
	return "s3"
}

func (s *S3Storage) Head(token string, filename string) (contentType string, contentLength uint64, err error) {
	key := fmt.Sprintf("%s/%s", token, filename)

	// content type , content length
	response, err := s.bucket.Head(key, map[string][]string{})
	if err != nil {
		return
	}

	contentType = response.Header.Get("Content-Type")

	contentLength, err = strconv.ParseUint(response.Header.Get("Content-Length"), 10, 0)
	if err != nil {
		return
	}

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
	response, err := s.bucket.GetResponse(key)
	if err != nil {
		return
	}

	contentType = response.Header.Get("Content-Type")
	contentLength, err = strconv.ParseUint(response.Header.Get("Content-Length"), 10, 0)
	if err != nil {
		return
	}

	reader = response.Body
	return
}

func (s *S3Storage) Put(token string, filename string, reader io.Reader, contentType string, contentLength uint64) (err error) {
	key := fmt.Sprintf("%s/%s", token, filename)

	var (
		multi *s3.Multi
		parts []s3.Part
	)

	if multi, err = s.bucket.InitMulti(key, contentType, s3.Private); err != nil {
		log.Printf(err.Error())
		return
	}

	// 20 mb parts
	partsChan := make(chan interface{})
	// partsChan := make(chan s3.Part)

	go func() {
		// maximize to 20 threads
		sem := make(chan int, 20)
		index := 1
		var wg sync.WaitGroup

		for {
			// buffered in memory because goamz s3 multi needs seekable reader
			var (
				buffer []byte = make([]byte, (1<<20)*10)
				count  int
				err    error
			)

			// Amazon expects parts of at least 5MB, except for the last one
			if count, err = io.ReadAtLeast(reader, buffer, (1<<20)*5); err != nil && err != io.ErrUnexpectedEOF && err != io.EOF {
				log.Printf(err.Error())
				return
			}

			// always send minimal 1 part
			if err == io.EOF && index > 1 {
				log.Printf("Waiting for all parts to finish uploading.")

				// wait for all parts to be finished uploading
				wg.Wait()

				// and close the channel
				close(partsChan)

				return
			}

			wg.Add(1)

			sem <- 1

			// using goroutines because of retries when upload fails
			go func(multi *s3.Multi, buffer []byte, index int) {
				log.Printf("Uploading part %d %d", index, len(buffer))

				defer func() {
					log.Printf("Finished part %d %d", index, len(buffer))

					wg.Done()

					<-sem
				}()

				partReader := bytes.NewReader(buffer)

				var part s3.Part

				if part, err = multi.PutPart(index, partReader); err != nil {
					log.Printf("Error while uploading part %d %d %s", index, len(buffer), err.Error())
					partsChan <- err
					return
				}

				log.Printf("Finished uploading part %d %d", index, len(buffer))

				partsChan <- part

			}(multi, buffer[:count], index)

			index++
		}
	}()

	// wait for all parts to be uploaded
	for part := range partsChan {
		switch part.(type) {
		case s3.Part:
			parts = append(parts, part.(s3.Part))
		case error:
			// abort multi upload
			log.Printf("Error during upload, aborting %s.", part.(error).Error())
			err = part.(error)

			multi.Abort()
			return
		}

	}

	log.Printf("Completing upload %d parts", len(parts))

	if err = multi.Complete(parts); err != nil {
		log.Printf("Error during completing upload %d parts %s", len(parts), err.Error())
		return
	}

	log.Printf("Completed uploading %d", len(parts))

	return
}
