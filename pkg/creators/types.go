package creators

import (
	"github.com/digitalmonsters/music/pkg/database"
	"gopkg.in/guregu/null.v4"
)

type BecomeMusicCreatorRequest struct {
	LibraryLink string `json:"library_link"`
}

type CreatorRequestsListRequest struct {
	UserId               null.Int                 `json:"user_id"`
	Statuses             []database.CreatorStatus `json:"statuses"`
	MaxThresholdExceeded bool                     `json:"max_threshold_exceeded"`
	OrderOption          OrderOption              `json:"order_option"`
	Limit                int                      `json:"limit"`
	Offset               int                      `json:"offset"`
}

type OrderOption int8

const (
	OrderOptionNone          = OrderOption(0)
	OrderOptionCreatedAtDesc = OrderOption(1)
	OrderOptionCreatedAtAsc  = OrderOption(2)
)

type CreatorRequestsListResponse struct {
	Items      []creatorListItem `json:"items"`
	TotalCount int64             `json:"total_count"`
}

type creatorListItem struct {
	database.Creator
	UserId    int64       `json:"user_id"`
	FirstName string      `json:"first_name"`
	LastName  string      `json:"last_name"`
	UserName  string      `json:"user_name"`
	Avatar    null.String `json:"avatar"`
}

type CreatorRequestApproveRequest struct {
	Ids []int64 `json:"ids"`
}

type CreatorRequestRejectRequest struct {
	Items []creatorRejectItem `json:"items"`
}

type creatorRejectItem struct {
	Id     int64 `json:"id"`
	Reason int64 `json:"reason"`
}

type UploadNewSongRequest struct {
	Id           null.Int    `json:"id"`
	Name         string      `json:"name"`
	LyricAuthor  null.String `json:"lyric_author"`
	MusicAuthor  string      `json:"music_author"`
	CategoryId   int64       `json:"category_id"`
	FullSongUrl  string      `json:"full_song_url"`
	ShortSongUrl string      `json:"short_song_url"`
	ImageUrl     string      `json:"image_url"`
	Hashtags     []string    `json:"hashtags"`
}
