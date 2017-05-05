package main

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"
)

type serviceHandler struct {
	sc serviceConfig
}

func (h serviceHandler) mapRequest(w http.ResponseWriter, r *http.Request) {
	body, err := ioutil.ReadAll(r.Body)
	if err != nil {
		writerBadRequest(w, err, "")
	}
	tid := r.Header.Get("X-Request-Id")

	m := relatedContentMapper{sc: h.sc, strContent: string(body), tid: tid}

	mappedRelatedContentBytes, _, err := h.mapRelatedContentRequest(&m)
	if err != nil {
		writerBadRequest(w, err, tid)
	}

	if mappedRelatedContentBytes == nil {
		return
	}

	w.Header().Add("Content-Type", "application/json")
	_, err = w.Write(mappedRelatedContentBytes)
	if err != nil {
		logger.serviceEvent(tid, err, "Writing response error.")
	}
}

func (h serviceHandler) mapRelatedContentRequest(m *relatedContentMapper) ([]byte, string, error) {
	if err := json.Unmarshal([]byte(m.strContent), &m.unmarshalled); err != nil {
		return nil, "", fmt.Errorf("Video JSON from Next couldn't be unmarshalled: %v. Skipping invalid JSON: %v", err.Error(), m.strContent)
	}
	return m.mapRelatedContent()
}

func writerBadRequest(w http.ResponseWriter, err error, tid string) {
	w.WriteHeader(http.StatusBadRequest)
	_, err2 := w.Write([]byte(err.Error()))
	if err2 != nil {
		logger.serviceEvent(tid, err, "Couldn't write Bad Request response.")
	}
	return
}
