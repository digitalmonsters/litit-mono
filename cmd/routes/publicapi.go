package routes

import (
	"github.com/digitalmonsters/comments/pkg/publicapi"
	"github.com/digitalmonsters/go-common/common"
	"github.com/digitalmonsters/go-common/error_codes"
	"github.com/digitalmonsters/go-common/router"
	"gorm.io/gorm"
	"net/http"
)

func InitPublicRoutes(httpRouter *router.HttpRouter, db *gorm.DB) error {
	if err := httpRouter.RegisterRestCmd(router.NewRestCommand(func(request []byte,
		executionData router.MethodExecutionData) (interface{}, *error_codes.ErrorWithCode) {
		publicapi.GetCommendById(db.WithContext(executionData.Context))
	}, "/{id}", http.MethodGet, common.AccessLevelPublic, false, false)); err != nil {
		return err
	}

	if err := httpRouter.RegisterRestCmd(router.NewRestCommand(func(request []byte,
		executionData router.MethodExecutionData) (interface{}, *error_codes.ErrorWithCode) {
		publicapi.DeleteCommentById()
	}, "/{id}", http.MethodDelete, common.AccessLevelPublic, true, true)); err != nil {
		return err
	}

	if err := httpRouter.RegisterRestCmd(router.NewRestCommand(func(request []byte,
		executionData router.MethodExecutionData) (interface{}, *error_codes.ErrorWithCode) {
		publicapi.UpdateCommentById()
	}, "/{id}", http.MethodPatch, common.AccessLevelPublic, true, true)); err != nil {
		return err
	}

	if err := httpRouter.RegisterRestCmd(router.NewRestCommand(func(request []byte,
		executionData router.MethodExecutionData) (interface{}, *error_codes.ErrorWithCode) {
		publicapi.GetRepliesByCommentId()
	}, "/{id}/replies", http.MethodGet, common.AccessLevelPublic, false, false)); err != nil {
		return err
	}

	if err := httpRouter.RegisterRestCmd(router.NewRestCommand(func(request []byte,
		executionData router.MethodExecutionData) (interface{}, *error_codes.ErrorWithCode) {
		publicapi.VoteComment()
	}, "/{id}/vote", http.MethodPost, common.AccessLevelPublic, true, false)); err != nil {
		return err
	}

	if err := httpRouter.RegisterRestCmd(router.NewRestCommand(func(request []byte,
		executionData router.MethodExecutionData) (interface{}, *error_codes.ErrorWithCode) {
		publicapi.ReportComment()
	}, "/{id}/report", http.MethodPost, common.AccessLevelPublic, true, false)); err != nil {
		return err
	}

	if err := httpRouter.RegisterRestCmd(router.NewRestCommand(func(request []byte,
		executionData router.MethodExecutionData) (interface{}, *error_codes.ErrorWithCode) {
		publicapi.GetCommentByTypeWithResourceId()
	}, "/content/{id}", http.MethodDelete, common.AccessLevelPublic, false, false)); err != nil {
		return err
	}

	if err := httpRouter.RegisterRestCmd(router.NewRestCommand(func(request []byte,
		executionData router.MethodExecutionData) (interface{}, *error_codes.ErrorWithCode) {
		publicapi.SendComment()
	}, "/{content_id}", http.MethodPost, common.AccessLevelPublic, true, true)); err != nil {
		return err
	}

	return nil
}
