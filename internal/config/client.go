package config

import (
	"github.com/block-p/paqetManager/internal/network"
)

// ClientConfig represents the complete paqet client configuration.
type ClientConfig struct {
	Role      string         `yaml:"role"`
	Log       Log            `yaml:"log"`
	Socks5    []Socks5Config `yaml:"socks5,omitempty"`
	Network   Network        `yaml:"network"`
	Server    Server         `yaml:"server"`
	Transport Transport      `yaml:"transport"`
}

// DefaultClientConfig creates a default ClientConfig initialized with network discovery.
func DefaultClientConfig(serverAddr, secretKey string, opts ...Option) (*ClientConfig, error) {
	info, err := network.GetFullInfo()
	if err != nil {
		return nil, err
	}

	netConfig := Network{
		Interface: info.Interface,
		IPv4: IPv4Config{
			Addr:      info.IpAddr + ":0",
			RouterMAC: info.MacAddr,
		},
	}
	cfg := &ClientConfig{
		Role: "client",
		Log: Log{
			Level: "info",
		},
		Socks5: []Socks5Config{
			{Listen: "127.0.0.1:1080"},
		},
		Network: netConfig,
		Server: Server{
			Addr: serverAddr,
		},
		Transport: Transport{
			Protocol: "kcp",
			Conn:     1,
			KCP: KCP{
				Mode: "fast",
				Key:  secretKey,
				MTU:  1350,
			},
		},
	}
	for _, opt := range opts {
		opt.applyClient(cfg)
	}
	return cfg, nil
}

// GetRole returns the configuration role ("client").
func (c *ClientConfig) GetRole() string {
	return c.Role
}

// GetNetwork returns a pointer to the Network configuration.
func (c *ClientConfig) GetNetwork() *Network {
	return &c.Network
}

// SetRouterMAC updates the router MAC address.
func (c *ClientConfig) SetRouterMAC(mac string) {
	c.Network.IPv4.RouterMAC = mac
}

// RefreshRouterMAC re-discovers the current gateway MAC address from the network.
func (c *ClientConfig) RefreshRouterMAC() error {
	info, err := network.GetFullInfo()
	if err != nil {
		return err
	}
	c.Network.IPv4.RouterMAC = info.MacAddr
	return nil
}

// WriteConfig serializes ClientConfig to YAML and writes it to a file.
func (c *ClientConfig) WriteConfig(name string) error {
	return writeYAML(name, c)
}
