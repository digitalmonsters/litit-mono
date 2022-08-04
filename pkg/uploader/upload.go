package uploader

import (
	"bytes"
	"crypto/md5"
	"fmt"
	"github.com/digitalmonsters/go-common/application"
	"github.com/digitalmonsters/go-common/s3"
	"github.com/digitalmonsters/music/configs"
	"github.com/pkg/errors"
	"github.com/tcolgate/mp3"
	"github.com/thoas/go-funk"
	"github.com/valyala/fasthttp"
	"gopkg.in/guregu/null.v4"
	"io"
	"path/filepath"
	"strings"
)

var extensionsForMusic = []string{"mp3"}
var extensionsForImage = []string{"jpg", "jpeg", "png"}

func FileUpload(cfg *configs.Settings, appConfig *application.Configurator[configs.AppConfig], uploadType UploadType, ctx *fasthttp.RequestCtx) (*uploadResponse, error) {
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

	f := strings.Split(header.Filename, ".")
	if len(f) < 2 {
		return nil, errors.New("invalid file format")
	}

	fileExtension := f[len(f)-1]

	if !checkFileExtension(uploadType, fileExtension) {
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

	var filePath string
	if uploadType >= UploadTypeCreatorsSongFull && uploadType <= UploadTypeCreatorsSongImage { //creator upload
		filePath = "creator"
	}

	var duration null.Float

	if fileExtension == "mp3" {
		duration = null.FloatFrom(getSongDuration(body))
		if uploadType == UploadTypeCreatorsSongFull {
			if int(duration.ValueOrZero()) > appConfig.Values.MUSIC_FULL_VERSION_MAX_DURATION {
				return nil, errors.New("song duration is greater than max song duration")
			}
		}

		if uploadType == UploadTypeCreatorsSongShort {
			if int(duration.ValueOrZero()) > appConfig.Values.MUSIC_SHORT_VERSION_MAX_DURATION {
				return nil, errors.New("song duration is greater than max song duration")
			}
		}
	}

	filePath = filepath.Join(filePath, uploadType.ToString(), filename)

	uploader := s3.NewUploader(&cfg.S3)
	if err := uploader.UploadObject(filePath, body, "application/octet-stream"); err != nil {
		return nil, err
	}

	return &uploadResponse{
		FileUrl:  filePath,
		Size:     size,
		Duration: duration,
	}, nil
}

func checkFileExtension(t UploadType, ext string) bool {
	switch t {
	case UploadTypeAdminMusic, UploadTypeCreatorsSongFull, UploadTypeCreatorsSongShort:
		return funk.ContainsString(extensionsForMusic, ext)
	case UploadTypeCreatorsSongImage:
		return funk.ContainsString(extensionsForImage, ext)
	default:
		return false
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
		duration += fr.Duration().Seconds()
	}

	return duration
}
