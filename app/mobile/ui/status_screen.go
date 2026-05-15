package ui

import (
	"time"

	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/canvas"
	"fyne.io/fyne/v2/container"
	"fyne.io/fyne/v2/widget"
)

// CreateStatusScreen creates the VPN status display screen
func CreateStatusScreen(isConnected func() bool, getError func() string) fyne.CanvasObject {
	// Status indicator
	statusCircle := canvas.NewCircle(ColorStopped)
	statusCircle.StrokeWidth = 2

	// Status text
	statusText := widget.NewLabel("Disconnected")
	statusText.Alignment = fyne.TextAlignCenter

	// Error message (if any)
	errorLabel := widget.NewLabel("")
	errorLabel.Alignment = fyne.TextAlignCenter

	// Update status periodically
	go func() {
		ticker := time.NewTicker(500 * time.Millisecond)
		defer ticker.Stop()

		for range ticker.C {
			connected := isConnected()
			if connected {
				statusCircle.FillColor = ColorRunning
				statusText.SetText("Connected")
				errorLabel.SetText("")
			} else {
				statusCircle.FillColor = ColorStopped
				statusText.SetText("Disconnected")
				if errMsg := getError(); errMsg != "" {
					errorLabel.SetText("Error: " + errMsg)
					statusCircle.FillColor = ColorError
				}
			}
			statusCircle.Refresh()
			statusText.Refresh()
			errorLabel.Refresh()
		}
	}()

	// Layout
	statusSection := container.NewVBox(
		statusCircle,
		statusText,
		errorLabel,
	)

	return container.NewVBox(
		widget.NewCard("Connection Status", "", statusSection),
	)
}
