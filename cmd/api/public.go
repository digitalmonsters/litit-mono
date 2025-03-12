package api

import (
	"github.com/digitalmonsters/ads-manager/pkg/database"
	"github.com/digitalmonsters/ads-manager/pkg/message"
	"github.com/digitalmonsters/ads-manager/utils"
	"github.com/digitalmonsters/go-common/error_codes"
	"github.com/digitalmonsters/go-common/router"
	"github.com/digitalmonsters/go-common/wrappers/user_go"
)

func InitPublicApi(httpRouter *router.HttpRouter, userGoWrapper user_go.IUserGoWrapper) error {
	getAdsMessagePath := "/ads/message/me"
	messagePath := "/message/me"
	adsAvailablePath := "/ads/available"

	if err := httpRouter.RegisterRestCmd(router.NewRestCommand(func(request []byte, executionData router.MethodExecutionData) (interface{}, *error_codes.ErrorWithCode) {
		messageType := database.MessageType(int(utils.ExtractInt64(executionData.GetUserValue, "type", 1, 50)))
		if messageType < database.MessageTypeMobile || messageType > database.MessageTypeWeb {
			messageType = database.MessageTypeMobile
		}

		resp, err := message.GetMessageForUser(executionData.UserId, messageType, database.GetDbWithContext(database.DbTypeReadonly, executionData.Context), userGoWrapper, executionData)
		if err != nil {
			return nil, error_codes.NewErrorWithCodeRef(err, error_codes.GenericServerError)
		}

		return resp, nil
	}, getAdsMessagePath, router.MethodGet).RequireIdentityValidation().Build()); err != nil {
		return err
	}

	if err := httpRouter.RegisterRestCmd(router.NewRestCommand(func(request []byte, executionData router.MethodExecutionData) (interface{}, *error_codes.ErrorWithCode) {
		messageType := database.MessageType(int(utils.ExtractInt64(executionData.GetUserValue, "type", 1, 50)))
		if messageType < database.MessageTypeMobile || messageType > database.MessageTypeWeb {
			messageType = database.MessageTypeMobile
		}

		resp, err := message.GetMessageForUser(executionData.UserId, messageType, database.GetDbWithContext(database.DbTypeReadonly, executionData.Context), userGoWrapper, executionData)
		if err != nil {
			return nil, error_codes.NewErrorWithCodeRef(err, error_codes.GenericServerError)
		}

		return resp, nil
	}, messagePath, router.MethodGet).RequireIdentityValidation().Build()); err != nil {
		return err
	}

	if err := httpRouter.RegisterRestCmd(router.NewRestCommand(func(request []byte, executionData router.MethodExecutionData) (interface{}, *error_codes.ErrorWithCode) {
		return adsAvailableResponse{
			IsAvailableForUser: message.IsAdsAvailableForUser(executionData.UserId),
		}, nil
	}, adsAvailablePath, router.MethodGet).Build()); err != nil {
		return err
	}

	return nil
}
