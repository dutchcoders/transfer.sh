package storage

import (
	"context"
	"fmt"
	"io"
	"log"
	"time"

	"github.com/Azure/azure-sdk-for-go/sdk/azidentity"
	"github.com/Azure/azure-sdk-for-go/sdk/storage/azblob"
	"github.com/Azure/azure-sdk-for-go/sdk/storage/azblob/bloberror"
)

type AzureStorage struct {
	Storage
	client        *azblob.Client
	containerName string
	logger        *log.Logger
}

func getCredentials() (*azidentity.DefaultAzureCredential, error) {
	return azidentity.NewDefaultAzureCredential(nil)
}

func NewAzureBlobStorage(ctx context.Context, storageAccountName string, containerName string, logger *log.Logger) (Storage, error) {
	credentials, err := getCredentials()
	if err != nil {
		return nil, err
	}

	serviceUrl := fmt.Sprintf("https://%s.blob.core.windows.net", storageAccountName)
	client, err := azblob.NewClient(serviceUrl, credentials, nil)
	if err != nil {
		return nil, err
	}

	azureStorage := &AzureStorage{
		client:        client,
		containerName: containerName,
		logger:        logger,
	}

	return azureStorage, nil
}

func (s *AzureStorage) Type() string {
	return "azure"
}

func (s *AzureStorage) Get(ctx context.Context, token string, filename string, _ *Range) (io.ReadCloser, uint64, error) {
	key := fmt.Sprintf("%s/%s", token, filename)

	resp, err := s.client.DownloadStream(ctx, s.containerName, key, nil)
	if err != nil {
		return nil, 0, err
	}
	return resp.Body, uint64(*resp.ContentLength), nil
}

func (s *AzureStorage) Head(ctx context.Context, token string, filename string) (contentLength uint64, err error) {
	key := fmt.Sprintf("%s/%s", token, filename)
	containerClient := s.client.ServiceClient().NewContainerClient(s.containerName)
	props, err := containerClient.NewBlobClient(key).GetProperties(ctx, nil)

	if err != nil {
		return 0, err
	}
	return uint64(*props.ContentLength), nil
}

func (s *AzureStorage) Put(ctx context.Context, token string, filename string, reader io.Reader, _ string, _ uint64) error {
	key := fmt.Sprintf("%s/%s", token, filename)
	_, err := s.client.UploadStream(ctx, s.containerName, key, reader, nil)
	return err
}

func (s *AzureStorage) Delete(ctx context.Context, token string, filename string) error {
	key := fmt.Sprintf("%s/%s", token, filename)
	_, err := s.client.DeleteBlob(ctx, s.containerName, key, nil)
	return err
}

func (s *AzureStorage) IsNotExist(err error) bool {
	return bloberror.HasCode(err, bloberror.BlobNotFound)
}

func (s *AzureStorage) Purge(ctx context.Context, days time.Duration) error {
	pager := s.client.NewListBlobsFlatPager(s.containerName, nil)

	for pager.More() {
		page, err := pager.NextPage(ctx)
		if err != nil {
			return err
		}

		for _, blob := range page.Segment.BlobItems {
			if time.Since(*blob.Properties.LastModified) > days {
				key := *blob.Name
				if _, err := s.client.DeleteBlob(ctx, s.containerName, key, nil); err != nil {
					return err
				}
			}
		}
	}
	return nil
}

func (s *AzureStorage) IsRangeSupported() bool {
	return false
}
