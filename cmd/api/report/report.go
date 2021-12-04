package report

import (
	"encoding/json"
	"github.com/digitalmonsters/comments/pkg/publicapi"
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

	return nil
}
