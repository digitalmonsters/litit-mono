package apm_helper

import (
	"github.com/digitalmonsters/go-common/boilerplate"
	"github.com/rs/zerolog/log"
	"go.elastic.co/apm"
	"go.elastic.co/apm/transport"
	"os"
)

type ApmConfig struct {
	LogLevel    string `json:"log_level"`
	ServiceName string `json:"service_name"`
	ServerUrls  string `json:"server_urls"`
}

type mockLogger struct {
}

func (mockLogger) Debugf(format string, args ...interface{}) {
	//log.Trace().Msgf(format, args...)
}

func (mockLogger) Errorf(format string, args ...interface{}) {
	//log.Warn().Msgf(format, args)
}

func SetupApmLogging(config *boilerplate.ApmConfig) {
	if config != nil && len(config.ServerUrls) > 0 && len(config.ServiceName) > 0 {
		apm.DefaultTracer.Close()
		transport.Default = nil

		_ = os.Setenv("ELASTIC_APM_CENTRAL_CONFIG", "false")
		_ = os.Setenv("ELASTIC_APM_SERVICE_NAME", config.ServiceName)
		_ = os.Setenv("ELASTIC_APM_SERVER_URL", config.ServerUrls)
		_ = os.Setenv("ELASTIC_APM_LOG_LEVEL", config.LogLevel)
		_ = os.Setenv("ELASTIC_APM_ENVIRONMENT", boilerplate.GetCurrentEnvironment().ToString())

		tr, _ := transport.InitDefault()

		t, _ := apm.NewTracer(config.ServiceName, "")

		t.SetLogger(mockLogger{})
		apm.DefaultTracer = t
		t.Transport = tr

		log.Info().Msg("APM Client started")
	} else {
		_ = os.Setenv("ELASTIC_APM_ACTIVE", "false")

		t, _ := apm.NewTracer("nil", "nil")
		apm.DefaultTracer = t

		log.Warn().Msg("Ignoring APM Client as no configuration specified")
	}
}
