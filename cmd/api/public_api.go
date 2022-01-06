package api

import (
	"encoding/json"
	"github.com/digitalmonsters/go-common/common"
	"github.com/digitalmonsters/go-common/error_codes"
	"github.com/digitalmonsters/go-common/router"
	"github.com/digitalmonsters/go-common/swagger"
	"github.com/digitalmonsters/music/pkg/database"
	"github.com/digitalmonsters/music/pkg/favorites"
)

func InitPublicApi(httpRouter *router.HttpRouter, apiDef map[string]swagger.ApiDescription) error {
	if err := httpRouter.RegisterRpcCommand(router.NewCommand("AddToFavorites", func(request []byte, executionData router.MethodExecutionData) (interface{}, *error_codes.ErrorWithCode) {
		var req favorites.AddToFavoritesRequest

		if err := json.Unmarshal(request, &req); err != nil {
			return nil, error_codes.NewErrorWithCodeRef(err, error_codes.GenericMappingError)
		}

		if err := favorites.AddToFavorites(req, database.GetDb(database.DbTypeMaster).WithContext(executionData.Context)); err != nil {
			return nil, error_codes.NewErrorWithCodeRef(err, error_codes.GenericServerError)
		}

		return "ok", nil
	}, common.AccessLevelRead, false, true)); err != nil {
		return err
	}

	if err := httpRouter.RegisterRpcCommand(router.NewCommand("RemoveFromFavorites", func(request []byte, executionData router.MethodExecutionData) (interface{}, *error_codes.ErrorWithCode) {
		var req favorites.RemoveFromFavoritesRequest

		if err := json.Unmarshal(request, &req); err != nil {
			return nil, error_codes.NewErrorWithCodeRef(err, error_codes.GenericMappingError)
		}

		if err := favorites.RemoveFromFavorites(req, database.GetDb(database.DbTypeMaster).WithContext(executionData.Context)); err != nil {
			return nil, error_codes.NewErrorWithCodeRef(err, error_codes.GenericServerError)
		}

		return "ok", nil
	}, common.AccessLevelRead, false, true)); err != nil {
		return err
	}

	apiDef["AddToFavorites"] = swagger.ApiDescription{
		Request:  favorites.AddToFavoritesRequest{},
		Response: nil,
		Tags:     []string{"favorites", "user"},
	}

	apiDef["RemoveFromFavorites"] = swagger.ApiDescription{
		Request:  favorites.RemoveFromFavoritesRequest{},
		Response: nil,
		Tags:     []string{"favorites", "user"},
	}

	return nil
}
