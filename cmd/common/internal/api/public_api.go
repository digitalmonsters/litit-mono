package api

import (
	"encoding/json"

	commonPkg "github.com/digitalmonsters/ads-manager/pkg/common"
	"github.com/digitalmonsters/ads-manager/pkg/database"
	"github.com/digitalmonsters/go-common/error_codes"
	"github.com/digitalmonsters/go-common/router"
)

func (a *apiApp) initPublicApi(httpRouter *router.HttpRouter) error {
	restCommands := []*router.RestCommand{
		a.publicListActionButtons(),
	}

	for _, c := range restCommands {
		if err := httpRouter.RegisterRestCmd(c); err != nil {
			return err
		}
	}

	return nil
}

func (a *apiApp) publicListActionButtons() *router.RestCommand {
	path := "/list_action_buttons"

	return router.NewRestCommand(func(request []byte, executionData router.MethodExecutionData) (interface{}, *error_codes.ErrorWithCode) {
		var req commonPkg.PublicListActionButtonsRequest
		if len(request) > 0 {
			if err := json.Unmarshal(request, &req); err != nil {
				return nil, error_codes.NewErrorWithCodeRef(err, error_codes.GenericMappingError)
			}
		}
		resp, err := a.commonService.PublicListActionButtons(req, database.GetDbWithContext(database.DbTypeReadonly, executionData.Context))
		if err != nil {
			return nil, error_codes.NewErrorWithCodeRef(err, error_codes.GenericServerError)
		}

		return resp, nil
	}, path, router.MethodPost).RequireIdentityValidation().Build()
}
