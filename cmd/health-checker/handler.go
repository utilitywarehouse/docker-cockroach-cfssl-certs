package main

import (
	"encoding/json"
	"fmt"
	"github.com/pkg/errors"
	"io"
	"net/http"
	"time"

	log "github.com/sirupsen/logrus"
)

func copyHeaders(from, to http.Header) {
	for header, values := range from {
		for _, value := range values {
			to.Add(header, value)
		}
	}
}

type handler struct {
	client  http.Client
	expTime time.Time
	host    string
	logger  *log.Entry
}

func (h *handler) writeError(writer http.ResponseWriter, message string) {
	writer.Header().Set("Content-Type", "application/json")
	writer.WriteHeader(http.StatusInternalServerError)
	resp, err := json.Marshal(map[string]string{
		"error": message,
	})
	if err != nil {
		h.logger.Error(errors.Wrap(err, "failed to create json error response"))
		return
	}

	_, err = writer.Write(resp)
	if err != nil {
		h.logger.Error(errors.Wrap(err, "failed to write response"))
		return
	}

}

func (h *handler) forwardRequest(writer http.ResponseWriter, request *http.Request) {
	url := request.URL
	url.Host = h.host
	url.Scheme = "http"

	proxyReq, err := http.NewRequest(request.Method, url.String(), request.Body)
	if err != nil {
		h.writeError(writer, fmt.Sprintf("failed to forward the request: %v", err))
		return
	}

	proxyReq.Header.Set("Host", request.Host)
	proxyReq.Header.Set("X-Forwarded-For", request.RemoteAddr)
	copyHeaders(request.Header, proxyReq.Header)

	proxyRes, err := h.client.Do(proxyReq)
	if err != nil {
		h.writeError(writer, fmt.Sprintf("failed to forward the request: %v", err))
		return
	}
	defer proxyRes.Body.Close()

	copyHeaders(proxyRes.Header, writer.Header())
	writer.WriteHeader(proxyRes.StatusCode)

	_, err = io.Copy(writer, proxyRes.Body)
	if err != nil {
		h.logger.Error(errors.Wrap(err, "failed to write response"))
		return
	}
}

func (h *handler) ServeHTTP(writer http.ResponseWriter, request *http.Request) {
	if request.Method != http.MethodGet {
		writer.WriteHeader(http.StatusMethodNotAllowed)
		return
	}

	// Check whether expiration time is before now, i.e. whether the certificate has already expired
	if h.expTime.Before(time.Now()) {
		h.writeError(writer, "node certificate expired")
		return
	}

	h.forwardRequest(writer, request)
}
