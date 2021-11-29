package hashtag

import "github.com/digitalmonsters/go-common/rpc"

type SimpleHashtag struct {
	Name       string `json:"name"`
	ViewsCount int    `json:"views_count"`
}

type ResponseData struct {
	Items      []SimpleHashtag `json:"items"`
	TotalCount int64           `json:"total_count"`
}

type HashtagsGetInternalResponseChan struct {
	Error *rpc.RpcError `json:"error"`
	Data  *ResponseData `json:"data"`
}

type GetHashtagsInternalRequest struct {
	Hashtags []string `json:"hashtags"`
	Limit    int      `json:"limit"`
	Offset   int      `json:"offset"`
}
