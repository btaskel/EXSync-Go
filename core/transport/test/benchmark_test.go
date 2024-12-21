package test

import (
	"EXSync/core/transport"
	"context"
	"crypto/rand"
	"crypto/rsa"
	"crypto/tls"
	"crypto/x509"
	"crypto/x509/pkix"
	"encoding/pem"
	"fmt"
	"github.com/quic-go/quic-go"
	"math/big"
	"testing"
	"time"
)

func BenchmarkTCPCipherRW(b *testing.B) {
	b.N = 1
	go func() {
		listener, err := transport.Listen("tcp", "127.0.0.1:10000", &transport.ConfOption{
			ConfTLS:          nil,
			ConfQUIC:         nil,
			AEADMethod:       "aes-128-gcm",
			AEADPassword:     "123456",
			CompressorMethod: "lz4",
		})
		if err != nil {
			panic(err)
		}
		var conn quic.Connection
		conn, err = listener.Accept(context.Background())
		if err != nil {
			return
		}
		stream, err := conn.AcceptStream(context.Background())
		if err != nil {
			return
		}
		buf := make([]byte, 4096)
		var n int
		var counter int
		for i := 0; i < b.N; i++ {
			counter++
			fmt.Println("read")
			n, err = stream.Read(buf[:n])
			if err != nil {
				panic(err)
			}
		}
		b.Log("server finish", counter)
	}()
	conn, err := transport.DialAddr(context.Background(), "tcp", "127.0.0.1:10000", &transport.ConfOption{
		ConfTLS:          nil,
		ConfQUIC:         nil,
		AEADMethod:       "aes-128-gcm",
		AEADPassword:     "123456",
		CompressorMethod: "lz4",
	})
	if err != nil {
		panic(err)
	}

	stream, err := conn.OpenStream()
	if err != nil {
		panic(err)
	}
	//buf := make([]byte, 4096)
	//_, err = rand.Read(buf)
	//if err != nil {
	//	panic(err)
	//}
	dst := []byte("测试")
	counter := 0
	for i := 0; i < b.N; i++ {
		counter++
		_, err = stream.Write(dst)
		if err != nil {
			panic(err)
		}
	}
	b.Log("client finish")
	fmt.Println("client finish", counter)
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
