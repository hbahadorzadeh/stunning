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

	// Update status periodically with proper cleanup
	ticker := time.NewTicker(500 * time.Millisecond)
	done := make(chan struct{})

	go func() {
		defer ticker.Stop()

		for {
			select {
			case <-ticker.C:
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
			case <-done:
				return
			}
		}
	}()

	// Store the done channel in the container for cleanup (if needed)
	// This allows callers to stop the ticker if needed

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
