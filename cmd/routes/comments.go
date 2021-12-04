package routes

import (
	"encoding/json"
	"github.com/digitalmonsters/comments/pkg/publicapi"
	"github.com/digitalmonsters/comments/utils"
	"github.com/digitalmonsters/go-common/common"
	"github.com/digitalmonsters/go-common/error_codes"
	"github.com/digitalmonsters/go-common/router"
	"github.com/digitalmonsters/go-common/wrappers/content"
	"github.com/digitalmonsters/go-common/wrappers/user"
	"github.com/pkg/errors"
	"gorm.io/gorm"
	"net/http"
)

func InitPublicRoutes(httpRouter *router.HttpRouter, db *gorm.DB, userWrapper user.IUserWrapper,
	contentWrapper content.IContentWrapper) error {

	if err := httpRouter.RegisterRestCmd(router.NewRestCommand(func(request []byte,
		executionData router.MethodExecutionData) (interface{}, *error_codes.ErrorWithCode) {
		commentId := utils.ExtractInt64(executionData.GetUserValue, "comment_id", 0, 0)

		if commentId <= 0 {
			return nil, error_codes.NewErrorWithCodeRef(errors.New("invalid comment_id"), error_codes.GenericValidationError)
		}

		if comment, err := publicapi.GetCommendById(db.WithContext(executionData.Context), commentId, executionData.UserId,
			userWrapper, executionData.ApmTransaction); err != nil {
			return nil, error_codes.NewErrorWithCodeRef(err, error_codes.GenericValidationError)
		} else {
			return comment, nil
		}
	}, "/{comment_id}", http.MethodGet, common.AccessLevelPublic, false, false)); err != nil {
		return err
	}

	if err := httpRouter.RegisterRestCmd(router.NewRestCommand(func(request []byte,
		executionData router.MethodExecutionData) (interface{}, *error_codes.ErrorWithCode) {
		commentId := utils.ExtractInt64(executionData.GetUserValue, "comment_id", 0, 0)

		if commentId <= 0 {
			return nil, error_codes.NewErrorWithCodeRef(errors.New("invalid comment_id"), error_codes.GenericValidationError)
		}

		if resp, err := publicapi.DeleteCommentById(db.WithContext(executionData.Context), commentId, executionData.UserId,
			contentWrapper, executionData.ApmTransaction); err != nil {
			return nil, error_codes.NewErrorWithCodeRef(err, error_codes.GenericValidationError)
		} else {
			return resp, nil
		}
	}, "/{comment_id}", http.MethodDelete, common.AccessLevelPublic, true, true)); err != nil {
		return err
	}

	if err := httpRouter.RegisterRestCmd(router.NewRestCommand(func(request []byte,
		executionData router.MethodExecutionData) (interface{}, *error_codes.ErrorWithCode) {
		commentId := utils.ExtractInt64(executionData.GetUserValue, "comment_id", 0, 0)

		var updateRequest updateCommentRequest

		if err := json.Unmarshal(request, &updateRequest); err != nil {
			return nil, error_codes.NewErrorWithCodeRef(err, error_codes.GenericMappingError)
		}

		if len(updateRequest.Comment) == 0 {
			return nil, error_codes.NewErrorWithCodeRef(errors.New("invalid comment length"), error_codes.GenericValidationError)
		}

		if commentId <= 0 {
			return nil, error_codes.NewErrorWithCodeRef(errors.New("invalid comment_id"), error_codes.GenericValidationError)
		}

		if resp, err := publicapi.UpdateCommentById(db.WithContext(executionData.Context), commentId,
			updateRequest.Comment, executionData.UserId); err != nil {
			return nil, error_codes.NewErrorWithCodeRef(err, error_codes.GenericValidationError)
		} else {
			return resp, nil
		}
	}, "/{comment_id}", http.MethodPatch, common.AccessLevelPublic, true, true)); err != nil {
		return err
	}

	if err := httpRouter.RegisterRestCmd(router.NewRestCommand(func(request []byte,
		executionData router.MethodExecutionData) (interface{}, *error_codes.ErrorWithCode) {
		commentId := utils.ExtractInt64(executionData.GetUserValue, "comment_id", 0, 0)

		if commentId <= 0 {
			return nil, error_codes.NewErrorWithCodeRef(errors.New("invalid comment_id"), error_codes.GenericValidationError)
		}

		count := utils.ExtractInt64(executionData.GetUserValue, "count", 10, 10)
		after := utils.ExtractString(executionData.GetUserValue, "after", "")

		if resp, err := publicapi.GetRepliesByCommentId(commentId, db.WithContext(executionData.Context), executionData.ApmTransaction,
			count, after); err != nil {
			return nil, error_codes.NewErrorWithCodeRef(err, error_codes.GenericValidationError)
		} else {
			return resp, nil
		}
	}, "/{comment_id}/replies", http.MethodGet, common.AccessLevelPublic, false, false)); err != nil {
		return err
	}

	if err := httpRouter.RegisterRestCmd(router.NewRestCommand(func(request []byte,
		executionData router.MethodExecutionData) (interface{}, *error_codes.ErrorWithCode) {
		commentId := utils.ExtractInt64(executionData.GetUserValue, "comment_id", 0, 0)

		if commentId <= 0 {
			return nil, error_codes.NewErrorWithCodeRef(errors.New("invalid comment_id"), error_codes.GenericValidationError)
		}

		publicapi.VoteComment()
	}, "/{comment_id}/vote", http.MethodPost, common.AccessLevelPublic, true, false)); err != nil {
		return err
	}

	if err := httpRouter.RegisterRestCmd(router.NewRestCommand(func(request []byte,
		executionData router.MethodExecutionData) (interface{}, *error_codes.ErrorWithCode) {
		commentId := utils.ExtractInt64(executionData.GetUserValue, "comment_id", 0, 0)

		if commentId <= 0 {
			return nil, error_codes.NewErrorWithCodeRef(errors.New("invalid comment_id"), error_codes.GenericValidationError)
		}

		var reportRequest reportCommentRequest

		if err := json.Unmarshal(request, &reportRequest); err != nil {
			return nil, error_codes.NewErrorWithCodeRef(err, error_codes.GenericMappingError)
		}

		if resp, err := publicapi.ReportComment(commentId, reportRequest.Details, db.WithContext(executionData.Context)); err != nil {
			return nil, error_codes.NewErrorWithCodeRef(err, error_codes.GenericServerError)
		} else {
			return resp, nil
		}
	}, "/{id}/report", http.MethodPost, common.AccessLevelPublic, true, false)); err != nil {
		return err
	}

	if err := httpRouter.RegisterRestCmd(router.NewRestCommand(func(request []byte,
		executionData router.MethodExecutionData) (interface{}, *error_codes.ErrorWithCode) {
		contentId := utils.ExtractInt64(executionData.GetUserValue, "content_id", 0, 0)

		if contentId <= 0 {
			return nil, error_codes.NewErrorWithCodeRef(errors.New("invalid comment_id"), error_codes.GenericValidationError)
		}

		parentId := utils.ExtractInt64(executionData.GetUserValue, "parent_id", 0, 0)

		count := utils.ExtractInt64(executionData.GetUserValue, "count", 10, 10)
		after := utils.ExtractString(executionData.GetUserValue, "after", "")
		sortOrder := utils.ExtractString(executionData.GetUserValue, "sort_order", "")

		if resp, err := publicapi.GetCommentByTypeWithResourceId(publicapi.GetCommentsByTypeWithResourceRequest{
			ContentId: contentId,
			ParentId:  parentId,
			After:     after,
			Count:     count,
			SortOrder: sortOrder,
		}, executionData.UserId, db.WithContext(executionData.Context), userWrapper, executionData.ApmTransaction); err != nil {
			return nil, error_codes.NewErrorWithCodeRef(err, error_codes.GenericServerError)
		} else {
			return resp, nil
		}
	}, "/content/{content_id}", http.MethodDelete, common.AccessLevelPublic, false, false)); err != nil {
		return err
	}

	if err := httpRouter.RegisterRestCmd(router.NewRestCommand(func(request []byte,
		executionData router.MethodExecutionData) (interface{}, *error_codes.ErrorWithCode) {
		contentId := utils.ExtractInt64(executionData.GetUserValue, "content_id", 0, 0)

		if contentId <= 0 {
			return nil, error_codes.NewErrorWithCodeRef(errors.New("invalid comment_id"), error_codes.GenericValidationError)
		}

		var createRequest createCommentRequest

		if err := json.Unmarshal(request, &createRequest); err != nil {
			return nil, error_codes.NewErrorWithCodeRef(err, error_codes.GenericMappingError)
		}

		if resp, err := publicapi.SendContentComment(db.WithContext(executionData.Context), contentId,
			createRequest.Comment, createRequest.ParentId, contentWrapper, executionData.ApmTransaction,
			executionData.UserId); err != nil {
			return nil, error_codes.NewErrorWithCodeRef(err, error_codes.GenericServerError)
		} else {
			return resp, nil
		}
	}, "/{content_id}", http.MethodPost, common.AccessLevelPublic, true, true)); err != nil {
		return err
	}

	return nil
}
