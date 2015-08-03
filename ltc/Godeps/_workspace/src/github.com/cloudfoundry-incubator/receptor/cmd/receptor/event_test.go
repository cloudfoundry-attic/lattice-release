package main_test

import (
	"encoding/json"
	"time"

	"github.com/cloudfoundry-incubator/bbs/models"
	"github.com/cloudfoundry-incubator/receptor"
	"github.com/cloudfoundry-incubator/receptor/serialization"
	"github.com/cloudfoundry-incubator/runtime-schema/bbs/bbserrors"
	oldmodels "github.com/cloudfoundry-incubator/runtime-schema/models"
	"github.com/tedsuo/ifrit/ginkgomon"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
)

var _ = Describe("Event", func() {
	var eventSource receptor.EventSource
	var events chan receptor.Event
	var done chan struct{}
	var oldDesiredLRP oldmodels.DesiredLRP

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
		primerLRP := oldmodels.DesiredLRP{
			ProcessGuid: "primer-guid",
			Domain:      "primer-domain",
			RootFS:      "primer:rootfs",
			Routes: map[string]*json.RawMessage{
				"router": &rawMessage,
			},
			Action: &oldmodels.RunAction{
				User: "me",
				Path: "true",
			},
		}

		err = legacyBBS.DesireLRP(logger, primerLRP)
		Expect(err).NotTo(HaveOccurred())

	PRIMING:
		for {
			select {
			case <-events:
				break PRIMING
			case <-time.After(50 * time.Millisecond):
				routeMsg := json.RawMessage([]byte(`{"port":8080,"hosts":["garbage-route"]}`))
				err = legacyBBS.UpdateDesiredLRP(logger, primerLRP.ProcessGuid, oldmodels.DesiredLRPUpdate{
					Routes: map[string]*json.RawMessage{
						"router": &routeMsg,
					},
				})
				Expect(err).NotTo(HaveOccurred())
			}
		}

		err = legacyBBS.RemoveDesiredLRPByProcessGuid(logger, primerLRP.ProcessGuid)
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
			routes := map[string]*json.RawMessage{"cf-router": &routeMessage}

			oldDesiredLRP = oldmodels.DesiredLRP{
				ProcessGuid: "some-guid",
				Domain:      "some-domain",
				RootFS:      "some:rootfs",
				Routes:      routes,
				Action: &oldmodels.RunAction{
					User:      "me",
					Dir:       "/tmp",
					Path:      "true",
					LogSource: "logs",
				},
			}
		})

		It("receives events", func() {
			By("creating a DesiredLRP")
			err := legacyBBS.DesireLRP(logger, oldDesiredLRP)
			Expect(err).NotTo(HaveOccurred())

			desiredLRP, err := bbsClient.DesiredLRPByProcessGuid(oldDesiredLRP.ProcessGuid)
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
			newRoutes := map[string]*json.RawMessage{
				"cf-router": &routeMessage,
			}
			err = legacyBBS.UpdateDesiredLRP(logger, oldDesiredLRP.ProcessGuid, oldmodels.DesiredLRPUpdate{Routes: newRoutes})
			Expect(err).NotTo(HaveOccurred())

			Eventually(events).Should(Receive(&event))

			desiredLRPChangedEvent, ok := event.(receptor.DesiredLRPChangedEvent)
			Expect(ok).To(BeTrue())
			Expect(desiredLRPChangedEvent.After.Routes).To(Equal(receptor.RoutingInfo(newRoutes)))

			By("removing the DesiredLRP")
			err = legacyBBS.RemoveDesiredLRPByProcessGuid(logger, oldDesiredLRP.ProcessGuid)
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

		var (
			legacyKey            oldmodels.ActualLRPKey
			legacyOldInstanceKey oldmodels.ActualLRPInstanceKey
			legacyNewInstanceKey oldmodels.ActualLRPInstanceKey
			legacyNetInfo        oldmodels.ActualLRPNetInfo
		)

		BeforeEach(func() {
			oldDesiredLRP = oldmodels.DesiredLRP{
				ProcessGuid: processGuid,
				Domain:      domain,
				RootFS:      "some:rootfs",
				Instances:   1,
				Action: &oldmodels.RunAction{
					Path: "true",
					User: "me",
				},
			}

			legacyKey = oldmodels.NewActualLRPKey(processGuid, 0, domain)
			legacyOldInstanceKey = oldmodels.NewActualLRPInstanceKey("instance-guid", "cell-id")
			legacyNewInstanceKey = oldmodels.NewActualLRPInstanceKey("other-instance-guid", "other-cell-id")
			legacyNetInfo = oldmodels.NewActualLRPNetInfo("1.1.1.1", []oldmodels.PortMapping{})

			key = models.NewActualLRPKey(processGuid, 0, domain)
			oldInstanceKey = models.NewActualLRPInstanceKey("instance-guid", "cell-id")
			newInstanceKey = models.NewActualLRPInstanceKey("other-instance-guid", "other-cell-id")
			netInfo = models.NewActualLRPNetInfo("1.1.1.1")
		})

		It("receives events", func() {
			By("creating a ActualLRP")
			err := legacyBBS.DesireLRP(logger, oldDesiredLRP)
			Expect(err).NotTo(HaveOccurred())

			actualLRPGroup, err := bbsClient.ActualLRPGroupByProcessGuidAndIndex(oldDesiredLRP.ProcessGuid, 0)
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
			_, err = bbsClient.ClaimActualLRP(processGuid, 0, &oldInstanceKey)
			Expect(err).NotTo(HaveOccurred())

			before := actualLRP
			actualLRPGroup, err = bbsClient.ActualLRPGroupByProcessGuidAndIndex(oldDesiredLRP.ProcessGuid, 0)
			Expect(err).NotTo(HaveOccurred())
			actualLRP = actualLRPGroup.Instance

			Eventually(func() receptor.Event {
				Eventually(events).Should(Receive(&event))
				return event
			}).Should(BeAssignableToTypeOf(receptor.ActualLRPChangedEvent{}))

			actualLRPChangedEvent := event.(receptor.ActualLRPChangedEvent)
			Expect(actualLRPChangedEvent.Before).To(Equal(serialization.ActualLRPProtoToResponse(before, false)))
			Expect(actualLRPChangedEvent.After).To(Equal(serialization.ActualLRPProtoToResponse(actualLRP, false)))

			By("evacuating the ActualLRP")
			_, err = legacyBBS.EvacuateRunningActualLRP(logger, legacyKey, legacyOldInstanceKey, legacyNetInfo, 0)
			Expect(err).To(Equal(bbserrors.ErrServiceUnavailable))

			evacuatingLRPGroup, err := bbsClient.ActualLRPGroupByProcessGuidAndIndex(oldDesiredLRP.ProcessGuid, 0)
			Expect(err).NotTo(HaveOccurred())
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
			_, err = bbsClient.StartActualLRP(&key, &newInstanceKey, &netInfo)
			Expect(err).NotTo(HaveOccurred())

			// discard instance -> RUNNING
			Eventually(func() receptor.Event {
				Eventually(events).Should(Receive(&event))
				return event
			}).Should(BeAssignableToTypeOf(receptor.ActualLRPChangedEvent{}))

			evacuatingBefore := evacuatingLRP
			_, err = legacyBBS.EvacuateRunningActualLRP(logger, legacyKey, legacyNewInstanceKey, legacyNetInfo, 0)
			Expect(err).To(Equal(bbserrors.ErrServiceUnavailable))

			evacuatingLRPGroup, err = bbsClient.ActualLRPGroupByProcessGuidAndIndex(oldDesiredLRP.ProcessGuid, 0)
			Expect(err).NotTo(HaveOccurred())
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
			actualLRPGroup, err = bbsClient.ActualLRPGroupByProcessGuidAndIndex(oldDesiredLRP.ProcessGuid, 0)
			Expect(err).NotTo(HaveOccurred())
			actualLRP = actualLRPGroup.Instance

			err = bbsClient.RemoveActualLRP(key.ProcessGuid, legacyKey.Index)
			Expect(err).NotTo(HaveOccurred())

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
			err = legacyBBS.RemoveEvacuatingActualLRP(logger, legacyKey, legacyNewInstanceKey)
			Expect(err).NotTo(HaveOccurred())

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
