package api

import (
	"encoding/json"
	"net/http"

	"github.com/digitalmonsters/go-common/error_codes"
	"github.com/digitalmonsters/go-common/extract"
	"github.com/digitalmonsters/go-common/router"
	"github.com/digitalmonsters/go-common/swagger"
	"github.com/digitalmonsters/go-common/wrappers/follow"
	"github.com/digitalmonsters/go-common/wrappers/user_go"
	"github.com/digitalmonsters/notification-handler/pkg/database"
	notificationPkg "github.com/digitalmonsters/notification-handler/pkg/notification"
	"github.com/google/uuid"
	ua "github.com/mileusna/useragent"
	"github.com/pkg/errors"
)

func InitNotificationApi(httpRouter *router.HttpRouter, apiDef map[string]swagger.ApiDescription, userGoWrapper user_go.IUserGoWrapper,
	followWrapper follow.IFollowWrapper) error {
	notificationsPath := "/mobile/v1/notifications"
	deleteNotificationPath := "/mobile/v1/notifications/{id}"
	readAllNotificationsPath := "/mobile/v1/notifications/reset"
	readNotificationPath := "/mobile/v1/notification/read"
	inAppNotificationsPath := "/mobile/v1/app/notifications"

	if err := httpRouter.RegisterRestCmd(router.NewRestCommand(func(request []byte,
		executionData router.MethodExecutionData) (interface{}, *error_codes.ErrorWithCode) {
		page := extract.String(executionData.GetUserValue, "page", "")
		typeGroup := notificationPkg.TypeGroup(extract.String(executionData.GetUserValue, "notification_type", string(notificationPkg.TypeGroupAll)))

		userId := executionData.UserId

		if userId <= 0 {
			return nil, error_codes.NewErrorWithCodeRef(errors.New("invalid user_id"), error_codes.GenericValidationError)
		}

		pushAdminSupported := true

		userAgent, ok := executionData.GetUserValue("User-Agent").(string)
		if ok {
			userAgentParsed := ua.Parse(userAgent)

			if userAgentParsed.IsIOS() {
				pushAdminSupported = false
			}
		}

		resp, err := notificationPkg.GetNotificationsLegacy(database.GetDb(database.DbTypeReadonly).WithContext(executionData.Context),
			executionData.UserId, page, typeGroup, pushAdminSupported, 10, userGoWrapper, followWrapper, executionData.Context)
		if err != nil {
			return nil, error_codes.NewErrorWithCodeRef(err, error_codes.GenericServerError)
		}

		return resp, nil
	}, notificationsPath, http.MethodGet).RequireIdentityValidation().Build()); err != nil {
		return err
	}

	if err := httpRouter.RegisterRestCmd(router.NewRestCommand(func(request []byte,
		executionData router.MethodExecutionData) (interface{}, *error_codes.ErrorWithCode) {
		userId := executionData.UserId

		if userId <= 0 {
			return nil, error_codes.NewErrorWithCodeRef(errors.New("invalid user_id"), error_codes.GenericValidationError)
		}

		resp, err := notificationPkg.GetInAppNotifications(userId, database.GetDb(database.DbTypeReadonly).WithContext(executionData.Context))
		if err != nil {
			return nil, error_codes.NewErrorWithCodeRef(err, error_codes.GenericServerError)
		}
		return resp, nil
	}, inAppNotificationsPath, http.MethodGet).RequireIdentityValidation().Build()); err != nil {
		return err
	}

	if err := httpRouter.RegisterRestCmd(router.NewRestCommand(func(request []byte,
		executionData router.MethodExecutionData) (interface{}, *error_codes.ErrorWithCode) {
		id, err := uuid.Parse(extract.String(executionData.GetUserValue, "id", uuid.Nil.String()))
		if err != nil || id.String() == uuid.Nil.String() {
			return nil, error_codes.NewErrorWithCodeRef(errors.New("invalid id"), error_codes.GenericValidationError)
		}

		userId := executionData.UserId

		if userId <= 0 {
			return nil, error_codes.NewErrorWithCodeRef(errors.New("invalid user_id"), error_codes.GenericValidationError)
		}

		if err = notificationPkg.DeleteNotification(database.GetDb(database.DbTypeMaster).WithContext(executionData.Context),
			executionData.UserId, id); err != nil {
			return nil, error_codes.NewErrorWithCodeRef(err, error_codes.GenericServerError)
		}

		return nil, nil
	}, deleteNotificationPath, http.MethodDelete).RequireIdentityValidation().Build()); err != nil {
		return err
	}

	if err := httpRouter.RegisterRestCmd(router.NewRestCommand(func(request []byte,
		executionData router.MethodExecutionData) (interface{}, *error_codes.ErrorWithCode) {
		userId := executionData.UserId

		if userId <= 0 {
			return nil, error_codes.NewErrorWithCodeRef(errors.New("invalid user_id"), error_codes.GenericValidationError)
		}

		if err := notificationPkg.ReadAllNotifications(database.GetDb(database.DbTypeMaster).WithContext(executionData.Context),
			executionData.UserId); err != nil {
			return nil, error_codes.NewErrorWithCodeRef(err, error_codes.GenericServerError)
		}

		return nil, nil
	}, readAllNotificationsPath, http.MethodPatch).RequireIdentityValidation().Build()); err != nil {
		return err
	}

	if err := httpRouter.RegisterRestCmd(router.NewRestCommand(func(request []byte,
		executionData router.MethodExecutionData) (interface{}, *error_codes.ErrorWithCode) {
		userId := executionData.UserId
		if userId <= 0 {
			return nil, error_codes.NewErrorWithCodeRef(errors.New("invalid user_id"), error_codes.GenericValidationError)
		}

		var req notificationPkg.ReadNotificationRequest
		if err := json.Unmarshal(request, &req); err != nil {
			return nil, error_codes.NewErrorWithCodeRef(err, error_codes.GenericMappingError)
		}

		if err := notificationPkg.ReadNotification(req, executionData.UserId, executionData.Context); err != nil {
			return nil, error_codes.NewErrorWithCodeRef(err, error_codes.GenericServerError)
		}

		return nil, nil
	}, readNotificationPath, http.MethodPost).RequireIdentityValidation().Build()); err != nil {
		return err
	}

	apiDef[notificationsPath] = swagger.ApiDescription{
		AdditionalSwaggerParameters: []swagger.ParameterDescription{
			{
				Name:        "page",
				In:          swagger.ParameterInQuery,
				Description: "page",
				Required:    false,
				Type:        "string",
			},
			{
				Name:        "notification_type",
				In:          swagger.ParameterInQuery,
				Description: "all|comment|system|following",
				Required:    false,
				Type:        "string",
			},
		},
		Response:          notificationPkg.NotificationsResponse{},
		MethodDescription: "user notifications",
		Tags:              []string{"notification"},
	}

	apiDef[deleteNotificationPath] = swagger.ApiDescription{
		AdditionalSwaggerParameters: []swagger.ParameterDescription{
			{
				Name:        "id",
				In:          swagger.ParameterInPath,
				Description: "notification id",
				Required:    true,
				Type:        "uuid",
			},
		},
		MethodDescription: "delete notification",
		Tags:              []string{"notification"},
	}

	apiDef[readAllNotificationsPath] = swagger.ApiDescription{
		MethodDescription: "read all notifications",
		Tags:              []string{"notification"},
	}

	apiDef[readNotificationPath] = swagger.ApiDescription{
		Request:           notificationPkg.ReadNotificationRequest{},
		MethodDescription: "read notification",
		Tags:              []string{"notification"},
	}

	return nil
}
