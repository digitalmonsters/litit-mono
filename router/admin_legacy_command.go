package router

import (
	"context"
	"strings"

	"github.com/digitalmonsters/go-common/common"
	"github.com/digitalmonsters/go-common/error_codes"
	"github.com/digitalmonsters/go-common/rpc"
	"github.com/digitalmonsters/go-common/translation"
	"github.com/digitalmonsters/go-common/wrappers/auth_go"
	"github.com/pkg/errors"
	"github.com/valyala/fasthttp"
	"go.elastic.co/apm"
)

type LegacyAdminCommand struct {
	methodName                string
	accessLevel               common.AccessLevel
	forceLog                  bool
	fn                        CommandFunc
	requireIdentityValidation bool
	allowBanned               bool
}

func NewLegacyAdminCommand(methodName string, fn CommandFunc) ICommand {
	return &LegacyAdminCommand{
		methodName:                strings.ToLower(methodName),
		accessLevel:               common.AccessLevelWrite,
		forceLog:                  true,
		fn:                        fn,
		requireIdentityValidation: true,
		allowBanned:               false,
	}
}

func (a LegacyAdminCommand) CanExecute(httpCtx *fasthttp.RequestCtx, ctx context.Context, auth auth_go.IAuthGoWrapper, userValidator UserExecutorValidator) (int64, bool, bool, bool, translation.Language, string, *rpc.ExtendedLocalRpcError) {
	userId, isGuest, isBanned, isPet2User, language, customUserIp, err := publicCanExecuteLogic(httpCtx, a.requireIdentityValidation, a.allowBanned, userValidator)

	if err != nil {
		return 0, isGuest, isBanned, isPet2User, language, customUserIp, err
	}

	httpCtx.Response.Header.Set("Access-Control-Allow-Credentials", "true")
	httpCtx.Response.Header.SetBytesV("Access-Control-Allow-Origin", httpCtx.Request.Header.Peek("Origin"))
	httpCtx.Response.Header.Set("Access-Control-Allow-Headers", "*")
	httpCtx.Response.Header.Set("Access-Control-Allow-Methods", "POST, GET, OPTIONS, PATCH, DELETE")

	if userId <= 0 {
		err := errors.New("legacy admin method requires identity validation")

		return 0, isGuest, isBanned, isPet2User, language, customUserIp, &rpc.ExtendedLocalRpcError{
			RpcError: rpc.RpcError{
				Code:        error_codes.MissingJwtToken,
				Message:     "legacy admin method requires identity validation",
				Hostname:    hostName,
				ServiceName: hostName,
			},
			LocalHandlingError: err,
		}
	}

	resp := <-auth.CheckLegacyAdmin(userId, apm.TransactionFromContext(ctx), false)

	if resp.Error != nil {
		return 0, isGuest, isBanned, isPet2User, language, customUserIp, &rpc.ExtendedLocalRpcError{
			RpcError: *resp.Error,
		}
	}

	if resp.Resp.IsAdmin || resp.Resp.IsSuperAdmin {
		return userId, isGuest, isBanned, isPet2User, language, customUserIp, nil
	}

	err1 := errors.New("user is not marked as admin")
	return 0, isGuest, isBanned, isPet2User, language, customUserIp, &rpc.ExtendedLocalRpcError{
		RpcError: rpc.RpcError{
			Code:        error_codes.InvalidJwtToken,
			Message:     err1.Error(),
			Stack:       "",
			Hostname:    hostName,
			ServiceName: hostName,
		},
		LocalHandlingError: err1,
	}
}

func (a LegacyAdminCommand) GetPath() string {
	return a.GetMethodName()
}

func (a LegacyAdminCommand) GetHttpMethod() string {
	return "post"
}

func (a LegacyAdminCommand) ForceLog() bool {
	return a.forceLog
}

func (a LegacyAdminCommand) GetObj() string {
	return ""
}

func (a LegacyAdminCommand) RequireIdentityValidation() bool {
	return a.requireIdentityValidation
}

func (a LegacyAdminCommand) AllowBanned() bool {
	return a.allowBanned
}

func (a LegacyAdminCommand) AccessLevel() common.AccessLevel {
	return a.accessLevel
}

func (a LegacyAdminCommand) GetMethodName() string {
	return a.methodName
}

func (a LegacyAdminCommand) GetFn() CommandFunc {
	return a.fn
}
