package fake_cluster_test_runner

import (
	"sync"
	"time"
)

type FakeClusterTestRunner struct {
	sync.RWMutex
	timeout      time.Duration
	verbose      bool
	runCallCount int
}

func (fake *FakeClusterTestRunner) Run(timeout time.Duration, verbose bool) {
	fake.Lock()
	defer fake.Unlock()

	fake.timeout = timeout
	fake.verbose = verbose
	fake.runCallCount++
}

func (fake *FakeClusterTestRunner) RunCallCount() int {
	fake.RLock()
	defer fake.RUnlock()
	return fake.runCallCount
}

func (fake *FakeClusterTestRunner) GetArgsForRun() (time.Duration, bool) {
	fake.RLock()
	defer fake.RUnlock()
	return fake.timeout, fake.verbose
}
