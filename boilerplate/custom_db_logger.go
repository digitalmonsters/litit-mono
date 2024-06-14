package boilerplate

import (
	"context"
	"github.com/pkg/errors"
	"github.com/rs/zerolog"
	gorm_logger "gorm.io/gorm/logger"
	"gorm.io/gorm/utils"
	"time"
)

type logger struct {
	SlowThreshold         time.Duration
	SkipErrRecordNotFound bool
	HasStackTrace         bool
}

type CustomLoggerConfig struct {
	SlowThreshold         time.Duration
	SkipErrRecordNotFound bool
	HasStackTrace         bool
}

func NewDbLogger(cfg CustomLoggerConfig) *logger {
	return &logger{
		SkipErrRecordNotFound: cfg.SkipErrRecordNotFound,
		SlowThreshold:         cfg.SlowThreshold,
		HasStackTrace:         cfg.HasStackTrace,
	}
}

func (l *logger) LogMode(level gorm_logger.LogLevel) gorm_logger.Interface {
	return l
}

func (l *logger) Info(ctx context.Context, s string, args ...interface{}) {
	zerolog.Ctx(ctx).Info().Msgf(s, args...)
}

func (l *logger) Warn(ctx context.Context, s string, args ...interface{}) {
	zerolog.Ctx(ctx).Warn().Msgf(s, args...)
}

func (l *logger) Error(ctx context.Context, s string, args ...interface{}) {
	zerolog.Ctx(ctx).Error().Msgf(s, args...)
}

func (l *logger) Trace(ctx context.Context, begin time.Time, fc func() (string, int64), err error) {
	elapsed := time.Since(begin)
	sql, _ := fc()
	fields := map[string]interface{}{
		"sql":      sql,
		"duration": elapsed,
	}
	if l.HasStackTrace {
		fields["stack_trace"] = utils.FileWithLineNum()
	}

	switch {
	case err != nil && (!errors.Is(err, gorm_logger.ErrRecordNotFound) || !l.SkipErrRecordNotFound):
		zerolog.Ctx(ctx).Error().Err(err).Fields(fields).Msg("[GORM] query error")
	case elapsed > l.SlowThreshold && l.SlowThreshold != 0:
		zerolog.Ctx(ctx).Warn().Fields(fields).Msgf("[GORM] slow query")
	default:
		zerolog.Ctx(ctx).Debug().Fields(fields).Msgf("[GORM] query")
	}
}
