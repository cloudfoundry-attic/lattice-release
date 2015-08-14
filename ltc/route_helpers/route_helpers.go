package route_helpers

import (
	"encoding/json"

	"github.com/cloudfoundry-incubator/receptor"
)

type Routes struct {
	AppRoutes AppRoutes
	TcpRoutes TcpRoutes
}

const AppRouter = "cf-router"

type AppRoutes []AppRoute

type AppRoute struct {
	Hostnames []string `json:"hostnames"`
	Port      uint16   `json:"port"`
}

const TcpRouter = "tcp-router"

type TcpRoutes []TcpRoute

type TcpRoute struct {
	ExternalPort uint16 `json:"external_port"`
	Port         uint16 `json:"container_port"`
}

func (r Routes) RoutingInfo() receptor.RoutingInfo {
	routingInfo := receptor.RoutingInfo{}

	if r.AppRoutes != nil {
		data, _ := json.Marshal(r.AppRoutes)
		appRoutingInfo := json.RawMessage(data)
		routingInfo[AppRouter] = &appRoutingInfo
	}

	if r.TcpRoutes != nil {
		data, _ := json.Marshal(r.TcpRoutes)
		tcpRoutingInfo := json.RawMessage(data)
		routingInfo[TcpRouter] = &tcpRoutingInfo
	}

	return routingInfo
}

func (l AppRoutes) RoutingInfo() receptor.RoutingInfo {
	data, _ := json.Marshal(l)
	routingInfo := json.RawMessage(data)
	return receptor.RoutingInfo{
		AppRouter: &routingInfo,
	}
}

func (l AppRoutes) HostnamesByPort() map[uint16][]string {
	routesByPort := make(map[uint16][]string)

	for _, route := range l {
		routesByPort[route.Port] = route.Hostnames
	}

	return routesByPort
}

func RoutesFromRoutingInfo(routingInfo receptor.RoutingInfo) Routes {
	var appRoutes AppRoutes
	var tcpRoutes TcpRoutes

	routes := Routes{}

	if routingInfo == nil {
		return Routes{}
	}

	data, found := routingInfo[TcpRouter]
	if found && data != nil {
		tcpRoutes = TcpRoutes{}
		err := json.Unmarshal(*data, &tcpRoutes)
		if err != nil {
			panic(err)
		}
		routes.TcpRoutes = tcpRoutes
	}

	data, found = routingInfo[AppRouter]
	if found && data != nil {
		appRoutes = AppRoutes{}
		err := json.Unmarshal(*data, &appRoutes)
		if err != nil {
			panic(err)
		}
		routes.AppRoutes = appRoutes
	}

	return routes

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
