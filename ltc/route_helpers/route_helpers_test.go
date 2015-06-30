package route_helpers_test

import (
	"encoding/json"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"

	"github.com/cloudfoundry-incubator/lattice/ltc/route_helpers"
	"github.com/cloudfoundry-incubator/receptor"
)

var _ = Describe("RoutingInfoHelpers", func() {
	var (
		route1 route_helpers.AppRoute
		route2 route_helpers.AppRoute
		route3 route_helpers.AppRoute

		routes route_helpers.AppRoutes
	)

	BeforeEach(func() {
		route1 = route_helpers.AppRoute{
			Hostnames: []string{"foo1.example.com", "bar1.examaple.com"},
			Port:      11111,
		}
		route2 = route_helpers.AppRoute{
			Hostnames: []string{"foo2.example.com", "bar2.examaple.com"},
			Port:      22222,
		}
		route3 = route_helpers.AppRoute{
			Hostnames: []string{"foo3.example.com", "bar3.examaple.com"},
			Port:      33333,
		}

		routes = route_helpers.AppRoutes{route1, route2, route3}
	})

	Describe("AppRoutes", func() {
		Describe("RoutingInfo", func() {
			var routingInfo receptor.RoutingInfo

			JustBeforeEach(func() {
				routingInfo = routes.RoutingInfo()
			})

			It("wraps the serialized routes with the correct key", func() {
				expectedBytes, err := json.Marshal(routes)
				Expect(err).ToNot(HaveOccurred())

				payload, err := routingInfo[route_helpers.AppRouter].MarshalJSON()
				Expect(err).ToNot(HaveOccurred())

				Expect(payload).To(MatchJSON(expectedBytes))
			})

			Context("when AppRoutes is empty", func() {
				BeforeEach(func() {
					routes = route_helpers.AppRoutes{}
				})

				It("marshals an empty list", func() {
					payload, err := routingInfo[route_helpers.AppRouter].MarshalJSON()
					Expect(err).ToNot(HaveOccurred())

					Expect(payload).To(MatchJSON(`[]`))
				})
			})
		})
	})

	Describe("AppRoutesFromRoutingInfo", func() {
		var routingInfo receptor.RoutingInfo

		Context("when the method returns a value", func() {
			var routesResult route_helpers.AppRoutes

			JustBeforeEach(func() {
				routesResult = route_helpers.AppRoutesFromRoutingInfo(routingInfo)
			})

			Context("when lattice app routes are present in the routing info", func() {
				BeforeEach(func() {
					routingInfo = routes.RoutingInfo()
				})

				It("returns the routes", func() {
					Expect(routes).To(Equal(routesResult))
				})
			})

			Context("when the result should be nil", func() {
				itReturnsNilRoutes := func() {
					It("returns nil routes", func() {
						Expect(routesResult).To(BeNil())
					})
				}

				Context("when the lattice routes are nil", func() {
					BeforeEach(func() {
						routingInfo = receptor.RoutingInfo{route_helpers.AppRouter: nil}
					})

					itReturnsNilRoutes()
				})

				Context("when lattice app routes are not present in the routing info", func() {
					BeforeEach(func() {
						routingInfo = receptor.RoutingInfo{}
					})

					itReturnsNilRoutes()
				})

				Context("when the routing info is nil", func() {
					BeforeEach(func() {
						routingInfo = nil
					})

					itReturnsNilRoutes()
				})
			})
		})

		Context("when the json.RawMessage is malformed", func() {
			BeforeEach(func() {
				routingInfo = receptor.RoutingInfo{}
				jsonMessage := json.RawMessage(`{"what": "up`)
				routingInfo[route_helpers.AppRouter] = &jsonMessage
			})

			It("panics at the disco", func() {
				appRoutesFromRoutingInfo := func() func() {
					return func() { route_helpers.AppRoutesFromRoutingInfo(routingInfo) }
				}

				Consistently(appRoutesFromRoutingInfo).Should(Panic(), "invalid json.RawMessage ought to panic")
			})
		})
	})

	Describe("HostnamesByPort", func() {
		It("returns map of ports to slice of hostnames", func() {
			expectedHostnamesByPort := map[uint16][]string{
				11111: []string{"foo1.example.com", "bar1.examaple.com"},
				22222: []string{"foo2.example.com", "bar2.examaple.com"},
				33333: []string{"foo3.example.com", "bar3.examaple.com"},
			}

			Expect(routes.HostnamesByPort()).To(Equal(expectedHostnamesByPort))
		})
	})
})
