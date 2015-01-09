package exit_handler

import (
	"os"
)

func New(signalChan chan os.Signal, systemExit func(code int)) *ExitHandler {
	return &ExitHandler{
		signalChan:      signalChan,
		systemExit:      systemExit,
		onExitFuncs:     make([]func(), 0),
		onExitFuncsChan: make(chan func()),
	}
}

type ExitHandler struct {
	onExitFuncs     []func()
	onExitFuncsChan chan func()
	signalChan      chan os.Signal
	systemExit      func(int)
}

func (e *ExitHandler) Run() {
	for {
		select {
		case signal := <-e.signalChan:
			if signal == os.Interrupt {
				for _, exitFunc := range e.onExitFuncs {
					exitFunc()
				}
				e.systemExit(130)
				return
			}
		case exitFunc := <-e.onExitFuncsChan:
			e.onExitFuncs = append(e.onExitFuncs, exitFunc)
		}
	}
}

func (e *ExitHandler) OnExit(exitFunc func()) {
	e.onExitFuncsChan <- exitFunc
}
