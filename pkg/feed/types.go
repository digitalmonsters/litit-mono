package feed

type CursorPaging struct {
	Before string `json:"before"`
	After  string `json:"after"`
}

type ContentFeedResponse struct {
	Data     []MusicFeedItem `json:"data"`
	Paging   CursorPaging    `json:"paging"`
	FeedType string          `json:"feed_type"`
}

type MusicFeedItemType string

const (
	ContentFeedItemMusic = "music"
)

type MusicFeedItem struct {
	Type MusicFeedItemType `json:"type"`
	Data interface{}       `json:"data"`
}
