package util

import (
	"errors"
	"fmt"
	"net"
)

func SetReceiveBuffer(c net.PacketConn, receiveBufferSize int) error {
	conn, ok := c.(interface{ SetReadBuffer(int) error })
	if !ok {
		return errors.New("connection doesn't allow setting of receive buffer size. Not a *net.UDPConn")
	}
	if err := conn.SetReadBuffer(receiveBufferSize); err != nil {
		return fmt.Errorf("failed to increase receive buffer size: %w", err)
	}
	return nil
}

// GetInterfaceFromIPorName returns the network interface based on the provided IP address or name.
// If the input is an IP address, it searches for an interface with that IP.
// If the input is a name, it attempts to find the interface by name.
func GetInterfaceFromIPorName(input string) (*net.Interface, error) {
	// Check if input is an IP address
	if ip := net.ParseIP(input); ip != nil {
		interfaces, err := net.Interfaces()
		if err != nil {
			return nil, fmt.Errorf("failed to get network interfaces: %w", err)
		}

		for _, iface := range interfaces {
			addrs, err := iface.Addrs()
			if err != nil {
				continue // Skip interfaces where addresses can't be obtained
			}

			for _, addr := range addrs {
				var ipNet *net.IPNet
				var ok bool
				if ipNet, ok = addr.(*net.IPNet); ok && ipNet.IP.Equal(ip) {
					return &iface, nil
				}
			}
		}

		return nil, fmt.Errorf("no interface found with IP %s", input)
	}

	// If input is not an IP, assume it's an interface name
	iface, err := net.InterfaceByName(input)
	if err != nil {
		return nil, fmt.Errorf("interface %s not found: %w", input, err)
	}

	return iface, nil
}
