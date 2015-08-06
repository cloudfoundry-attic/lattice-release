package fake_exit_handler

import "sync"

type FakeExitHandler struct {
	sync.RWMutex
	exitFunc       func()
	ExitCalledWith []int
}

func (f *FakeExitHandler) OnExit(exitHandler func()) {
	f.Lock()
	defer f.Unlock()
	f.exitFunc = exitHandler
}

func (*FakeExitHandler) Run() {}

func (f *FakeExitHandler) Exit(code int) {
	f.Lock()
	defer f.Unlock()
	f.ExitCalledWith = append(f.ExitCalledWith, code)
	if f.exitFunc != nil {
		f.exitFunc()
	}
}
