package creator

import (
	"encoding/json"
	"github.com/digitalmonsters/go-common/common"
	"github.com/digitalmonsters/go-common/error_codes"
	"github.com/digitalmonsters/go-common/router"
	"github.com/digitalmonsters/go-common/swagger"
	"github.com/digitalmonsters/go-common/wrappers/user_go"
	"github.com/digitalmonsters/music/configs"
	"github.com/digitalmonsters/music/pkg/creators"
	"github.com/digitalmonsters/music/pkg/creators/categories"
	"github.com/digitalmonsters/music/pkg/creators/moderation"
	"github.com/digitalmonsters/music/pkg/creators/moods"
	"github.com/digitalmonsters/music/pkg/creators/reject_reasons"
	"github.com/digitalmonsters/music/pkg/database"
)

func InitAdminApi(adminEndpoint router.IRpcEndpoint, apiDef map[string]swagger.ApiDescription, cfg configs.Settings, userGoWrapper user_go.IUserGoWrapper, creatorsService *creators.Service) error {
	if err := adminEndpoint.RegisterRpcCommand(router.NewAdminCommand("CreatorRequestsListAdmin", func(request []byte, executionData router.MethodExecutionData) (interface{}, *error_codes.ErrorWithCode) {
		var req creators.CreatorRequestsListRequest
		if err := json.Unmarshal(request, &req); err != nil {
			return nil, error_codes.NewErrorWithCodeRef(err, error_codes.GenericMappingError)
		}

		res, err := creatorsService.CreatorRequestsList(req, database.GetDb(database.DbTypeReadonly).WithContext(executionData.Context), cfg.Creators.MaxThresholdHours, executionData.ApmTransaction, userGoWrapper)
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

		res, err := creatorsService.CreatorRequestApprove(req, database.GetDb(database.DbTypeMaster).WithContext(executionData.Context))
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

		res, err := creatorsService.CreatorRequestReject(req, database.GetDb(database.DbTypeMaster).WithContext(executionData.Context))
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

	if err := adminEndpoint.RegisterRpcCommand(router.NewAdminCommand("CategoriesUpsertAdmin", func(request []byte, executionData router.MethodExecutionData) (interface{}, *error_codes.ErrorWithCode) {
		var req categories.UpsertRequest
		if err := json.Unmarshal(request, &req); err != nil {
			return nil, error_codes.NewErrorWithCodeRef(err, error_codes.GenericMappingError)
		}

		res, err := categories.Upsert(req, database.GetDb(database.DbTypeMaster).WithContext(executionData.Context))
		if err != nil {
			return nil, error_codes.NewErrorWithCodeRef(err, error_codes.GenericServerError)
		}

		return res, nil
	}, common.AccessLevelWrite, "music:category:crud:upsert")); err != nil {
		return err
	}

	if err := adminEndpoint.RegisterRpcCommand(router.NewAdminCommand("CategoriesListAdmin", func(request []byte, executionData router.MethodExecutionData) (interface{}, *error_codes.ErrorWithCode) {
		var req categories.ListRequest
		if err := json.Unmarshal(request, &req); err != nil {
			return nil, error_codes.NewErrorWithCodeRef(err, error_codes.GenericMappingError)
		}

		res, err := categories.AdminList(req, database.GetDb(database.DbTypeReadonly).WithContext(executionData.Context))
		if err != nil {
			return nil, error_codes.NewErrorWithCodeRef(err, error_codes.GenericServerError)
		}

		return res, nil
	}, common.AccessLevelRead, "music:category:list")); err != nil {
		return err
	}

	if err := adminEndpoint.RegisterRpcCommand(router.NewAdminCommand("CategoriesDeleteAdmin", func(request []byte, executionData router.MethodExecutionData) (interface{}, *error_codes.ErrorWithCode) {
		var req categories.DeleteRequest
		if err := json.Unmarshal(request, &req); err != nil {
			return nil, error_codes.NewErrorWithCodeRef(err, error_codes.GenericMappingError)
		}

		err := categories.Delete(req, database.GetDb(database.DbTypeMaster).WithContext(executionData.Context))
		if err != nil {
			return nil, error_codes.NewErrorWithCodeRef(err, error_codes.GenericServerError)
		}

		return "ok", nil
	}, common.AccessLevelWrite, "music:category:crud:delete")); err != nil {
		return err
	}

	if err := adminEndpoint.RegisterRpcCommand(router.NewAdminCommand("MoodsUpsertAdmin", func(request []byte, executionData router.MethodExecutionData) (interface{}, *error_codes.ErrorWithCode) {
		var req moods.UpsertRequest
		if err := json.Unmarshal(request, &req); err != nil {
			return nil, error_codes.NewErrorWithCodeRef(err, error_codes.GenericMappingError)
		}

		res, err := moods.Upsert(req, database.GetDb(database.DbTypeMaster).WithContext(executionData.Context))
		if err != nil {
			return nil, error_codes.NewErrorWithCodeRef(err, error_codes.GenericServerError)
		}

		return res, nil
	}, common.AccessLevelWrite, "music:mood:crud:upsert")); err != nil {
		return err
	}

	if err := adminEndpoint.RegisterRpcCommand(router.NewAdminCommand("MoodsListAdmin", func(request []byte, executionData router.MethodExecutionData) (interface{}, *error_codes.ErrorWithCode) {
		var req moods.ListRequest
		if err := json.Unmarshal(request, &req); err != nil {
			return nil, error_codes.NewErrorWithCodeRef(err, error_codes.GenericMappingError)
		}

		res, err := moods.AdminList(req, database.GetDb(database.DbTypeReadonly).WithContext(executionData.Context))
		if err != nil {
			return nil, error_codes.NewErrorWithCodeRef(err, error_codes.GenericServerError)
		}

		return res, nil
	}, common.AccessLevelRead, "music:mood:list")); err != nil {
		return err
	}

	if err := adminEndpoint.RegisterRpcCommand(router.NewAdminCommand("MoodsDeleteAdmin", func(request []byte, executionData router.MethodExecutionData) (interface{}, *error_codes.ErrorWithCode) {
		var req moods.DeleteRequest
		if err := json.Unmarshal(request, &req); err != nil {
			return nil, error_codes.NewErrorWithCodeRef(err, error_codes.GenericMappingError)
		}

		err := moods.Delete(req, database.GetDb(database.DbTypeMaster).WithContext(executionData.Context))
		if err != nil {
			return nil, error_codes.NewErrorWithCodeRef(err, error_codes.GenericServerError)
		}

		return "ok", nil
	}, common.AccessLevelWrite, "music:mood:crud:delete")); err != nil {
		return err
	}

	if err := adminEndpoint.RegisterRpcCommand(router.NewAdminCommand("RejectMusic", func(request []byte, executionData router.MethodExecutionData) (interface{}, *error_codes.ErrorWithCode) {
		var req moderation.RejectMusicRequest
		if err := json.Unmarshal(request, &req); err != nil {
			return nil, error_codes.NewErrorWithCodeRef(err, error_codes.GenericMappingError)
		}

		err := moderation.RejectMusic(req, database.GetDb(database.DbTypeMaster).WithContext(executionData.Context))
		if err != nil {
			return nil, error_codes.NewErrorWithCodeRef(err, error_codes.GenericServerError)
		}

		return "ok", nil
	}, common.AccessLevelWrite, "music:moderation:reject")); err != nil {
		return err
	}

	if err := adminEndpoint.RegisterRpcCommand(router.NewAdminCommand("ApproveMusic", func(request []byte, executionData router.MethodExecutionData) (interface{}, *error_codes.ErrorWithCode) {
		var req moderation.ApproveMusicRequest
		if err := json.Unmarshal(request, &req); err != nil {
			return nil, error_codes.NewErrorWithCodeRef(err, error_codes.GenericMappingError)
		}

		err := moderation.ApproveMusic(req, database.GetDb(database.DbTypeMaster).WithContext(executionData.Context))
		if err != nil {
			return nil, error_codes.NewErrorWithCodeRef(err, error_codes.GenericServerError)
		}

		return "ok", nil
	}, common.AccessLevelWrite, "music:moderation:approve")); err != nil {
		return err
	}

	if err := adminEndpoint.RegisterRpcCommand(router.NewAdminCommand("CreatorSongsList", func(request []byte, executionData router.MethodExecutionData) (interface{}, *error_codes.ErrorWithCode) {
		var req moderation.ListRequest
		if err := json.Unmarshal(request, &req); err != nil {
			return nil, error_codes.NewErrorWithCodeRef(err, error_codes.GenericMappingError)
		}

		return moderation.List(req, database.GetDb(database.DbTypeReadonly).WithContext(executionData.Context), userGoWrapper, executionData.ApmTransaction)
	}, common.AccessLevelRead, "music:moderation:list")); err != nil {
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

	apiDef["CategoriesUpsertAdmin"] = swagger.ApiDescription{
		Request:  categories.UpsertRequest{},
		Response: []database.Category{},
		Tags:     []string{"categories"},
	}

	apiDef["CategoriesListAdmin"] = swagger.ApiDescription{
		Request:  categories.ListRequest{},
		Response: categories.ListResponse{},
		Tags:     []string{"categories"},
	}

	apiDef["MoodsDeleteAdmin"] = swagger.ApiDescription{
		Request:  moods.DeleteRequest{},
		Response: nil,
		Tags:     []string{"moods"},
	}

	apiDef["MoodsUpsertAdmin"] = swagger.ApiDescription{
		Request:  moods.UpsertRequest{},
		Response: []database.Mood{},
		Tags:     []string{"moods"},
	}

	apiDef["MoodsListAdmin"] = swagger.ApiDescription{
		Request:  moods.ListRequest{},
		Response: moods.ListResponse{},
		Tags:     []string{"moods"},
	}

	apiDef["CategoriesDeleteAdmin"] = swagger.ApiDescription{
		Request:  categories.DeleteRequest{},
		Response: nil,
		Tags:     []string{"categories"},
	}

	apiDef["RejectMusic"] = swagger.ApiDescription{
		Request: moderation.RejectMusicRequest{},
		Tags:    []string{"moderation"},
	}

	apiDef["ApproveMusic"] = swagger.ApiDescription{
		Request: moderation.ApproveMusicRequest{},
		Tags:    []string{"moderation"},
	}

	apiDef["CreatorSongsList"] = swagger.ApiDescription{
		Request:  moderation.ListRequest{},
		Response: moderation.ListResponse{},
		Tags:     []string{"moderation"},
	}

	return nil
}
