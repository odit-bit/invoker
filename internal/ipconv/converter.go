package ipconv

import (
	"fmt"
	"net"
)

// possible IPv4 addresses (0.0.0.0 to 255.255.255.255).

func Uint64(ipAddr string) (uint64, error) {
	ip := net.ParseIP(ipAddr)
	if ip == nil {
		return 0, fmt.Errorf("Invalid IP address: %s", ipAddr)
	}
	// Convert the IP address to a 32-bit unsigned integer (big.Int)
	ipBytes := ip.To4()
	ipNumber := uint64(ipBytes[0])<<24 | uint64(ipBytes[1])<<16 | uint64(ipBytes[2])<<8 | uint64(ipBytes[3])
	return ipNumber, nil
}

func String(ipNum uint64) (string, error) {
	ipBytes := make(net.IP, 4)
	ipBytes[0] = byte(ipNum >> 24)
	ipBytes[1] = byte(ipNum >> 16)
	ipBytes[2] = byte(ipNum >> 8)
	ipBytes[3] = byte(ipNum)

	return ipBytes.String(), nil
}
