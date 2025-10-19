package utils

import (
	"crypto/md5"
	"encoding/json"
	"fmt"
	"github.com/RichardKnop/machinery/v1"
	"github.com/RichardKnop/machinery/v1/backends/result"
	"github.com/RichardKnop/machinery/v1/tasks"
	"github.com/pkg/errors"
)

var TaskIsInProgress = errors.New("task is in progress")

func RegisterPeriodicTask(server *machinery.Server, spec string, name string, args []tasks.Arg, checkDuplicates bool) error {
	orchestratorTaskName := fmt.Sprintf("orchestrator:%v", name)

	if err := server.RegisterTask(orchestratorTaskName, func() error {
		_, err := SendTask(server, name, args, checkDuplicates)

		return errors.WithStack(err)
	}); err != nil {
		return errors.WithStack(err)
	}

	if err := server.RegisterPeriodicTask(spec, orchestratorTaskName, &tasks.Signature{
		Name:       orchestratorTaskName,
		RetryCount: 3,
	}); err != nil {
		return errors.WithStack(err)
	}

	return nil
}

func SendTask(server *machinery.Server, taskName string, args []tasks.Arg, checkDuplicates bool) (*result.AsyncResult, error) {
	if server == nil {
		return nil, errors.WithStack(errors.New("machinery is null"))
	}

	uuid := taskName

	if !checkDuplicates {
		uuid = "" // ignore, as machinery will generate it own id
	}

	if len(args) > 0 {
		b, err := json.Marshal(args)

		if err != nil {
			return nil, errors.WithStack(err)
		}

		uuid = fmt.Sprintf("%v_%x", uuid, md5.Sum(b))
	}

	if checkDuplicates {
		if err := checkForDuplicates(server, uuid); err != nil {
			return nil, err
		}
	}

	signature := &tasks.Signature{
		Name:       taskName,
		UUID:       uuid,
		Args:       args,
		RetryCount: 3,
	}

	return server.SendTask(signature)
}

func SendChain(server *machinery.Server, tasksDict map[string][]tasks.Arg, checkDuplicates bool) (*result.ChainAsyncResult, error) {
	if server == nil {
		return nil, errors.WithStack(errors.New("machinery is null"))
	}

	var tasksSignatures []*tasks.Signature

	for taskName, args := range tasksDict {
		uuid := taskName

		if !checkDuplicates {
			uuid = "" // ignore, as machinery will generate it own id
		}

		if len(args) > 0 {
			b, err := json.Marshal(args)

			if err != nil {
				return nil, errors.WithStack(err)
			}

			uuid = fmt.Sprintf("%v_%x", uuid, md5.Sum(b))
		}

		if checkDuplicates {
			if err := checkForDuplicates(server, uuid); err != nil {
				return nil, err
			}
		}

		tasksSignatures = append(tasksSignatures, &tasks.Signature{
			Name:       taskName,
			UUID:       uuid,
			Args:       args,
			RetryCount: 3,
		})

	}

	chain, err := tasks.NewChain(tasksSignatures...)

	if err != nil {
		return nil, err
	}

	return server.SendChain(chain)
}

func checkForDuplicates(server *machinery.Server, uuid string) error {

	state, _ := server.GetBackend().GetState(uuid)

	if state != nil && (state.State == "PENDING" || state.State == "RECEIVED" || state.State == "STARTED") {
		return TaskIsInProgress
	}

	return nil

}
