package fake_log_reader

import (
	"github.com/cloudfoundry/noaa/events"
	"sync"
)

type FakeLogReader struct {
	sync.RWMutex
	StopChan       chan struct{}
	logs           []*events.LogMessage
	errors         []error
	logTailStopped bool
	appGuid        string
}

func (f *FakeLogReader) TailLogs(appGuid string, logCallback func(*events.LogMessage), errorCallback func(error), stopChan chan struct{}) {
	for _, log := range f.logs {
		logCallback(log)
	}

	for _, err := range f.errors {
		errorCallback(err)
	}

	f.Lock()
	defer f.Unlock()
	f.appGuid = appGuid

	go func() {
		select {
		case <-stopChan:
			f.Lock()
			defer f.Unlock()
			f.logTailStopped = true
			return
		}
	}()
}

func (f *FakeLogReader) GetAppGuid() string {
	f.RLock()
	defer f.RUnlock()
	return f.appGuid
}

func (f *FakeLogReader) IsLogTailStopped() bool {
	f.RLock()
	defer f.RUnlock()
	return f.logTailStopped
}

func (f *FakeLogReader) AddLog(log *events.LogMessage) {
	f.logs = append(f.logs, log)
}

func (f *FakeLogReader) AddError(err error) {
	f.errors = append(f.errors, err)
}
