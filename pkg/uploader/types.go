package uploader

import "gopkg.in/guregu/null.v4"

type uploadResponse struct {
	FileUrl  string     `json:"file_url"`
	Size     int64      `json:"size"`
	Duration null.Float `json:"duration"`
}

type UploadType int

const (
	UploadTypeNone          = UploadType(0)
	UploadTypeMusic         = UploadType(1)
	UploadTypeCreatorsMusic = UploadType(2)
)

func (t UploadType) ToString() string {
	switch t {
	case UploadTypeMusic:
		return "music"
	case UploadTypeCreatorsMusic:
		return "creators"
	default:
		return "unk"
	}
}
