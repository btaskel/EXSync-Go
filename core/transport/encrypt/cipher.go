package encrypt

import (
	"crypto/cipher"
	"crypto/rand"
	"errors"
)

const (
	Null      = ""
	Aes128Gcm = "aes-128-gcm"
	Aes192Gcm = "aes-192-gcm"
	Aes256Gcm = "aes-256-gcm"

	Xchacha20IetfPoly1305 = "xchacha20-ietf-poly1305"
	Chacha20IetfPoly1305  = "chacha20-ietf-poly1305"
)

type cipherInfo struct {
	KeyLen  int
	ivLen   int
	lossLen int
	newAEAD func(key []byte) (*cipher.AEAD, error)
}

func (c *cipherInfo) GetIvLen() int {
	return c.ivLen
}
func (c *cipherInfo) GetKeyLen() int {
	return c.KeyLen
}

var cipherMethod = map[string]*cipherInfo{
	Null:                  {0, 0, 0, nil},
	Aes128Gcm:             {16, 12, 28, newAESGCM}, // iv: 12, tag: 16
	Aes192Gcm:             {24, 16, 28, newAESGCM},
	Aes256Gcm:             {32, 16, 28, newAESGCM},
	Chacha20IetfPoly1305:  {32, 12, 16, newChacha20IETFPoly1305}, // iv: 12, tag: 4
	Xchacha20IetfPoly1305: {32, 24, 16, newXChacha20IETFPoly1305X},
}

type Cipher struct {
	key  []byte
	enc  *cipher.AEAD
	dec  *cipher.AEAD
	Info *cipherInfo
}

var (
	ErrEmptyPassword = errors.New("ErrEmptyPassword")
)

// NewCipher 创建加/解密器
func NewCipher(method, password string) (c *Cipher, err error) {
	if password == "" {
		return nil, ErrEmptyPassword
	}

	ci, ok := cipherMethod[method]
	if !ok {
		return nil, errors.New("Unsupported encryption method: " + method)
	}

	passwordHash, err := getPasswordHash(password, ci.KeyLen)
	if err != nil {
		return nil, err
	}
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
			return
		}
		c.enc, err = c.Info.newAEAD(c.key)
	}
	return
}

func (c *Cipher) InitDecrypt() (err error) {
	if c.dec == nil {
		if c.Info.newAEAD == nil {
			return
		}
		c.dec, err = c.Info.newAEAD(c.key)
	}
	return
}

// Encrypt 传入明文，并将密文传回dst。将nonce写入到前N个字节中去
func (c *Cipher) Encrypt(plaintext []byte) error {
	_, err := rand.Read(plaintext[2 : c.Info.ivLen+1])
	if err != nil {
		return err
	}
	(*c.enc).Seal(plaintext, plaintext[2:c.Info.ivLen+1], plaintext[c.Info.ivLen+3:], nil)
	return nil
}

// Decrypt 传入密文，并明文传回dst
func (c *Cipher) Decrypt(ciphertext []byte) error {
	_, err := (*c.dec).Open(nil, ciphertext[2:(*c.dec).NonceSize()], ciphertext, nil)
	if err != nil {
		return err
	}
	return nil
}
