package api

import (
	"encoding/json"
	"fmt"
	"github.com/digitalmonsters/go-common/error_codes"
	"github.com/digitalmonsters/go-common/router"
	"github.com/digitalmonsters/go-common/swagger"
	"github.com/digitalmonsters/music/cmd/api"
	"github.com/digitalmonsters/music/pkg/creators"
	"github.com/digitalmonsters/music/pkg/database"
	"github.com/digitalmonsters/music/utils"
	"github.com/pkg/errors"
	"net/http"
)

func (c *creatorApp) initPublicApi() error {
	restCommands := []*router.RestCommand{
		c.sendCreatorRequest(),
		c.addSong(),
		c.getRequestStatus(),
		c.getMySongs(),
		c.getUserSongs(),
	}

	for _, command := range restCommands {
		if err := c.httpRouter.RegisterRestCmd(command); err != nil {
			return err
		}
	}
	return nil
}

func (c *creatorApp) sendCreatorRequest() *router.RestCommand {
	path := "/creators/request/send"

	c.apiDef[path] = swagger.ApiDescription{
		Request:           creators.BecomeMusicCreatorRequest{},
		Response:          api.SuccessResponse{},
		MethodDescription: "become music creator",
		Tags:              []string{"creator"},
	}

	return router.NewRestCommand(func(request []byte, executionData router.MethodExecutionData) (interface{}, *error_codes.ErrorWithCode) {
		var req creators.BecomeMusicCreatorRequest

		if err := json.Unmarshal(request, &req); err != nil {
			return nil, error_codes.NewErrorWithCodeRef(err, error_codes.GenericMappingError)
		}

		if len(req.LibraryLink) == 0 {
			return nil, error_codes.NewErrorWithCodeRef(errors.New("library_link is required"), error_codes.GenericValidationError)
		}

		if err := c.creatorsService.BecomeMusicCreator(req, database.GetDbWithContext(database.DbTypeMaster, executionData.Context), executionData, c.userGoWrapper); err != nil {
			return nil, error_codes.NewErrorWithCodeRef(err, error_codes.GenericServerError)
		}

		return api.SuccessResponse{
			Success: true,
		}, nil
	}, path, http.MethodPost).RequireIdentityValidation().Build()
}

func (c *creatorApp) addSong() *router.RestCommand {
	path := "/creators/song/add"

	c.apiDef[path] = swagger.ApiDescription{
		Request:           creators.UploadNewSongRequest{},
		Response:          database.CreatorSong{},
		MethodDescription: "add new song",
		Tags:              []string{"creator"},
	}

	return router.NewRestCommand(func(request []byte, executionData router.MethodExecutionData) (interface{}, *error_codes.ErrorWithCode) {
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

		if len(req.Hashtags) > c.appConfig.Values.MUSIC_MAX_HASHTAGS_COUNT {
			return nil, error_codes.NewErrorWithCodeRef(fmt.Errorf("max hashtags limit is %v", c.appConfig.Values.MUSIC_MAX_HASHTAGS_COUNT), error_codes.GenericValidationError)
		}

		resp, err := c.creatorsService.UploadNewSong(req, c.contentWrapper, database.GetDbWithContext(database.DbTypeMaster, executionData.Context), executionData)
		if err != nil {
			return nil, error_codes.NewErrorWithCodeRef(err, error_codes.GenericServerError)
		}

		return resp, nil
	}, path, http.MethodPost).RequireIdentityValidation().Build()
}

func (c *creatorApp) getRequestStatus() *router.RestCommand {
	path := "/creators/request/status"

	c.apiDef[path] = swagger.ApiDescription{
		Request:           nil,
		Response:          creators.CheckRequestStatusResponse{},
		MethodDescription: "check creator status",
		Tags:              []string{"creator"},
	}

	return router.NewRestCommand(func(request []byte, executionData router.MethodExecutionData) (interface{}, *error_codes.ErrorWithCode) {
		resp, err := c.creatorsService.CheckRequestStatus(executionData.UserId, database.GetDbWithContext(database.DbTypeReadonly, executionData.Context))
		if err != nil {
			return nil, error_codes.NewErrorWithCodeRef(err, error_codes.GenericServerError)
		}

		return resp, nil
	}, path, http.MethodGet).RequireIdentityValidation().Build()
}

func (c *creatorApp) getMySongs() *router.RestCommand {
	path := "/me/songs"

	c.apiDef[path] = swagger.ApiDescription{
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
		Response:          creators.SongsListResponse{},
		MethodDescription: "my songs list",
		Tags:              []string{"creator", "song"},
	}

	return router.NewRestCommand(func(request []byte, executionData router.MethodExecutionData) (interface{}, *error_codes.ErrorWithCode) {
		cursor := utils.ExtractString(executionData.GetUserValue, "cursor", "")
		count := utils.ExtractInt64(executionData.GetUserValue, "count", 10, 50)

		return c.creatorsService.SongsList(creators.SongsListRequest{
			UserId: executionData.UserId,
			Count:  int(count),
			Cursor: cursor,
		}, executionData.UserId, database.GetDbWithContext(database.DbTypeReadonly, executionData.Context), executionData)
	}, path, http.MethodGet).RequireIdentityValidation().Build()
}

func (c *creatorApp) getUserSongs() *router.RestCommand {
	path := "/{user_id}/songs"

	c.apiDef[path] = swagger.ApiDescription{
		AdditionalSwaggerParameters: []swagger.ParameterDescription{
			{
				Name:        "user_id",
				In:          swagger.ParameterInPath,
				Description: "requested user id",
				Required:    true,
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
		Response:          creators.SongsListResponse{},
		MethodDescription: "my songs list",
		Tags:              []string{"creator", "song"},
	}

	return router.NewRestCommand(func(request []byte, executionData router.MethodExecutionData) (interface{}, *error_codes.ErrorWithCode) {
		cursor := utils.ExtractString(executionData.GetUserValue, "cursor", "")
		count := utils.ExtractInt64(executionData.GetUserValue, "count", 10, 50)

		userId := utils.ExtractInt64(executionData.GetUserValue, "user_id", 0, 0)
		if userId <= 0 {
			return nil, error_codes.NewErrorWithCodeRef(errors.New("invalid user_id"), error_codes.GenericValidationError)
		}

		return c.creatorsService.SongsList(creators.SongsListRequest{
			UserId: userId,
			Count:  int(count),
			Cursor: cursor,
		}, executionData.UserId, database.GetDbWithContext(database.DbTypeReadonly, executionData.Context), executionData)
	}, path, http.MethodGet).Build()
}
