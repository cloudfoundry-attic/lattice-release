package exit_handler

import (
	"os"
	"sync"

	"github.com/cloudfoundry-incubator/lattice/ltc/exit_handler/exit_codes"
)

func New(signalChan chan os.Signal, systemExit func(code int)) ExitHandler {
	return &exitHandler{
		signalChan:  signalChan,
		systemExit:  systemExit,
		onExitFuncs: make([]func(), 0),
		exitCode:    exit_codes.SigInt,
	}
}

type ExitHandler interface {
	Run()
	OnExit(exitFunc func())
	Exit(code int)
}

type exitHandler struct {
	onExitFuncs []func()
	signalChan  chan os.Signal
	systemExit  func(int)
	exitCode    int
	sync.RWMutex
}

func (e *exitHandler) Run() {
	for {
		select {
		case signal := <-e.signalChan:
			if signal == os.Interrupt {
				e.Exit(e.exitCode)
			}
		}
	}
}

func (e *exitHandler) OnExit(exitFunc func()) {
	defer e.Unlock()
	e.Lock()
	e.onExitFuncs = append(e.onExitFuncs, exitFunc)
}

func (e *exitHandler) Exit(code int) {
	defer e.RUnlock()
	e.RLock()
	for _, exitFunc := range e.onExitFuncs {
		exitFunc()
	}
	e.systemExit(code)
}
