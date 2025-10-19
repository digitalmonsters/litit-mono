package s3

import (
	"time"
)

type UploaderMock struct {
	GetObjectSignedUrlFn func(path string, urlExpiration time.Duration) (string, error)
	PutObjectSignedUrlFn func(path string, urlExpiration time.Duration, acl string) (string, error)
	GetObjectSizeFn      func(path string) (int64, error)
	UploadObjectFn       func(path string, data []byte, contentType string) error
}

func (u *UploaderMock) GetObjectSignedUrl(path string, urlExpiration time.Duration) (string, error) {
	return u.GetObjectSignedUrlFn(path, urlExpiration)
}
func (u *UploaderMock) PutObjectSignedUrl(path string, urlExpiration time.Duration, acl string) (string, error) {
	return u.PutObjectSignedUrlFn(path, urlExpiration, acl)
}
func (u *UploaderMock) GetObjectSize(path string) (int64, error) {
	return u.GetObjectSizeFn(path)
}
func (u *UploaderMock) UploadObject(path string, data []byte, contentType string) error {
	return u.UploadObjectFn(path, data, contentType)
}

func GetMock() IUploader { // for compiler errors
	return &UploaderMock{}
}
