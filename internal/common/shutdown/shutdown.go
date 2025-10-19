package shutdown

import (
	"github.com/digitalmonsters/go-common/boilerplate"
	"os"
	"strconv"
	"time"
)

func RunGracefulShutdown(minShutdownTimeSeconds int, callers []func() error) {
	startAt := time.Now()

	for _, c := range callers {
		_ = c()
	}

	diff := time.Since(startAt).Seconds()

	if diff < float64(minShutdownTimeSeconds) {
		time.Sleep(time.Duration(float64(minShutdownTimeSeconds)-diff) * time.Second)
	}
}

func GetGracefulSleepDuration() int {
	currentSec := os.Getenv("APP_GRACEFUL_SHUTDOWN_SEC")
	// trigger build latest latest
	if len(currentSec) > 0 {
		if v, err := strconv.Atoi(currentSec); err == nil {
			return v
		}
	}

	if boilerplate.GetCurrentEnvironment() == boilerplate.Dev || boilerplate.GetCurrentEnvironment() == boilerplate.Ci {
		return 0
	}

	return 20
}
