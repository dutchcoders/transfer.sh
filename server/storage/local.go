package storage

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"log"
	"os"
	"path/filepath"
)

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
