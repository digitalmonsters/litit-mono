package template

import (
	"encoding/json"
	"github.com/digitalmonsters/go-common/common"
	"github.com/digitalmonsters/go-common/error_codes"
	"github.com/digitalmonsters/go-common/router"
	"github.com/digitalmonsters/go-common/swagger"
	"github.com/digitalmonsters/notification-handler/pkg/database"
	"github.com/digitalmonsters/notification-handler/pkg/template"
)

func (a templateApp) initAdminApi(endpoint router.IRpcEndpoint) error {
	commands := []router.ICommand{
		a.editTemplate(),
		a.listTemplates(),
	}

	for _, c := range commands {
		if err := endpoint.RegisterRpcCommand(c); err != nil {
			return err
		}
	}

	return nil
}

func (a templateApp) editTemplate() router.ICommand {
	method := "EditTemplate"

	a.apiDef[method] = swagger.ApiDescription{
		Request: template.EditTemplateRequest{},
		Tags:    []string{"template"},
	}

	return router.NewAdminCommand(method, func(request []byte,
		executionData router.MethodExecutionData) (interface{}, *error_codes.ErrorWithCode) {
		var req template.EditTemplateRequest

		if err := json.Unmarshal(request, &req); err != nil {
			return nil, error_codes.NewErrorWithCodeRef(err, error_codes.GenericMappingError)
		}

		tx := database.GetDbWithContext(database.DbTypeMaster, executionData.Context).Begin()
		defer tx.Rollback()

		if err := a.templateService.EditTemplate(req, tx); err != nil {
			return nil, error_codes.NewErrorWithCodeRef(err, error_codes.GenericServerError)
		}

		if err := tx.Commit().Error; err != nil {
			return nil, error_codes.NewErrorWithCodeRef(err, error_codes.GenericServerError)
		}

		return nil, nil
	}, common.AccessLevelWrite, "notification_template")
}

func (a templateApp) listTemplates() router.ICommand {
	method := "ListTemplates"

	a.apiDef[method] = swagger.ApiDescription{
		Request:  template.ListTemplatesRequest{},
		Response: template.ListTemplatesResponse{},
		Tags:     []string{"template"},
	}

	return router.NewAdminCommand(method, func(request []byte,
		executionData router.MethodExecutionData) (interface{}, *error_codes.ErrorWithCode) {
		var req template.ListTemplatesRequest

		if err := json.Unmarshal(request, &req); err != nil {
			return nil, error_codes.NewErrorWithCodeRef(err, error_codes.GenericMappingError)
		}

		resp, err := a.templateService.ListTemplates(req, database.GetDbWithContext(database.DbTypeReadonly, executionData.Context))
		if err != nil {
			return nil, error_codes.NewErrorWithCodeRef(err, error_codes.GenericServerError)
		}

		return resp, nil
	}, common.AccessLevelRead, "notification_template")
}
