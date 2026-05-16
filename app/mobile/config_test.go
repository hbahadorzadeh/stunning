package main

import (
	"testing"
)

func TestAddConfig(t *testing.T) {
	cm := &ConfigManager{configs: make(map[string]*VPNConfig)}

	config, err := cm.AddConfig("Test VPN", "test.example.com:443", "tls")
	if err != nil {
		t.Fatalf("AddConfig failed: %v", err)
	}
	if config.Name != "Test VPN" {
		t.Errorf("Expected name 'Test VPN', got '%s'", config.Name)
	}
	if config.Server != "test.example.com:443" {
		t.Errorf("Expected server 'test.example.com:443', got '%s'", config.Server)
	}
}

func TestAddConfigDuplicate(t *testing.T) {
	cm := &ConfigManager{configs: make(map[string]*VPNConfig)}

	cm.AddConfig("Test VPN", "test.example.com:443", "tls")
	_, err := cm.AddConfig("Test VPN", "other.example.com:443", "https")

	if err == nil {
		t.Errorf("Expected error for duplicate config, got nil")
	}
}

func TestGetAllConfigs(t *testing.T) {
	cm := &ConfigManager{configs: make(map[string]*VPNConfig)}

	cm.AddConfig("VPN1", "server1.com:443", "tls")
	cm.AddConfig("VPN2", "server2.com:443", "https")

	configs := cm.GetAllConfigs()
	if len(configs) != 2 {
		t.Errorf("Expected 2 configs, got %d", len(configs))
	}
}

func TestSetActiveConfig(t *testing.T) {
	cm := &ConfigManager{configs: make(map[string]*VPNConfig)}

	cfg1, _ := cm.AddConfig("VPN1", "server1.com:443", "tls")
	cfg2, _ := cm.AddConfig("VPN2", "server2.com:443", "https")

	cm.SetActiveConfig(cfg1.ID)
	active := cm.GetActiveConfig()

	if active == nil || active.ID != cfg1.ID {
		t.Errorf("Expected active config to be VPN1, got %v", active)
	}

	cm.SetActiveConfig(cfg2.ID)
	active = cm.GetActiveConfig()

	if active == nil || active.ID != cfg2.ID {
		t.Errorf("Expected active config to be VPN2, got %v", active)
	}
}

func TestDeleteConfig(t *testing.T) {
	cm := &ConfigManager{configs: make(map[string]*VPNConfig)}

	cfg, _ := cm.AddConfig("VPN1", "server1.com:443", "tls")
	err := cm.DeleteConfig(cfg.ID)

	if err != nil {
		t.Errorf("DeleteConfig failed: %v", err)
	}

	configs := cm.GetAllConfigs()
	if len(configs) != 0 {
		t.Errorf("Expected 0 configs after delete, got %d", len(configs))
	}
}

func TestUpdateConfig(t *testing.T) {
	cm := &ConfigManager{configs: make(map[string]*VPNConfig)}

	cfg, _ := cm.AddConfig("VPN1", "server1.com:443", "tls")
	updated, err := cm.UpdateConfig(cfg.ID, "Updated VPN", "newserver.com:8443", "https")

	if err != nil {
		t.Errorf("UpdateConfig failed: %v", err)
	}

	if updated.Name != "Updated VPN" {
		t.Errorf("Expected name 'Updated VPN', got '%s'", updated.Name)
	}

	if updated.Server != "newserver.com:8443" {
		t.Errorf("Expected server 'newserver.com:8443', got '%s'", updated.Server)
	}
}
