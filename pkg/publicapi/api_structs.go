package publicapi

type GetCommentsByTypeWithResourceRequest struct {
	ContentId int64
	ParentId  int64
	After     string // cursor
	Count     int64  // Limit
	SortOrder string
}

type CursorPaging struct {
	HasNext bool   `json:"hasNext"`
	Next    string `json:"next"`
}

type GetCommentsByTypeWithResourceResponse struct {
	Comments []Comment    `json:"comments"`
	Paging   CursorPaging `json:"paging"`
}

type BlockedUserType string

const (
	BlockedUser   BlockedUserType = "BLOCKED USER"
	BlockedByUser BlockedUserType = "YOUR PROFILE IS BLOCKED BY USER"
)
