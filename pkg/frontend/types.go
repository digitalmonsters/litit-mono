package frontend

type Playlist struct {
	Id         int64  `json:"id"`
	Name       string `json:"name"`
	Color      string `json:"color"`
	SongsCount int    `json:"songs_count"`
}

type Song struct {
	Id       string  `json:"id"`
	Title    string  `json:"title"`
	Artist   string  `json:"artist"`
	Url      string  `json:"url"`
	ImageUrl string  `json:"image_url"`
	Genre    string  `json:"genre"`
	Duration float64 `json:"duration"`
}
