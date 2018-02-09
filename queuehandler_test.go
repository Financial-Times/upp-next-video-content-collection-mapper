package main

import (
	"io/ioutil"
	"testing"
	"time"

	"github.com/Financial-Times/message-queue-go-producer/producer"
	"github.com/Financial-Times/message-queue-gonsumer/consumer"
	"github.com/stretchr/testify/assert"

	log "github.com/Financial-Times/go-logger/test"
)

var lastModified = time.Now().Format(dateFormat)

func init() {
	log.NewTestHook("test")
}

type mockMessageProducer struct {
	message    string
	sendCalled bool
}

func TestQueueConsume(t *testing.T) {
	assert := assert.New(t)
	tests := []struct {
		fileName        string
		originSystem    string
		tid             string
		expectedMsgSent bool
		expectedContent string
	}{
		{
			"next-video-input.json",
			nextVideoOrigin,
			"1234",
			true,
			newStringMappedContent(t, "c4cde316-128c-11e7-80f4-13e067d5072c", "1234", lastModified),
		},
		{
			"next-video-input.json",
			"other",
			"1234",
			false,
			"",
		},
		{
			"next-video-input.json",
			nextVideoOrigin,
			"",
			false,
			"",
		},
		{
			"invalid-format.json",
			nextVideoOrigin,
			"1234",
			false,
			"",
		},
		{
			"next-video-invalid-related-input.json",
			nextVideoOrigin,
			"1234",
			false,
			"",
		},
		{
			"next-video-empty-related-input.json",
			nextVideoOrigin,
			"1234",
			true,
			newStringMappedContent(t, "", "1234", lastModified),
		},
	}

	for _, test := range tests {
		mockMsgProducer := mockMessageProducer{}
		var msgProducer = &mockMsgProducer
		h := queueHandler{
			sc:              serviceConfig{},
			messageProducer: msgProducer,
		}

		msg := consumer.Message{
			Headers: createHeaders(test.originSystem, test.tid, lastModified),
			Body:    string(getBytes(test.fileName, t)),
		}
		h.queueConsume(msg)

		assert.Equal(test.expectedMsgSent, mockMsgProducer.sendCalled,
			"Message sending check is wrong. Input JSON file: %s, Origin-System-Id: %s, X-Request-Id: %s", test.fileName, test.originSystem, test.tid)
		assert.Equal(test.expectedContent, mockMsgProducer.message,
			"Marshalled content wrong. Input JSON file: %s, Origin-System-Id: %s, X-Request-Id: %s", test.fileName, test.originSystem, test.tid)
	}
}

func createHeaders(originSystem string, requestID string, msgDate string) map[string]string {
	var result = make(map[string]string)
	result["Origin-System-Id"] = originSystem
	result["X-Request-Id"] = requestID
	result["Message-Timestamp"] = msgDate
	return result
}

func (mock *mockMessageProducer) SendMessage(uuid string, message producer.Message) error {
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
