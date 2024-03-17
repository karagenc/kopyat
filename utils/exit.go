package utils

import (
	"os"
	"sync"
)

type ExitFunc func()

var (
	exitFuncs   []ExitFunc
	exitFuncsMu sync.Mutex
)

func AddExitHandler(f ExitFunc) {
	exitFuncsMu.Lock()
	exitFuncs = append(exitFuncs, f)
	exitFuncsMu.Unlock()
}

func ExitHandlers() []ExitFunc {
	exitFuncsMu.Lock()
	exitFuncs := exitFuncs
	exitFuncsMu.Unlock()
	return exitFuncs
}

func RunExitHandlers() {
	exitFuncsMu.Lock()
	exitFuncs := exitFuncs
	exitFuncsMu.Unlock()
	for _, f := range exitFuncs {
		f()
	}
}

func Exit(code int) {
	exitFuncsMu.Lock()
	exitFuncs := exitFuncs
	exitFuncsMu.Unlock()
	for _, f := range exitFuncs {
		f()
	}
	os.Exit(code)
}
