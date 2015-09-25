package handlers

import (
	"encoding/json"
	"errors"
	"fmt"
	"net/http"

	"github.com/cloudfoundry-incubator/bbs"
	"github.com/cloudfoundry-incubator/bbs/models"
	"github.com/cloudfoundry-incubator/receptor"
	"github.com/cloudfoundry-incubator/receptor/serialization"
	"github.com/pivotal-golang/lager"
)

type DesiredLRPHandler struct {
	bbs    bbs.Client
	logger lager.Logger
}

func NewDesiredLRPHandler(bbs bbs.Client, logger lager.Logger) *DesiredLRPHandler {
	return &DesiredLRPHandler{
		bbs:    bbs,
		logger: logger.Session("desired-lrp-handler"),
	}
}

func (h *DesiredLRPHandler) Create(w http.ResponseWriter, r *http.Request) {
	log := h.logger.Session("create")
	desireLRPRequest := receptor.DesiredLRPCreateRequest{}

	err := json.NewDecoder(r.Body).Decode(&desireLRPRequest)
	if err != nil {
		log.Error("invalid-json", err)
		writeBadRequestResponse(w, receptor.InvalidJSON, err)
		return
	}

	desiredLRP := serialization.DesiredLRPFromRequest(desireLRPRequest)

	err = h.bbs.DesireLRP(desiredLRP)
	if err != nil {
		bbsError := models.ConvertError(err)
		switch bbsError.Type {
		case models.Error_InvalidRequest:
			log.Error("lrp-request-invalid", err)
			writeBadRequestResponse(w, receptor.InvalidLRP, err)
			return
		case models.Error_ResourceExists:
			log.Error("lrp-request-invalid", err)
			writeDesiredLRPAlreadyExistsResponse(w, desiredLRP.ProcessGuid)
			return
		}

		log.Error("desire-lrp-failed", err)
		writeUnknownErrorResponse(w, err)
		return
	}

	w.WriteHeader(http.StatusCreated)
}

func (h *DesiredLRPHandler) Get(w http.ResponseWriter, r *http.Request) {
	processGuid := r.FormValue(":process_guid")
	logger := h.logger.Session("get", lager.Data{
		"ProcessGuid": processGuid,
	})

	if processGuid == "" {
		err := errors.New("process_guid missing from request")
		logger.Error("missing-process-guid", err)
		writeBadRequestResponse(w, receptor.InvalidRequest, err)
		return
	}

	desiredLRP, err := h.bbs.DesiredLRPByProcessGuid(processGuid)
	if err != nil {
		bbsError := models.ConvertError(err)
		if bbsError.Type == models.Error_ResourceNotFound {
			writeDesiredLRPNotFoundResponse(w, processGuid)
			return
		}

		logger.Error("unknown-error", err)
		writeUnknownErrorResponse(w, err)
		return
	}

	writeJSONResponse(w, http.StatusOK, serialization.DesiredLRPProtoToResponse(desiredLRP))
}

func (h *DesiredLRPHandler) Update(w http.ResponseWriter, r *http.Request) {
	processGuid := r.FormValue(":process_guid")
	logger := h.logger.Session("update", lager.Data{
		"ProcessGuid": processGuid,
	})

	if processGuid == "" {
		err := errors.New("process_guid missing from request")
		logger.Error("missing-process-guid", err)
		writeBadRequestResponse(w, receptor.InvalidRequest, err)
		return
	}

	desireLRPRequest := receptor.DesiredLRPUpdateRequest{}

	err := json.NewDecoder(r.Body).Decode(&desireLRPRequest)
	if err != nil {
		logger.Error("invalid-json", err)
		writeBadRequestResponse(w, receptor.InvalidJSON, err)
		return
	}

	update := serialization.DesiredLRPUpdateFromRequest(desireLRPRequest)

	updateAttempts := 0
	for updateAttempts < 2 {
		err = h.bbs.UpdateDesiredLRP(processGuid, update)
		bbsError := models.ConvertError(err)
		if bbsError == nil || bbsError.Type != models.Error_ResourceConflict {
			// we only want to retry on compare and swap errors
			break
		}

		updateAttempts++
		logger.Error("failed-to-compare-and-swap", err, lager.Data{"Attempt": updateAttempts})
	}

	if err != nil {
		bbsError := models.ConvertError(err)
		switch bbsError.Type {
		case models.Error_ResourceNotFound:
			logger.Error("desired-lrp-not-found", err)
			writeDesiredLRPNotFoundResponse(w, processGuid)
			return
		case models.Error_ResourceConflict:
			logger.Error("failed-to-compare-and-swap", err)
			writeCompareAndSwapFailedResponse(w, processGuid)
			return
		default:
			logger.Error("unknown-error", err)
			writeUnknownErrorResponse(w, err)
			return
		}
	}

	w.WriteHeader(http.StatusNoContent)
}

func (h *DesiredLRPHandler) Delete(w http.ResponseWriter, req *http.Request) {
	processGuid := req.FormValue(":process_guid")
	logger := h.logger.Session("delete", lager.Data{
		"ProcessGuid": processGuid,
	})

	if processGuid == "" {
		err := errors.New("process_guid missing from request")
		logger.Error("missing-process-guid", err)
		writeBadRequestResponse(w, receptor.InvalidRequest, err)
		return
	}

	err := h.bbs.RemoveDesiredLRP(processGuid)
	if err != nil {
		bbsError := models.ConvertError(err)
		if bbsError.Type == models.Error_ResourceNotFound {
			writeDesiredLRPNotFoundResponse(w, processGuid)
			return
		}

		logger.Error("unknown-error", err)
		writeUnknownErrorResponse(w, err)
		return
	}

	w.WriteHeader(http.StatusNoContent)
}

func (h *DesiredLRPHandler) GetAll(w http.ResponseWriter, req *http.Request) {
	domain := req.FormValue("domain")
	logger := h.logger.Session("get-all", lager.Data{
		"domain": domain,
	})

	filter := models.DesiredLRPFilter{Domain: domain}
	desiredLRPs, err := h.bbs.DesiredLRPs(filter)

	writeDesiredLRPProtoResponse(w, logger, desiredLRPs, err)
}

func writeDesiredLRPProtoResponse(w http.ResponseWriter, logger lager.Logger, desiredLRPs []*models.DesiredLRP, err error) {
	if err != nil {
		logger.Error("failed-to-fetch-desired-lrps", err)
		writeUnknownErrorResponse(w, err)
		return
	}

	responses := make([]receptor.DesiredLRPResponse, 0, len(desiredLRPs))
	for _, desiredLRP := range desiredLRPs {
		responses = append(responses, serialization.DesiredLRPProtoToResponse(desiredLRP))
	}

	writeJSONResponse(w, http.StatusOK, responses)
}

func writeCompareAndSwapFailedResponse(w http.ResponseWriter, processGuid string) {
	writeJSONResponse(w, http.StatusInternalServerError, receptor.Error{
		Type:    receptor.ResourceConflict,
		Message: fmt.Sprintf("Desired LRP with guid '%s' failed to update", processGuid),
	})
}

func writeDesiredLRPNotFoundResponse(w http.ResponseWriter, processGuid string) {
	writeJSONResponse(w, http.StatusNotFound, receptor.Error{
		Type:    receptor.DesiredLRPNotFound,
		Message: fmt.Sprintf("Desired LRP with guid '%s' not found", processGuid),
	})
}

func writeDesiredLRPAlreadyExistsResponse(w http.ResponseWriter, processGuid string) {
	writeJSONResponse(w, http.StatusConflict, receptor.Error{
		Type:    receptor.DesiredLRPAlreadyExists,
		Message: fmt.Sprintf("Desired LRP with guid '%s' already exists", processGuid),
	})
}
