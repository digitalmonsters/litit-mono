package uploader

import "gopkg.in/guregu/null.v4"

type uploadResponse struct {
	FileUrl  string     `json:"file_url"`
	Size     int64      `json:"size"`
	Duration null.Float `json:"duration"`
}

type UploadType int

const (
	UploadTypeNone              = UploadType(0)
	UploadTypeAdminMusic        = UploadType(1)
	UploadTypeCreatorsSongFull  = UploadType(2)
	UploadTypeCreatorsSongShort = UploadType(3)
	UploadTypeCreatorsSongImage = UploadType(4)
)

func (t UploadType) ToString() string {
	switch t {
	case UploadTypeAdminMusic:
		return "music"
	case UploadTypeCreatorsSongFull:
		return "full"
	case UploadTypeCreatorsSongShort:
		return "short"
	case UploadTypeCreatorsSongImage:
		return "image"
	default:
		return "unk"
	}
}
