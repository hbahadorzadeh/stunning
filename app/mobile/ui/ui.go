package ui

import (
	"image/color"

	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/canvas"
)

var iconBytes []byte

// SetIcon sets the app icon
func SetIcon(bytes []byte) {
	iconBytes = bytes
}

// GetIconImage returns the app icon as a Fyne image
func GetIconImage() *canvas.Image {
	res := fyne.NewStaticResource("icon.png", iconBytes)
	return canvas.NewImageFromResource(res)
}

// ColorRunning is the color for running status
var ColorRunning = color.RGBA{R: 0, G: 200, B: 0, A: 255}

// ColorStopped is the color for stopped status
var ColorStopped = color.RGBA{R: 200, G: 0, B: 0, A: 255}

// ColorError is the color for error status
var ColorError = color.RGBA{R: 255, G: 165, B: 0, A: 255}
