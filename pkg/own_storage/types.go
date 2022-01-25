package own_storage

import "gopkg.in/guregu/null.v4"

type AddSongsToOwnStorageRequest struct {
	Items []OwnSongItem `json:"items"`
}

type OwnSongItem struct {
	Id          null.Int `json:"id"`
	Title       string   `json:"title"`
	Description string   `json:"description"`
	Artist      string   `json:"artist"`
	ImageUrl    string   `json:"image_url"`
	FileUrl     string   `json:"file_url"`
	Genre       string   `json:"genre"`
	Duration    float64  `json:"duration"`
}
