package storage

import (
	"context"
	"fmt"
	"io"
	"log"
	"time"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/awserr"
	"github.com/aws/aws-sdk-go/aws/credentials"
	"github.com/aws/aws-sdk-go/aws/session"
	"github.com/aws/aws-sdk-go/service/s3"
	"github.com/aws/aws-sdk-go/service/s3/s3manager"
)

// S3Storage is a storage backed by AWS S3
type S3Storage struct {
	Storage
	bucket      string
	session     *session.Session
	s3          *s3.S3
	logger      *log.Logger
	purgeDays   time.Duration
	noMultipart bool
}

// NewS3Storage is the factory for S3Storage
func NewS3Storage(accessKey, secretKey, bucketName string, purgeDays int, region, endpoint string, disableMultipart bool, forcePathStyle bool, logger *log.Logger) (*S3Storage, error) {
	sess := getAwsSession(accessKey, secretKey, region, endpoint, forcePathStyle)

	return &S3Storage{
		bucket:      bucketName,
		s3:          s3.New(sess),
		session:     sess,
		logger:      logger,
		noMultipart: disableMultipart,
		purgeDays:   time.Duration(purgeDays*24) * time.Hour,
	}, nil
}

// Type returns the storage type
func (s *S3Storage) Type() string {
	return "s3"
}

// Head retrieves content length of a file from storage
func (s *S3Storage) Head(ctx context.Context, token string, filename string) (contentLength uint64, err error) {
	key := fmt.Sprintf("%s/%s", token, filename)

	headRequest := &s3.HeadObjectInput{
		Bucket: aws.String(s.bucket),
		Key:    aws.String(key),
	}

	// content type , content length
	response, err := s.s3.HeadObjectWithContext(ctx, headRequest)
	if err != nil {
		return
	}

	if response.ContentLength != nil {
		contentLength = uint64(*response.ContentLength)
	}

	return
}

// Purge cleans up the storage
func (s *S3Storage) Purge(context.Context, time.Duration) (err error) {
	// NOOP expiration is set at upload time
	return nil
}

// IsNotExist indicates if a file doesn't exist on storage
func (s *S3Storage) IsNotExist(err error) bool {
	if err == nil {
		return false
	}

	if aerr, ok := err.(awserr.Error); ok {
		switch aerr.Code() {
		case s3.ErrCodeNoSuchKey:
			return true
		}
	}

	return false
}

// Get retrieves a file from storage
func (s *S3Storage) Get(ctx context.Context, token string, filename string, rng *Range) (reader io.ReadCloser, contentLength uint64, err error) {
	key := fmt.Sprintf("%s/%s", token, filename)

	getRequest := &s3.GetObjectInput{
		Bucket: aws.String(s.bucket),
		Key:    aws.String(key),
	}

	if rng != nil {
		getRequest.Range = aws.String(rng.Range())
	}

	response, err := s.s3.GetObjectWithContext(ctx, getRequest)
	if err != nil {
		return
	}

	if response.ContentLength != nil {
		contentLength = uint64(*response.ContentLength)
	}
	if rng != nil && response.ContentRange != nil {
		rng.SetContentRange(*response.ContentRange)
	}

	reader = response.Body
	return
}

// Delete removes a file from storage
func (s *S3Storage) Delete(ctx context.Context, token string, filename string) (err error) {
	metadata := fmt.Sprintf("%s/%s.metadata", token, filename)
	deleteRequest := &s3.DeleteObjectInput{
		Bucket: aws.String(s.bucket),
		Key:    aws.String(metadata),
	}

	_, err = s.s3.DeleteObjectWithContext(ctx, deleteRequest)
	if err != nil {
		return
	}

	key := fmt.Sprintf("%s/%s", token, filename)
	deleteRequest = &s3.DeleteObjectInput{
		Bucket: aws.String(s.bucket),
		Key:    aws.String(key),
	}

	_, err = s.s3.DeleteObjectWithContext(ctx, deleteRequest)

	return
}

// Put saves a file on storage
func (s *S3Storage) Put(ctx context.Context, token string, filename string, reader io.Reader, contentType string, _ uint64) (err error) {
	key := fmt.Sprintf("%s/%s", token, filename)

	s.logger.Printf("Uploading file %s to S3 Bucket", filename)
	var concurrency int
	if !s.noMultipart {
		concurrency = 20
	} else {
		concurrency = 1
	}

	// Create an uploader with the session and custom options
	uploader := s3manager.NewUploader(s.session, func(u *s3manager.Uploader) {
		u.Concurrency = concurrency // default is 5
		u.LeavePartsOnError = false
	})

	var expire *time.Time
	if s.purgeDays.Hours() > 0 {
		expire = aws.Time(time.Now().Add(s.purgeDays))
	}

	_, err = uploader.UploadWithContext(ctx, &s3manager.UploadInput{
		Bucket:      aws.String(s.bucket),
		Key:         aws.String(key),
		Body:        reader,
		Expires:     expire,
		ContentType: aws.String(contentType),
	})

	return
}

func (s *S3Storage) IsRangeSupported() bool { return true }

func getAwsSession(accessKey, secretKey, region, endpoint string, forcePathStyle bool) *session.Session {
	return session.Must(session.NewSession(&aws.Config{
		Region:           aws.String(region),
		Endpoint:         aws.String(endpoint),
		Credentials:      credentials.NewStaticCredentials(accessKey, secretKey, ""),
		S3ForcePathStyle: aws.Bool(forcePathStyle),
	}))
}
