package ui

import (
	"fmt"
	"image/color"
	"sync"
	"time"

	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/canvas"
	"fyne.io/fyne/v2/container"
	"fyne.io/fyne/v2/widget"
	"github.com/hbahadorzadeh/stunning/core"
)

type TunnelStatus struct {
	Name    string
	Running bool
	RxBytes uint64
	TxBytes uint64
	Uptime  string
	Config  core.TunnelConfig
}

func CreateMainWindow(myApp fyne.App, getTunnels func() []TunnelStatus, startTunnel func(string) error,
	stopTunnel func(string) error, deleteTunnel func(string) error, saveTunnel func(string, core.TunnelConfig) error) fyne.Window {

	win := myApp.NewWindow("Stunning Tunnel Manager")

	// Toolbar buttons
	addBtn := widget.NewButton("Add", nil)
	removeBtn := widget.NewButton("Remove", nil)
	startBtn := widget.NewButton("Start", nil)
	stopBtn := widget.NewButton("Stop", nil)
	aboutBtn := widget.NewButton("About", nil)

	// Status bar
	statusBar := widget.NewLabel("Ready")

	// Details panel
	selectedName := ""
	detailsContainer := container.NewVBox()
	var statusView *StatusView
	statusViewMu := &sync.Mutex{}

	// Tunnel list - custom list
	tunnelList := widget.NewList(
		func() int {
			return len(getTunnels())
		},
		func() fyne.CanvasObject {
			return container.NewHBox(
				widget.NewLabel("Name"),
				canvas.NewCircle(color.RGBA{R: 200, G: 200, B: 200, A: 255}),
			)
		},
		func(id widget.ListItemID, obj fyne.CanvasObject) {
			tunnels := getTunnels()
			if id >= len(tunnels) {
				return
			}
			status := tunnels[id]

			hbox := obj.(*fyne.Container)
			label := hbox.Objects[0].(*widget.Label)
			label.SetText(status.Name)

			circle := hbox.Objects[1].(*canvas.Circle)
			circle.StrokeWidth = 2
			if status.Running {
				circle.FillColor = color.RGBA{R: 0, G: 200, B: 0, A: 255}
			} else {
				circle.FillColor = color.RGBA{R: 200, G: 0, B: 0, A: 255}
			}
			circle.Resize(fyne.NewSize(12, 12))
		},
	)

	tunnelList.OnSelected = func(id widget.ListItemID) {
		tunnels := getTunnels()
		if id >= len(tunnels) {
			return
		}
		selectedName = tunnels[id].Name
		status := tunnels[id]

		statusViewMu.Lock()
		statusView = CreateStatusView(status)
		statusViewMu.Unlock()

		detailsContainer.Objects = nil
		detailsContainer.Objects = append(detailsContainer.Objects, statusView.Container)
		detailsContainer.Refresh()
	}

	// Update button callbacks
	addBtn.OnTapped = func() {
		CreateTunnelDialog(myApp, func(config core.TunnelConfig, name string) error {
			if err := saveTunnel(name, config); err != nil {
				return err
			}
			tunnelList.Refresh()
			return nil
		})
	}

	removeBtn.OnTapped = func() {
		if selectedName != "" {
			deleteTunnel(selectedName)
			tunnelList.Refresh()
			detailsContainer.Objects = nil
			detailsContainer.Refresh()
			selectedName = ""
		}
	}

	startBtn.OnTapped = func() {
		if selectedName != "" {
			startTunnel(selectedName)
			tunnelList.Refresh()
		}
	}

	stopBtn.OnTapped = func() {
		if selectedName != "" {
			stopTunnel(selectedName)
			tunnelList.Refresh()
		}
	}

	aboutBtn.OnTapped = func() {
		CreateAboutDialog(myApp)
	}

	// Build toolbar
	toolbar := container.NewHBox(addBtn, removeBtn, startBtn, stopBtn, aboutBtn)

	// Build main layout with split
	topSplit := container.NewHSplit(
		tunnelList,
		detailsContainer,
	)
	topSplit.SetOffset(0.3)

	// Main border layout
	mainContent := container.NewBorder(
		toolbar,   // top
		statusBar, // bottom
		nil, nil,
		topSplit,
	)

	win.SetContent(mainContent)

	// Periodic refresh (update stats every 500ms)
	ticker := time.NewTicker(500 * time.Millisecond)
	refreshCancel := make(chan struct{})

	go func() {
		defer ticker.Stop()
		for {
			select {
			case <-ticker.C:
				tunnelList.Refresh()
				if selectedName != "" {
					statusViewMu.Lock()
					sv := statusView
					statusViewMu.Unlock()

					if sv != nil {
						tunnels := getTunnels()
						for i := range tunnels {
							if tunnels[i].Name == selectedName {
								sv.Update(tunnels[i])
								break
							}
						}
					}
				}
			case <-refreshCancel:
				return
			}
		}
	}()

	win.SetOnClosed(func() {
		close(refreshCancel)
	})

	return win
}

type StatusView struct {
	Container *fyne.Container
	status    TunnelStatus
	mu        sync.Mutex
}

func CreateStatusView(status TunnelStatus) *StatusView {
	sv := &StatusView{
		status: status,
	}
	sv.rebuild()
	return sv
}

func (sv *StatusView) rebuild() {
	sv.mu.Lock()
	defer sv.mu.Unlock()

	nameText := canvas.NewText(sv.status.Name, color.White)
	nameText.TextSize = 24
	nameText.TextStyle.Bold = true

	var statusBadge *canvas.Text
	if sv.status.Running {
		statusBadge = canvas.NewText("RUNNING", color.RGBA{R: 0, G: 200, B: 0, A: 255})
	} else {
		statusBadge = canvas.NewText("STOPPED", color.RGBA{R: 200, G: 0, B: 0, A: 255})
	}
	statusBadge.TextStyle.Bold = true

	modeText := canvas.NewText(fmt.Sprintf("Mode: %s", sv.status.Config.ServiceMode), color.White)
	serverTypeText := canvas.NewText(fmt.Sprintf("Server Type: %s", sv.status.Config.ServerType), color.White)
	interfaceTypeText := canvas.NewText(fmt.Sprintf("Interface Type: %s", sv.status.Config.InterfaceType), color.White)
	listenText := canvas.NewText(fmt.Sprintf("Listen: %s", sv.status.Config.Listen), color.White)
	connectText := canvas.NewText(fmt.Sprintf("Connect: %s", sv.status.Config.Connect), color.White)

	rxText := canvas.NewText(fmt.Sprintf("RX: %d bytes", sv.status.RxBytes), color.White)
	txText := canvas.NewText(fmt.Sprintf("TX: %d bytes", sv.status.TxBytes), color.White)
	uptimeText := canvas.NewText(fmt.Sprintf("Uptime: %s", sv.status.Uptime), color.White)

	copyBtn := widget.NewButton("Copy Config", func() {
		// Copy JSON to clipboard
	})

	sv.Container = container.NewVBox(
		nameText,
		statusBadge,
		modeText,
		serverTypeText,
		interfaceTypeText,
		listenText,
		connectText,
		rxText,
		txText,
		uptimeText,
		copyBtn,
	)
}

func (sv *StatusView) Update(status TunnelStatus) {
	sv.mu.Lock()
	sv.status = status
	sv.rebuild()
	sv.mu.Unlock()
	sv.Container.Refresh()
}
