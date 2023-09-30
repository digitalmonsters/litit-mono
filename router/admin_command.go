package router

import (
	"context"
	"strconv"
	"strings"

	"github.com/digitalmonsters/go-common/boilerplate"
	"github.com/digitalmonsters/go-common/common"
	"github.com/digitalmonsters/go-common/error_codes"
	"github.com/digitalmonsters/go-common/rpc"
	"github.com/digitalmonsters/go-common/translation"
	"github.com/digitalmonsters/go-common/wrappers/auth"
	"github.com/digitalmonsters/go-common/wrappers/auth_go"
	"github.com/pkg/errors"
	"github.com/valyala/fasthttp"
	"go.elastic.co/apm"
)

type AdminCommand struct {
	methodName                string
	accessLevel               common.AccessLevel
	forceLog                  bool
	fn                        CommandFunc
	requireIdentityValidation bool
	allowBanned               bool
	obj                       string
}

func NewAdminCommand(methodName string, fn CommandFunc, accessLevel common.AccessLevel, rbacObj string) ICommand {
	return &AdminCommand{
		methodName:                strings.ToLower(methodName),
		accessLevel:               accessLevel,
		forceLog:                  true,
		fn:                        fn,
		obj:                       rbacObj,
		requireIdentityValidation: true,
		allowBanned:               false,
	}
}

func (a AdminCommand) CanExecute(httpCtx *fasthttp.RequestCtx, ctx context.Context, authWrapper auth_go.IAuthGoWrapper, userValidator UserExecutorValidator) (int64, bool, bool, translation.Language, *rpc.ExtendedLocalRpcError) {
	currentUserId := int64(0)
	language := translation.DefaultUserLanguage
	httpCtx.Response.Header.Set("Access-Control-Allow-Credentials", "true")
	httpCtx.Response.Header.SetBytesV("Access-Control-Allow-Origin", httpCtx.Request.Header.Peek("Origin"))
	httpCtx.Response.Header.Set("Access-Control-Allow-Headers", "*")
	httpCtx.Response.Header.Set("Access-Control-Allow-Methods", "POST, GET, OPTIONS, PATCH, DELETE")

	if externalAuthValue := httpCtx.Request.Header.Peek("X-Ext-Authz-Check-Result"); strings.EqualFold(string(externalAuthValue), "allowed") { // external auth
		if userIdHead := httpCtx.Request.Header.Peek("Admin-Id"); len(userIdHead) > 0 {
			if userIdParsed, err := strconv.ParseInt(string(userIdHead), 10, 64); err != nil {
				err = errors.Wrapf(err, "can not parse str to int for admin-id. input string %v.", userIdHead)
				return 0, false, false, language, &rpc.ExtendedLocalRpcError{
					RpcError: rpc.RpcError{
						Code:        error_codes.InvalidJwtToken,
						Message:     err.Error(),
						Hostname:    hostName,
						ServiceName: hostName,
					},
					LocalHandlingError: err,
				}
			} else {
				currentUserId = userIdParsed
			}
		}
	}

	if currentUserId == 0 { // TODO temporary remove after fix for istio fallback auth
		if jwtAuthData := httpCtx.Request.Header.Peek("Authorization-Admin"); len(jwtAuthData) > 0 {
			forwardAuthWrapper := auth.NewAuthWrapper(boilerplate.WrapperConfig{
				ApiUrl:     "http://forward-auth",
				TimeoutSec: 3,
			})

			resp := <-forwardAuthWrapper.ParseNewAdminToken(string(jwtAuthData), false, apm.TransactionFromContext(ctx), false)

			if resp.Error != nil {
				return 0, false, false, language, &rpc.ExtendedLocalRpcError{
					RpcError: *resp.Error,
				}
			}

			currentUserId = resp.Resp.UserId
		}
	}

	if currentUserId == 0 {
		err := errors.New("new admin method requires new admin authorization header")

		return 0, false, false, language, &rpc.ExtendedLocalRpcError{
			RpcError: rpc.RpcError{
				Code:        error_codes.MissingJwtToken,
				Message:     err.Error(),
				Hostname:    hostName,
				ServiceName: hostName,
			},
			LocalHandlingError: err,
		}
	}

	if a.accessLevel == common.AccessLevelPublic {
		return currentUserId, false, false, language, nil
	}

	ch := <-authWrapper.CheckAdminPermissions(currentUserId, a.obj, apm.TransactionFromContext(ctx), false)

	if ch.Error != nil {
		return 0, false, false, language, &rpc.ExtendedLocalRpcError{
			RpcError: *ch.Error,
		}
	}

	if ch.Resp.HasAccess {
		return currentUserId, false, false, language, nil
	}

	err := errors.New("admin user does not have access to this method")

	return 0, false, false, language, &rpc.ExtendedLocalRpcError{
		RpcError: rpc.RpcError{
			Code:        error_codes.InvalidJwtToken,
			Message:     err.Error(),
			Hostname:    hostName,
			ServiceName: hostName,
		},
		LocalHandlingError: err,
	}
}

func (a AdminCommand) GetPath() string {
	return a.GetMethodName()
}

func (a AdminCommand) GetHttpMethod() string {
	return "post"
}

func (a AdminCommand) ForceLog() bool {
	return a.forceLog
}

func (a AdminCommand) GetObj() string {
	return a.obj
}

func (a AdminCommand) RequireIdentityValidation() bool {
	return a.requireIdentityValidation
}

func (a AdminCommand) AllowBanned() bool {
	return a.allowBanned
}

func (a AdminCommand) AccessLevel() common.AccessLevel {
	return a.accessLevel
}

func (a AdminCommand) GetMethodName() string {
	return a.methodName
}

func (a AdminCommand) GetFn() CommandFunc {
	return a.fn
}
