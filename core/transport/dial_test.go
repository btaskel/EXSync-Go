package transport

import (
	"EXSync/core/transport/encrypt"
	"context"
	"crypto/rand"
	"crypto/rsa"
	"crypto/tls"
	"crypto/x509"
	"crypto/x509/pkix"
	"encoding/pem"
	"fmt"
	"math/big"
	"testing"
	"time"
)

// TestTCPWithCipher pass
func TestTCPWithCipher(t *testing.T) {
	go server(t, "tcp", true, true, nil)
	time.Sleep(1 * time.Second)
	client(t, "tcp", true, true, nil)
	time.Sleep(4 * time.Second)
	fmt.Println("---end---")
}

// generateSelfSignedCertificate 生成一个自签名证书。
func generateSelfSignedCertificate() (tls.Certificate, error) {
	priv, err := rsa.GenerateKey(rand.Reader, 2048)
	if err != nil {
		return tls.Certificate{}, err
	}

	template := x509.Certificate{
		SerialNumber: big.NewInt(1),
		Subject: pkix.Name{
			Organization: []string{"Acme Co"},
		},
		NotBefore: time.Now(),
		NotAfter:  time.Now().Add(time.Hour * 24 * 180),

		KeyUsage:              x509.KeyUsageKeyEncipherment | x509.KeyUsageDigitalSignature,
		ExtKeyUsage:           []x509.ExtKeyUsage{x509.ExtKeyUsageServerAuth},
		BasicConstraintsValid: true,
	}

	derBytes, err := x509.CreateCertificate(rand.Reader, &template, &template, &priv.PublicKey, priv)
	if err != nil {
		return tls.Certificate{}, err
	}

	certPEM := pem.EncodeToMemory(&pem.Block{Type: "CERTIFICATE", Bytes: derBytes})
	privPEM := pem.EncodeToMemory(&pem.Block{Type: "RSA PRIVATE KEY", Bytes: x509.MarshalPKCS1PrivateKey(priv)})

	cert, err := tls.X509KeyPair(certPEM, privPEM)
	if err != nil {
		return tls.Certificate{}, err
	}

	return cert, nil
}

// TestQUIC pass
func TestQUIC(t *testing.T) {

	cert, err := generateSelfSignedCertificate()
	if err != nil {
		fmt.Println(err)
		return
	}
	go server(t, "quic", false, true, &tls.Config{Certificates: []tls.Certificate{cert}})
	time.Sleep(1 * time.Second)
	client(t, "quic", false, true, &tls.Config{InsecureSkipVerify: true})
	time.Sleep(5 * time.Second)
	fmt.Println("---end---")
}

func TestTCPWithTLS(t *testing.T) {
	cert, err := generateSelfSignedCertificate()
	if err != nil {
		fmt.Println(err)
		return
	}
	go server(t, "tcp", false, true, &tls.Config{Certificates: []tls.Certificate{cert}})
	time.Sleep(1 * time.Second)
	client(t, "tcp", false, true, &tls.Config{InsecureSkipVerify: true})
	time.Sleep(5 * time.Second)
	fmt.Println("---end---")
}

func optionParse(cipherFlag, compressorFlag bool, ConfTLS *tls.Config) ConfOption {
	var compressorMethod string
	if compressorFlag {
		compressorMethod = "lz4"
	}
	var cipherMethod string
	if cipherFlag {
		cipherMethod = encrypt.Aes256Gcm
	}
	return ConfOption{
		ConfTLS:          ConfTLS,
		ConfQUIC:         nil,
		AEADMethod:       cipherMethod,
		AEADPassword:     "123456",
		CompressorMethod: compressorMethod,
	}
}

func client(t *testing.T, network string, cipherFlag, compressorFlag bool, ConfTLS *tls.Config) {
	option := optionParse(cipherFlag, compressorFlag, ConfTLS)
	fmt.Println("test: ", option.ConfTLS, option.AEADMethod, option.CompressorMethod, option.AEADPassword)
	conn, err := DialAddr(context.Background(), network, "127.0.0.1:5003", &option)
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
	buf := make([]byte, 4096)
	buf[0] = 50
	buf[1] = 51
	buf[2] = 52
	buf[3] = 53
	n, err := stream.Write(buf[:4])
	if err != nil {
		panic(err)
	}
	fmt.Println("client write:", n)
	conn.CloseWithError(400, "123456")
}

// server 启动一个监听指定网络和地址的服务器。
// 它接受传入的连接并从连接的第一个流中读取数据。
// network: 网络类型（例如，“tcp”，“quic”）。
// cipherFlag: 一个布尔值，指示是否使用加密。
// compressorFlag: 一个布尔值，指示是否使用压缩。
// ConfTLS: 用于安全连接的 TLS 配置。
func server(t *testing.T, network string, cipherFlag, compressorFlag bool, ConfTLS *tls.Config) {
	option := optionParse(cipherFlag, compressorFlag, ConfTLS)
	listener, err := Listen(network, "127.0.0.1:5003", &option)
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
	//conn.CloseWithError(0, "")
}
