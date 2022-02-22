package creator

import (
	"encoding/json"
	"github.com/digitalmonsters/go-common/error_codes"
	"github.com/digitalmonsters/go-common/router"
	"github.com/digitalmonsters/go-common/swagger"
	"github.com/digitalmonsters/music/cmd/api"
	"github.com/digitalmonsters/music/pkg/creators"
	"github.com/digitalmonsters/music/pkg/database"
	"github.com/pkg/errors"
	"net/http"
)

func InitPublicApi(publicRouter *router.HttpRouter, apiDef map[string]swagger.ApiDescription) error {
	if err := publicRouter.RegisterRestCmd(router.NewRestCommand(func(request []byte, executionData router.MethodExecutionData) (interface{}, *error_codes.ErrorWithCode) {
		var req creators.BecomeMusicCreatorRequest

		if err := json.Unmarshal(request, &req); err != nil {
			return nil, error_codes.NewErrorWithCodeRef(err, error_codes.GenericMappingError)
		}

		if len(req.LibraryLink) == 0 {
			return nil, error_codes.NewErrorWithCodeRef(errors.New("library_link is required"), error_codes.GenericValidationError)
		}

		if err := creators.BecomeMusicCreator(req, database.GetDbWithContext(database.DbTypeMaster, executionData.Context), executionData); err != nil {
			return nil, error_codes.NewErrorWithCodeRef(err, error_codes.GenericServerError)
		}

		return api.SuccessResponse{
			Success: true,
		}, nil
	}, "/creators/request/send", http.MethodPost, true, false)); err != nil {
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

		if len(req.MusicAuthor) == 0 {
			return nil, error_codes.NewErrorWithCodeRef(errors.New("music_author is required"), error_codes.GenericValidationError)
		}

		resp, err := creators.UploadNewSong(req, database.GetDbWithContext(database.DbTypeMaster, executionData.Context), executionData)
		if err != nil {
			return nil, error_codes.NewErrorWithCodeRef(err, error_codes.GenericServerError)
		}

		return resp, nil
	}, "/creators/song/add", http.MethodPost, true, false)); err != nil {
		return err
	}

	apiDef["/creators/request/send"] = swagger.ApiDescription{
		Request:           creators.BecomeMusicCreatorRequest{},
		Response:          api.SuccessResponse{},
		MethodDescription: "become music creator",
		Tags:              []string{"creator"},
	}

	apiDef["/creators/song/add"] = swagger.ApiDescription{
		Request:           creators.UploadNewSongRequest{},
		Response:          database.CreatorSong{},
		MethodDescription: "add new song",
		Tags:              []string{"creator"},
	}

	return nil
}
