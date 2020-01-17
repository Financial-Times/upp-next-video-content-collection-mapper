package main

import (
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
	"strings"
	"time"

	"github.com/Financial-Times/message-queue-go-producer/producer"
	consumer "github.com/Financial-Times/message-queue-gonsumer"
	uuid "github.com/satori/go.uuid"

	log "github.com/Financial-Times/go-logger"
)

const (
	nextVideoOrigin  = "http://cmdb.ft.com/systems/next-video-editor"
	dateFormat       = "2006-01-02T15:04:05.000Z0700"
	generatedMsgType = "cms-content-published"
)

type queueHandler struct {
	sc              serviceConfig
	httpCl          *http.Client
	consumerConfig  consumer.QueueConfig
	producerConfig  producer.MessageProducerConfig
	messageConsumer consumer.MessageConsumer
	messageProducer producer.MessageProducer
}

func (h *queueHandler) init() {
	h.messageProducer = producer.NewMessageProducer(h.producerConfig)
	h.messageConsumer = consumer.NewConsumer(h.consumerConfig, h.queueConsume, h.httpCl, h.sc.log)
}

func (h *queueHandler) queueConsume(m consumer.Message) {
	if m.Headers["Origin-System-Id"] != nextVideoOrigin {
		log.WithTransactionID(m.Headers["X-Request-Id"]).WithField("queue_topic", h.consumerConfig.Topic).Infof("Ignoring message with different Origin-System-Id: %v", m.Headers["Origin-System-Id"])
		return
	}
	if strings.Contains(m.Headers["Content-Type"], "audio") {
		log.WithTransactionID(m.Headers["X-Request-Id"]).WithField("queue_topic", h.consumerConfig.Topic).Infof("Ignoring message with Content-Type: %v", m.Headers["Content-Type"])
		return
	}
	lastModified := m.Headers["Message-Timestamp"]
	if lastModified == "" {
		lastModified = time.Now().Format(dateFormat)
	}

	vm := relatedContentMapper{sc: h.sc, strContent: m.Body, tid: m.Headers["X-Request-Id"], lastModified: lastModified}
	marshalledEvent, videoUUID, err := h.mapNextVideoAnnotationsMessage(&vm)
	if err != nil {
		log.WithTransactionID(vm.tid).WithUUID(videoUUID).
			WithField("queue_name", h.consumerConfig.Queue).
			WithField("queue_topic", h.consumerConfig.Topic).
			WithError(err).Warn("Error mapping the message from queue")
		return
	}

	if marshalledEvent == nil {
		return
	}

	headers := createHeader(m.Headers, lastModified)
	msgToSend := string(marshalledEvent)
	err = h.messageProducer.SendMessage("", producer.Message{Headers: headers, Body: msgToSend})
	if err != nil {
		log.WithTransactionID(vm.tid).WithUUID(videoUUID).
			WithField("queue_name", h.consumerConfig.Queue).
			WithField("queue_topic", h.consumerConfig.Topic).
			WithError(err).Warn("Error sending transformed message to queue")
		return
	}

	log.WithTransactionID(vm.tid).WithUUID(videoUUID).
		WithField("queue_name", h.consumerConfig.Queue).
		WithField("queue_topic", h.consumerConfig.Topic).
		WithError(err).
		Infof("Mapped and sent: [%v]", msgToSend)
}

func (h *queueHandler) mapNextVideoAnnotationsMessage(vm *relatedContentMapper) ([]byte, string, error) {
	log.WithField("queue_topic", h.consumerConfig.Topic).Info("Start mapping next video message.")
	if err := json.Unmarshal([]byte(vm.strContent), &vm.unmarshalled); err != nil {
		return nil, "", fmt.Errorf("Video JSON from Next couldn't be unmarshalled: %v. Skipping invalid JSON: %v", err.Error(), vm.strContent)
	}
	if vm.tid == "" {
		return nil, "", errors.New("X-Request-Id not found in kafka message headers. Skipping message")
	}
	return vm.mapRelatedContent()
}

func createHeader(origMsgHeaders map[string]string, lastModified string) map[string]string {
	return map[string]string{
		"X-Request-Id":      origMsgHeaders["X-Request-Id"],
		"Message-Timestamp": lastModified,
		"Message-Id":        uuid.NewV4().String(),
		"Message-Type":      generatedMsgType,
		"Content-Type":      "application/json",
		"Origin-System-Id":  origMsgHeaders["Origin-System-Id"],
	}
}
