package azure_blob

import (
	"context"
	"fmt"
	"os"
	"path"
	"time"

	"github.com/Azure/azure-sdk-for-go/sdk/azcore/to"
	"github.com/Azure/azure-sdk-for-go/sdk/azidentity"
	"github.com/Azure/azure-sdk-for-go/sdk/storage/azblob"
	"github.com/Azure/azure-sdk-for-go/sdk/storage/azblob/sas"
	"github.com/digitalmonsters/go-common/boilerplate"
)

type IAzureBlobObject interface {
	GetObjectSignedUrl(fileName, containerName string, urlExpiration time.Duration) (string, error)
	PutObjectSignedUrl(fileName, containerName string, urlExpiration time.Duration) (string, error)
	GetObjectSize(fileName, containerName string) (int64, error)
	UploadObject(fileName, containerName string, data []byte, contentType string) error
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

func (u *AzureBlobObject) GetObjectSignedUrl(fileName, containerName string, urlExpiration time.Duration) (string, error) {

	cred, _ := azblob.NewSharedKeyCredential(u.config.StorageAccountName, u.config.StorageAccountKey)

	sasQueryParams, err := sas.BlobSignatureValues{
		Protocol:      sas.ProtocolHTTPS,
		StartTime:     time.Now().UTC(),
		ExpiryTime:    time.Now().UTC().Add(urlExpiration),
		Permissions:   to.Ptr(sas.BlobPermissions{Read: true, Create: false, Write: false, Tag: false}).String(),
		ContainerName: containerName,
		BlobName:      fileName,
	}.SignWithSharedKey(cred)

	if err != nil {
		return "", err
	}

	signedUrl := fmt.Sprintf("https://%s.blob.core.windows.net/?%s", u.config.StorageAccountName, sasQueryParams.Encode())

	return signedUrl, nil
}

func (u *AzureBlobObject) PutObjectSignedUrl(fileName, containerName string, urlExpiration time.Duration) (string, error) {

	cred, _ := azblob.NewSharedKeyCredential(u.config.StorageAccountName, u.config.StorageAccountKey)

	sasQueryParams, err := sas.BlobSignatureValues{
		Protocol:      sas.ProtocolHTTPS,
		StartTime:     time.Now().UTC(),
		ExpiryTime:    time.Now().UTC().Add(urlExpiration),
		Permissions:   to.Ptr(sas.BlobPermissions{Read: true, Create: true, Write: true, Tag: true}).String(),
		ContainerName: containerName,
		BlobName:      fileName,
	}.SignWithSharedKey(cred)

	if err != nil {
		return "", err
	}

	signedUrl := fmt.Sprintf("https://%s.blob.core.windows.net/?%s", u.config.StorageAccountName, sasQueryParams.Encode())

	return signedUrl, nil
}

func (u *AzureBlobObject) GetObjectSize(fileName, containerName string) (int64, error) {
	client, err := u.getClient()
	if err != nil {
		return 0, err
	}

	blobClient := client.ServiceClient().NewContainerClient(containerName).NewBlockBlobClient(fileName)
	p, err := blobClient.GetProperties(context.Background(), nil)
	if err != nil {
		return 0, err
	}

	return *p.ContentLength, nil
}

func (u *AzureBlobObject) UploadObject(fileName, containerName string, data []byte, contentType string) error {
	client, err := u.getClient()
	if err != nil {
		return err
	}

	blobClient := client.ServiceClient().NewContainerClient(containerName).NewBlockBlobClient(fileName)
	_, err = blobClient.UploadBuffer(context.Background(), data, &azblob.UploadBufferOptions{
		BlockSize:   int64(4 * 1024),
		Concurrency: uint16(3),
		// If Progress is non-nil, this function is called periodically as bytes are uploaded.
		Progress: func(bytesTransferred int64) {
			fmt.Println(bytesTransferred)
		},
	})

	if err != nil {
		return err
	}

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
