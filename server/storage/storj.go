package storage

import (
	"context"
	"errors"
	"io"
	"log"
	"time"

	"storj.io/common/fpath"
	"storj.io/common/storj"
	"storj.io/uplink"
)

// StorjStorage is a storage backed by Storj
type StorjStorage struct {
	Storage
	project   *uplink.Project
	bucket    *uplink.Bucket
	purgeDays time.Duration
	logger    *log.Logger
}

// NewStorjStorage is the factory for StorjStorage
func NewStorjStorage(ctx context.Context, access, bucket string, purgeDays int, logger *log.Logger) (*StorjStorage, error) {
	var instance StorjStorage
	var err error

	ctx = fpath.WithTempData(ctx, "", true)

	uplConf := &uplink.Config{
		UserAgent: "transfer-sh",
	}

	parsedAccess, err := uplink.ParseAccess(access)
	if err != nil {
		return nil, err
	}

	instance.project, err = uplConf.OpenProject(ctx, parsedAccess)
	if err != nil {
		return nil, err
	}

	instance.bucket, err = instance.project.EnsureBucket(ctx, bucket)
	if err != nil {
		//Ignoring the error to return the one that occurred first, but try to clean up.
		_ = instance.project.Close()
		return nil, err
	}

	instance.purgeDays = time.Duration(purgeDays*24) * time.Hour

	instance.logger = logger

	return &instance, nil
}

// Type returns the storage type
func (s *StorjStorage) Type() string {
	return "storj"
}

// Head retrieves content length of a file from storage
func (s *StorjStorage) Head(ctx context.Context, token string, filename string) (contentLength uint64, err error) {
	key := storj.JoinPaths(token, filename)

	obj, err := s.project.StatObject(fpath.WithTempData(ctx, "", true), s.bucket.Name, key)
	if err != nil {
		return 0, err
	}

	contentLength = uint64(obj.System.ContentLength)

	return
}

// Get retrieves a file from storage
func (s *StorjStorage) Get(ctx context.Context, token string, filename string, rng *Range) (reader io.ReadCloser, contentLength uint64, err error) {
	key := storj.JoinPaths(token, filename)

	s.logger.Printf("Getting file %s from Storj Bucket", filename)

	var options *uplink.DownloadOptions
	if rng != nil {
		options = new(uplink.DownloadOptions)
		options.Offset = int64(rng.Start)
		if rng.Limit > 0 {
			options.Length = int64(rng.Limit)
		} else {
			options.Length = -1
		}
	}

	download, err := s.project.DownloadObject(fpath.WithTempData(ctx, "", true), s.bucket.Name, key, options)
	if err != nil {
		return nil, 0, err
	}

	contentLength = uint64(download.Info().System.ContentLength)
	if rng != nil {
		contentLength = rng.AcceptLength(contentLength)
	}

	reader = download
	return
}

// Delete removes a file from storage
func (s *StorjStorage) Delete(ctx context.Context, token string, filename string) (err error) {
	key := storj.JoinPaths(token, filename)

	s.logger.Printf("Deleting file %s from Storj Bucket", filename)

	_, err = s.project.DeleteObject(fpath.WithTempData(ctx, "", true), s.bucket.Name, key)

	return
}

// Purge cleans up the storage
func (s *StorjStorage) Purge(context.Context, time.Duration) (err error) {
	// NOOP expiration is set at upload time
	return nil
}

// Put saves a file on storage
func (s *StorjStorage) Put(ctx context.Context, token string, filename string, reader io.Reader, contentType string, contentLength uint64) (err error) {
	key := storj.JoinPaths(token, filename)

	s.logger.Printf("Uploading file %s to Storj Bucket", filename)

	var uploadOptions *uplink.UploadOptions
	if s.purgeDays.Hours() > 0 {
		uploadOptions = &uplink.UploadOptions{Expires: time.Now().Add(s.purgeDays)}
	}

	writer, err := s.project.UploadObject(fpath.WithTempData(ctx, "", true), s.bucket.Name, key, uploadOptions)
	if err != nil {
		return err
	}

	n, err := io.Copy(writer, reader)
	if err != nil || uint64(n) != contentLength {
		//Ignoring the error to return the one that occurred first, but try to clean up.
		_ = writer.Abort()
		return err
	}
	err = writer.SetCustomMetadata(ctx, uplink.CustomMetadata{"content-type": contentType})
	if err != nil {
		//Ignoring the error to return the one that occurred first, but try to clean up.
		_ = writer.Abort()
		return err
	}

	err = writer.Commit()
	return err
}

func (s *StorjStorage) IsRangeSupported() bool { return true }

// IsNotExist indicates if a file doesn't exist on storage
func (s *StorjStorage) IsNotExist(err error) bool {
	return errors.Is(err, uplink.ErrObjectNotFound)
}
