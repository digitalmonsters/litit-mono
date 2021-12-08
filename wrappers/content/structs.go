package content

import (
	"github.com/digitalmonsters/go-common/rpc"
	"gopkg.in/guregu/null.v4"
)

type SimpleContent struct {
	Id            int64    `json:"id"`
	Duration      int      `json:"duration"`
	AgeRestricted bool     `json:"age_restricted"`
	AuthorId      int64    `json:"author_id"`
	CategoryId    null.Int `json:"category_id"`
	Hashtags      []string `json:"hashtags"`
	Width         int      `json:"width"`
	Height        int      `json:"height"`
	VideoId       string   `json:"video_id"`
}

//goland:noinspection ALL
type ContentGetInternalResponseChan struct {
	Error *rpc.RpcError           `json:"error"`
	Items map[int64]SimpleContent `json:"items"`
}

type ContentGetInternalRequest struct {
	IncludeDeleted bool    `json:"include_deleted"`
	ContentIds     []int64 `json:"content_ids"`
}
