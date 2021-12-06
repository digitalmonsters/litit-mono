package vote

import (
	"encoding/json"
	"github.com/digitalmonsters/comments/pkg/vote"
	"github.com/digitalmonsters/comments/utils"
	"github.com/digitalmonsters/go-common/common"
	"github.com/digitalmonsters/go-common/error_codes"
	"github.com/digitalmonsters/go-common/router"
	"github.com/digitalmonsters/go-common/swagger"
	"github.com/pkg/errors"
	"gorm.io/gorm"
	"net/http"
)

func Init(httpRouter *router.HttpRouter, db *gorm.DB, def map[string]swagger.ApiDescription) error {
	if err := httpRouter.RegisterRestCmd(router.NewRestCommand(func(request []byte,
		executionData router.MethodExecutionData) (interface{}, *error_codes.ErrorWithCode) {
		commentId := utils.ExtractInt64(executionData.GetUserValue, "comment_id", 0, 0)

		if commentId <= 0 {
			return nil, error_codes.NewErrorWithCodeRef(errors.New("invalid comment_id"), error_codes.GenericValidationError)
		}

		var reportRequest voteRequest

		if err := json.Unmarshal(request, &reportRequest); err != nil {
			return nil, error_codes.NewErrorWithCodeRef(err, error_codes.GenericMappingError)
		}

		if _, err := vote.VoteComment(db.WithContext(executionData.Context), commentId,
			reportRequest.VoteUp, executionData.UserId); err != nil {
			return nil, error_codes.NewErrorWithCodeRef(err, error_codes.GenericServerError)
		} else {
			return successResponse{
				Success: true,
			}, nil
		}
	}, "/{comment_id}/vote", http.MethodPost, common.AccessLevelPublic, true, false)); err != nil {
		return err
	}

	def["/{comment_id}/vote"] = swagger.ApiDescription{
		Request: voteRequest{},
		AdditionalSwaggerParameters: []swagger.ParameterDescription{
			{
				Name:        "comment_id",
				In:          swagger.ParameterInPath,
				Description: "comment_id",
				Required:    true,
				Type:        "integer",
			},
		},
		Response:          successResponse{},
		MethodDescription: "report comment",
		Tags:              []string{"report"},
	}

	return nil
}
