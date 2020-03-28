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
	"strconv"
	"strings"
	"time"

	"golang.org/x/oauth2"
	"golang.org/x/oauth2/google"
	"google.golang.org/api/drive/v3"
	"google.golang.org/api/googleapi"
)

const GDriveRootConfigFile = "root_id.conf"
const GDriveTokenJsonFile = "token.json"
const GDriveDirectoryMimeType = "application/vnd.google-apps.folder"

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
