package main

import (
	"net/http"

	fthealth "github.com/Financial-Times/go-fthealth/v1_1"
	"github.com/Financial-Times/message-queue-go-producer/producer"
	consumer "github.com/Financial-Times/message-queue-gonsumer"
	"github.com/Financial-Times/service-status-go/gtg"

	"time"
)

type HealthCheck struct {
	consumer      consumer.MessageConsumer
	producer      producer.MessageProducer
	appSystemCode string
	appName       string
	panicGuide    string
}

func NewHealthCheck(p producer.MessageProducer, c consumer.MessageConsumer, appName, appSystemCode, panicGuide string) *HealthCheck {
	return &HealthCheck{
		consumer:      c,
		producer:      p,
		appName:       appName,
		appSystemCode: appSystemCode,
		panicGuide:    panicGuide,
	}
}

func (h *HealthCheck) Health() func(w http.ResponseWriter, r *http.Request) {
	checks := []fthealth.Check{h.readQueueCheck(), h.writeQueueCheck()}
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
		ID:               "read-message-queue-proxy-reachable",
		Name:             "Read Message Queue Proxy Reachable",
		Severity:         2,
		BusinessImpact:   "Related content from published Next videos will not be processed, clients will not see them within content.",
		TechnicalSummary: "Read message queue proxy is not reachable/healthy",
		PanicGuide:       h.panicGuide,
		Checker:          h.consumer.ConnectivityCheck,
	}
}

func (h *HealthCheck) writeQueueCheck() fthealth.Check {
	return fthealth.Check{
		ID:               "write-message-queue-proxy-reachable",
		Name:             "Write Message Queue Proxy Reachable",
		Severity:         2,
		BusinessImpact:   "Related content from published Next videos will not be processed, clients will not see them within content.",
		TechnicalSummary: "Write message queue proxy is not reachable/healthy",
		PanicGuide:       h.panicGuide,
		Checker:          h.producer.ConnectivityCheck,
	}
}

func (h *HealthCheck) GTG() gtg.Status {
	consumerCheck := func() gtg.Status {
		return gtgCheck(h.consumer.ConnectivityCheck)
	}
	producerCheck := func() gtg.Status {
		return gtgCheck(h.producer.ConnectivityCheck)
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
