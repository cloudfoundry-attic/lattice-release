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

type DesiredLRPHandler struct {
	bbs    Bbs.ReceptorBBS
	logger lager.Logger
}

func NewDesiredLRPHandler(bbs Bbs.ReceptorBBS, logger lager.Logger) *DesiredLRPHandler {
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

	err = h.bbs.DesireLRP(log, desiredLRP)
	if err != nil {
		if _, ok := err.(models.ValidationError); ok {
			log.Error("lrp-request-invalid", err)
			writeBadRequestResponse(w, receptor.InvalidLRP, err)
			return
		}

		if err == bbserrors.ErrStoreResourceExists {
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
	log := h.logger.Session("get", lager.Data{
		"ProcessGuid": processGuid,
	})

	if processGuid == "" {
		err := errors.New("process_guid missing from request")
		log.Error("missing-process-guid", err)
		writeBadRequestResponse(w, receptor.InvalidRequest, err)
		return
	}

	desiredLRP, err := h.bbs.DesiredLRPByProcessGuid(processGuid)
	if err == bbserrors.ErrStoreResourceNotFound {
		writeDesiredLRPNotFoundResponse(w, processGuid)
		return
	}

	if err != nil {
		log.Error("unknown-error", err)
		writeUnknownErrorResponse(w, err)
		return
	}

	writeJSONResponse(w, http.StatusOK, serialization.DesiredLRPToResponse(desiredLRP))
}

func (h *DesiredLRPHandler) Update(w http.ResponseWriter, r *http.Request) {
	processGuid := r.FormValue(":process_guid")
	log := h.logger.Session("update", lager.Data{
		"ProcessGuid": processGuid,
	})

	if processGuid == "" {
		err := errors.New("process_guid missing from request")
		log.Error("missing-process-guid", err)
		writeBadRequestResponse(w, receptor.InvalidRequest, err)
		return
	}

	desireLRPRequest := receptor.DesiredLRPUpdateRequest{}

	err := json.NewDecoder(r.Body).Decode(&desireLRPRequest)
	if err != nil {
		log.Error("invalid-json", err)
		writeBadRequestResponse(w, receptor.InvalidJSON, err)
		return
	}

	update := serialization.DesiredLRPUpdateFromRequest(desireLRPRequest)

	err = h.bbs.UpdateDesiredLRP(log, processGuid, update)
	if err == bbserrors.ErrStoreResourceNotFound {
		writeDesiredLRPNotFoundResponse(w, processGuid)
		return
	}

	if err != nil {
		log.Error("unknown-error", err)
		writeUnknownErrorResponse(w, err)
		return
	}

	w.WriteHeader(http.StatusNoContent)
}

func (h *DesiredLRPHandler) Delete(w http.ResponseWriter, req *http.Request) {
	processGuid := req.FormValue(":process_guid")
	log := h.logger.Session("delete", lager.Data{
		"ProcessGuid": processGuid,
	})

	if processGuid == "" {
		err := errors.New("process_guid missing from request")
		log.Error("missing-process-guid", err)
		writeBadRequestResponse(w, receptor.InvalidRequest, err)
		return
	}

	err := h.bbs.RemoveDesiredLRPByProcessGuid(log, processGuid)
	if err == bbserrors.ErrStoreResourceNotFound {
		writeDesiredLRPNotFoundResponse(w, processGuid)
		return
	}

	if err != nil {
		log.Error("unknown-error", err)
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

	var desiredLRPs []models.DesiredLRP
	var err error

	if domain == "" {
		desiredLRPs, err = h.bbs.DesiredLRPs()
	} else {
		desiredLRPs, err = h.bbs.DesiredLRPsByDomain(domain)
	}

	writeDesiredLRPResponse(w, logger, desiredLRPs, err)
}

func writeDesiredLRPResponse(w http.ResponseWriter, logger lager.Logger, desiredLRPs []models.DesiredLRP, err error) {
	if err != nil {
		logger.Error("failed-to-fetch-desired-lrps", err)
		writeUnknownErrorResponse(w, err)
		return
	}

	responses := make([]receptor.DesiredLRPResponse, 0, len(desiredLRPs))
	for _, desiredLRP := range desiredLRPs {
		responses = append(responses, serialization.DesiredLRPToResponse(desiredLRP))
	}

	writeJSONResponse(w, http.StatusOK, responses)
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
