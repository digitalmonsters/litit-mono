package frontend

type Playlist struct {
	Id         int64  `json:"id"`
	Name       string `json:"name"`
	Color      string `json:"color"`
	SongsCount int    `json:"songs_count"`
}

type Song struct {
	Id         int64   `json:"id"`
	Title      string  `json:"title"`
	Artist     string  `json:"artist"`
	Url        string  `json:"url"`
	IsFavorite bool    `json:"is_favorite"`
	ImageUrl   string  `json:"image_url"`
	Genre      string  `json:"genre"`
	Duration   float64 `json:"duration"`
}

type Category struct {
	Id   int64  `json:"id"`
	Name string `json:"name"`
}

type Mood struct {
	Id   int64  `json:"id"`
	Name string `json:"name"`
}
