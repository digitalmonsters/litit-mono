package vote

import "gopkg.in/guregu/null.v4"

type voteRequest struct {
	VoteUp null.Bool `json:"vote_up"`
}

type successResponse struct {
	Success bool `json:"success"`
}
