package proxy

import (
	"EXSync/core/internal/config"
	loger "EXSync/core/log"
	"fmt"
	"golang.org/x/net/proxy"
	"net"
	"os"
	"time"
)

var Socks5 = setProxy()

func setProxy() proxy.Dialer {
	addr := fmt.Sprintf("%s:%d", config.Config.Server.Proxy.Hostname, config.Config.Server.Proxy.Port)
	dialer := &net.Dialer{Timeout: 5 * time.Second}
	socks5, err := proxy.SOCKS5("tcp", addr, nil, dialer)
	if err != nil {
		loger.Log.Fatalf("setProxy: Proxy server settings error! %s", addr)
		os.Exit(1)
	}
	return socks5
}
