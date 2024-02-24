package encryption

import (
	"EXSync/core/internal/modules/hashext"
	loger "EXSync/core/log"
	"crypto/aes"
	"crypto/cipher"
	"crypto/rand"
	"encoding/base64"
	"errors"
	"fmt"
	"io"
)

const dataLength = 4096 // 加密后允许的最大长度

func NewGCM(key string) (gcm *Gcm, err error) {
	block, err := aes.NewCipher([]byte(hashext.GetSha256(key)[:16]))
	if err != nil {
		loger.Log.Errorf("NewGCM: Failed to create Cipher with key %s! %s", key, err)
		return nil, err
	}
	aesGCM, err_ := cipher.NewGCM(block)
	if err_ != nil {
		loger.Log.Errorf("NewGCM: Failed to create GCM with key %s! %s", key, err)
		return nil, err
	}
	return &Gcm{Key: []byte(key), aesGCM: aesGCM}, nil
}

type Gcm struct {
	Key    []byte
	aesGCM cipher.AEAD
}

// AesGcmEncrypt 使用aes-ctr加密一个byte数组
func (g *Gcm) AesGcmEncrypt(data []byte) (res []byte, err error) {
	if len(data)-40 > dataLength {
		loger.Log.Errorf("AesGcmEncrypt: An error occurred while encrypting data!%s", err)
		return nil, errors.New("lengthError")
	} else {
		nonce := make([]byte, g.aesGCM.NonceSize()) // nonce size: 12, tag size: 16
		if _, err_ := io.ReadFull(rand.Reader, nonce); err_ != nil {
			loger.Log.Errorf("AesGcmEncrypt: An error occurred while encrypting data!%s", err)
			return nil, err_
		}
		fmt.Println("Nonce", len(nonce))
		ciphertext := g.aesGCM.Seal(nil, nonce, data, nil)
		return append(nonce, ciphertext...), nil
	}
}

// AesGcmDecrypt 解密使用aes-ctr加密的byte数组
func (g *Gcm) AesGcmDecrypt(data []byte) ([]byte, error) {
	nonceSize := g.aesGCM.NonceSize()
	if len(data) < nonceSize {
		loger.Log.Error("AesGcmDecrypt: Decrypted data length is too small!")
		return nil, errors.New("ciphertextTooShort")
	}

	nonce, ciphertext := data[:nonceSize], data[nonceSize:]
	plaintext, err := g.aesGCM.Open(nil, nonce, ciphertext, nil)
	if err != nil {
		loger.Log.Errorf("AesGcmDecrypt: An error occurred while decrypting data! %s", err)
		return nil, err
	}
	return plaintext, nil
}

// B64GCMEncrypt 使用aes-ctr加密并转换为base64
func (g *Gcm) B64GCMEncrypt(data []byte) (string, error) {
	ciphertext, err := g.AesGcmEncrypt(data)
	if err != nil {
		loger.Log.Errorf("B64GCMEncrypt: An error occurred! %s", err)
		return "", err
	}
	return base64.StdEncoding.EncodeToString(ciphertext), nil
}

// B64GCMDecrypt 解密一个使用aes-ctr base64转码的字符串
func (g *Gcm) B64GCMDecrypt(data string) ([]byte, error) {
	ciphertext, err := base64.StdEncoding.DecodeString(data)
	if err != nil {
		loger.Log.Errorf("B64GCMEncrypt: An error occurred! %s", err)
		return nil, err
	}
	result, err := g.AesGcmDecrypt(ciphertext)
	if err != nil {
		loger.Log.Errorf("B64GCMEncrypt: An error occurred! %s", err)
		return nil, err
	}
	return result, nil
}

// StrB64GCMEncrypt 使用aes-ctr加密并转换为base64
func (g *Gcm) StrB64GCMEncrypt(data string) (string, error) {
	ciphertext, err := g.AesGcmEncrypt([]byte(data))
	if err != nil {
		loger.Log.Errorf("StrB64GCMEncrypt: An error occurred! %s", err)
		return "", err
	}
	return base64.StdEncoding.EncodeToString(ciphertext), nil
}

// StrB64GCMDecrypt 解密一个使用aes-ctr base64转码的字符串
func (g *Gcm) StrB64GCMDecrypt(data string) (string, error) {
	ciphertext, err := base64.StdEncoding.DecodeString(data)
	if err != nil {
		loger.Log.Errorf("StrB64GCMDecrypt: An error occurred! %s", err)
		return "", err
	}
	result, err := g.AesGcmDecrypt(ciphertext)
	if err != nil {
		loger.Log.Errorf("StrB64GCMDecrypt: An error occurred! %s", err)
		return "", err
	}
	return string(result), nil
}
