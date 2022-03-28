package api

import (
	"encoding/json"
	"github.com/digitalmonsters/go-common/error_codes"
	"github.com/digitalmonsters/go-common/extract"
	"github.com/digitalmonsters/go-common/router"
	"github.com/digitalmonsters/go-common/swagger"
	"github.com/digitalmonsters/notification-handler/pkg/database"
	tokenPkg "github.com/digitalmonsters/notification-handler/pkg/token"
	"github.com/pkg/errors"
	"net/http"
)

func InitTokenApi(httpRouter *router.HttpRouter, apiDef map[string]swagger.ApiDescription) error {
	createTokenPath := "/mobile/v1/push/token"
	deleteTokenPath := "/mobile/v1/push/token/{device_id}"

	if err := httpRouter.RegisterRestCmd(router.NewRestCommand(func(request []byte,
		executionData router.MethodExecutionData) (interface{}, *error_codes.ErrorWithCode) {
		userId := executionData.UserId

		if userId <= 0 {
			return nil, error_codes.NewErrorWithCodeRef(errors.New("invalid user_id"), error_codes.GenericValidationError)
		}

		var req tokenPkg.TokenCreateRequest

		if err := json.Unmarshal(request, &req); err != nil {
			return nil, error_codes.NewErrorWithCodeRef(err, error_codes.GenericMappingError)
		}

		if err := tokenPkg.CreateToken(database.GetDb(database.DbTypeMaster).WithContext(executionData.Context),
			executionData.UserId, req); err != nil {
			return nil, error_codes.NewErrorWithCodeRef(err, error_codes.GenericServerError)
		}

		return nil, nil
	}, createTokenPath, http.MethodPost, true, false)); err != nil {
		return err
	}

	if err := httpRouter.RegisterRestCmd(router.NewRestCommand(func(request []byte,
		executionData router.MethodExecutionData) (interface{}, *error_codes.ErrorWithCode) {
		userId := executionData.UserId

		if userId <= 0 {
			return nil, error_codes.NewErrorWithCodeRef(errors.New("invalid user_id"), error_codes.GenericValidationError)
		}

		deviceId := extract.String(executionData.GetUserValue, "device_id", "")

		if err := tokenPkg.DeleteToken(database.GetDb(database.DbTypeMaster).WithContext(executionData.Context),
			executionData.UserId, deviceId); err != nil {
			return nil, error_codes.NewErrorWithCodeRef(err, error_codes.GenericServerError)
		}

		return nil, nil
	}, deleteTokenPath, http.MethodDelete, true, false)); err != nil {
		return err
	}

	apiDef[createTokenPath] = swagger.ApiDescription{
		Request:           tokenPkg.TokenCreateRequest{},
		MethodDescription: "create token",
		Tags:              []string{"token"},
	}

	apiDef[deleteTokenPath] = swagger.ApiDescription{
		AdditionalSwaggerParameters: []swagger.ParameterDescription{
			{
				Name:        "device_id",
				In:          swagger.ParameterInPath,
				Description: "device id",
				Required:    true,
				Type:        "string",
			},
		},
		MethodDescription: "delete token",
		Tags:              []string{"token"},
	}

	return nil
}
