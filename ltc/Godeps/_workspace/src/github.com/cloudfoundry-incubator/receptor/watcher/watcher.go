package watcher

import (
	"os"
	"sync"
	"time"

	"github.com/cloudfoundry-incubator/receptor"
	"github.com/cloudfoundry-incubator/receptor/event"
	"github.com/cloudfoundry-incubator/receptor/serialization"
	"github.com/cloudfoundry-incubator/runtime-schema/bbs"
	"github.com/cloudfoundry-incubator/runtime-schema/models"
	"github.com/pivotal-golang/clock"
	"github.com/pivotal-golang/lager"
	"github.com/tedsuo/ifrit"
)

type Watcher ifrit.Runner

type watcher struct {
	bbs               bbs.ReceptorBBS
	hub               event.Hub
	clock             clock.Clock
	retryWaitDuration time.Duration
	logger            lager.Logger
}

func NewWatcher(
	bbs bbs.ReceptorBBS,
	hub event.Hub,
	clock clock.Clock,
	retryWaitDuration time.Duration,
	logger lager.Logger,
) Watcher {
	return &watcher{
		bbs:               bbs,
		hub:               hub,
		clock:             clock,
		retryWaitDuration: retryWaitDuration,
		logger:            logger,
	}
}

func (w *watcher) Run(signals <-chan os.Signal, ready chan<- struct{}) error {
	logger := w.logger.Session("running-watcher")
	logger.Info("starting")

	hubNotification := make(chan int, 1)
	hubSize := 0

	logger.Info("registering-callback-with-hub")
	w.hub.RegisterCallback(func(size int) {
		hubNotification <- size
	})
	logger.Info("registered-callback-with-hub")

	close(ready)
	logger.Info("started")
	defer logger.Info("finished")

	var desiredStop, actualStop chan<- bool
	var desiredErrors, actualErrors <-chan error

	reWatchTimerDesired := w.clock.NewTimer(w.retryWaitDuration)
	defer reWatchTimerDesired.Stop()
	reWatchTimerDesired.Stop()

	reWatchTimerActual := w.clock.NewTimer(w.retryWaitDuration)
	defer reWatchTimerActual.Stop()
	reWatchTimerActual.Stop()

	var reWatchActual <-chan time.Time
	var reWatchDesired <-chan time.Time

	for {
		select {
		case hubSize = <-hubNotification:
			if hubSize == 0 {
				if desiredStop != nil {
					logger.Info("stopping-desired-watch-from-hub-notification")
					desiredStop <- true
					desiredStop = nil
					desiredErrors = nil
				}
				if actualStop != nil {
					logger.Info("stopping-actual-watch-from-hub-notification")
					actualStop <- true
					actualStop = nil
					actualErrors = nil
				}
			} else {
				wg := sync.WaitGroup{}

				if desiredStop == nil {
					logger.Info("rewatching-desired-from-hub-notification")

					wg.Add(1)
					go func() {
						defer wg.Done()
						desiredStop, desiredErrors = w.watchDesired(logger)
						logger.Debug("finished-rewatching-desired-from-hub-notification")
					}()
				}

				if actualStop == nil {
					logger.Info("rewatching-actual-from-hub-notification")

					wg.Add(1)
					go func() {
						defer wg.Done()
						actualStop, actualErrors = w.watchActual(logger)
						logger.Debug("finished-rewatching-actual-from-hub-notification")
					}()
				}

				wg.Wait()
			}

		case err, ok := <-desiredErrors:
			if ok {
				reWatchTimerDesired.Reset(w.retryWaitDuration)
				reWatchDesired = reWatchTimerDesired.C()
			}
			if err != nil {
				logger.Error("desired-watch-failed", err)
			}
			desiredErrors = nil
			desiredStop = nil

		case err, ok := <-actualErrors:
			if ok {
				reWatchTimerActual.Reset(w.retryWaitDuration)
				reWatchActual = reWatchTimerActual.C()
			}
			if err != nil {
				logger.Error("actual-watch-failed", err)
			}
			actualErrors = nil
			actualStop = nil

		case <-reWatchDesired:
			reWatchDesired = nil

			if desiredStop == nil && hubSize > 0 {
				logger.Info("rewatching-desired")
				desiredStop, desiredErrors = w.watchDesired(logger)
			}

		case <-reWatchActual:
			reWatchActual = nil
			if actualStop == nil && hubSize > 0 {
				logger.Info("rewatching-actual")
				actualStop, actualErrors = w.watchActual(logger)
			}

		case <-signals:
			logger.Info("stopping")
			if desiredStop != nil {
				desiredStop <- true
				desiredStop = nil
			}
			if actualStop != nil {
				actualStop <- true
				actualStop = nil
			}
			return nil
		}
	}

	return nil
}

func (w *watcher) watchDesired(logger lager.Logger) (chan<- bool, <-chan error) {
	return w.bbs.WatchForDesiredLRPChanges(logger,
		func(created models.DesiredLRP) {
			logger.Debug("handling-desired-create")
			w.hub.Emit(receptor.NewDesiredLRPCreatedEvent(serialization.DesiredLRPToResponse(created)))
		},
		func(changed models.DesiredLRPChange) {
			logger.Debug("handling-desired-change")
			w.hub.Emit(receptor.NewDesiredLRPChangedEvent(
				serialization.DesiredLRPToResponse(changed.Before),
				serialization.DesiredLRPToResponse(changed.After),
			))
		},
		func(deleted models.DesiredLRP) {
			logger.Debug("handling-desired-delete")
			w.hub.Emit(receptor.NewDesiredLRPRemovedEvent(serialization.DesiredLRPToResponse(deleted)))
		})
}

func (w *watcher) watchActual(logger lager.Logger) (chan<- bool, <-chan error) {
	return w.bbs.WatchForActualLRPChanges(logger,
		func(created models.ActualLRP, evacuating bool) {
			logger.Debug("handling-actual-create")
			w.hub.Emit(receptor.NewActualLRPCreatedEvent(serialization.ActualLRPToResponse(created, evacuating)))
		},
		func(changed models.ActualLRPChange, evacuating bool) {
			logger.Debug("handling-actual-change")
			w.hub.Emit(receptor.NewActualLRPChangedEvent(
				serialization.ActualLRPToResponse(changed.Before, evacuating),
				serialization.ActualLRPToResponse(changed.After, evacuating),
			))
		},
		func(deleted models.ActualLRP, evacuating bool) {
			logger.Debug("handling-actual-delete")
			w.hub.Emit(receptor.NewActualLRPRemovedEvent(serialization.ActualLRPToResponse(deleted, evacuating)))
		})
}

