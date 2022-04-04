package comments

import (
	"encoding/json"
	"github.com/digitalmonsters/comments/cmd/api/comments/notifiers/comment"
	"github.com/digitalmonsters/comments/cmd/api/comments/notifiers/content_comments_counter"
	"github.com/digitalmonsters/comments/cmd/api/comments/notifiers/user_comments_counter"
	"github.com/digitalmonsters/comments/pkg/comments"
	"github.com/digitalmonsters/comments/utils"
	"github.com/digitalmonsters/go-common/apm_helper"
	"github.com/digitalmonsters/go-common/error_codes"
	"github.com/digitalmonsters/go-common/router"
	"github.com/digitalmonsters/go-common/swagger"
	"github.com/digitalmonsters/go-common/wrappers/content"
	"github.com/digitalmonsters/go-common/wrappers/user_go"
	"github.com/pkg/errors"
	"gorm.io/gorm"
	"net/http"
	"strings"
)

func Init(httpRouter *router.HttpRouter, db *gorm.DB, userWrapper user_go.IUserGoWrapper, contentWrapper content.IContentWrapper,
	apiDef map[string]swagger.ApiDescription, commentNotifier *comment.Notifier, contentCommentsNotifier *content_comments_counter.Notifier,
	userCommentsNotifier *user_comments_counter.Notifier) error {

	if err := httpRouter.RegisterRestCmd(router.NewRestCommand(func(request []byte,
		executionData router.MethodExecutionData) (interface{}, *error_codes.ErrorWithCode) {
		commentId := utils.ExtractInt64(executionData.GetUserValue, "comment_id", 0, 0)

		if commentId <= 0 {
			return nil, error_codes.NewErrorWithCodeRef(errors.New("invalid comment_id"), error_codes.GenericValidationError)
		}

		if comment, err := comments.GetCommentById(db.WithContext(executionData.Context), commentId, executionData.UserId,
			userWrapper, executionData.ApmTransaction); err != nil {
			return nil, error_codes.NewErrorWithCodeRef(err, error_codes.GenericValidationError)
		} else {
			return commentToFrontendCommentWithCursorResponse(*comment), nil
		}
	}, "/{comment_id}", http.MethodGet, false, false)); err != nil {
		return err
	}

	apiDef["/{comment_id}"] = swagger.ApiDescription{
		Response:          frontendCommentResponse{},
		MethodDescription: "get comment by id",
		AdditionalSwaggerParameters: []swagger.ParameterDescription{
			{
				Name:        "comment_id",
				In:          swagger.ParameterInPath,
				Description: "comment_id",
				Required:    true,
				Type:        "integer",
			},
		},
		Tags: []string{"comment"},
	}

	if err := httpRouter.RegisterRestCmd(router.NewRestCommand(func(request []byte,
		executionData router.MethodExecutionData) (interface{}, *error_codes.ErrorWithCode) {

		if executionData.IsGuest {
			return nil, error_codes.NewErrorWithCodeRef(errors.New("registration required"), error_codes.RegistrationRequiredError)
		}

		commentId := utils.ExtractInt64(executionData.GetUserValue, "delete_comment_id", 0, 0)

		if commentId <= 0 {
			return nil, error_codes.NewErrorWithCodeRef(errors.New("invalid comment_id"), error_codes.GenericValidationError)
		}

		apm_helper.AddApmLabel(executionData.ApmTransaction, "user_id", executionData.UserId)
		apm_helper.AddApmLabel(executionData.ApmTransaction, "comment_id", commentId)

		if _, err := comments.DeleteCommentById(db.WithContext(executionData.Context), commentId, executionData.UserId,
			contentWrapper, executionData.ApmTransaction, commentNotifier, contentCommentsNotifier, userCommentsNotifier); err != nil {
			return nil, error_codes.NewErrorWithCodeRef(err, error_codes.GenericValidationError)
		} else {
			return successResponse{
				Success: true,
			}, nil
		}
	}, "/{delete_comment_id}", http.MethodDelete, true, true)); err != nil {
		return err
	}

	apiDef["/{delete_comment_id}"] = swagger.ApiDescription{
		Response:          successResponse{},
		MethodDescription: "delete comment by id",
		AdditionalSwaggerParameters: []swagger.ParameterDescription{
			{
				Name:        "delete_comment_id",
				In:          swagger.ParameterInPath,
				Description: "comment_id",
				Required:    true,
				Type:        "integer",
			},
		},
		Tags: []string{"comment"},
	}

	if err := httpRouter.RegisterRestCmd(router.NewRestCommand(func(request []byte,
		executionData router.MethodExecutionData) (interface{}, *error_codes.ErrorWithCode) {
		commentId := utils.ExtractInt64(executionData.GetUserValue, "update_comment_id", 0, 0)

		if executionData.IsGuest {
			return nil, error_codes.NewErrorWithCodeRef(errors.New("registration required"), error_codes.RegistrationRequiredError)
		}

		var updateRequest updateCommentRequest

		if err := json.Unmarshal(request, &updateRequest); err != nil {
			return nil, error_codes.NewErrorWithCodeRef(err, error_codes.GenericMappingError)
		}

		apm_helper.AddApmLabel(executionData.ApmTransaction, "user_id", executionData.UserId)
		apm_helper.AddApmLabel(executionData.ApmTransaction, "comment_id", commentId)

		if len(updateRequest.Comment) == 0 {
			return nil, error_codes.NewErrorWithCodeRef(errors.New("invalid comment length"), error_codes.GenericValidationError)
		}

		if commentId <= 0 {
			return nil, error_codes.NewErrorWithCodeRef(errors.New("invalid comment_id"), error_codes.GenericValidationError)
		}

		if _, err := comments.UpdateCommentById(db.WithContext(executionData.Context), commentId,
			updateRequest.Comment, executionData.UserId, contentWrapper, commentNotifier, executionData.ApmTransaction); err != nil {
			return nil, error_codes.NewErrorWithCodeRef(err, error_codes.GenericValidationError)
		} else {
			return successResponse{
				Success: true,
			}, nil
		}
	}, "/{update_comment_id}", http.MethodPatch, true, true)); err != nil {
		return err
	}

	apiDef["/{update_comment_id}"] = swagger.ApiDescription{
		Response:          successResponse{},
		MethodDescription: "update comment by id",
		AdditionalSwaggerParameters: []swagger.ParameterDescription{
			{
				Name:        "update_comment_id",
				In:          swagger.ParameterInPath,
				Description: "comment_id",
				Required:    true,
				Type:        "integer",
			},
		},
		Tags: []string{"comment"},
	}

	if err := httpRouter.RegisterRestCmd(router.NewRestCommand(func(request []byte,
		executionData router.MethodExecutionData) (interface{}, *error_codes.ErrorWithCode) {
		commentId := utils.ExtractInt64(executionData.GetUserValue, "comment_id", 0, 0)

		if commentId <= 0 {
			return nil, error_codes.NewErrorWithCodeRef(errors.New("invalid comment_id"), error_codes.GenericValidationError)
		}

		count := utils.ExtractInt64(executionData.GetUserValue, "count", 10, 50)
		after := utils.ExtractString(executionData.GetUserValue, "after", "")
		before := utils.ExtractString(executionData.GetUserValue, "before", "")
		sortOrder := utils.ExtractString(executionData.GetUserValue, "sort_order", "")

		apm_helper.AddApmLabel(executionData.ApmTransaction, "user_id", executionData.UserId)
		apm_helper.AddApmLabel(executionData.ApmTransaction, "comment_id", commentId)

		if resp, err := comments.GetCommentsByResourceId(comments.GetCommentsByTypeWithResourceRequest{
			After:      after,
			Before:     before,
			Count:      count,
			SortOrder:  sortOrder,
			ResourceId: commentId,
		}, executionData.UserId, db.WithContext(executionData.Context), userWrapper, executionData.ApmTransaction,
			comments.ResourceTypeParentComment); err != nil {
			return nil, error_codes.NewErrorWithCodeRef(err, error_codes.GenericValidationError)
		} else {
			return resp, nil
		}
	}, "/{comment_id}/replies", http.MethodGet, false, false)); err != nil {
		return err
	}

	apiDef["/{comment_id}/replies"] = swagger.ApiDescription{
		Response:          frontendCommentPaginationResponse{},
		MethodDescription: "get replies by comment id",
		AdditionalSwaggerParameters: []swagger.ParameterDescription{
			{
				Name:        "comment_id",
				In:          swagger.ParameterInPath,
				Description: "comment_id",
				Required:    true,
				Type:        "integer",
			},
			{
				Name:        "count",
				In:          swagger.ParameterInQuery,
				Description: "count per page",
				Required:    true,
				Type:        "integer",
			},
			{
				Name:        "after",
				In:          swagger.ParameterInQuery,
				Description: "cursor pagination",
				Required:    false,
				Type:        "string",
			},
			{
				Name:        "before",
				In:          swagger.ParameterInQuery,
				Description: "cursor pagination",
				Required:    false,
				Type:        "string",
			},
			{
				Name:        "sort_order",
				In:          swagger.ParameterInQuery,
				Description: "sort. top_reactions  newest most_replied least_popular oldest",
				Required:    false,
				Type:        "string",
			},
		},
		Tags: []string{"comment"},
	}

	if err := httpRouter.RegisterRestCmd(router.NewRestCommand(func(request []byte,
		executionData router.MethodExecutionData) (interface{}, *error_codes.ErrorWithCode) {
		resourceId := utils.ExtractInt64(executionData.GetUserValue, "resource_id", 0, 0)
		resourceType := utils.ExtractString(executionData.GetUserValue, "type", "content")

		if resourceId <= 0 {
			return nil, error_codes.NewErrorWithCodeRef(errors.New("invalid resource_id"), error_codes.GenericValidationError)
		}

		mappedResourceType := comments.ResourceTypeContent

		switch strings.ToLower(resourceType) {
		case "profile":
			mappedResourceType = comments.ResourceTypeProfile
		}

		count := utils.ExtractInt64(executionData.GetUserValue, "count", 10, 50)
		after := utils.ExtractString(executionData.GetUserValue, "after", "")
		before := utils.ExtractString(executionData.GetUserValue, "before", "")
		sortOrder := utils.ExtractString(executionData.GetUserValue, "sort_order", "")

		if resp, err := comments.GetCommentsByResourceId(comments.GetCommentsByTypeWithResourceRequest{
			ResourceId: resourceId,
			After:      after,
			Before:     before,
			Count:      count,
			SortOrder:  sortOrder,
		}, executionData.UserId, db.WithContext(executionData.Context), userWrapper, executionData.ApmTransaction, mappedResourceType); err != nil {
			return nil, error_codes.NewErrorWithCodeRef(err, error_codes.GenericServerError)
		} else {
			return commentsWithPagingToFrontendPaginationResponse(*resp), nil
		}
	}, "/{type}/{resource_id}", http.MethodGet, false, false)); err != nil {
		return err
	}

	apiDef["/{type}/{resource_id}"] = swagger.ApiDescription{
		Response:          frontendCommentPaginationResponse{},
		MethodDescription: "get comments by resource",
		AdditionalSwaggerParameters: []swagger.ParameterDescription{
			{
				Name:        "type",
				In:          swagger.ParameterInPath,
				Description: "resource type. comment || profile",
				Required:    true,
				Type:        "integer",
			},
			{
				Name:        "resource_id",
				In:          swagger.ParameterInPath,
				Description: "resource_id",
				Required:    true,
				Type:        "integer",
			},
			{
				Name:        "count",
				In:          swagger.ParameterInQuery,
				Description: "count per page",
				Required:    true,
				Type:        "integer",
			},
			{
				Name:        "after",
				In:          swagger.ParameterInQuery,
				Description: "cursor pagination",
				Required:    false,
				Type:        "string",
			},
			{
				Name:        "before",
				In:          swagger.ParameterInQuery,
				Description: "cursor pagination",
				Required:    false,
				Type:        "string",
			},
			{
				Name:        "sort_order",
				In:          swagger.ParameterInQuery,
				Description: "sort. top_reactions  newest most_replied least_popular oldest",
				Required:    false,
				Type:        "string",
			},
		},
		Tags: []string{"comment"},
	}

	if err := httpRouter.RegisterRestCmd(router.NewRestCommand(func(request []byte,
		executionData router.MethodExecutionData) (interface{}, *error_codes.ErrorWithCode) {
		contentId := utils.ExtractInt64(executionData.GetUserValue, "content_id_to_create_comment_on", 0, 0)

		if executionData.IsGuest {
			return nil, error_codes.NewErrorWithCodeRef(errors.New("registration required"), error_codes.RegistrationRequiredError)
		}

		if contentId <= 0 {
			return nil, error_codes.NewErrorWithCodeRef(errors.New("invalid comment_id"), error_codes.GenericValidationError)
		}

		var createRequest createCommentRequest

		if err := json.Unmarshal(request, &createRequest); err != nil {
			return nil, error_codes.NewErrorWithCodeRef(err, error_codes.GenericMappingError)
		}

		if resp, err := comments.CreateComment(db.WithContext(executionData.Context), contentId,
			createRequest.Comment, createRequest.ParentId, contentWrapper, userWrapper, executionData.ApmTransaction,
			executionData.UserId, commentNotifier, contentCommentsNotifier, userCommentsNotifier); err != nil {
			return nil, error_codes.NewErrorWithCodeRef(err, error_codes.GenericServerError)
		} else {
			return createCommentResponse{
				Id:        resp.Id,
				Comment:   resp.Comment,
				AuthorId:  resp.AuthorId,
				ContentId: resp.ContentId.ValueOrZero(),
			}, nil
		}
	}, "/content/{content_id_to_create_comment_on}", http.MethodPost, true, true)); err != nil {
		return err
	}

	apiDef["/content/{content_id_to_create_comment_on}"] = swagger.ApiDescription{
		Response:          createCommentResponse{},
		MethodDescription: "create comment on content",
		AdditionalSwaggerParameters: []swagger.ParameterDescription{
			{
				Name:        "content_id_to_create_comment_on",
				In:          swagger.ParameterInPath,
				Description: "content_id",
				Required:    true,
				Type:        "integer",
			},
		},
		Tags: []string{"comment"},
	}

	if err := httpRouter.RegisterRestCmd(router.NewRestCommand(func(request []byte,
		executionData router.MethodExecutionData) (interface{}, *error_codes.ErrorWithCode) {
		profileId := utils.ExtractInt64(executionData.GetUserValue, "profile_id_to_create_comment_on", 0, 0)

		if executionData.IsGuest {
			return nil, error_codes.NewErrorWithCodeRef(errors.New("registration required"), error_codes.RegistrationRequiredError)
		}

		if profileId <= 0 {
			return nil, error_codes.NewErrorWithCodeRef(errors.New("invalid profile_id"), error_codes.GenericValidationError)
		}

		var createRequest createCommentRequest

		if err := json.Unmarshal(request, &createRequest); err != nil {
			return nil, error_codes.NewErrorWithCodeRef(err, error_codes.GenericMappingError)
		}

		if resp, err := comments.CreateCommentOnProfile(db.WithContext(executionData.Context), profileId,
			createRequest.Comment, createRequest.ParentId, userWrapper, executionData.ApmTransaction,
			executionData.UserId, commentNotifier, contentCommentsNotifier, userCommentsNotifier); err != nil {
			return nil, error_codes.NewErrorWithCodeRef(err, error_codes.GenericServerError)
		} else {
			return createCommentOnProfileResponse{
				Id:       resp.Id,
				Comment:  resp.Comment,
				AuthorId: resp.AuthorId,
			}, nil
		}
	}, "/profile/{profile_id_to_create_comment_on}", http.MethodPost, true, true)); err != nil {
		return err
	}

	apiDef["/profile/{profile_id_to_create_comment_on}"] = swagger.ApiDescription{
		Response:          createCommentOnProfileResponse{},
		MethodDescription: "create comment on content",
		AdditionalSwaggerParameters: []swagger.ParameterDescription{
			{
				Name:        "profile_id_to_create_comment_on",
				In:          swagger.ParameterInPath,
				Description: "profile_id",
				Required:    true,
				Type:        "integer",
			},
		},
		Tags: []string{"comment"},
	}

	return nil
}
