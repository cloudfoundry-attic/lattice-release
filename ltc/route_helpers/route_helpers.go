package route_helpers

import (
	"encoding/json"
	"fmt"
	"strconv"

	"github.com/cloudfoundry-incubator/receptor"
)

type Routes struct {
	AppRoutes     AppRoutes
	TcpRoutes     TcpRoutes
	DiegoSSHRoute *DiegoSSHRoute
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

const DiegoSSHRouter = "diego-ssh"

type DiegoSSHRoute struct {
	Port       uint16 `json:"container_port"`
	PrivateKey string `json:"private_key"`
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

	if r.DiegoSSHRoute != nil {
		data, _ := json.Marshal(r.DiegoSSHRoute)
		diegoSSHRoutingInfo := json.RawMessage(data)
		routingInfo[DiegoSSHRouter] = &diegoSSHRoutingInfo
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
	routes := Routes{}

	if routingInfo == nil {
		return Routes{}
	}

	data, found := routingInfo[TcpRouter]
	if found && data != nil {
		tcpRoutes := TcpRoutes{}
		err := json.Unmarshal(*data, &tcpRoutes)
		if err != nil {
			panic(err)
		}
		routes.TcpRoutes = tcpRoutes
	}

	data, found = routingInfo[AppRouter]
	if found && data != nil {
		appRoutes := AppRoutes{}
		err := json.Unmarshal(*data, &appRoutes)
		if err != nil {
			panic(err)
		}
		routes.AppRoutes = appRoutes
	}

	data, found = routingInfo[DiegoSSHRouter]
	if found && data != nil {
		diegoSSHRoute := &DiegoSSHRoute{}
		err := json.Unmarshal(*data, diegoSSHRoute)
		if err != nil {
			panic(err)
		}
		routes.DiegoSSHRoute = diegoSSHRoute
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

func BuildDefaultRoutingInfo(appName string, exposedPorts []uint16, primaryPort uint16, systemDomain string) AppRoutes {
	appRoutes := AppRoutes{}

	for _, port := range exposedPorts {
		hostnames := []string{}
		if port == primaryPort {
			hostnames = append(hostnames, fmt.Sprintf("%s.%s", appName, systemDomain))
		}

		hostnames = append(hostnames, fmt.Sprintf("%s-%s.%s", appName, strconv.Itoa(int(port)), systemDomain))
		appRoutes = append(appRoutes, AppRoute{
			Hostnames: hostnames,
			Port:      port,
		})
	}

	return appRoutes
}

func GetPrimaryPort(monitorPort uint16, exposedPorts []uint16) uint16 {
	primaryPort := uint16(0)
	if monitorPort != 0 {
		primaryPort = monitorPort
	} else if len(exposedPorts) > 0 {
		primaryPort = exposedPorts[0]
	}
	return primaryPort
}
