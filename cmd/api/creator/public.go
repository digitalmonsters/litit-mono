package creator

import (
	"encoding/json"
	"github.com/digitalmonsters/go-common/error_codes"
	"github.com/digitalmonsters/go-common/router"
	"github.com/digitalmonsters/go-common/swagger"
	"github.com/digitalmonsters/go-common/wrappers/user_go"
	"github.com/digitalmonsters/music/cmd/api"
	"github.com/digitalmonsters/music/pkg/creators"
	"github.com/digitalmonsters/music/pkg/creators/categories"
	"github.com/digitalmonsters/music/pkg/creators/moods"
	"github.com/digitalmonsters/music/pkg/database"
	"github.com/digitalmonsters/music/utils"
	"github.com/pkg/errors"
	"gopkg.in/guregu/null.v4"
	"net/http"
)

func InitPublicApi(publicRouter *router.HttpRouter, apiDef map[string]swagger.ApiDescription, creatorsService *creators.Service, userGoWrapper user_go.IUserGoWrapper) error {
	if err := publicRouter.RegisterRestCmd(router.NewRestCommand(func(request []byte, executionData router.MethodExecutionData) (interface{}, *error_codes.ErrorWithCode) {
		var req creators.BecomeMusicCreatorRequest

		if err := json.Unmarshal(request, &req); err != nil {
			return nil, error_codes.NewErrorWithCodeRef(err, error_codes.GenericMappingError)
		}

		if len(req.LibraryLink) == 0 {
			return nil, error_codes.NewErrorWithCodeRef(errors.New("library_link is required"), error_codes.GenericValidationError)
		}

		if err := creatorsService.BecomeMusicCreator(req, database.GetDbWithContext(database.DbTypeMaster, executionData.Context), executionData, userGoWrapper); err != nil {
			return nil, error_codes.NewErrorWithCodeRef(err, error_codes.GenericServerError)
		}

		return api.SuccessResponse{
			Success: true,
		}, nil
	}, "/creators/request/send", http.MethodPost).RequireIdentityValidation().Build()); err != nil {
		return err
	}

	if err := publicRouter.RegisterRestCmd(router.NewRestCommand(func(request []byte, executionData router.MethodExecutionData) (interface{}, *error_codes.ErrorWithCode) {
		var req creators.UploadNewSongRequest

		if err := json.Unmarshal(request, &req); err != nil {
			return nil, error_codes.NewErrorWithCodeRef(err, error_codes.GenericMappingError)
		}

		if len(req.Name) == 0 {
			return nil, error_codes.NewErrorWithCodeRef(errors.New("name is required"), error_codes.GenericValidationError)
		}

		if len(req.ShortSongUrl) == 0 {
			return nil, error_codes.NewErrorWithCodeRef(errors.New("short_song_url is required"), error_codes.GenericValidationError)
		}

		if len(req.FullSongUrl) == 0 {
			return nil, error_codes.NewErrorWithCodeRef(errors.New("full_song_url is required"), error_codes.GenericValidationError)
		}

		if req.CategoryId == 0 {
			return nil, error_codes.NewErrorWithCodeRef(errors.New("category_id is required"), error_codes.GenericValidationError)
		}

		if req.MoodId == 0 {
			return nil, error_codes.NewErrorWithCodeRef(errors.New("mood_id is required"), error_codes.GenericValidationError)
		}

		if len(req.MusicAuthor) == 0 {
			return nil, error_codes.NewErrorWithCodeRef(errors.New("music_author is required"), error_codes.GenericValidationError)
		}

		resp, err := creatorsService.UploadNewSong(req, database.GetDbWithContext(database.DbTypeMaster, executionData.Context), executionData)
		if err != nil {
			return nil, error_codes.NewErrorWithCodeRef(err, error_codes.GenericServerError)
		}

		return resp, nil
	}, "/creators/song/add", http.MethodPost).RequireIdentityValidation().Build()); err != nil {
		return err
	}

	if err := publicRouter.RegisterRestCmd(router.NewRestCommand(func(request []byte, executionData router.MethodExecutionData) (interface{}, *error_codes.ErrorWithCode) {
		resp, err := creatorsService.CheckRequestStatus(executionData.UserId, database.GetDbWithContext(database.DbTypeReadonly, executionData.Context))
		if err != nil {
			return nil, error_codes.NewErrorWithCodeRef(err, error_codes.GenericServerError)
		}

		return resp, nil
	}, "/creators/request/status", http.MethodGet).RequireIdentityValidation().Build()); err != nil {
		return err
	}

	if err := publicRouter.RegisterRestCmd(router.NewRestCommand(func(request []byte, executionData router.MethodExecutionData) (interface{}, *error_codes.ErrorWithCode) {
		cursor := utils.ExtractString(executionData.GetUserValue, "cursor", "")
		categoryName := utils.ExtractString(executionData.GetUserValue, "name", "")
		count := utils.ExtractInt64(executionData.GetUserValue, "count", 10, 50)

		resp, err := categories.PublicList(categories.PublicListRequest{
			Name:   null.StringFrom(categoryName),
			Count:  int(count),
			Cursor: cursor,
		}, database.GetDbWithContext(database.DbTypeReadonly, executionData.Context))

		if err != nil {
			return nil, error_codes.NewErrorWithCodeRef(err, error_codes.GenericServerError)
		}

		return resp, nil
	}, "/categories/list", http.MethodGet).Build()); err != nil {
		return err
	}

	if err := publicRouter.RegisterRestCmd(router.NewRestCommand(func(request []byte, executionData router.MethodExecutionData) (interface{}, *error_codes.ErrorWithCode) {
		cursor := utils.ExtractString(executionData.GetUserValue, "cursor", "")
		categoryName := utils.ExtractString(executionData.GetUserValue, "name", "")
		count := utils.ExtractInt64(executionData.GetUserValue, "count", 10, 50)

		resp, err := moods.PublicList(moods.PublicListRequest{
			Name:   null.StringFrom(categoryName),
			Count:  int(count),
			Cursor: cursor,
		}, database.GetDbWithContext(database.DbTypeReadonly, executionData.Context))

		if err != nil {
			return nil, error_codes.NewErrorWithCodeRef(err, error_codes.GenericServerError)
		}

		return resp, nil
	}, "/moods/list", http.MethodGet).Build()); err != nil {
		return err
	}

	apiDef["/creators/request/send"] = swagger.ApiDescription{
		Request:           creators.BecomeMusicCreatorRequest{},
		Response:          api.SuccessResponse{},
		MethodDescription: "become music creator",
		Tags:              []string{"creator"},
	}

	apiDef["/creators/request/status"] = swagger.ApiDescription{
		Request:           nil,
		Response:          creators.CheckRequestStatusResponse{},
		MethodDescription: "check creator status",
		Tags:              []string{"creator"},
	}

	apiDef["/creators/song/add"] = swagger.ApiDescription{
		Request:           creators.UploadNewSongRequest{},
		Response:          database.CreatorSong{},
		MethodDescription: "add new song",
		Tags:              []string{"creator"},
	}

	apiDef["/categories/list"] = swagger.ApiDescription{
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
			{
				Name:        "name",
				In:          swagger.ParameterInQuery,
				Description: "name filter for categories",
				Required:    false,
				Type:        "string",
			},
		},
		Response:          categories.PublicListResponse{},
		MethodDescription: "categories list",
		Tags:              []string{"category"},
	}

	apiDef["/moods/list"] = swagger.ApiDescription{
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
			{
				Name:        "name",
				In:          swagger.ParameterInQuery,
				Description: "name filter for categories",
				Required:    false,
				Type:        "string",
			},
		},
		Response:          moods.PublicListResponse{},
		MethodDescription: "moods list",
		Tags:              []string{"mood"},
	}

	return nil
}
