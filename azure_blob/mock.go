package azure_blob

import "time"

type AzureBlobObjectMock struct {
	GetObjectSignedUrlFn func(path string, urlExpiration time.Duration) (string, error)
	PutObjectSignedUrlFn func(path string, urlExpiration time.Duration, acl string) (string, error)
	GetObjectSizeFn      func(path string) (int64, error)
	UploadObjectFn       func(path string, data []byte, contentType string) error
	ListBlobsFn          func(containerName string) ([]string, error)
	DownloadFn           func(blobName string, destination string, containerName string) error
	DeleteBlobFn         func(blobName string, containerName string) error
}

func (u *AzureBlobObjectMock) GetObjectSignedUrl(path string, urlExpiration time.Duration) (string, error) {
	return u.GetObjectSignedUrlFn(path, urlExpiration)
}
func (u *AzureBlobObjectMock) PutObjectSignedUrl(path string, urlExpiration time.Duration, acl string) (string, error) {
	return u.PutObjectSignedUrlFn(path, urlExpiration, acl)
}
func (u *AzureBlobObjectMock) GetObjectSize(path string) (int64, error) {
	return u.GetObjectSizeFn(path)
}
func (u *AzureBlobObjectMock) UploadObject(path string, data []byte, contentType string) error {
	return u.UploadObjectFn(path, data, contentType)
}
func (u *AzureBlobObjectMock) ListBlobs(containerName string) ([]string, error) {
	return u.ListBlobsFn(containerName)
}
func (u *AzureBlobObjectMock) Download(blobName string, destination string, containerName string) error {
	return u.DownloadFn(blobName, destination, containerName)
}
func (u *AzureBlobObjectMock) DeleteBlob(blobName string, containerName string) error {
	return u.DeleteBlobFn(blobName, containerName)
}

func GetMock() IAzureBlobObject { // for compiler errors
	return &AzureBlobObjectMock{}
}
