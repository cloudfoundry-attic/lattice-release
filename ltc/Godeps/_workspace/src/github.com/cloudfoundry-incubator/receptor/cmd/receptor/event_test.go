package main_test

import (
	"encoding/json"
	"time"

	"github.com/cloudfoundry-incubator/receptor"
	"github.com/cloudfoundry-incubator/receptor/serialization"
	"github.com/cloudfoundry-incubator/runtime-schema/bbs/bbserrors"
	"github.com/cloudfoundry-incubator/runtime-schema/models"
	"github.com/tedsuo/ifrit/ginkgomon"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
)

var _ = Describe("Event", func() {
	var eventSource receptor.EventSource
	var events chan receptor.Event
	var done chan struct{}
	var desiredLRP models.DesiredLRP

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
		primerLRP := models.DesiredLRP{
			ProcessGuid: "primer-guid",
			Domain:      "primer-domain",
			RootFS:      "primer:rootfs",
			Routes: map[string]*json.RawMessage{
				"router": &rawMessage,
			},
			Action: &models.RunAction{
				Path: "true",
			},
		}

		err = bbs.DesireLRP(logger, primerLRP)
		Expect(err).NotTo(HaveOccurred())

	PRIMING:
		for {
			select {
			case <-events:
				break PRIMING
			case <-time.After(50 * time.Millisecond):
				routeMsg := json.RawMessage([]byte(`{"port":8080,"hosts":["garbage-route"]}`))
				err = bbs.UpdateDesiredLRP(logger, primerLRP.ProcessGuid, models.DesiredLRPUpdate{
					Routes: map[string]*json.RawMessage{
						"router": &routeMsg,
					},
				})
				Expect(err).NotTo(HaveOccurred())
			}
		}

		err = bbs.RemoveDesiredLRPByProcessGuid(logger, primerLRP.ProcessGuid)
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

			desiredLRP = models.DesiredLRP{
				ProcessGuid: "some-guid",
				Domain:      "some-domain",
				RootFS:      "some:rootfs",
				Routes:      routes,
				Action: &models.RunAction{
					Path: "true",
				},
			}
		})

		It("receives events", func() {
			By("creating a DesiredLRP")
			err := bbs.DesireLRP(logger, desiredLRP)
			Expect(err).NotTo(HaveOccurred())

			desiredLRP, err := bbs.DesiredLRPByProcessGuid(desiredLRP.ProcessGuid)
			Expect(err).NotTo(HaveOccurred())

			var event receptor.Event
			Eventually(events).Should(Receive(&event))

			desiredLRPCreatedEvent, ok := event.(receptor.DesiredLRPCreatedEvent)
			Expect(ok).To(BeTrue())

			Expect(desiredLRPCreatedEvent.DesiredLRPResponse).To(Equal(serialization.DesiredLRPToResponse(desiredLRP)))

			By("updating an existing DesiredLRP")
			routeMessage := json.RawMessage([]byte(`[{"port":8080,"hostnames":["new-route"]}]`))
			newRoutes := map[string]*json.RawMessage{
				"cf-router": &routeMessage,
			}
			err = bbs.UpdateDesiredLRP(logger, desiredLRP.ProcessGuid, models.DesiredLRPUpdate{Routes: newRoutes})
			Expect(err).NotTo(HaveOccurred())

			Eventually(events).Should(Receive(&event))

			desiredLRPChangedEvent, ok := event.(receptor.DesiredLRPChangedEvent)
			Expect(ok).To(BeTrue())
			Expect(desiredLRPChangedEvent.After.Routes).To(Equal(receptor.RoutingInfo(newRoutes)))

			By("removing the DesiredLRP")
			err = bbs.RemoveDesiredLRPByProcessGuid(logger, desiredLRP.ProcessGuid)
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
			instanceKey    models.ActualLRPInstanceKey
			newInstanceKey models.ActualLRPInstanceKey
			netInfo        models.ActualLRPNetInfo
		)

		BeforeEach(func() {
			desiredLRP = models.DesiredLRP{
				ProcessGuid: processGuid,
				Domain:      domain,
				RootFS:      "some:rootfs",
				Instances:   1,
				Action: &models.RunAction{
					Path: "true",
				},
			}

			key = models.NewActualLRPKey(processGuid, 0, domain)
			instanceKey = models.NewActualLRPInstanceKey("instance-guid", "cell-id")
			newInstanceKey = models.NewActualLRPInstanceKey("other-instance-guid", "other-cell-id")
			netInfo = models.NewActualLRPNetInfo("1.1.1.1", []models.PortMapping{})
		})

		It("receives events", func() {
			By("creating a ActualLRP")
			err := bbs.DesireLRP(logger, desiredLRP)
			Expect(err).NotTo(HaveOccurred())

			actualLRPGroup, err := bbs.ActualLRPGroupByProcessGuidAndIndex(desiredLRP.ProcessGuid, 0)
			Expect(err).NotTo(HaveOccurred())
			actualLRP := *actualLRPGroup.Instance

			var event receptor.Event
			Eventually(func() receptor.Event {
				Eventually(events).Should(Receive(&event))
				return event
			}).Should(BeAssignableToTypeOf(receptor.ActualLRPCreatedEvent{}))

			actualLRPCreatedEvent := event.(receptor.ActualLRPCreatedEvent)
			Expect(actualLRPCreatedEvent.ActualLRPResponse).To(Equal(serialization.ActualLRPToResponse(actualLRP, false)))

			By("updating the existing ActualLR")
			err = bbs.ClaimActualLRP(logger, key, instanceKey)
			Expect(err).NotTo(HaveOccurred())

			before := actualLRP
			actualLRPGroup, err = bbs.ActualLRPGroupByProcessGuidAndIndex(desiredLRP.ProcessGuid, 0)
			Expect(err).NotTo(HaveOccurred())
			actualLRP = *actualLRPGroup.Instance

			Eventually(func() receptor.Event {
				Eventually(events).Should(Receive(&event))
				return event
			}).Should(BeAssignableToTypeOf(receptor.ActualLRPChangedEvent{}))

			actualLRPChangedEvent := event.(receptor.ActualLRPChangedEvent)
			Expect(actualLRPChangedEvent.Before).To(Equal(serialization.ActualLRPToResponse(before, false)))
			Expect(actualLRPChangedEvent.After).To(Equal(serialization.ActualLRPToResponse(actualLRP, false)))

			By("evacuating the ActualLRP")
			_, err = bbs.EvacuateRunningActualLRP(logger, key, instanceKey, netInfo, 0)
			Expect(err).To(Equal(bbserrors.ErrServiceUnavailable))

			evacuatingLRP, err := bbs.EvacuatingActualLRPByProcessGuidAndIndex(desiredLRP.ProcessGuid, 0)
			Expect(err).NotTo(HaveOccurred())

			Eventually(func() receptor.Event {
				Eventually(events).Should(Receive(&event))
				return event
			}).Should(BeAssignableToTypeOf(receptor.ActualLRPCreatedEvent{}))

			actualLRPCreatedEvent = event.(receptor.ActualLRPCreatedEvent)
			Expect(actualLRPCreatedEvent.ActualLRPResponse).To(Equal(serialization.ActualLRPToResponse(evacuatingLRP, true)))

			// discard instance -> UNCLAIMED
			Eventually(func() receptor.Event {
				Eventually(events).Should(Receive(&event))
				return event
			}).Should(BeAssignableToTypeOf(receptor.ActualLRPChangedEvent{}))

			By("starting and then evacuating the ActualLRP on another cell")
			err = bbs.StartActualLRP(logger, key, newInstanceKey, netInfo)
			Expect(err).NotTo(HaveOccurred())

			// discard instance -> RUNNING
			Eventually(func() receptor.Event {
				Eventually(events).Should(Receive(&event))
				return event
			}).Should(BeAssignableToTypeOf(receptor.ActualLRPChangedEvent{}))

			evacuatingBefore := evacuatingLRP
			_, err = bbs.EvacuateRunningActualLRP(logger, key, newInstanceKey, netInfo, 0)
			Expect(err).To(Equal(bbserrors.ErrServiceUnavailable))

			evacuatingLRP, err = bbs.EvacuatingActualLRPByProcessGuidAndIndex(desiredLRP.ProcessGuid, 0)
			Expect(err).NotTo(HaveOccurred())

			Eventually(func() receptor.Event {
				Eventually(events).Should(Receive(&event))
				return event
			}).Should(BeAssignableToTypeOf(receptor.ActualLRPChangedEvent{}))

			actualLRPChangedEvent = event.(receptor.ActualLRPChangedEvent)
			Expect(actualLRPChangedEvent.Before).To(Equal(serialization.ActualLRPToResponse(evacuatingBefore, true)))
			Expect(actualLRPChangedEvent.After).To(Equal(serialization.ActualLRPToResponse(evacuatingLRP, true)))

			// discard instance -> UNCLAIMED
			Eventually(func() receptor.Event {
				Eventually(events).Should(Receive(&event))
				return event
			}).Should(BeAssignableToTypeOf(receptor.ActualLRPChangedEvent{}))

			By("removing the instance ActualLRP")
			actualLRPGroup, err = bbs.ActualLRPGroupByProcessGuidAndIndex(desiredLRP.ProcessGuid, 0)
			Expect(err).NotTo(HaveOccurred())
			actualLRP = *actualLRPGroup.Instance

			err = bbs.RemoveActualLRP(logger, key, models.ActualLRPInstanceKey{})
			Expect(err).NotTo(HaveOccurred())

			Eventually(func() receptor.Event {
				Eventually(events).Should(Receive(&event))
				return event
			}).Should(BeAssignableToTypeOf(receptor.ActualLRPRemovedEvent{}))

			actualLRPRemovedEvent := event.(receptor.ActualLRPRemovedEvent)
			Expect(actualLRPRemovedEvent.ActualLRPResponse).To(Equal(serialization.ActualLRPToResponse(actualLRP, false)))

			By("removing the evacuating ActualLRP")
			err = bbs.RemoveEvacuatingActualLRP(logger, key, newInstanceKey)
			Expect(err).NotTo(HaveOccurred())

			Eventually(func() receptor.Event {
				Eventually(events).Should(Receive(&event))
				return event
			}).Should(BeAssignableToTypeOf(receptor.ActualLRPRemovedEvent{}))

			Expect(event).To(BeAssignableToTypeOf(receptor.ActualLRPRemovedEvent{}))
			actualLRPRemovedEvent = event.(receptor.ActualLRPRemovedEvent)
			Expect(actualLRPRemovedEvent.ActualLRPResponse).To(Equal(serialization.ActualLRPToResponse(evacuatingLRP, true)))
		})
	})
})
