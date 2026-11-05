package device

import (
	"bytes"
	_ "embed"
	"encoding/csv"
	"fmt"
	"log"
	"net"
	"sync"
)

//go:embed oui.csv
var ouiCSV []byte

var (
	vendors     map[string]string
	vendorsOnce sync.Once
)

func loadVendors() {
	vendors = make(map[string]string)
	r := csv.NewReader(bytes.NewReader(ouiCSV))
	r.FieldsPerRecord = -1
	records, err := r.ReadAll()
	if err != nil {
		log.Println("loading OUI database:", err)
		return
	}
	for _, rec := range records {
		if len(rec) < 3 {
			continue
		}
		vendors[rec[1]] = rec[2]
	}
}

func Vendor(mac net.HardwareAddr) string {
	if len(mac) < 3 {
		return "unknown"
	}
	if mac[0]&0x02 != 0 {
		return "random"
	}
	vendorsOnce.Do(loadVendors)
	key := fmt.Sprintf("%02X%02X%02X", mac[0], mac[1], mac[2])
	if vendor, exists := vendors[key]; exists {
		return vendor
	}
	return "unknown"
}
