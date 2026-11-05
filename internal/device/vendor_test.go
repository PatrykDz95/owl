package device

import (
	"net"
	"strings"
	"testing"
)

func TestVendor(t *testing.T) {
	tests := []struct {
		mac  string
		want string
	}{
		{"00:00:0c:12:34:56", "Cisco"},
		{"02:11:22:33:44:55", "random"},
	}
	for _, tt := range tests {
		mac, _ := net.ParseMAC(tt.mac)
		got := Vendor(mac)
		if !strings.Contains(got, tt.want) {
			t.Errorf("Vendor(%s) = %q, want %q", tt.mac, got, tt.want)
		}
	}
}
