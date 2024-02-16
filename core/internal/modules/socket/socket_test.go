package socket

import (
	"EXSync/core/internal/modules/timechannel"
	"EXSync/core/option/exsync/comm"
	"fmt"
	"net"
	"sync"
	"testing"
	"time"
)

func TestNewSession(t *testing.T) {
	tc := timechannel.NewTimeChannel()
	defer tc.Close()

	server := func() {
		conn, err := net.Listen("tcp", "127.0.0.1:5000")
		if err != nil {
			fmt.Println("s1:", err)
		}
		defer conn.Close()
		accept, err := conn.Accept()
		if err != nil {
			fmt.Println("s2:", err)
			return
		}
		session, err := NewSession(tc, accept, nil, "12345678", nil)
		defer session.Close()
		if err != nil {
			fmt.Println("s3:", err)
			return
		}
		fmt.Println("开始发送")
		err = session.SendDataP([]byte("测试的文字"))
		if err != nil {
			fmt.Println("s4:", err)
			return
		}
	}
	client := func() {
		conn, err := net.DialTimeout("tcp", "127.0.0.1:5000", 3*time.Second)
		if err != nil {
			fmt.Println("c1:", err)
		}
		defer conn.Close()
		//session, err := NewSession(tc, conn, nil, "12345678", nil)
		err = tc.CreateRecv("12345678")
		if err != nil {
			fmt.Println("c2:", err)
			return
		}
		buf := make([]byte, 4096)
		n, err := conn.Read(buf)
		if err != nil {
			return
		}
		tc.Set("12345678", buf[:n])

		result, err := tc.GetTimeout("12345678", 2)
		if err != nil {
			fmt.Println("c3:", err)
			return
		}
		fmt.Println(string(result))
	}
	wait := sync.WaitGroup{}
	wait.Add(2)
	go func() {
		server()
		wait.Done()
	}()
	time.Sleep(1 * time.Second)
	go func() {
		client()
		wait.Done()
	}()
	wait.Wait()

}

func TestSession_SendCommand(t *testing.T) {

	server := func() {
		tc := timechannel.NewTimeChannel()
		defer tc.Close()
		conn, err := net.Listen("tcp", "127.0.0.1:5000")
		if err != nil {
			fmt.Println("s1:", err)
		}
		defer conn.Close()
		accept, err := conn.Accept()
		if err != nil {
			fmt.Println("s2:", err)
			return
		}
		session, err := NewSession(tc, accept, nil, "12345678", nil)
		defer session.Close()
		if err != nil {
			fmt.Println("s3:", err)
			return
		}
		fmt.Println("开始发送")
		command := comm.Command{
			Command: "abc",
			Type:    "edf",
			Method:  "123",
			Data: map[string]any{
				"flag": true,
			},
		}
		result, err := session.SendCommand(command, true, true)
		if err != nil {
			fmt.Println("s4:", err)
			return
		}
		data, ok := result["data"].(map[string]any)
		if !ok {
			fmt.Println("result[\"data\"].(map[string]any)转换失败了")
		}
		flag, ok := data["flag"].(bool)
		if !ok {
			fmt.Println("data[\"flag\"].(bool)转换失败了")
		}
		fmt.Println("client答复标识：", flag)
	}
	client := func() {
		tc := timechannel.NewTimeChannel()
		defer tc.Close()
		conn, err := net.DialTimeout("tcp", "127.0.0.1:5000", 3*time.Second)
		if err != nil {
			fmt.Println("c1:", err)
		}
		defer conn.Close()
		session, err := NewSession(tc, conn, nil, "12345678", nil)
		result, ok := session.Recv()
		if !ok {
			fmt.Println("c2:", err)
		}
		flag, ok := result.Data["flag"].(bool)
		fmt.Println("server答复标识: ", flag) //
		command := comm.Command{
			Command: "abc",
			Type:    "edf",
			Method:  "123",
			Data: map[string]any{
				"flag": true,
			},
		}
		_, err = session.SendCommand(command, true, true)
		if err != nil {
			fmt.Println("c3:", err)
			return
		}
	}
	wait := sync.WaitGroup{}
	wait.Add(2)
	go func() {
		server()
		wait.Done()
	}()
	time.Sleep(1 * time.Second)
	go func() {
		client()
		wait.Done()
	}()
	wait.Wait()
}
