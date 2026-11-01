package storage

import (
	"context"
	"errors"
	"fmt"
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
		return nil, fmt.Errorf("webdav connect to %s: %w", url, err)
	}
	return &WebDAVStorage{client: c, basePath: basePath, logger: logger}, nil
}

// Type returns the storage type
func (s *WebDAVStorage) Type() string { return "webdav" }

// Head retrieves content length of a file from storage
func (s *WebDAVStorage) Head(_ context.Context, token, filename string) (uint64, error) {
	p := ObjectPath(s.basePath, token, filename)
	fi, err := s.client.Stat(p)
	if err != nil {
		return 0, fmt.Errorf("webdav head %s: %w", p, err)
	}
	return uint64(fi.Size()), nil
}

// Get retrieves a file from storage
func (s *WebDAVStorage) Get(_ context.Context, token, filename string, rng *Range) (io.ReadCloser, uint64, error) {
	p := ObjectPath(s.basePath, token, filename)

	// Stat first: the size is needed to validate the range, and a missing
	// file is reported before any body is opened.
	fi, err := s.client.Stat(p)
	if err != nil {
		return nil, 0, fmt.Errorf("webdav get %s: %w", p, err)
	}
	contentLength := uint64(fi.Size())

	if rng != nil {
		contentLength = rng.AcceptLength(contentLength)
		// Only ask the WebDAV server for a range we know is satisfiable,
		// otherwise fall through and serve the whole file as the other
		// providers do.
		if rng.ContentRange() != "" {
			rc, err := s.client.ReadStreamRange(p, int64(rng.Start), int64(rng.Limit))
			if err != nil {
				return nil, 0, fmt.Errorf("webdav get %s (%s): %w", p, rng.Range(), err)
			}
			return rc, contentLength, nil
		}
	}

	rc, err := s.client.ReadStream(p)
	if err != nil {
		return nil, 0, fmt.Errorf("webdav get %s: %w", p, err)
	}
	return rc, contentLength, nil
}

// Delete removes a file from storage
func (s *WebDAVStorage) Delete(_ context.Context, token, filename string) error {
	p := ObjectPath(s.basePath, token, filename)
	if err := s.client.Remove(p); err != nil {
		return fmt.Errorf("webdav delete %s: %w", p, err)
	}
	return nil
}

// Purge cleans up the storage (noop for webdav)
func (s *WebDAVStorage) Purge(context.Context, time.Duration) error { return nil }

// Put saves a file on storage
func (s *WebDAVStorage) Put(_ context.Context, token, filename string, reader io.Reader, _ string, _ uint64) error {
	dir := path.Join(s.basePath, token)
	if err := s.client.MkdirAll(dir, 0755); err != nil {
		return fmt.Errorf("webdav mkdir %s: %w", dir, err)
	}
	p := ObjectPath(s.basePath, token, filename)
	if err := s.client.WriteStream(p, reader, 0644); err != nil {
		return fmt.Errorf("webdav put %s: %w", p, err)
	}
	return nil
}

// IsNotExist indicates if a file doesn't exist on storage
func (s *WebDAVStorage) IsNotExist(err error) bool {
	// gowebdav reports every failure as an *os.PathError, so only a 404
	// status means the file is missing. gowebdav.IsErrNotFound does not
	// unwrap, hence the errors.As.
	var pathErr *os.PathError
	return errors.As(err, &pathErr) && gowebdav.IsErrNotFound(pathErr)
}

// IsRangeSupported returns true: ranges are forwarded to the WebDAV server
// as Range requests, and gowebdav emulates them when the server ignores them
func (s *WebDAVStorage) IsRangeSupported() bool { return true }
