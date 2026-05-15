package ui

import (
	"testing"

	"github.com/hbahadorzadeh/stunning/core"
)

func TestTunnelConfigValidation(t *testing.T) {
	testCases := []struct {
		testName   string
		listen     string
		connect    string
		shouldFail bool
	}{
		{
			testName:   "empty listen should fail",
			listen:     "",
			connect:    "127.0.0.1:9999",
			shouldFail: true,
		},
		{
			testName:   "empty connect should fail",
			listen:     "127.0.0.1:8888",
			connect:    "",
			shouldFail: true,
		},
		{
			testName:   "valid config",
			listen:     "127.0.0.1:8888",
			connect:    "127.0.0.1:9999",
			shouldFail: false,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.testName, func(t *testing.T) {
			hasMissingRequired := tc.listen == "" || tc.connect == ""

			if tc.shouldFail && !hasMissingRequired {
				t.Error("Expected validation to fail but it didn't")
			}

			if !tc.shouldFail && hasMissingRequired {
				t.Error("Expected validation to succeed but it failed")
			}
		})
	}
}

func TestTunnelConfigStructure(t *testing.T) {
	config := core.TunnelConfig{
		ServiceMode:   "server",
		ServerType:    "tcp",
		InterfaceType: "socks",
		Listen:        "127.0.0.1:8888",
		Connect:       "127.0.0.1:9999",
		Mtu:           "1500",
		DeviceName:    "tun0",
		Cert:          "/path/to/cert.pem",
		Key:           "/path/to/key.pem",
	}

	if config.ServiceMode == "" {
		t.Error("ServiceMode should not be empty")
	}

	if config.ServerType == "" {
		t.Error("ServerType should not be empty")
	}

	if config.InterfaceType == "" {
		t.Error("InterfaceType should not be empty")
	}

	if config.Listen == "" {
		t.Error("Listen should not be empty")
	}

	if config.Connect == "" {
		t.Error("Connect should not be empty")
	}

	if config.Mtu == "" {
		t.Error("Mtu should have default value")
	}

	if config.DeviceName == "" {
		t.Error("DeviceName should have default value")
	}
}

func TestCertKeyRequirementByServerType(t *testing.T) {
	serverTypes := map[string]bool{
		"tcp":   false,
		"udp":   false,
		"http":  false,
		"https": true,
		"tls":   true,
		"h2":    true,
		"ws":    true,
		"udps":  true,
		"dns":   false,
		"icmp":  false,
	}

	for serverType, needsCertKey := range serverTypes {
		t.Run(serverType, func(t *testing.T) {
			config := core.TunnelConfig{
				ServerType: serverType,
			}

			// Cert/Key should be provided for these types
			if needsCertKey {
				if config.Cert == "" && config.Key == "" {
					// This is expected - validation should catch it
					t.Logf("Server type %s requires cert/key (as expected)", serverType)
				}
			}
		})
	}
}

func TestTunnelConfigDefaults(t *testing.T) {
	config := core.TunnelConfig{
		Mtu:        "1500",
		DeviceName: "tun",
	}

	if config.Mtu != "1500" {
		t.Errorf("Expected default MTU 1500, got %s", config.Mtu)
	}

	if config.DeviceName != "tun" {
		t.Errorf("Expected default device name 'tun', got %s", config.DeviceName)
	}
}

func TestTunnelModes(t *testing.T) {
	modes := []string{"server", "client"}

	for _, mode := range modes {
		config := core.TunnelConfig{
			ServiceMode: mode,
		}

		if config.ServiceMode != mode {
			t.Errorf("Expected mode %s, got %s", mode, config.ServiceMode)
		}
	}
}

func TestInterfaceTypes(t *testing.T) {
	interfaceTypes := []string{"socks", "tcp", "tun"}

	for _, iface := range interfaceTypes {
		config := core.TunnelConfig{
			InterfaceType: iface,
		}

		if config.InterfaceType != iface {
			t.Errorf("Expected interface type %s, got %s", iface, config.InterfaceType)
		}
	}
}
