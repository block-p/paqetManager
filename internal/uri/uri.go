package uri

import (
	"encoding/base64"
	"fmt"
	"strconv"
	"strings"

	"github.com/block-p/paqetManager/internal/config"
)

// MakeUri generates a paqet URI from a ServerConfig.
func MakeUri(server config.ServerConfig) string {
	parts := []string{
		base64.StdEncoding.EncodeToString([]byte(server.Network.IPv4.Addr)),
		base64.StdEncoding.EncodeToString([]byte(server.Transport.KCP.Mode)),
		base64.StdEncoding.EncodeToString([]byte(strconv.Itoa(server.Transport.Conn))),
		base64.StdEncoding.EncodeToString([]byte(server.Transport.KCP.Block)),
		base64.StdEncoding.EncodeToString([]byte(server.Transport.KCP.Key)),
		base64.StdEncoding.EncodeToString([]byte(strconv.Itoa(server.Transport.KCP.MTU))),
	}
	return "paqet://" + strings.Join(parts, ".")
}

// UriData holds the decoded configuration parameters from a paqet URI.
type UriData struct {
	Addr  string
	Mode  string
	Conn  int
	Block string
	Key   string
	Mtu   int
}

func decodeBase64(s string) ([]byte, error) {
	// Support standard base64 with padding
	if data, err := base64.StdEncoding.DecodeString(s); err == nil {
		return data, nil
	}
	// Support unpadded raw standard base64
	if data, err := base64.RawStdEncoding.DecodeString(s); err == nil {
		return data, nil
	}
	// Support URL-safe base64 with padding
	if data, err := base64.URLEncoding.DecodeString(s); err == nil {
		return data, nil
	}
	// Support unpadded raw URL-safe base64
	return base64.RawURLEncoding.DecodeString(s)
}

// DecodeUri parses and decodes a paqet URI into UriData.
func DecodeUri(uri string) (UriData, error) {
	if !strings.HasPrefix(uri, "paqet://") {
		return UriData{}, fmt.Errorf("invalid URI scheme: must start with 'paqet://'")
	}

	trimmed := strings.TrimPrefix(uri, "paqet://")
	u := strings.Split(trimmed, ".")
	if len(u) < 5 {
		return UriData{}, fmt.Errorf("invalid URI format: expected at least 5 parts, got %d", len(u))
	}

	addr, err := decodeBase64(u[0])
	if err != nil {
		return UriData{}, fmt.Errorf("failed to decode addr: %w", err)
	}

	mode, err := decodeBase64(u[1])
	if err != nil {
		return UriData{}, fmt.Errorf("failed to decode mode: %w", err)
	}

	connstr, err := decodeBase64(u[2])
	if err != nil {
		return UriData{}, fmt.Errorf("failed to decode conn: %w", err)
	}
	conn, err := strconv.Atoi(string(connstr))
	if err != nil {
		return UriData{}, fmt.Errorf("invalid conn number %q: %w", string(connstr), err)
	}

	block, err := decodeBase64(u[3])
	if err != nil {
		return UriData{}, fmt.Errorf("failed to decode block: %w", err)
	}

	key, err := decodeBase64(u[4])
	if err != nil {
		return UriData{}, fmt.Errorf("failed to decode key: %w", err)
	}

	mtu := 1350
	if len(u) >= 6 {
		mtubyte, err := decodeBase64(u[5])
		if err != nil {
			return UriData{}, fmt.Errorf("failed to decode mtu: %w", err)
		}
		if parsedMtu, err := strconv.Atoi(string(mtubyte)); err == nil && parsedMtu > 0 {
			mtu = parsedMtu
		}
	}

	return UriData{
		Addr:  string(addr),
		Mode:  string(mode),
		Conn:  conn,
		Block: string(block),
		Key:   string(key),
		Mtu:   mtu,
	}, nil
}
