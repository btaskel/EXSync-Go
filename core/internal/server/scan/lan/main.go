package lan

import (
	"errors"
	"fmt"
	"net"
	"runtime"
	"strings"
)

var IphalfStr []string

func ScanDevices() ([]string, error) {
	runtime.GOMAXPROCS(4)
	err := Getlocaladdr()
	if err != nil {
		return nil, err
	}
	var onlineHosts []string
	for _, ch := range IphalfStr {
		onlineHosts = append(onlineHosts, Task(ch)...)
	}
	return onlineHosts, nil
}

//func loginit() {
//	logFile, err := os.Create("log.txt")
//	if err != nil {
//		log.Fatalln("open file error !")
//	}
//	debugLog = log.New(logFile, "[Debug]", log.Llongfile)
//}

func Getlocaladdr() error {
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
	host := fmt.Sprintf("%s.%s.%s", a[0], a[1], a[2])
	//debugLog.Print("获取本机地址:", ip, "------->  处理为网段:", host)
	IphalfStr = append(IphalfStr, host)
	return nil
}
