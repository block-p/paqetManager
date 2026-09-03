# paqetManager

A supervisor and automated setup daemon for [paqet](https://github.com/hanselime/paqet) (high-speed network tunneling over KCP).

`paqetManager` bundles architecture-specific `paqet` binaries, automatically detects network routing and router MAC addresses (resolving ARP dynamically), generates configuration files, and supervises the child process with periodic router MAC refreshes and graceful restarts.

---

## Features

- **Self-Contained Deployment**: Bundles `paqet_linux_amd64` and `paqet_linux_arm64` via Go `embed`. Installs automatically to `/opt/paqet/paqet`.
- **Automatic Network Discovery**: Parses `/proc/net/route` and `/proc/net/arp` to auto-detect the default network interface, IP, and gateway MAC. Wakes up the ARP cache automatically via UDP probe.
- **Continuous Supervision & MAC Refresh**: Monitors the `paqet` process and automatically updates the router MAC address every hour, triggering a graceful restart to keep the tunnel stable across gateway shifts.
- **One-Click Shareable URIs**: Generates compact `paqet://` base64 URIs on server startup for instant client configuration.
- **Built-in SOCKS5 Inbound**: Configures a local SOCKS5 proxy on `127.0.0.1:1080` out-of-the-box in client mode.

---

## Prerequisites

- Linux OS (`amd64` or `arm64`)
- Root privileges (`sudo`) required for raw network interface bindings and writing to `/opt/paqet/`

---

## Installation & Building

```bash
git clone https://github.com/block-p/paqetManager.git
cd paqetManager

# Build binary
go build -o paqetManager cmd/main.go
```

---

## Usage Guide

### 1. Server Mode

Start a server on a given port. An authentication key is automatically generated using UUID v4 if not provided.

```bash
sudo ./paqetManager server -port 9999
```

#### Flags for Server:
| Flag | Default | Description |
|------|---------|-------------|
| `-port` | `9999` | UDP port for the server to listen on |
| `-key` | Auto (UUID) | Secret key for KCP encryption |
| `-conn` | `1` | Number of multiplexed KCP connections (1–256) |
| `-mode` | `normal` | KCP mode (`normal`, `fast1`, `fast2`, `fast3`, `manual`) |

On startup, the server prints a connection URI:
```text
paqet://OTEuMjI4LjE4Ni4xMTI6OTk5OQ==.bm9ybWFs.MQ==.YWVz.ZDBhM2Y1Zjkt....MTM1MA==
```
Keep this URI to configure clients easily.

---

### 2. Client Mode

#### Method A: Using Connection URI (Recommended)
Pass the generated `paqet://` URI directly:

```bash
sudo ./paqetManager client paqet://<URI>
# Or shorthand:
sudo ./paqetManager c paqet://<URI>
```

#### Method B: Using CLI Flags
```bash
sudo ./paqetManager client -addr 1.2.3.4:9999 -key "<YOUR_SECRET_KEY>"
# Optional flags:
#   -conn 2
#   -mode fast
```

Once connected, local traffic can be routed through SOCKS5 proxy:
```text
127.0.0.1:1080
```

---

### 3. Running as a Standalone Supervisor

If a config already exists in `/opt/paqet/config.yaml`, run without arguments:

```bash
sudo ./paqetManager
```

`paqetManager` will:
1. Load `/opt/paqet/config.yaml`.
2. Inspect and refresh the local router MAC address.
3. Start and supervise `paqet`.
4. Refresh MAC and gracefully restart every 1 hour or instantly upon unexpected exits.

---

### 4. Running as a Systemd Service

Create `/etc/systemd/system/paqet.service`:

```ini
[Unit]
Description=Paqet Tunnel Manager
After=network.target

[Service]
Type=simple
ExecStart=/usr/local/bin/paqetManager
Restart=always
RestartSec=5
LimitNOFILE=65535

[Install]
WantedBy=multi-user.target
```

Enable and start:
```bash
sudo mv paqetManager /usr/local/bin/
sudo systemctl daemon-reload
sudo systemctl enable --now paqet
```

---

## Configuration Reference (`/opt/paqet/config.yaml`)

### Server Example
```yaml
role: server
log:
  level: info
listen:
  addr: :9999
network:
  interface: eth0
  ipv4:
    addr: 91.228.186.112:9999
    router_mac: 92:7a:4e:fe:e1:8b
transport:
  protocol: kcp
  conn: 1
  kcp:
    mode: normal
    mtu: 1350
    block: aes
    key: d0a3f5f9-b883-4a0b-8d76-e578f7e7f7b1
```

### Client Example
```yaml
role: client
log:
  level: info
socks5:
  - listen: 127.0.0.1:1080
network:
  interface: eth0
  ipv4:
    addr: 192.168.1.50:0
    router_mac: aa:bb:cc:dd:ee:ff
server:
  addr: 91.228.186.112:9999
transport:
  protocol: kcp
  conn: 1
  kcp:
    mode: fast
    mtu: 1350
    key: d0a3f5f9-b883-4a0b-8d76-e578f7e7f7b1
```
