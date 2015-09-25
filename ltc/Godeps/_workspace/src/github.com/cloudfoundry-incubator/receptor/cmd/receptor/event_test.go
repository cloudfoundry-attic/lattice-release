package main_test

import (
	"encoding/json"
	"time"

	"github.com/cloudfoundry-incubator/bbs/models"
	"github.com/cloudfoundry-incubator/receptor"
	"github.com/cloudfoundry-incubator/receptor/serialization"
	"github.com/tedsuo/ifrit/ginkgomon"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
)

var _ = Describe("Event", func() {
	var (
		eventSource receptor.EventSource
		events      chan receptor.Event
		done        chan struct{}
		desiredLRP  *models.DesiredLRP
	)

	JustBeforeEach(func() {
		receptorProcess = ginkgomon.Invoke(receptorRunner)

		var err error
		eventSource, err = client.SubscribeToEvents()
		Expect(err).NotTo(HaveOccurred())

		events = make(chan receptor.Event)
		done = make(chan struct{})

		go func() {
			defer close(done)
			for {
				event, err := eventSource.Next()
				if err != nil {
					close(events)
					return
				}
				events <- event
			}
		}()

		rawMessage := json.RawMessage([]byte(`{"port":8080,"hosts":["primer-route"]}`))
		primerLRP := &models.DesiredLRP{
			ProcessGuid: "primer-guid",
			Domain:      "primer-domain",
			RootFs:      "primer:rootfs",
			Routes: &models.Routes{
				"router": &rawMessage,
			},
			Action: models.WrapAction(&models.RunAction{
				User: "me",
				Path: "true",
			}),
		}

		err = bbsClient.DesireLRP(primerLRP)
		Expect(err).NotTo(HaveOccurred())

	PRIMING:
		for {
			select {
			case <-events:
				break PRIMING
			case <-time.After(50 * time.Millisecond):
				routeMsg := json.RawMessage([]byte(`{"port":8080,"hosts":["garbage-route"]}`))
				err = bbsClient.UpdateDesiredLRP(primerLRP.ProcessGuid, &models.DesiredLRPUpdate{
					Routes: &models.Routes{
						"router": &routeMsg,
					},
				})
				Expect(err).NotTo(HaveOccurred())
			}
		}

		err = bbsClient.RemoveDesiredLRP(primerLRP.ProcessGuid)
		Expect(err).NotTo(HaveOccurred())

		var event receptor.Event
		for {
			Eventually(events).Should(Receive(&event))
			if event.EventType() == receptor.EventTypeDesiredLRPRemoved {
				break
			}
		}
	})

	AfterEach(func() {
		ginkgomon.Kill(receptorProcess)
		err := eventSource.Close()
		Expect(err).NotTo(HaveOccurred())
		Eventually(done).Should(BeClosed())
	})

	Describe("Desired LRPs", func() {
		BeforeEach(func() {
			routeMessage := json.RawMessage([]byte(`[{"port":8080,"hostnames":["original-route"]}]`))
			routes := &models.Routes{"cf-router": &routeMessage}

			desiredLRP = &models.DesiredLRP{
				ProcessGuid: "some-guid",
				Domain:      "some-domain",
				RootFs:      "some:rootfs",
				Routes:      routes,
				Action: models.WrapAction(&models.RunAction{
					User:      "me",
					Dir:       "/tmp",
					Path:      "true",
					LogSource: "logs",
				}),
			}
		})

		It("receives events", func() {
			By("creating a DesiredLRP")
			err := bbsClient.DesireLRP(desiredLRP)
			Expect(err).NotTo(HaveOccurred())

			desiredLRP, err := bbsClient.DesiredLRPByProcessGuid(desiredLRP.ProcessGuid)
			Expect(err).NotTo(HaveOccurred())

			var event receptor.Event
			Eventually(events).Should(Receive(&event))

			desiredLRPCreatedEvent, ok := event.(receptor.DesiredLRPCreatedEvent)
			Expect(ok).To(BeTrue())

			actualJSON, _ := json.Marshal(desiredLRPCreatedEvent.DesiredLRPResponse)
			expectedJSON, _ := json.Marshal(serialization.DesiredLRPProtoToResponse(desiredLRP))
			Expect(actualJSON).To(MatchJSON(expectedJSON))

			By("updating an existing DesiredLRP")
			routeMessage := json.RawMessage([]byte(`[{"port":8080,"hostnames":["new-route"]}]`))
			newRoutes := &models.Routes{
				"cf-router": &routeMessage,
			}
			err = bbsClient.UpdateDesiredLRP(desiredLRP.ProcessGuid, &models.DesiredLRPUpdate{Routes: newRoutes})
			Expect(err).NotTo(HaveOccurred())

			Eventually(events).Should(Receive(&event))

			desiredLRPChangedEvent, ok := event.(receptor.DesiredLRPChangedEvent)
			Expect(ok).To(BeTrue())
			Expect(desiredLRPChangedEvent.After.Routes).To(Equal(receptor.RoutingInfo(*newRoutes)))

			By("removing the DesiredLRP")
			err = bbsClient.RemoveDesiredLRP(desiredLRP.ProcessGuid)
			Expect(err).NotTo(HaveOccurred())

			Eventually(events).Should(Receive(&event))

			desiredLRPRemovedEvent, ok := event.(receptor.DesiredLRPRemovedEvent)
			Expect(ok).To(BeTrue())
			Expect(desiredLRPRemovedEvent.DesiredLRPResponse.ProcessGuid).To(Equal(desiredLRP.ProcessGuid))
		})
	})

	Describe("Actual LRPs", func() {
		const (
			processGuid = "some-process-guid"
			domain      = "some-domain"
		)

		var (
			key            models.ActualLRPKey
			oldInstanceKey models.ActualLRPInstanceKey
			newInstanceKey models.ActualLRPInstanceKey
			netInfo        models.ActualLRPNetInfo
		)

		BeforeEach(func() {
			desiredLRP = &models.DesiredLRP{
				ProcessGuid: processGuid,
				Domain:      domain,
				RootFs:      "some:rootfs",
				Instances:   1,
				Action: models.WrapAction(&models.RunAction{
					Path: "true",
					User: "me",
				}),
			}

			key = models.NewActualLRPKey(processGuid, 0, domain)
			oldInstanceKey = models.NewActualLRPInstanceKey("instance-guid", "cell-id")
			newInstanceKey = models.NewActualLRPInstanceKey("other-instance-guid", "other-cell-id")
			netInfo = models.NewActualLRPNetInfo("1.1.1.1")
		})

		It("receives events", func() {
			By("creating a ActualLRP")
			err := bbsClient.DesireLRP(desiredLRP)
			Expect(err).NotTo(HaveOccurred())

			actualLRPGroup, err := bbsClient.ActualLRPGroupByProcessGuidAndIndex(desiredLRP.ProcessGuid, 0)
			Expect(err).NotTo(HaveOccurred())
			actualLRP := actualLRPGroup.Instance

			var event receptor.Event
			Eventually(func() receptor.Event {
				Eventually(events).Should(Receive(&event))
				return event
			}).Should(BeAssignableToTypeOf(receptor.ActualLRPCreatedEvent{}))

			actualLRPCreatedEvent := event.(receptor.ActualLRPCreatedEvent)
			Expect(actualLRPCreatedEvent.ActualLRPResponse).To(Equal(serialization.ActualLRPProtoToResponse(actualLRP, false)))

			By("updating the existing ActualLRP")
			bbsErr := bbsClient.ClaimActualLRP(processGuid, 0, &oldInstanceKey)
			Expect(err).NotTo(HaveOccurred())

			before := actualLRP
			actualLRPGroup, bbsErr = bbsClient.ActualLRPGroupByProcessGuidAndIndex(desiredLRP.ProcessGuid, 0)
			Expect(bbsErr).NotTo(HaveOccurred())
			actualLRP = actualLRPGroup.Instance

			Eventually(func() receptor.Event {
				Eventually(events).Should(Receive(&event))
				return event
			}).Should(BeAssignableToTypeOf(receptor.ActualLRPChangedEvent{}))

			actualLRPChangedEvent := event.(receptor.ActualLRPChangedEvent)
			Expect(actualLRPChangedEvent.Before).To(Equal(serialization.ActualLRPProtoToResponse(before, false)))
			Expect(actualLRPChangedEvent.After).To(Equal(serialization.ActualLRPProtoToResponse(actualLRP, false)))

			By("evacuating the ActualLRP")
			_, bbsErr = bbsClient.EvacuateRunningActualLRP(&key, &oldInstanceKey, &netInfo, 0)
			// This will cause an auction to be submitted. We expect this to fail
			// because there is no auctioneer running.
			Expect(bbsErr).To(Equal(models.ErrUnknownError))

			evacuatingLRPGroup, bbsErr := bbsClient.ActualLRPGroupByProcessGuidAndIndex(desiredLRP.ProcessGuid, 0)
			Expect(bbsErr).NotTo(HaveOccurred())
			evacuatingLRP := evacuatingLRPGroup.Evacuating

			Eventually(func() receptor.Event {
				Eventually(events).Should(Receive(&event))
				return event
			}).Should(BeAssignableToTypeOf(receptor.ActualLRPCreatedEvent{}))

			// this is a necessary hack until we migrate other things to protobufs or pointer structs
			actualLRPCreatedEvent = event.(receptor.ActualLRPCreatedEvent)
			response := actualLRPCreatedEvent.ActualLRPResponse
			response.Ports = nil
			Expect(response).To(Equal(serialization.ActualLRPProtoToResponse(evacuatingLRP, true)))

			// discard instance -> UNCLAIMED
			Eventually(func() receptor.Event {
				Eventually(events).Should(Receive(&event))
				return event
			}).Should(BeAssignableToTypeOf(receptor.ActualLRPChangedEvent{}))

			By("starting and then evacuating the ActualLRP on another cell")
			bbsErr = bbsClient.StartActualLRP(&key, &newInstanceKey, &netInfo)
			Expect(bbsErr).NotTo(HaveOccurred())

			// discard instance -> RUNNING
			Eventually(func() receptor.Event {
				Eventually(events).Should(Receive(&event))
				return event
			}).Should(BeAssignableToTypeOf(receptor.ActualLRPChangedEvent{}))

			evacuatingBefore := evacuatingLRP
			_, bbsErr = bbsClient.EvacuateRunningActualLRP(&key, &newInstanceKey, &netInfo, 0)
			Expect(bbsErr).To(Equal(models.ErrUnknownError))

			evacuatingLRPGroup, bbsErr = bbsClient.ActualLRPGroupByProcessGuidAndIndex(desiredLRP.ProcessGuid, 0)
			Expect(bbsErr).NotTo(HaveOccurred())
			evacuatingLRP = evacuatingLRPGroup.Evacuating

			Expect(err).NotTo(HaveOccurred())

			Eventually(func() receptor.Event {
				Eventually(events).Should(Receive(&event))
				return event
			}).Should(BeAssignableToTypeOf(receptor.ActualLRPChangedEvent{}))

			actualLRPChangedEvent = event.(receptor.ActualLRPChangedEvent)
			response = actualLRPChangedEvent.Before
			response.Ports = nil
			Expect(response).To(Equal(serialization.ActualLRPProtoToResponse(evacuatingBefore, true)))

			response = actualLRPChangedEvent.After
			response.Ports = nil
			Expect(response).To(Equal(serialization.ActualLRPProtoToResponse(evacuatingLRP, true)))

			// discard instance -> UNCLAIMED
			Eventually(func() receptor.Event {
				Eventually(events).Should(Receive(&event))
				return event
			}).Should(BeAssignableToTypeOf(receptor.ActualLRPChangedEvent{}))

			By("removing the instance ActualLRP")
			actualLRPGroup, bbsErr = bbsClient.ActualLRPGroupByProcessGuidAndIndex(desiredLRP.ProcessGuid, 0)
			Expect(bbsErr).NotTo(HaveOccurred())
			actualLRP = actualLRPGroup.Instance

			bbsErr = bbsClient.RemoveActualLRP(key.ProcessGuid, int(key.Index))
			Expect(bbsErr).NotTo(HaveOccurred())

			Eventually(func() receptor.Event {
				Eventually(events).Should(Receive(&event))
				return event
			}).Should(BeAssignableToTypeOf(receptor.ActualLRPRemovedEvent{}))

			// this is a necessary hack until we migrate other things to protobufs or pointer structs
			actualLRPRemovedEvent := event.(receptor.ActualLRPRemovedEvent)
			response = actualLRPRemovedEvent.ActualLRPResponse
			response.Ports = nil
			Expect(response).To(Equal(serialization.ActualLRPProtoToResponse(actualLRP, false)))

			By("removing the evacuating ActualLRP")
			bbsErr = bbsClient.RemoveEvacuatingActualLRP(&key, &newInstanceKey)
			Expect(bbsErr).NotTo(HaveOccurred())

			Eventually(func() receptor.Event {
				Eventually(events).Should(Receive(&event))
				return event
			}).Should(BeAssignableToTypeOf(receptor.ActualLRPRemovedEvent{}))

			Expect(event).To(BeAssignableToTypeOf(receptor.ActualLRPRemovedEvent{}))

			// this is a necessary hack until we migrate other things to protobufs or pointer structs
			actualLRPRemovedEvent = event.(receptor.ActualLRPRemovedEvent)
			response = actualLRPRemovedEvent.ActualLRPResponse
			response.Ports = nil
			Expect(response).To(Equal(serialization.ActualLRPProtoToResponse(evacuatingLRP, true)))
		})
	})
})
