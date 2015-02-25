package handlers

import (
	"net/http"

	"github.com/cloudfoundry-incubator/receptor"
	"github.com/cloudfoundry-incubator/receptor/serialization"
	Bbs "github.com/cloudfoundry-incubator/runtime-schema/bbs"
	"github.com/pivotal-golang/lager"
)

type CellHandler struct {
	bbs    Bbs.ReceptorBBS
	logger lager.Logger
}

func NewCellHandler(bbs Bbs.ReceptorBBS, logger lager.Logger) *CellHandler {
	return &CellHandler{
		bbs:    bbs,
		logger: logger.Session("cell-handler"),
	}
}

func (h *CellHandler) GetAll(w http.ResponseWriter, req *http.Request) {
	logger := h.logger.Session("get-all")

	cellPresences, err := h.bbs.Cells()
	if err != nil {
		logger.Error("failed-to-fetch-cells", err)
		writeUnknownErrorResponse(w, err)
		return
	}

	responses := make([]receptor.CellResponse, 0, len(cellPresences))
	for _, cellPresence := range cellPresences {
		responses = append(responses, serialization.CellPresenceToCellResponse(cellPresence))
	}

	writeJSONResponse(w, http.StatusOK, responses)
}
