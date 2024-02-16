package lan

import (
	"fmt"
	"testing"
)

func TestScanDevices(t *testing.T) {
	devices, err := ScanDevices()
	if err != nil {
		return
	}
	fmt.Println(devices)
}
