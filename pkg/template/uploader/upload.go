package uploader

import (
	"context"
	"strings"
	"time"

	"github.com/digitalmonsters/go-common/azure_blob"
	"github.com/digitalmonsters/go-common/wrappers/content_uploader"
	"github.com/google/uuid"
	"github.com/pkg/errors"
	"github.com/thoas/go-funk"
	"github.com/valyala/fasthttp"
	"go.elastic.co/apm"
)

var extensionsForImage = []string{"jpg", "jpeg", "png", "bmp"}

func GetSignedUrl(path string, azureUploader azure_blob.IAzureBlobObject) (string, error) {
	if url, err := azureUploader.PutObjectSignedUrl(path, 10*time.Minute, ""); err != nil {
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

	//fileId := fmt.Sprintf("%x", md5.Sum(body))
	// filename := fmt.Sprintf("%s.%s", fileId, f[len(f)-1])

	// cfg := configs.GetConfig()
	// filePath := filepath.Join(cfg.AzureBlob.CdnDirectory, filename)
	// uploader := azure_blob.NewAzureBlobObject(&cfg.AzureBlob)

	// signedUrl, err := GetSignedUrl(filePath, uploader)
	// if err != nil {
	// 	return nil, err
	// }

	imagePath := getImagePath()
	respChan := uploaderWrapper.UploadContentInternal(imagePath, "application/octet-stream", body, apmTx, true)

	if err := <-respChan; err != nil {
		return nil, err
	}

	return &uploadResponse{
		FileUrl: getImageUrl(imagePath),
		Size:    size,
	}, nil
}

func getImagePath() string {
	return "image/notification/" + uuid.NewString() + ".jpg"
}

func getImageUrl(path string) string {
	return "https://litit-images.b-cdn.net/" + path
}
