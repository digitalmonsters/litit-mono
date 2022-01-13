package router

import (
	"context"
	"fmt"
	"github.com/pkg/errors"
	"github.com/prometheus/client_golang/prometheus/promhttp"
	"go.elastic.co/apm"
)

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

	promhttp.Handler()
	c.commands[command.GetMethodName()] = command

	return nil
}

func (c *CommandExecutor) GetCommand(methodName string) (*Command, error) {
	if v, ok := c.commands[methodName]; ok {
		return v, nil
	}

	return nil, errors.New("command not found")
}
