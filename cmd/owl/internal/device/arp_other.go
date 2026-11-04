//go:build !windows

package device

import (
	"errors"
	"net"
)

func LookupMAC(ip net.IP) (net.HardwareAddr, error) {
	return nil, errors.New("LookupMAC: not implemented on this OS")
}
