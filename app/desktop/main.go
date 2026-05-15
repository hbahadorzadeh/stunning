package main

import (
	_ "embed"
	"fmt"
	"log"
	"sync"
	"time"

	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/app"
	"github.com/hbahadorzadeh/stunning/app/desktop/ui"
	"github.com/hbahadorzadeh/stunning/core"
)

//go:embed assets/icon.png
var iconBytes []byte

var (
	tunnelsMutex sync.RWMutex
	tunnels      = make(map[string]*TunnelInstance)
)

type TunnelInstance struct {
	config core.TunnelConfig
	tunnel core.Tunnel
	done   chan struct{}
	stats  TunnelStats
}

type TunnelStats struct {
	mu        sync.RWMutex
	rxBytes   uint64
	txBytes   uint64
	startTime time.Time
}

func main() {
	myApp := app.New()

	// Note: Icon can't be set directly from canvas.Image
	// Icon is set in window resources or kept as embedded bytes

	// Share icon with UI
	ui.SetAboutIcon(iconBytes)

	// Create main window
	mainWindow := ui.CreateMainWindow(myApp, getTunnels, startTunnel, stopTunnel, deleteTunnel, saveTunnel)
	mainWindow.Resize(fyne.NewSize(1000, 700))
	mainWindow.ShowAndRun()
}

func getTunnels() []ui.TunnelStatus {
	tunnelsMutex.RLock()
	defer tunnelsMutex.RUnlock()

	var result []ui.TunnelStatus
	for name, inst := range tunnels {
		isAlive := inst.tunnel != nil && inst.tunnel.IsAlive()
		var rxBytes, txBytes uint64
		var uptime string

		if inst.stats.startTime != (time.Time{}) {
			duration := time.Since(inst.stats.startTime)
			uptime = formatUptime(duration)

			inst.stats.mu.RLock()
			rxBytes = inst.stats.rxBytes
			txBytes = inst.stats.txBytes
			inst.stats.mu.RUnlock()
		}

		result = append(result, ui.TunnelStatus{
			Name:      name,
			Running:   isAlive,
			RxBytes:   rxBytes,
			TxBytes:   txBytes,
			Uptime:    uptime,
			Config:    inst.config,
		})
	}
	return result
}

func startTunnel(name string) error {
	tunnelsMutex.Lock()
	defer tunnelsMutex.Unlock()

	if inst, exists := tunnels[name]; exists && inst.tunnel != nil && inst.tunnel.IsAlive() {
		return nil // Already running
	}

	if _, exists := tunnels[name]; !exists {
		return fmt.Errorf("tunnel %s not found", name)
	}

	conf := tunnels[name].config
	tunnel := core.TunnelFactory(name, conf)

	inst := tunnels[name]
	inst.tunnel = tunnel
	inst.done = make(chan struct{})
	inst.stats.startTime = time.Now()
	inst.stats.rxBytes = 0
	inst.stats.txBytes = 0

	go func() {
		defer func() {
			if r := recover(); r != nil {
				log.Printf("Tunnel %s panicked: %v\n", name, r)
			}
			tunnelsMutex.Lock()
			inst.tunnel = nil
			tunnelsMutex.Unlock()
			close(inst.done)
		}()
		tunnel.ListenAndServer()
	}()

	return nil
}

func stopTunnel(name string) error {
	tunnelsMutex.Lock()
	defer tunnelsMutex.Unlock()

	inst, exists := tunnels[name]
	if !exists || inst.tunnel == nil {
		return nil
	}

	// Mark as stopped
	inst.tunnel = nil
	inst.stats.startTime = time.Time{}

	return nil
}

func deleteTunnel(name string) error {
	tunnelsMutex.Lock()
	defer tunnelsMutex.Unlock()

	if inst, exists := tunnels[name]; exists {
		if inst.tunnel != nil {
			inst.tunnel = nil
		}
		delete(tunnels, name)
	}
	return nil
}

func saveTunnel(name string, config core.TunnelConfig) error {
	tunnelsMutex.Lock()
	defer tunnelsMutex.Unlock()

	if _, exists := tunnels[name]; !exists {
		tunnels[name] = &TunnelInstance{
			config: config,
			stats: TunnelStats{
				startTime: time.Time{},
			},
		}
	} else {
		tunnels[name].config = config
	}
	return nil
}

func formatUptime(d time.Duration) string {
	hours := int(d.Hours())
	minutes := int(d.Minutes()) % 60
	seconds := int(d.Seconds()) % 60

	if hours > 0 {
		return fmt.Sprintf("%dh %dm %ds", hours, minutes, seconds)
	}
	if minutes > 0 {
		return fmt.Sprintf("%dm %ds", minutes, seconds)
	}
	return fmt.Sprintf("%ds", seconds)
}
