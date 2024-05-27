package transport

import (
	"crypto/rand"
	"net"
)

var (
	SocketSize           = 4096
	streamIDSize         = 4                          // byte
	defaultControlStream = string([]byte{0, 0, 0, 1}) // create Stream and delete Stream

	StreamInUse = map[string]struct{}{
		defaultControlStream: {},
	} // use Stream id
)

// getTCPStreamID 获取一个未被占用的StreamID
func getTCPStreamID() (string, error) {
	slice := make([]byte, streamIDSize)
	_, err := rand.Read(slice)
	if err != nil {
		return "", err
	}
	id := string(slice)
	if _, ok := StreamInUse[id]; ok {
		id, err = getTCPStreamID()
		if err != nil {
			return "", err
		}
		return id, nil
	}
	return id, nil
}

type TCPStream struct {
	*TCPWithCipher
	id        string
	idByte    []byte
	rn, wn    int
	DataStart int
}

func pickStaticStream() {

}

// newTCPStream with Cipher
func newTCPStream(twc *TCPWithCipher) (Stream, error) {
	tcpStreamID, err := getTCPStreamID()
	if err != nil {
		return nil, err
	}

	err = twc.mc.CreateRecv(tcpStreamID)
	if err != nil {
		return nil, err
	}

	return &TCPStream{TCPWithCipher: twc,
		DataStart: twc.cipher.Info.GetIvLen() + streamIDSize - 1 + 2, // +2 DataLen
		id:        tcpStreamID,
		idByte:    []byte(tcpStreamID),
	}, nil
}

// Read Stream With Cipher
// dataLen -> cipherText -> uncompressedText -> multiplexing -> plainText
func (c *TCPStream) Read(b []byte) (int, error) {
	c.rn, c.err = c.mc.Pull(c.id, &b)
	if c.err != nil {
		return 0, c.err
	}
	return 0, nil
}

// Write Stream With Cipher
func (c *TCPStream) Write(b []byte) (int, error) {
	// 增加长度首部
	b[0] = byte((len(b) >> 8) & 255)
	b[1] = byte(len(b) & 255)

	// 增加流id
	copy(b[2:streamIDSize-1], c.idByte)
	if c.cipher.Info.KeyLen == 0 {
		c.wn, c.err = c.Conn.Write(b)
	}

	// 进行加密
	c.err = c.cipher.Encrypt(b)
	if c.err != nil {
		return 0, c.err
	}

	c.wn, c.err = c.Conn.Write(b)
	if c.err != nil {
		return c.wn, c.err
	}
	return c.wn, c.err
}

func (c *TCPStream) RemoteAddr() net.Addr {
	return c.Conn.RemoteAddr()
}

func (c *TCPStream) LocalAddr() net.Addr {
	return c.Conn.LocalAddr()
}

// Close 关闭一个多路复用数据流
func (c *TCPStream) Close() error {
	c.mc.Del(c.id)
	delete(StreamInUse, c.id)
	return nil
}

func (c *TCPStream) GetIVLen() int {
	return c.cipher.Info.GetIvLen()
}

func (c *TCPStream) GetKeyLen() int {
	return c.cipher.Info.GetKeyLen()
}

// NewDataSlice 创建一个数据收发切片，并返回数据起始位置
func (c *TCPStream) NewDataSlice() ([]byte, int) {
	slice := make([]byte, SocketSize)
	return slice, c.cipher.Info.GetIvLen() + 2 + 4 + c.TCPWithCipher.compressorLoss - 1
}
