package router

import (
	"context"
	"fmt"
	"github.com/digitalmonsters/go-common/common"
	"github.com/digitalmonsters/go-common/error_codes"
	"github.com/pkg/errors"
	"go.elastic.co/apm"
	"strings"
)

type CommandFunc func(request []byte, executionData MethodExecutionData) (interface{}, *error_codes.ErrorWithCode)

type MethodExecutionData struct {
	ApmTransaction *apm.Transaction
	Context        context.Context
	UserId         int64
	getUserValueFn func(key string) interface{}
}

func (m MethodExecutionData) GetUserValue(key string) interface{} {
	if m.getUserValueFn != nil {
		return m.getUserValueFn(key)
	}

	return nil
}

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

func (c Command) ForceLog() bool {
	if c.forceLog {
		return true
	}

	if c.AccessLevel() > common.AccessLevelRead {
		return true
	}

	return false
}

type CommandExecutor struct {
	commands map[string]*Command
}

func NewCommandExecutor() *CommandExecutor {
	return &CommandExecutor{commands: map[string]*Command{}}
}

func (c *CommandExecutor) AddCommand(command *Command) error {
	if _, ok := c.commands[command.GetMethodName()]; ok {
		return errors.New(fmt.Sprintf("command with same name already registered [%v]", command.GetMethodName()))
	}

	c.commands[command.GetMethodName()] = command

	return nil
}

func (c *CommandExecutor) GetCommand(methodName string) (*Command, error) {
	if v, ok := c.commands[methodName]; ok {
		return v, nil
	}

	return nil, errors.New("command not found")
}
