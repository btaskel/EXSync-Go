package protocol

//
//import (
//	"context"
//	"crypto/rand"
//	"crypto/rsa"
//	"crypto/tls"
//	"crypto/x509"
//	"encoding/pem"
//	"github.com/quic-go/quic-go"
//	"io"
//	"math/big"
//	"net"
//)
//
//func main() {
//	go func() { // 启动服务器
//		listener, err := quic.ListenAddr("localhost:4242", generateTLSConfig(), nil)
//		if err != nil {
//			panic(err)
//		}
//		sess, err := listener.Accept(context.Background())
//		if err != nil {
//			panic(err)
//		}
//		stream, err := sess.AcceptStream(context.Background())
//		if err != nil {
//			panic(err)
//		}
//		io.Copy()
//		// ...
//	}()
//
//	// 启动客户端
//	raddr, err := net.ResolveUDPAddr("udp", "localhost:4242")
//	if err != nil {
//		panic(err)
//	}
//	sess, err := quic.Dial(raddr, &tls.Config{InsecureSkipVerify: true}, nil)
//	if err != nil {
//		panic(err)
//	}
//	stream, err := sess.OpenStreamSync(context.Background())
//	if err != nil {
//		panic(err)
//	}
//	// ...
//}
//
//func generateTLSConfig() *tls.Config {
//	key, err := rsa.GenerateKey(rand.Reader, 2048)
//	if err != nil {
//		panic(err)
//	}
//	template := x509.Certificate{SerialNumber: big.NewInt(1)}
//	certDER, err := x509.CreateCertificate(rand.Reader, &template, &template, &key.PublicKey, key)
//	if err != nil {
//		panic(err)
//	}
//	keyPEM := pem.EncodeToMemory(&pem.Block{Type: "RSA PRIVATE KEY", Bytes: x509.MarshalPKCS1PrivateKey(key)})
//	certPEM := pem.EncodeToMemory(&pem.Block{Type: "CERTIFICATE", Bytes: certDER})
//
//	tlsCert, err := tls.X509KeyPair(certPEM, keyPEM)
//	if err != nil {
//		panic(err)
//	}
//	return &tls.Config{Certificates: []tls.Certificate{tlsCert}, NextProtos: []string{"quic-echo-example"}}
//}
