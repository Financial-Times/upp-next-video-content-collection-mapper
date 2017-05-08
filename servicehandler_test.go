package main

import (
	"github.com/stretchr/testify/assert"
	"io/ioutil"
	"net/http"
	"net/http/httptest"
	"os"
	"testing"
)

func init() {
	logger = newAppLogger("test")
}

func TestMapRequest(t *testing.T) {
	h := serviceHandler{
		sc: serviceConfig{},
	}

	assert := assert.New(t)
	tests := []struct {
		fileName           string
		expectedContent    string
		expectedHTTPStatus int
	}{
		{
			"next-video-input.json",
			newStringMappedContent(t, "c4cde316-128c-11e7-80f4-13e067d5072c", "", ""),
			http.StatusOK,
		},
		{
			"next-video-invalid-related-input.json",
			"",
			http.StatusBadRequest,
		},
		{
			"invalid-format.json",
			"",
			http.StatusBadRequest,
		},
	}

	for _, test := range tests {
		fileReader := getReader(test.fileName, t)
		req, _ := http.NewRequest("POST", "http://next-video-content-collection-mapper.ft.com/map", fileReader)
		w := httptest.NewRecorder()

		h.mapRequest(w, req)

		body, err := ioutil.ReadAll(w.Body)

		switch {
		case err != nil:
			assert.Fail(err.Error())
		case test.expectedHTTPStatus != http.StatusOK:
			assert.Equal(test.expectedHTTPStatus, http.StatusBadRequest, "HTTP status wrong. Input JSON: %s", test.fileName)
		default:
			assert.Equal(test.expectedHTTPStatus, http.StatusOK, "HTTP status wrong. Input JSON: %s", test.fileName)
			assert.Equal(test.expectedContent, string(body), "Marshalled content wrong. Input JSON: %s", test.fileName)
		}
	}
}

func getReader(fileName string, t *testing.T) *os.File {
	file, err := os.Open("test-resources/" + fileName)
	if err != nil {
		assert.Fail(t, err.Error())
		return nil
	}

	return file
}
