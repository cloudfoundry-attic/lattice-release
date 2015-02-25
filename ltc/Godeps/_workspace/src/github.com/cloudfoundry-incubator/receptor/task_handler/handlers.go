package task_handler

import (
	"net/http"

	"github.com/cloudfoundry-incubator/receptor/handlers"
	"github.com/cloudfoundry-incubator/runtime-schema/models"
	"github.com/cloudfoundry-incubator/runtime-schema/routes"
	"github.com/pivotal-golang/lager"
	"github.com/tedsuo/rata"
)

func New(enqueue chan<- models.Task, logger lager.Logger) http.Handler {
	taskHandler := NewHandler(enqueue, logger)

	actions := rata.Handlers{
		// internal Tasks
		routes.CompleteTasks: taskHandler,
	}

	handler, err := rata.NewRouter(routes.CompleteTasksRoutes, actions)
	if err != nil {
		panic("unable to create router: " + err.Error())
	}

	handler = handlers.LogWrap(handler, logger)

	return handler
}
