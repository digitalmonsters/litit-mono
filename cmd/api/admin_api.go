package api

import (
	"encoding/json"
	"github.com/digitalmonsters/go-common/common"
	"github.com/digitalmonsters/go-common/error_codes"
	"github.com/digitalmonsters/go-common/router"
	"github.com/digitalmonsters/go-common/swagger"
	"github.com/digitalmonsters/music/pkg/database"
	"github.com/digitalmonsters/music/pkg/playlist"
	"github.com/digitalmonsters/music/pkg/song"
)

func InitAuthApi(httpRouter *router.HttpRouter, apiDef map[string]swagger.ApiDescription) error {
	if err := httpRouter.RegisterRpcCommand(router.NewCommand("UpsertPlaylistAdmin", func(request []byte, executionData router.MethodExecutionData) (interface{}, *error_codes.ErrorWithCode) {
		var req playlist.UpsertPlaylistRequest

		if err := json.Unmarshal(request, &req); err != nil {
			return nil, error_codes.NewErrorWithCodeRef(err, error_codes.GenericMappingError)
		}

		res, err := playlist.UpsertPlaylist(req, database.GetDb(database.DbTypeMaster).WithContext(executionData.Context))
		if err != nil {
			return nil, error_codes.NewErrorWithCodeRef(err, error_codes.GenericServerError)
		}

		return res, nil
	}, common.AccessLevelRead, false, true)); err != nil {
		return err
	}

	if err := httpRouter.RegisterRpcCommand(router.NewCommand("DeletePlaylistsBulkAdmin", func(request []byte, executionData router.MethodExecutionData) (interface{}, *error_codes.ErrorWithCode) {
		var req playlist.DeletePlaylistsBulkRequest

		if err := json.Unmarshal(request, &req); err != nil {
			return nil, error_codes.NewErrorWithCodeRef(err, error_codes.GenericMappingError)
		}

		err := playlist.DeletePlaylistsBulk(req, database.GetDb(database.DbTypeMaster).WithContext(executionData.Context))
		if err != nil {
			return nil, error_codes.NewErrorWithCodeRef(err, error_codes.GenericServerError)
		}

		return "ok", nil
	}, common.AccessLevelRead, false, true)); err != nil {
		return err
	}

	if err := httpRouter.RegisterRpcCommand(router.NewCommand("PlaylistListingAdmin", func(request []byte, executionData router.MethodExecutionData) (interface{}, *error_codes.ErrorWithCode) {
		var req playlist.PlaylistListingAdminRequest

		if err := json.Unmarshal(request, &req); err != nil {
			return nil, error_codes.NewErrorWithCodeRef(err, error_codes.GenericMappingError)
		}

		resp, err := playlist.PlaylistListingAdmin(req, database.GetDb(database.DbTypeReadonly).WithContext(executionData.Context))
		if err != nil {
			return nil, error_codes.NewErrorWithCodeRef(err, error_codes.GenericServerError)
		}

		return resp, nil
	}, common.AccessLevelRead, false, true)); err != nil {
		return err
	}

	if err := httpRouter.RegisterRpcCommand(router.NewCommand("AddSongToPlaylistBulkAdmin", func(request []byte, executionData router.MethodExecutionData) (interface{}, *error_codes.ErrorWithCode) {
		var req song.AddSongToPlaylistRequest

		if err := json.Unmarshal(request, &req); err != nil {
			return nil, error_codes.NewErrorWithCodeRef(err, error_codes.GenericMappingError)
		}

		err := song.AddSongToPlaylistBulk(req, database.GetDb(database.DbTypeMaster).WithContext(executionData.Context))
		if err != nil {
			return nil, error_codes.NewErrorWithCodeRef(err, error_codes.GenericServerError)
		}

		return "ok", nil
	}, common.AccessLevelRead, false, true)); err != nil {
		return err
	}

	if err := httpRouter.RegisterRpcCommand(router.NewCommand("DeleteSongFromPlaylistsBulkAdmin", func(request []byte, executionData router.MethodExecutionData) (interface{}, *error_codes.ErrorWithCode) {
		var req song.DeleteSongsFromPlaylistBulkRequest

		if err := json.Unmarshal(request, &req); err != nil {
			return nil, error_codes.NewErrorWithCodeRef(err, error_codes.GenericMappingError)
		}

		err := song.DeleteSongFromPlaylistsBulk(req, database.GetDb(database.DbTypeMaster).WithContext(executionData.Context))
		if err != nil {
			return nil, error_codes.NewErrorWithCodeRef(err, error_codes.GenericServerError)
		}

		return "ok", nil
	}, common.AccessLevelRead, false, true)); err != nil {
		return err
	}

	apiDef["UpsertPlaylistAdmin"] = swagger.ApiDescription{
		Request:  playlist.UpsertPlaylistRequest{},
		Response: database.Playlist{},
		Tags:     []string{"upsert", "playlist", "admin"},
	}

	apiDef["DeletePlaylistsBulkAdmin"] = swagger.ApiDescription{
		Request:  playlist.DeletePlaylistsBulkRequest{},
		Response: nil,
		Tags:     []string{"delete", "playlist", "bulk", "admin"},
	}

	apiDef["PlaylistListingAdmin"] = swagger.ApiDescription{
		Request:  playlist.PlaylistListingAdminRequest{},
		Response: playlist.PlaylistListingAdminResponse{},
		Tags:     []string{"list", "playlist", "bulk", "admin"},
	}

	apiDef["AddSongToPlaylistBulkAdmin"] = swagger.ApiDescription{
		Request:  song.AddSongToPlaylistRequest{},
		Response: nil,
		Tags:     []string{"song", "playlist", "add"},
	}

	apiDef["DeleteSongFromPlaylistsBulkAdmin"] = swagger.ApiDescription{
		Request:  song.DeleteSongsFromPlaylistBulkRequest{},
		Response: nil,
		Tags:     []string{"song", "playlist", "delete"},
	}

	return nil
}
