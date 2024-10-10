package encrypt

import (
	"crypto/aes"
	"crypto/cipher"
	"golang.org/x/crypto/chacha20poly1305"
)

func newAESGCM(key []byte) (*cipher.AEAD, error) {
	c, err := aes.NewCipher(key)
	if err != nil {
		return nil, err
	}
	var aead cipher.AEAD
	aead, err = cipher.NewGCM(c)
	if err != nil {
		return nil, err
	}
	return &aead, nil
}

func newChacha20IETFPoly1305(key []byte) (*cipher.AEAD, error) {
	aead, err := chacha20poly1305.New(key)
	if err != nil {
		return nil, err
	}
	return &aead, nil
}

func newXChacha20IETFPoly1305X(key []byte) (*cipher.AEAD, error) {
	aead, err := chacha20poly1305.NewX(key)
	if err != nil {
		return nil, err
	}
	return &aead, nil
}
