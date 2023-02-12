package storage

import (
	"context"
	"fmt"
	"io"
	"log"
	"os"
	"path/filepath"
	"time"
)

// LocalStorage is a local storage
type LocalStorage struct {
	Storage
	basedir string
	logger  *log.Logger
}

// NewLocalStorage is the factory for LocalStorage
func NewLocalStorage(basedir string, logger *log.Logger) (*LocalStorage, error) {
	return &LocalStorage{basedir: basedir, logger: logger}, nil
}

// Type returns the storage type
func (s *LocalStorage) Type() string {
	return "local"
}

// Head retrieves content length of a file from storage
func (s *LocalStorage) Head(_ context.Context, token string, filename string) (contentLength uint64, err error) {
	path := filepath.Join(s.basedir, token, filename)

	var fi os.FileInfo
	if fi, err = os.Lstat(path); err != nil {
		return
	}

	contentLength = uint64(fi.Size())

	return
}

// Get retrieves a file from storage
func (s *LocalStorage) Get(_ context.Context, token string, filename string, rng *Range) (reader io.ReadCloser, contentLength uint64, err error) {
	path := filepath.Join(s.basedir, token, filename)

	var file *os.File

	// content type , content length
	if file, err = os.Open(path); err != nil {
		return
	}
	reader = file

	var fi os.FileInfo
	if fi, err = os.Lstat(path); err != nil {
		return
	}

	contentLength = uint64(fi.Size())
	if rng != nil {
		contentLength = rng.AcceptLength(contentLength)
		if _, err = file.Seek(int64(rng.Start), 0); err != nil {
			return
		}
	}

	return
}

// Delete removes a file from storage
func (s *LocalStorage) Delete(_ context.Context, token string, filename string) (err error) {
	metadata := filepath.Join(s.basedir, token, fmt.Sprintf("%s.metadata", filename))
	_ = os.Remove(metadata)

	path := filepath.Join(s.basedir, token, filename)
	err = os.Remove(path)
	return
}

// Purge cleans up the storage
func (s *LocalStorage) Purge(_ context.Context, days time.Duration) (err error) {
	err = filepath.Walk(s.basedir,
		func(path string, info os.FileInfo, err error) error {
			if err != nil {
				return err
			}
			if info.IsDir() {
				return nil
			}

			if info.ModTime().Before(time.Now().Add(-1 * days)) {
				err = os.Remove(path)
				return err
			}

			return nil
		})

	return
}

// IsNotExist indicates if a file doesn't exist on storage
func (s *LocalStorage) IsNotExist(err error) bool {
	if err == nil {
		return false
	}

	return os.IsNotExist(err)
}

// Put saves a file on storage
func (s *LocalStorage) Put(_ context.Context, token string, filename string, reader io.Reader, contentType string, contentLength uint64) error {
	var f io.WriteCloser
	var err error

	path := filepath.Join(s.basedir, token)

	if err = os.MkdirAll(path, 0700); err != nil && !os.IsExist(err) {
		return err
	}

	f, err = os.OpenFile(filepath.Join(path, filename), os.O_RDWR|os.O_CREATE|os.O_TRUNC, 0600)
	defer CloseCheck(f)

	if err != nil {
		return err
	}

	if _, err = io.Copy(f, reader); err != nil {
		return err
	}

	return nil
}

func (s *LocalStorage) IsRangeSupported() bool { return true }
