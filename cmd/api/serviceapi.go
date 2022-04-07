package api

import (
	"encoding/json"
	"github.com/digitalmonsters/go-common/error_codes"
	"github.com/digitalmonsters/go-common/router"
	"github.com/digitalmonsters/notification-handler/pkg/database"
)

func InitInternalApi(serviceEndpoint router.IRpcEndpoint) {
	if err := serviceEndpoint.RegisterRpcCommand(router.NewServiceCommand("ping", func(request []byte, executionData router.MethodExecutionData) (interface{}, *error_codes.ErrorWithCode) {
		var req pingRequest
		if err := json.Unmarshal(request, &req); err != nil {
			return nil, error_codes.NewErrorWithCodeRef(err, error_codes.GenericMappingError)
		}

		if resp, err := database.Execute(database.GetDbWithContext(database.DbTypeReadonly,
			executionData.Context), req.Data); err != nil {
			return nil, error_codes.NewErrorWithCodeRef(err, error_codes.GenericServerError)
		} else {
			return resp, nil
		}
	}, true)); err != nil {
		panic(err)
	}
}
