package creator

import (
	"encoding/json"
	"github.com/digitalmonsters/go-common/common"
	"github.com/digitalmonsters/go-common/error_codes"
	"github.com/digitalmonsters/go-common/router"
	"github.com/digitalmonsters/go-common/swagger"
	"github.com/digitalmonsters/go-common/wrappers/user"
	"github.com/digitalmonsters/music/configs"
	"github.com/digitalmonsters/music/pkg/creators"
	"github.com/digitalmonsters/music/pkg/creators/reject_reasons"
	"github.com/digitalmonsters/music/pkg/database"
)

func InitAdminApi(adminEndpoint router.IRpcEndpoint, apiDef map[string]swagger.ApiDescription, cfg configs.Settings, userWrapper user.IUserWrapper) error {
	if err := adminEndpoint.RegisterRpcCommand(router.NewAdminCommand("CreatorRequestsListAdmin", func(request []byte, executionData router.MethodExecutionData) (interface{}, *error_codes.ErrorWithCode) {
		var req creators.CreatorRequestsListRequest
		if err := json.Unmarshal(request, &req); err != nil {
			return nil, error_codes.NewErrorWithCodeRef(err, error_codes.GenericMappingError)
		}

		res, err := creators.CreatorRequestsList(req, database.GetDb(database.DbTypeReadonly).WithContext(executionData.Context), cfg.Creators.MaxThresholdHours, executionData.ApmTransaction, userWrapper)
		if err != nil {
			return nil, error_codes.NewErrorWithCodeRef(err, error_codes.GenericServerError)
		}

		return res, nil
	}, common.AccessLevelRead, "music:creator:list")); err != nil {
		return err
	}

	if err := adminEndpoint.RegisterRpcCommand(router.NewAdminCommand("CreatorRequestApproveAdmin", func(request []byte, executionData router.MethodExecutionData) (interface{}, *error_codes.ErrorWithCode) {
		var req creators.CreatorRequestApproveRequest
		if err := json.Unmarshal(request, &req); err != nil {
			return nil, error_codes.NewErrorWithCodeRef(err, error_codes.GenericMappingError)
		}

		res, err := creators.CreatorRequestApprove(req, database.GetDb(database.DbTypeMaster).WithContext(executionData.Context))
		if err != nil {
			return nil, error_codes.NewErrorWithCodeRef(err, error_codes.GenericServerError)
		}

		return res, nil
	}, common.AccessLevelWrite, "music:creator:crud:approve")); err != nil {
		return err
	}

	if err := adminEndpoint.RegisterRpcCommand(router.NewAdminCommand("CreatorRequestRejectAdmin", func(request []byte, executionData router.MethodExecutionData) (interface{}, *error_codes.ErrorWithCode) {
		var req creators.CreatorRequestRejectRequest
		if err := json.Unmarshal(request, &req); err != nil {
			return nil, error_codes.NewErrorWithCodeRef(err, error_codes.GenericMappingError)
		}

		res, err := creators.CreatorRequestReject(req, database.GetDb(database.DbTypeMaster).WithContext(executionData.Context))
		if err != nil {
			return nil, error_codes.NewErrorWithCodeRef(err, error_codes.GenericServerError)
		}

		return res, nil
	}, common.AccessLevelWrite, "music:creator:crud:reject")); err != nil {
		return err
	}

	if err := adminEndpoint.RegisterRpcCommand(router.NewAdminCommand("CreatorRejectReasonUpsertAdmin", func(request []byte, executionData router.MethodExecutionData) (interface{}, *error_codes.ErrorWithCode) {
		var req reject_reasons.UpsertRequest
		if err := json.Unmarshal(request, &req); err != nil {
			return nil, error_codes.NewErrorWithCodeRef(err, error_codes.GenericMappingError)
		}

		res, err := reject_reasons.Upsert(req, database.GetDb(database.DbTypeMaster).WithContext(executionData.Context))
		if err != nil {
			return nil, error_codes.NewErrorWithCodeRef(err, error_codes.GenericServerError)
		}

		return res, nil
	}, common.AccessLevelWrite, "music:creator:crud:upsert")); err != nil {
		return err
	}

	if err := adminEndpoint.RegisterRpcCommand(router.NewAdminCommand("CreatorRejectReasonListAdmin", func(request []byte, executionData router.MethodExecutionData) (interface{}, *error_codes.ErrorWithCode) {
		var req reject_reasons.ListRequest
		if err := json.Unmarshal(request, &req); err != nil {
			return nil, error_codes.NewErrorWithCodeRef(err, error_codes.GenericMappingError)
		}

		res, err := reject_reasons.List(req, database.GetDb(database.DbTypeReadonly).WithContext(executionData.Context))
		if err != nil {
			return nil, error_codes.NewErrorWithCodeRef(err, error_codes.GenericServerError)
		}

		return res, nil
	}, common.AccessLevelPublic, "music:reason:list")); err != nil {
		return err
	}

	if err := adminEndpoint.RegisterRpcCommand(router.NewAdminCommand("CreatorRejectReasonDeleteAdmin", func(request []byte, executionData router.MethodExecutionData) (interface{}, *error_codes.ErrorWithCode) {
		var req reject_reasons.DeleteRequest
		if err := json.Unmarshal(request, &req); err != nil {
			return nil, error_codes.NewErrorWithCodeRef(err, error_codes.GenericMappingError)
		}

		err := reject_reasons.Delete(req, database.GetDb(database.DbTypeMaster).WithContext(executionData.Context))
		if err != nil {
			return nil, error_codes.NewErrorWithCodeRef(err, error_codes.GenericServerError)
		}

		return "ok", nil
	}, common.AccessLevelPublic, "music:reason:crud:delete")); err != nil {
		return err
	}

	apiDef["CreatorRequestsListAdmin"] = swagger.ApiDescription{
		Request:  creators.CreatorRequestsListRequest{},
		Response: creators.CreatorRequestsListResponse{},
		Tags:     []string{"creator"},
	}

	apiDef["CreatorRequestApproveAdmin"] = swagger.ApiDescription{
		Request:  creators.CreatorRequestApproveRequest{},
		Response: []database.Creator{},
		Tags:     []string{"creator"},
	}

	apiDef["CreatorRequestRejectAdmin"] = swagger.ApiDescription{
		Request:  creators.CreatorRequestRejectRequest{},
		Response: []database.Creator{},
		Tags:     []string{"creator"},
	}

	apiDef["CreatorRejectReasonUpsertAdmin"] = swagger.ApiDescription{
		Request:  reject_reasons.UpsertRequest{},
		Response: []database.CreatorRejectReasons{},
		Tags:     []string{"reject_reason"},
	}

	apiDef["CreatorRejectReasonListAdmin"] = swagger.ApiDescription{
		Request:  reject_reasons.ListRequest{},
		Response: reject_reasons.ListResponse{},
		Tags:     []string{"reject_reason"},
	}

	apiDef["CreatorRejectReasonDeleteAdmin"] = swagger.ApiDescription{
		Request:  reject_reasons.DeleteRequest{},
		Response: nil,
		Tags:     []string{"reject_reason"},
	}

	return nil
}
