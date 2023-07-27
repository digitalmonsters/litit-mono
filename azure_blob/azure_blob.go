package azure_blob

import (
	"context"
	"fmt"
	"os"
	"path"

	"github.com/Azure/azure-sdk-for-go/sdk/azidentity"
	"github.com/Azure/azure-sdk-for-go/sdk/storage/azblob"
	"github.com/digitalmonsters/go-common/boilerplate"
)

type IAzureBlobObject interface {
	Upload(fileName, containerName string) error
	ListBlobs(containerName string) error
	Download(blobName string, destination string, containerName string) error
	DeleteBlob(blobName string, containerName string) error
}

type AzureBlobObject struct {
	config       *boilerplate.AzureBlobConfig
	azblobClient *azblob.Client
}

func NewAzureBlobObject(cfg *boilerplate.AzureBlobConfig) IAzureBlobObject {
	u := &AzureBlobObject{
		config: cfg,
	}
	return u
}

func (u *AzureBlobObject) Upload(fileName, containerName string) error {
	client, err := u.getClient()
	if err != nil {
		return err
	}

	file, err := os.OpenFile(fileName, os.O_RDONLY, 0)
	if err != nil {
		return err
	}
	defer file.Close()

	_, err = client.UploadFile(context.Background(), containerName, fileName, file, nil)
	return err
}

func (u *AzureBlobObject) ListBlobs(containerName string) error {

	client, err := u.getClient()
	if err != nil {
		return err
	}
	pager := client.NewListBlobsFlatPager(containerName, nil)
	for pager.More() {
		page, err := pager.NextPage(context.Background())
		if err != nil {
			return err
		}
		for _, blob := range page.Segment.BlobItems {
			fmt.Println(*blob.Name)
		}
	}
	return nil
}

func (u *AzureBlobObject) Download(blobName string, destination string, containerName string) error {
	client, err := u.getClient()
	if err != nil {
		return err
	}
	target := path.Join(destination, blobName)
	d, err := os.Create(target)
	if err != nil {
		return err
	}
	defer d.Close()

	_, err = client.DownloadFile(context.Background(), containerName, blobName, d, nil)
	return err
}

func (u *AzureBlobObject) DeleteBlob(blobName string, containerName string) error {
	client, err := u.getClient()
	if err != nil {
		return err
	}

	_, err = client.DeleteBlob(context.Background(), containerName, blobName, nil)
	return err
}

func (u *AzureBlobObject) getClient() (*azblob.Client, error) {

	if u.azblobClient == nil {
		url := fmt.Sprintf("https://%s.blob.core.windows.net/", u.config.StorageAccountName)
		if u.config.UseCliAuth {
			cred, err := azidentity.NewAzureCLICredential(nil)
			if err != nil {
				return nil, err
			}
			if client, err := azblob.NewClient(url, cred, nil); err != nil {
				return nil, err
			} else {
				u.azblobClient = client
				return u.azblobClient, nil
			}
		}

		cred, err := azblob.NewSharedKeyCredential(u.config.StorageAccountName, u.config.StorageAccountKey)
		if err != nil {
			return nil, err
		}
		if client, err := azblob.NewClientWithSharedKeyCredential(url, cred, nil); err != nil {
			return nil, err
		} else {
			u.azblobClient = client
			return u.azblobClient, nil
		}
	}

	return u.azblobClient, nil
}
