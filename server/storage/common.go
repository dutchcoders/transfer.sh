package storage

import (
	"io"
	"time"
)

type Metadata struct {
	// ContentType is the original uploading content type
	ContentType string
	// ContentLength contains the length of the actual object
	ContentLength int64
	// Downloads is the actual number of downloads
	Downloads int
	// MaxDownloads contains the maximum numbers of downloads
	MaxDownloads int
	// MaxDate contains the max age of the file
	MaxDate time.Time
	// DeletionToken contains the token to match against for deletion
	DeletionToken string
	// Secret as knowledge to delete file
	Secret string
}

type Storage interface {
	Get(token string, filename string) (reader io.ReadCloser, metaData Metadata, err error)
	Head(token string, filename string) (metadata Metadata, err error)
	Meta(token string, filename string, metadata Metadata) error
	Put(token string, filename string, reader io.Reader, metadata Metadata) error
	Delete(token string, filename string) error
	IsNotExist(err error) bool
	DeleteExpired() error

	Type() string
}
