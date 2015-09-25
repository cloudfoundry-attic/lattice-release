package handlers

import (
	"errors"
	"fmt"
	"net/http"
	"strconv"

	"github.com/cloudfoundry-incubator/bbs"
	"github.com/cloudfoundry-incubator/bbs/models"
	"github.com/cloudfoundry-incubator/receptor"
	"github.com/cloudfoundry-incubator/receptor/serialization"
	"github.com/pivotal-golang/lager"
)

type ActualLRPHandler struct {
	bbs    bbs.Client
	logger lager.Logger
}

func NewActualLRPHandler(bbs bbs.Client, logger lager.Logger) *ActualLRPHandler {
	return &ActualLRPHandler{
		bbs:    bbs,
		logger: logger.Session("actual-lrp-handler"),
	}
}

func (h *ActualLRPHandler) GetAll(w http.ResponseWriter, req *http.Request) {
	domain := req.FormValue("domain")
	logger := h.logger.Session("get-all", lager.Data{
		"domain": domain,
	})

	filter := models.ActualLRPFilter{Domain: domain}
	actualLRPGroups, err := h.bbs.ActualLRPGroups(filter)

	if err != nil {
		logger.Error("failed-to-fetch-actual-lrp-groups", err)
		writeUnknownErrorResponse(w, err)
		return
	}

	responses := make([]receptor.ActualLRPResponse, 0, len(actualLRPGroups))
	for _, actualLRPGroup := range actualLRPGroups {
		lrp, evacuating := actualLRPGroup.Resolve()
		responses = append(responses, serialization.ActualLRPProtoToResponse(lrp, evacuating))
	}

	writeJSONResponse(w, http.StatusOK, responses)
}

func (h *ActualLRPHandler) GetAllByProcessGuid(w http.ResponseWriter, req *http.Request) {
	processGuid := req.FormValue(":process_guid")
	logger := h.logger.Session("get-all-by-process-guid", lager.Data{
		"ProcessGuid": processGuid,
	})

	if processGuid == "" {
		err := errors.New("process_guid missing from request")
		logger.Error("missing-process-guid", err)
		writeBadRequestResponse(w, receptor.InvalidRequest, err)
		return
	}

	actualLRPGroupsByIndex, err := h.bbs.ActualLRPGroupsByProcessGuid(processGuid)
	if err != nil {
		logger.Error("failed-to-fetch-actual-lrp-groups-by-process-guid", err)
		writeUnknownErrorResponse(w, err)
		return
	}

	responses := make([]receptor.ActualLRPResponse, 0, len(actualLRPGroupsByIndex))
	for _, actualLRPGroup := range actualLRPGroupsByIndex {
		lrp, evacuating := actualLRPGroup.Resolve()
		responses = append(responses, serialization.ActualLRPProtoToResponse(lrp, evacuating))
	}

	writeJSONResponse(w, http.StatusOK, responses)
}

func (h *ActualLRPHandler) GetByProcessGuidAndIndex(w http.ResponseWriter, req *http.Request) {
	processGuid := req.FormValue(":process_guid")
	indexString := req.FormValue(":index")

	logger := h.logger.Session("get-by-process-guid-and-index", lager.Data{
		"ProcessGuid": processGuid,
		"Index":       indexString,
	})

	if processGuid == "" {
		err := errors.New("process_guid missing from request")
		logger.Error("missing-process-guid", err)
		writeBadRequestResponse(w, receptor.InvalidRequest, err)
		return
	}

	if indexString == "" {
		err := errors.New("index missing from request")
		logger.Error("missing-index", err)
		writeBadRequestResponse(w, receptor.InvalidRequest, err)
		return
	}

	var err error

	index, indexErr := strconv.Atoi(indexString)
	if indexErr != nil {
		err = errors.New("index not a number")
		logger.Error("invalid-index", err)
		writeBadRequestResponse(w, receptor.InvalidRequest, err)
		return
	}

	actualLRPGroup, err := h.bbs.ActualLRPGroupByProcessGuidAndIndex(processGuid, index)
	if err != nil {
		bbsError := models.ConvertError(err)
		if bbsError.Type == models.Error_ResourceNotFound {
			writeJSONResponse(w, http.StatusNotFound, nil)
		} else {
			logger.Error("failed-to-fetch-actual-lrps-by-process-guid", err)
			writeUnknownErrorResponse(w, err)
		}
		return
	}

	actualLRP, evacuating := actualLRPGroup.Resolve()

	writeJSONResponse(w, http.StatusOK, serialization.ActualLRPProtoToResponse(actualLRP, evacuating))
}

func (h *ActualLRPHandler) KillByProcessGuidAndIndex(w http.ResponseWriter, req *http.Request) {
	processGuid := req.FormValue(":process_guid")
	indexString := req.FormValue(":index")
	logger := h.logger.Session("kill-by-process-guid-and-index", lager.Data{
		"ProcessGuid": processGuid,
		"Index":       indexString,
	})

	if processGuid == "" {
		err := errors.New("process_guid missing from request")
		logger.Error("missing-process-guid", err)
		writeBadRequestResponse(w, receptor.InvalidRequest, err)
		return
	}

	if indexString == "" {
		err := errors.New("index missing from request")
		logger.Error("missing-index", err)
		writeBadRequestResponse(w, receptor.InvalidRequest, err)
		return
	}

	index, err := strconv.Atoi(indexString)
	if err != nil {
		err = errors.New("index not a number")
		logger.Error("invalid-index", err)
		writeBadRequestResponse(w, receptor.InvalidRequest, err)
		return
	}

	actualLRPGroup, err := h.bbs.ActualLRPGroupByProcessGuidAndIndex(processGuid, index)
	if err != nil {
		bbsError := models.ConvertError(err)
		if bbsError.Type == models.Error_ResourceNotFound {
			responseErr := fmt.Errorf("process-guid '%s' does not exist or has no instance at index %d", processGuid, index)
			logger.Error("no-instances-to-delete", responseErr)
			writeJSONResponse(w, http.StatusNotFound, receptor.Error{
				Type:    receptor.ActualLRPIndexNotFound,
				Message: responseErr.Error(),
			})
		} else {
			logger.Error("failed-to-fetch-actual-lrps-by-process-guid", err)
			writeUnknownErrorResponse(w, err)
		}
		return
	}

	actualLRP, _ := actualLRPGroup.Resolve()
	h.bbs.RetireActualLRP(&actualLRP.ActualLRPKey)

	w.WriteHeader(http.StatusNoContent)
}
