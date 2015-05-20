package task_handler

import (
	"bytes"
	"encoding/json"
	"fmt"
	"net/http"
	"os"
	"regexp"

	"github.com/cloudfoundry-incubator/cf_http"
	"github.com/cloudfoundry-incubator/receptor/serialization"
	"github.com/cloudfoundry-incubator/runtime-schema/bbs"
	"github.com/cloudfoundry-incubator/runtime-schema/models"
	"github.com/pivotal-golang/lager"
	"github.com/tedsuo/ifrit"
	"github.com/tedsuo/ifrit/grouper"
)

const MAX_RETRIES = 3

const POOL_SIZE = 20

func NewTaskWorkerPool(receptorBBS bbs.ReceptorBBS, logger lager.Logger) (ifrit.Runner, chan<- models.Task) {
	taskQueue := make(chan models.Task, POOL_SIZE)

	members := make(grouper.Members, POOL_SIZE)

	for i := 0; i < POOL_SIZE; i++ {
		name := fmt.Sprintf("task-worker-%d", i)
		members[i].Name = name
		members[i].Runner = newTaskWorker(taskQueue, receptorBBS, logger.Session(name))
	}

	return grouper.NewParallel(os.Interrupt, members), taskQueue
}

func newTaskWorker(taskQueue <-chan models.Task, receptorBBS bbs.ReceptorBBS, logger lager.Logger) *taskWorker {
	return &taskWorker{
		taskQueue:   taskQueue,
		receptorBBS: receptorBBS,
		logger:      logger,
		httpClient:  cf_http.NewClient(),
	}
}

type taskWorker struct {
	taskQueue   <-chan models.Task
	receptorBBS bbs.ReceptorBBS
	logger      lager.Logger
	httpClient  *http.Client
}

func (t *taskWorker) Run(signals <-chan os.Signal, ready chan<- struct{}) error {
	t.logger.Debug("starting")
	close(ready)
	for {
		select {
		case task := <-t.taskQueue:
			t.handleCompletedTask(task)
		case <-signals:
			t.logger.Debug("exited")
			return nil
		}
	}
}

func (t *taskWorker) handleCompletedTask(task models.Task) {
	logger := t.logger.WithData(lager.Data{"task-guid": task.TaskGuid})

	if task.CompletionCallbackURL != nil {
		logger.Info("resolving-task")
		err := t.receptorBBS.ResolvingTask(logger, task.TaskGuid)
		if err != nil {
			logger.Error("marking-task-as-resolving-failed", err)
			return
		}

		logger = logger.WithData(lager.Data{"callback_url": task.CompletionCallbackURL.String()})

		json, err := json.Marshal(serialization.TaskToResponse(task))
		if err != nil {
			logger.Error("marshalling-task-failed", err)
			return
		}

		var statusCode int

		for i := 0; i < MAX_RETRIES; i++ {
			request, err := http.NewRequest("POST", task.CompletionCallbackURL.String(), bytes.NewReader(json))
			if err != nil {
				logger.Error("building-request-failed", err)
				return
			}

			request.Header.Set("Content-Type", "application/json")

			response, err := t.httpClient.Do(request)
			if err != nil {
				matched, _ := regexp.MatchString("use of closed network connection", err.Error())
				if matched {
					continue
				}
				logger.Error("doing-request-failed", err)
				return
			}

			statusCode = response.StatusCode
			if shouldResolve(statusCode) {
				err = t.receptorBBS.ResolveTask(logger, task.TaskGuid)
				if err != nil {
					logger.Error("resolving-task-failed", err)
					return
				}

				logger.Info("resolved-task", lager.Data{"status_code": statusCode})
				return
			}
		}

		logger.Info("callback-failed", lager.Data{"status_code": statusCode})
	}
}

func shouldResolve(status int) bool {
	switch status {
	case http.StatusServiceUnavailable, http.StatusGatewayTimeout:
		return false
	default:
		return true
	}
}
