package transport

import (
	"EXSync/core/internal/config"
	"EXSync/core/log"
	"EXSync/core/transport/compress"
	"EXSync/core/transport/encrypt"
	"context"
	"encoding/binary"
	"errors"
	"net"
	"time"
)

// TCPDial Client Dial
func TCPDial(address string, timeout time.Duration) (Conn, error) {
	conn, err := net.DialTimeout("tcp", address, timeout)
	if err != nil {
		return nil, err
	}
	return NewTCPWithCipher(conn)
}

// TCPListen Server listener
func TCPListen(address string) (listener Listener, err error) {
	l, err := net.Listen("tcp", address)
	if err != nil {
		return
	}
	return &TCPListener{Listener: l}, nil
}

type TCPListener struct {
	net.Listener
}

func (c *TCPListener) Accept(ctx context.Context) (Conn, error) {
	conn, err := c.Listener.Accept()
	if err != nil {
		return nil, err
	}
	return NewTCPWithCipher(conn)
}

type TCPWithCipher struct {
	net.Conn
	mc     *MapChannel
	cipher *encrypt.Cipher

	compressor     compress.Compress
	compressorLoss int

	err error
}

func NewTCPWithCipher(conn net.Conn) (tcpConn *TCPWithCipher, err error) {
	cip, err := encrypt.NewCipher(config.Config.Server.Setting.Encryption, config.Config.Server.Addr.Password)
	if err != nil {
		return
	}
	compressor, n, err := compress.NewCompress("", SocketSize-2-4)
	if err != nil {
		return nil, err
	}
	tcpConn = &TCPWithCipher{
		Conn:   conn,
		mc:     NewTimeChannel(),
		cipher: cip,

		compressor:     compressor,
		compressorLoss: n,
	}
	if err != nil {
		return
	}
	go tcpConn.tcpStreamRecv()
	return tcpConn, nil
}

func (c *TCPWithCipher) tcpStreamRecv() {
	defer func() {
		c.err = errors.New("tcpStreamRecv stopped")
	}()

	dataBuf := make([]byte, SocketSize)
	dataLenBuf := make([]byte, 2)
	var n int
	var dataLen uint16

	compressDstBuf := c.compressor.GetDstBuf()
	for {
		// 处理粘包
		_, c.err = c.Conn.Read(dataLenBuf)
		if c.err != nil {
			logger.Debug(c.err)
			return
		}
		dataLen = binary.BigEndian.Uint16(dataLenBuf)
		n, c.err = c.Conn.Read(dataBuf[:dataLen])
		if c.err != nil {
			return
		}

		// 解密内容
		if c.cipher != nil {
			c.err = c.cipher.Decrypt(dataBuf[:dataLen])
			if c.err != nil {
				continue
			}
		}

		// 解压缩
		if c.compressor != nil {
			n, c.err = c.compressor.UnCompressData(dataBuf[:dataLen], compressDstBuf)
			if c.err != nil {
				return
			}
			c.err = c.mc.Push(string(compressDstBuf[:n][:4]), compressDstBuf[4:n])
			if c.err != nil {
				logger.Errorf("tcpStreamRecv %s->%s: timeout!", c.RemoteAddr(), c.LocalAddr())
				return
			}
		} else {
			c.err = c.mc.Push(string(dataBuf[:dataLen][:4]), dataBuf[:dataLen][4:])
			if c.err != nil {
				logger.Errorf("tcpStreamRecv %s->%s: timeout!", c.RemoteAddr(), c.LocalAddr())
				return
			}
		}
	}
}

// OpenStream 创建多路复用数据流
func (c *TCPWithCipher) OpenStream(ctx context.Context) (Stream, error) {
	return newTCPStream(c)
}

// Close 释放TCPWithCipher
func (c *TCPWithCipher) Close() error {
	c.mc.Close()
	err := c.Conn.Close()
	if err != nil {
		return err
	}
	return nil
}
