package main

import (
	"io/ioutil"
	"testing"
	"time"

	"github.com/Financial-Times/go-logger/v2"
	"github.com/Financial-Times/kafka-client-go/v3"
	"github.com/stretchr/testify/assert"
)

var lastModified = time.Now().Format(dateFormat)

type mockMessageProducer struct {
	message    string
	sendCalled bool
}

func TestQueueConsume(t *testing.T) {
	tests := []struct {
		fileName        string
		originSystem    string
		contentType     string
		tid             string
		expectedMsgSent bool
		expectedContent string
	}{
		{
			"next-video-input.json",
			nextVideoOrigin,
			"application/json",
			"1234",
			true,
			newStringMappedContent(t, "c4cde316-128c-11e7-80f4-13e067d5072c", "1234", lastModified, false),
		},
		{
			"next-video-input.json",
			"other",
			"application/json",
			"1234",
			false,
			"",
		},
		{
			"next-video-input.json",
			nextVideoOrigin,
			"application/json",
			"",
			false,
			"",
		},
		{
			"invalid-format.json",
			nextVideoOrigin,
			"application/json",
			"1234",
			false,
			"",
		},
		{
			"next-video-invalid-related-input.json",
			nextVideoOrigin,
			"application/json",
			"1234",
			false,
			"",
		},
		{
			"next-video-empty-related-input.json",
			nextVideoOrigin,
			"application/json",
			"1234",
			true,
			newStringMappedContent(t, "", "1234", lastModified, false),
		},
		{
			"next-video-empty-related-input.json",
			nextVideoOrigin,
			"audio",
			"1234",
			false,
			"",
		},
		{
			"next-video-empty-related-input.json",
			nextVideoOrigin,
			"application/vnd.ft-upp-audio",
			"1234",
			false,
			"",
		},
	}

	for _, test := range tests {
		mockMsgProducer := mockMessageProducer{}
		var msgProducer = &mockMsgProducer
		log := logger.NewUPPLogger("video-mapper", "Debug")
		h := queueHandler{
			sc:              serviceConfig{},
			messageProducer: msgProducer,
			log:             log,
		}

		msg := kafka.FTMessage{
			Headers: createHeaders(test.originSystem, test.contentType, test.tid, lastModified),
			Body:    string(getBytes(test.fileName, t)),
		}
		h.queueConsume(msg)

		assert.Equal(t, test.expectedMsgSent, mockMsgProducer.sendCalled,
			"Message sending check is wrong. Input JSON file: %s, Origin-System-Id: %s, X-Request-Id: %s", test.fileName, test.originSystem, test.tid)
		assert.Equal(t, test.expectedContent, mockMsgProducer.message,
			"Marshalled content wrong. Input JSON file: %s, Origin-System-Id: %s, X-Request-Id: %s", test.fileName, test.originSystem, test.tid)
	}
}

func createHeaders(originSystem string, contentType string, requestID string, msgDate string) map[string]string {
	var result = make(map[string]string)
	result["Origin-System-Id"] = originSystem
	result["Content-Type"] = contentType
	result["X-Request-Id"] = requestID
	result["Message-Timestamp"] = msgDate
	return result
}

func (mock *mockMessageProducer) SendMessage(message kafka.FTMessage) error {
	mock.message = message.Body
	mock.sendCalled = true
	return nil
}

func (mock *mockMessageProducer) ConnectivityCheck() (string, error) {
	// do nothing
	return "", nil
}

func getBytes(fileName string, t *testing.T) []byte {
	bytes, err := ioutil.ReadFile("test-resources/" + fileName)
	if err != nil {
		assert.Fail(t, err.Error())
		return nil
	}

	return bytes
}
