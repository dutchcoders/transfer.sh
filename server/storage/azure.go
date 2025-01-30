package storage

import (
	"context"
	"fmt"
	"io"
	"log"
	"time"

	"github.com/Azure/azure-sdk-for-go/sdk/azidentity"
	"github.com/Azure/azure-sdk-for-go/sdk/storage/azblob"
	"github.com/Azure/azure-sdk-for-go/sdk/storage/azblob/container"
)

type AzureStorage struct {
	Storage
	client          *azblob.Client
	containerClient *container.Client
	containerName   string
	logger          *log.Logger
}

func getCredentials() (*azidentity.DefaultAzureCredential, error) {
	return azidentity.NewDefaultAzureCredential(nil)
}

func NewAzureBlobStorage(ctx context.Context, storageAccountName string, containerName string, logger *log.Logger) (Storage, error) {
	logger.Println("Creating Azure Blob Storage")

	if storageAccountName == "" {
		return nil, fmt.Errorf("missing storage account name")
	}

	if containerName == "" {
		return nil, fmt.Errorf("missing container name")
	}

	credentials, err := getCredentials()
	if err != nil {
		return nil, err
	}

	serviceUrl := "https://" + storageAccountName + ".blob.core.windows.net"
	client, err := azblob.NewClient(serviceUrl, credentials, nil)
	if err != nil {
		return nil, err
	}

	containerClient := client.ServiceClient().NewContainerClient(containerName)

	azureStorage := &AzureStorage{
		client:          client,
		containerClient: containerClient,
		containerName:   containerName,
		logger:          logger,
	}
	return azureStorage, nil
}

func (s *AzureStorage) Type() string {
	return "azure"
}

func (s *AzureStorage) Get(ctx context.Context, token string, filename string, _ *Range) (io.ReadCloser, uint64, error) {
	key := fmt.Sprintf("%s/%s", token, filename)
	blobClient := s.containerClient.NewBlobClient(key)
	resp, err := blobClient.DownloadStream(ctx, nil)
	if err != nil {
		return nil, 0, err
	}
	return resp.Body, uint64(*resp.ContentLength), nil
}

func (s *AzureStorage) Head(ctx context.Context, token string, filename string) (contentLength uint64, err error) {
	key := fmt.Sprintf("%s/%s", token, filename)
	props, err := s.containerClient.NewBlobClient(key).GetProperties(ctx, nil)
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
	blobClient := s.containerClient.NewBlobClient(key)
	_, err := blobClient.Delete(ctx, nil)
	if err != nil {
		s.logger.Printf("Failed to delete blob %s: %v", key, err)
		return err
	}
	s.logger.Printf("Successfully deleted blob %s", key)
	return nil
}

func (s *AzureStorage) IsNotExist(err error) bool {
	return err != nil
}

func (s *AzureStorage) Purge(ctx context.Context, days time.Duration) error {
	s.logger.Printf("Purging blobs older than %v days", days)
	pager := s.containerClient.NewListBlobsFlatPager(nil)

	for pager.More() {
		page, err := pager.NextPage(ctx)
		if err != nil {
			return fmt.Errorf("failed to list blobs: %w", err)
		}

		for _, blob := range page.Segment.BlobItems {
			if time.Since(*blob.Properties.LastModified) > days {
				blobClient := s.containerClient.NewBlobClient(*blob.Name)
				_, err := blobClient.Delete(ctx, nil)
				if err != nil {
					s.logger.Printf("Failed to delete blob %s: %v", *blob.Name, err)
					continue
				}
				s.logger.Printf("Successfully deleted blob %s", *blob.Name)
			}
		}
	}
	return nil
}

func (s *AzureStorage) IsRangeSupported() bool {
	return true
}
