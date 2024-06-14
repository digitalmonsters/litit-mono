package boilerplate

import (
	"fmt"
	"github.com/RichardKnop/machinery/v1"
	"github.com/RichardKnop/machinery/v1/brokers/eager"
	"github.com/RichardKnop/machinery/v1/config"
	eagerlock "github.com/RichardKnop/machinery/v1/locks/eager"
	lockiface "github.com/RichardKnop/machinery/v1/locks/iface"
	redislock "github.com/RichardKnop/machinery/v1/locks/redis"
	"github.com/pkg/errors"
	"github.com/rs/zerolog"
	"strconv"
	"strings"
)

// NewServer creates Server instance
func NewServer(cnf *config.Config) (*machinery.Server, error) {
	broker, err := machinery.BrokerFactory(cnf)
	if err != nil {
		return nil, err
	}

	// Backend is optional so we ignore the error
	backend, _ := machinery.BackendFactory(cnf)

	// Init lock
	lock, err := LockFactory(cnf)
	if err != nil {
		return nil, err
	}

	srv := machinery.NewServerWithBrokerBackendLock(cnf, broker, backend, lock)

	// init for eager-mode
	eager, ok := broker.(eager.Mode)
	if ok {
		// we don't have to call worker.Launch in eager mode
		eager.AssignWorker(srv.NewWorker("eager", 0))
	}
	var logger = zerolog.DefaultContextLogger
	SetMachineryZeroLogLogger(*logger)
	return srv, nil
}

// LockFactory creates a new object of iface.Lock
// Currently supported lock is redis
func LockFactory(cnf *config.Config) (lockiface.Lock, error) {
	if strings.HasPrefix(cnf.Lock, "eager") {
		return eagerlock.New(), nil
	}
	if cnf.TLSConfig != nil {
		if strings.HasPrefix(cnf.Lock, "rediss://") {
			parts := strings.Split(cnf.Lock, "rediss://")
			if len(parts) != 2 {
				return nil, fmt.Errorf(
					"Redis broker connection string should be in format rediss://host:port, instead got %s",
					cnf.Lock,
				)
			}

			if strings.Contains(parts[1], ",") {
				return nil, errors.New(", is not allowed in locks")
			}

			dbNumber := 0
			host := parts[1]

			sp := strings.Split(host, "/")

			if len(sp) == 2 {
				host = sp[0] // thi is host

				if dbNum, err := strconv.Atoi(sp[1]); err != nil {
					return nil, errors.WithStack(err)
				} else {
					dbNumber = dbNum
				}
			}

			locks := strings.Split(host, ",")
			return New(cnf, locks, dbNumber, 3), nil
		}
	} else {
		if strings.HasPrefix(cnf.Lock, "redis://") {
			parts := strings.Split(cnf.Lock, "redis://")
			if len(parts) != 2 {
				return nil, fmt.Errorf(
					"Redis broker connection string should be in format redis://host:port, instead got %s",
					cnf.Lock,
				)
			}

			if strings.Contains(parts[1], ",") {
				return nil, errors.New(", is not allowed in locks")
			}

			dbNumber := 0
			host := parts[1]

			sp := strings.Split(host, "/")

			if len(sp) == 2 {
				host = sp[0] // thi is host

				if dbNum, err := strconv.Atoi(sp[1]); err != nil {
					return nil, errors.WithStack(err)
				} else {
					dbNumber = dbNum
				}
			}

			locks := strings.Split(host, ",")
			return redislock.New(cnf, locks, dbNumber, 3), nil
		}
	}

	// Lock is required for periodic tasks to work, therefor return in memory lock in case none is configured
	return eagerlock.New(), nil
}
