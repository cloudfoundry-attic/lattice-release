package route_helpers

import (
	"encoding/json"

	"github.com/cloudfoundry-incubator/receptor"
	"strings"
)

const AppRouter = "cf-router"

type AppRoutes []AppRoute

type AppRoute struct {
	Hostnames []string `json:"hostnames"`
	Port      uint16   `json:"port"`
}

func (l AppRoutes) RoutingInfo() receptor.RoutingInfo {
	data, _ := json.Marshal(l)
	routingInfo := json.RawMessage(data)
	return receptor.RoutingInfo{
		AppRouter: &routingInfo,
	}
}

func (l AppRoutes) String() string {

	routeStrings := []string{}
	for _, route := range l {
		routeStrings = append(routeStrings, route.String())
	}

	return strings.Join(routeStrings, "| ")
}

func (r AppRoute) String() string {
	return strings.Join(r.Hostnames, ", ")
}

func AppRoutesFromRoutingInfo(routingInfo receptor.RoutingInfo) AppRoutes {
	if routingInfo == nil {
		return nil
	}

	data, found := routingInfo[AppRouter]
	if !found {
		return nil
	}

	if data == nil {
		return nil
	}

	routes := AppRoutes{}
	err := json.Unmarshal(*data, &routes)
	if err != nil {
		panic(err)
	}

	return routes
}
