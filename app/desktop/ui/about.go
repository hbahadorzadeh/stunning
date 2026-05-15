package ui

import (
	"bytes"
	_ "embed"
	"image/color"
	"net/url"

	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/canvas"
	"fyne.io/fyne/v2/container"
	"fyne.io/fyne/v2/dialog"
	"fyne.io/fyne/v2/widget"
)

// SetAboutIcon allows setting icon from main package
func SetAboutIcon(data []byte) {
	aboutIconBytes = data
}

var aboutIconBytes []byte

func CreateAboutDialog(app fyne.App) {
	// Load icon if available
	var iconImg *canvas.Image
	if len(aboutIconBytes) > 0 {
		iconImg = canvas.NewImageFromReader(bytes.NewReader(aboutIconBytes), "icon.png")
	}

	// Create dialog content
	var content fyne.CanvasObject

	if iconImg != nil {
		// Icon (128x128)
		iconImg.Resize(fyne.NewSize(128, 128))
		icon := container.NewCenter(iconImg)

		title := canvas.NewText("Stunning Tunnel Manager", color.White)
		title.TextSize = 20
		title.TextStyle.Bold = true

		version := canvas.NewText("v0.1.0", color.RGBA{R: 180, G: 180, B: 180, A: 255})
		version.TextSize = 12

		description := widget.NewLabel("A modern, cross-platform network tunneling application")
		description.Alignment = fyne.TextAlignCenter

		githubURL, _ := url.Parse("https://github.com/hbahadorzadeh/stunning")
		githubLink := widget.NewHyperlink("GitHub Repository", githubURL)

		closeBtn := widget.NewButton("Close", func() {
			// Dialog will close automatically
		})

		content = container.NewVBox(
			container.NewCenter(icon),
			container.NewCenter(title),
			container.NewCenter(version),
			description,
			container.NewCenter(githubLink),
			container.NewCenter(closeBtn),
		)
	} else {
		title := canvas.NewText("Stunning Tunnel Manager", color.White)
		title.TextSize = 20
		title.TextStyle.Bold = true

		version := canvas.NewText("v0.1.0", color.RGBA{R: 180, G: 180, B: 180, A: 255})
		version.TextSize = 12

		description := widget.NewLabel("A modern, cross-platform network tunneling application")
		description.Alignment = fyne.TextAlignCenter

		githubURL, _ := url.Parse("https://github.com/hbahadorzadeh/stunning")
		githubLink := widget.NewHyperlink("GitHub Repository", githubURL)

		closeBtn := widget.NewButton("Close", func() {
			// Dialog will close automatically
		})

		content = container.NewVBox(
			title,
			version,
			description,
			githubLink,
			closeBtn,
		)
	}

	aboutDialog := dialog.NewCustom(
		"About Stunning",
		"Close",
		content,
		app.Driver().AllWindows()[0],
	)
	aboutDialog.Show()
}
