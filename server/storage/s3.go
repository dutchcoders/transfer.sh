package storage

import (
	"fmt"
	"io"
	"log"
	"strconv"
	"time"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/awserr"
	"github.com/aws/aws-sdk-go/aws/credentials"
	"github.com/aws/aws-sdk-go/aws/session"
	"github.com/aws/aws-sdk-go/service/s3"
	"github.com/aws/aws-sdk-go/service/s3/s3manager"
)

type S3Storage struct {
	Storage
	bucket      string
	session     *session.Session
	s3          *s3.S3
	logger      *log.Logger
	noMultipart bool
}

func NewS3Storage(accessKey, secretKey, bucketName, region, endpoint string, logger *log.Logger, disableMultipart bool, forcePathStyle bool) (*S3Storage, error) {
	sess := getAwsSession(accessKey, secretKey, region, endpoint, forcePathStyle)

	return &S3Storage{bucket: bucketName, s3: s3.New(sess), session: sess, logger: logger, noMultipart: disableMultipart}, nil
}

func (s *S3Storage) Type() string {
	return "s3"
}

func (s *S3Storage) Head(token string, filename string) (metadata Metadata, err error) {
	key := fmt.Sprintf("%s/%s", token, filename)

	headRequest := &s3.HeadObjectInput{
		Bucket: aws.String(s.bucket),
		Key:    aws.String(key),
	}

	// content type , content length
	response, err := s.s3.HeadObject(headRequest)
	if err != nil {
		return Metadata{}, err
	}

	downloads, err := strconv.Atoi(*response.Metadata["downloads"])
	if err != nil {
		return Metadata{}, err
	}
	maxdownloads, err := strconv.Atoi(*response.Metadata["maxDownloads"])
	if err != nil {
		return Metadata{}, err
	}
	expires, err := time.Parse("2020-02-02 02:02:02", *response.Expires)
	if err != nil {
		return Metadata{}, err
	}

	metadata = Metadata{
		ContentType:   "",
		ContentLength: *response.ContentLength,
		Downloads:     downloads,
		MaxDownloads:  maxdownloads,
		MaxDate:       expires,
		DeletionToken: *response.Metadata["deletionToken"],
		Secret:        *response.Metadata["deletionSecret"],
	}
	return metadata, nil
}

func (s *S3Storage) Meta(token string, filename string, metadata Metadata) error {
	key := fmt.Sprintf("%s/%s", token, filename)

	input := &s3.CopyObjectInput{
		Bucket:            aws.String(s.bucket),
		CopySource:        aws.String(key),
		Key:               aws.String(key),
		MetadataDirective: aws.String("REPLACE"),
		Metadata: map[string]*string{
			"downloads":      aws.String(strconv.Itoa(metadata.Downloads)),
			"maxDownloads":   aws.String(strconv.Itoa(metadata.MaxDownloads)),
			"deletionToken":  aws.String(metadata.DeletionToken),
			"deletionSecret": aws.String(metadata.Secret),
		},
		ContentType: aws.String(metadata.ContentType),
		Expires:     aws.Time(metadata.MaxDate),
	}

	_, err := s.s3.CopyObject(input)
	if err != nil {
		return err
	}
	return nil
}

func (s *S3Storage) Get(token string, filename string) (reader io.ReadCloser, metadata Metadata, err error) {
	key := fmt.Sprintf("%s/%s", token, filename)

	getRequest := &s3.GetObjectInput{
		Bucket: aws.String(s.bucket),
		Key:    aws.String(key),
	}

	response, err := s.s3.GetObject(getRequest)
	if err != nil {
		return
	}

	downloads, err := strconv.Atoi(*response.Metadata["downloads"])
	if err != nil {
		return nil, Metadata{}, err
	}
	maxdownloads, err := strconv.Atoi(*response.Metadata["maxDownloads"])
	if err != nil {
		return nil, Metadata{}, err
	}
	expires, err := time.Parse("2020-02-02 02:02:02", *response.Expires)
	if err != nil {
		return nil, Metadata{}, err
	}

	metadata = Metadata{
		ContentType:   "",
		ContentLength: *response.ContentLength,
		Downloads:     downloads,
		MaxDownloads:  maxdownloads,
		MaxDate:       expires,
		DeletionToken: *response.Metadata["deletionToken"],
		Secret:        *response.Metadata["deletionSecret"],
	}

	reader = response.Body
	return
}

func (s *S3Storage) Delete(token string, filename string) (err error) {
	metadata := fmt.Sprintf("%s/%s.metadata", token, filename)
	deleteRequest := &s3.DeleteObjectInput{
		Bucket: aws.String(s.bucket),
		Key:    aws.String(metadata),
	}

	_, err = s.s3.DeleteObject(deleteRequest)
	if err != nil {
		return
	}

	key := fmt.Sprintf("%s/%s", token, filename)
	deleteRequest = &s3.DeleteObjectInput{
		Bucket: aws.String(s.bucket),
		Key:    aws.String(key),
	}

	_, err = s.s3.DeleteObject(deleteRequest)

	return
}

func (s *S3Storage) Put(token string, filename string, reader io.Reader, metadata Metadata) (err error) {
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

	_, err = uploader.Upload(&s3manager.UploadInput{
		Bucket: aws.String(s.bucket),
		Key:    aws.String(key),
		Body:   reader,
		Metadata: map[string]*string{
			"downloads":      aws.String(strconv.Itoa(metadata.Downloads)),
			"maxDownloads":   aws.String(strconv.Itoa(metadata.MaxDownloads)),
			"deletionToken":  aws.String(metadata.DeletionToken),
			"deletionSecret": aws.String(metadata.Secret),
		},
		ContentType: aws.String(metadata.ContentType),
		Expires:     aws.Time(metadata.MaxDate),
	})

	return
}

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

func (s *S3Storage) DeleteExpired() error {
	// not necessary, as S3 has expireDate on files to automatically delete the them
	return nil
}

func getAwsSession(accessKey, secretKey, region, endpoint string, forcePathStyle bool) *session.Session {
	return session.Must(session.NewSession(&aws.Config{
		Region:           aws.String(region),
		Endpoint:         aws.String(endpoint),
		Credentials:      credentials.NewStaticCredentials(accessKey, secretKey, ""),
		S3ForcePathStyle: aws.Bool(forcePathStyle),
	}))
}
