package router

import (
	"github.com/digitalmonsters/go-common/common"
	"github.com/digitalmonsters/go-common/rpc"
	"github.com/digitalmonsters/go-common/wrappers/auth_go"
	"github.com/valyala/fasthttp"
	"go.elastic.co/apm"
	"strings"
)

type ServiceCommand struct {
	methodName                string
	accessLevel               common.AccessLevel
	forceLog                  bool
	fn                        CommandFunc
	requireIdentityValidation bool
	obj                       string
}

func NewServiceCommand(methodName string, fn CommandFunc, forceLog bool) ICommand {
	return &ServiceCommand{
		methodName:                strings.ToLower(methodName),
		accessLevel:               common.AccessLevelPublic,
		forceLog:                  forceLog,
		fn:                        fn,
		obj:                       "",
		requireIdentityValidation: false,
	}
}

func (a ServiceCommand) CanExecute(ctx *fasthttp.RequestCtx, apmTransaction *apm.Transaction, auth auth_go.IAuthGoWrapper) (int64, bool, *rpc.RpcError) {
	return 0, false, nil
}

func (a ServiceCommand) ForceLog() bool {
	return a.forceLog
}

func (a ServiceCommand) GetObj() string {
	return a.obj
}

func (a ServiceCommand) RequireIdentityValidation() bool {
	return a.requireIdentityValidation
}

func (a ServiceCommand) AccessLevel() common.AccessLevel {
	return a.accessLevel
}

func (a ServiceCommand) GetMethodName() string {
	return a.methodName
}

func (a ServiceCommand) GetFn() CommandFunc {
	return a.fn
}

func (a ServiceCommand) GetPath() string {
	return a.GetMethodName()
}

func (a ServiceCommand) GetHttpMethod() string {
	return "post"
}
