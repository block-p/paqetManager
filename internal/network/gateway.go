package network

import (
	"bufio"
	"fmt"
	"os"
	"strings"
)

const (
	RTF_UP      = 0x0001 // Route is up
	RTF_GATEWAY = 0x0002 // Route uses a gateway
)

type Gateway struct {
	Interface string
	LocalIp   string
}

func GetDefaultGateway() (Gateway, error) {
	routefile, err := os.Open("/proc/net/route")
	if err != nil {
		return Gateway{}, err
	}
	scanner := bufio.NewScanner(routefile)
	//skip the first line
	scanner.Scan()
	for scanner.Scan() {
		gw := strings.Fields(scanner.Text())
		if len(gw) < 3 {
			continue
		}
		iface := gw[0]
		dst := gw[1]
		gateway := gw[2]

		if dst == "00000000" && gateway != "00000000" {
			return Gateway{
				Interface: iface,
				LocalIp:   gateway,
			}, nil
		}
	}
	if scanner.Err() != nil {
		return Gateway{}, scanner.Err()
	}
	return Gateway{}, fmt.Errorf("Network interface not found")

}
