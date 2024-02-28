package lan

import (
	"testing"
)

func TestScanDevices(t *testing.T) {
	_, err := ScanDevices()
	if err != nil {
		return
	}
	//fmt.Println(devices)
	//fmt.Println(IphalfStr)
	//err := getLocalAddr()
	//if err != nil {
	//	return
	//}
	//fmt.Println("______")
	//for _, addr := range IphalfStr {
	//	fmt.Println(addr)
	//}
	//	169.254.44
	//192.168.1
	//10.221.20
	//169.254.212
	//169.254.218
	//169.254.45
	//192.168.1
	//169.254.146

	//169.254
	//192.168.1
	//10
	//169.254
	//169.254
	//169.254
	//192.168.1
	//169.254
}
