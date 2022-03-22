package creators

import (
	"github.com/digitalmonsters/music/pkg/database"
	"gopkg.in/guregu/null.v4"
	"gorm.io/gorm"
	"time"
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
	Id           int64                  `json:"id"`
	Status       database.CreatorStatus `json:"status"`
	RejectReason null.String            `json:"reject_reason"`
	LibraryUrl   string                 `json:"library_url"`
	SlaExpired   bool                   `json:"sla_expired"`
	UserId       int64                  `json:"user_id"`
	FirstName    string                 `json:"first_name"`
	LastName     string                 `json:"last_name"`
	UserName     string                 `json:"user_name"`
	Avatar       null.String            `json:"avatar"`
	CreatedAt    time.Time              `json:"created_at"`
	ApprovedAt   null.Time              `json:"approved_at"`
	DeletedAt    gorm.DeletedAt         `json:"deleted_at"`
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
	Id                null.Int    `json:"id"`
	Name              string      `json:"name"`
	LyricAuthor       null.String `json:"lyric_author"`
	MusicAuthor       string      `json:"music_author"`
	CategoryId        int64       `json:"category_id"`
	FullSongUrl       string      `json:"full_song_url"`
	FullSongDuration  float64     `json:"full_song_duration"`
	ShortSongUrl      string      `json:"short_song_url"`
	ShortSongDuration float64     `json:"short_song_duration"`
	ImageUrl          string      `json:"image_url"`
	Hashtags          []string    `json:"hashtags"`
}
