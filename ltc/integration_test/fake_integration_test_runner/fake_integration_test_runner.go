package fake_integration_test_runner

import (
	"sync"
	"time"
)

func NewFakeIntegrationTestRunner() *FakeIntegrationTestRunner {
	return &FakeIntegrationTestRunner{}
}

type FakeIntegrationTestRunner struct {
	sync.RWMutex
	timeout      time.Duration
	verbose      bool
	cliHelp      bool
	runCallCount int
}

func (fake *FakeIntegrationTestRunner) Run(timeout time.Duration, verbose bool, cliHelp bool) {
	fake.Lock()
	defer fake.Unlock()

	fake.timeout = timeout
	fake.verbose = verbose
	fake.cliHelp = cliHelp
	fake.runCallCount++
}

func (fake *FakeIntegrationTestRunner) RunCallCount() int {
	fake.RLock()
	defer fake.RUnlock()
	return fake.runCallCount
}

func (fake *FakeIntegrationTestRunner) GetArgsForRun() (time.Duration, bool, bool) {
	fake.RLock()
	defer fake.RUnlock()
	return fake.timeout, fake.verbose, fake.cliHelp
}
