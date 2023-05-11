package storage

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"io/ioutil"
	"log"
	"net/http"
	"os"
	"path/filepath"
	"strings"
	"time"

	"golang.org/x/oauth2"
	"golang.org/x/oauth2/google"
	"google.golang.org/api/drive/v3"
	"google.golang.org/api/googleapi"
	"google.golang.org/api/option"
)

// GDrive is a storage backed by GDrive
type GDrive struct {
	service         *drive.Service
	rootID          string
	basedir         string
	localConfigPath string
	chunkSize       int
	logger          *log.Logger
}

const gDriveRootConfigFile = "root_id.conf"
const gDriveTokenJSONFile = "token.json"
const gDriveDirectoryMimeType = "application/vnd.google-apps.folder"

// NewGDriveStorage is the factory for GDrive
func NewGDriveStorage(ctx context.Context, clientJSONFilepath string, localConfigPath string, basedir string, chunkSize int, logger *log.Logger) (*GDrive, error) {

	b, err := ioutil.ReadFile(clientJSONFilepath)
	if err != nil {
		return nil, err
	}

	// If modifying these scopes, delete your previously saved client_secret.json.
	config, err := google.ConfigFromJSON(b, drive.DriveScope, drive.DriveMetadataScope)
	if err != nil {
		return nil, err
	}

	httpClient := getGDriveClient(ctx, config, localConfigPath, logger)

	srv, err := drive.NewService(ctx, option.WithHTTPClient(httpClient))
	if err != nil {
		return nil, err
	}

	storage := &GDrive{service: srv, basedir: basedir, rootID: "", localConfigPath: localConfigPath, chunkSize: chunkSize, logger: logger}
	err = storage.setupRoot()
	if err != nil {
		return nil, err
	}

	return storage, nil
}

func (s *GDrive) setupRoot() error {
	rootFileConfig := filepath.Join(s.localConfigPath, gDriveRootConfigFile)

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
		MimeType: gDriveDirectoryMimeType,
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

	q := fmt.Sprintf("'%s' in parents and name='%s' and mimeType='%s' and trashed=false", s.rootID, token, gDriveDirectoryMimeType)
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
		return "", fmt.Errorf("cannot find file %s/%s", token, filename)
	}

	q = fmt.Sprintf("'%s' in parents and name='%s' and mimeType!='%s' and trashed=false", tokenID, filename, gDriveDirectoryMimeType)
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
		return "", fmt.Errorf("cannot find file %s/%s", token, filename)
	}

	return fileID, nil
}

// Type returns the storage type
func (s *GDrive) Type() string {
	return "gdrive"
}

// Head retrieves content length of a file from storage
func (s *GDrive) Head(ctx context.Context, token string, filename string) (contentLength uint64, err error) {
	var fileID string
	fileID, err = s.findID(filename, token)
	if err != nil {
		return
	}

	var fi *drive.File
	if fi, err = s.service.Files.Get(fileID).Context(ctx).Fields("size").Do(); err != nil {
		return
	}

	contentLength = uint64(fi.Size)

	return
}

// Get retrieves a file from storage
func (s *GDrive) Get(ctx context.Context, token string, filename string, rng *Range) (reader io.ReadCloser, contentLength uint64, err error) {
	var fileID string
	fileID, err = s.findID(filename, token)
	if err != nil {
		return
	}

	var fi *drive.File
	fi, err = s.service.Files.Get(fileID).Fields("size", "md5Checksum").Do()
	if err != nil {
		return
	}
	if !s.hasChecksum(fi) {
		err = fmt.Errorf("cannot find file %s/%s", token, filename)
		return
	}

	contentLength = uint64(fi.Size)

	fileGetCall := s.service.Files.Get(fileID)
	if rng != nil {
		header := fileGetCall.Header()
		header.Set("Range", rng.Range())
	}

	var res *http.Response
	res, err = fileGetCall.Context(ctx).Download()
	if err != nil {
		return
	}

	if rng != nil {
		reader = res.Body
		rng.AcceptLength(contentLength)
		return
	}

	reader = res.Body

	return
}

// Delete removes a file from storage
func (s *GDrive) Delete(ctx context.Context, token string, filename string) (err error) {
	metadata, _ := s.findID(fmt.Sprintf("%s.metadata", filename), token)
	_ = s.service.Files.Delete(metadata).Do()

	var fileID string
	fileID, err = s.findID(filename, token)
	if err != nil {
		return
	}

	err = s.service.Files.Delete(fileID).Context(ctx).Do()
	return
}

// Purge cleans up the storage
func (s *GDrive) Purge(ctx context.Context, days time.Duration) (err error) {
	nextPageToken := ""

	expirationDate := time.Now().Add(-1 * days).Format(time.RFC3339)
	q := fmt.Sprintf("'%s' in parents and modifiedTime < '%s' and mimeType!='%s' and trashed=false", s.rootID, expirationDate, gDriveDirectoryMimeType)
	l, err := s.list(nextPageToken, q)
	if err != nil {
		return err
	}

	for 0 < len(l.Files) {
		for _, fi := range l.Files {
			err = s.service.Files.Delete(fi.Id).Context(ctx).Do()
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
func (s *GDrive) Put(ctx context.Context, token string, filename string, reader io.Reader, contentType string, contentLength uint64) error {
	dirID, err := s.findID("", token)
	if err != nil {
		return err
	}

	if dirID == "" {
		dir := &drive.File{
			Name:     token,
			Parents:  []string{s.rootID},
			MimeType: gDriveDirectoryMimeType,
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

	_, err = s.service.Files.Create(dst).Context(ctx).Media(reader, googleapi.ChunkSize(s.chunkSize)).Do()

	if err != nil {
		return err
	}

	return nil
}

func (s *GDrive) IsRangeSupported() bool { return true }

// Retrieve a token, saves the token, then returns the generated client.
func getGDriveClient(ctx context.Context, config *oauth2.Config, localConfigPath string, logger *log.Logger) *http.Client {
	tokenFile := filepath.Join(localConfigPath, gDriveTokenJSONFile)
	tok, err := gDriveTokenFromFile(tokenFile)
	if err != nil {
		tok = getGDriveTokenFromWeb(ctx, config, logger)
		saveGDriveToken(tokenFile, tok, logger)
	}

	return config.Client(ctx, tok)
}

// Request a token from the web, then returns the retrieved token.
func getGDriveTokenFromWeb(ctx context.Context, config *oauth2.Config, logger *log.Logger) *oauth2.Token {
	authURL := config.AuthCodeURL("state-token", oauth2.AccessTypeOffline)
	fmt.Printf("Go to the following link in your browser then type the "+
		"authorization code: \n%v\n", authURL)

	var authCode string
	if _, err := fmt.Scan(&authCode); err != nil {
		logger.Fatalf("Unable to read authorization code %v", err)
	}

	tok, err := config.Exchange(ctx, authCode)
	if err != nil {
		logger.Fatalf("Unable to retrieve token from web %v", err)
	}
	return tok
}

// Retrieves a token from a local file.
func gDriveTokenFromFile(file string) (*oauth2.Token, error) {
	f, err := os.Open(file)
	defer CloseCheck(f)
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
	defer CloseCheck(f)
	if err != nil {
		logger.Fatalf("Unable to cache oauth token: %v", err)
	}

	err = json.NewEncoder(f).Encode(token)
	if err != nil {
		logger.Fatalf("Unable to encode oauth token: %v", err)
	}
}
