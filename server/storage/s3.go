package storage

import (
	"context"
	"errors"
	"fmt"
	"io"
	"log"
	"time"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/config"
	"github.com/aws/aws-sdk-go-v2/credentials"
	"github.com/aws/aws-sdk-go-v2/feature/s3/transfermanager"
	"github.com/aws/aws-sdk-go-v2/service/s3"
	"github.com/aws/aws-sdk-go-v2/service/s3/types"
)

const (
	s3CredentialsTypeLegacy                    = "legacy"
	s3CredentialsTypeDefaultSDKCredentialChain = "default-sdk-credential-chain"
)

// S3Storage is a storage backed by AWS S3
type S3Storage struct {
	Storage
	bucket      string
	s3          *s3.Client
	logger      *log.Logger
	purgeDays   time.Duration
	noMultipart bool
}

// NewS3Storage is the factory for S3Storage
func NewS3Storage(ctx context.Context, credentialsType, accessKey, secretKey, bucketName string, purgeDays int, region, endpoint string, disableMultipart bool, forcePathStyle bool, logger *log.Logger) (*S3Storage, error) {
	cfg, err := getAwsConfig(ctx, credentialsType, accessKey, secretKey)
	if err != nil {
		return nil, err
	}

	client := s3.NewFromConfig(cfg, func(o *s3.Options) {
		o.Region = region
		o.UsePathStyle = forcePathStyle
		if len(endpoint) > 0 {
			o.BaseEndpoint = aws.String(endpoint)
		}
	})

	return &S3Storage{
		bucket:      bucketName,
		s3:          client,
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
	response, err := s.s3.HeadObject(ctx, headRequest)
	if err != nil {
		return
	}

	contentLength = uint64(aws.ToInt64(response.ContentLength))

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

	var nkerr *types.NoSuchKey
	return errors.As(err, &nkerr)
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

	response, err := s.s3.GetObject(ctx, getRequest)
	if err != nil {
		return
	}

	contentLength = uint64(aws.ToInt64(response.ContentLength))
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

	_, err = s.s3.DeleteObject(ctx, deleteRequest)
	if err != nil {
		return
	}

	key := fmt.Sprintf("%s/%s", token, filename)
	deleteRequest = &s3.DeleteObjectInput{
		Bucket: aws.String(s.bucket),
		Key:    aws.String(key),
	}

	_, err = s.s3.DeleteObject(ctx, deleteRequest)

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

	// Create a transfer manager with custom options; failed multipart uploads are always aborted
	uploader := transfermanager.New(s.s3, func(o *transfermanager.Options) {
		o.Concurrency = concurrency // default is 5
		// transfermanager does not inherit the client setting, keep S3-compatible backends working
		o.RequestChecksumCalculation = aws.RequestChecksumCalculationWhenRequired
	})

	var expire *time.Time
	if s.purgeDays.Hours() > 0 {
		expire = aws.Time(time.Now().Add(s.purgeDays))
	}

	_, err = uploader.UploadObject(ctx, &transfermanager.UploadObjectInput{
		Bucket:      aws.String(s.bucket),
		Key:         aws.String(key),
		Body:        reader,
		Expires:     expire,
		ContentType: aws.String(contentType),
	})

	return
}

func (s *S3Storage) IsRangeSupported() bool { return true }

func getAwsConfig(ctx context.Context, credentialsType, accessKey, secretKey string) (aws.Config, error) {
	options := []func(*config.LoadOptions) error{
		config.WithRequestChecksumCalculation(aws.RequestChecksumCalculationWhenRequired),
		config.WithResponseChecksumValidation(aws.ResponseChecksumValidationWhenRequired),
	}

	switch credentialsType {
	case s3CredentialsTypeLegacy:
		if accessKey == "" {
			return aws.Config{}, errors.New("access-key not set")
		}
		if secretKey == "" {
			return aws.Config{}, errors.New("secret-key not set")
		}
		options = append(options, config.WithCredentialsProvider(credentials.StaticCredentialsProvider{
			Value: aws.Credentials{
				AccessKeyID:     accessKey,
				SecretAccessKey: secretKey,
				SessionToken:    "",
			},
		}))
	case s3CredentialsTypeDefaultSDKCredentialChain:
	default:
		return aws.Config{}, fmt.Errorf("unsupported S3 credentials type %q", credentialsType)
	}

	return config.LoadDefaultConfig(ctx, options...)
}
