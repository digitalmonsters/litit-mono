package uploader

import (
	"bytes"
	"crypto/md5"
	"encoding/json"
	"fmt"
	"github.com/digitalmonsters/go-common/s3"
	"github.com/digitalmonsters/music/configs"
	"github.com/pkg/errors"
	"github.com/tcolgate/mp3"
	"github.com/valyala/fasthttp"
	"gopkg.in/guregu/null.v4"
	"io"
	"path/filepath"
	"strings"
)

func FileUpload(cfg *configs.Settings, uploadType UploadType, ctx *fasthttp.RequestCtx) ([]byte, error) {
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

	var fileUrl string

	header := files[0]
	size := header.Size

	f := strings.Split(header.Filename, ".")

	if len(f) < 2 {
		return nil, errors.New("invalid file format")
	}

	if f[len(f)-1] == "mp3" {
		return nil, errors.New("mp3 format is only available for upload")
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
	filePath := filepath.Join(cfg.S3.CdnDirectory, filename)
	fileUrl = fmt.Sprintf("%v/%v", cfg.S3.CdnUrl, filename)

	uploader := s3.NewUploader(&cfg.S3)
	if err := uploader.UploadObject(filePath, body, "application/octet-stream"); err != nil {
		return nil, err
	}

	if respBytes, err := json.Marshal(&uploadResponse{
		FileUrl:  fileUrl,
		Size:     size,
		Duration: null.FloatFrom(getSongDuration(body)),
	}); err != nil {
		return nil, err
	} else {
		return respBytes, nil
	}
}

func getSongDuration(body []byte) float64 {
	duration := 0.0
	r := bytes.NewReader(body)
	d := mp3.NewDecoder(r)
	var fr mp3.Frame
	skipped := 0

	for {
		if err := d.Decode(&fr, &skipped); err != nil {
			if err == io.EOF {
				break
			}
			break
		}
		duration = duration + fr.Duration().Seconds()
	}

	return duration
}
