package exit_handler

import (
	"github.com/pivotal-cf-experimental/lattice-cli/exit_handler/exit_codes"
	"os"
)

func New(signalChan chan os.Signal, systemExit func(code int)) ExitHandler {
	return &exitHandler{
		signalChan:      signalChan,
		systemExit:      systemExit,
		onExitFuncs:     make([]func(), 0),
		onExitFuncsChan: make(chan func()),
		exitCode:        exit_codes.SigInt,
	}
}

type ExitHandler interface {
	Run()
	OnExit(exitFunc func())
	Exit(code int)
}

type exitHandler struct {
	onExitFuncs     []func()
	onExitFuncsChan chan func()
	signalChan      chan os.Signal
	systemExit      func(int)
	exitCode        int
}

func (e *exitHandler) Run() {
	for {
		select {
		case signal := <-e.signalChan:
			if signal == os.Interrupt {
				for _, exitFunc := range e.onExitFuncs {
					exitFunc()
				}
				e.systemExit(e.exitCode)
				return
			}
		case exitFunc := <-e.onExitFuncsChan:
			e.onExitFuncs = append(e.onExitFuncs, exitFunc)
		}
	}
}

func (e *exitHandler) OnExit(exitFunc func()) {
	e.onExitFuncsChan <- exitFunc
}

func (e *exitHandler) Exit(code int) {
	e.exitCode = code
	e.signalChan <- os.Interrupt
}
