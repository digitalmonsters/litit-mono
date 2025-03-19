package admin_api

import (
	"encoding/json"

	"github.com/digitalmonsters/comments/pkg/database"
	"github.com/digitalmonsters/comments/pkg/report"
	"github.com/digitalmonsters/go-common/common"
	"github.com/digitalmonsters/go-common/error_codes"
	"github.com/digitalmonsters/go-common/router"
)

func (a *adminApiApp) initAdminApi(endpoint router.IRpcEndpoint) error {
	commands := []router.ICommand{
		a.getReportedUserProfileComments(),
		a.getReportedVideoComments(),
		a.GetReportsForComment(),
		a.ApproveRejectReportedComment(),
	}

	for _, c := range commands {
		if err := endpoint.RegisterRpcCommand(c); err != nil {
			return err
		}
	}

	return nil
}

func (a *adminApiApp) getReportedUserProfileComments() router.ICommand {
	method := "GetReportedUserProfileComments"

	return router.NewAdminCommand(method, func(request []byte, executionData router.MethodExecutionData) (interface{}, *error_codes.ErrorWithCode) {
		var req report.GetReportedUserProfileCommentsRequest

		if err := json.Unmarshal(request, &req); err != nil {
			return nil, error_codes.NewErrorWithCodeRef(err, error_codes.GenericMappingError)
		}

		resp, err := report.GetReportedUserProfileComments(req, database.GetDb(), a.userWrapper, executionData.Context)
		if err != nil {
			return nil, error_codes.NewErrorWithCodeRef(err, error_codes.GenericServerError)
		}

		return resp, nil
	}, common.AccessLevelRead, "report:comments:profile")
}

func (a *adminApiApp) getReportedVideoComments() router.ICommand {
	method := "GetReportedVideoComments"

	return router.NewAdminCommand(method, func(request []byte, executionData router.MethodExecutionData) (interface{}, *error_codes.ErrorWithCode) {
		var req report.GetReportedVideoCommentsRequest

		if err := json.Unmarshal(request, &req); err != nil {
			return nil, error_codes.NewErrorWithCodeRef(err, error_codes.GenericMappingError)
		}

		resp, err := report.GetReportedVideoComments(req, database.GetDb(), a.userWrapper, a.contentWrapper, executionData.Context, executionData.ApmTransaction)
		if err != nil {
			return nil, error_codes.NewErrorWithCodeRef(err, error_codes.GenericServerError)
		}

		return resp, nil
	}, common.AccessLevelRead, "report:comments:content")
}

func (a *adminApiApp) GetReportsForComment() router.ICommand {
	method := "GetReportsForComment"

	return router.NewAdminCommand(method, func(request []byte, executionData router.MethodExecutionData) (interface{}, *error_codes.ErrorWithCode) {
		var req report.GetReportsForCommentRequest

		if err := json.Unmarshal(request, &req); err != nil {
			return nil, error_codes.NewErrorWithCodeRef(err, error_codes.GenericMappingError)
		}

		resp, err := report.GetReportsForComment(req, database.GetDb(), a.userWrapper, executionData.Context)
		if err != nil {
			return nil, error_codes.NewErrorWithCodeRef(err, error_codes.GenericServerError)
		}

		return resp, nil
	}, common.AccessLevelRead, "report:comments:view")
}

func (a *adminApiApp) ApproveRejectReportedComment() router.ICommand {
	method := "ApproveRejectReportedComment"

	return router.NewAdminCommand(method, func(request []byte, executionData router.MethodExecutionData) (interface{}, *error_codes.ErrorWithCode) {
		var req report.ApproveRejectReportedCommentRequest

		if err := json.Unmarshal(request, &req); err != nil {
			return nil, error_codes.NewErrorWithCodeRef(err, error_codes.GenericMappingError)
		}

		resp, err := report.ApproveRejectReportedComment(req, database.GetDb(), executionData.Context)
		if err != nil {
			return nil, error_codes.NewErrorWithCodeRef(err, error_codes.GenericServerError)
		}

		return resp, nil
	}, common.AccessLevelWrite, "report:comments:approve")
}
