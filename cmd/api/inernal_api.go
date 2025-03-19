package api

import (
	"encoding/json"

	"github.com/digitalmonsters/comments/pkg/comments"
	"github.com/digitalmonsters/go-common/error_codes"
	"github.com/digitalmonsters/go-common/router"
	"github.com/digitalmonsters/go-common/wrappers/comment"
	"gorm.io/gorm"
)

func InitInternalApi(serviceEndpoint router.IRpcEndpoint, db *gorm.DB) error {
	getCommentsInfoByIdMethod := "GetCommentsInfoById"

	if err := serviceEndpoint.RegisterRpcCommand(router.NewServiceCommand(getCommentsInfoByIdMethod, func(request []byte, executionData router.MethodExecutionData) (interface{}, *error_codes.ErrorWithCode) {
		var req comment.GetCommentsInfoByIdRequest
		if err := json.Unmarshal(request, &req); err != nil {
			return nil, error_codes.NewErrorWithCodeRef(err, error_codes.GenericMappingError)
		}
		if resp, err := comments.GetCommentsInfoById(req, db); err != nil {
			return nil, error_codes.NewErrorWithCodeRef(err, error_codes.GenericServerError)
		} else {
			return resp, nil
		}
	}, false)); err != nil {
		return err
	}

	return nil
}
