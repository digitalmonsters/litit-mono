package music

import (
	"encoding/json"
	"github.com/digitalmonsters/go-common/error_codes"
	"github.com/digitalmonsters/go-common/router"
	"github.com/digitalmonsters/go-common/swagger"
	"github.com/digitalmonsters/music/cmd/api"
	"github.com/digitalmonsters/music/pkg/database"
	"github.com/digitalmonsters/music/pkg/music_source"
	"github.com/digitalmonsters/music/pkg/own_storage"
	"github.com/digitalmonsters/music/pkg/playlist"
	"github.com/digitalmonsters/music/pkg/song"
)

func InitAdminApi(adminEndpoint router.IRpcEndpoint, apiDef map[string]swagger.ApiDescription, musicStorageService *music_source.MusicStorageService) error {
	if err := adminEndpoint.RegisterRpcCommand(router.NewLegacyAdminCommand("UpsertPlaylistAdmin", func(request []byte, executionData router.MethodExecutionData) (interface{}, *error_codes.ErrorWithCode) {
		var req playlist.UpsertPlaylistRequest

		if err := json.Unmarshal(request, &req); err != nil {
			return nil, error_codes.NewErrorWithCodeRef(err, error_codes.GenericMappingError)
		}

		res, err := playlist.UpsertPlaylist(req, database.GetDb(database.DbTypeMaster).WithContext(executionData.Context))
		if err != nil {
			return nil, error_codes.NewErrorWithCodeRef(err, error_codes.GenericServerError)
		}

		return res, nil
	})); err != nil {
		return err
	}

	if err := adminEndpoint.RegisterRpcCommand(router.NewLegacyAdminCommand("DeletePlaylistsBulkAdmin", func(request []byte, executionData router.MethodExecutionData) (interface{}, *error_codes.ErrorWithCode) {
		var req playlist.DeletePlaylistsBulkRequest

		if err := json.Unmarshal(request, &req); err != nil {
			return nil, error_codes.NewErrorWithCodeRef(err, error_codes.GenericMappingError)
		}

		err := playlist.DeletePlaylistsBulk(req, database.GetDb(database.DbTypeMaster).WithContext(executionData.Context))
		if err != nil {
			return nil, error_codes.NewErrorWithCodeRef(err, error_codes.GenericServerError)
		}

		return "ok", nil
	})); err != nil {
		return err
	}

	if err := adminEndpoint.RegisterRpcCommand(router.NewLegacyAdminCommand("PlaylistListingAdmin", func(request []byte, executionData router.MethodExecutionData) (interface{}, *error_codes.ErrorWithCode) {
		var req playlist.PlaylistListingAdminRequest

		if err := json.Unmarshal(request, &req); err != nil {
			return nil, error_codes.NewErrorWithCodeRef(err, error_codes.GenericMappingError)
		}

		resp, err := playlist.PlaylistListingAdmin(req, database.GetDb(database.DbTypeReadonly).WithContext(executionData.Context))
		if err != nil {
			return nil, error_codes.NewErrorWithCodeRef(err, error_codes.GenericServerError)
		}

		return resp, nil
	})); err != nil {
		return err
	}

	if err := adminEndpoint.RegisterRpcCommand(router.NewLegacyAdminCommand("AddSongToPlaylistBulkAdmin", func(request []byte, executionData router.MethodExecutionData) (interface{}, *error_codes.ErrorWithCode) {
		var req song.AddSongToPlaylistRequest

		if err := json.Unmarshal(request, &req); err != nil {
			return nil, error_codes.NewErrorWithCodeRef(err, error_codes.GenericMappingError)
		}

		err := song.AddSongToPlaylistBulk(req, database.GetDb(database.DbTypeMaster).WithContext(executionData.Context), executionData.ApmTransaction, musicStorageService)
		if err != nil {
			return nil, error_codes.NewErrorWithCodeRef(err, error_codes.GenericServerError)
		}

		return "ok", nil
	})); err != nil {
		return err
	}

	if err := adminEndpoint.RegisterRpcCommand(router.NewLegacyAdminCommand("DeleteSongFromPlaylistsBulkAdmin", func(request []byte, executionData router.MethodExecutionData) (interface{}, *error_codes.ErrorWithCode) {
		var req song.DeleteSongsFromPlaylistBulkRequest

		if err := json.Unmarshal(request, &req); err != nil {
			return nil, error_codes.NewErrorWithCodeRef(err, error_codes.GenericMappingError)
		}

		err := song.DeleteSongFromPlaylistsBulk(req, database.GetDb(database.DbTypeMaster).WithContext(executionData.Context))
		if err != nil {
			return nil, error_codes.NewErrorWithCodeRef(err, error_codes.GenericServerError)
		}

		return "ok", nil
	})); err != nil {
		return err
	}

	if err := adminEndpoint.RegisterRpcCommand(router.NewLegacyAdminCommand("PlaylistSongListAdmin", func(request []byte, executionData router.MethodExecutionData) (interface{}, *error_codes.ErrorWithCode) {
		var req song.PlaylistSongListRequest

		if err := json.Unmarshal(request, &req); err != nil {
			return nil, error_codes.NewErrorWithCodeRef(err, error_codes.GenericMappingError)
		}

		resp, err := song.PlaylistSongListAdmin(req, database.GetDb(database.DbTypeReadonly).WithContext(executionData.Context))
		if err != nil {
			return nil, error_codes.NewErrorWithCodeRef(err, error_codes.GenericServerError)
		}

		return resp, nil
	})); err != nil {
		return err
	}

	if err := adminEndpoint.RegisterRpcCommand(router.NewLegacyAdminCommand("AllSongsListAdmin", func(request []byte, executionData router.MethodExecutionData) (interface{}, *error_codes.ErrorWithCode) {
		var req music_source.ListMusicRequest

		if err := json.Unmarshal(request, &req); err != nil {
			return nil, error_codes.NewErrorWithCodeRef(err, error_codes.GenericMappingError)
		}

		resp, err := musicStorageService.ListMusic(req, database.GetDb(database.DbTypeReadonly).WithContext(executionData.Context), executionData.ApmTransaction)
		if err != nil {
			return nil, error_codes.NewErrorWithCodeRef(err, error_codes.GenericServerError)
		}

		return resp, nil
	})); err != nil {
		return err
	}

	if err := adminEndpoint.RegisterRpcCommand(router.NewLegacyAdminCommand("UpsertSongsToOwnStorageBulk", func(request []byte, executionData router.MethodExecutionData) (interface{}, *error_codes.ErrorWithCode) {
		var req own_storage.AddSongsToOwnStorageRequest

		if err := json.Unmarshal(request, &req); err != nil {
			return nil, error_codes.NewErrorWithCodeRef(err, error_codes.GenericMappingError)
		}

		resp, err := own_storage.UpsertSongsToOwnStorageBulk(req, database.GetDb(database.DbTypeMaster).WithContext(executionData.Context))
		if err != nil {
			return nil, error_codes.NewErrorWithCodeRef(err, error_codes.GenericServerError)
		}

		return resp, nil
	})); err != nil {
		return err
	}

	if err := adminEndpoint.RegisterRpcCommand(router.NewLegacyAdminCommand("DeleteSongsFromOwnStorageBulk", func(request []byte, executionData router.MethodExecutionData) (interface{}, *error_codes.ErrorWithCode) {
		var req own_storage.DeleteSongsFromOwnStorageRequest

		if err := json.Unmarshal(request, &req); err != nil {
			return nil, error_codes.NewErrorWithCodeRef(err, error_codes.GenericMappingError)
		}

		if err := own_storage.DeleteSongsFromOwnStorageBulk(req, database.GetDb(database.DbTypeMaster).WithContext(executionData.Context)); err != nil {
			return nil, error_codes.NewErrorWithCodeRef(err, error_codes.GenericServerError)
		}

		return api.SuccessResponse{Success: true}, nil
	})); err != nil {
		return err
	}

	if err := adminEndpoint.RegisterRpcCommand(router.NewLegacyAdminCommand("OwnStorageMusicList", func(request []byte, executionData router.MethodExecutionData) (interface{}, *error_codes.ErrorWithCode) {
		var req own_storage.OwnStorageMusicListRequest

		if err := json.Unmarshal(request, &req); err != nil {
			return nil, error_codes.NewErrorWithCodeRef(err, error_codes.GenericMappingError)
		}

		resp, err := own_storage.OwnStorageMusicList(req, database.GetDb(database.DbTypeReadonly).WithContext(executionData.Context))
		if err != nil {
			return nil, error_codes.NewErrorWithCodeRef(err, error_codes.GenericServerError)
		}

		return resp, nil
	})); err != nil {
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

	apiDef["PlaylistSongListAdmin"] = swagger.ApiDescription{
		Request:  song.PlaylistSongListRequest{},
		Response: song.PlaylistSongListResponse{},
		Tags:     []string{"song", "playlist", "list"},
	}

	apiDef["AllSongsListAdmin"] = swagger.ApiDescription{
		Request:  music_source.ListMusicRequest{},
		Response: music_source.ListMusicResponse{},
		Tags:     []string{"songs", "list", "admin"},
	}

	apiDef["UpsertSongsToOwnStorageBulk"] = swagger.ApiDescription{
		Request:  own_storage.AddSongsToOwnStorageRequest{},
		Response: []database.MusicStorage{},
		Tags:     []string{"upsert", "song", "own_storage", "admin"},
	}

	apiDef["DeleteSongsFromOwnStorageBulk"] = swagger.ApiDescription{
		Request:  own_storage.DeleteSongsFromOwnStorageRequest{},
		Response: api.SuccessResponse{},
		Tags:     []string{"delete", "song", "own_storage", "admin"},
	}

	apiDef["OwnStorageMusicList"] = swagger.ApiDescription{
		Request:  own_storage.OwnStorageMusicListRequest{},
		Response: own_storage.OwnStorageMusicListResponse{},
		Tags:     []string{"list", "song", "own_storage", "admin"},
	}

	return nil
}
