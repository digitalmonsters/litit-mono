package api

import (
	"encoding/json"
	"log"
	"net/http"

	"github.com/digitalmonsters/go-common/error_codes"
	"github.com/digitalmonsters/go-common/extract"
	"github.com/digitalmonsters/go-common/router"
	"github.com/digitalmonsters/go-common/wrappers/auth_go"
	"github.com/digitalmonsters/go-common/wrappers/follow"
	"github.com/digitalmonsters/go-common/wrappers/user_go"
	"github.com/digitalmonsters/notification-handler/pkg/database"
	notificationPkg "github.com/digitalmonsters/notification-handler/pkg/notification"
	"github.com/google/uuid"
	ua "github.com/mileusna/useragent"
	"github.com/pkg/errors"
)

func InitNotificationApi(httpRouter *router.HttpRouter, userGoWrapper user_go.IUserGoWrapper, authWrapper auth_go.IAuthGoWrapper,
	followWrapper follow.IFollowWrapper) error {
	notificationsPath := "/mobile/v1/notifications"
	deleteNotificationPath := "/mobile/v1/notifications/{id}"
	readAllNotificationsPath := "/mobile/v1/notifications/reset"
	readNotificationPath := "/mobile/v1/notification/read"
	inAppNotificationsPath := "/mobile/v1/app/notifications"
	notificationAnalyticsPath := "/v1/notification/analytics"
	// notificationTracker := "/mobile/v1/tracknotification"

	if err := httpRouter.RegisterRestCmd(router.NewRestCommand(func(request []byte,
		executionData router.MethodExecutionData) (interface{}, *error_codes.ErrorWithCode) {
		log.Println("Started")
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

		db := database.GetDb(database.DbTypeMaster).WithContext(executionData.Context)

		if err := notificationPkg.NotificationEventAPILog(executionData.UserId, req.NotificationId, executionData.DeviceId, db); err != nil {
		}

		return nil, nil
	}, readNotificationPath, http.MethodPost).RequireIdentityValidation().Build()); err != nil {
		return err
	}

	if err := httpRouter.RegisterRestCmd(router.NewRestCommand(func(request []byte,
		executionData router.MethodExecutionData) (interface{}, *error_codes.ErrorWithCode) {

		db := database.GetDb(database.DbTypeMaster).WithContext(executionData.Context)

		data, err := notificationPkg.NotificationAnalyticsByDevice(db)
		if err != nil {
			return nil, error_codes.NewErrorWithCodeRef(err, error_codes.GenericServerError)
		}

		return data, nil
	}, notificationAnalyticsPath, http.MethodGet).RequireIdentityValidation().Build()); err != nil {
		return err
	}

	return nil
}
