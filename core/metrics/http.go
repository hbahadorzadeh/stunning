package metrics

import (
	"encoding/json"
	"fmt"
	"net"
	"net/http"
	"strconv"
)

// HTTPServer provides HTTP endpoints for metrics
type HTTPServer struct {
	collector *MetricsCollector
	server    *http.Server
	listener  net.Listener
}

// NewHTTPServer creates a new HTTP metrics server
func NewHTTPServer(addr string, collector *MetricsCollector) (*HTTPServer, error) {
	if collector == nil {
		collector = globalCollector
	}

	hs := &HTTPServer{
		collector: collector,
	}

	mux := http.NewServeMux()

	// Prometheus metrics endpoint
	mux.HandleFunc("/metrics", func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "text/plain; version=0.0.4")
		fmt.Fprint(w, hs.collector.ExportPrometheus())
	})

	// Prometheus endpoint (alternate path)
	mux.HandleFunc("/prometheus", func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "text/plain; version=0.0.4")
		fmt.Fprint(w, hs.collector.ExportPrometheus())
	})

	// JSON metrics endpoint
	mux.HandleFunc("/api/metrics", func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(hs.collector.ExportJSON())
	})

	// JSON metrics for specific tunnel
	mux.HandleFunc("/api/metrics/", func(w http.ResponseWriter, r *http.Request) {
		tunnelName := r.URL.Path[len("/api/metrics/"):]
		m := hs.collector.GetMetrics(tunnelName)
		if m == nil {
			w.WriteHeader(http.StatusNotFound)
			fmt.Fprintf(w, `{"error": "tunnel not found: %s"}`, tunnelName)
			return
		}
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(m.ExportJSON(tunnelName))
	})

	// Health check
	mux.HandleFunc("/health", func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		fmt.Fprintf(w, `{"status": "ok"}`)
	})

	hs.server = &http.Server{
		Addr:    addr,
		Handler: mux,
	}

	var err error
	hs.listener, err = net.Listen("tcp", addr)
	if err != nil {
		return nil, err
	}

	// Start server in background
	go hs.server.Serve(hs.listener)

	return hs, nil
}

// GetAddr returns the address the server is listening on
func (hs *HTTPServer) GetAddr() string {
	if hs.listener == nil {
		return ""
	}
	return hs.listener.Addr().String()
}

// Close closes the HTTP server
func (hs *HTTPServer) Close() error {
	if hs.server != nil {
		return hs.server.Close()
	}
	return nil
}

// GetMetricsPort extracts port from metrics address
func GetMetricsPort(addr string) int {
	_, portStr, err := net.SplitHostPort(addr)
	if err != nil {
		return 0
	}
	port, _ := strconv.Atoi(portStr)
	return port
}
