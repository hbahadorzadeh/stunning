package ui

import (
	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/container"
	"fyne.io/fyne/v2/widget"
)

// ConfigManager interface for managing VPN configs
type ConfigManager interface {
	AddConfig(name, server, protocol string) (interface{}, error)
	GetAllConfigs() []interface{}
	DeleteConfig(id string) error
	SetActiveConfig(id string) error
	GetActiveConfig() interface{}
	UpdateConfig(id, name, server, protocol string) (interface{}, error)
}

// ConfigData represents a VPN configuration
type ConfigData struct {
	ID       string
	Name     string
	Server   string
	Protocol string
	Active   bool
}

// CreateConfigScreen creates the VPN configuration management screen
func CreateConfigScreen(configMgr ConfigManager) fyne.CanvasObject {
	var selectedIndex int
	var configItems []*ConfigData
	var configItemButtons []*widget.Button

	// Refresh the config list
	refreshConfigList := func() {
		configItems = []*ConfigData{}
		configItemButtons = []*widget.Button{}

		configs := configMgr.GetAllConfigs()
		for i, cfg := range configs {
			if config, ok := cfg.(*ConfigData); ok {
				configItems = append(configItems, config)

				// Create button for this config
				btnText := config.Name + " (" + config.Protocol + ")"
				if config.Active {
					btnText = "✓ " + btnText
				}
				btn := widget.NewButton(btnText, func(index int) func() {
					return func() {
						selectedIndex = index
					}
				}(i))
				configItemButtons = append(configItemButtons, btn)
			}
		}
	}

	refreshConfigList()

	// Config list container
	configListBox := container.NewVBox(
		widget.NewLabel("Saved Configurations:"),
	)
	refreshConfigListUI := func() {
		configListBox.RemoveAll()
		configListBox.Add(widget.NewLabel("Saved Configurations:"))
		for _, btn := range configItemButtons {
			configListBox.Add(btn)
		}
		configListBox.Refresh()
	}
	refreshConfigListUI()

	// Config form for adding
	nameEntry := widget.NewEntry()
	nameEntry.SetPlaceHolder("Configuration name")

	serverEntry := widget.NewEntry()
	serverEntry.SetPlaceHolder("Server address (e.g., vpn.example.com:443)")

	protocolSelect := widget.NewSelect(
		[]string{"tcp", "tls", "h2", "wss", "https", "udp"},
		func(s string) {},
	)
	protocolSelect.SetSelected("tls")

	statusLabel := widget.NewLabel("No active configuration")

	addButton := widget.NewButton("Add Configuration", func() {
		if nameEntry.Text != "" && serverEntry.Text != "" && protocolSelect.Selected != "" {
			configMgr.AddConfig(nameEntry.Text, serverEntry.Text, protocolSelect.Selected)
			nameEntry.SetText("")
			serverEntry.SetText("")
			protocolSelect.SetSelected("tls")
			refreshConfigList()
			refreshConfigListUI()
		}
	})

	deleteButton := widget.NewButton("Delete Selected", func() {
		if selectedIndex >= 0 && selectedIndex < len(configItems) {
			configMgr.DeleteConfig(configItems[selectedIndex].ID)
			refreshConfigList()
			refreshConfigListUI()
		}
	})

	activateButton := widget.NewButton("Activate Selected", func() {
		if selectedIndex >= 0 && selectedIndex < len(configItems) {
			configMgr.SetActiveConfig(configItems[selectedIndex].ID)
			refreshConfigList()
			refreshConfigListUI()
		}
	})

	// Update status label
	updateStatus := func() {
		activeConfig := configMgr.GetActiveConfig()
		if config, ok := activeConfig.(*ConfigData); ok && config != nil {
			statusLabel.SetText("Active: " + config.Name + " (" + config.Server + ")")
		} else {
			statusLabel.SetText("No active configuration")
		}
		statusLabel.Refresh()
	}

	updateStatus()

	// Main form
	form := container.NewVBox(
		widget.NewCard("Configuration List", "", container.NewVBox(
			configListBox,
			container.NewHBox(activateButton, deleteButton),
			statusLabel,
		)),
		widget.NewCard("Add New Configuration", "", container.NewVBox(
			nameEntry,
			serverEntry,
			protocolSelect,
			addButton,
		)),
	)

	return container.NewScroll(form)
}
