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

// Purge deletes the uploads under basePath older than days. Only files laid
// out as transfer.sh writes them are touched, basePath/<token>/<file> with a
// <file>.metadata next to it, because a WebDAV base path is often a folder on
// an existing share (Nextcloud, a NAS) that also holds files of its own.
// Emptied token directories are left in place, like the local provider does:
// a WebDAV DELETE on a collection is recursive, so removing one could take
// down a file uploaded after it was listed.
func (s *WebDAVStorage) Purge(ctx context.Context, days time.Duration) error {
	tokens, err := s.client.ReadDir(s.basePath)
	if err != nil {
		if s.IsNotExist(err) {
			// Nothing was ever uploaded, so there is nothing to purge.
			return nil
		}
		return fmt.Errorf("webdav purge %s: %w", s.basePath, err)
	}

	cutoff := time.Now().Add(-days)
	for _, token := range tokens {
		if !token.IsDir() {
			continue
		}
		if err := s.purgeToken(ctx, path.Join(s.basePath, token.Name()), cutoff); err != nil {
			return err
		}
	}
	return nil
}

// purgeToken deletes, in one token directory, every upload whose data file was
// last modified before cutoff, together with its metadata. The data file
// dates the upload: the metadata is rewritten on each download to count it.
func (s *WebDAVStorage) purgeToken(ctx context.Context, dir string, cutoff time.Time) error {
	entries, err := s.client.ReadDir(dir)
	if err != nil {
		return fmt.Errorf("webdav purge %s: %w", dir, err)
	}

	names := make(map[string]bool, len(entries))
	for _, fi := range entries {
		if !fi.IsDir() {
			names[fi.Name()] = true
		}
	}

	for _, fi := range entries {
		if err := ctx.Err(); err != nil {
			return err
		}

		metadata := fi.Name() + ".metadata"
		if fi.IsDir() || !names[metadata] || !webdavModifiedBefore(fi, cutoff) {
			continue
		}
		// Data first: a leftover metadata file is harmless, while metadata
		// removed first would leave a file that never expires. gowebdav
		// already treats a 404 as success, so an upload deleted by its owner
		// in the meantime is not an error.
		for _, name := range []string{fi.Name(), metadata} {
			p := path.Join(dir, name)
			if err := s.client.Remove(p); err != nil {
				return fmt.Errorf("webdav purge %s: %w", p, err)
			}
		}
	}
	return nil
}

// webdavModifiedBefore reports whether fi is known to be older than cutoff.
// A server that does not send getlastmodified gives no usable date (gowebdav
// falls back to the Unix epoch), and such a file is never considered old.
func webdavModifiedBefore(fi os.FileInfo, cutoff time.Time) bool {
	mtime := fi.ModTime()
	if mtime.IsZero() || mtime.Unix() == 0 {
		return false
	}
	return mtime.Before(cutoff)
}

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
