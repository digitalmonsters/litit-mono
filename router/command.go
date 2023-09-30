package router

import (
	"context"
	"strconv"
	"strings"

	"github.com/digitalmonsters/go-common/common"
	"github.com/digitalmonsters/go-common/error_codes"
	"github.com/digitalmonsters/go-common/rpc"
	"github.com/digitalmonsters/go-common/translation"
	"github.com/digitalmonsters/go-common/wrappers/auth_go"
	"github.com/pkg/errors"
	"github.com/valyala/fasthttp"
)

type ICommand interface {
	RequireIdentityValidation() bool
	AllowBanned() bool
	AccessLevel() common.AccessLevel
	GetMethodName() string
	GetFn() CommandFunc
	ForceLog() bool
	GetPath() string
	GetHttpMethod() string
	GetObj() string
	CanExecute(httpCtx *fasthttp.RequestCtx, ctx context.Context, auth auth_go.IAuthGoWrapper,
		userValidator UserExecutorValidator) (userId int64, isGuest bool, isBanned bool, language translation.Language, err *rpc.ExtendedLocalRpcError)
}

type CommandFunc func(request []byte, executionData MethodExecutionData) (interface{}, *error_codes.ErrorWithCode)

type Command struct {
	methodName                string
	forceLog                  bool
	fn                        CommandFunc
	requireIdentityValidation bool
	allowBanned               bool
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
		allowBanned:               false,
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

func (c Command) AllowBanned() bool {
	return c.allowBanned
}

func (c Command) GetHttpMethod() string {
	return "post"
}

func (c Command) GetFn() CommandFunc {
	return c.fn
}

func (c Command) CanExecute(httpCtx *fasthttp.RequestCtx, ctx context.Context, auth auth_go.IAuthGoWrapper, userValidator UserExecutorValidator) (int64, bool, bool, translation.Language, *rpc.ExtendedLocalRpcError) {
	return publicCanExecuteLogic(httpCtx, c.requireIdentityValidation, c.allowBanned, userValidator)
}

func publicCanExecuteLogic(ctx *fasthttp.RequestCtx, requireIdentityValidation bool, allowBanned bool, userValidator UserExecutorValidator) (int64, bool, bool, translation.Language, *rpc.ExtendedLocalRpcError) {
	var userId int64
	var isGuest bool
	var isBanned bool
	language := translation.DefaultUserLanguage

	if externalAuthValue := ctx.Request.Header.Peek("X-Ext-Authz-Check-Result"); strings.EqualFold(string(externalAuthValue), "allowed") {
		if userIdHead := ctx.Request.Header.Peek("User-Id"); len(userIdHead) > 0 {
			if userIdParsed, err := strconv.ParseInt(string(userIdHead), 10, 64); err != nil {
				err = errors.Wrapf(err, "can not parse str to int for user-id. input string %v", userIdHead)

				return 0, isGuest, isBanned, language, &rpc.ExtendedLocalRpcError{
					RpcError: rpc.RpcError{
						Code:        error_codes.InvalidJwtToken,
						Message:     err.Error(),
						Hostname:    hostName,
						ServiceName: hostName,
					},
					LocalHandlingError: err,
				}
			} else {
				userId = userIdParsed
			}
		}
	}

	if userId > 0 {
		usersResp, err := userValidator.Validate(userId, ctx)

		if err != nil {
			err = errors.Wrap(err, "can not get user info from auth service")

			return 0, isGuest, false, language, &rpc.ExtendedLocalRpcError{
				RpcError: rpc.RpcError{
					Code:        error_codes.GenericServerError,
					Message:     err.Error(),
					Hostname:    hostName,
					ServiceName: hostName,
				},
				LocalHandlingError: err,
			}
		}

		language = usersResp.Language

		isBanned = usersResp.BannedTill.Valid

		if !allowBanned && isBanned {
			err := errors.WithStack(errors.New("user banned"))

			return 0, isGuest, isBanned, language, &rpc.ExtendedLocalRpcError{
				RpcError: rpc.RpcError{
					Code:        error_codes.Forbidden,
					Message:     err.Error(),
					Hostname:    hostName,
					ServiceName: hostName,
				},
				LocalHandlingError: err,
			}
		}

		if usersResp.Deleted {
			err := errors.WithStack(errors.New("user deleted"))

			return 0, isGuest, isBanned, language, &rpc.ExtendedLocalRpcError{
				RpcError: rpc.RpcError{
					Code:        error_codes.Forbidden,
					Message:     err.Error(),
					Hostname:    hostName,
					ServiceName: hostName,
				},
				LocalHandlingError: err,
			}
		}

		if isGuestHeader := ctx.Request.Header.Peek("Is-Guest"); len(isGuestHeader) > 0 {
			if parsedIsGuest, err := strconv.ParseBool(string(isGuestHeader)); err != nil {
				err = errors.Wrapf(err, "can not parse str to int for is-guest. input string %v", isGuestHeader)

				return 0, usersResp.Guest, isBanned, language, &rpc.ExtendedLocalRpcError{
					RpcError: rpc.RpcError{
						Code:        error_codes.InvalidJwtToken,
						Message:     err.Error(),
						Hostname:    hostName,
						ServiceName: hostName,
					},
					LocalHandlingError: err,
				}
			} else {
				isGuest = parsedIsGuest
			}
		}
	}

	if requireIdentityValidation && userId <= 0 {
		err := errors.New("public method requires identity validation")

		return 0, isGuest, isBanned, language, &rpc.ExtendedLocalRpcError{
			RpcError: rpc.RpcError{
				Code:        error_codes.MissingJwtToken,
				Message:     err.Error(),
				Hostname:    hostName,
				ServiceName: hostName,
			},
			LocalHandlingError: err,
		}
	}

	ctx.Response.Header.Set("Access-Control-Allow-Credentials", "true")
	ctx.Response.Header.SetBytesV("Access-Control-Allow-Origin", ctx.Request.Header.Peek("Origin"))
	ctx.Response.Header.Set("Access-Control-Allow-Headers", "*")
	ctx.Response.Header.Set("Access-Control-Allow-Methods", "POST, GET, OPTIONS, PATCH, DELETE")

	return userId, isGuest, isBanned, language, nil
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
