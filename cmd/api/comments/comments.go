package comments

import (
	"encoding/json"
	"github.com/digitalmonsters/comments/pkg/comments"
	"github.com/digitalmonsters/comments/utils"
	"github.com/digitalmonsters/go-common/common"
	"github.com/digitalmonsters/go-common/error_codes"
	"github.com/digitalmonsters/go-common/router"
	"github.com/digitalmonsters/go-common/swagger"
	"github.com/digitalmonsters/go-common/wrappers/content"
	"github.com/digitalmonsters/go-common/wrappers/user"
	"github.com/digitalmonsters/go-common/wrappers/user_block"
	"github.com/pkg/errors"
	"gorm.io/gorm"
	"net/http"
)

func Init(httpRouter *router.HttpRouter, db *gorm.DB, userWrapper user.IUserWrapper, contentWrapper content.IContentWrapper,
	userBlockWrapper user_block.IUserBlockWrapper, apiDef map[string]swagger.ApiDescription) error {

	if err := httpRouter.RegisterRestCmd(router.NewRestCommand(func(request []byte,
		executionData router.MethodExecutionData) (interface{}, *error_codes.ErrorWithCode) {
		commentId := utils.ExtractInt64(executionData.GetUserValue, "comment_id", 0, 0)

		if commentId <= 0 {
			return nil, error_codes.NewErrorWithCodeRef(errors.New("invalid comment_id"), error_codes.GenericValidationError)
		}

		if comment, err := comments.GetCommendById(db.WithContext(executionData.Context), commentId, executionData.UserId,
			userWrapper, executionData.ApmTransaction); err != nil {
			return nil, error_codes.NewErrorWithCodeRef(err, error_codes.GenericValidationError)
		} else {
			return commentToFrontendCommentResponse(*comment), nil
		}
	}, "/{comment_id}", http.MethodGet, common.AccessLevelPublic, false, false)); err != nil {
		return err
	}

	apiDef["/{comment_id}"] = swagger.ApiDescription{
		Response:          frontendCommentResponse{},
		MethodDescription: "get comment by id",
		Tags:              []string{"comment"},
	}

	if err := httpRouter.RegisterRestCmd(router.NewRestCommand(func(request []byte,
		executionData router.MethodExecutionData) (interface{}, *error_codes.ErrorWithCode) {
		commentId := utils.ExtractInt64(executionData.GetUserValue, "delete_comment_id", 0, 0)

		if commentId <= 0 {
			return nil, error_codes.NewErrorWithCodeRef(errors.New("invalid comment_id"), error_codes.GenericValidationError)
		}

		if resp, err := comments.DeleteCommentById(db.WithContext(executionData.Context), commentId, executionData.UserId,
			contentWrapper, executionData.ApmTransaction); err != nil {
			return nil, error_codes.NewErrorWithCodeRef(err, error_codes.GenericValidationError)
		} else {
			return resp, nil
		}
	}, "/{delete_comment_id}", http.MethodDelete, common.AccessLevelPublic, true, true)); err != nil {
		return err
	}

	apiDef["/{delete_comment_id}"] = swagger.ApiDescription{
		Response:          comments.SimpleComment{},
		MethodDescription: "delete comment by id",
		Tags:              []string{"comment"},
	}

	if err := httpRouter.RegisterRestCmd(router.NewRestCommand(func(request []byte,
		executionData router.MethodExecutionData) (interface{}, *error_codes.ErrorWithCode) {
		commentId := utils.ExtractInt64(executionData.GetUserValue, "update_comment_id", 0, 0)

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

		if _, err := comments.UpdateCommentById(db.WithContext(executionData.Context), commentId,
			updateRequest.Comment, executionData.UserId); err != nil {
			return nil, error_codes.NewErrorWithCodeRef(err, error_codes.GenericValidationError)
		} else {
			return successResponse{
				Success: true,
			}, nil
		}
	}, "/{update_comment_id}", http.MethodPatch, common.AccessLevelPublic, true, true)); err != nil {
		return err
	}

	apiDef["/{update_comment_id}"] = swagger.ApiDescription{
		Response:          successResponse{},
		MethodDescription: "update comment by id",
		Tags:              []string{"comment"},
	}

	if err := httpRouter.RegisterRestCmd(router.NewRestCommand(func(request []byte,
		executionData router.MethodExecutionData) (interface{}, *error_codes.ErrorWithCode) {
		commentId := utils.ExtractInt64(executionData.GetUserValue, "comment_id", 0, 0)

		if commentId <= 0 {
			return nil, error_codes.NewErrorWithCodeRef(errors.New("invalid comment_id"), error_codes.GenericValidationError)
		}

		count := utils.ExtractInt64(executionData.GetUserValue, "count", 10, 10)
		after := utils.ExtractString(executionData.GetUserValue, "after", "")
		before := utils.ExtractString(executionData.GetUserValue, "before", "")
		sortOrder := utils.ExtractString(executionData.GetUserValue, "sort_order", "")

		if resp, err := comments.GetCommentsByContent(comments.GetCommentsByTypeWithResourceRequest{
			ParentId:  commentId,
			After:     after,
			Before:    before,
			Count:     count,
			SortOrder: sortOrder,
		}, executionData.UserId, db.WithContext(executionData.Context), userWrapper, executionData.ApmTransaction); err != nil {
			return nil, error_codes.NewErrorWithCodeRef(err, error_codes.GenericValidationError)
		} else {
			return resp, nil
		}
	}, "/{comment_id}/replies", http.MethodGet, common.AccessLevelPublic, false, false)); err != nil {
		return err
	}

	apiDef["/{comment_id}/replies"] = swagger.ApiDescription{
		Response:          frontendCommentPaginationResponse{},
		MethodDescription: "get replies by comment id",
		Tags:              []string{"comment"},
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
		before := utils.ExtractString(executionData.GetUserValue, "before", "")
		sortOrder := utils.ExtractString(executionData.GetUserValue, "sort_order", "")

		if resp, err := comments.GetCommentsByContent(comments.GetCommentsByTypeWithResourceRequest{
			ContentId: contentId,
			ParentId:  parentId,
			After:     after,
			Before:    before,
			Count:     count,
			SortOrder: sortOrder,
		}, executionData.UserId, db.WithContext(executionData.Context), userWrapper, executionData.ApmTransaction); err != nil {
			return nil, error_codes.NewErrorWithCodeRef(err, error_codes.GenericServerError)
		} else {
			return commentsWithPagingToFrontendPaginationResponse(*resp), nil
		}
	}, "/content/{content_id}", http.MethodGet, common.AccessLevelPublic, false, false)); err != nil {
		return err
	}

	apiDef["/content/{content_id}"] = swagger.ApiDescription{
		Response:          frontendCommentPaginationResponse{},
		MethodDescription: "get comments by content",
		Tags:              []string{"comment"},
	}

	if err := httpRouter.RegisterRestCmd(router.NewRestCommand(func(request []byte,
		executionData router.MethodExecutionData) (interface{}, *error_codes.ErrorWithCode) {
		contentId := utils.ExtractInt64(executionData.GetUserValue, "content_id_to_create_comment_on", 0, 0)

		if contentId <= 0 {
			return nil, error_codes.NewErrorWithCodeRef(errors.New("invalid comment_id"), error_codes.GenericValidationError)
		}

		var createRequest createCommentRequest

		if err := json.Unmarshal(request, &createRequest); err != nil {
			return nil, error_codes.NewErrorWithCodeRef(err, error_codes.GenericMappingError)
		}

		if resp, err := comments.CreateComment(db.WithContext(executionData.Context), contentId,
			createRequest.Comment, createRequest.ParentId, contentWrapper, userBlockWrapper, executionData.ApmTransaction,
			executionData.UserId); err != nil {
			return nil, error_codes.NewErrorWithCodeRef(err, error_codes.GenericServerError)
		} else {
			return createCommentResponse{
				Id:        resp.Id,
				Comment:   resp.Comment,
				AuthorId:  resp.AuthorId,
				ContentId: resp.ContentId,
			}, nil
		}
	}, "/content/{content_id_to_create_comment_on}", http.MethodPost, common.AccessLevelPublic, true, true)); err != nil {
		return err
	}

	apiDef["/content/{content_id_to_create_comment_on}"] = swagger.ApiDescription{
		Response:          createCommentResponse{},
		MethodDescription: "create comment on content",
		Tags:              []string{"comment"},
	}

	return nil
}
