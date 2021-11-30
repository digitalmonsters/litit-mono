package router

import (
	"github.com/digitalmonsters/go-common/common"
	"github.com/digitalmonsters/go-common/error_codes"
	"strings"
)

type ICommand interface {
	RequireIdentityValidation() bool
	AccessLevel() common.AccessLevel
	GetFn() CommandFunc
}

type CommandFunc func(request []byte, executionData MethodExecutionData) (interface{}, *error_codes.ErrorWithCode)

type Command struct {
	methodName                string
	accessLevel               common.AccessLevel
	forceLog                  bool
	fn                        CommandFunc
	requireIdentityValidation bool
}

func (c *Command) Execute(request []byte, data MethodExecutionData) (interface{}, *error_codes.ErrorWithCode) {
	return c.fn(request, data)
}

func NewCommand(methodName string, fn CommandFunc, accessLevel common.AccessLevel, forceLog bool, requireIdentityValidation bool) *Command {
	return &Command{
		methodName:                strings.ToLower(methodName),
		accessLevel:               accessLevel,
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

func (c Command) AccessLevel() common.AccessLevel {
	return c.accessLevel
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

func (c Command) ForceLog() bool {
	if c.forceLog {
		return true
	}

	if c.AccessLevel() > common.AccessLevelRead {
		return true
	}

	return false
}