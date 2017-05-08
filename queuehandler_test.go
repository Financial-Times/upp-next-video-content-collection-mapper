package main

import (
	"encoding/json"
	"github.com/Financial-Times/message-queue-go-producer/producer"
	"github.com/Financial-Times/message-queue-gonsumer/consumer"
	"github.com/stretchr/testify/assert"
	"io/ioutil"
	"testing"
	"time"
)

const videoUUID = "e2290d14-7e80-4db8-a715-949da4de9a07"

var lastModified = time.Now().Format(dateFormat)

func init() {
	logger = newAppLogger("test")
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

func newStringMappedContent(t *testing.T, itemUUID string, tid string, msgDate string) string {
	ccUUID := NewNameUUIDFromBytes([]byte(videoUUID)).String()
	var cc ContentCollection
	if itemUUID != "" {
		items := []Item{{itemUUID}}
		cc = ContentCollection{
			UUID:             ccUUID,
			Items:            items,
			PublishReference: tid,
			LastModified:     msgDate,
			CollectionType:   collectionType,
		}
	}

	mc := MappedContent{
		Payload:      cc,
		ContentURI:   contentURIPrefix + ccUUID,
		LastModified: msgDate,
		UUID:         ccUUID,
	}

	marshalledContent, err := json.Marshal(mc)
	if err != nil {
		assert.Fail(t, err.Error())
	}
	return string(marshalledContent)

}
