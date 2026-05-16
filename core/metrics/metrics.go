// Package metrics provides Prometheus-compatible metrics for tunnels
package metrics

import (
	"fmt"
	"sync"
	"sync/atomic"
	"time"
)

// Metrics tracks tunnel statistics
type Metrics struct {
	mu                 sync.RWMutex
	BytesReceived      atomic.Int64
	BytesSent          atomic.Int64
	ConnectionsTotal   atomic.Int64
	ConnectionsCurrent atomic.Int64
	ErrorsTotal        atomic.Int64
	StartTime          time.Time
	LastActivity       time.Time
}

// NewMetrics creates a new metrics instance
func NewMetrics() *Metrics {
	return &Metrics{
		StartTime:    time.Now(),
		LastActivity: time.Now(),
	}
}

// RecordBytesReceived records bytes received
func (m *Metrics) RecordBytesReceived(bytes int64) {
	m.BytesReceived.Add(bytes)
	m.updateLastActivity()
}

// RecordBytesSent records bytes sent
func (m *Metrics) RecordBytesSent(bytes int64) {
	m.BytesSent.Add(bytes)
	m.updateLastActivity()
}

// RecordConnectionOpened increments connection count
func (m *Metrics) RecordConnectionOpened() {
	m.ConnectionsTotal.Add(1)
	m.ConnectionsCurrent.Add(1)
	m.updateLastActivity()
}

// RecordConnectionClosed decrements current connection count
func (m *Metrics) RecordConnectionClosed() {
	m.ConnectionsCurrent.Add(-1)
	m.updateLastActivity()
}

// RecordError increments error count
func (m *Metrics) RecordError() {
	m.ErrorsTotal.Add(1)
	m.updateLastActivity()
}

// Uptime returns uptime in seconds
func (m *Metrics) Uptime() int64 {
	return int64(time.Since(m.StartTime).Seconds())
}

// updateLastActivity updates last activity timestamp
func (m *Metrics) updateLastActivity() {
	m.mu.Lock()
	m.LastActivity = time.Now()
	m.mu.Unlock()
}

// GetLastActivity returns last activity time
func (m *Metrics) GetLastActivity() time.Time {
	m.mu.RLock()
	defer m.mu.RUnlock()
	return m.LastActivity
}

// Export returns metrics in Prometheus text format
func (m *Metrics) Export(tunnelName string) string {
	uptime := m.Uptime()
	rxBytes := m.BytesReceived.Load()
	txBytes := m.BytesSent.Load()
	connTotal := m.ConnectionsTotal.Load()
	connCurrent := m.ConnectionsCurrent.Load()
	errors := m.ErrorsTotal.Load()

	return fmt.Sprintf(`# HELP tunnel_uptime_seconds Tunnel uptime in seconds
# TYPE tunnel_uptime_seconds gauge
tunnel_uptime_seconds{tunnel="%s"} %d
# HELP tunnel_bytes_received_total Total bytes received
# TYPE tunnel_bytes_received_total counter
tunnel_bytes_received_total{tunnel="%s"} %d
# HELP tunnel_bytes_sent_total Total bytes sent
# TYPE tunnel_bytes_sent_total counter
tunnel_bytes_sent_total{tunnel="%s"} %d
# HELP tunnel_connections_total Total connections opened
# TYPE tunnel_connections_total counter
tunnel_connections_total{tunnel="%s"} %d
# HELP tunnel_connections_current Current active connections
# TYPE tunnel_connections_current gauge
tunnel_connections_current{tunnel="%s"} %d
# HELP tunnel_errors_total Total errors
# TYPE tunnel_errors_total counter
tunnel_errors_total{tunnel="%s"} %d
`, tunnelName, uptime, tunnelName, rxBytes, tunnelName, txBytes, tunnelName, connTotal, tunnelName, connCurrent, tunnelName, errors)
}

// ExportJSON returns metrics in JSON format
func (m *Metrics) ExportJSON(tunnelName string) map[string]interface{} {
	return map[string]interface{}{
		"tunnel":              tunnelName,
		"uptime_seconds":      m.Uptime(),
		"bytes_received":      m.BytesReceived.Load(),
		"bytes_sent":          m.BytesSent.Load(),
		"connections_total":   m.ConnectionsTotal.Load(),
		"connections_current": m.ConnectionsCurrent.Load(),
		"errors_total":        m.ErrorsTotal.Load(),
		"last_activity":       m.GetLastActivity().Unix(),
	}
}

// MetricsCollector collects metrics from multiple tunnels
type MetricsCollector struct {
	tunnels map[string]*Metrics
	mu      sync.RWMutex
}

// NewMetricsCollector creates a new metrics collector
func NewMetricsCollector() *MetricsCollector {
	return &MetricsCollector{
		tunnels: make(map[string]*Metrics),
	}
}

// Register registers a tunnel's metrics
func (mc *MetricsCollector) Register(name string, metrics *Metrics) {
	mc.mu.Lock()
	defer mc.mu.Unlock()
	mc.tunnels[name] = metrics
}

// Unregister unregisters a tunnel's metrics
func (mc *MetricsCollector) Unregister(name string) {
	mc.mu.Lock()
	defer mc.mu.Unlock()
	delete(mc.tunnels, name)
}

// GetMetrics returns metrics for a tunnel
func (mc *MetricsCollector) GetMetrics(name string) *Metrics {
	mc.mu.RLock()
	defer mc.mu.RUnlock()
	return mc.tunnels[name]
}

// ExportPrometheus exports all metrics in Prometheus format
func (mc *MetricsCollector) ExportPrometheus() string {
	mc.mu.RLock()
	defer mc.mu.RUnlock()

	output := "# Stunning Tunnel Metrics\n\n"
	for name, m := range mc.tunnels {
		output += m.Export(name) + "\n"
	}
	return output
}

// ExportJSON exports all metrics in JSON format
func (mc *MetricsCollector) ExportJSON() map[string]interface{} {
	mc.mu.RLock()
	defer mc.mu.RUnlock()

	result := make(map[string]interface{})
	for name, m := range mc.tunnels {
		result[name] = m.ExportJSON(name)
	}
	return result
}

// Global metrics collector instance
var globalCollector = NewMetricsCollector()

// GetGlobalCollector returns the global metrics collector
func GetGlobalCollector() *MetricsCollector {
	return globalCollector
}
