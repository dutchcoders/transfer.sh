package storage

import (
	"context"
	"io"
	"log"
	"os"
	"path"
	"time"

	"github.com/studio-b12/gowebdav"
)

// WebDAVStorage is a storage backed by a WebDAV server
// basePath is the root directory within the WebDAV server
// where files will be stored.
type WebDAVStorage struct {
	client   *gowebdav.Client
	basePath string
	logger   *log.Logger
}

// NewWebDAVStorage creates a new WebDAVStorage
func NewWebDAVStorage(url, basePath, username, password string, logger *log.Logger) (*WebDAVStorage, error) {
	c := gowebdav.NewClient(url, username, password)
	if err := c.Connect(); err != nil {
		if logger != nil {
			logger.Printf("webdav connect error: %v", err)
		}
		return nil, err
	}
	return &WebDAVStorage{client: c, basePath: basePath, logger: logger}, nil
}

// Type returns the storage type
func (s *WebDAVStorage) Type() string { return "webdav" }

func (s *WebDAVStorage) fullPath(token, filename string) string {
	return path.Join(s.basePath, token, filename)
}

// Head retrieves content length of a file from storage
func (s *WebDAVStorage) Head(_ context.Context, token, filename string) (uint64, error) {
	fi, err := s.client.Stat(s.fullPath(token, filename))
	if err != nil {
		if s.logger != nil {
			s.logger.Printf("webdav head %s/%s error: %v", token, filename, err)
		}
		return 0, err
	}
	return uint64(fi.Size()), nil
}

// Get retrieves a file from storage
func (s *WebDAVStorage) Get(_ context.Context, token, filename string, rng *Range) (io.ReadCloser, uint64, error) {
	p := s.fullPath(token, filename)
	var rc io.ReadCloser
	var err error
	if rng != nil {
		rc, err = s.client.ReadStreamRange(p, int64(rng.Start), int64(rng.Limit))
	} else {
		rc, err = s.client.ReadStream(p)
	}
	if err != nil {
		if s.logger != nil {
			s.logger.Printf("webdav get %s/%s error: %v", token, filename, err)
		}
		return nil, 0, err
	}
	fi, err := s.client.Stat(p)
	if err != nil {
		rc.Close()
		if s.logger != nil {
			s.logger.Printf("webdav stat %s/%s error: %v", token, filename, err)
		}
		return nil, 0, err
	}
	size := uint64(fi.Size())
	if rng != nil {
		size = rng.AcceptLength(size)
	}
	return rc, size, nil
}

// Delete removes a file from storage
func (s *WebDAVStorage) Delete(_ context.Context, token, filename string) error {
	if err := s.client.Remove(s.fullPath(token, filename)); err != nil {
		if s.logger != nil {
			s.logger.Printf("webdav delete %s/%s error: %v", token, filename, err)
		}
		return err
	}
	return nil
}

// Purge cleans up the storage (noop for webdav)
func (s *WebDAVStorage) Purge(context.Context, time.Duration) error { return nil }

// Put saves a file on storage
func (s *WebDAVStorage) Put(_ context.Context, token, filename string, reader io.Reader, _ string, _ uint64) error {
	dir := path.Join(s.basePath, token)
	if err := s.client.MkdirAll(dir, 0755); err != nil {
		if s.logger != nil {
			s.logger.Printf("webdav mkdir %s error: %v", dir, err)
		}
		return err
	}
	if err := s.client.WriteStream(s.fullPath(token, filename), reader, 0644); err != nil {
		if s.logger != nil {
			s.logger.Printf("webdav put %s/%s error: %v", token, filename, err)
		}
		return err
	}
	return nil
}

// IsNotExist indicates if a file doesn't exist on storage
func (s *WebDAVStorage) IsNotExist(err error) bool {
	if err == nil {
		return false
	}
	if _, ok := err.(*os.PathError); ok {
		return true
	}
	return false
}

func (s *WebDAVStorage) IsRangeSupported() bool { return true }
