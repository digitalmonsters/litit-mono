package boilerplate

import (
	"fmt"
	"github.com/RichardKnop/machinery/v1/log"
	"github.com/rs/zerolog"
)

func SetMachineryZeroLogLogger(logger zerolog.Logger) {
	log.SetInfo(NewMachineryZeroLogLogger(logger, zerolog.InfoLevel))
	log.SetDebug(NewMachineryZeroLogLogger(logger, zerolog.DebugLevel))
	log.SetError(NewMachineryZeroLogLogger(logger, zerolog.ErrorLevel))
	log.SetWarning(NewMachineryZeroLogLogger(logger, zerolog.WarnLevel))
	log.SetFatal(NewMachineryZeroLogLogger(logger, zerolog.FatalLevel))
}

func NewMachineryZeroLogLogger(logger zerolog.Logger, level zerolog.Level) MachineryZeroLogLogger {
	return MachineryZeroLogLogger{
		logger: logger,
		level:  level,
	}
}

type MachineryZeroLogLogger struct {
	logger zerolog.Logger
	level  zerolog.Level
}

func (m MachineryZeroLogLogger) Print(i ...interface{}) {
	for _, message := range i {
		m.logger.WithLevel(m.level).Msg(fmt.Sprint(message))
	}
}

func (m MachineryZeroLogLogger) Printf(s string, i ...interface{}) {
	m.logger.WithLevel(m.level).Msgf(s, i...)
}

func (m MachineryZeroLogLogger) Println(i ...interface{}) {
	for _, message := range i {
		m.logger.WithLevel(m.level).Msg(fmt.Sprint(message))
	}
}

func (m MachineryZeroLogLogger) Fatal(i ...interface{}) {
	for _, message := range i {
		m.logger.WithLevel(m.level).Msg(fmt.Sprint(message))
	}
}

func (m MachineryZeroLogLogger) Fatalf(s string, i ...interface{}) {
	m.logger.WithLevel(m.level).Msgf(s, i...)
}

func (m MachineryZeroLogLogger) Fatalln(i ...interface{}) {
	for _, message := range i {
		m.logger.WithLevel(m.level).Msg(fmt.Sprint(message))
	}
}

func (m MachineryZeroLogLogger) Panic(i ...interface{}) {
	for _, message := range i {
		m.logger.WithLevel(m.level).Msg(fmt.Sprint(message))
	}
}

func (m MachineryZeroLogLogger) Panicf(s string, i ...interface{}) {
	m.logger.WithLevel(m.level).Msgf(s, i...)
}

func (m MachineryZeroLogLogger) Panicln(i ...interface{}) {
	for _, message := range i {
		m.logger.WithLevel(m.level).Msg(fmt.Sprint(message))
	}
}
