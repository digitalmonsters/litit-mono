package kafka_listener

import (
	"github.com/digitalmonsters/go-common/apm_helper"
	"github.com/pkg/errors"
	"github.com/segmentio/kafka-go"
)

type CommandFunc func(executionData ExecutionData, request ...kafka.Message) []kafka.Message

type Command struct {
	fancyName string
	forceLog  bool
	fn        CommandFunc
}

func (c Command) GetFancyName() string {
	return c.fancyName
}
func (c *Command) ForceLog() bool {
	return c.forceLog
}

func NewCommand(fancyName string, fn CommandFunc, forceLog bool) *Command {
	cmd := &Command{
		fancyName: fancyName,
		forceLog:  forceLog,
		fn:        fn,
	}

	return cmd
}

func (c *Command) Execute(executionData ExecutionData, request ...kafka.Message) (successfullyProcessed []kafka.Message) {
	defer func() {
		if er := recover(); er != nil {
			var err error

			switch x := er.(type) {
			case string:
				err = errors.New(x)
			case error:
				err = errors.WithStack(x)
			default:
				// Fallback err (per specs, error strings should be lowercase w/o punctuation
				err = errors.New("unknown panic")
			}

			apm_helper.LogError(err, executionData.Context)
		}
	}()

	return c.fn(executionData, request...)
}
