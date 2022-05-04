package s3

import (
	"bytes"
	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/session"
	"github.com/aws/aws-sdk-go/service/s3"
	"github.com/digitalmonsters/go-common/boilerplate"
	"time"
)

type IUploader interface {
	GetObjectSignedUrl(path string, urlExpiration time.Duration) (string, error)
	PutObjectSignedUrl(path string, urlExpiration time.Duration, acl string) (string, error)
	GetObjectSize(path string) (int64, error)
	UploadObject(path string, data []byte, contentType string) error
}

type Uploader struct {
	config   *boilerplate.S3Config
	session  *session.Session
	s3Client *s3.S3
}

func NewUploader(cfg *boilerplate.S3Config) IUploader {
	u := &Uploader{
		config: cfg,
	}
	return u
}

func (u *Uploader) GetObjectSignedUrl(path string, urlExpiration time.Duration) (string, error) {
	client, err := u.getClient()
	if err != nil {
		return "", err
	}

	req, _ := client.GetObjectRequest(&s3.GetObjectInput{
		Bucket: aws.String(u.config.Bucket),
		Key:    aws.String(path),
	})

	signedUrl, err := req.Presign(urlExpiration)
	if err != nil {
		return "", err
	}

	return signedUrl, nil
}

func (u *Uploader) PutObjectSignedUrl(path string, urlExpiration time.Duration, acl string) (string, error) {
	client, err := u.getClient()
	if err != nil {
		return "", err
	}

	putReq := &s3.PutObjectInput{
		Bucket: aws.String(u.config.Bucket),
		Key:    aws.String(path),
	}

	if len(acl) > 0 {
		putReq.ACL = &acl
	}

	req, _ := client.PutObjectRequest(putReq)

	signedUrl, err := req.Presign(urlExpiration)
	if err != nil {
		return "", err
	}

	return signedUrl, nil
}

func (u *Uploader) GetObjectSize(path string) (int64, error) {
	client, err := u.getClient()
	if err != nil {
		return 0, err
	}

	fileMetadata, err := client.HeadObject(&s3.HeadObjectInput{
		Bucket: aws.String(u.config.Bucket),
		Key:    aws.String(path),
	})
	if err != nil {
		return 0, err
	}

	return *fileMetadata.ContentLength, nil
}

func (u *Uploader) UploadObject(path string, data []byte, contentType string) error {
	client, err := u.getClient()
	if err != nil {
		return err
	}
	_, err = client.PutObject(&s3.PutObjectInput{
		Body:        bytes.NewReader(data),
		Key:         aws.String(path),
		Bucket:      aws.String(u.config.Bucket),
		ContentType: aws.String(contentType),
	})
	return err
}

func (u *Uploader) getClient() (*s3.S3, error) {
	if u.session == nil {
		if sess, err := session.NewSession(&aws.Config{Region: aws.String(u.config.Region)}); err != nil {
			return nil, err
		} else {
			u.session = sess
		}
	}

	if u.s3Client == nil {
		u.s3Client = s3.New(u.session)
	}
	return u.s3Client, nil
}
