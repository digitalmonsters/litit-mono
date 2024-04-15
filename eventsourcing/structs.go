package eventsourcing

import (
	"encoding/json"
	"fmt"
	"time"

	"github.com/lib/pq"
	"gopkg.in/guregu/null.v4"
)

type LikeEvent struct {
	UserId      int64       `json:"user_id"`
	ContentId   int64       `json:"content_id"`
	ContentType ContentType `json:"content_type"`
	Like        bool        `json:"like"`
	CreatedAt   int64       `json:"created_at"`
}

func (l LikeEvent) GetPublishKey() string {
	return fmt.Sprintf("{\"content_id\":%v,\"user_id\":%v}", l.ContentId, l.UserId)
}

type UserCategoryEvent struct {
	UserId     int64 `json:"user_id"`
	CategoryId int64 `json:"category_id"`
	Subscribed bool  `json:"subscribed"`
	CreatedAt  int64 `json:"created_at"`
}

func (l UserCategoryEvent) GetPublishKey() string {
	return fmt.Sprintf("{\"category_id\":%v,\"user_id\":%v}", l.CategoryId, l.UserId)
}

type UserHashtagEvent struct {
	UserId     int64  `json:"user_id"`
	Hashtag    string `json:"hashtag"`
	Subscribed bool   `json:"subscribed"`
}

func (l UserHashtagEvent) GetPublishKey() string {
	name := l.Hashtag
	if v, _ := json.Marshal(name); len(v) > 0 {
		name = string(v)
	}
	return fmt.Sprintf("{\"hashtag\":\"%v\",\"user_id\":%v}", name, l.UserId)
}

type ViewEvent struct {
	UserId              int64       `json:"user_id"`
	UserCountryCode     string      `json:"user_country_code"`
	UserIp              string      `json:"user_ip"`
	ContentId           int64       `json:"content_id"`
	ContentType         ContentType `json:"content_type"`
	ContentAuthorId     int64       `json:"content_author_id"`
	ContentTotalViews   int64       `json:"content_total_views"`
	WatchTime           int         `json:"watch_time"`
	WatchPercent        int8        `json:"watch_percent"`
	SharerId            null.Int    `json:"sharer_id"`
	SharerCode          null.String `json:"sharer_code"`
	IsSharedView        bool        `json:"is_shared_view"`
	IsBot               bool        `json:"is_bot"`
	CreatedAt           time.Time   `json:"created_at"`
	UseTokenomicVersion int8        `json:"use_tokenomic_version"`
	ListenType          ListenType  `json:"listen_type"`
	SourceView          SourceView  `json:"source_view"`
}

func (v ViewEvent) GetPublishKey() string {
	return fmt.Sprintf("%v_%v", v.UserId, v.ContentId)
}

type ListenType int

const (
	ListenTypeNone  = ListenType(0)
	ListenTypeShort = ListenType(1)
	ListenTypeFull  = ListenType(2)
)

type FollowEvent struct {
	ToUserId  int64 `json:"to_user_id"`
	UserId    int64 `json:"user_id"`
	Follow    bool  `json:"follow"`
	CreatedAt int64 `json:"created_at"`
}

func (l FollowEvent) GetPublishKey() string {
	return fmt.Sprintf("{\"user_id\":%v,\"to_user_id\":%v}", l.UserId, l.ToUserId)
}

type ContentType int

const (
	ContentTypeVideo    = ContentType(0)
	ContentTypeSpot     = ContentType(1)
	ContentTypeBio      = ContentType(2)
	ContentTypeMusic    = ContentType(3)
	ContentTypeDogs     = ContentType(4)
	ContentTypePets     = ContentType(5)
	ContentTypePreview  = ContentType(6)
	ContentTypeDogsSpot = ContentType(7)
	ContentTypeCatsSpot = ContentType(8)
)

func (c ContentType) IsPetProfileRequired(value ContentType) bool {
	if value == ContentTypeDogsSpot || value == ContentTypeCatsSpot {
		return true
	}
	return false
}

func (c ContentType) IsPetSpotFeed(value ContentType) bool {
	if value == ContentTypeDogsSpot || value == ContentTypeCatsSpot {
		return true
	}
	return false
}

type UploadStatus int

const (
	UploadStatusNotStarted = UploadStatus(0)
	UploadStatusInProgress = UploadStatus(1)
	UploadStatusCompleted  = UploadStatus(2)
)

type ContentEvent struct {
	Id               int64          `json:"id"`
	UserId           int64          `json:"user_id"`
	VideoId          string         `json:"video_id"`
	PageUrl          string         `json:"page_url"`
	Title            string         `json:"title"`
	Artist           string         `json:"artist"`
	Description      string         `json:"description"`
	CategoryId       null.Int       `json:"category_id"`
	SubcategoryId    null.Int       `json:"subcategory_id"`
	Duration         null.Float     `json:"duration"`
	AgeRestricted    bool           `json:"age_restricted"`
	Whitelisted      bool           `json:"whitelisted"`
	WhitelistedById  null.Int       `json:"whitelisted_by_id"`
	WhitelistedAt    null.Time      `json:"whitelisted_at"`
	Approved         bool           `json:"approved"`
	ApprovedById     null.Int       `json:"approved_by_id"`
	Reason           string         `json:"reason"`
	LiveAt           null.Time      `json:"live_at"`
	Flagged          bool           `json:"flagged"`
	Unlisted         bool           `json:"unlisted"`
	Suspended        bool           `json:"suspended"`
	SuspendedById    null.Int       `json:"suspended_by_id"`
	SuspendedAt      null.Time      `json:"suspended_at"`
	Deleted          bool           `json:"deleted"`
	DeletedById      null.Int       `json:"deleted_by_id"`
	DeletedAt        null.Time      `json:"deleted_at"`
	CreatedAt        time.Time      `json:"created_at"`
	UpdatedAt        time.Time      `json:"updated_at"`
	OhwApplicationId null.Int       `json:"ohw_application_id"`
	Hashtags         []string       `json:"hashtags"`
	HashtagsArray    pq.StringArray `json:"hashtags_array"`
	AllowComments    bool           `json:"allow_comments"`
	NotToRepeat      bool           `json:"not_to_repeat"`
	VideoShareLink   string         `json:"video_share_link"`
	Draft            bool           `json:"draft"`
	Width            int            `json:"width"`
	Height           int            `json:"height"`
	UploadStatus     UploadStatus   `json:"upload_status"`
	ByAdmin          bool           `json:"by_admin"`
	Fps              string         `json:"fps"`
	Bitrate          string         `json:"bitrate"`
	LikesCount       int64          `json:"likes_count"`
	WatchCount       int64          `json:"watch_count"`
	SharesCount      int64          `json:"shares_count"`
	CommentsCount    int64          `json:"comments_count"`
	RejectReason     null.Int       `json:"reject_reason"`
	IsNewVisible     bool           `json:"is_new_visible"`
	ContentType      ContentType    `json:"content_type"`
	BaseChangeEvent
}

func (c ContentEvent) GetPublishKey() string {
	return fmt.Sprintf("%v", c.Id)
}

type ContentPredictions struct {
	Porn       float64 `json:"porn"`
	Sexy       float64 `json:"sexy"`
	Hentai     float64 `json:"hentai"`
	Neutral    float64 `json:"neutral"`
	Drawing    float64 `json:"drawing"`
	Monochrome float64 `json:"monochrome"`
}
type ContentPredictionsEvent struct {
	BaseChangeEvent
	ContentId   int64              `json:"content_id"`
	Predictions ContentPredictions `json:"predictions"`
}

type BaseChangeEvent struct {
	CrudOperation       ChangeEvenType `json:"crud_operation"`
	CrudOperationReason string         `json:"crud_operation_reason"`
}

func NewBaseChangeEvent(crudOperation ChangeEvenType) BaseChangeEvent {
	return BaseChangeEvent{
		CrudOperation:       crudOperation,
		CrudOperationReason: "",
	}
}

func NewBaseChangeEventWithReason(crudOperation ChangeEvenType, reason string) BaseChangeEvent {
	return BaseChangeEvent{
		CrudOperation:       crudOperation,
		CrudOperationReason: reason,
	}
}

type ChangeEvenType string

const (
	ChangeEventTypeCreated = ChangeEvenType("created")
	ChangeEventTypeUpdated = ChangeEvenType("updated")
	ChangeEventTypeDeleted = ChangeEvenType("deleted")
)

type DisLikeEvent struct {
	UserId      int64       `json:"user_id"`
	ContentId   int64       `json:"content_id"`
	ContentType ContentType `json:"content_type"`
	Dislike     bool        `json:"dislike"`
	CreatedAt   int64       `json:"created_at"`
}

func (l DisLikeEvent) GetPublishKey() string {
	return fmt.Sprintf("{\"content_id\":%v,\"user_id\":%v}", l.ContentId, l.UserId)
}

type LoveEvent struct {
	UserId      int64       `json:"user_id"`
	ContentId   int64       `json:"content_id"`
	ContentType ContentType `json:"content_type"`
	Love        bool        `json:"love"`
	CreatedAt   int64       `json:"created_at"`
}

func (l LoveEvent) GetPublishKey() string {
	return fmt.Sprintf("{\"content_id\":%v,\"user_id\":%v}", l.ContentId, l.UserId)
}

type ContentUserStatsEvent struct {
	UserId int64 `json:"user_id"`
	Shares int   `json:"shares"`
	BaseChangeEvent
}

func (e ContentUserStatsEvent) GetPublishKey() string {
	return fmt.Sprintf("%v", e.UserId)
}
