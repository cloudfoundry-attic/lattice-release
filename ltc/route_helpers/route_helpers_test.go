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
		appRoute1 route_helpers.AppRoute
		appRoute2 route_helpers.AppRoute
		appRoute3 route_helpers.AppRoute

		appRoutes route_helpers.AppRoutes
	)

	BeforeEach(func() {
		appRoute1 = route_helpers.AppRoute{
			Hostnames: []string{"foo1.example.com", "bar1.examaple.com"},
			Port:      11111,
		}
		appRoute2 = route_helpers.AppRoute{
			Hostnames: []string{"foo2.example.com", "bar2.examaple.com"},
			Port:      22222,
		}
		appRoute3 = route_helpers.AppRoute{
			Hostnames: []string{"foo3.example.com", "bar3.examaple.com"},
			Port:      33333,
		}

		appRoutes = route_helpers.AppRoutes{appRoute1, appRoute2, appRoute3}

	})

	Describe("AppRoutes", func() {
		Describe("RoutingInfo", func() {
			var routingInfo receptor.RoutingInfo

			JustBeforeEach(func() {
				routingInfo = appRoutes.RoutingInfo()
			})

			It("wraps the serialized routes with the correct key", func() {
				expectedBytes, err := json.Marshal(appRoutes)
				Expect(err).ToNot(HaveOccurred())

				payload, err := routingInfo[route_helpers.AppRouter].MarshalJSON()
				Expect(err).ToNot(HaveOccurred())

				Expect(payload).To(MatchJSON(expectedBytes))
			})

			Context("when AppRoutes is empty", func() {
				BeforeEach(func() {
					appRoutes = route_helpers.AppRoutes{}
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
					routingInfo = appRoutes.RoutingInfo()
				})

				It("returns the routes", func() {
					Expect(appRoutes).To(Equal(routesResult))
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

			Expect(appRoutes.HostnamesByPort()).To(Equal(expectedHostnamesByPort))
		})
	})

	Describe("Routes", func() {

		var (
			routes    route_helpers.Routes
			tcpRoute1 route_helpers.TcpRoute
			tcpRoute2 route_helpers.TcpRoute
		)

		BeforeEach(func() {

			tcpRoute1 = route_helpers.TcpRoute{
				ExternalPort: 50000,
				Port:         5222,
			}
			tcpRoute2 = route_helpers.TcpRoute{
				ExternalPort: 51000,
				Port:         5223,
			}
			routes = route_helpers.Routes{
				AppRoutes: route_helpers.AppRoutes{
					appRoute1, appRoute2,
				},
				TcpRoutes: route_helpers.TcpRoutes{
					tcpRoute1, tcpRoute2,
				},
			}
		})

		Describe("RoutingInfo", func() {
			var routingInfo receptor.RoutingInfo

			JustBeforeEach(func() {
				routingInfo = routes.RoutingInfo()
			})

			It("wraps the serialized routes with the correct key", func() {
				expectedBytes, err := json.Marshal(route_helpers.AppRoutes{appRoute1, appRoute2})
				Expect(err).ToNot(HaveOccurred())

				payload, err := routingInfo[route_helpers.AppRouter].MarshalJSON()
				Expect(err).ToNot(HaveOccurred())

				Expect(payload).To(MatchJSON(expectedBytes))

				expectedBytes, err = json.Marshal(route_helpers.TcpRoutes{tcpRoute1, tcpRoute2})
				Expect(err).ToNot(HaveOccurred())

				payload, err = routingInfo[route_helpers.TcpRouter].MarshalJSON()
				Expect(err).ToNot(HaveOccurred())

				Expect(payload).To(MatchJSON(expectedBytes))
			})

			Context("when Routes is empty", func() {
				BeforeEach(func() {
					routes = route_helpers.Routes{
						AppRoutes: route_helpers.AppRoutes{},
						TcpRoutes: route_helpers.TcpRoutes{},
					}
				})

				It("marshals an empty list", func() {
					payload, err := routingInfo[route_helpers.AppRouter].MarshalJSON()
					Expect(err).ToNot(HaveOccurred())

					Expect(payload).To(MatchJSON(`[]`))

					payload, err = routingInfo[route_helpers.TcpRouter].MarshalJSON()
					Expect(err).ToNot(HaveOccurred())

					Expect(payload).To(MatchJSON(`[]`))
				})
			})
		})

		Describe("RoutesFromRoutingInfo", func() {
			var routingInfo receptor.RoutingInfo

			Context("when the method returns a value", func() {
				var routesResult route_helpers.Routes

				JustBeforeEach(func() {
					routesResult = route_helpers.RoutesFromRoutingInfo(routingInfo)
				})

				Context("when lattice app routes are present in the routing info", func() {
					BeforeEach(func() {
						routingInfo = routes.RoutingInfo()
					})

					It("returns the routes", func() {
						Expect(routesResult).To(Equal(routes))
					})
				})

				Context("when the result should be nil", func() {
					itReturnsEmptyRoutes := func() {
						It("returns nil routes", func() {
							Expect(routesResult).To(BeZero())
						})
					}

					Context("when the both http and tcp routes are nil", func() {
						BeforeEach(func() {
							routingInfo = receptor.RoutingInfo{
								route_helpers.AppRouter: nil,
								route_helpers.TcpRouter: nil,
							}
						})

						itReturnsEmptyRoutes()
					})

					Context("when lattice app and tcp routes are not present in the routing info", func() {
						BeforeEach(func() {
							routingInfo = receptor.RoutingInfo{}
						})

						itReturnsEmptyRoutes()
					})

					Context("when the routing info is nil", func() {
						BeforeEach(func() {
							routingInfo = nil
						})

						itReturnsEmptyRoutes()
					})

					Context("when the http routes are nil", func() {
						var onlyTcpRoutes route_helpers.Routes
						BeforeEach(func() {
							onlyTcpRoutes = route_helpers.Routes{
								TcpRoutes: route_helpers.TcpRoutes{tcpRoute1},
							}
							routingInfo = onlyTcpRoutes.RoutingInfo()
						})

						It("returns only tcp routes", func() {
							Expect(routesResult).To(Equal(onlyTcpRoutes))
						})
					})

					Context("when the tcp routes are nil", func() {
						var onlyHttpRoutes route_helpers.Routes
						BeforeEach(func() {
							onlyHttpRoutes = route_helpers.Routes{
								AppRoutes: route_helpers.AppRoutes{appRoute1},
							}
							routingInfo = onlyHttpRoutes.RoutingInfo()
						})

						It("returns only http routes", func() {
							Expect(routesResult).To(Equal(onlyHttpRoutes))
						})
					})
				})
			})

			Context("when the json.RawMessage is malformed", func() {
				BeforeEach(func() {
					routingInfo = receptor.RoutingInfo{}
					jsonMessage := json.RawMessage(`{"what": "up`)
					routingInfo[route_helpers.TcpRouter] = &jsonMessage
				})

				It("panics", func() {
					routesFromRoutingInfo := func() func() {
						return func() { route_helpers.RoutesFromRoutingInfo(routingInfo) }
					}
					Consistently(routesFromRoutingInfo).Should(Panic(), "invalid json.RawMessage ought to panic")
				})
			})
		})
	})
})
