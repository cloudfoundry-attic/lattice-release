package handlers

import (
	"encoding/json"
	"errors"
	"fmt"
	"net/http"

	"github.com/cloudfoundry-incubator/receptor"
	"github.com/cloudfoundry-incubator/receptor/serialization"
	Bbs "github.com/cloudfoundry-incubator/runtime-schema/bbs"
	"github.com/cloudfoundry-incubator/runtime-schema/bbs/bbserrors"
	"github.com/cloudfoundry-incubator/runtime-schema/models"
	"github.com/pivotal-golang/lager"
)

type TaskHandler struct {
	bbs    Bbs.ReceptorBBS
	logger lager.Logger
}

func NewTaskHandler(bbs Bbs.ReceptorBBS, logger lager.Logger) *TaskHandler {
	return &TaskHandler{
		bbs:    bbs,
		logger: logger.Session("task-handler"),
	}
}

func (h *TaskHandler) Create(w http.ResponseWriter, r *http.Request) {
	log := h.logger.Session("create")
	taskRequest := receptor.TaskCreateRequest{}

	err := json.NewDecoder(r.Body).Decode(&taskRequest)
	if err != nil {
		log.Error("invalid-json", err)
		writeJSONResponse(w, http.StatusBadRequest, receptor.Error{
			Type:    receptor.InvalidJSON,
			Message: err.Error(),
		})
		return
	}

	task, err := serialization.TaskFromRequest(taskRequest)
	if err != nil {
		log.Error("task-request-invalid", err)
		writeJSONResponse(w, http.StatusBadRequest, receptor.Error{
			Type:    receptor.InvalidTask,
			Message: err.Error(),
		})
		return
	}

	log.Debug("creating-task", lager.Data{"task-guid": task.TaskGuid})

	err = h.bbs.DesireTask(log, task)
	if err != nil {
		log.Error("failed-to-desire-task", err)

		if _, ok := err.(models.ValidationError); ok {
			writeJSONResponse(w, http.StatusBadRequest, receptor.Error{
				Type:    receptor.InvalidTask,
				Message: err.Error(),
			})
			return
		}

		if err == bbserrors.ErrStoreResourceExists {
			writeJSONResponse(w, http.StatusConflict, receptor.Error{
				Type:    receptor.TaskGuidAlreadyExists,
				Message: "task already exists",
			})
		} else {
			writeUnknownErrorResponse(w, err)
		}
		return
	}

	log.Info("created", lager.Data{"task-guid": task.TaskGuid})
	w.WriteHeader(http.StatusCreated)
}

func (h *TaskHandler) GetAll(w http.ResponseWriter, req *http.Request) {
	domain := req.FormValue("domain")
	logger := h.logger.Session("get-all", lager.Data{
		"domain": domain,
	})

	var tasks []models.Task
	var err error

	if domain == "" {
		tasks, err = h.bbs.Tasks(logger)
	} else {
		tasks, err = h.bbs.TasksByDomain(logger, domain)
	}

	writeTaskResponse(w, logger, tasks, err)
}

func (h *TaskHandler) GetByGuid(w http.ResponseWriter, req *http.Request) {
	guid := req.FormValue(":task_guid")
	logger := h.logger.Session("get-by-guid", lager.Data{
		"TaskGuid": guid,
	})

	if guid == "" {
		err := errors.New("task_guid missing from request")
		logger.Error("missing-task-guid", err)
		writeBadRequestResponse(w, receptor.InvalidRequest, err)
		return
	}

	task, err := h.bbs.TaskByGuid(logger, guid)
	if err == bbserrors.ErrStoreResourceNotFound {
		h.logger.Error("failed-to-fetch-task", err)
		writeTaskNotFoundResponse(w, guid)
		return
	}

	if err != nil {
		if err == bbserrors.ErrStoreResourceNotFound {
			h.logger.Error("failed-to-fetch-task", err)
			writeTaskNotFoundResponse(w, guid)
			return
		}

		h.logger.Error("failed-to-fetch-task", err)
		writeUnknownErrorResponse(w, err)
		return
	}

	writeJSONResponse(w, http.StatusOK, serialization.TaskToResponse(task))
}

func (h *TaskHandler) Delete(w http.ResponseWriter, req *http.Request) {
	guid := req.FormValue(":task_guid")

	err := h.bbs.ResolvingTask(h.logger, guid)
	if err != nil {
		switch err.(type) {
		case bbserrors.TaskStateTransitionError:
			h.logger.Error("invalid-task-state-transition", err)
			writeJSONResponse(w, http.StatusConflict, receptor.Error{
				Type:    receptor.TaskNotDeletable,
				Message: "This task has not been completed. Please retry when it is completed.",
			})
			return
		default:
			if err == bbserrors.ErrStoreResourceNotFound {
				h.logger.Error("task-not-found", err)
				writeTaskNotFoundResponse(w, guid)
				return
			}
			h.logger.Error("failed-to-mark-task-resolving", err)
			writeUnknownErrorResponse(w, err)
			return
		}
	}

	err = h.bbs.ResolveTask(h.logger, guid)
	if err != nil {
		h.logger.Error("failed-to-resolve-task", err)
		writeUnknownErrorResponse(w, err)
	}
}

func (h *TaskHandler) Cancel(w http.ResponseWriter, req *http.Request) {
	guid := req.FormValue(":task_guid")

	err := h.bbs.CancelTask(h.logger, guid)

	switch err {
	case nil:
	case bbserrors.ErrStoreResourceNotFound:
		h.logger.Error("failed-to-cancel-task", err)
		writeTaskNotFoundResponse(w, guid)
	default:
		h.logger.Error("failed-to-fetch-task", err)
		writeUnknownErrorResponse(w, err)
	}
}

func writeTaskResponse(w http.ResponseWriter, logger lager.Logger, tasks []models.Task, err error) {
	if err != nil {
		logger.Error("failed-to-fetch-tasks", err)
		writeUnknownErrorResponse(w, err)
		return
	}

	taskResponses := make([]receptor.TaskResponse, 0, len(tasks))
	for _, task := range tasks {
		taskResponses = append(taskResponses, serialization.TaskToResponse(task))
	}

	writeJSONResponse(w, http.StatusOK, taskResponses)
}

func writeTaskNotFoundResponse(w http.ResponseWriter, taskGuid string) {
	writeJSONResponse(w, http.StatusNotFound, receptor.Error{
		Type:    receptor.TaskNotFound,
		Message: fmt.Sprintf("task with guid '%s' not found", taskGuid),
	})
}
