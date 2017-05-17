package main

import (
	"github.com/Financial-Times/message-queue-go-producer/producer"
	"github.com/Financial-Times/message-queue-gonsumer/consumer"
	"github.com/gorilla/handlers"
	"github.com/gorilla/mux"
	"github.com/stretchr/testify/assert"
	"net/http"
	"net/http/httptest"
	"testing"
)

const (
	statusOK int = 1 + iota
	statusNA
)

var queueServerMock *httptest.Server

func init() {
	logger = newAppLogger("test")
}

func TestCheckMessageQueueAvailability(t *testing.T) {
	assert := assert.New(t)

	startQueueServerMock(statusOK)
	defer queueServerMock.Close()

	hc := healthConfig{
		httpCl:       &http.Client{},
		consumerConf: newConsumerConfig(queueServerMock.URL),
		producerConf: newProducerConfig(queueServerMock.URL),
	}

	hs := newHealthService(&hc)

	result, err := hs.checkAggregateMessageQueueProxiesReachable()
	assert.Nil(err, "Error not expected.")
	assert.Equal("Ok", result, "Message queue availability status is wrong")
}

func TestCheckMessageQueueNonAvailability(t *testing.T) {
	assert := assert.New(t)

	startQueueServerMock(statusNA)
	defer queueServerMock.Close()

	hc := healthConfig{
		httpCl:       &http.Client{},
		consumerConf: newConsumerConfig(queueServerMock.URL),
		producerConf: newProducerConfig(queueServerMock.URL),
	}

	hs := newHealthService(&hc)

	_, err := hs.checkAggregateMessageQueueProxiesReachable()
	assert.Equal(true, err != nil, "Error was expected.")
}

func TestCheckMessageQueueWrongQueueURL(t *testing.T) {
	assert := assert.New(t)

	startQueueServerMock(statusOK)
	defer queueServerMock.Close()

	tests := []struct {
		consumerConfig consumer.QueueConfig
		producerConfig producer.MessageProducerConfig
	}{
		{
			newConsumerConfig("wrong url"),
			newProducerConfig(queueServerMock.URL),
		},
		{
			newConsumerConfig(queueServerMock.URL),
			newProducerConfig("wrong url"),
		},
	}

	for _, test := range tests {
		hc := healthConfig{
			httpCl:       &http.Client{},
			consumerConf: test.consumerConfig,
			producerConf: test.producerConfig,
		}

		hs := newHealthService(&hc)

		_, err := hs.checkAggregateMessageQueueProxiesReachable()
		assert.Equal(true, err != nil, "Error was expected for input consumer [%v], producer [%v]", test.consumerConfig, test.producerConfig)
	}
}

func startQueueServerMock(status int) {
	router := mux.NewRouter()
	var getContent http.HandlerFunc

	switch status {
	case statusOK:
		getContent = statusOKHandler
	case statusNA:
		getContent = internalErrorHandler
	}

	router.Path("/topics").Handler(handlers.MethodHandler{"GET": http.HandlerFunc(getContent)})

	queueServerMock = httptest.NewServer(router)
}

func statusOKHandler(w http.ResponseWriter, r *http.Request) {
	w.WriteHeader(http.StatusOK)
}

func internalErrorHandler(w http.ResponseWriter, r *http.Request) {
	w.WriteHeader(http.StatusInternalServerError)
}

func newConsumerConfig(addr string) consumer.QueueConfig {
	return consumer.QueueConfig{
		Addrs:            []string{addr},
		Queue:            "queue",
		AuthorizationKey: "auth",
	}
}

func newProducerConfig(addr string) producer.MessageProducerConfig {
	return producer.MessageProducerConfig{
		Addr:          addr,
		Queue:         "queue",
		Authorization: "auth",
	}
}
