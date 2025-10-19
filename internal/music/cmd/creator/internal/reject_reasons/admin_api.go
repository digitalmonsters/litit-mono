package reject_reasons

import (
	"encoding/json"
	"github.com/digitalmonsters/go-common/common"
	"github.com/digitalmonsters/go-common/error_codes"
	"github.com/digitalmonsters/go-common/router"
	"github.com/digitalmonsters/go-common/swagger"
	"github.com/digitalmonsters/music/pkg/creators/reject_reasons"
	"github.com/digitalmonsters/music/pkg/database"
)

func (c *rejectReasonsApp) initAdminApi() error {
	adminCommands := []router.ICommand{
		c.creatorRejectReasonUpsert(),
		c.creatorRejectReasonList(),
		c.creatorRejectReasonDelete(),
	}

	for _, command := range adminCommands {
		if err := c.httpRouter.GetRpcAdminEndpoint().RegisterRpcCommand(command); err != nil {
			return err
		}
	}
	return nil
}

func (c *rejectReasonsApp) creatorRejectReasonUpsert() router.ICommand {
	method := "CreatorRejectReasonUpsertAdmin"

	c.apiDef[method] = swagger.ApiDescription{
		Request:  reject_reasons.UpsertRequest{},
		Response: []database.CreatorRejectReasons{},
		Tags:     []string{"reject_reason"},
	}

	return router.NewAdminCommand(method, func(request []byte, executionData router.MethodExecutionData) (interface{}, *error_codes.ErrorWithCode) {
		var req reject_reasons.UpsertRequest
		if err := json.Unmarshal(request, &req); err != nil {
			return nil, error_codes.NewErrorWithCodeRef(err, error_codes.GenericMappingError)
		}

		res, err := reject_reasons.Upsert(req, database.GetDb(database.DbTypeMaster).WithContext(executionData.Context))
		if err != nil {
			return nil, error_codes.NewErrorWithCodeRef(err, error_codes.GenericServerError)
		}

		return res, nil
	}, common.AccessLevelWrite, "music:creator:crud:upsert")
}

func (c *rejectReasonsApp) creatorRejectReasonList() router.ICommand {
	method := "CreatorRejectReasonListAdmin"

	c.apiDef[method] = swagger.ApiDescription{
		Request:  reject_reasons.ListRequest{},
		Response: reject_reasons.ListResponse{},
		Tags:     []string{"reject_reason"},
	}

	return router.NewAdminCommand(method, func(request []byte, executionData router.MethodExecutionData) (interface{}, *error_codes.ErrorWithCode) {
		var req reject_reasons.ListRequest
		if err := json.Unmarshal(request, &req); err != nil {
			return nil, error_codes.NewErrorWithCodeRef(err, error_codes.GenericMappingError)
		}

		res, err := reject_reasons.List(req, database.GetDb(database.DbTypeReadonly).WithContext(executionData.Context))
		if err != nil {
			return nil, error_codes.NewErrorWithCodeRef(err, error_codes.GenericServerError)
		}

		return res, nil
	}, common.AccessLevelPublic, "music:reason:list")
}

func (c *rejectReasonsApp) creatorRejectReasonDelete() router.ICommand {
	method := "CreatorRejectReasonDeleteAdmin"

	c.apiDef[method] = swagger.ApiDescription{
		Request:  reject_reasons.DeleteRequest{},
		Response: nil,
		Tags:     []string{"reject_reason"},
	}

	return router.NewAdminCommand(method, func(request []byte, executionData router.MethodExecutionData) (interface{}, *error_codes.ErrorWithCode) {
		var req reject_reasons.DeleteRequest
		if err := json.Unmarshal(request, &req); err != nil {
			return nil, error_codes.NewErrorWithCodeRef(err, error_codes.GenericMappingError)
		}

		err := reject_reasons.Delete(req, database.GetDb(database.DbTypeMaster).WithContext(executionData.Context))
		if err != nil {
			return nil, error_codes.NewErrorWithCodeRef(err, error_codes.GenericServerError)
		}

		return "ok", nil
	}, common.AccessLevelPublic, "music:reason:crud:delete")
}
