package uploader

import (
	"context"
	"crypto/md5"
	"fmt"
	"github.com/digitalmonsters/go-common/s3"
	"github.com/digitalmonsters/go-common/wrappers/content_uploader"
	"github.com/digitalmonsters/notification-handler/configs"
	"github.com/digitalmonsters/notification-handler/pkg/utils"
	"github.com/pkg/errors"
	"github.com/thoas/go-funk"
	"github.com/valyala/fasthttp"
	"go.elastic.co/apm"
	"path/filepath"
	"strings"
	"time"
)

var extensionsForImage = []string{"jpg", "jpeg", "png", "bmp"}

func GetSignedUrl(path string, s3Uploader s3.IUploader) (string, error) {
	if url, err := s3Uploader.PutObjectSignedUrl(path, 10*time.Minute, ""); err != nil {
		return "", err
	} else {
		return url, nil
	}
}

func FileUpload(ctx *fasthttp.RequestCtx, uploaderWrapper content_uploader.IContentUploaderWrapper, apmTx *apm.Transaction, appCtx context.Context) (*uploadResponse, error) {
	m, err := ctx.Request.MultipartForm()
	if err != nil {
		return nil, err
	}

	files := m.File["File"]
	if len(files) == 0 {
		return nil, errors.New("no file found")
	}
	if len(files) > 1 {
		return nil, errors.New("multiple file upload not supported")
	}

	header := files[0]
	size := header.Size

	if size > 1048576 { // 1mb
		return nil, errors.New("max image size 1 MB allowed")
	}

	f := strings.Split(header.Filename, ".")

	if len(f) < 2 {
		return nil, errors.New("invalid file format")
	}

	fileExtension := f[len(f)-1]

	if !funk.ContainsString(extensionsForImage, fileExtension) {
		return nil, errors.New("wrong file extension")
	}

	openedFile, err := header.Open()
	if err != nil {
		return nil, err
	}

	defer openedFile.Close()

	body := make([]byte, size)

	_, err = openedFile.Read(body)
	if err != nil {
		return nil, err
	}

	fileId := fmt.Sprintf("%x", md5.Sum(body))
	filename := fmt.Sprintf("%s.%s", fileId, f[len(f)-1])

	cfg := configs.GetConfig()
	filePath := filepath.Join(cfg.S3.CdnDirectory, filename)
	uploader := s3.NewUploader(&cfg.S3)

	signedUrl, err := GetSignedUrl(filePath, uploader)
	if err != nil {
		return nil, err
	}

	respChan := uploaderWrapper.UploadContentInternal(utils.RetrieveContentUploaderPath(appCtx, signedUrl), "application/octet-stream", body, apmTx, true)

	if err := <-respChan; err != nil {
		return nil, err
	}

	return &uploadResponse{
		FileUrl: fmt.Sprintf("%v/%v", cfg.CdnBase, filePath),
		Size:    size,
	}, nil
}
