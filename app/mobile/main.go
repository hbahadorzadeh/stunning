package main

import (
	_ "embed"

	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/app"
	"fyne.io/fyne/v2/container"
	"github.com/hbahadorzadeh/stunning/app/mobile/ui"
	"github.com/hbahadorzadeh/stunning/app/mobile/vpn"
)

//go:embed assets/icon.png
var iconBytes []byte

func main() {
	myApp := app.New()

	// Initialize VPN provider
	vpn.SetProvider(vpn.NewStubProvider())

	// Create main window
	mainWindow := myApp.NewWindow("Stunning VPN")
	mainWindow.Resize(fyne.NewSize(400, 600))

	// Share icon with UI
	ui.SetIcon(iconBytes)

	// Create tab-based navigation
	connectScreen := ui.CreateConnectScreen(
		func(addr, proto string) error { return vpn.Connect(addr, proto) },
		func() error { return vpn.Disconnect() },
	)
	statusScreen := ui.CreateStatusScreen(vpn.IsConnected, vpn.GetError)
	aboutScreen := ui.CreateAboutScreen()

	tabs := container.NewAppTabs(
		container.NewTabItem("Connect", connectScreen),
		container.NewTabItem("Status", statusScreen),
		container.NewTabItem("About", aboutScreen),
	)
	tabs.SetTabLocation(container.TabLocationBottom)

	mainWindow.SetContent(tabs)
	mainWindow.ShowAndRun()
}
