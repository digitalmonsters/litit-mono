package api

import (
	"github.com/digitalmonsters/go-common/common"
	"github.com/digitalmonsters/go-common/error_codes"
	"github.com/digitalmonsters/go-common/router"
	"github.com/digitalmonsters/go-common/swagger"
	"github.com/digitalmonsters/music/pkg/database"
	"github.com/digitalmonsters/music/pkg/favorites"
	"github.com/digitalmonsters/music/pkg/playlist"
	"github.com/digitalmonsters/music/pkg/popular"
	"github.com/digitalmonsters/music/utils"
	"github.com/pkg/errors"
	"gopkg.in/guregu/null.v4"
	"net/http"
)

func InitPublicApi(httpRouter *router.HttpRouter, apiDef map[string]swagger.ApiDescription) error {
	if err := httpRouter.RegisterRestCmd(router.NewRestCommand(func(request []byte,
		executionData router.MethodExecutionData) (interface{}, *error_codes.ErrorWithCode) {
		userId := executionData.UserId

		if userId <= 0 {
			return nil, error_codes.NewErrorWithCodeRef(errors.New("invalid user_id"), error_codes.GenericValidationError)
		}

		songId := utils.ExtractString(executionData.GetUserValue, "song_id", "")

		if len(songId) <= 0 {
			return nil, error_codes.NewErrorWithCodeRef(errors.New("invalid song_id"), error_codes.GenericValidationError)
		}

		req := favorites.AddToFavoritesRequest{
			UserId: userId,
			SongId: songId,
		}

		if err := favorites.AddToFavorites(req, database.GetDb(database.DbTypeMaster).WithContext(executionData.Context)); err != nil {
			return nil, error_codes.NewErrorWithCodeRef(err, error_codes.GenericServerError)
		} else {
			return successResponse{
				Success: true,
			}, nil
		}
	}, "/song/favorites/add/{song_id}", http.MethodPost, common.AccessLevelPublic, true, false)); err != nil {
		return err
	}

	if err := httpRouter.RegisterRestCmd(router.NewRestCommand(func(request []byte,
		executionData router.MethodExecutionData) (interface{}, *error_codes.ErrorWithCode) {
		userId := executionData.UserId

		if userId <= 0 {
			return nil, error_codes.NewErrorWithCodeRef(errors.New("invalid user_id"), error_codes.GenericValidationError)
		}

		songId := utils.ExtractString(executionData.GetUserValue, "song_id", "")

		if len(songId) <= 0 {
			return nil, error_codes.NewErrorWithCodeRef(errors.New("invalid song_id"), error_codes.GenericValidationError)
		}

		req := favorites.RemoveFromFavoritesRequest{
			UserId: userId,
			SongId: songId,
		}

		if err := favorites.RemoveFromFavorites(req, database.GetDb(database.DbTypeMaster).WithContext(executionData.Context)); err != nil {
			return nil, error_codes.NewErrorWithCodeRef(err, error_codes.GenericServerError)
		} else {
			return successResponse{
				Success: true,
			}, nil
		}
	}, "/song/favorites/remove/{song_id}", http.MethodPost, common.AccessLevelPublic, true, false)); err != nil {
		return err
	}

	if err := httpRouter.RegisterRestCmd(router.NewRestCommand(func(request []byte,
		executionData router.MethodExecutionData) (interface{}, *error_codes.ErrorWithCode) {
		userId := executionData.UserId

		if userId <= 0 {
			return nil, error_codes.NewErrorWithCodeRef(errors.New("invalid user_id"), error_codes.GenericValidationError)
		}

		playlistName := utils.ExtractString(executionData.GetUserValue, "name", "")
		count := utils.ExtractInt64(executionData.GetUserValue, "count", 10, 50)
		cursor := utils.ExtractString(executionData.GetUserValue, "cursor", "")

		resp, err := playlist.PlaylistListingPublic(playlist.PlayListListingPublicRequest{
			Name:   null.StringFrom(playlistName),
			Count:  int(count),
			Cursor: cursor,
		}, database.GetDb(database.DbTypeReadonly).WithContext(executionData.Context))

		if err != nil {
			return nil, error_codes.NewErrorWithCodeRef(err, error_codes.GenericServerError)
		}

		return resp, nil
	}, "/playlist", http.MethodGet, common.AccessLevelPublic, true, false)); err != nil {
		return err
	}

	if err := httpRouter.RegisterRestCmd(router.NewRestCommand(func(request []byte,
		executionData router.MethodExecutionData) (interface{}, *error_codes.ErrorWithCode) {
		userId := executionData.UserId

		if userId <= 0 {
			return nil, error_codes.NewErrorWithCodeRef(errors.New("invalid user_id"), error_codes.GenericValidationError)
		}

		playlistId := utils.ExtractInt64(executionData.GetUserValue, "playlist_id", 0, 0)
		if playlistId <= 0 {
			return nil, error_codes.NewErrorWithCodeRef(errors.New("invalid playlist_id"), error_codes.GenericValidationError)
		}

		count := utils.ExtractInt64(executionData.GetUserValue, "count", 10, 50)
		cursor := utils.ExtractString(executionData.GetUserValue, "cursor", "")

		resp, err := playlist.PlaylistSongsListPublic(playlist.PlaylistSongsListPublicRequest{
			PlaylistId: playlistId,
			Count:      int(count),
			Cursor:     cursor,
		}, database.GetDb(database.DbTypeReadonly).WithContext(executionData.Context))

		if err != nil {
			return nil, error_codes.NewErrorWithCodeRef(err, error_codes.GenericServerError)
		}

		return resp, nil
	}, "/playlist/{playlist_id}", http.MethodGet, common.AccessLevelPublic, true, false)); err != nil {
		return err
	}

	if err := httpRouter.RegisterRestCmd(router.NewRestCommand(func(request []byte,
		executionData router.MethodExecutionData) (interface{}, *error_codes.ErrorWithCode) {
		userId := executionData.UserId

		if userId <= 0 {
			return nil, error_codes.NewErrorWithCodeRef(errors.New("invalid user_id"), error_codes.GenericValidationError)
		}

		count := utils.ExtractInt64(executionData.GetUserValue, "count", 10, 50)
		cursor := utils.ExtractString(executionData.GetUserValue, "cursor", "")

		resp, err := favorites.FavoriteSongsList(favorites.FavoriteSongsListRequest{
			Count:  int(count),
			Cursor: cursor,
		}, executionData.UserId, database.GetDb(database.DbTypeReadonly).WithContext(executionData.Context))

		if err != nil {
			return nil, error_codes.NewErrorWithCodeRef(err, error_codes.GenericServerError)
		}

		return resp, nil
	}, "/song/favorites", http.MethodGet, common.AccessLevelPublic, true, false)); err != nil {
		return err
	}

	if err := httpRouter.RegisterRestCmd(router.NewRestCommand(func(request []byte,
		executionData router.MethodExecutionData) (interface{}, *error_codes.ErrorWithCode) {
		userId := executionData.UserId

		if userId <= 0 {
			return nil, error_codes.NewErrorWithCodeRef(errors.New("invalid user_id"), error_codes.GenericValidationError)
		}

		count := utils.ExtractInt64(executionData.GetUserValue, "count", 10, 50)
		cursor := utils.ExtractString(executionData.GetUserValue, "cursor", "")

		resp, err := popular.GetPopularSongs(popular.GetPopularSongsRequest{
			Count:  int(count),
			Cursor: cursor,
		}, database.GetDb(database.DbTypeReadonly).WithContext(executionData.Context))

		if err != nil {
			return nil, error_codes.NewErrorWithCodeRef(err, error_codes.GenericServerError)
		}

		return resp, nil
	}, "/song/popular", http.MethodGet, common.AccessLevelPublic, true, false)); err != nil {
		return err
	}

	apiDef["/song/favorites/add/{song_id}"] = swagger.ApiDescription{
		AdditionalSwaggerParameters: []swagger.ParameterDescription{
			{
				Name:        "song_id",
				In:          swagger.ParameterInPath,
				Description: "song_id",
				Required:    true,
				Type:        "string",
			},
		},
		Response:          successResponse{},
		MethodDescription: "add song ro favorites",
		Tags:              []string{"favorites", "public"},
	}

	apiDef["/song/favorites/remove/{song_id}"] = swagger.ApiDescription{
		AdditionalSwaggerParameters: []swagger.ParameterDescription{
			{
				Name:        "song_id",
				In:          swagger.ParameterInPath,
				Description: "song_id",
				Required:    true,
				Type:        "string",
			},
		},
		Response:          successResponse{},
		MethodDescription: "remove song from favorites",
		Tags:              []string{"favorites", "public"},
	}

	apiDef["/playlist"] = swagger.ApiDescription{
		AdditionalSwaggerParameters: []swagger.ParameterDescription{
			{
				Name:        "name",
				In:          swagger.ParameterInQuery,
				Description: "name filter for playlists",
				Required:    false,
				Type:        "string",
			},
			{
				Name:        "count",
				In:          swagger.ParameterInQuery,
				Description: "count per page",
				Required:    false,
				Type:        "integer",
			},
			{
				Name:        "cursor",
				In:          swagger.ParameterInQuery,
				Description: "cursor position",
				Required:    false,
				Type:        "string",
			},
		},
		Response:          playlist.PlayListListingPublicResponse{},
		MethodDescription: "playlists list",
		Tags:              []string{"playlist", "list", "public"},
	}

	apiDef["/playlist/{playlist_id}"] = swagger.ApiDescription{
		AdditionalSwaggerParameters: []swagger.ParameterDescription{
			{
				Name:        "playlist_id",
				In:          swagger.ParameterInPath,
				Description: "selected playlist id",
				Required:    true,
				Type:        "integer",
			},
			{
				Name:        "count",
				In:          swagger.ParameterInQuery,
				Description: "count per page",
				Required:    false,
				Type:        "integer",
			},
			{
				Name:        "cursor",
				In:          swagger.ParameterInQuery,
				Description: "cursor position",
				Required:    false,
				Type:        "string",
			},
		},
		Response:          playlist.PlaylistSongsListPublicResponse{},
		MethodDescription: "playlist songs list",
		Tags:              []string{"songs", "list", "public"},
	}

	apiDef["/song/favorites"] = swagger.ApiDescription{
		AdditionalSwaggerParameters: []swagger.ParameterDescription{
			{
				Name:        "count",
				In:          swagger.ParameterInQuery,
				Description: "count per page",
				Required:    false,
				Type:        "integer",
			},
			{
				Name:        "cursor",
				In:          swagger.ParameterInQuery,
				Description: "cursor position",
				Required:    false,
				Type:        "string",
			},
		},
		Response:          favorites.FavoriteSongsListResponse{},
		MethodDescription: "favorite songs list",
		Tags:              []string{"favorite", "song", "list", "public"},
	}

	apiDef["/song/popular"] = swagger.ApiDescription{
		AdditionalSwaggerParameters: []swagger.ParameterDescription{
			{
				Name:        "count",
				In:          swagger.ParameterInQuery,
				Description: "count per page",
				Required:    false,
				Type:        "integer",
			},
			{
				Name:        "cursor",
				In:          swagger.ParameterInQuery,
				Description: "cursor position",
				Required:    false,
				Type:        "string",
			},
		},
		Response:          popular.GetPopularSongsResponse{},
		MethodDescription: "popular songs list",
		Tags:              []string{"popular", "song", "list", "public"},
	}

	return nil
}
