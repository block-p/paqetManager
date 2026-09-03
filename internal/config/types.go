package config

import (
	"bufio"
	"os"

	"gopkg.in/yaml.v3"
)

// Log represents the logging configuration.
type Log struct {
	Level string `yaml:"level"` // none, debug, info, warn, error, fatal
}

// Listen represents the server listening address.
type Listen struct {
	Addr string `yaml:"addr"` // e.g. ":8080"
}

// IPv4Config represents IPv4 address and router MAC configuration.
type IPv4Config struct {
	Addr      string `yaml:"addr"`       // e.g. "91.228.186.112:0"
	RouterMAC string `yaml:"router_mac"` // e.g. "92:7a:4e:fe:e1:8b"
}

// Network represents the network interface and IPv4 settings.
type Network struct {
	Interface string     `yaml:"interface"` // e.g. "eth0"
	IPv4      IPv4Config `yaml:"ipv4"`
}

// Transport represents the transport protocol configuration.
type Transport struct {
	Protocol string `yaml:"protocol"` // "kcp"
	Conn     int    `yaml:"conn"`     // 1-256
	KCP      KCP    `yaml:"kcp"`
}

// KCP represents KCP protocol parameters.
type KCP struct {
	Mode  string `yaml:"mode"`            // normal, fast1, fast2, fast3, manual
	MTU   int    `yaml:"mtu"`             // 50-1500
	Block string `yaml:"block,omitempty"` // aes, aes-128, etc.
	Key   string `yaml:"key"`             // password
}

// Socks5Config represents client SOCKS5 inbound proxy settings.
type Socks5Config struct {
	Listen   string `yaml:"listen"`             // e.g. "127.0.0.1:1080"
	Username string `yaml:"username,omitempty"` // Optional SOCKS5 auth
	Password string `yaml:"password,omitempty"` // Optional SOCKS5 auth
}

// Server represents the remote server address for a client.
type Server struct {
	Addr string `yaml:"addr"` // e.g. "127.0.0.1:8080"
}

// Config represents a common interface implemented by both ServerConfig and ClientConfig.
type Config interface {
	GetRole() string
	GetNetwork() *Network
	SetRouterMAC(mac string)
	RefreshRouterMAC() error
	WriteConfig(filePath string) error
}

// Option provides a unified functional option interface that can configure
// both ServerConfig and ClientConfig without name collisions.
type Option interface {
	applyServer(*ServerConfig)
	applyClient(*ClientConfig)
}

type optionFunc struct {
	serverFn func(*ServerConfig)
	clientFn func(*ClientConfig)
}

func (f optionFunc) applyServer(s *ServerConfig) {
	if f.serverFn != nil {
		f.serverFn(s)
	}
}

func (f optionFunc) applyClient(c *ClientConfig) {
	if f.clientFn != nil {
		f.clientFn(c)
	}
}

func writeYAML(filePath string, data any) error {
	yml, err := yaml.Marshal(data)
	if err != nil {
		return err
	}
	fs, err := os.Create(filePath)
	if err != nil {
		return err
	}
	defer fs.Close()

	writer := bufio.NewWriter(fs)
	if _, err = writer.Write(yml); err != nil {
		return err
	}
	return writer.Flush()
}
