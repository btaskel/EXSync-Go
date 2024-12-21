package encrypt

import (
	"crypto/cipher"
	"crypto/rand"
	"errors"
	"fmt"
)

const (
	Aes128Gcm = "aes-128-gcm"
	Aes192Gcm = "aes-192-gcm"
	Aes256Gcm = "aes-256-gcm"

	Xchacha20IetfPoly1305 = "xchacha20-ietf-poly1305"
	Chacha20IetfPoly1305  = "chacha20-ietf-poly1305"
)

// CipherInfo 结构包含以下字段：
// keyLen: 密钥长度（以字节为单位）。
// ivLen: 初始化向量（IV）长度（以字节为单位）。
// lossLen: 加密损耗长度（以字节为单位）。
// newAEAD: 一个函数，用于根据给定的密钥创建新的 AEAD 实例。
type CipherInfo struct {
	keyLen  int
	ivLen   int
	lossLen int
	newAEAD func(key []byte) (*cipher.AEAD, error)
}

func (c *CipherInfo) GetIvLen() int {
	return c.ivLen
}
func (c *CipherInfo) GetKeyLen() int {
	return c.keyLen
}
func (c *CipherInfo) GetLossLen() int {
	return c.lossLen
}

// cipherMethod 是一个映射，将加密方法名称映射到相应的 CipherInfo 结构。
var cipherMethod = map[string]CipherInfo{
	Aes128Gcm:             {16, 12, 28, newAESGCM}, // iv: 12, tag: 16
	Aes192Gcm:             {24, 12, 28, newAESGCM},
	Aes256Gcm:             {32, 12, 28, newAESGCM},
	Chacha20IetfPoly1305:  {32, 12, 16, newChacha20IETFPoly1305}, // iv: 12, tag: 4
	Xchacha20IetfPoly1305: {32, 24, 16, newXChacha20IETFPoly1305X},
}

type Cipher struct {
	key  []byte
	enc  *cipher.AEAD
	dec  *cipher.AEAD
	Info CipherInfo
}

// NewCipher 创建加/解密器
func NewCipher(method, password string) (c *Cipher, err error) {
	if password == "" {
		return nil, errors.New("ErrEmptyPassword")
	}

	ci, ok := cipherMethod[method]
	if !ok {
		return nil, errors.New("Unsupported encryption method: " + method)
	}

	passwordHash, err := getPasswordHash(password, ci.keyLen)
	if err != nil {
		return nil, err
	}
	fmt.Println("pwHash:", passwordHash)
	c = &Cipher{
		key:  passwordHash,
		Info: ci,
	}

	err = c.InitEncrypt()
	if err != nil {
		return nil, err
	}
	err = c.InitDecrypt()
	if err != nil {
		return nil, err
	}

	return
}

func (c *Cipher) InitEncrypt() (err error) {
	if c.enc == nil {
		if c.Info.newAEAD == nil {
			return errors.New("newAEAD error")
		}
		c.enc, err = c.Info.newAEAD(c.key)
	}
	return
}

func (c *Cipher) InitDecrypt() (err error) {
	if c.dec == nil {
		if c.Info.newAEAD == nil {
			return errors.New("newAEAD error")
		}
		c.dec, err = c.Info.newAEAD(c.key)
	}
	return
}

// Encrypt 接受所有未加密数据，从加密损耗处开始将后续所有数据进行加密，同时会填补nonce与tag
func (c *Cipher) Encrypt(src, dst []byte) error {
	if _, err := rand.Read(dst[2 : c.Info.ivLen+2]); err != nil {
		return err
	}
	(*c.enc).Seal(dst[:c.Info.ivLen+2], dst[2:c.Info.ivLen+2], src[2+c.Info.lossLen:], nil)
	return nil
}

// Decrypt 接受所有已加密数据，会自动分割IV与TAG
func (c *Cipher) Decrypt(src, dst []byte) error {
	_, err := (*c.dec).Open(dst[:c.Info.lossLen+2], src[2:2+c.Info.ivLen], src[2+c.Info.ivLen:], nil)
	return err
}
