package frontend

import (
	"gopkg.in/guregu/null.v4"
	"time"
)

type VideoSubcategoryModel struct {
	Emojis    string   `json:"emojis"`
	Id        int64    `json:"id"`
	Name      string   `json:"name"`
	ParentId  null.Int `json:"parent_id"`
	SortOrder int      `json:"sort_order"`
	Status    int      `json:"status"`
}

type VideoUserModel struct {
	Avatar    string `json:"avatar"`
	FirstName string `json:"firstname"`
	Id        int64  `json:"id"`
	LastName  string `json:"lastname"`
	UserName  string `json:"username"`
	Verified  bool   `json:"verified"`
}

type ContentModel struct {
	Id int64 `json:"id"`

	AnimUrl            string `json:"anim_url"`
	CommentsCount      int64  `json:"comments_count"`
	IsCreatorFollowing bool   `json:"is_creator_following"`
	IsFollowing        bool   `json:"is_following"`
	IsVertical         bool   `json:"is_vertical"`
	LikedByMe          bool   `json:"liked_by_me"`
	Unlisted           bool   `json:"unlisted"`

	Subcategory VideoSubcategoryModel `json:"subcategory"`
	User        VideoUserModel        `json:"user"`

	UserId           int64        `json:"user_id"`
	VideoId          string       `json:"video_id"`
	Thumbnail        string       `json:"thumbnail"`
	VideoUrl         string       `json:"video_url"`
	PageUrl          string       `json:"page_url"`
	Title            null.String  `json:"title"`
	Artist           null.String  `json:"artist"`
	Description      string       `json:"description"`
	CategoryId       null.Int     `json:"category_id"`
	SubcategoryId    null.Int     `json:"subcategory_id"`
	Duration         float64      `json:"duration"`
	AgeRestricted    bool         `json:"age_restricted"`
	LiveAt           null.Time    `json:"live_at"`
	LiveAtTs         int64        `json:"live_at_ts"`
	Flagged          bool         `json:"flagged"`
	CreatedAt        time.Time    `json:"created_at"`
	CreatedAtTs      int64        `json:"created_at_ts"`
	UpdatedAt        time.Time    `json:"updated_at"`
	UpdatedAtTs      int64        `json:"updated_at_ts"`
	OhwApplicationId null.Int     `json:"ohw_application_id"`
	HashtagsArray    []string     `json:"hashtags_array"`
	AllowComments    bool         `json:"allow_comments"`
	NotToRepeat      bool         `json:"not_to_repeat"`
	VideoShareLink   string       `json:"video_share_link"`
	RejectReason     RejectReason `json:"reject_reason"`
	Draft            bool         `json:"draft"`

	Width        int   `json:"width"`
	Height       int   `json:"height"`
	UploadStatus int   `json:"upload_status"`
	LikesCount   int64 `json:"likes_count"`
	WatchCount   int64 `json:"watch_count"`
	ShareCount   int64 `json:"shares_count"`
}

type RejectReason int

const (
	RejectReasonNone                       = RejectReason(0)
	RejectReasonFakeIdentity               = RejectReason(1)
	RejectReasonOffensive                  = RejectReason(2)
	RejectReasonHateSpeech                 = RejectReason(3)
	RejectReasonHateNudityOrSexualActivity = RejectReason(4)
	RejectReasonViolence                   = RejectReason(5)
	RejectReasonHarassment                 = RejectReason(6)
)

type ContentWithPointsCount struct {
	ContentModel
	PointsCount float64 `json:"points_count"`
}
