package api

import (
	"github.com/digitalmonsters/go-common/common"
	"github.com/digitalmonsters/go-common/error_codes"
	"github.com/digitalmonsters/go-common/router"
	"github.com/digitalmonsters/go-common/swagger"
	"github.com/digitalmonsters/music/pkg/database"
	"github.com/digitalmonsters/music/pkg/favorites"
	"github.com/digitalmonsters/music/utils"
	"github.com/pkg/errors"
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
	}, "/favorites/add/{song_id}", http.MethodPost, common.AccessLevelPublic, true, false)); err != nil {
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
	}, "/favorites/remove/{song_id}", http.MethodPost, common.AccessLevelPublic, true, false)); err != nil {
		return err
	}

	apiDef["/favorites/add/{song_id}"] = swagger.ApiDescription{
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

	apiDef["/favorites/remove/{song_id}"] = swagger.ApiDescription{
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
	return nil
}
