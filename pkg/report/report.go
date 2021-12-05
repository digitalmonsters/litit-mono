package report

import (
	"github.com/digitalmonsters/comments/pkg/database"
	"github.com/pkg/errors"
	"gorm.io/gorm"
)

func ReportComment(commentId int64, details string, db *gorm.DB, currentUserId int64) (*database.Report, error) {
	var existingReport database.Report

	if err := db.Where("comment_id = ? and reporter_id = ? and report_type = ?", commentId,
		currentUserId, "comment").Find(&existingReport).Error; err != nil {
		return nil, errors.WithStack(err)
	}

	if existingReport.Id > 0 { // existing report
		return &existingReport, nil
	}

	existingReport.CommentId = commentId
	existingReport.Type = "comment"
	existingReport.ReporterId = currentUserId
	existingReport.Detail = details

	if err := db.Create(&existingReport).Error; err != nil {
		return nil, errors.WithStack(err)
	}

	return &existingReport, nil
}
