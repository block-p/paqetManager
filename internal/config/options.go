package config

// Helper constructors for common options targeting sub-structures
func withTransport(fn func(*Transport)) Option {
	return optionFunc{
		serverFn: func(s *ServerConfig) { fn(&s.Transport) },
		clientFn: func(c *ClientConfig) { fn(&c.Transport) },
	}
}

func withLog(fn func(*Log)) Option {
	return optionFunc{
		serverFn: func(s *ServerConfig) { fn(&s.Log) },
		clientFn: func(c *ClientConfig) { fn(&c.Log) },
	}
}

func withNetwork(fn func(*Network)) Option {
	return optionFunc{
		serverFn: func(s *ServerConfig) { fn(&s.Network) },
		clientFn: func(c *ClientConfig) { fn(&c.Network) },
	}
}

// --- Common Options (Apply to both ServerConfig and ClientConfig) ---

// WithMode sets the KCP mode (normal, fast1, fast2, fast3, manual).
func WithMode(mode string) Option {
	validModes := map[string]bool{
		"normal": true, "fast1": true, "fast2": true, "fast3": true, "manual": true,
	}
	if !validModes[mode] {
		mode = "normal"
	}
	return withTransport(func(t *Transport) {
		t.KCP.Mode = mode
	})
}

// WithConn sets the number of connections (1-256).
func WithConn(conn int) Option {
	if conn < 1 || conn > 256 {
		conn = 1
	}
	return withTransport(func(t *Transport) {
		t.Conn = conn
	})
}

// WithKey sets the secret key for KCP encryption.
func WithKey(key string) Option {
	return withTransport(func(t *Transport) {
		t.KCP.Key = key
	})
}

// WithMTU sets the MTU for KCP (50-1500).
func WithMTU(mtu int) Option {
	if mtu < 50 || mtu > 1500 {
		mtu = 1350
	}
	return withTransport(func(t *Transport) {
		t.KCP.MTU = mtu
	})
}

// WithBlock sets the encryption block algorithm (e.g. "aes", "aes-128").
func WithBlock(block string) Option {
	return withTransport(func(t *Transport) {
		t.KCP.Block = block
	})
}

// WithLogLevel sets the logging level (none, debug, info, warn, error, fatal).
func WithLogLevel(level string) Option {
	return withLog(func(l *Log) {
		l.Level = level
	})
}

// WithInterface sets the network interface name.
func WithInterface(iface string) Option {
	return withNetwork(func(n *Network) {
		n.Interface = iface
	})
}

// WithRouterMAC sets the router MAC address.
func WithRouterMAC(mac string) Option {
	return withNetwork(func(n *Network) {
		n.IPv4.RouterMAC = mac
	})
}

// --- Server-Specific Options ---

// WithListenAddr sets the listening address for the server.
func WithListenAddr(addr string) Option {
	return optionFunc{
		serverFn: func(s *ServerConfig) {
			s.Listen.Addr = addr
		},
	}
}

// --- Client-Specific Options ---

// WithServerAddr sets the remote server address for the client.
func WithServerAddr(addr string) Option {
	return optionFunc{
		clientFn: func(c *ClientConfig) {
			c.Server.Addr = addr
		},
	}
}

// WithSocks5 adds a SOCKS5 proxy configuration to the client.
func WithSocks5(listen, username, password string) Option {
	return optionFunc{
		clientFn: func(c *ClientConfig) {
			c.Socks5 = append(c.Socks5, Socks5Config{
				Listen:   listen,
				Username: username,
				Password: password,
			})
		},
	}
}
