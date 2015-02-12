package fake_integration_test_runner

import (
	"io"
	"sync"

	"fmt"
	"time"
)

func NewFakeIntegrationTestRunner(outputWriter io.Writer) *FakeIntegrationTestRunner {
	return &FakeIntegrationTestRunner{testOutputWriter: outputWriter}
}

type FakeIntegrationTestRunner struct {
	sync.RWMutex
	testOutputWriter io.Writer
	timeout          time.Duration
	verbose          bool
}

func (fake *FakeIntegrationTestRunner) Run(timeout time.Duration, verbose bool) {
	fake.Lock()
	defer fake.Unlock()

	fake.timeout = timeout
	fake.verbose = verbose

	fmt.Fprintf(fake.testOutputWriter, "Running fake integration tests!!!\n")
}

func (fake *FakeIntegrationTestRunner) GetArgsForRun() (time.Duration, bool) {
	fake.RLock()
	defer fake.RUnlock()
	return fake.timeout, fake.verbose
}
