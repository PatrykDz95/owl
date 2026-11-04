package device

import (
	"bufio"
	"bytes"
	"errors"
	"net"
	"os/exec"
	"strings"
)

func LookupMAC(ip net.IP) (net.HardwareAddr, error) {
	out, err := exec.Command("arp", "-a").Output()
	if err != nil {
		return nil, err
	}

	scanner := bufio.NewScanner(bytes.NewReader(out))
	for scanner.Scan() {
		fields := strings.Fields(scanner.Text())
		if len(fields) >= 2 && net.ParseIP(fields[0]).Equal(ip) {
			return net.ParseMAC(fields[1])
		}
	}
	return nil, errors.New("not found in ARP table")
}
