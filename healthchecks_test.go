package main

import (
	"errors"
	"net/http/httptest"
	"testing"

	"github.com/Financial-Times/go-logger/v2"
	"github.com/Financial-Times/kafka-client-go/v3"
	"github.com/stretchr/testify/assert"
)

func initializeHealthCheck(isProducerConnectionHealthy bool, isConsumerConnectionHealthy bool, isConsumerNotLagging bool) *HealthCheck {
	return &HealthCheck{
		consumer: &mockConsumerInstance{isConnectionHealthy: isConsumerConnectionHealthy, isNotLagging: isConsumerNotLagging},
		producer: &mockProducerInstance{isConnectionHealthy: isProducerConnectionHealthy},
	}
}

func TestNewHealthCheck(t *testing.T) {
	log := logger.NewUPPLogger("video-mapper", "Debug")
	hc := NewHealthCheck(
		kafka.NewProducer(kafka.ProducerConfig{}, log),
		kafka.NewConsumer(kafka.ConsumerConfig{}, []*kafka.Topic{}, log),
		"appName",
		"appSystemCode",
		"panicGuide",
	)

	assert.NotNil(t, hc.consumer)
	assert.NotNil(t, hc.producer)
	assert.Equal(t, "appName", hc.appName)
	assert.Equal(t, "appSystemCode", hc.appSystemCode)
	assert.Equal(t, "panicGuide", hc.panicGuide)
}

func TestHappyHealthCheck(t *testing.T) {
	hc := initializeHealthCheck(true, true, true)

	req := httptest.NewRequest("GET", "http://example.com/__health", nil)
	w := httptest.NewRecorder()

	hc.Health()(w, req)

	assert.Equal(t, 200, w.Code, "It should return HTTP 200 OK")
	assert.Contains(t, w.Body.String(), `"name":"Read Message Queue Reachable","ok":true`, "Read message queue healthcheck should be happy")
	assert.Contains(t, w.Body.String(), `"name":"Write Message Queue Reachable","ok":true`, "Write message queue healthcheck should be happy")
	assert.Contains(t, w.Body.String(), `"name":"Read Message Queue Is Not Lagging","ok":true`, "Read message queue is nog lagging healthcheck should be happy")
}

func TestHealthCheckWithUnhappyConsumer(t *testing.T) {
	hc := initializeHealthCheck(true, false, false)

	req := httptest.NewRequest("GET", "http://example.com/__health", nil)
	w := httptest.NewRecorder()

	hc.Health()(w, req)

	assert.Equal(t, 200, w.Code, "It should return HTTP 200 OK")
	assert.Contains(t, w.Body.String(), `"name":"Read Message Queue Reachable","ok":false`, "Read message queue healthcheck should be unhappy")
	assert.Contains(t, w.Body.String(), `"name":"Write Message Queue Reachable","ok":true`, "Write message queue ealthcheck should be happy")
	assert.Contains(t, w.Body.String(), `"name":"Read Message Queue Is Not Lagging","ok":false`, "Read message queue is nog lagging healthcheck should be unhappy")
}

func TestHealthCheckWithUnhappyProducer(t *testing.T) {
	hc := initializeHealthCheck(false, true, true)

	req := httptest.NewRequest("GET", "http://example.com/__health", nil)
	w := httptest.NewRecorder()

	hc.Health()(w, req)

	assert.Equal(t, 200, w.Code, "It should return HTTP 200 OK")
	assert.Contains(t, w.Body.String(), `"name":"Read Message Queue Reachable","ok":true`, "Read message queue healthcheck should be happy")
	assert.Contains(t, w.Body.String(), `"name":"Write Message Queue Reachable","ok":false`, "Write message queue healthcheck should be unhappy")
	assert.Contains(t, w.Body.String(), `"name":"Read Message Queue Is Not Lagging","ok":true`, "Read message queue is nog lagging healthcheck should be happy")
}

func TestUnhappyHealthCheck(t *testing.T) {
	hc := initializeHealthCheck(false, false, false)

	req := httptest.NewRequest("GET", "http://example.com/__health", nil)
	w := httptest.NewRecorder()

	hc.Health()(w, req)

	assert.Equal(t, 200, w.Code, "It should return HTTP 200 OK")
	assert.Contains(t, w.Body.String(), `"name":"Read Message Queue Reachable","ok":false`, "Read message queue healthcheck should be unhappy")
	assert.Contains(t, w.Body.String(), `"name":"Write Message Queue Reachable","ok":false`, "Write message queue healthcheck should be unhappy")
	assert.Contains(t, w.Body.String(), `"name":"Read Message Queue Is Not Lagging","ok":false`, "Read message queue is nog lagging healthcheck should be unhappy")
}

func TestGTGHappyFlow(t *testing.T) {
	hc := initializeHealthCheck(true, true, true)

	status := hc.GTG()
	assert.True(t, status.GoodToGo)
	assert.Empty(t, status.Message)
}

func TestGTGBrokenConsumer(t *testing.T) {
	hc := initializeHealthCheck(true, false, false)

	status := hc.GTG()
	assert.False(t, status.GoodToGo)
	assert.Equal(t, "error connecting to the queue", status.Message)
}

func TestGTGBrokenProducer(t *testing.T) {
	hc := initializeHealthCheck(false, true, true)

	status := hc.GTG()
	assert.False(t, status.GoodToGo)
	assert.Equal(t, "error connecting to the queue", status.Message)
}

type mockProducerInstance struct {
	isConnectionHealthy bool
}

type mockConsumerInstance struct {
	isConnectionHealthy bool
	isNotLagging        bool
}

func (p *mockProducerInstance) ConnectivityCheck() error {
	if p.isConnectionHealthy {
		return nil
	}

	return errors.New("error connecting to the queue")
}

func (c *mockConsumerInstance) ConnectivityCheck() error {
	if c.isConnectionHealthy {
		return nil
	}
	return errors.New("error connecting to the queue")
}

func (c *mockConsumerInstance) MonitorCheck() error {
	if c.isNotLagging {
		return nil
	}
	return errors.New("consumer is lagging")
}
