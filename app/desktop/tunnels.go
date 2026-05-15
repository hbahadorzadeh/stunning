package main

import (
	"sync"
	"time"

	"github.com/hbahadorzadeh/stunning/core"
)

// TunnelManager handles tunnel lifecycle and statistics
type TunnelManager struct {
	mu      sync.RWMutex
	tunnels map[string]*TunnelInstance
}

// NewTunnelManager creates a new tunnel manager
func NewTunnelManager() *TunnelManager {
	return &TunnelManager{
		tunnels: make(map[string]*TunnelInstance),
	}
}

// StartTunnel starts a tunnel by name
func (tm *TunnelManager) StartTunnel(name string) error {
	tm.mu.Lock()
	defer tm.mu.Unlock()

	inst, exists := tm.tunnels[name]
	if !exists {
		return nil
	}

	if inst.tunnel != nil && inst.tunnel.IsAlive() {
		return nil // Already running
	}

	tunnel := core.TunnelFactory(name, inst.config)
	inst.tunnel = tunnel
	inst.done = make(chan struct{})
	inst.stats.startTime = time.Now()
	inst.stats.rxBytes = 0
	inst.stats.txBytes = 0

	go func() {
		defer func() {
			if r := recover(); r != nil {
				// Log panic but don't crash
			}
			tm.mu.Lock()
			inst.tunnel = nil
			tm.mu.Unlock()
			close(inst.done)
		}()
		tunnel.ListenAndServer()
	}()

	return nil
}

// StopTunnel stops a tunnel by name
func (tm *TunnelManager) StopTunnel(name string) error {
	tm.mu.Lock()
	defer tm.mu.Unlock()

	inst, exists := tm.tunnels[name]
	if !exists || inst.tunnel == nil {
		return nil
	}

	inst.tunnel = nil
	inst.stats.startTime = time.Time{}

	return nil
}

// GetTunnel returns a specific tunnel if it exists
func (tm *TunnelManager) GetTunnel(name string) (core.Tunnel, bool) {
	tm.mu.RLock()
	defer tm.mu.RUnlock()

	inst, exists := tm.tunnels[name]
	if !exists {
		return nil, false
	}
	return inst.tunnel, inst.tunnel != nil
}

// IsAlive checks if a tunnel is alive
func (tm *TunnelManager) IsAlive(name string) bool {
	tm.mu.RLock()
	defer tm.mu.RUnlock()

	inst, exists := tm.tunnels[name]
	if !exists || inst.tunnel == nil {
		return false
	}
	return inst.tunnel.IsAlive()
}
