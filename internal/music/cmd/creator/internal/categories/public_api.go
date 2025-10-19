package categories

import (
	"github.com/digitalmonsters/go-common/error_codes"
	"github.com/digitalmonsters/go-common/router"
	"github.com/digitalmonsters/go-common/swagger"
	"github.com/digitalmonsters/music/pkg/creators/categories"
	"github.com/digitalmonsters/music/pkg/database"
	"github.com/digitalmonsters/music/utils"
	"gopkg.in/guregu/null.v4"
	"net/http"
)

func (c *categoriesApp) initPublicApi() error {
	restCommands := []*router.RestCommand{
		c.getCategoriesList(),
	}

	for _, command := range restCommands {
		if err := c.httpRouter.RegisterRestCmd(command); err != nil {
			return err
		}
	}
	return nil
}

func (c *categoriesApp) getCategoriesList() *router.RestCommand {
	path := "/categories/list"

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

	return router.NewRestCommand(func(request []byte, executionData router.MethodExecutionData) (interface{}, *error_codes.ErrorWithCode) {
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
	}, path, http.MethodGet).Build()
}
