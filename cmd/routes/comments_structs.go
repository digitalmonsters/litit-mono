package routes

import "gopkg.in/guregu/null.v4"

type updateCommentRequest struct {
	Comment string `json:"comment"`
}

type reportCommentRequest struct {
	Type    string `json:"type"`
	Details string `json:"details"`
}

type createCommentRequest struct {
	ParentId null.Int `json:"parent_id"`
	Comment  string   `json:"comment"`
}
