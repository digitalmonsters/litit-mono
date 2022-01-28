package uploader

import (
	"bytes"
	"crypto/md5"
	"encoding/json"
	"fmt"
	"github.com/digitalmonsters/go-common/s3"
	"github.com/digitalmonsters/music/configs"
	"github.com/pkg/errors"
	"github.com/rs/zerolog/log"
	"github.com/tcolgate/mp3"
	"github.com/valyala/fasthttp"
	"gopkg.in/guregu/null.v4"
	"io"
	"path/filepath"
	"strings"
)

func FileUpload(cfg *configs.Settings, ctx *fasthttp.RequestCtx) ([]byte, error) {
	m, err := ctx.Request.MultipartForm()
	if err != nil {
		return nil, err
	}

	log.Info().Msg("[FileUpload] from Multipart")
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

	log.Info().Msg("[FileUpload] file format")
	if len(f) < 2 {
		return nil, errors.New("invalid file format")
	}

	log.Info().Msg("[FileUpload] try open")
	openedFile, err := header.Open()
	if err != nil {
		return nil, err
	}

	defer openedFile.Close()

	body := make([]byte, size)

	log.Info().Msg("[FileUpload] try read")
	_, err = openedFile.Read(body)
	if err != nil {
		return nil, err
	}

	fileId := fmt.Sprintf("%x", md5.Sum(body))

	filename := fmt.Sprintf("%s.%s", fileId, f[len(f)-1])
	filePath := filepath.Join(cfg.S3.CdnDirectory, filename)
	fileUrl = fmt.Sprintf("%v/%v", cfg.S3.CdnUrl, filename)

	log.Info().Msg("[FileUpload] try upload")
	uploader := s3.NewUploader(&cfg.S3)
	if err := uploader.UploadObject(filePath, body, "application/octet-stream"); err != nil {
		return nil, err
	}

	resp := &uploadResponse{
		FileUrl: fileUrl,
		Size:    size,
	}

	if f[len(f)-1] == "mp3" {
		resp.Duration = null.FloatFrom(getSongDuration(body))
	}

	if respBytes, err := json.Marshal(resp); err != nil {
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
