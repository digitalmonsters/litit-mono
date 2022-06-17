package uploader

import (
	"crypto/md5"
	"fmt"
	"github.com/digitalmonsters/go-common/s3"
	"github.com/digitalmonsters/notification-handler/configs"
	"github.com/pkg/errors"
	"github.com/thoas/go-funk"
	"github.com/valyala/fasthttp"
	"path/filepath"
	"strings"
)

var extensionsForImage = []string{"jpg", "jpeg", "png", "bmp"}

func FileUpload(ctx *fasthttp.RequestCtx) (*uploadResponse, error) {
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
	if err = uploader.UploadObject(filePath, body, "application/octet-stream"); err != nil {
		return nil, err
	}

	return &uploadResponse{
		FileUrl: filepath.Join(cfg.CdnBase, cfg.S3.CdnDirectory, filename),
		Size:    size,
	}, nil
}
