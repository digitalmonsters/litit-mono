package report

import (
	"context"

	"github.com/digitalmonsters/comments/pkg/database"
	"github.com/digitalmonsters/go-common/wrappers/content"
	"github.com/digitalmonsters/go-common/wrappers/user_go"
	"github.com/pkg/errors"
	"go.elastic.co/apm"
	"gorm.io/gorm"
)

func ReportComment(commentId int64, details string, tx *gorm.DB, currentUserId int64, fromReqType string) (*database.Report, error) {
	var existingReport database.Report
	if err := tx.Where("comment_id = ? and reporter_id = ? and report_type = ?", commentId,
		currentUserId, "comment").Find(&existingReport).Error; err != nil {
		return nil, errors.WithStack(err)
	}

	if existingReport.Id > 0 { // existing report
		return &existingReport, nil
	}

	var comment database.Comment

	if err := tx.Model(database.Comment{}).
		Where("id = ?", commentId).First(&comment).Error; err != nil {
		return nil, err
	}

	existingReport.CommentId = commentId
	existingReport.ReportType = "comment"
	existingReport.ContentId = comment.ContentId
	existingReport.UserId = comment.ProfileId
	existingReport.ReporterId = currentUserId
	existingReport.Detail = details
	existingReport.Type = fromReqType

	if err := tx.Create(&existingReport).Error; err != nil {
		return nil, errors.WithStack(err)
	}

	comment.NumReports += 1

	if err := tx.Model(&comment).Update("num_reports", comment.NumReports).Error; err != nil {
		return nil, err
	}

	return &existingReport, nil
}

func GetReportedUserProfileComments(req GetReportedUserProfileCommentsRequest, db *gorm.DB, userWrapper user_go.IUserGoWrapper, ctx context.Context) (*GetReportedUserProfileCommentsResponse, error) {
	query := db.Model(database.Comment{}).Where("num_reports > 0").Where("profile_id is not null")

	if req.Approved.Valid {
		if req.Approved.ValueOrZero() {
			query = query.Where("Active = true")
		}
	}

	if req.Rejected.Valid {
		if req.Rejected.ValueOrZero() {
			query = query.Where("Active = false")
		}
	}

	if len(req.CommenterIds) > 0 {
		query = query.Where("author_id in ?", req.CommenterIds)
	}

	if len(req.UserIds) > 0 {
		query = query.Where("profile_id in ?", req.UserIds)
	}

	var totalCount int64
	if err := query.Count(&totalCount).Error; err != nil {
		return nil, errors.WithStack(err)
	}
	var comments []database.Comment
	if err := query.Order("num_reports desc").
		Limit(req.Limit).Offset(req.Offset).Find(&comments).Error; err != nil {
		return nil, errors.WithStack(err)
	}
	return &GetReportedUserProfileCommentsResponse{
		TotalCount: totalCount,
		Items:      mapDbCommentsToReportedUserProfileCommentModels(comments, userWrapper, ctx),
	}, nil
}

func GetReportsForComment(req GetReportsForCommentRequest, db *gorm.DB, userWrapper user_go.IUserGoWrapper, ctx context.Context) (*GetReportsForCommentResponse, error) {
	query := db.Model(database.Report{}).Where("comment_id = ?", req.CommentId)

	if len(req.ReportedByIds) > 0 {
		query = query.Where("reporter_id in ?", req.ReportedByIds)
	}

	var totalCount int64
	if err := query.Count(&totalCount).Error; err != nil {
		return nil, errors.WithStack(err)
	}

	var reports []database.Report

	if sortingArr := req.Sorting; len(sortingArr) > 0 {
		for _, sorting := range sortingArr {
			sortOrder := " asc"
			if !sorting.IsAscending {
				sortOrder = " desc"
			}
			query = query.Order(string(sorting.Field) + sortOrder)
		}
	} else {
		query = query.Order("created_at desc")
	}
	if err := query.Limit(req.Limit).Offset(req.Offset).
		Find(&reports).Error; err != nil {
		return nil, errors.WithStack(err)
	}
	return &GetReportsForCommentResponse{
		TotalCount: totalCount,
		Items:      mapDbReportsToReportForCommentModels(reports, userWrapper, ctx),
	}, nil
}

func GetReportedVideoComments(req GetReportedVideoCommentsRequest, db *gorm.DB, userWrapper user_go.IUserGoWrapper, contentWrapper content.IContentWrapper, ctx context.Context, apmTx *apm.Transaction) (*GetReportedVideoCommentsResponse, error) {
	query := db.Model(database.Comment{}).Where("num_reports > 0").Where("content_id is not null")

	if req.Approved.Valid {
		if req.Approved.ValueOrZero() {
			query = query.Where("Active = true")
		}
	}

	if req.Rejected.Valid {
		if req.Rejected.ValueOrZero() {
			query = query.Where("Active = false")
		}
	}

	if len(req.CommenterIds) > 0 {
		query = query.Where("author_id in ?", req.CommenterIds)
	}

	if len(req.ContentIds) > 0 {
		query = query.Where("content_id in ?", req.ContentIds)
	}

	var totalCount int64
	if err := query.Count(&totalCount).Error; err != nil {
		return nil, errors.WithStack(err)
	}
	var comments []database.Comment
	if err := query.Order("num_reports desc").
		Limit(req.Limit).Offset(req.Offset).Find(&comments).Error; err != nil {
		return nil, errors.WithStack(err)
	}
	return &GetReportedVideoCommentsResponse{
		TotalCount: totalCount,
		Items:      mapDbCommentsToReportedVideoCommentModels(comments, userWrapper, contentWrapper, ctx, apmTx),
	}, nil
}

func ApproveRejectReportedComment(req ApproveRejectReportedCommentRequest, db *gorm.DB, ctx context.Context) (*ApproveRejectReportedCommentResponse, error) {

	var comment database.Comment

	if err := db.Model(&comment).Where("Id = ?", req.CommentID).Updates(map[string]interface{}{"Active": req.Approve}).Error; err != nil {
		return nil, err
	}

	return &ApproveRejectReportedCommentResponse{
		Success: true,
	}, nil

}
