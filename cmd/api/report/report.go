package report

import (
	"encoding/json"
	"github.com/digitalmonsters/comments/pkg/report"
	"github.com/digitalmonsters/comments/utils"
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

		if executionData.IsGuest {
			return nil, error_codes.NewErrorWithCodeRef(errors.New("registration required"), error_codes.RegistrationRequiredError)
		}

		if commentId <= 0 {
			return nil, error_codes.NewErrorWithCodeRef(errors.New("invalid comment_id"), error_codes.GenericValidationError)
		}

		var reportRequest reportCommentRequest

		if err := json.Unmarshal(request, &reportRequest); err != nil {
			return nil, error_codes.NewErrorWithCodeRef(err, error_codes.GenericMappingError)
		}

		tx := db.WithContext(executionData.Context)
		defer tx.Rollback()

		if _, err := report.ReportComment(commentId, reportRequest.Details, tx,
			executionData.UserId, reportRequest.Type); err != nil {
			return nil, error_codes.NewErrorWithCodeRef(err, error_codes.GenericServerError)
		} else {
			if err := tx.Commit().Error; err != nil {
				return nil, error_codes.NewErrorWithCodeRef(errors.WithStack(err), error_codes.GenericServerError)
			}
			return successResponse{
				Success: true,
			}, nil
		}
	}, "/{comment_id}/report", http.MethodPost).RequireIdentityValidation().Build()); err != nil {
		return err
	}

	def["/{comment_id}/report"] = swagger.ApiDescription{
		Request:  reportCommentRequest{},
		Response: successResponse{},
		AdditionalSwaggerParameters: []swagger.ParameterDescription{
			{
				Name:        "comment_id",
				In:          swagger.ParameterInPath,
				Description: "comment_id",
				Required:    true,
				Type:        "integer",
			},
		},
		MethodDescription: "report comment",
		Tags:              []string{"report"},
	}

	return nil
}
