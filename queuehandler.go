package main

import (
	"encoding/json"
	"errors"
	"fmt"
	"strings"
	"time"

	"github.com/Financial-Times/go-logger/v2"
	"github.com/Financial-Times/kafka-client-go/v3"
	"github.com/google/uuid"
)

const (
	nextVideoOrigin  = "http://cmdb.ft.com/systems/next-video-editor"
	dateFormat       = "2006-01-02T15:04:05.000Z0700"
	generatedMsgType = "cms-content-published"
)

type messageProducer interface {
	SendMessage(message kafka.FTMessage) error
}

type queueHandler struct {
	sc              serviceConfig
	messageProducer messageProducer
	log             *logger.UPPLogger
}

func (h *queueHandler) queueConsume(m kafka.FTMessage) {
	if m.Headers["Origin-System-Id"] != nextVideoOrigin {
		h.log.WithTransactionID(m.Headers["X-Request-Id"]).Infof("Ignoring message with different Origin-System-Id: %v", m.Headers["Origin-System-Id"])
		return
	}
	if strings.Contains(m.Headers["Content-Type"], "audio") {
		h.log.WithTransactionID(m.Headers["X-Request-Id"]).Infof("Ignoring message with Content-Type: %v", m.Headers["Content-Type"])
		return
	}
	lastModified := m.Headers["Message-Timestamp"]
	if lastModified == "" {
		lastModified = time.Now().Format(dateFormat)
	}

	vm := relatedContentMapper{
		sc:           h.sc,
		strContent:   m.Body,
		tid:          m.Headers["X-Request-Id"],
		lastModified: lastModified,
		log:          h.log,
	}
	marshalledEvent, videoUUID, err := h.mapNextVideoAnnotationsMessage(&vm)
	if err != nil {
		h.log.WithTransactionID(vm.tid).WithUUID(videoUUID).
			WithError(err).Warn("Error mapping the message from queue")
		return
	}

	if marshalledEvent == nil {
		return
	}

	headers := createHeader(m.Headers, lastModified)
	msgToSend := string(marshalledEvent)
	err = h.messageProducer.SendMessage(kafka.FTMessage{Headers: headers, Body: msgToSend})
	if err != nil {
		h.log.WithTransactionID(vm.tid).WithUUID(videoUUID).
			WithError(err).Warn("Error sending transformed message to queue")
		return
	}

	h.log.WithTransactionID(vm.tid).WithUUID(videoUUID).
		WithError(err).
		Infof("Mapped and sent: [%v]", msgToSend)
}

func (h *queueHandler) mapNextVideoAnnotationsMessage(vm *relatedContentMapper) ([]byte, string, error) {
	h.log.Info("Start mapping next video message.")
	if err := json.Unmarshal([]byte(vm.strContent), &vm.unmarshalled); err != nil {
		return nil, "", fmt.Errorf("video JSON from Next couldn't be unmarshalled: %v. Skipping invalid JSON: %v", err.Error(), vm.strContent)
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
		"Message-Id":        uuid.New().String(),
		"Message-Type":      generatedMsgType,
		"Content-Type":      "application/json",
		"Origin-System-Id":  origMsgHeaders["Origin-System-Id"],
	}
}
