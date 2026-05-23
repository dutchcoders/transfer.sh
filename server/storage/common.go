package storage

import (
	"context"
	"fmt"
	"io"
	"regexp"
	"strconv"
	"time"
)

type Range struct {
	Start        uint64
	Limit        uint64
	contentRange string
}

// Range Reconstructs Range header and returns it
func (r *Range) Range() string {
	if r.Limit > 0 {
		return fmt.Sprintf("bytes=%d-%d", r.Start, r.Start+r.Limit-1)
	} else {
		return fmt.Sprintf("bytes=%d-", r.Start)
	}
}

// AcceptLength Tries to accept given range
// returns newContentLength if range was satisfied, otherwise returns given contentLength
func (r *Range) AcceptLength(contentLength uint64) (newContentLength uint64) {
	newContentLength = contentLength
	if r.Limit == 0 {
		r.Limit = newContentLength - r.Start
	}
	if contentLength < r.Start {
		return
	}
	if r.Limit > contentLength-r.Start {
		return
	}
	r.contentRange = fmt.Sprintf("bytes %d-%d/%d", r.Start, r.Start+r.Limit-1, contentLength)
	newContentLength = r.Limit
	return
}

func (r *Range) SetContentRange(cr string) {
	r.contentRange = cr
}

// Returns accepted Content-Range header. If range wasn't accepted empty string is returned
func (r *Range) ContentRange() string {
	return r.contentRange
}

var rexp *regexp.Regexp = regexp.MustCompile(`^bytes=([0-9]+)-([0-9]*)$`)

// Parses HTTP Range header and returns struct on success
// only bytes=start-finish supported
func ParseRange(rng string) *Range {
	if rng == "" {
		return nil
	}

	matches := rexp.FindAllStringSubmatch(rng, -1)
	if len(matches) != 1 || len(matches[0]) != 3 {
		return nil
	}
	if len(matches[0][0]) != len(rng) || len(matches[0][1]) == 0 {
		return nil
	}

	start, err := strconv.ParseUint(matches[0][1], 10, 64)
	if err != nil {
		return nil
	}

	if len(matches[0][2]) == 0 {
		return &Range{Start: start, Limit: 0}
	}

	finish, err := strconv.ParseUint(matches[0][2], 10, 64)
	if err != nil {
		return nil
	}
	if finish < start || finish+1 < finish {
		return nil
	}

	return &Range{Start: start, Limit: finish - start + 1}
}

// Storage is the interface for storage operation
type Storage interface {
	// Get retrieves a file from storage
	Get(ctx context.Context, token string, filename string, rng *Range) (reader io.ReadCloser, contentLength uint64, err error)
	// Head retrieves content length of a file from storage
	Head(ctx context.Context, token string, filename string) (contentLength uint64, err error)
	// Put saves a file on storage
	Put(ctx context.Context, token string, filename string, reader io.Reader, contentType string, contentLength uint64) error
	// Delete removes a file from storage
	Delete(ctx context.Context, token string, filename string) error
	// IsNotExist indicates if a file doesn't exist on storage
	IsNotExist(err error) bool
	// Purge cleans up the storage
	Purge(ctx context.Context, days time.Duration) error
	// Whether storage supports Get with Range header
	IsRangeSupported() bool
	// Type returns the storage type
	Type() string
}

func CloseCheck(c io.Closer) {
	if c == nil {
		return
	}

	if err := c.Close(); err != nil {
		fmt.Println("Received close error:", err)
	}
}
