package server

import (
	"bytes"
	"fmt"
	"io"
	"log"
	"mime"
	"os"
	"path/filepath"
	"strconv"
	"sync"

	"encoding/json"
	"github.com/goamz/goamz/s3"

	"golang.org/x/net/context"
	"golang.org/x/oauth2"
	"golang.org/x/oauth2/google"
	"google.golang.org/api/drive/v3"
	"google.golang.org/api/googleapi"
	"io/ioutil"
	"net/http"
	"strings"
)

type Storage interface {
	Get(token string, filename string) (reader io.ReadCloser, contentType string, contentLength uint64, err error)
	Head(token string, filename string) (contentType string, contentLength uint64, err error)
	Put(token string, filename string, reader io.Reader, contentType string, contentLength uint64) error
	Delete(token string, filename string) error
	IsNotExist(err error) bool

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

func (s *LocalStorage) Head(token string, filename string) (contentType string, contentLength uint64, err error) {
	path := filepath.Join(s.basedir, token, filename)

	var fi os.FileInfo
	if fi, err = os.Lstat(path); err != nil {
		return
	}

	contentLength = uint64(fi.Size())

	contentType = mime.TypeByExtension(filepath.Ext(filename))

	return
}

func (s *LocalStorage) Get(token string, filename string) (reader io.ReadCloser, contentType string, contentLength uint64, err error) {
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

	contentType = mime.TypeByExtension(filepath.Ext(filename))

	return
}

func (s *LocalStorage) Delete(token string, filename string) (err error) {
	metadata := filepath.Join(s.basedir, token, fmt.Sprintf("%s.metadata", filename))
	os.Remove(metadata)

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

func (s *LocalStorage) Put(token string, filename string, reader io.Reader, contentType string, contentLength uint64) error {
	var f io.WriteCloser
	var err error

	path := filepath.Join(s.basedir, token)

	if err = os.Mkdir(path, 0700); err != nil && !os.IsExist(err) {
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

type S3Storage struct {
	Storage
	bucket *s3.Bucket
	logger *log.Logger
}

func NewS3Storage(accessKey, secretKey, bucketName, endpoint string, logger *log.Logger) (*S3Storage, error) {
	bucket, err := getBucket(accessKey, secretKey, bucketName, endpoint)
	if err != nil {
		return nil, err
	}

	return &S3Storage{bucket: bucket, logger: logger}, nil
}

func (s *S3Storage) Type() string {
	return "s3"
}

func (s *S3Storage) Head(token string, filename string) (contentType string, contentLength uint64, err error) {
	key := fmt.Sprintf("%s/%s", token, filename)

	// content type , content length
	response, err := s.bucket.Head(key, map[string][]string{})
	if err != nil {
		return
	}

	contentType = response.Header.Get("Content-Type")

	contentLength, err = strconv.ParseUint(response.Header.Get("Content-Length"), 10, 0)
	if err != nil {
		return
	}

	return
}

func (s *S3Storage) IsNotExist(err error) bool {
	if err == nil {
		return false
	}

	s.logger.Printf("IsNotExist: %s, %#v", err.Error(), err)

	b := (err.Error() == "The specified key does not exist.")
	b = b || (err.Error() == "Access Denied")
	return b
}

func (s *S3Storage) Get(token string, filename string) (reader io.ReadCloser, contentType string, contentLength uint64, err error) {
	key := fmt.Sprintf("%s/%s", token, filename)

	// content type , content length
	response, err := s.bucket.GetResponse(key)
	if err != nil {
		return
	}

	contentType = response.Header.Get("Content-Type")
	contentLength, err = strconv.ParseUint(response.Header.Get("Content-Length"), 10, 0)
	if err != nil {
		return
	}

	reader = response.Body
	return
}

func (s *S3Storage) Delete(token string, filename string) (err error) {
	metadata := fmt.Sprintf("%s/%s.metadata", token, filename)
	s.bucket.Del(metadata)

	key := fmt.Sprintf("%s/%s", token, filename)
	err = s.bucket.Del(key)

	return
}

func (s *S3Storage) Put(token string, filename string, reader io.Reader, contentType string, contentLength uint64) (err error) {
	key := fmt.Sprintf("%s/%s", token, filename)

	var (
		multi *s3.Multi
		parts []s3.Part
	)

	if multi, err = s.bucket.InitMulti(key, contentType, s3.Private); err != nil {
		s.logger.Printf(err.Error())
		return
	}

	// 20 mb parts
	partsChan := make(chan interface{})
	// partsChan := make(chan s3.Part)

	go func() {
		// maximize to 20 threads
		sem := make(chan int, 20)
		index := 1
		var wg sync.WaitGroup

		for {
			// buffered in memory because goamz s3 multi needs seekable reader
			var (
				buffer []byte = make([]byte, (1<<20)*10)
				count  int
				err    error
			)

			// Amazon expects parts of at least 5MB, except for the last one
			if count, err = io.ReadAtLeast(reader, buffer, (1<<20)*5); err != nil && err != io.ErrUnexpectedEOF && err != io.EOF {
				s.logger.Printf(err.Error())
				return
			}

			// always send minimal 1 part
			if err == io.EOF && index > 1 {
				s.logger.Printf("Waiting for all parts to finish uploading.")

				// wait for all parts to be finished uploading
				wg.Wait()

				// and close the channel
				close(partsChan)

				return
			}

			wg.Add(1)

			sem <- 1

			// using goroutines because of retries when upload fails
			go func(multi *s3.Multi, buffer []byte, index int) {
				s.logger.Printf("Uploading part %d %d", index, len(buffer))

				defer func() {
					s.logger.Printf("Finished part %d %d", index, len(buffer))

					wg.Done()

					<-sem
				}()

				partReader := bytes.NewReader(buffer)

				var part s3.Part

				if part, err = multi.PutPart(index, partReader); err != nil {
					s.logger.Printf("Error while uploading part %d %d %s", index, len(buffer), err.Error())
					partsChan <- err
					return
				}

				s.logger.Printf("Finished uploading part %d %d", index, len(buffer))

				partsChan <- part

			}(multi, buffer[:count], index)

			index++
		}
	}()

	// wait for all parts to be uploaded
	for part := range partsChan {
		switch part.(type) {
		case s3.Part:
			parts = append(parts, part.(s3.Part))
		case error:
			// abort multi upload
			s.logger.Printf("Error during upload, aborting %s.", part.(error).Error())
			err = part.(error)

			multi.Abort()
			return
		}

	}

	s.logger.Printf("Completing upload %d parts", len(parts))

	if err = multi.Complete(parts); err != nil {
		s.logger.Printf("Error during completing upload %d parts %s", len(parts), err.Error())
		return
	}

	s.logger.Printf("Completed uploading %d", len(parts))

	return
}

type GDrive struct {
	service         *drive.Service
	rootId          string
	basedir         string
	localConfigPath string
	logger          *log.Logger
}

func NewGDriveStorage(clientJsonFilepath string, localConfigPath string, basedir string, logger *log.Logger) (*GDrive, error) {
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

	storage := &GDrive{service: srv, basedir: basedir, rootId: "", localConfigPath: localConfigPath, logger: logger}
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
	for 0 < len(l.Files) {
		if err != nil {
			return "", err
		}

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

	for 0 < len(l.Files) {
		if err != nil {
			return "", err
		}

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

func (s *GDrive) Head(token string, filename string) (contentType string, contentLength uint64, err error) {
	var fileId string
	fileId, err = s.findId(filename, token)
	if err != nil {
		return
	}

	var fi *drive.File
	if fi, err = s.service.Files.Get(fileId).Fields("mimeType", "size").Do(); err != nil {
		return
	}

	contentLength = uint64(fi.Size)

	contentType = fi.MimeType

	return
}

func (s *GDrive) Get(token string, filename string) (reader io.ReadCloser, contentType string, contentLength uint64, err error) {
	var fileId string
	fileId, err = s.findId(filename, token)
	if err != nil {
		return
	}

	var fi *drive.File
	fi, err = s.service.Files.Get(fileId).Fields("mimeType", "size", "md5Checksum").Do()
	if !s.hasChecksum(fi) {
		err = fmt.Errorf("Cannot find file %s/%s", token, filename)
		return
	}

	contentLength = uint64(fi.Size)
	contentType = fi.MimeType

	ctx := context.Background()
	var res *http.Response
	res, err = s.service.Files.Get(fileId).Context(ctx).Download()
	if err != nil {
		return
	}

	reader = res.Body

	return
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
	if err == nil {
		return false
	}

	if err != nil {
		if e, ok := err.(*googleapi.Error); ok {
			return e.Code == http.StatusNotFound
		}
	}

	return false
}

func (s *GDrive) Put(token string, filename string, reader io.Reader, contentType string, contentLength uint64) error {
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
		MimeType: contentType,
	}

	ctx := context.Background()
	_, err = s.service.Files.Create(dst).Context(ctx).Media(reader).Do()

	if err != nil {
		return err
	}

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
