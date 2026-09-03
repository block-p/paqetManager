package config

import (
	"fmt"
	"os"

	"gopkg.in/yaml.v3"
)

// LoadConfig loads either a ServerConfig or ClientConfig from a YAML file based on its "role" field.
func LoadConfig(filePath string) (Config, error) {
	data, err := os.ReadFile(filePath)
	if err != nil {
		return nil, fmt.Errorf("failed to read config file %q: %w", filePath, err)
	}

	var probe struct {
		Role string `yaml:"role"`
	}
	if err := yaml.Unmarshal(data, &probe); err != nil {
		return nil, fmt.Errorf("failed to parse YAML in %q: %w", filePath, err)
	}

	switch probe.Role {
	case "server":
		var s ServerConfig
		if err := yaml.Unmarshal(data, &s); err != nil {
			return nil, fmt.Errorf("failed to parse server config in %q: %w", filePath, err)
		}
		return &s, nil
	case "client":
		var c ClientConfig
		if err := yaml.Unmarshal(data, &c); err != nil {
			return nil, fmt.Errorf("failed to parse client config in %q: %w", filePath, err)
		}
		return &c, nil
	default:
		return nil, fmt.Errorf("unknown or missing role %q in config file %q", probe.Role, filePath)
	}
}

// UpdateConfigFileMAC updates the RouterMAC in any config file (server or client) and writes it back to disk.
func UpdateConfigFileMAC(filePath string, newMAC string) error {
	cfg, err := LoadConfig(filePath)
	if err != nil {
		return err
	}
	cfg.SetRouterMAC(newMAC)
	return cfg.WriteConfig(filePath)
}

// RefreshConfigFileMAC re-discovers the current gateway MAC from the network and updates the config file on disk.
func RefreshConfigFileMAC(filePath string) error {
	cfg, err := LoadConfig(filePath)
	if err != nil {
		return err
	}
	if err := cfg.RefreshRouterMAC(); err != nil {
		return fmt.Errorf("failed to refresh MAC from network: %w", err)
	}
	return cfg.WriteConfig(filePath)
}
