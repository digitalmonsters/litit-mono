package report

import (
	"github.com/digitalmonsters/comments/pkg/database"
	"github.com/pkg/errors"
	"gorm.io/gorm"
)

func ReportComment(commentId int64, details string, db *gorm.DB, currentUserId int64, fromReqType string) (*database.Report, error) {
	var existingReport database.Report
	if err := db.Where("comment_id = ? and reporter_id = ? and report_type = ?", commentId,
		currentUserId, "comment").Find(&existingReport).Error; err != nil {
		return nil, errors.WithStack(err)
	}

	if existingReport.Id > 0 { // existing report
		return &existingReport, nil
	}

	var contentId int64
	if err := db.Model(database.Comment{}).Select("content_id").
		Where("id = ?", commentId).First(&contentId).Error; err != nil {
		return nil, err
	}

	existingReport.CommentId = commentId
	existingReport.ContentId = contentId
	existingReport.ReportType = "comment"
	existingReport.ReporterId = currentUserId
	existingReport.Detail = details
	existingReport.Type = fromReqType

	if err := db.Create(&existingReport).Error; err != nil {
		return nil, errors.WithStack(err)
	}

	return &existingReport, nil
}
