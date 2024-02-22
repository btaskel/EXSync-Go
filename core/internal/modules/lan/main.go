package lan

import (
	"errors"
	"fmt"
	"net"
	"strings"
	"strconv"
)

var IphalfStr []string

func ScanDevices() ([]string, error) {
	err := getLocalAddr()
	if err != nil {
		return nil, err
	}
	var onlineHosts []string
	for _, ch := range IphalfStr {
		onlineHosts = append(onlineHosts, Task(ch)...)
	}
	return onlineHosts, nil
}

func getLocalAddr() error {
	addrs, err := net.InterfaceAddrs()

	if err != nil {
		return err
	}
	for _, address := range addrs {
		// 检查ip地址判断是否回环地址
		if ipnet, ok := address.(*net.IPNet); ok && !ipnet.IP.IsLoopback() {
			if ipnet.IP.To4() != nil {
				a := strings.Split(ipnet.IP.String(), ".")
				if a[3] == "1" {
					continue
				}
				err = AddressProcessing(ipnet.IP.String())
				if err != nil {
					return err
				}
			}
		}
	}
	return nil
}

func AddressProcessing(ip string) error {
	a := strings.Split(ip, ".")
	if len(a) != 4 {
		//debugLog.Print("检查地址错误！")
		return errors.New("检查地址错误！")
	}
	num, err := strconv.Atoi(a[0])
	if err != nil {
		return err
	}

	host := fmt.Sprintf("%s.%s.%s", a[0], a[1], a[2])
	if num >= 128 && num < 192 {
		host = fmt.Sprintf("%s.%s", a[0], a[1])
	} else if num < 128 {
		host = fmt.Sprintf("%s", a[0])
	}
	//debugLog.Print("获取本机地址:", ip, "------->  处理为网段:", host)
	IphalfStr = append(IphalfStr, host)
	return nil
}
