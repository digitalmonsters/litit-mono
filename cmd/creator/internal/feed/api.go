package feed

import (
	"github.com/digitalmonsters/go-common/apm_helper"
	"github.com/digitalmonsters/go-common/error_codes"
	"github.com/digitalmonsters/go-common/extract"
	"github.com/digitalmonsters/go-common/router"
	"github.com/digitalmonsters/go-common/swagger"
	"github.com/digitalmonsters/music/pkg/database"
	"github.com/digitalmonsters/music/pkg/frontend"
	"net/http"
)

func (f *feedApp) initPublicApi() error {
	restCommands := []*router.RestCommand{
		f.getFeed(),
	}

	for _, command := range restCommands {
		if err := f.httpRouter.RegisterRestCmd(command); err != nil {
			return err
		}
	}
	return nil
}

func (f *feedApp) getFeed() *router.RestCommand {
	path := "/feed"

	f.apiDef[path] = swagger.ApiDescription{
		AdditionalSwaggerParameters: []swagger.ParameterDescription{
			{
				Name:        "count",
				In:          swagger.ParameterInQuery,
				Description: "songs count in feed",
				Required:    false,
				Type:        "integer",
			},
		},
		Response:          []frontend.CreatorSongModel{},
		MethodDescription: "music feed",
		Tags:              []string{"feed"},
	}

	return router.NewRestCommand(func(request []byte, executionData router.MethodExecutionData) (interface{}, *error_codes.ErrorWithCode) {
		count := extract.Int64(executionData.GetUserValue, "count", 10, 20)

		if executionData.UserId > 0 {
			apm_helper.AddApmLabel(executionData.ApmTransaction, "user_id", executionData.UserId)
		}

		return f.musicFeedService.GetFeed(database.GetDbWithContext(database.DbTypeReadonly, executionData.Context),
			executionData.UserId, int(count), executionData)
	}, path, http.MethodGet).RequireIdentityValidation().Build()
}
