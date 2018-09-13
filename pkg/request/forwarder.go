package request

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"

	"github.com/cloudflare/cfssl/log"
	"github.com/pkg/errors"
)

func copyHeaders(from, to http.Header) {
	for header, values := range from {
		for _, value := range values {
			to.Add(header, value)
		}
	}
}

// Forwarder implements `http.Handler` interface and forwards all incoming requests
// to the provided host using the provided scheme and http client
type Forwarder struct {
	HTTPClient    http.Client
	ForwardScheme string
	ForwardHost   string
}

func (forwarder *Forwarder) writeError(writer http.ResponseWriter, message string) {
	writer.Header().Set("Content-Type", "application/json")
	writer.WriteHeader(http.StatusInternalServerError)
	resp, err := json.Marshal(map[string]string{
		"error": message,
	})

	if err != nil {
		log.Error(errors.Wrap(err, "failed to create json error response"))
		return
	}

	_, err = writer.Write(resp)
	if err != nil {
		log.Error(errors.Wrap(err, "failed to write response"))
		return
	}
}

// ServeHTTP forwards the request to the specified `ForwardHost`.
func (forwarder *Forwarder) ServeHTTP(writer http.ResponseWriter, request *http.Request) {
	url := request.URL
	url.Host = forwarder.ForwardHost
	url.Scheme = "https"

	proxyReq, err := http.NewRequest(request.Method, url.String(), request.Body)
	if err != nil {
		forwarder.writeError(writer, fmt.Sprintf("failed to forward the request: %v", err))
		return
	}

	proxyReq.Header.Set("Host", request.Host)
	proxyReq.Header.Set("X-Forwarded-For", request.RemoteAddr)
	copyHeaders(request.Header, proxyReq.Header)

	proxyRes, err := forwarder.HTTPClient.Do(proxyReq)
	if err != nil {
		forwarder.writeError(writer, fmt.Sprintf("failed to forward the request: %v", err))
		return
	}
	defer proxyRes.Body.Close()

	copyHeaders(proxyRes.Header, writer.Header())
	writer.WriteHeader(proxyRes.StatusCode)

	_, err = io.Copy(writer, proxyRes.Body)
	if err != nil {
		log.Error(errors.Wrap(err, "failed to write response"))
		return
	}
}
