package transport

import (
	"EXSync/core/transport/encrypt"
	"context"
	"fmt"
	"testing"
	"time"
)

func TestListen(t *testing.T) {
	go server(t)
	time.Sleep(1 * time.Second)
	client(t)
	fmt.Println("---end---")
	time.Sleep(5 * time.Second)
}

func client(t *testing.T) {
	conn, err := Dial(context.Background(), "tcp", "127.0.0.1:5002", ConfOption{
		ConfTLS:          nil,
		ConfQUIC:         nil,
		AEADMethod:       encrypt.Aes192Gcm,
		AEADPassword:     "123456",
		CompressorMethod: "lz4",
	})
	if err != nil {
		t.Fatal(err)
		return
	}
	stream, err := conn.OpenStream()
	if err != nil {
		t.Fatal("OpenStream: ", err)
		return
	}
	//buf := []byte{50, 51, 52, 53}
	buf, n := stream.GetBuf()
	fmt.Println("write-start-n:", n)
	buf[n] = 50
	buf[n+1] = 51
	buf[n+2] = 52
	buf[n+3] = 53
	n, err = stream.Write(buf[:n+4])
	if err != nil {
		t.Fatal(err)
		return
	}
	fmt.Println("client write:", n)
}

func server(t *testing.T) {
	listener, err := Listen("tcp", "127.0.0.1:5002", ConfOption{
		ConfTLS:          nil,
		ConfQUIC:         nil,
		AEADMethod:       encrypt.Aes192Gcm,
		AEADPassword:     "123456",
		CompressorMethod: "lz4",
	})
	if err != nil {
		t.Fatal(err)
		return
	}

	conn, err := listener.Accept(context.Background())
	if err != nil {
		t.Fatal(err)
		return
	}
	fmt.Println("Accepted")

	stream, err := conn.AcceptStream(context.Background())
	if err != nil {
		t.Fatal(err)
		return
	}
	fmt.Println("AcceptStream")

	buf := make([]byte, 4096)
	fmt.Println("stream-pointer", stream == nil)
	n, err := stream.Read(buf)
	if err != nil {
		t.Fatal(err)
		return
	}
	fmt.Println("result : ", buf[:n])
	conn.Close()
}
