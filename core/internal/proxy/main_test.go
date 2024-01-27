package proxy

import (
	"fmt"
	"testing"
)

func TestSetProxy(t *testing.T) {
	socks := setProxy()
	_, err := socks.Dial("tcp", "127.0.0.1:5500")
	if err != nil {
		fmt.Println(err)
		return
	}
	fmt.Println("over")
}
