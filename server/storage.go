package server

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"io/ioutil"
	"log"
	"net/http"
	"os"
	"path/filepath"
	"strconv"
	"strings"
	"time"

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

type LocalStorage struct {
	Storage
	basedir string
	logger  *log.Logger
}

func NewLocalStorage(basedir string, logger *log.Logger) (*LocalStorage, error) {
	return &LocalStorage{basedir: basedir, logger: logger}, nil
}

func (s *LocalStorage) Type() string {
	return "local"
}

func (s *LocalStorage) Get(token string, filename string) (reader io.ReadCloser, metadata Metadata, err error) {
	path := filepath.Join(s.basedir, token, filename)

	// content type , content length
	reader, err = os.Open(path)
	if err != nil {
		return nil, Metadata{}, err
	}

	metadata, err = s.Head(token, filename)
	if err != nil {
		return nil, Metadata{}, err
	}
	return reader, metadata, nil
}

func (s *LocalStorage) Head(token string, filename string) (metadata Metadata, err error) {
	path := filepath.Join(s.basedir, token, filename)

	fi, err := os.Open(path)
	if err != nil {
		return
	}

	err = json.NewDecoder(fi).Decode(&metadata)
	if err != nil {
		return Metadata{}, err
	}
	return metadata, nil
}

func (s *LocalStorage) Meta(token string, filename string, metadata Metadata) error {
	return s.putMetadata(token, filename, metadata)
}

func (s *LocalStorage) Put(token string, filename string, reader io.Reader, metadata Metadata) error {
	err := s.putMetadata(token, filename, metadata)
	if err != nil {
		return err
	}

	err = s.put(token, filename, reader)
	if err != nil {
		//Delete the metadata if the put failed
		_ = s.Delete(token, fmt.Sprintf("%s.metadata", filename))
	}
	return err
}

func (s *LocalStorage) put(token string, filename string, reader io.Reader) error {
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

	_, err = io.Copy(f, reader)
	return err
}

func (s *LocalStorage) putMetadata(token string, filename string, metadata Metadata) error {
	buffer := &bytes.Buffer{}
	if err := json.NewEncoder(buffer).Encode(metadata); err != nil {
		log.Printf("%s", err.Error())
		return err
	} else if err := s.put(token, filename, buffer); err != nil {
		log.Printf("%s", err.Error())

		return nil
	}
	return nil
}

func (s *LocalStorage) Delete(token string, filename string) (err error) {
	metadata := filepath.Join(s.basedir, token, fmt.Sprintf("%s.metadata", filename))
	_ = os.Remove(metadata)

	path := filepath.Join(s.basedir, token, filename)
	err = os.Remove(path)
	return
}

func (s *LocalStorage) IsNotExist(err error) bool {
	if err == nil {
		return false
	}

	return os.IsNotExist(err)
}

func (s *LocalStorage) DeleteExpired() error {
	return nil
}

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

type GDrive struct {
	service         *drive.Service
	rootId          string
	basedir         string
	localConfigPath string
	chunkSize       int
	logger          *log.Logger
}

func NewGDriveStorage(clientJsonFilepath string, localConfigPath string, basedir string, chunkSize int, logger *log.Logger) (*GDrive, error) {
	b, err := ioutil.ReadFile(clientJsonFilepath)
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
	storage := &GDrive{service: srv, basedir: basedir, rootId: "", localConfigPath: localConfigPath, chunkSize: chunkSize, logger: logger}
	err = storage.setupRoot()
	if err != nil {
		return nil, err
	}

	return storage, nil
}

const GDriveRootConfigFile = "root_id.conf"
const GDriveTokenJsonFile = "token.json"
const GDriveDirectoryMimeType = "application/vnd.google-apps.folder"

func (s *GDrive) setupRoot() error {
	rootFileConfig := filepath.Join(s.localConfigPath, GDriveRootConfigFile)

	rootId, err := ioutil.ReadFile(rootFileConfig)
	if err != nil && !os.IsNotExist(err) {
		return err
	}

	if string(rootId) != "" {
		s.rootId = string(rootId)
		return nil
	}

	dir := &drive.File{
		Name:     s.basedir,
		MimeType: GDriveDirectoryMimeType,
	}

	di, err := s.service.Files.Create(dir).Fields("id").Do()
	if err != nil {
		return err
	}

	s.rootId = di.Id
	err = ioutil.WriteFile(rootFileConfig, []byte(s.rootId), os.FileMode(0600))
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

func (s *GDrive) findId(filename string, token string) (string, error) {
	filename = strings.Replace(filename, `'`, `\'`, -1)
	filename = strings.Replace(filename, `"`, `\"`, -1)

	fileId, tokenId, nextPageToken := "", "", ""

	q := fmt.Sprintf("'%s' in parents and name='%s' and mimeType='%s' and trashed=false", s.rootId, token, GDriveDirectoryMimeType)
	l, err := s.list(nextPageToken, q)
	if err != nil {
		return "", err
	}

	for 0 < len(l.Files) {
		for _, fi := range l.Files {
			tokenId = fi.Id
			break
		}

		if l.NextPageToken == "" {
			break
		}

		l, err = s.list(l.NextPageToken, q)
	}

	if filename == "" {
		return tokenId, nil
	} else if tokenId == "" {
		return "", fmt.Errorf("Cannot find file %s/%s", token, filename)
	}

	q = fmt.Sprintf("'%s' in parents and name='%s' and mimeType!='%s' and trashed=false", tokenId, filename, GDriveDirectoryMimeType)
	l, err = s.list(nextPageToken, q)
	if err != nil {
		return "", err
	}

	for 0 < len(l.Files) {
		for _, fi := range l.Files {

			fileId = fi.Id
			break
		}

		if l.NextPageToken == "" {
			break
		}

		l, err = s.list(l.NextPageToken, q)
	}

	if fileId == "" {
		return "", fmt.Errorf("Cannot find file %s/%s", token, filename)
	}

	return fileId, nil
}

func (s *GDrive) Type() string {
	return "gdrive"
}

func (s *GDrive) Get(token string, filename string) (reader io.ReadCloser, metadata Metadata, err error) {
	var fileId string
	fileId, err = s.findId(filename, token)
	if err != nil {
		return
	}

	var fi *drive.File
	fi, err = s.service.Files.Get(fileId).Do()
	if !s.hasChecksum(fi) {
		err = fmt.Errorf("Cannot find file %s/%s", token, filename)
		return
	}
	if err != nil {
		return nil, Metadata{}, err
	}

	downloads, err := strconv.Atoi(fi.Properties["downloads"])
	if err != nil {
		return nil, Metadata{}, err
	}
	maxdownloads, err := strconv.Atoi(fi.Properties["maxDownloads"])
	if err != nil {
		return nil, Metadata{}, err
	}
	expires, err := time.Parse("2020-02-02 02:02:02", fi.Properties["expires"])
	if err != nil {
		return nil, Metadata{}, err
	}

	metadata = Metadata{
		ContentType:   "",
		ContentLength: fi.Size,
		Downloads:     downloads,
		MaxDownloads:  maxdownloads,
		MaxDate:       expires,
		DeletionToken: fi.Properties["deletionToken"],
		Secret:        fi.Properties["deletionSecret"],
	}

	ctx := context.Background()
	var res *http.Response
	res, err = s.service.Files.Get(fileId).Context(ctx).Download()
	if err != nil {
		return
	}

	reader = res.Body

	return
}

func (s *GDrive) Head(token string, filename string) (metadata Metadata, err error) {
	var fileId string
	fileId, err = s.findId(filename, token)
	if err != nil {
		return
	}

	var fi *drive.File
	if fi, err = s.service.Files.Get(fileId).Do(); err != nil {
		return
	}

	downloads, err := strconv.Atoi(fi.Properties["downloads"])
	if err != nil {
		return Metadata{}, err
	}
	maxdownloads, err := strconv.Atoi(fi.Properties["maxDownloads"])
	if err != nil {
		return Metadata{}, err
	}
	expires, err := time.Parse("2020-02-02 02:02:02", fi.Properties["expires"])
	if err != nil {
		return Metadata{}, err
	}

	metadata = Metadata{
		ContentType:   "",
		ContentLength: fi.Size,
		Downloads:     downloads,
		MaxDownloads:  maxdownloads,
		MaxDate:       expires,
		DeletionToken: fi.Properties["deletionToken"],
		Secret:        fi.Properties["deletionSecret"],
	}

	return
}

func (s *GDrive) Meta(token string, filename string, metadata Metadata) error {
	return nil
}

func (s *GDrive) Put(token string, filename string, reader io.Reader, metadata Metadata) error {
	dirId, err := s.findId("", token)
	if err != nil {
		return err
	}

	if dirId == "" {
		dir := &drive.File{
			Name:     token,
			Parents:  []string{s.rootId},
			MimeType: GDriveDirectoryMimeType,
		}

		di, err := s.service.Files.Create(dir).Fields("id").Do()
		if err != nil {
			return err
		}

		dirId = di.Id
	}

	// Instantiate empty drive file
	dst := &drive.File{
		Name:     filename,
		Parents:  []string{dirId},
		MimeType: metadata.ContentType,
		Properties: map[string]string{
			"downloads":      strconv.Itoa(metadata.Downloads),
			"maxDownloads":   strconv.Itoa(metadata.MaxDownloads),
			"deletionToken":  metadata.DeletionToken,
			"deletionSecret": metadata.Secret,
			"expires":        metadata.MaxDate.String(),
		},
	}

	ctx := context.Background()
	_, err = s.service.Files.Create(dst).Context(ctx).Media(reader, googleapi.ChunkSize(s.chunkSize)).Do()

	if err != nil {
		return err
	}

	return nil
}

func (s *GDrive) Delete(token string, filename string) (err error) {
	metadata, _ := s.findId(fmt.Sprintf("%s.metadata", filename), token)
	s.service.Files.Delete(metadata).Do()

	var fileId string
	fileId, err = s.findId(filename, token)
	if err != nil {
		return
	}

	err = s.service.Files.Delete(fileId).Do()
	return
}

func (s *GDrive) IsNotExist(err error) bool {
	if err != nil {
		if e, ok := err.(*googleapi.Error); ok {
			return e.Code == http.StatusNotFound
		}
	}

	return false
}

func (s *GDrive) DeleteExpired() error {
	return nil
}

// Retrieve a token, saves the token, then returns the generated client.
func getGDriveClient(config *oauth2.Config, localConfigPath string, logger *log.Logger) *http.Client {
	tokenFile := filepath.Join(localConfigPath, GDriveTokenJsonFile)
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
