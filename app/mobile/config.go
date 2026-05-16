package main

import (
	"encoding/json"
	"errors"
	"sync"
)

// VPNConfig represents a saved VPN configuration
type VPNConfig struct {
	ID       string `json:"id"`
	Name     string `json:"name"`
	Server   string `json:"server"`
	Protocol string `json:"protocol"`
	Active   bool   `json:"active"`
}

// ConfigManager manages VPN configurations
type ConfigManager struct {
	mu      sync.RWMutex
	configs map[string]*VPNConfig
}

var configManager = &ConfigManager{
	configs: make(map[string]*VPNConfig),
}

// AddConfig adds a new VPN configuration
func (cm *ConfigManager) AddConfig(name, server, protocol string) (*VPNConfig, error) {
	cm.mu.Lock()
	defer cm.mu.Unlock()

	if name == "" || server == "" || protocol == "" {
		return nil, errors.New("missing required fields")
	}

	// Generate simple ID based on name
	id := name
	if _, exists := cm.configs[id]; exists {
		return nil, errors.New("configuration already exists")
	}

	config := &VPNConfig{
		ID:       id,
		Name:     name,
		Server:   server,
		Protocol: protocol,
		Active:   false,
	}

	cm.configs[id] = config
	return config, nil
}

// GetConfig retrieves a configuration by ID
func (cm *ConfigManager) GetConfig(id string) (*VPNConfig, error) {
	cm.mu.RLock()
	defer cm.mu.RUnlock()

	config, exists := cm.configs[id]
	if !exists {
		return nil, errors.New("configuration not found")
	}
	return config, nil
}

// GetAllConfigs returns all configurations
func (cm *ConfigManager) GetAllConfigs() []*VPNConfig {
	cm.mu.RLock()
	defer cm.mu.RUnlock()

	configs := make([]*VPNConfig, 0, len(cm.configs))
	for _, config := range cm.configs {
		configs = append(configs, config)
	}
	return configs
}

// UpdateConfig updates an existing configuration
func (cm *ConfigManager) UpdateConfig(id, name, server, protocol string) (*VPNConfig, error) {
	cm.mu.Lock()
	defer cm.mu.Unlock()

	config, exists := cm.configs[id]
	if !exists {
		return nil, errors.New("configuration not found")
	}

	if name != "" {
		config.Name = name
	}
	if server != "" {
		config.Server = server
	}
	if protocol != "" {
		config.Protocol = protocol
	}

	return config, nil
}

// DeleteConfig removes a configuration
func (cm *ConfigManager) DeleteConfig(id string) error {
	cm.mu.Lock()
	defer cm.mu.Unlock()

	if _, exists := cm.configs[id]; !exists {
		return errors.New("configuration not found")
	}

	delete(cm.configs, id)
	return nil
}

// SetActiveConfig sets a configuration as active
func (cm *ConfigManager) SetActiveConfig(id string) error {
	cm.mu.Lock()
	defer cm.mu.Unlock()

	// Deactivate all other configs
	for _, config := range cm.configs {
		config.Active = false
	}

	config, exists := cm.configs[id]
	if !exists {
		return errors.New("configuration not found")
	}

	config.Active = true
	return nil
}

// GetActiveConfig returns the active configuration
func (cm *ConfigManager) GetActiveConfig() *VPNConfig {
	cm.mu.RLock()
	defer cm.mu.RUnlock()

	for _, config := range cm.configs {
		if config.Active {
			return config
		}
	}
	return nil
}

// ExportConfig exports configuration as JSON
func (c *VPNConfig) ToJSON() string {
	data, _ := json.Marshal(c)
	return string(data)
}

// GetManager returns the global config manager
func GetManager() *ConfigManager {
	return configManager
}
