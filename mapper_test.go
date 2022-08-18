package main

import (
	"encoding/json"
	"io/ioutil"
	"testing"

	"github.com/Financial-Times/go-logger/v2"
	"github.com/stretchr/testify/assert"
)

const (
	testVideoUUID             = "e2290d14-7e80-4db8-a715-949da4de9a07"
	testContentCollectionUUID = "e2290d14-7e80-4db8-19d1-ea8e75cf09e8"
)

var testMap = make(map[string]interface{})

func init() {
	testMap["string"] = "value1"
	testMap["nullstring"] = nil
	testMap["bool"] = true

	var objArray = make([]interface{}, 0)
	var obj = make(map[string]interface{})
	obj["field1"] = "test"
	obj["field2"] = true
	objArray = append(objArray, obj)
	testMap["objArray"] = objArray

	testMap["emptyObjArray"] = make([]interface{}, 0)
}

func TestBuildRelatedItems(t *testing.T) {
	log := logger.NewUPPLogger("video-mapper", "Debug")
	m := relatedContentMapper{log: log}
	tests := []struct {
		nextRelatedItem []map[string]interface{}
		expectedItems   []Item
	}{
		{
			[]map[string]interface{}{
				newRelatedItem("c4cde316-128c-11e7-80f4-13e067d5072c"),
				newRelatedItem("e2290d14-7e80-4db8-a715-949da4de9a07"),
			},
			[]Item{
				{UUID: "c4cde316-128c-11e7-80f4-13e067d5072c"},
				{UUID: "e2290d14-7e80-4db8-a715-949da4de9a07"},
			},
		},
		{
			[]map[string]interface{}{
				newRelatedItem(nil),
			},
			[]Item{},
		},
	}
	for _, test := range tests {
		items := m.retrieveRelatedItems(test.nextRelatedItem, "")
		assert.Equal(t, test.expectedItems, items, "Related items are wrong. Test input: [%v]", test.nextRelatedItem)
	}
}

func TestMapNextVideoRelatedContentHappyFlows(t *testing.T) {
	tests := []struct {
		fileName          string
		expectedContent   string
		expectedVideoUUID string
		expectedErrStatus bool
	}{
		{
			"next-video-input.json",
			newStringMappedContent(t, "c4cde316-128c-11e7-80f4-13e067d5072c", "", ""),
			testVideoUUID,
			false,
		},
		{
			"next-video-delete-input.json",
			newStringMappedContent(t, "", "", ""),
			testVideoUUID,
			false,
		},
	}

	for _, test := range tests {
		nextVideo, err := readContent(test.fileName)
		if err != nil {
			assert.Fail(t, err.Error())
		}
		m := relatedContentMapper{
			sc:           serviceConfig{},
			unmarshalled: nextVideo,
		}

		marshalledContent, videoUUID, err := m.mapRelatedContent()

		assert.Equal(t, test.expectedContent, string(marshalledContent), "Marshalled content wrong. Input JSON: %s", test.fileName)
		assert.Equal(t, test.expectedVideoUUID, videoUUID, "Video UUID wrong. Input JSON: %s", test.fileName)
		assert.Equal(t, test.expectedErrStatus, err != nil, "Error status wrong. Input JSON: %s", test.fileName)
	}
}

func TestMapNextVideoRelatedContentMissingFields(t *testing.T) {
	log := logger.NewUPPLogger("video-mapper", "Debug")
	tests := []struct {
		fileName          string
		expectedErrStatus bool
	}{
		{
			"next-video-no-related-input.json",
			false,
		},
		{
			"next-video-empty-related-input.json",
			false,
		},
		{
			"next-video-invalid-related-input.json",
			true,
		},
		{
			"next-video-related-no-item-id-input.json",
			false,
		},
		{
			"next-video-no-videouuid-input.json",
			true,
		},
	}

	for _, test := range tests {
		nextVideo, err := readContent(test.fileName)
		if err != nil {
			assert.Fail(t, err.Error())
		}
		m := relatedContentMapper{unmarshalled: nextVideo, log: log}
		_, _, err = m.mapRelatedContent()
		assert.Equal(t, test.expectedErrStatus, err != nil, "Error status wrong. Input JSON: %s", test.fileName)
	}
}

func TestGetRequiredStringField(t *testing.T) {
	tests := []struct {
		key           string
		expectedValue interface{}
		expectedIsErr bool
	}{
		{
			"string",
			"value1",
			false,
		},
		{
			"nullstring",
			"",
			true,
		},
		{
			"bool",
			"",
			true,
		},
		{
			"no_key",
			"",
			true,
		},
	}

	for _, test := range tests {
		result, err := getRequiredStringField(test.key, testMap)
		assert.Equal(t, test.expectedValue, result, "Value is wrong. Map key: %s", test.key)
		assert.Equal(t, test.expectedIsErr, err != nil, "Error status is wrong. Map key: %s", test.key)
	}
}

func TestGetObjectsArrayField(t *testing.T) {
	log := logger.NewUPPLogger("video-mapper", "Debug")
	m := relatedContentMapper{log: log}
	var objArray = make([]map[string]interface{}, 0)
	var objValue = make(map[string]interface{})
	objValue["field1"] = "test"
	objValue["field2"] = true
	objArray = append(objArray, objValue)
	tests := []struct {
		key           string
		expectedValue []map[string]interface{}
		expectedIsErr bool
	}{
		{
			"objArray",
			objArray,
			false,
		},
		{
			"string",
			nil,
			true,
		},
		{
			"no_key",
			make([]map[string]interface{}, 0),
			false,
		},
		{
			"emptyObjArray",
			make([]map[string]interface{}, 0),
			false,
		},
	}

	for _, test := range tests {
		result, err := getObjectsArrayField(test.key, testMap, "videoUUID", &m)
		assert.Equal(t, test.expectedValue, result, "Value is wrong. Map key: %s", test.key)
		assert.Equal(t, test.expectedIsErr, err != nil, "Error status is wrong. Map key: %s", test.key)
	}
}

func readContent(fileName string) (map[string]interface{}, error) {
	data, err := ioutil.ReadFile("test-resources/" + fileName)
	if err != nil {
		return nil, err
	}

	var result = make(map[string]interface{})
	err = json.Unmarshal(data, &result)
	if err != nil {
		return nil, err
	}
	return result, nil
}

func newRelatedItem(id interface{}) map[string]interface{} {
	var obj = make(map[string]interface{})
	if id != nil {
		obj[relatedItemIDField] = id
	}
	return obj
}

func newStringMappedContent(t *testing.T, itemUUID string, tid string, msgDate string) string {
	var cc ContentCollection
	if itemUUID != "" {
		items := []Item{{itemUUID}}
		cc = ContentCollection{
			UUID:             testContentCollectionUUID,
			Items:            items,
			PublishReference: tid,
			LastModified:     msgDate,
			CollectionType:   collectionType,
		}
	}

	mc := MappedContent{
		Payload:      cc,
		ContentURI:   contentURIPrefix + testContentCollectionUUID,
		LastModified: msgDate,
		UUID:         testContentCollectionUUID,
	}

	marshalledContent, err := json.Marshal(mc)
	if err != nil {
		assert.Fail(t, err.Error())
	}
	return string(marshalledContent)
}
