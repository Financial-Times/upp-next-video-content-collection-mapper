package main

import (
	"errors"
	"net/http"
	"net/http/httptest"
	"testing"

	logger2 "github.com/Financial-Times/go-logger/v2"
	"github.com/Financial-Times/message-queue-go-producer/producer"
	consumer "github.com/Financial-Times/message-queue-gonsumer"
	"github.com/stretchr/testify/assert"
)

func initializeHealthCheck(isProducerConnectionHealthy bool, isConsumerConnectionHealthy bool) *HealthCheck {
	return &HealthCheck{
		consumer: &mockConsumerInstance{isConnectionHealthy: isConsumerConnectionHealthy},
		producer: &mockProducerInstance{isConnectionHealthy: isProducerConnectionHealthy},
	}
}

func TestNewHealthCheck(t *testing.T) {
	logConfig := logger2.KeyNamesConfig{KeyTime: "@time"}
	l := logger2.NewUPPLogger("Test", "PANIC", logConfig)
	hc := NewHealthCheck(
		producer.NewMessageProducer(producer.MessageProducerConfig{}),
		consumer.NewConsumer(consumer.QueueConfig{}, func(m consumer.Message) {}, http.DefaultClient, l),
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
	hc := initializeHealthCheck(true, true)

	req := httptest.NewRequest("GET", "http://example.com/__health", nil)
	w := httptest.NewRecorder()

	hc.Health()(w, req)

	assert.Equal(t, 200, w.Code, "It should return HTTP 200 OK")
	assert.Contains(t, w.Body.String(), `"name":"Read Message Queue Proxy Reachable","ok":true`, "Read message queue proxy healthcheck should be happy")
	assert.Contains(t, w.Body.String(), `"name":"Write Message Queue Proxy Reachable","ok":true`, "Write message queue proxy healthcheck should be happy")
}

func TestHealthCheckWithUnhappyConsumer(t *testing.T) {
	hc := initializeHealthCheck(true, false)

	req := httptest.NewRequest("GET", "http://example.com/__health", nil)
	w := httptest.NewRecorder()

	hc.Health()(w, req)

	assert.Equal(t, 200, w.Code, "It should return HTTP 200 OK")
	assert.Contains(t, w.Body.String(), `"name":"Read Message Queue Proxy Reachable","ok":false`, "Read message queue proxy healthcheck should be unhappy")
	assert.Contains(t, w.Body.String(), `"name":"Write Message Queue Proxy Reachable","ok":true`, "Write message queue proxy healthcheck should be happy")
}

func TestHealthCheckWithUnhappyProducer(t *testing.T) {
	hc := initializeHealthCheck(false, true)

	req := httptest.NewRequest("GET", "http://example.com/__health", nil)
	w := httptest.NewRecorder()

	hc.Health()(w, req)

	assert.Equal(t, 200, w.Code, "It should return HTTP 200 OK")
	assert.Contains(t, w.Body.String(), `"name":"Read Message Queue Proxy Reachable","ok":true`, "Read message queue proxy healthcheck should be happy")
	assert.Contains(t, w.Body.String(), `"name":"Write Message Queue Proxy Reachable","ok":false`, "Write message queue proxy healthcheck should be unhappy")
}

func TestUnhappyHealthCheck(t *testing.T) {
	hc := initializeHealthCheck(false, false)

	req := httptest.NewRequest("GET", "http://example.com/__health", nil)
	w := httptest.NewRecorder()

	hc.Health()(w, req)

	assert.Equal(t, 200, w.Code, "It should return HTTP 200 OK")
	assert.Contains(t, w.Body.String(), `"name":"Read Message Queue Proxy Reachable","ok":false`, "Read message queue proxy healthcheck should be unhappy")
	assert.Contains(t, w.Body.String(), `"name":"Write Message Queue Proxy Reachable","ok":false`, "Write message queue proxy healthcheck should be unhappy")
}

func TestGTGHappyFlow(t *testing.T) {
	hc := initializeHealthCheck(true, true)

	status := hc.GTG()
	assert.True(t, status.GoodToGo)
	assert.Empty(t, status.Message)
}

func TestGTGBrokenConsumer(t *testing.T) {
	hc := initializeHealthCheck(true, false)

	status := hc.GTG()
	assert.False(t, status.GoodToGo)
	assert.Equal(t, "Error connecting to the queue", status.Message)
}

func TestGTGBrokenProducer(t *testing.T) {
	hc := initializeHealthCheck(false, true)

	status := hc.GTG()
	assert.False(t, status.GoodToGo)
	assert.Equal(t, "Error connecting to the queue", status.Message)
}

type mockProducerInstance struct {
	isConnectionHealthy bool
}

type mockConsumerInstance struct {
	isConnectionHealthy bool
}

func (p *mockProducerInstance) SendMessage(string, producer.Message) error {
	return nil
}

func (p *mockProducerInstance) ConnectivityCheck() (string, error) {
	if p.isConnectionHealthy {
		return "", nil
	}

	return "", errors.New("Error connecting to the queue")
}

func (c *mockConsumerInstance) Start() {
}

func (c *mockConsumerInstance) Stop() {
}

func (c *mockConsumerInstance) ConnectivityCheck() (string, error) {
	if c.isConnectionHealthy {
		return "", nil
	}

	return "", errors.New("Error connecting to the queue")
}
