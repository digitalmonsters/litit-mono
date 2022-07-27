package watch

import (
	"fmt"
	"github.com/digitalmonsters/go-common/eventsourcing"
	"github.com/digitalmonsters/go-common/rpc"
	"gopkg.in/guregu/null.v4"
)

type LastWatchesByUserRecord struct {
	ContentId  int64   `json:"content_id"`
	Duration   int     `json:"duration"`
	IsFullView bool    `json:"is_full_view"`
	Percent    float64 `json:"percent"`
}

type LastWatcherByUserResponseChan struct {
	Error *rpc.RpcError                       `json:"error"`
	Items map[int64][]LastWatchesByUserRecord `json:"items"`
}

type GetLatestWatchesByUserRequest struct {
	LimitPerUser int     `json:"limit_per_user"`
	UserIds      []int64 `json:"user_ids"`
	MinPercent   float64 `json:"min_percent"`
}

type CategoryInfo struct {
	Id         int64  `json:"id"`
	Name       string `json:"name"`
	ViewsCount int64  `json:"views_count"`
}

type GetCategoriesByViewsRequest struct {
	Limit  int64
	Offset int64
}

type GetCategoriesResponseChan struct {
	Error *rpc.RpcError  `json:"error"`
	Items []CategoryInfo `json:"items"`
}

type AddViewsRequest struct {
	ViewEvents []AddViewRecord `json:"view_events"` // todo
}

type AddViewRecord struct {
	UserId              int64                      `json:"user_id"`
	UserCountryCode     string                     `json:"user_country_code"`
	ContentId           int64                      `json:"content_id"`
	ContentType         eventsourcing.ContentEvent `json:"content_type"`
	Duration            int                        `json:"duration"`
	UserIp              string                     `json:"user_ip"`
	SharerId            null.Int                   `json:"sharer_id"`
	ShareCode           null.String                `json:"share_code"`
	AdsId               null.Int                   `json:"ads_id"`
	IsSharedView        bool                       `json:"is_shared_view"`
	CreatedAt           int64                      `json:"created_at"`
	IsGuest             bool                       `json:"is_guest"`
	IsBot               bool                       `json:"is_bot"`
	UseTokenomicVersion int8                       `json:"use_tokenomic_version"`
}

func (l AddViewRecord) GetPublishKey() string {
	return fmt.Sprintf("{\"content_id\":%v,\"user_id\":%v}", l.ContentId, l.UserId)
}

type AddViewsResponse struct {
	Success bool `json:"success"`
}

type GetUsersTotalTimeWatchingInternalRequest struct {
	UserIds []int64 `json:"user_ids"`
}
