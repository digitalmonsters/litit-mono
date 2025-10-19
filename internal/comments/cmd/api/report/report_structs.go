package report

type reportCommentRequest struct {
	Type    string `json:"type"`
	Details string `json:"details"`
}

type successResponse struct {
	Success bool `json:"success"`
}
