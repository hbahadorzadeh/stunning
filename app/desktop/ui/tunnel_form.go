package ui

import (
	"fmt"
	"os"

	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/container"
	"fyne.io/fyne/v2/dialog"
	"fyne.io/fyne/v2/widget"
	"github.com/hbahadorzadeh/stunning/core"
)

func CreateTunnelDialog(app fyne.App, onSubmit func(core.TunnelConfig, string) error) {
	nameEntry := widget.NewEntry()
	nameEntry.SetPlaceHolder("Tunnel name")

	serviceModeSelect := widget.NewSelect([]string{"server", "client"}, func(s string) {})
	serviceModeSelect.SetSelected("server")

	serverTypeSelect := widget.NewSelect([]string{
		"tcp", "tls", "https", "h2", "ws", "udps", "dns", "icmp", "http", "udp",
	}, func(s string) {})
	serverTypeSelect.SetSelected("tcp")

	interfaceTypeSelect := widget.NewSelect([]string{
		"socks", "tcp", "tun",
	}, func(s string) {})
	interfaceTypeSelect.SetSelected("socks")

	listenEntry := widget.NewEntry()
	listenEntry.SetPlaceHolder("0.0.0.0:8888")

	connectEntry := widget.NewEntry()
	connectEntry.SetPlaceHolder("127.0.0.1:9999")

	certEntry := widget.NewEntry()
	certEntry.SetPlaceHolder("Path to certificate")

	certBtn := widget.NewButton("Browse Cert", func() {
		fd := dialog.NewFileOpen(func(reader fyne.URIReadCloser, err error) {
			if err == nil && reader != nil {
				certEntry.SetText(reader.URI().Path())
				reader.Close()
			}
		}, app.Driver().AllWindows()[0])
		fd.Show()
	})

	keyEntry := widget.NewEntry()
	keyEntry.SetPlaceHolder("Path to key")

	keyBtn := widget.NewButton("Browse Key", func() {
		fd := dialog.NewFileOpen(func(reader fyne.URIReadCloser, err error) {
			if err == nil && reader != nil {
				keyEntry.SetText(reader.URI().Path())
				reader.Close()
			}
		}, app.Driver().AllWindows()[0])
		fd.Show()
	})

	mtuEntry := widget.NewEntry()
	mtuEntry.SetText("1500")
	mtuEntry.SetPlaceHolder("1500")

	deviceNameEntry := widget.NewEntry()
	deviceNameEntry.SetText("tun")
	deviceNameEntry.SetPlaceHolder("tun")

	errorLabel := widget.NewLabel("")

	form := widget.NewForm(
		widget.NewFormItem("Name", nameEntry),
		widget.NewFormItem("Service Mode", serviceModeSelect),
		widget.NewFormItem("Server Type", serverTypeSelect),
		widget.NewFormItem("Interface Type", interfaceTypeSelect),
		widget.NewFormItem("Listen Address", listenEntry),
		widget.NewFormItem("Connect Address", connectEntry),
		widget.NewFormItem("Certificate", container.NewHBox(certEntry, certBtn)),
		widget.NewFormItem("Key", container.NewHBox(keyEntry, keyBtn)),
		widget.NewFormItem("MTU", mtuEntry),
		widget.NewFormItem("Device Name", deviceNameEntry),
	)

	submitBtn := widget.NewButton("Submit", func() {
		// Validation
		if nameEntry.Text == "" {
			errorLabel.SetText("Name is required")
			return
		}

		if listenEntry.Text == "" {
			errorLabel.SetText("Listen address is required")
			return
		}

		if connectEntry.Text == "" {
			errorLabel.SetText("Connect address is required")
			return
		}

		// Check if cert/key is required
		serverType := serverTypeSelect.Selected
		if serverType == "tls" || serverType == "https" || serverType == "h2" || serverType == "ws" || serverType == "udps" {
			if certEntry.Text == "" {
				errorLabel.SetText(fmt.Sprintf("%s server type requires a certificate", serverType))
				return
			}
			if keyEntry.Text == "" {
				errorLabel.SetText(fmt.Sprintf("%s server type requires a key", serverType))
				return
			}

			// Verify files exist
			if _, err := os.Stat(certEntry.Text); os.IsNotExist(err) {
				errorLabel.SetText("Certificate file does not exist")
				return
			}

			if _, err := os.Stat(keyEntry.Text); os.IsNotExist(err) {
				errorLabel.SetText("Key file does not exist")
				return
			}
		}

		config := core.TunnelConfig{
			Cert:          certEntry.Text,
			Connect:       connectEntry.Text,
			DeviceName:    deviceNameEntry.Text,
			InterfaceType: interfaceTypeSelect.Selected,
			Key:           keyEntry.Text,
			Listen:        listenEntry.Text,
			Mtu:           mtuEntry.Text,
			ServerType:    serverTypeSelect.Selected,
			ServiceMode:   serviceModeSelect.Selected,
		}

		if err := onSubmit(config, nameEntry.Text); err != nil {
			errorLabel.SetText(fmt.Sprintf("Error: %v", err))
			return
		}

		// Close the dialog if submit was successful
		for _, window := range app.Driver().AllWindows() {
			if window.Title() == "New Tunnel" {
				window.Close()
			}
		}
	})

	cancelBtn := widget.NewButton("Cancel", func() {
		for _, window := range app.Driver().AllWindows() {
			if window.Title() == "New Tunnel" {
				window.Close()
			}
		}
	})

	buttonBox := container.NewHBox(submitBtn, cancelBtn)

	dialogContent := container.NewVBox(
		errorLabel,
		form,
		buttonBox,
	)

	dialogWin := app.NewWindow("New Tunnel")
	dialogWin.SetContent(dialogContent)
	dialogWin.Resize(fyne.NewSize(500, 600))
	dialogWin.Show()
}
