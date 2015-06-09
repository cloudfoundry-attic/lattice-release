package app_examiner

import (
	"github.com/cloudfoundry/noaa"
	"github.com/cloudfoundry/sonde-go/events"
)

//go:generate counterfeiter -o fake_noaa_consumer/fake_noaa_consumer.go . NoaaConsumer
type NoaaConsumer interface {
	GetContainerMetrics(appGuid, token string) ([]*events.ContainerMetric, error)
}

type noaaConsumer struct {
	consumer *noaa.Consumer
}

func NewNoaaConsumer(consumer *noaa.Consumer) NoaaConsumer {
	return &noaaConsumer{
		consumer: consumer,
	}
}

func (n *noaaConsumer) GetContainerMetrics(appGuid, token string) ([]*events.ContainerMetric, error) {
	return n.consumer.ContainerMetrics(appGuid, token)
}
