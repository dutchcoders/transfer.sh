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

	"github.com/goamz/goamz/s3"
	"encoding/json"

	"golang.org/x/oauth2"
	"golang.org/x/net/context"
	"golang.org/x/oauth2/google"
	"google.golang.org/api/drive/v3"
	"google.golang.org/api/googleapi"
	"net/http"
	"io/ioutil"
	"time"
)

type Storage interface {
	Get(token string, filename string) (reader io.ReadCloser, contentType string, contentLength uint64, err error)
	Head(token string, filename string) (contentType string, contentLength uint64, err error)
	Put(token string, filename string, reader io.Reader, contentType string, contentLength uint64) error
	IsNotExist(err error) bool

	Type() string
}

type LocalStorage struct {
	Storage
	basedir string
}

func NewLocalStorage(basedir string) (*LocalStorage, error) {
	return &LocalStorage{basedir: basedir}, nil
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
		fmt.Printf("%s", err)
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
}

func NewS3Storage(accessKey, secretKey, bucketName, endpoint string) (*S3Storage, error) {
	bucket, err := getBucket(accessKey, secretKey, bucketName, endpoint)
	if err != nil {
		return nil, err
	}

	return &S3Storage{bucket: bucket}, nil
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

	log.Printf("IsNotExist: %s, %#v", err.Error(), err)

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

func (s *S3Storage) Put(token string, filename string, reader io.Reader, contentType string, contentLength uint64) (err error) {
	key := fmt.Sprintf("%s/%s", token, filename)

	var (
		multi *s3.Multi
		parts []s3.Part
	)

	if multi, err = s.bucket.InitMulti(key, contentType, s3.Private); err != nil {
		log.Printf(err.Error())
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
				log.Printf(err.Error())
				return
			}

			// always send minimal 1 part
			if err == io.EOF && index > 1 {
				log.Printf("Waiting for all parts to finish uploading.")

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
				log.Printf("Uploading part %d %d", index, len(buffer))

				defer func() {
					log.Printf("Finished part %d %d", index, len(buffer))

					wg.Done()

					<-sem
				}()

				partReader := bytes.NewReader(buffer)

				var part s3.Part

				if part, err = multi.PutPart(index, partReader); err != nil {
					log.Printf("Error while uploading part %d %d %s", index, len(buffer), err.Error())
					partsChan <- err
					return
				}

				log.Printf("Finished uploading part %d %d", index, len(buffer))

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
			log.Printf("Error during upload, aborting %s.", part.(error).Error())
			err = part.(error)

			multi.Abort()
			return
		}

	}

	log.Printf("Completing upload %d parts", len(parts))

	if err = multi.Complete(parts); err != nil {
		log.Printf("Error during completing upload %d parts %s", len(parts), err.Error())
		return
	}

	log.Printf("Completed uploading %d", len(parts))

	return
}

type GDrive struct {
	service *drive.Service
	basedir string
}

func NewGDriveStorage(clientJsonFilepath string, basedir string) (*GDrive, error) {
	b, err := ioutil.ReadFile(clientJsonFilepath)
	if err != nil {
		return nil, err
	}

	// If modifying these scopes, delete your previously saved client_secret.json.
	config, err := google.ConfigFromJSON(b, drive.DriveScope, drive.DriveMetadataScope)
	if err != nil {
		return nil, err
	}

	srv, err := drive.New(getGDriveClient(config))
	if err != nil {
		return nil, err
	}

	return &GDrive{service: srv, basedir: basedir}, nil
}

const GDriveTimeoutTimerInterval = time.Second * 10
const GDriveDirectoryMimeType = "application/vnd.google-apps.folder"

type gDriveTimeoutReaderWrapper func(io.Reader) io.Reader

func (s *GDrive) getTimeoutReader(r io.Reader, cancel context.CancelFunc, timeout time.Duration) io.Reader {
	return &GDriveTimeoutReader{
		reader:         r,
		cancel:         cancel,
		mutex:          &sync.Mutex{},
		maxIdleTimeout: timeout,
	}
}

type GDriveTimeoutReader struct {
	reader         io.Reader
	cancel         context.CancelFunc
	lastActivity   time.Time
	timer          *time.Timer
	mutex          *sync.Mutex
	maxIdleTimeout time.Duration
	done           bool
}

func (r *GDriveTimeoutReader) Read(p []byte) (int, error) {
	if r.timer == nil {
		r.startTimer()
	}

	r.mutex.Lock()

	// Read
	n, err := r.reader.Read(p)

	r.lastActivity = time.Now()
	r.done = (err != nil)

	r.mutex.Unlock()

	if r.done {
		r.stopTimer()
	}

	return n, err
}

func (r *GDriveTimeoutReader) Close() error {
	return r.reader.(io.ReadCloser).Close()
}

func (r *GDriveTimeoutReader) startTimer() {
	r.mutex.Lock()
	defer r.mutex.Unlock()

	if !r.done {
		r.timer = time.AfterFunc(GDriveTimeoutTimerInterval, r.timeout)
	}
}

func (r *GDriveTimeoutReader) stopTimer() {
	r.mutex.Lock()
	defer r.mutex.Unlock()

	if r.timer != nil {
		r.timer.Stop()
	}
}

func (r *GDriveTimeoutReader) timeout() {
	r.mutex.Lock()

	if r.done {
		r.mutex.Unlock()
		return
	}

	if time.Since(r.lastActivity) > r.maxIdleTimeout {
		r.cancel()
		r.mutex.Unlock()
		return
	}

	r.mutex.Unlock()
	r.startTimer()
}

func (s *GDrive) getTimeoutReaderWrapperContext(timeout time.Duration) (gDriveTimeoutReaderWrapper, context.Context) {
	ctx, cancel := context.WithCancel(context.TODO())
	wrapper := func(r io.Reader) io.Reader {
		// Return untouched reader if timeout is 0
		if timeout == 0 {
			return r
		}

		return s.getTimeoutReader(r, cancel, timeout)
	}
	return wrapper, ctx
}

func (s *GDrive) hasChecksum(f *drive.File) bool {
	return f.Md5Checksum != ""
}

func (s *GDrive) list(nextPageToken string, q string) (*drive.FileList, error){
	return s.service.Files.List().Fields("nextPageToken, files(id, name, mimeType)").Q(q).PageToken(nextPageToken).Do()
}

func (s *GDrive) findId(filename string, token string) (string, error) {
	fileId, rootId, tokenId, nextPageToken := "", "", "", ""

	q := fmt.Sprintf("name='%s' and trashed=false", s.basedir)
	l, err := s.list(nextPageToken, q)
	for 0 < len(l.Files) {
		if err != nil {
			return "", err
		}

		for _, fi := range l.Files {
			rootId = fi.Id
			break
		}

		if l.NextPageToken == "" {
			break
		}

		l, err = s.list(l.NextPageToken, q)
	}

	if token == "" {
		if rootId == "" {
			return "", fmt.Errorf("Cannot find file %s/%s", token, filename)
		}

		return rootId, nil
	}

	q = fmt.Sprintf("'%s' in parents and name='%s' and mimeType='%s' and trashed=false", rootId, token, GDriveDirectoryMimeType)
	l, err = s.list(nextPageToken, q)
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

	contentType = mime.TypeByExtension(fi.MimeType)

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
		return
	}


	contentLength = uint64(fi.Size)

	contentType = mime.TypeByExtension(fi.MimeType)

	// Get timeout reader wrapper and context
	timeoutReaderWrapper, ctx := s.getTimeoutReaderWrapperContext(time.Duration(10))

	var res *http.Response
	res, err = s.service.Files.Get(fileId).Context(ctx).Download()
	if err != nil {
		return
	}

	reader = timeoutReaderWrapper(res.Body).(io.ReadCloser)

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
	rootId, err := s.findId("", "")
	if err != nil {
		return err
	}

	dirId, err := s.findId("", token)
	if err != nil {
		return err
	}


	if dirId == "" {
		dir := &drive.File{
			Name:        token,
			Parents: 	 []string{rootId},
			MimeType:    GDriveDirectoryMimeType,
		}

		di, err := s.service.Files.Create(dir).Fields("id").Do()
		if err != nil {
			return err
		}

		dirId = di.Id
	}

	// Wrap reader in timeout reader
	timeoutReaderWrapper, ctx := s.getTimeoutReaderWrapperContext(time.Duration(10))

	// Instantiate empty drive file
	dst := &drive.File{
		Name: filename,
		Parents: []string{dirId},
		MimeType: contentType,
	}

	_, err = s.service.Files.Create(dst).Context(ctx).Media(timeoutReaderWrapper(reader)).Do()
	if err != nil {
		return err
	}

	return nil
}


// Retrieve a token, saves the token, then returns the generated client.
func getGDriveClient(config *oauth2.Config) *http.Client {
	tokenFile := "token.json"
	tok, err := gDriveTokenFromFile(tokenFile)
	if err != nil {
		tok = getGDriveTokenFromWeb(config)
		saveGDriveToken(tokenFile, tok)
	}
	return config.Client(context.Background(), tok)
}

// Request a token from the web, then returns the retrieved token.
func getGDriveTokenFromWeb(config *oauth2.Config) *oauth2.Token {
	authURL := config.AuthCodeURL("state-token", oauth2.AccessTypeOffline)
	fmt.Printf("Go to the following link in your browser then type the "+
		"authorization code: \n%v\n", authURL)

	var authCode string
	if _, err := fmt.Scan(&authCode); err != nil {
		log.Fatalf("Unable to read authorization code %v", err)
	}

	tok, err := config.Exchange(oauth2.NoContext, authCode)
	if err != nil {
		log.Fatalf("Unable to retrieve token from web %v", err)
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
func saveGDriveToken(path string, token *oauth2.Token) {
	fmt.Printf("Saving credential file to: %s\n", path)
	f, err := os.OpenFile(path, os.O_RDWR|os.O_CREATE|os.O_TRUNC, 0600)
	defer f.Close()
	if err != nil {
		log.Fatalf("Unable to cache oauth token: %v", err)
	}
	json.NewEncoder(f).Encode(token)
}
