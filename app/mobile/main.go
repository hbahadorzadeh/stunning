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

// ConfigManagerAdapter adapts ConfigManager to ui.ConfigManager interface
type ConfigManagerAdapter struct {
	impl *ConfigManager
}

// NewConfigManagerAdapter creates an adapter
func NewConfigManagerAdapter(cm *ConfigManager) *ConfigManagerAdapter {
	return &ConfigManagerAdapter{impl: cm}
}

// AddConfig adds a config and returns it
func (ca *ConfigManagerAdapter) AddConfig(name, server, protocol string) (interface{}, error) {
	config, err := ca.impl.AddConfig(name, server, protocol)
	if err != nil {
		return nil, err
	}
	return &ui.ConfigData{
		ID:       config.ID,
		Name:     config.Name,
		Server:   config.Server,
		Protocol: config.Protocol,
		Active:   config.Active,
	}, nil
}

// GetAllConfigs returns all configs as interface{}
func (ca *ConfigManagerAdapter) GetAllConfigs() []interface{} {
	configs := ca.impl.GetAllConfigs()
	result := make([]interface{}, len(configs))
	for i, cfg := range configs {
		result[i] = &ui.ConfigData{
			ID:       cfg.ID,
			Name:     cfg.Name,
			Server:   cfg.Server,
			Protocol: cfg.Protocol,
			Active:   cfg.Active,
		}
	}
	return result
}

// DeleteConfig deletes a config
func (ca *ConfigManagerAdapter) DeleteConfig(id string) error {
	return ca.impl.DeleteConfig(id)
}

// SetActiveConfig sets a config as active
func (ca *ConfigManagerAdapter) SetActiveConfig(id string) error {
	return ca.impl.SetActiveConfig(id)
}

// GetActiveConfig returns the active config
func (ca *ConfigManagerAdapter) GetActiveConfig() interface{} {
	cfg := ca.impl.GetActiveConfig()
	if cfg == nil {
		return nil
	}
	return &ui.ConfigData{
		ID:       cfg.ID,
		Name:     cfg.Name,
		Server:   cfg.Server,
		Protocol: cfg.Protocol,
		Active:   cfg.Active,
	}
}

// UpdateConfig updates a config
func (ca *ConfigManagerAdapter) UpdateConfig(id, name, server, protocol string) (interface{}, error) {
	config, err := ca.impl.UpdateConfig(id, name, server, protocol)
	if err != nil {
		return nil, err
	}
	return &ui.ConfigData{
		ID:       config.ID,
		Name:     config.Name,
		Server:   config.Server,
		Protocol: config.Protocol,
		Active:   config.Active,
	}, nil
}

func main() {
	myApp := app.New()

	// VPN provider is initialized via build tags (vpn_stub.go, vpn_android.go, vpn_ios.go)
	// No explicit initialization needed here

	// Create main window
	mainWindow := myApp.NewWindow("Stunning VPN")
	mainWindow.Resize(fyne.NewSize(400, 600))

	// Share icon with UI
	ui.SetIcon(iconBytes)

	// Initialize with sample config for testing
	configMgrImpl := GetManager()
	configMgrImpl.AddConfig("Work Server", "vpn.company.com:443", "tls")
	configMgrImpl.AddConfig("Home Server", "home.vpn:8443", "https")

	// Create tab-based navigation with multiple configs support
	configScreen := ui.CreateConfigScreen(NewConfigManagerAdapter(configMgrImpl))
	statusScreen := ui.CreateStatusScreen(vpn.IsConnected, vpn.GetError)
	aboutScreen := ui.CreateAboutScreen()

	tabs := container.NewAppTabs(
		container.NewTabItem("Configs", configScreen),
		container.NewTabItem("Status", statusScreen),
		container.NewTabItem("About", aboutScreen),
	)
	tabs.SetTabLocation(container.TabLocationBottom)

	mainWindow.SetContent(tabs)
	mainWindow.ShowAndRun()
}
