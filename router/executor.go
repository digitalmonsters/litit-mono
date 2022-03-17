package router

import (
	"context"
	"fmt"
	"github.com/pkg/errors"
	"go.elastic.co/apm"
)

type MethodExecutionData struct {
	ApmTransaction *apm.Transaction
	Context        context.Context
	UserId         int64
	IsGuest        bool
	UserIp         string
	getUserValueFn func(key string) interface{}
}

func (m MethodExecutionData) GetUserValue(key string) interface{} {
	if m.getUserValueFn != nil {
		return m.getUserValueFn(key)
	}

	return nil
}

type CommandExecutor struct {
	commands map[string]ICommand
}

func NewCommandExecutor() *CommandExecutor {
	return &CommandExecutor{commands: map[string]ICommand{}}
}

func (c *CommandExecutor) AddCommand(command ICommand) error {
	if _, ok := c.commands[command.GetMethodName()]; ok {
		return errors.New(fmt.Sprintf("command with same name already registered [%v]", command.GetMethodName()))
	}

	c.commands[command.GetMethodName()] = command

	return nil
}

func (c *CommandExecutor) GetCommand(methodName string) (ICommand, error) {
	if v, ok := c.commands[methodName]; ok {
		return v, nil
	}

	return nil, errors.New("command not found")
}
