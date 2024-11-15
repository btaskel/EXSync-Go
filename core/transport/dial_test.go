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
	time.Sleep(5 * time.Second)
	fmt.Println("---end---")
}

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
		cipherMethod = encrypt.Aes192Gcm
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
	conn, err := DialAddr(context.Background(), network, "127.0.0.1:5002", &option)
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

func server(t *testing.T, network string, cipherFlag, compressorFlag bool, ConfTLS *tls.Config) {
	option := optionParse(cipherFlag, compressorFlag, ConfTLS)
	listener, err := Listen(network, "127.0.0.1:5002", &option)
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
	err = conn.Close()
	if err != nil {
		fmt.Println(err)
		return
	}
}
