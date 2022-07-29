package moods

import (
	"encoding/json"
	"github.com/digitalmonsters/go-common/common"
	"github.com/digitalmonsters/go-common/error_codes"
	"github.com/digitalmonsters/go-common/router"
	"github.com/digitalmonsters/go-common/swagger"
	"github.com/digitalmonsters/music/pkg/creators/moods"
	"github.com/digitalmonsters/music/pkg/database"
)

func (c *moodsApp) initAdminApi() error {
	adminCommands := []router.ICommand{
		c.moodsUpsert(),
		c.moodsList(),
		c.moodsDelete(),
	}

	for _, command := range adminCommands {
		if err := c.httpRouter.GetRpcAdminEndpoint().RegisterRpcCommand(command); err != nil {
			return err
		}
	}
	return nil
}

func (c *moodsApp) moodsUpsert() router.ICommand {
	method := "MoodsUpsertAdmin"

	c.apiDef[method] = swagger.ApiDescription{
		Request:  moods.UpsertRequest{},
		Response: []database.Mood{},
		Tags:     []string{"moods"},
	}

	return router.NewAdminCommand(method, func(request []byte, executionData router.MethodExecutionData) (interface{}, *error_codes.ErrorWithCode) {
		var req moods.UpsertRequest
		if err := json.Unmarshal(request, &req); err != nil {
			return nil, error_codes.NewErrorWithCodeRef(err, error_codes.GenericMappingError)
		}

		res, err := moods.Upsert(req, database.GetDb(database.DbTypeMaster).WithContext(executionData.Context))
		if err != nil {
			return nil, error_codes.NewErrorWithCodeRef(err, error_codes.GenericServerError)
		}

		return res, nil
	}, common.AccessLevelWrite, "music:mood:crud:upsert")
}

func (c *moodsApp) moodsList() router.ICommand {
	method := "MoodsListAdmin"

	c.apiDef[method] = swagger.ApiDescription{
		Request:  moods.ListRequest{},
		Response: moods.ListResponse{},
		Tags:     []string{"moods"},
	}
	return router.NewAdminCommand(method, func(request []byte, executionData router.MethodExecutionData) (interface{}, *error_codes.ErrorWithCode) {
		var req moods.ListRequest
		if err := json.Unmarshal(request, &req); err != nil {
			return nil, error_codes.NewErrorWithCodeRef(err, error_codes.GenericMappingError)
		}

		res, err := moods.AdminList(req, database.GetDb(database.DbTypeReadonly).WithContext(executionData.Context))
		if err != nil {
			return nil, error_codes.NewErrorWithCodeRef(err, error_codes.GenericServerError)
		}

		return res, nil
	}, common.AccessLevelRead, "music:mood:list")
}

func (c *moodsApp) moodsDelete() router.ICommand {
	method := "MoodsDeleteAdmin"

	c.apiDef[method] = swagger.ApiDescription{
		Request:  moods.DeleteRequest{},
		Response: nil,
		Tags:     []string{"moods"},
	}

	return router.NewAdminCommand(method, func(request []byte, executionData router.MethodExecutionData) (interface{}, *error_codes.ErrorWithCode) {
		var req moods.DeleteRequest
		if err := json.Unmarshal(request, &req); err != nil {
			return nil, error_codes.NewErrorWithCodeRef(err, error_codes.GenericMappingError)
		}

		err := moods.Delete(req, database.GetDb(database.DbTypeMaster).WithContext(executionData.Context))
		if err != nil {
			return nil, error_codes.NewErrorWithCodeRef(err, error_codes.GenericServerError)
		}

		return "ok", nil
	}, common.AccessLevelWrite, "music:mood:crud:delete")
}
