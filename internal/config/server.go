package config

import (
	"github.com/block-p/paqetManager/internal/network"
	"github.com/google/uuid"
)

// ServerConfig represents the complete paqet server configuration.
type ServerConfig struct {
	Role      string    `yaml:"role"`
	Log       Log       `yaml:"log"`
	Listen    Listen    `yaml:"listen"`
	Network   Network   `yaml:"network"`
	Transport Transport `yaml:"transport"`
}

// GenerateDefaultConfig creates a default ServerConfig initialized with network discovery.
func GenerateDefaultConfig(port string, opts ...Option) (*ServerConfig, error) {
	info, err := network.GetFullInfo()
	if err != nil {
		return nil, err
	}

	netConfig := Network{
		Interface: info.Interface,
		IPv4: IPv4Config{
			Addr:      info.IpAddr + ":" + port,
			RouterMAC: info.MacAddr,
		},
	}
	kcp := KCP{
		Mode:  "normal",
		MTU:   1350,
		Block: "aes",
		Key:   uuid.New().String(),
	}
	cfg := &ServerConfig{
		Role:      "server",
		Log:       Log{Level: "info"},
		Listen:    Listen{Addr: ":" + port},
		Network:   netConfig,
		Transport: Transport{Protocol: "kcp", Conn: 1, KCP: kcp},
	}
	for _, opt := range opts {
		opt.applyServer(cfg)
	}
	return cfg, nil
}

// GetRole returns the configuration role ("server").
func (c *ServerConfig) GetRole() string {
	return c.Role
}

// GetNetwork returns a pointer to the Network configuration.
func (c *ServerConfig) GetNetwork() *Network {
	return &c.Network
}

// SetRouterMAC updates the router MAC address.
func (c *ServerConfig) SetRouterMAC(mac string) {
	c.Network.IPv4.RouterMAC = mac
}

// RefreshRouterMAC re-discovers the current gateway MAC address from the network.
func (c *ServerConfig) RefreshRouterMAC() error {
	info, err := network.GetFullInfo()
	if err != nil {
		return err
	}
	c.Network.IPv4.RouterMAC = info.MacAddr
	return nil
}

// WriteConfig serializes ServerConfig to YAML and writes it to a file.
func (c *ServerConfig) WriteConfig(name string) error {
	return writeYAML(name, c)
}
