package network

import (
	"bufio"
	"fmt"
	"os"
	"strings"
)

type Arp struct {
	Interface string
	IpAddress string
	MacAdress string
}

func GetArp(ipAddress string, iface string) (Arp, error) {

	arpfile, err := os.Open("/proc/net/arp")
	if err != nil {
		return Arp{}, err
	}
	scanner := bufio.NewScanner(arpfile)

	//skip the first line
	scanner.Scan()
	for scanner.Scan() {
		fields := strings.Fields(scanner.Text())
		if len(fields) < 3 {
			continue
		}
		ipaddr := fields[0]
		if ipaddr == ipAddress && fields[5] == iface {
			macaddr := fields[3]
			iface := fields[5]
			return Arp{Interface: iface, IpAddress: ipaddr, MacAdress: macaddr}, nil
		}
	}
	if scanner.Err() != nil {
		return Arp{}, scanner.Err()
	}
	return Arp{}, fmt.Errorf("Arp not found")
}
