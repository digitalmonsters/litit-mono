package boilerplate_testing

import (
	"fmt"
	"runtime"
	"runtime/debug"
	"strings"
)

func SetMockAppConfig[T any](mockAppConfigs map[string]T, mockAppConfig T) {
	callerName := getCallerName(true, 3)
	if len(callerName) == 0 {
		panic("cannot get caller info")
	}

	mockAppConfigs[callerName] = mockAppConfig
}

func getCallerName(last bool, skip int) string {
	pc, _, _, ok := runtime.Caller(skip)
	if !ok {
		panic("cannot get caller info")
	}

	details := runtime.FuncForPC(pc)

	if !last {
		return details.Name()
	}

	callerNameSplit := strings.Split(details.Name(), ".")
	lastCallerName := strings.TrimSpace(callerNameSplit[len(callerNameSplit)-1])

	return lastCallerName
}

func GetMockAppConfig[T any](mockAppConfigs map[string]T) T {
	stack := string(debug.Stack())

	for k, v := range mockAppConfigs {
		if !strings.Contains(stack, k) {
			continue
		}

		return v
	}

	panic(fmt.Sprintf("missing config for stack: %v", stack))
}
