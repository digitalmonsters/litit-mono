package categories

import (
	"encoding/json"
	"github.com/digitalmonsters/go-common/common"
	"github.com/digitalmonsters/go-common/error_codes"
	"github.com/digitalmonsters/go-common/router"
	"github.com/digitalmonsters/go-common/swagger"
	"github.com/digitalmonsters/music/pkg/creators/categories"
	"github.com/digitalmonsters/music/pkg/database"
)

func (c *categoriesApp) initAdminApi() error {
	adminCommands := []router.ICommand{
		c.categoriesUpsert(),
		c.categoriesList(),
		c.categoriesDelete(),
	}

	for _, command := range adminCommands {
		if err := c.httpRouter.GetRpcAdminEndpoint().RegisterRpcCommand(command); err != nil {
			return err
		}
	}
	return nil
}

func (c *categoriesApp) categoriesUpsert() router.ICommand {
	method := "CategoriesUpsertAdmin"

	c.apiDef[method] = swagger.ApiDescription{
		Request:  categories.UpsertRequest{},
		Response: []database.Category{},
		Tags:     []string{"categories"},
	}
	return router.NewAdminCommand(method, func(request []byte, executionData router.MethodExecutionData) (interface{}, *error_codes.ErrorWithCode) {
		var req categories.UpsertRequest
		if err := json.Unmarshal(request, &req); err != nil {
			return nil, error_codes.NewErrorWithCodeRef(err, error_codes.GenericMappingError)
		}

		res, err := categories.Upsert(req, database.GetDb(database.DbTypeMaster).WithContext(executionData.Context))
		if err != nil {
			return nil, error_codes.NewErrorWithCodeRef(err, error_codes.GenericServerError)
		}

		return res, nil
	}, common.AccessLevelWrite, "music:category:crud:upsert")
}

func (c *categoriesApp) categoriesList() router.ICommand {
	method := "CategoriesListAdmin"

	c.apiDef[method] = swagger.ApiDescription{
		Request:  categories.ListRequest{},
		Response: categories.ListResponse{},
		Tags:     []string{"categories"},
	}
	return router.NewAdminCommand(method, func(request []byte, executionData router.MethodExecutionData) (interface{}, *error_codes.ErrorWithCode) {
		var req categories.ListRequest
		if err := json.Unmarshal(request, &req); err != nil {
			return nil, error_codes.NewErrorWithCodeRef(err, error_codes.GenericMappingError)
		}

		res, err := categories.AdminList(req, database.GetDb(database.DbTypeReadonly).WithContext(executionData.Context))
		if err != nil {
			return nil, error_codes.NewErrorWithCodeRef(err, error_codes.GenericServerError)
		}

		return res, nil
	}, common.AccessLevelRead, "music:category:list")
}

func (c *categoriesApp) categoriesDelete() router.ICommand {
	method := "CategoriesDeleteAdmin"

	c.apiDef[method] = swagger.ApiDescription{
		Request:  categories.DeleteRequest{},
		Response: nil,
		Tags:     []string{"categories"},
	}
	return router.NewAdminCommand(method, func(request []byte, executionData router.MethodExecutionData) (interface{}, *error_codes.ErrorWithCode) {
		var req categories.DeleteRequest
		if err := json.Unmarshal(request, &req); err != nil {
			return nil, error_codes.NewErrorWithCodeRef(err, error_codes.GenericMappingError)
		}

		err := categories.Delete(req, database.GetDb(database.DbTypeMaster).WithContext(executionData.Context))
		if err != nil {
			return nil, error_codes.NewErrorWithCodeRef(err, error_codes.GenericServerError)
		}

		return "ok", nil
	}, common.AccessLevelWrite, "music:category:crud:delete")
}
