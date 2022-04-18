package router

import (
	"context"
	"fmt"
	"github.com/digitalmonsters/go-common/common"
	"github.com/digitalmonsters/go-common/error_codes"
	"github.com/digitalmonsters/go-common/rpc"
	"github.com/digitalmonsters/go-common/wrappers/auth_go"
	"github.com/valyala/fasthttp"
	"strconv"
	"strings"
)

type ICommand interface {
	RequireIdentityValidation() bool
	AccessLevel() common.AccessLevel
	GetMethodName() string
	GetFn() CommandFunc
	ForceLog() bool
	GetPath() string
	GetHttpMethod() string
	GetObj() string
	CanExecute(httpCtx *fasthttp.RequestCtx, ctx context.Context, auth auth_go.IAuthGoWrapper) (userId int64, isGuest bool, err *rpc.RpcError)
}

type CommandFunc func(request []byte, executionData MethodExecutionData) (interface{}, *error_codes.ErrorWithCode)

type Command struct {
	methodName                string
	forceLog                  bool
	fn                        CommandFunc
	requireIdentityValidation bool
}

func (c *Command) Execute(request []byte, data MethodExecutionData) (interface{}, *error_codes.ErrorWithCode) {
	return c.fn(request, data)
}

func NewCommand(methodName string, fn CommandFunc, forceLog bool, requireIdentityValidation bool) ICommand {
	return &Command{
		methodName:                strings.ToLower(methodName),
		forceLog:                  forceLog,
		fn:                        fn,
		requireIdentityValidation: requireIdentityValidation,
	}
}

func (c Command) GetMethodName() string {
	return c.methodName
}

func (c Command) GetPath() string { // for rest
	return c.GetMethodName()
}

func (c Command) GetObj() string {
	return ""
}

func (c Command) AccessLevel() common.AccessLevel {
	return common.AccessLevelPublic
}

func (c Command) RequireIdentityValidation() bool {
	return c.requireIdentityValidation
}

func (c Command) GetHttpMethod() string {
	return "post"
}

func (c Command) GetFn() CommandFunc {
	return c.fn
}

func (c Command) CanExecute(httpCtx *fasthttp.RequestCtx, ctx context.Context, auth auth_go.IAuthGoWrapper) (int64, bool, *rpc.RpcError) {
	return publicCanExecuteLogic(httpCtx, c.requireIdentityValidation)
}

func publicCanExecuteLogic(ctx *fasthttp.RequestCtx, requireIdentityValidation bool) (int64, bool, *rpc.RpcError) {
	var userId int64
	var isGuest bool

	if externalAuthValue := ctx.Request.Header.Peek("X-Ext-Authz-Check-Result"); strings.EqualFold(string(externalAuthValue), "allowed") {
		if userIdHead := ctx.Request.Header.Peek("User-Id"); len(userIdHead) > 0 {
			if userIdParsed, err := strconv.ParseInt(string(userIdHead), 10, 64); err != nil {
				return 0, isGuest, &rpc.RpcError{
					Code:        error_codes.InvalidJwtToken,
					Message:     fmt.Sprintf("can not parse str to int for user-id. input string %v. [%v]", userIdHead, err.Error()),
					Hostname:    hostName,
					ServiceName: hostName,
				}
			} else {
				userId = userIdParsed
			}
		}
	}

	if userId > 0 {
		if isGuestHeader := ctx.Request.Header.Peek("Is-Guest"); len(isGuestHeader) > 0 {
			if parsedIsGuest, err := strconv.ParseBool(string(isGuestHeader)); err != nil {
				return 0, isGuest, &rpc.RpcError{
					Code:        error_codes.InvalidJwtToken,
					Message:     fmt.Sprintf("can not parse str to int for is-guest. input string %v. [%v]", isGuestHeader, err.Error()),
					Hostname:    hostName,
					ServiceName: hostName,
				}
			} else {
				isGuest = parsedIsGuest
			}
		}
	}

	if requireIdentityValidation && userId <= 0 {
		return 0, isGuest, &rpc.RpcError{
			Code:        error_codes.MissingJwtToken,
			Message:     "public method requires identity validation",
			Hostname:    hostName,
			ServiceName: hostName,
		}
	}

	return userId, isGuest, nil
}

func (c Command) ForceLog() bool {
	if c.forceLog {
		return true
	}

	if c.AccessLevel() > common.AccessLevelRead {
		return true
	}

	return false
}
