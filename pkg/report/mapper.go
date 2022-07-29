package report

import (
	"context"
	"github.com/digitalmonsters/comments/pkg/database"
	"github.com/digitalmonsters/go-common/apm_helper"
	"github.com/digitalmonsters/go-common/frontend"
	"github.com/digitalmonsters/go-common/wrappers/content"
	"github.com/digitalmonsters/go-common/wrappers/user_go"
	"github.com/samber/lo"
	"go.elastic.co/apm"
)

func mapDbCommentsToReportedUserProfileCommentModels(dbComments []database.Comment, userWrapper user_go.IUserGoWrapper, ctx context.Context) []*ReportedUserProfileCommentModel {
	respItems := make([]*ReportedUserProfileCommentModel, len(dbComments))
	userIds := make([]int64, 0)

	for i, dbComment := range dbComments {
		respItems[i] = &ReportedUserProfileCommentModel{
			CommentId:   dbComment.Id,
			UserId:      dbComment.ProfileId.Int64,
			Comment:     dbComment.Comment,
			CommenterId: dbComment.AuthorId,
			Reports:     dbComment.NumReports,
		}
		if !lo.Contains(userIds, respItems[i].UserId) {
			userIds = append(userIds, respItems[i].UserId)
		}
		if !lo.Contains(userIds, respItems[i].CommenterId) {
			userIds = append(userIds, respItems[i].CommenterId)
		}
	}

	usersMapChan := userWrapper.GetUsers(userIds, ctx, false)
	usersMapChanResp := <-usersMapChan
	if usersMapChanResp.Error != nil {
		apm_helper.LogError(usersMapChanResp.Error.ToError(), ctx)
	} else {
		if usersMap := usersMapChanResp.Response; len(usersMap) > 0 {
			for i, item := range respItems {
				if val, ok := usersMap[item.UserId]; ok {
					respItems[i].UserAvatar = val.Avatar
					respItems[i].UserUsername = val.Username
				}
				if val, ok := usersMap[item.CommenterId]; ok {
					respItems[i].CommenterAvatar = val.Avatar
					respItems[i].CommenterUsername = val.Username
				}
			}
		}
	}

	return respItems
}

func mapDbReportsToReportForCommentModels(dbReports []database.Report, userWrapper user_go.IUserGoWrapper, ctx context.Context) []*ReportForCommentModel {
	respItems := make([]*ReportForCommentModel, len(dbReports))
	userIds := make([]int64, 0)

	for i, dbReport := range dbReports {
		respItems[i] = &ReportForCommentModel{
			Id:         dbReport.Id,
			Type:       dbReport.Type,
			ReportType: dbReport.ReportType,
			ReporterId: dbReport.ReporterId,
			Detail:     dbReport.Detail,
			CreatedAt:  dbReport.CreatedAt,
		}
		if !lo.Contains(userIds, respItems[i].ReporterId) {
			userIds = append(userIds, respItems[i].ReporterId)
		}
	}

	usersMapChan := userWrapper.GetUsers(userIds, ctx, false)
	usersMapChanResp := <-usersMapChan
	if usersMapChanResp.Error != nil {
		apm_helper.LogError(usersMapChanResp.Error.ToError(), ctx)
	} else {
		if usersMap := usersMapChanResp.Response; len(usersMap) > 0 {
			for i, item := range respItems {
				if val, ok := usersMap[item.ReporterId]; ok {
					respItems[i].ReporterUsername = val.Username
				}
			}
		}
	}

	return respItems
}

func mapDbCommentsToReportedVideoCommentModels(dbComments []database.Comment, userWrapper user_go.IUserGoWrapper, contentWrapper content.IContentWrapper, ctx context.Context, apmTx *apm.Transaction) []*ReportedVideoCommentModel {
	respItems := make([]*ReportedVideoCommentModel, len(dbComments))
	userIds := make([]int64, 0)
	contentIds := make([]int64, 0)

	for i, dbComment := range dbComments {
		respItems[i] = &ReportedVideoCommentModel{
			Content:     frontend.ContentModel{},
			CommentId:   dbComment.Id,
			Comment:     dbComment.Comment,
			CommenterId: dbComment.AuthorId,
			Reports:     dbComment.NumReports,
			ContentId:   dbComment.ContentId.Int64,
		}
		if !lo.Contains(userIds, respItems[i].CommenterId) {
			userIds = append(userIds, respItems[i].CommenterId)
		}
		if !lo.Contains(contentIds, respItems[i].ContentId) {
			contentIds = append(contentIds, respItems[i].ContentId)
		}
	}

	usersMapChan := userWrapper.GetUsers(userIds, ctx, false)
	usersMapChanResp := <-usersMapChan
	if usersMapChanResp.Error != nil {
		apm_helper.LogError(usersMapChanResp.Error.ToError(), ctx)
	} else {
		if usersMap := usersMapChanResp.Response; len(usersMap) > 0 {
			for i, item := range respItems {
				if val, ok := usersMap[item.CommenterId]; ok {
					respItems[i].CommenterAvatar = val.Avatar
					respItems[i].CommenterUsername = val.Username
				}
			}
		}
	}

	contentMapChan := contentWrapper.GetInternalAdminModels(contentIds, apmTx, false)
	contentMapChanResp := <-contentMapChan
	if contentMapChanResp.Error != nil {
		apm_helper.LogError(usersMapChanResp.Error.ToError(), ctx)
	} else {
		if contentMap := contentMapChanResp.Response; len(contentMap) > 0 {
			for i, item := range respItems {
				if val, ok := contentMap[item.ContentId]; ok {
					respItems[i].Content = val
				}
			}
		}
	}

	return respItems
}
