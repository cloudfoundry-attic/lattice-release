package handlers

import (
	"errors"
	"fmt"
	"net/http"
	"strconv"
	"strings"
	"time"

	"github.com/cloudfoundry-incubator/bbs"
	"github.com/cloudfoundry-incubator/receptor"
	Bbs "github.com/cloudfoundry-incubator/runtime-schema/bbs"
	"github.com/cloudfoundry-incubator/runtime-schema/models"
	"github.com/pivotal-golang/lager"
)

type DomainHandler struct {
	legacyBBS Bbs.ReceptorBBS
	bbs       bbs.Client
	logger    lager.Logger
}

var (
	ErrDomainMissing = errors.New("domain missing from request")
	ErrMaxAgeMissing = errors.New("max-age directive missing from request")
)

func NewDomainHandler(bbs bbs.Client, legacyBBS Bbs.ReceptorBBS, logger lager.Logger) *DomainHandler {
	return &DomainHandler{
		bbs:       bbs,
		legacyBBS: legacyBBS,
		logger:    logger.Session("domain-handler"),
	}
}

func (h *DomainHandler) Upsert(w http.ResponseWriter, req *http.Request) {
	domain := req.FormValue(":domain")
	logger := h.logger.Session("upsert", lager.Data{
		"Domain": domain,
	})

	if domain == "" {
		logger.Error("missing-domain", ErrDomainMissing)
		writeBadRequestResponse(w, receptor.InvalidRequest, ErrDomainMissing)
		return
	}

	ttl := 0

	cacheControl := req.Header["Cache-Control"]
	if cacheControl != nil {
		var maxAge string
		for _, directive := range cacheControl {
			if strings.HasPrefix(directive, "max-age=") {
				maxAge = directive
				break
			}
		}
		if maxAge == "" {
			logger.Error("missing-max-age-directive", ErrMaxAgeMissing)
			writeBadRequestResponse(w, receptor.InvalidRequest, ErrMaxAgeMissing)
			return
		}

		var err error
		ttl, err = strconv.Atoi(maxAge[8:])
		if err != nil {
			err := fmt.Errorf("invalid-max-age-directive: %s", maxAge)
			logger.Error("invalid-max-age-directive", err)
			writeBadRequestResponse(w, receptor.InvalidRequest, err)
			return
		}
	}

	err := h.bbs.UpsertDomain(domain, time.Second*time.Duration(ttl))
	if err != nil {
		if _, ok := err.(models.ValidationError); ok {
			logger.Error("failed-to-upsert-domain", err)
			writeBadRequestResponse(w, receptor.InvalidDomain, err)
			return
		}

		logger.Error("upsert-domain-failed", err)
		writeUnknownErrorResponse(w, err)
		return
	}

	w.WriteHeader(http.StatusNoContent)
}

func (h *DomainHandler) GetAll(w http.ResponseWriter, req *http.Request) {
	logger := h.logger.Session("get-all")

	domains, err := h.bbs.Domains()
	if err != nil {
		logger.Error("failed-to-fetch-domains", err)
		writeUnknownErrorResponse(w, err)
		return
	}

	writeJSONResponse(w, http.StatusOK, domains)
}
