package router

import (
	"fmt"
	"github.com/digitalmonsters/go-common/common"
	"github.com/digitalmonsters/go-common/error_codes"
	"github.com/digitalmonsters/go-common/rpc"
	"github.com/digitalmonsters/go-common/wrappers/auth_go"
	"github.com/valyala/fasthttp"
	"go.elastic.co/apm"
	"strconv"
	"strings"
)

type AdminCommand struct {
	methodName                string
	accessLevel               common.AccessLevel
	forceLog                  bool
	fn                        CommandFunc
	requireIdentityValidation bool
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
	}
}

func (a AdminCommand) CanExecute(ctx *fasthttp.RequestCtx, apmTransaction *apm.Transaction, auth auth_go.IAuthGoWrapper) (int64, *rpc.RpcError) {
	if externalAuthValue := ctx.Request.Header.Peek("X-Ext-Authz-Check-Result"); strings.EqualFold(string(externalAuthValue), "allowed") {
		if userIdHead := ctx.Request.Header.Peek("Admin-Id"); len(userIdHead) > 0 {
			if userIdParsed, err := strconv.ParseInt(string(userIdHead), 10, 64); err != nil {
				return 0, &rpc.RpcError{
					Code:        error_codes.InvalidJwtToken,
					Message:     fmt.Sprintf("can not parse str to int for admin-id. input string %v. [%v]", userIdHead, err.Error()),
					Hostname:    hostName,
					ServiceName: hostName,
				}
			} else {
				if a.accessLevel == common.AccessLevelPublic {
					return userIdParsed, nil
				}
				
				ch := <-auth.CheckAdminPermissions(userIdParsed, a.obj, a.AccessLevel(), apmTransaction, false)

				if ch.Error != nil {
					return 0, ch.Error
				}

				if ch.Resp.HasAccess {
					return userIdParsed, nil
				}

				return 0, &rpc.RpcError{
					Code:        error_codes.InvalidJwtToken,
					Message:     "admin user does not have access to this method",
					Hostname:    hostName,
					ServiceName: hostName,
				}
			}
		}
	}

	return 0, &rpc.RpcError{
		Code:        error_codes.MissingJwtToken,
		Message:     "new admin method requires new admin authorization header",
		Hostname:    hostName,
		ServiceName: hostName,
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

func (a AdminCommand) AccessLevel() common.AccessLevel {
	return a.accessLevel
}

func (a AdminCommand) GetMethodName() string {
	return a.methodName
}

func (a AdminCommand) GetFn() CommandFunc {
	return a.fn
}
