package server

import (
	"encoding/json"
	"errors"
	"fmt"
	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/awserr"
	"github.com/aws/aws-sdk-go/aws/session"
	"github.com/aws/aws-sdk-go/service/s3"
	"github.com/aws/aws-sdk-go/service/s3/s3manager"
	"golang.org/x/net/context"
	"golang.org/x/oauth2"
	"golang.org/x/oauth2/google"
	"google.golang.org/api/drive/v3"
	"google.golang.org/api/googleapi"
	"io"
	"io/ioutil"
	"log"
	"net/http"
	"os"
	"path/filepath"
	"strings"
	"time"

	"storj.io/common/storj"
	"storj.io/uplink"
)

// Storage is the interface for storage operation
type Storage interface {
	// Get retrieves a file from storage
	Get(token string, filename string) (reader io.ReadCloser, contentLength uint64, err error)
	// Head retrieves content length of a file from storage
	Head(token string, filename string) (contentLength uint64, err error)
	// Put saves a file on storage
	Put(token string, filename string, reader io.Reader, contentType string, contentLength uint64) error
	// Delete removes a file from storage
	Delete(token string, filename string) error
	// IsNotExist indicates if a file doesn't exist on storage
	IsNotExist(err error) bool
	// Purge cleans up the storage
	Purge(days time.Duration) error

	// Type returns the storage type
	Type() string
}

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
func (s *LocalStorage) Head(token string, filename string) (contentLength uint64, err error) {
	path := filepath.Join(s.basedir, token, filename)

	var fi os.FileInfo
	if fi, err = os.Lstat(path); err != nil {
		return
	}

	contentLength = uint64(fi.Size())

	return
}

// Get retrieves a file from storage
func (s *LocalStorage) Get(token string, filename string) (reader io.ReadCloser, contentLength uint64, err error) {
	path := filepath.Join(s.basedir, token, filename)

	// content type , content length
	if reader, err = os.Open(path); err != nil {
		return
	}

	var fi os.FileInfo
	if fi, err = os.Lstat(path); err != nil {
		return
	}

	contentLength = uint64(fi.Size())

	return
}

// Delete removes a file from storage
func (s *LocalStorage) Delete(token string, filename string) (err error) {
	metadata := filepath.Join(s.basedir, token, fmt.Sprintf("%s.metadata", filename))
	os.Remove(metadata)

	path := filepath.Join(s.basedir, token, filename)
	err = os.Remove(path)
	return
}

// Purge cleans up the storage
func (s *LocalStorage) Purge(days time.Duration) (err error) {
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
func (s *LocalStorage) Put(token string, filename string, reader io.Reader, contentType string, contentLength uint64) error {
	var f io.WriteCloser
	var err error

	path := filepath.Join(s.basedir, token)

	if err = os.MkdirAll(path, 0700); err != nil && !os.IsExist(err) {
		return err
	}

	if f, err = os.OpenFile(filepath.Join(path, filename), os.O_RDWR|os.O_CREATE|os.O_TRUNC, 0600); err != nil {
		return err
	}

	defer f.Close()

	if _, err = io.Copy(f, reader); err != nil {
		return err
	}

	return nil
}

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
func (s *S3Storage) Head(token string, filename string) (contentLength uint64, err error) {
	key := fmt.Sprintf("%s/%s", token, filename)

	headRequest := &s3.HeadObjectInput{
		Bucket: aws.String(s.bucket),
		Key:    aws.String(key),
	}

	// content type , content length
	response, err := s.s3.HeadObject(headRequest)
	if err != nil {
		return
	}

	if response.ContentLength != nil {
		contentLength = uint64(*response.ContentLength)
	}

	return
}

// Purge cleans up the storage
func (s *S3Storage) Purge(days time.Duration) (err error) {
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
func (s *S3Storage) Get(token string, filename string) (reader io.ReadCloser, contentLength uint64, err error) {
	key := fmt.Sprintf("%s/%s", token, filename)

	getRequest := &s3.GetObjectInput{
		Bucket: aws.String(s.bucket),
		Key:    aws.String(key),
	}

	response, err := s.s3.GetObject(getRequest)
	if err != nil {
		return
	}

	if response.ContentLength != nil {
		contentLength = uint64(*response.ContentLength)
	}

	reader = response.Body
	return
}

// Delete removes a file from storage
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

// Put saves a file on storage
func (s *S3Storage) Put(token string, filename string, reader io.Reader, contentType string, contentLength uint64) (err error) {
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

	_, err = uploader.Upload(&s3manager.UploadInput{
		Bucket:  aws.String(s.bucket),
		Key:     aws.String(key),
		Body:    reader,
		Expires: expire,
	})

	return
}

// GDrive is a storage backed by GDrive
type GDrive struct {
	service         *drive.Service
	rootID          string
	basedir         string
	localConfigPath string
	chunkSize       int
	logger          *log.Logger
}

// NewGDriveStorage is the factory for GDrive
func NewGDriveStorage(clientJSONFilepath string, localConfigPath string, basedir string, chunkSize int, logger *log.Logger) (*GDrive, error) {
	b, err := ioutil.ReadFile(clientJSONFilepath)
	if err != nil {
		return nil, err
	}

	// If modifying these scopes, delete your previously saved client_secret.json.
	config, err := google.ConfigFromJSON(b, drive.DriveScope, drive.DriveMetadataScope)
	if err != nil {
		return nil, err
	}

	srv, err := drive.New(getGDriveClient(config, localConfigPath, logger))
	if err != nil {
		return nil, err
	}

	chunkSize = chunkSize * 1024 * 1024
	storage := &GDrive{service: srv, basedir: basedir, rootID: "", localConfigPath: localConfigPath, chunkSize: chunkSize, logger: logger}
	err = storage.setupRoot()
	if err != nil {
		return nil, err
	}

	return storage, nil
}

const gdriveRootConfigFile = "root_id.conf"
const gdriveTokenJSONFile = "token.json"
const gdriveDirectoryMimeType = "application/vnd.google-apps.folder"

func (s *GDrive) setupRoot() error {
	rootFileConfig := filepath.Join(s.localConfigPath, gdriveRootConfigFile)

	rootID, err := ioutil.ReadFile(rootFileConfig)
	if err != nil && !os.IsNotExist(err) {
		return err
	}

	if string(rootID) != "" {
		s.rootID = string(rootID)
		return nil
	}

	dir := &drive.File{
		Name:     s.basedir,
		MimeType: gdriveDirectoryMimeType,
	}

	di, err := s.service.Files.Create(dir).Fields("id").Do()
	if err != nil {
		return err
	}

	s.rootID = di.Id
	err = ioutil.WriteFile(rootFileConfig, []byte(s.rootID), os.FileMode(0600))
	if err != nil {
		return err
	}

	return nil
}

func (s *GDrive) hasChecksum(f *drive.File) bool {
	return f.Md5Checksum != ""
}

func (s *GDrive) list(nextPageToken string, q string) (*drive.FileList, error) {
	return s.service.Files.List().Fields("nextPageToken, files(id, name, mimeType)").Q(q).PageToken(nextPageToken).Do()
}

func (s *GDrive) findID(filename string, token string) (string, error) {
	filename = strings.Replace(filename, `'`, `\'`, -1)
	filename = strings.Replace(filename, `"`, `\"`, -1)

	fileID, tokenID, nextPageToken := "", "", ""

	q := fmt.Sprintf("'%s' in parents and name='%s' and mimeType='%s' and trashed=false", s.rootID, token, gdriveDirectoryMimeType)
	l, err := s.list(nextPageToken, q)
	if err != nil {
		return "", err
	}

	for 0 < len(l.Files) {
		for _, fi := range l.Files {
			tokenID = fi.Id
			break
		}

		if l.NextPageToken == "" {
			break
		}

		l, err = s.list(l.NextPageToken, q)
		if err != nil {
			return "", err
		}
	}

	if filename == "" {
		return tokenID, nil
	} else if tokenID == "" {
		return "", fmt.Errorf("Cannot find file %s/%s", token, filename)
	}

	q = fmt.Sprintf("'%s' in parents and name='%s' and mimeType!='%s' and trashed=false", tokenID, filename, gdriveDirectoryMimeType)
	l, err = s.list(nextPageToken, q)
	if err != nil {
		return "", err
	}

	for 0 < len(l.Files) {
		for _, fi := range l.Files {

			fileID = fi.Id
			break
		}

		if l.NextPageToken == "" {
			break
		}

		l, err = s.list(l.NextPageToken, q)
		if err != nil {
			return "", err
		}
	}

	if fileID == "" {
		return "", fmt.Errorf("Cannot find file %s/%s", token, filename)
	}

	return fileID, nil
}

// Type returns the storage type
func (s *GDrive) Type() string {
	return "gdrive"
}

// Head retrieves content length of a file from storage
func (s *GDrive) Head(token string, filename string) (contentLength uint64, err error) {
	var fileID string
	fileID, err = s.findID(filename, token)
	if err != nil {
		return
	}

	var fi *drive.File
	if fi, err = s.service.Files.Get(fileID).Fields("size").Do(); err != nil {
		return
	}

	contentLength = uint64(fi.Size)

	return
}

// Get retrieves a file from storage
func (s *GDrive) Get(token string, filename string) (reader io.ReadCloser, contentLength uint64, err error) {
	var fileID string
	fileID, err = s.findID(filename, token)
	if err != nil {
		return
	}

	var fi *drive.File
	fi, err = s.service.Files.Get(fileID).Fields("size", "md5Checksum").Do()
	if !s.hasChecksum(fi) {
		err = fmt.Errorf("Cannot find file %s/%s", token, filename)
		return
	}

	contentLength = uint64(fi.Size)

	ctx := context.Background()
	var res *http.Response
	res, err = s.service.Files.Get(fileID).Context(ctx).Download()
	if err != nil {
		return
	}

	reader = res.Body

	return
}

// Delete removes a file from storage
func (s *GDrive) Delete(token string, filename string) (err error) {
	metadata, _ := s.findID(fmt.Sprintf("%s.metadata", filename), token)
	s.service.Files.Delete(metadata).Do()

	var fileID string
	fileID, err = s.findID(filename, token)
	if err != nil {
		return
	}

	err = s.service.Files.Delete(fileID).Do()
	return
}

// Purge cleans up the storage
func (s *GDrive) Purge(days time.Duration) (err error) {
	nextPageToken := ""

	expirationDate := time.Now().Add(-1 * days).Format(time.RFC3339)
	q := fmt.Sprintf("'%s' in parents and modifiedTime < '%s' and mimeType!='%s' and trashed=false", s.rootID, expirationDate, gdriveDirectoryMimeType)
	l, err := s.list(nextPageToken, q)
	if err != nil {
		return err
	}

	for 0 < len(l.Files) {
		for _, fi := range l.Files {
			err = s.service.Files.Delete(fi.Id).Do()
			if err != nil {
				return
			}
		}

		if l.NextPageToken == "" {
			break
		}

		l, err = s.list(l.NextPageToken, q)
		if err != nil {
			return
		}
	}

	return
}

// IsNotExist indicates if a file doesn't exist on storage
func (s *GDrive) IsNotExist(err error) bool {
	if err == nil {
		return false
	}

	if e, ok := err.(*googleapi.Error); ok {
		return e.Code == http.StatusNotFound
	}

	return false
}

// Put saves a file on storage
func (s *GDrive) Put(token string, filename string, reader io.Reader, contentType string, contentLength uint64) error {
	dirID, err := s.findID("", token)
	if err != nil {
		return err
	}

	if dirID == "" {
		dir := &drive.File{
			Name:     token,
			Parents:  []string{s.rootID},
			MimeType: gdriveDirectoryMimeType,
		}

		di, err := s.service.Files.Create(dir).Fields("id").Do()
		if err != nil {
			return err
		}

		dirID = di.Id
	}

	// Instantiate empty drive file
	dst := &drive.File{
		Name:     filename,
		Parents:  []string{dirID},
		MimeType: contentType,
	}

	ctx := context.Background()
	_, err = s.service.Files.Create(dst).Context(ctx).Media(reader, googleapi.ChunkSize(s.chunkSize)).Do()

	if err != nil {
		return err
	}

	return nil
}

// Retrieve a token, saves the token, then returns the generated client.
func getGDriveClient(config *oauth2.Config, localConfigPath string, logger *log.Logger) *http.Client {
	tokenFile := filepath.Join(localConfigPath, gdriveTokenJSONFile)
	tok, err := gDriveTokenFromFile(tokenFile)
	if err != nil {
		tok = getGDriveTokenFromWeb(config, logger)
		saveGDriveToken(tokenFile, tok, logger)
	}

	return config.Client(context.Background(), tok)
}

// Request a token from the web, then returns the retrieved token.
func getGDriveTokenFromWeb(config *oauth2.Config, logger *log.Logger) *oauth2.Token {
	authURL := config.AuthCodeURL("state-token", oauth2.AccessTypeOffline)
	fmt.Printf("Go to the following link in your browser then type the "+
		"authorization code: \n%v\n", authURL)

	var authCode string
	if _, err := fmt.Scan(&authCode); err != nil {
		logger.Fatalf("Unable to read authorization code %v", err)
	}

	tok, err := config.Exchange(context.TODO(), authCode)
	if err != nil {
		logger.Fatalf("Unable to retrieve token from web %v", err)
	}
	return tok
}

// Retrieves a token from a local file.
func gDriveTokenFromFile(file string) (*oauth2.Token, error) {
	f, err := os.Open(file)
	defer f.Close()
	if err != nil {
		return nil, err
	}
	tok := &oauth2.Token{}
	err = json.NewDecoder(f).Decode(tok)
	return tok, err
}

// Saves a token to a file path.
func saveGDriveToken(path string, token *oauth2.Token, logger *log.Logger) {
	logger.Printf("Saving credential file to: %s\n", path)
	f, err := os.OpenFile(path, os.O_RDWR|os.O_CREATE|os.O_TRUNC, 0600)
	defer f.Close()
	if err != nil {
		logger.Fatalf("Unable to cache oauth token: %v", err)
	}

	json.NewEncoder(f).Encode(token)
}

// StorjStorage is a storage backed by Storj
type StorjStorage struct {
	Storage
	project   *uplink.Project
	bucket    *uplink.Bucket
	purgeDays time.Duration
	logger    *log.Logger
}

// NewStorjStorage is the factory for StorjStorage
func NewStorjStorage(access, bucket string, purgeDays int, logger *log.Logger) (*StorjStorage, error) {
	var instance StorjStorage
	var err error

	ctx := context.TODO()

	parsedAccess, err := uplink.ParseAccess(access)
	if err != nil {
		return nil, err
	}

	instance.project, err = uplink.OpenProject(ctx, parsedAccess)
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
func (s *StorjStorage) Head(token string, filename string) (contentLength uint64, err error) {
	key := storj.JoinPaths(token, filename)

	ctx := context.TODO()

	obj, err := s.project.StatObject(ctx, s.bucket.Name, key)
	if err != nil {
		return 0, err
	}

	contentLength = uint64(obj.System.ContentLength)

	return
}

// Get retrieves a file from storage
func (s *StorjStorage) Get(token string, filename string) (reader io.ReadCloser, contentLength uint64, err error) {
	key := storj.JoinPaths(token, filename)

	s.logger.Printf("Getting file %s from Storj Bucket", filename)

	ctx := context.TODO()

	download, err := s.project.DownloadObject(ctx, s.bucket.Name, key, nil)
	if err != nil {
		return nil, 0, err
	}

	contentLength = uint64(download.Info().System.ContentLength)

	reader = download
	return
}

// Delete removes a file from storage
func (s *StorjStorage) Delete(token string, filename string) (err error) {
	key := storj.JoinPaths(token, filename)

	s.logger.Printf("Deleting file %s from Storj Bucket", filename)

	ctx := context.TODO()

	_, err = s.project.DeleteObject(ctx, s.bucket.Name, key)

	return
}

// Purge cleans up the storage
func (s *StorjStorage) Purge(days time.Duration) (err error) {
	// NOOP expiration is set at upload time
	return nil
}

// Put saves a file on storage
func (s *StorjStorage) Put(token string, filename string, reader io.Reader, contentType string, contentLength uint64) (err error) {
	key := storj.JoinPaths(token, filename)

	s.logger.Printf("Uploading file %s to Storj Bucket", filename)

	ctx := context.TODO()

	var uploadOptions *uplink.UploadOptions
	if s.purgeDays.Hours() > 0 {
		uploadOptions = &uplink.UploadOptions{Expires: time.Now().Add(s.purgeDays)}
	}

	writer, err := s.project.UploadObject(ctx, s.bucket.Name, key, uploadOptions)
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

// IsNotExist indicates if a file doesn't exist on storage
func (s *StorjStorage) IsNotExist(err error) bool {
	return errors.Is(err, uplink.ErrObjectNotFound)
}
