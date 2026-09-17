package netinfo

import (
	"fmt"
	"net"
	"os"
	"path/filepath"
	"strings"
)

type Interface struct {
	Name      string   `json:"name"`
	Status    string   `json:"status"`
	IP4       []string `json:"ip4"`
	IP6       []string `json:"ip6"`
	MAC       string   `json:"mac"`
	Speed     string   `json:"speed"`
	RXBytes   uint64   `json:"rx_bytes"`
	TXBytes   uint64   `json:"tx_bytes"`
	RXPackets uint64   `json:"rx_packets"`
	TXPackets uint64   `json:"tx_packets"`
}

func List() ([]Interface, error) {
	ifaces, err := net.Interfaces()
	if err != nil {
		return nil, err
	}

	var result []Interface
	for _, iface := range ifaces {
		info := Interface{
			Name:   iface.Name,
			Status: ifaceStatus(iface.Flags),
			MAC:    iface.HardwareAddr.String(),
		}

		addrs, err := iface.Addrs()
		if err == nil {
			for _, addr := range addrs {
				ip, _, _ := net.ParseCIDR(addr.String())
				if ip == nil {
					continue
				}
				if ip.To4() != nil {
					info.IP4 = append(info.IP4, ip.String())
				} else {
					info.IP6 = append(info.IP6, ip.String())
				}
			}
		}

		statsPath := fmt.Sprintf("/sys/class/net/%s/statistics", iface.Name)
		if data, err := os.ReadFile(filepath.Join(statsPath, "rx_bytes")); err == nil {
			fmt.Sscanf(strings.TrimSpace(string(data)), "%d", &info.RXBytes)
		}
		if data, err := os.ReadFile(filepath.Join(statsPath, "tx_bytes")); err == nil {
			fmt.Sscanf(strings.TrimSpace(string(data)), "%d", &info.TXBytes)
		}
		if data, err := os.ReadFile(filepath.Join(statsPath, "rx_packets")); err == nil {
			fmt.Sscanf(strings.TrimSpace(string(data)), "%d", &info.RXPackets)
		}
		if data, err := os.ReadFile(filepath.Join(statsPath, "tx_packets")); err == nil {
			fmt.Sscanf(strings.TrimSpace(string(data)), "%d", &info.TXPackets)
		}

		speedPath := fmt.Sprintf("/sys/class/net/%s/speed", iface.Name)
		if data, err := os.ReadFile(speedPath); err == nil {
			info.Speed = strings.TrimSpace(string(data)) + " Mbps"
		}

		result = append(result, info)
	}
	return result, nil
}

func ifaceStatus(flags net.Flags) string {
	if flags&net.FlagUp != 0 {
		return "up"
	}
	return "down"
}
