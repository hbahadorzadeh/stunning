package ui

import (
	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/container"
	"fyne.io/fyne/v2/widget"
)

// CreateConnectScreen creates the VPN connection screen
func CreateConnectScreen(connect func(string, string) error, disconnect func() error) fyne.CanvasObject {
	// Server address input
	serverEntry := widget.NewEntry()
	serverEntry.SetPlaceHolder("Enter server address")
	serverEntry.OnChanged = func(s string) {}

	// Protocol selector
	protocolSelect := widget.NewSelect(
		[]string{"tcp", "tls", "h2", "wss", "udp"},
		func(s string) {},
	)
	protocolSelect.SetSelected("tcp")

	// Status label
	statusLabel := widget.NewLabel("Disconnected")
	statusLabel.Alignment = fyne.TextAlignCenter

	// Connect button
	connectBtn := widget.NewButton("Connect", func() {
		if serverEntry.Text != "" && protocolSelect.Selected != "" {
			err := connect(serverEntry.Text, protocolSelect.Selected)
			if err == nil {
				statusLabel.SetText("Connecting...")
			} else {
				statusLabel.SetText("Connection failed: " + err.Error())
			}
			statusLabel.Refresh()
		}
	})

	// Disconnect button
	disconnectBtn := widget.NewButton("Disconnect", func() {
		disconnect()
		statusLabel.SetText("Disconnected")
		statusLabel.Refresh()
	})

	// Form container
	form := container.NewVBox(
		widget.NewCard("Server", "", serverEntry),
		widget.NewCard("Protocol", "", protocolSelect),
		widget.NewCard("Status", "", statusLabel),
		container.NewHBox(connectBtn, disconnectBtn),
	)

	return container.NewScroll(form)
}
