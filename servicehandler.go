package main

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"

	"github.com/Financial-Times/go-logger/v2"
)

type serviceHandler struct {
	sc  serviceConfig
	log *logger.UPPLogger
}

func (h serviceHandler) mapRequest(w http.ResponseWriter, r *http.Request) {
	body, err := ioutil.ReadAll(r.Body)
	if err != nil {
		writerBadRequest(w, err, "", h.log)
	}
	tid := r.Header.Get("X-Request-Id")

	m := relatedContentMapper{sc: h.sc, strContent: string(body), tid: tid, log: h.log}

	mappedRelatedContentBytes, err := h.mapRelatedContentRequest(&m)
	if err != nil {
		writerBadRequest(w, err, tid, h.log)
	}

	if mappedRelatedContentBytes == nil {
		return
	}

	w.Header().Add("Content-Type", "application/json")
	_, err = w.Write(mappedRelatedContentBytes)
	if err != nil {
		h.log.WithError(err).WithTransactionID(tid).Error("Writing response error.")
	}
}

func (h serviceHandler) mapRelatedContentRequest(m *relatedContentMapper) ([]byte, error) {
	if err := json.Unmarshal([]byte(m.strContent), &m.unmarshalled); err != nil {
		return nil, fmt.Errorf("Video JSON from Next couldn't be unmarshalled: %v. Skipping invalid JSON: %v", err.Error(), m.strContent)
	}
	mappedRelatedContentBytes, _, err := m.mapRelatedContent()
	return mappedRelatedContentBytes, err
}

func writerBadRequest(w http.ResponseWriter, err error, tid string, log *logger.UPPLogger) {
	w.WriteHeader(http.StatusBadRequest)
	_, err2 := w.Write([]byte(err.Error()))
	if err2 != nil {
		log.WithError(err).WithTransactionID(tid).Error("Writing response error.")
	}
}
