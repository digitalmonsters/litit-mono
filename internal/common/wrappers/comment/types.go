package comment

import (
	"github.com/digitalmonsters/go-common/rpc"
	"gopkg.in/guregu/null.v4"
)

type GetCommentsInfoByIdRequest struct {
	CommentIds []int64 `json:"comment_ids"`
}

type GetCommentsInfoByIdResponseChan struct {
	Error *rpc.RpcError
	Items map[int64]CommentsInfoById `json:"items"`
}

type CommentsInfoById struct {
	ParentAuthorId null.Int `json:"parent_author_id"`
}
