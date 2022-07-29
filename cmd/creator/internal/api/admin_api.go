package api

import (
	"encoding/json"
	"github.com/digitalmonsters/go-common/common"
	"github.com/digitalmonsters/go-common/error_codes"
	"github.com/digitalmonsters/go-common/router"
	"github.com/digitalmonsters/go-common/swagger"
	"github.com/digitalmonsters/music/pkg/creators"
	"github.com/digitalmonsters/music/pkg/creators/moderation"
	"github.com/digitalmonsters/music/pkg/database"
)

func (c *creatorApp) initAdminApi() error {
	adminCommands := []router.ICommand{
		c.creatorRequestListAdmin(),
		c.creatorRequestApproveAdmin(),
		c.creatorRequestRejectAdmin(),
		c.rejectCreatorSong(),
		c.approveCreatorSong(),
		c.creatorSongsList(),
	}

	for _, command := range adminCommands {
		if err := c.httpRouter.GetRpcAdminEndpoint().RegisterRpcCommand(command); err != nil {
			return err
		}
	}
	return nil
}

func (c *creatorApp) creatorRequestListAdmin() router.ICommand {
	method := "CreatorRequestsListAdmin"

	c.apiDef[method] = swagger.ApiDescription{
		Request:  creators.CreatorRequestsListRequest{},
		Response: creators.CreatorRequestsListResponse{},
		Tags:     []string{"creator"},
	}

	return router.NewAdminCommand(method, func(request []byte, executionData router.MethodExecutionData) (interface{}, *error_codes.ErrorWithCode) {
		var req creators.CreatorRequestsListRequest
		if err := json.Unmarshal(request, &req); err != nil {
			return nil, error_codes.NewErrorWithCodeRef(err, error_codes.GenericMappingError)
		}

		res, err := c.creatorsService.CreatorRequestsList(req, database.GetDb(database.DbTypeReadonly).WithContext(executionData.Context), c.creatorsCfg.MaxThresholdHours, executionData.Context, c.userGoWrapper)
		if err != nil {
			return nil, error_codes.NewErrorWithCodeRef(err, error_codes.GenericServerError)
		}

		return res, nil
	}, common.AccessLevelRead, "music:creator:list")
}

func (c *creatorApp) creatorRequestApproveAdmin() router.ICommand {
	method := "CreatorRequestApproveAdmin"

	c.apiDef[method] = swagger.ApiDescription{
		Request:  creators.CreatorRequestApproveRequest{},
		Response: []database.Creator{},
		Tags:     []string{"creator"},
	}

	return router.NewAdminCommand(method, func(request []byte, executionData router.MethodExecutionData) (interface{}, *error_codes.ErrorWithCode) {
		var req creators.CreatorRequestApproveRequest
		if err := json.Unmarshal(request, &req); err != nil {
			return nil, error_codes.NewErrorWithCodeRef(err, error_codes.GenericMappingError)
		}

		res, err := c.creatorsService.CreatorRequestApprove(req, database.GetDb(database.DbTypeMaster).WithContext(executionData.Context))
		if err != nil {
			return nil, error_codes.NewErrorWithCodeRef(err, error_codes.GenericServerError)
		}

		return res, nil
	}, common.AccessLevelWrite, "music:creator:crud:approve")
}

func (c *creatorApp) creatorRequestRejectAdmin() router.ICommand {
	method := "CreatorRequestRejectAdmin"

	c.apiDef[method] = swagger.ApiDescription{
		Request:  creators.CreatorRequestRejectRequest{},
		Response: []database.Creator{},
		Tags:     []string{"creator"},
	}

	return router.NewAdminCommand(method, func(request []byte, executionData router.MethodExecutionData) (interface{}, *error_codes.ErrorWithCode) {
		var req creators.CreatorRequestRejectRequest
		if err := json.Unmarshal(request, &req); err != nil {
			return nil, error_codes.NewErrorWithCodeRef(err, error_codes.GenericMappingError)
		}

		res, err := c.creatorsService.CreatorRequestReject(req, database.GetDb(database.DbTypeMaster).WithContext(executionData.Context))
		if err != nil {
			return nil, error_codes.NewErrorWithCodeRef(err, error_codes.GenericServerError)
		}

		return res, nil
	}, common.AccessLevelWrite, "music:creator:crud:reject")
}

func (c *creatorApp) rejectCreatorSong() router.ICommand {
	method := "RejectMusic"

	c.apiDef[method] = swagger.ApiDescription{
		Request: moderation.RejectMusicRequest{},
		Tags:    []string{"moderation"},
	}
	return router.NewAdminCommand(method, func(request []byte, executionData router.MethodExecutionData) (interface{}, *error_codes.ErrorWithCode) {
		var req moderation.RejectMusicRequest
		if err := json.Unmarshal(request, &req); err != nil {
			return nil, error_codes.NewErrorWithCodeRef(err, error_codes.GenericMappingError)
		}

		err := moderation.RejectMusic(req, database.GetDb(database.DbTypeMaster).WithContext(executionData.Context))
		if err != nil {
			return nil, error_codes.NewErrorWithCodeRef(err, error_codes.GenericServerError)
		}

		return "ok", nil
	}, common.AccessLevelWrite, "music:moderation:reject")
}

func (c *creatorApp) approveCreatorSong() router.ICommand {
	method := "ApproveMusic"

	c.apiDef[method] = swagger.ApiDescription{
		Request: moderation.ApproveMusicRequest{},
		Tags:    []string{"moderation"},
	}

	return router.NewAdminCommand(method, func(request []byte, executionData router.MethodExecutionData) (interface{}, *error_codes.ErrorWithCode) {
		var req moderation.ApproveMusicRequest
		if err := json.Unmarshal(request, &req); err != nil {
			return nil, error_codes.NewErrorWithCodeRef(err, error_codes.GenericMappingError)
		}

		err := moderation.ApproveMusic(req, database.GetDb(database.DbTypeMaster).WithContext(executionData.Context))
		if err != nil {
			return nil, error_codes.NewErrorWithCodeRef(err, error_codes.GenericServerError)
		}

		return "ok", nil
	}, common.AccessLevelWrite, "music:moderation:approve")
}

func (c *creatorApp) creatorSongsList() router.ICommand {
	method := "CreatorSongsList"

	c.apiDef[method] = swagger.ApiDescription{
		Request:  moderation.ListRequest{},
		Response: moderation.ListResponse{},
		Tags:     []string{"moderation"},
	}

	return router.NewAdminCommand(method, func(request []byte, executionData router.MethodExecutionData) (interface{}, *error_codes.ErrorWithCode) {
		var req moderation.ListRequest
		if err := json.Unmarshal(request, &req); err != nil {
			return nil, error_codes.NewErrorWithCodeRef(err, error_codes.GenericMappingError)
		}

		return moderation.List(req, database.GetDb(database.DbTypeReadonly).WithContext(executionData.Context), c.userGoWrapper, executionData.Context)
	}, common.AccessLevelRead, "music:moderation:list")
}
