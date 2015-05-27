package handlers

import (
	"net/http"

	"github.com/cloudfoundry-incubator/receptor"
	"github.com/cloudfoundry-incubator/receptor/event"
	Bbs "github.com/cloudfoundry-incubator/runtime-schema/bbs"
	"github.com/pivotal-golang/lager"
	"github.com/tedsuo/rata"
)

func New(bbs Bbs.ReceptorBBS, hub event.Hub, logger lager.Logger, username, password string, corsEnabled bool) http.Handler {
	taskHandler := NewTaskHandler(bbs, logger)
	desiredLRPHandler := NewDesiredLRPHandler(bbs, logger)
	actualLRPHandler := NewActualLRPHandler(bbs, logger)
	cellHandler := NewCellHandler(bbs, logger)
	domainHandler := NewDomainHandler(bbs, logger)
	eventStreamHandler := NewEventStreamHandler(hub, logger)
	authCookieHandler := NewAuthCookieHandler(logger)

	actions := rata.Handlers{
		// Tasks
		receptor.CreateTaskRoute: route(taskHandler.Create),
		receptor.TasksRoute:      route(taskHandler.GetAll),
		receptor.GetTaskRoute:    route(taskHandler.GetByGuid),
		receptor.DeleteTaskRoute: route(taskHandler.Delete),
		receptor.CancelTaskRoute: route(taskHandler.Cancel),

		// DesiredLRPs
		receptor.CreateDesiredLRPRoute: route(desiredLRPHandler.Create),
		receptor.GetDesiredLRPRoute:    route(desiredLRPHandler.Get),
		receptor.UpdateDesiredLRPRoute: route(desiredLRPHandler.Update),
		receptor.DeleteDesiredLRPRoute: route(desiredLRPHandler.Delete),
		receptor.DesiredLRPsRoute:      route(desiredLRPHandler.GetAll),

		// ActualLRPs
		receptor.ActualLRPsRoute:                         route(actualLRPHandler.GetAll),
		receptor.ActualLRPsByProcessGuidRoute:            route(actualLRPHandler.GetAllByProcessGuid),
		receptor.ActualLRPByProcessGuidAndIndexRoute:     route(actualLRPHandler.GetByProcessGuidAndIndex),
		receptor.KillActualLRPByProcessGuidAndIndexRoute: route(actualLRPHandler.KillByProcessGuidAndIndex),

		// Cells
		receptor.CellsRoute: route(cellHandler.GetAll),

		// Domains
		receptor.UpsertDomainRoute: route(domainHandler.Upsert),
		receptor.DomainsRoute:      route(domainHandler.GetAll),

		// Event Streaming
		receptor.EventStream: route(eventStreamHandler.EventStream),

		// Authentication Cookie
		receptor.GenerateCookie: route(authCookieHandler.GenerateCookie),
	}

	handler, err := rata.NewRouter(receptor.Routes, actions)
	if err != nil {
		panic("unable to create router: " + err.Error())
	}

	if username != "" {
		handler = CookieAuthWrap(BasicAuthWrap(handler, username, password), receptor.AuthorizationCookieName)
	}

	if corsEnabled {
		handler = CORSWrapper(handler)
	}

	return handler
}

func route(f func(w http.ResponseWriter, r *http.Request)) http.Handler {
	return http.HandlerFunc(f)
}
