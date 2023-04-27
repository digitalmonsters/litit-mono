package report

import (
	"time"

	"github.com/digitalmonsters/go-common/frontend"
	"gopkg.in/guregu/null.v4"
)

type GetReportedUserProfileCommentsRequest struct {
	Limit        int       `json:"limit"`
	Offset       int       `json:"offset"`
	CommenterIds []int64   `json:"commenter_ids"`
	UserIds      []int64   `json:"user_ids"`
	Approved     null.Bool `json:"approved"`
	Rejected     null.Bool `json:"rejected"`
}

type GetReportedUserProfileCommentsResponse struct {
	TotalCount int64                              `json:"total_count"`
	Items      []*ReportedUserProfileCommentModel `json:"items"`
}

type ReportedUserProfileCommentModel struct {
	CommentId         int64       `json:"comment_id"`
	UserAvatar        null.String `json:"user_avatar"`
	UserUsername      string      `json:"user_username"`
	UserId            int64       `json:"user_id"`
	Comment           string      `json:"comment"`
	CommenterAvatar   null.String `json:"commenter_avatar"`
	CommenterUsername string      `json:"commenter_username"`
	CommenterId       int64       `json:"commenter_id"`
	Reports           int64       `json:"reports"`
	Active            bool        `json:"active"`
}

type GetReportsForCommentRequest struct {
	Limit         int              `json:"limit"`
	Offset        int              `json:"offset"`
	CommentId     int64            `json:"comment_id"`
	Sorting       []ReportsSorting `json:"sorting"`
	ReportedByIds []int64          `json:"reported_by_ids"`
}

type GetReportsForCommentResponse struct {
	TotalCount int64                    `json:"total_count"`
	Items      []*ReportForCommentModel `json:"items"`
}

type GetReportedVideoCommentsRequest struct {
	Limit        int       `json:"limit"`
	Offset       int       `json:"offset"`
	CommenterIds []int64   `json:"commenter_ids"`
	ContentIds   []int64   `json:"content_ids"`
	Approved     null.Bool `json:"approved"`
	Rejected     null.Bool `json:"rejected"`
}

type GetReportedVideoCommentsResponse struct {
	TotalCount int64                        `json:"total_count"`
	Items      []*ReportedVideoCommentModel `json:"items"`
}

type ReportedVideoCommentModel struct {
	Content           frontend.ContentModel `json:"content"`
	ContentId         int64                 `json:"content_id"`
	CommentId         int64                 `json:"comment_id"`
	Comment           string                `json:"comment"`
	CommenterAvatar   null.String           `json:"commenter_avatar"`
	CommenterUsername string                `json:"commenter_username"`
	CommenterId       int64                 `json:"commenter_id"`
	Reports           int64                 `json:"reports"`
	Active            bool                  `json:"active"`
}

type ReportForCommentModel struct {
	Id               int       `json:"id"`
	Type             string    `json:"type"`
	ReportType       string    `json:"report_type"`
	ReporterId       int64     `json:"reporter_id"`
	Detail           string    `json:"detail"`
	CreatedAt        time.Time `json:"created_at"`
	ReporterUsername string    `json:"reporter_username"`
}

type ReportsSortField string

type ReportsSorting struct {
	Field       ReportsSortField `json:"field"`
	IsAscending bool             `json:"is_ascending"`
}

const (
	ReportsSortFieldDate ReportsSortField = "created_at"
)

type ApproveRejectReportedCommentRequest struct {
	CommentID int  `json:"comment_id"`
	Approve   bool `json:"approve"`
}

type ApproveRejectReportedCommentResponse struct {
	Success bool `json:"success"`
}
