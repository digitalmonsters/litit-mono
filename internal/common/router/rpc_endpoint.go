package router

import (
	"github.com/rs/zerolog/log"
	"reflect"
)

type IRpcEndpoint interface {
	RegisterRpcCommand(command ICommand) error
	GetCommand(methodName string) (ICommand, error)
	GetRegisteredCommands() []ICommand
}

type rpcEndpointPublic struct {
	executor *CommandExecutor
}

func newRpcEndpointPublic() *rpcEndpointPublic {
	return &rpcEndpointPublic{executor: NewCommandExecutor()}
}

func (r rpcEndpointPublic) GetRegisteredCommands() []ICommand {
	var commands []ICommand

	for _, c := range r.executor.commands {
		commands = append(commands, c)
	}

	return commands
}

func (r *rpcEndpointPublic) RegisterRpcCommand(command ICommand) error {
	cmd := reflect.TypeOf(command)

	if cmd.String() != "*router.Command" && cmd.String() != "*router.LegacyAdminCommand" {
		log.Fatal().Msg("only *router.Command or *router.LegacyAdminCommand can be registered with RegisterRpcCommand")
	}

	return r.executor.AddCommand(command)
}

func (r rpcEndpointPublic) GetCommand(methodName string) (ICommand, error) {
	return r.executor.GetCommand(methodName)
}

type rpcEndpointAdmin struct {
	executor *CommandExecutor
}

func newRpcEndpointAdmin() *rpcEndpointAdmin {
	return &rpcEndpointAdmin{executor: NewCommandExecutor()}
}

func (r *rpcEndpointAdmin) RegisterRpcCommand(command ICommand) error {
	if _, ok := command.(*AdminCommand); !ok {
		log.Fatal().Msg("only AdminCommand can be registered with RegisterRpcCommand")
	}

	return r.executor.AddCommand(command)
}

func (r rpcEndpointAdmin) GetRegisteredCommands() []ICommand {
	var commands []ICommand

	for _, c := range r.executor.commands {
		commands = append(commands, c)
	}

	return commands
}

func (r rpcEndpointAdmin) GetCommand(methodName string) (ICommand, error) {
	return r.executor.GetCommand(methodName)
}

type rpcEndpointService struct {
	executor *CommandExecutor
}

func newRpcEndpointService() *rpcEndpointService {
	return &rpcEndpointService{executor: NewCommandExecutor()}
}

func (r rpcEndpointService) GetRegisteredCommands() []ICommand {
	var commands []ICommand

	for _, c := range r.executor.commands {
		commands = append(commands, c)
	}

	return commands
}

func (r *rpcEndpointService) RegisterRpcCommand(command ICommand) error {
	if _, ok := command.(*ServiceCommand); !ok {
		log.Fatal().Msg("only ServiceCommand can be registered with RegisterRpcCommand")
	}

	return r.executor.AddCommand(command)
}

func (r rpcEndpointService) GetCommand(methodName string) (ICommand, error) {
	return r.executor.GetCommand(methodName)
}
