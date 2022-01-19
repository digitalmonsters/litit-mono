package uploader

import "gopkg.in/guregu/null.v4"

type uploadResponse struct {
	FileUrl  string     `json:"file_url"`
	Size     int64      `json:"size"`
	Duration null.Float `json:"duration"`
}
