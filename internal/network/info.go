package network

import (
	"encoding/hex"
	"fmt"
	"net"
	"time"
)

type Info struct {
	Interface string
	IpAddr    string
	MacAddr   string
}

func GetFullInfo() (Info, error) {
	gw, err := GetDefaultGateway()
	if err != nil {
		return Info{}, err
	}
	ip, err := HexToIP(gw.LocalIp)
	if err != nil {
		return Info{}, err
	}
	err = WakeUpArp(ip.String())
	if err != nil {
		return Info{}, err
	}
	arp, err := GetArp(ip.String(), gw.Interface)
	if err != nil {
		return Info{}, err
	}

	localip, err := GetIp()
	if err != nil {
		return Info{}, err
	}
	return Info{
		Interface: arp.Interface,
		IpAddr:    localip,
		MacAddr:   arp.MacAdress,
	}, nil
}

func HexToIP(hexStr string) (net.IP, error) {
	// 1. Decode the hex string into 4 bytes
	b, err := hex.DecodeString(hexStr)
	if err != nil {
		return nil, err
	}
	if len(b) != 4 {
		return nil, fmt.Errorf("invalid IPv4 hex length: %s", hexStr)
	}
	// 2. Reverse order (b[3], b[2], b[1], b[0]) to get the correct IPv4
	ip := net.IPv4(b[3], b[2], b[1], b[0])
	return ip, nil
}

func GetIp() (string, error) {
	conn, err := net.Dial("udp", "8.8.8.8:80")
	if err != nil {
		return "", err
	}
	defer conn.Close()

	addr := conn.LocalAddr().(*net.UDPAddr)
	return addr.IP.String(), nil
}

func WakeUpArp(gwip string) error {
	conn, err := net.DialTimeout("udp", gwip+":12345", 200*time.Millisecond)
	if err != nil {
		return err
	}
	defer conn.Close()
	_, err = conn.Write([]byte{0})
	if err != nil {
		return err
	}
	time.Sleep(100 * time.Millisecond)
	return nil
}
