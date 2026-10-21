// SPDX-License-Identifier: MIT
// SPDX-FileCopyrightText: 2026 Steadybit GmbH

package e2e

import (
	"crypto/tls"
	"fmt"
	stdlog "log"
	"net"
	"net/http"
	"net/http/httptest"
	"os"
	"strings"
	"sync/atomic"

	"github.com/rs/zerolog/log"
)

const mockServiceId = "5a8f8d3e-2b8a-4c6f-9c3e-1f2a3b4c5d6e"

// mockKongServer serves the parts of the Kong admin API used by the discoveries.
type mockKongServer struct {
	http *httptest.Server
	// tlsHandshakeErrors counts handshakes the client aborted, e.g. because it does not trust the certificate.
	tlsHandshakeErrors atomic.Int64
}

// errorLogWriter forwards the mock's server error log to stderr and counts TLS handshake errors.
type errorLogWriter struct{ mock *mockKongServer }

func (w errorLogWriter) Write(p []byte) (int, error) {
	if strings.Contains(string(p), "TLS handshake error") {
		w.mock.tlsHandshakeErrors.Add(1)
	}
	return os.Stderr.Write(p)
}

// createMockKongServerWithTLS serves HTTPS with the given certificate.
func createMockKongServerWithTLS(cert tls.Certificate) *mockKongServer {
	listener, err := net.Listen("tcp", "0.0.0.0:0")
	if err != nil {
		panic(fmt.Sprintf("httptest: failed to listen: %v", err))
	}
	mux := http.NewServeMux()
	mock := &mockKongServer{}

	server := httptest.Server{Listener: listener, Config: &http.Server{Handler: mux, ErrorLog: stdlog.New(errorLogWriter{mock}, "", stdlog.LstdFlags)}}
	server.TLS = &tls.Config{Certificates: []tls.Certificate{cert}}
	server.StartTLS()
	log.Info().Str("url", server.URL).Msg("Started Mock-Server")
	mock.http = &server

	mux.HandleFunc("GET /services", func(w http.ResponseWriter, _ *http.Request) {
		writeJson(w, fmt.Sprintf(`{"data":[{"id":%q,"name":"mock-service","protocol":"http","host":"example.com","port":80,"enabled":true}],"offset":null}`, mockServiceId))
	})
	mux.HandleFunc("GET /services/{id}/routes", func(w http.ResponseWriter, _ *http.Request) {
		writeJson(w, `{"data":[],"offset":null}`)
	})
	return mock
}

func writeJson(w http.ResponseWriter, body string) {
	w.Header().Set("Content-Type", "application/json")
	_, _ = w.Write([]byte(body))
}
