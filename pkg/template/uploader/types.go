package uploader

type uploadResponse struct {
	FileUrl string `json:"file_url"`
	Size    int64  `json:"size"`
}
