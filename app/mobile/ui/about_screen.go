package ui

import (
	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/container"
	"fyne.io/fyne/v2/widget"
)

// CreateAboutScreen creates the about/info screen
func CreateAboutScreen() fyne.CanvasObject {
	// Icon
	iconImg := GetIconImage()
	iconImg.Resize(fyne.NewSize(128, 128))

	// Title
	titleText := widget.NewLabel("Stunning VPN")
	titleText.Alignment = fyne.TextAlignCenter
	titleText.TextStyle.Bold = true

	// Version
	versionText := widget.NewLabel("Version 1.0.0")
	versionText.Alignment = fyne.TextAlignCenter

	// Description
	descText := widget.NewLabel("A high-performance tunnel and VPN solution\nwith support for multiple protocols")
	descText.Alignment = fyne.TextAlignCenter

	// GitHub link
	githubBtn := widget.NewButton("Visit GitHub", func() {
		// In a real app, this would open the browser
		// exec.Command("xdg-open", "https://github.com/hbahadorzadeh/stunning").Run()
	})

	// Credits
	creditsText := widget.NewLabel("Built with Fyne and Go")

	content := container.NewVBox(
		iconImg,
		titleText,
		versionText,
		descText,
		githubBtn,
		creditsText,
	)

	return container.NewScroll(content)
}
