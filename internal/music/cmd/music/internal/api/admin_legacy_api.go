package api

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

func (m *musicApp) initAdminLegacyApi() error {
	legacyCommands := []router.ICommand{
		m.playlistsListAdminLegacy(),
		m.upsertPlaylistsLegacy(),
		m.deletePlaylistsLegacy(),
		m.addSongToPlaylistAdminLegacy(),
		m.deleteSongFromPlaylistAdminLegacy(),
		m.playlistSongListAdminLegacy(),
		m.allSongsListAdminLegacy(),
		m.ownStorageMusicListLegacy(),
		m.upsertSongsToOwnStorageLegacy(),
		m.deleteSongsFromOwnStorageLegacy(),
	}

	for _, command := range legacyCommands {
		if err := m.httpRouter.GetRpcAdminLegacyEndpoint().RegisterRpcCommand(command); err != nil {
			return err
		}
	}
	return nil
}

func (m *musicApp) upsertPlaylistsLegacy() router.ICommand {
	method := "UpsertPlaylistAdmin"

	m.apiDef[method] = swagger.ApiDescription{
		Request:  playlist.UpsertPlaylistRequest{},
		Response: database.Playlist{},
		Tags:     []string{"upsert", "playlist", "admin"},
	}

	return router.NewLegacyAdminCommand(method, func(request []byte, executionData router.MethodExecutionData) (interface{}, *error_codes.ErrorWithCode) {
		var req playlist.UpsertPlaylistRequest

		if err := json.Unmarshal(request, &req); err != nil {
			return nil, error_codes.NewErrorWithCodeRef(err, error_codes.GenericMappingError)
		}

		res, err := playlist.UpsertPlaylist(req, database.GetDb(database.DbTypeMaster).WithContext(executionData.Context))
		if err != nil {
			return nil, error_codes.NewErrorWithCodeRef(err, error_codes.GenericServerError)
		}

		return res, nil
	})
}

func (m *musicApp) deletePlaylistsLegacy() router.ICommand {
	method := "DeletePlaylistsBulkAdmin"

	m.apiDef[method] = swagger.ApiDescription{
		Request:  playlist.DeletePlaylistsBulkRequest{},
		Response: nil,
		Tags:     []string{"delete", "playlist", "bulk", "admin"},
	}

	return router.NewLegacyAdminCommand(method, func(request []byte, executionData router.MethodExecutionData) (interface{}, *error_codes.ErrorWithCode) {
		var req playlist.DeletePlaylistsBulkRequest

		if err := json.Unmarshal(request, &req); err != nil {
			return nil, error_codes.NewErrorWithCodeRef(err, error_codes.GenericMappingError)
		}

		err := playlist.DeletePlaylistsBulk(req, database.GetDb(database.DbTypeMaster).WithContext(executionData.Context))
		if err != nil {
			return nil, error_codes.NewErrorWithCodeRef(err, error_codes.GenericServerError)
		}

		return "ok", nil
	})
}

func (m *musicApp) playlistsListAdminLegacy() router.ICommand {
	method := "PlaylistListingAdmin"

	m.apiDef[method] = swagger.ApiDescription{
		Request:  playlist.PlaylistListingAdminRequest{},
		Response: playlist.PlaylistListingAdminResponse{},
		Tags:     []string{"list", "playlist", "bulk", "admin"},
	}

	return router.NewLegacyAdminCommand(method, func(request []byte, executionData router.MethodExecutionData) (interface{}, *error_codes.ErrorWithCode) {
		var req playlist.PlaylistListingAdminRequest

		if err := json.Unmarshal(request, &req); err != nil {
			return nil, error_codes.NewErrorWithCodeRef(err, error_codes.GenericMappingError)
		}

		resp, err := playlist.PlaylistListingAdmin(req, database.GetDb(database.DbTypeReadonly).WithContext(executionData.Context))
		if err != nil {
			return nil, error_codes.NewErrorWithCodeRef(err, error_codes.GenericServerError)
		}

		return resp, nil
	})
}

func (m *musicApp) addSongToPlaylistAdminLegacy() router.ICommand {
	method := "AddSongToPlaylistBulkAdmin"

	m.apiDef[method] = swagger.ApiDescription{
		Request:  song.AddSongToPlaylistRequest{},
		Response: nil,
		Tags:     []string{"song", "playlist", "add"},
	}

	return router.NewLegacyAdminCommand(method, func(request []byte, executionData router.MethodExecutionData) (interface{}, *error_codes.ErrorWithCode) {
		var req song.AddSongToPlaylistRequest

		if err := json.Unmarshal(request, &req); err != nil {
			return nil, error_codes.NewErrorWithCodeRef(err, error_codes.GenericMappingError)
		}

		err := song.AddSongToPlaylistBulk(req, database.GetDb(database.DbTypeMaster).WithContext(executionData.Context), executionData.ApmTransaction, m.musicStorageService, executionData.Context)
		if err != nil {
			return nil, error_codes.NewErrorWithCodeRef(err, error_codes.GenericServerError)
		}

		return "ok", nil
	})
}

func (m *musicApp) deleteSongFromPlaylistAdminLegacy() router.ICommand {
	method := "DeleteSongFromPlaylistsBulkAdmin"

	m.apiDef[method] = swagger.ApiDescription{
		Request:  song.DeleteSongsFromPlaylistBulkRequest{},
		Response: nil,
		Tags:     []string{"song", "playlist", "delete"},
	}

	return router.NewLegacyAdminCommand(method, func(request []byte, executionData router.MethodExecutionData) (interface{}, *error_codes.ErrorWithCode) {
		var req song.DeleteSongsFromPlaylistBulkRequest

		if err := json.Unmarshal(request, &req); err != nil {
			return nil, error_codes.NewErrorWithCodeRef(err, error_codes.GenericMappingError)
		}

		err := song.DeleteSongFromPlaylistsBulk(req, database.GetDb(database.DbTypeMaster).WithContext(executionData.Context))
		if err != nil {
			return nil, error_codes.NewErrorWithCodeRef(err, error_codes.GenericServerError)
		}

		return "ok", nil
	})
}

func (m *musicApp) playlistSongListAdminLegacy() router.ICommand {
	method := "PlaylistSongListAdmin"

	m.apiDef[method] = swagger.ApiDescription{
		Request:  song.PlaylistSongListRequest{},
		Response: song.PlaylistSongListResponse{},
		Tags:     []string{"song", "playlist", "list"},
	}

	return router.NewLegacyAdminCommand(method, func(request []byte, executionData router.MethodExecutionData) (interface{}, *error_codes.ErrorWithCode) {
		var req song.PlaylistSongListRequest

		if err := json.Unmarshal(request, &req); err != nil {
			return nil, error_codes.NewErrorWithCodeRef(err, error_codes.GenericMappingError)
		}

		resp, err := song.PlaylistSongListAdmin(req, database.GetDb(database.DbTypeReadonly).WithContext(executionData.Context))
		if err != nil {
			return nil, error_codes.NewErrorWithCodeRef(err, error_codes.GenericServerError)
		}

		return resp, nil
	})
}

func (m *musicApp) allSongsListAdminLegacy() router.ICommand {
	method := "AllSongsListAdmin"

	m.apiDef[method] = swagger.ApiDescription{
		Request:  music_source.ListMusicRequest{},
		Response: music_source.ListMusicResponse{},
		Tags:     []string{"songs", "list", "admin"},
	}

	return router.NewLegacyAdminCommand(method, func(request []byte, executionData router.MethodExecutionData) (interface{}, *error_codes.ErrorWithCode) {
		var req music_source.ListMusicRequest

		if err := json.Unmarshal(request, &req); err != nil {
			return nil, error_codes.NewErrorWithCodeRef(err, error_codes.GenericMappingError)
		}

		resp, err := m.musicStorageService.ListMusic(req, database.GetDb(database.DbTypeReadonly).WithContext(executionData.Context), executionData.ApmTransaction, executionData.Context)
		if err != nil {
			return nil, error_codes.NewErrorWithCodeRef(err, error_codes.GenericServerError)
		}

		return resp, nil
	})
}

func (m *musicApp) upsertSongsToOwnStorageLegacy() router.ICommand {
	method := "UpsertSongsToOwnStorageBulk"

	m.apiDef[method] = swagger.ApiDescription{
		Request:  own_storage.AddSongsToOwnStorageRequest{},
		Response: []database.MusicStorage{},
		Tags:     []string{"upsert", "song", "own_storage", "admin"},
	}

	return router.NewLegacyAdminCommand(method, func(request []byte, executionData router.MethodExecutionData) (interface{}, *error_codes.ErrorWithCode) {
		var req own_storage.AddSongsToOwnStorageRequest

		if err := json.Unmarshal(request, &req); err != nil {
			return nil, error_codes.NewErrorWithCodeRef(err, error_codes.GenericMappingError)
		}

		resp, err := own_storage.UpsertSongsToOwnStorageBulk(req, database.GetDb(database.DbTypeMaster).WithContext(executionData.Context))
		if err != nil {
			return nil, error_codes.NewErrorWithCodeRef(err, error_codes.GenericServerError)
		}

		return resp, nil
	})
}

func (m *musicApp) deleteSongsFromOwnStorageLegacy() router.ICommand {
	method := "DeleteSongsFromOwnStorageBulk"

	m.apiDef[method] = swagger.ApiDescription{
		Request:  own_storage.DeleteSongsFromOwnStorageRequest{},
		Response: api.SuccessResponse{},
		Tags:     []string{"delete", "song", "own_storage", "admin"},
	}

	return router.NewLegacyAdminCommand(method, func(request []byte, executionData router.MethodExecutionData) (interface{}, *error_codes.ErrorWithCode) {
		var req own_storage.DeleteSongsFromOwnStorageRequest

		if err := json.Unmarshal(request, &req); err != nil {
			return nil, error_codes.NewErrorWithCodeRef(err, error_codes.GenericMappingError)
		}

		if err := own_storage.DeleteSongsFromOwnStorageBulk(req, database.GetDb(database.DbTypeMaster).WithContext(executionData.Context)); err != nil {
			return nil, error_codes.NewErrorWithCodeRef(err, error_codes.GenericServerError)
		}

		return api.SuccessResponse{Success: true}, nil
	})
}

func (m *musicApp) ownStorageMusicListLegacy() router.ICommand {
	method := "OwnStorageMusicList"

	m.apiDef[method] = swagger.ApiDescription{
		Request:  own_storage.OwnStorageMusicListRequest{},
		Response: own_storage.OwnStorageMusicListResponse{},
		Tags:     []string{"list", "song", "own_storage", "admin"},
	}

	return router.NewLegacyAdminCommand(method, func(request []byte, executionData router.MethodExecutionData) (interface{}, *error_codes.ErrorWithCode) {
		var req own_storage.OwnStorageMusicListRequest

		if err := json.Unmarshal(request, &req); err != nil {
			return nil, error_codes.NewErrorWithCodeRef(err, error_codes.GenericMappingError)
		}

		resp, err := own_storage.OwnStorageMusicList(req, database.GetDb(database.DbTypeReadonly).WithContext(executionData.Context))
		if err != nil {
			return nil, error_codes.NewErrorWithCodeRef(err, error_codes.GenericServerError)
		}

		return resp, nil
	})
}
