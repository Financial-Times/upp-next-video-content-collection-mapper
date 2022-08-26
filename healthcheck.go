package main

import (
	"net/http"
	"time"

	fthealth "github.com/Financial-Times/go-fthealth/v1_1"
	"github.com/Financial-Times/kafka-client-go/v3"
	"github.com/Financial-Times/service-status-go/gtg"
)

type messageConsumerHealthcheck interface {
	ConnectivityCheck() error
	MonitorCheck() error
}

type messageProducerHealthcheck interface {
	ConnectivityCheck() error
}
type HealthCheck struct {
	consumer      messageConsumerHealthcheck
	producer      messageProducerHealthcheck
	appSystemCode string
	appName       string
	panicGuide    string
}

func NewHealthCheck(p messageProducerHealthcheck, c messageConsumerHealthcheck, appName, appSystemCode, panicGuide string) *HealthCheck {
	return &HealthCheck{
		consumer:      c,
		producer:      p,
		appName:       appName,
		appSystemCode: appSystemCode,
		panicGuide:    panicGuide,
	}
}

func (h *HealthCheck) Health() func(w http.ResponseWriter, r *http.Request) {
	checks := []fthealth.Check{h.readQueueCheck(), h.readQueueLagCheck(), h.writeQueueCheck()}
	hc := fthealth.TimedHealthCheck{
		HealthCheck: fthealth.HealthCheck{
			SystemCode:  h.appSystemCode,
			Name:        h.appName,
			Description: serviceDescription,
			Checks:      checks,
		},
		Timeout: 10 * time.Second,
	}
	return fthealth.Handler(hc)
}

func (h *HealthCheck) readQueueCheck() fthealth.Check {
	return fthealth.Check{
		ID:               "read-message-queue-reachable",
		Name:             "Read Message Queue Reachable",
		Severity:         2,
		BusinessImpact:   "Related content from published Next videos will not be processed, clients will not see them within content.",
		TechnicalSummary: "Read message queue is not reachable/healthy",
		PanicGuide:       h.panicGuide,
		Checker:          h.checkIfKafkaIsReachableFromConsumer,
	}
}

func (h *HealthCheck) writeQueueCheck() fthealth.Check {
	return fthealth.Check{
		ID:               "write-message-queue-reachable",
		Name:             "Write Message Queue Reachable",
		Severity:         2,
		BusinessImpact:   "Related content from published Next videos will not be processed, clients will not see them within content.",
		TechnicalSummary: "Write message queue is not reachable/healthy",
		PanicGuide:       h.panicGuide,
		Checker:          h.checkIfKafkaIsReachableFromProducer,
	}
}

func (h *HealthCheck) readQueueLagCheck() fthealth.Check {
	return fthealth.Check{
		ID:               "read-message-queue-lagging",
		Name:             "Read Message Queue Is Not Lagging",
		Severity:         3,
		BusinessImpact:   "Related content from published Next videos will be processed with latency.",
		TechnicalSummary: kafka.LagTechnicalSummary,
		PanicGuide:       h.panicGuide,
		Checker:          h.checkIfConsumerIsLagging,
	}
}

func (h *HealthCheck) GTG() gtg.Status {
	consumerCheck := func() gtg.Status {
		return gtgCheck(h.checkIfKafkaIsReachableFromConsumer)
	}
	producerCheck := func() gtg.Status {
		return gtgCheck(h.checkIfKafkaIsReachableFromProducer)
	}

	return gtg.FailFastParallelCheck([]gtg.StatusChecker{
		consumerCheck,
		producerCheck,
	})()
}

func gtgCheck(handler func() (string, error)) gtg.Status {
	if _, err := handler(); err != nil {
		return gtg.Status{GoodToGo: false, Message: err.Error()}
	}
	return gtg.Status{GoodToGo: true}
}

func (h *HealthCheck) checkIfKafkaIsReachableFromConsumer() (string, error) {
	err := h.consumer.ConnectivityCheck()
	if err != nil {
		return "", err
	}
	return "OK", nil
}

func (h *HealthCheck) checkIfConsumerIsLagging() (string, error) {
	err := h.consumer.MonitorCheck()
	if err != nil {
		return "", err
	}
	return "OK", nil
}

func (h *HealthCheck) checkIfKafkaIsReachableFromProducer() (string, error) {
	err := h.producer.ConnectivityCheck()
	if err != nil {
		return "", err
	}
	return "OK", nil
}
