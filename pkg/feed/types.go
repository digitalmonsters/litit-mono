package feed

import "github.com/digitalmonsters/music/pkg/frontend"

type CursorPaging struct {
	Before string `json:"before"`
	After  string `json:"after"`
}

type ContentFeedResponse struct {
	Data     []frontend.CreatorSongModel `json:"data"`
	Paging   CursorPaging                `json:"paging"`
	FeedType string                      `json:"feed_type"`
}
