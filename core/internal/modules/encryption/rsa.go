package encryption

import (
	loger "EXSync/core/log"
	"crypto/rand"
	"crypto/rsa"
	"encoding/base64"
)

// GenerateKey 生成 RSA 密钥对
func GenerateKey() (publicKey *rsa.PublicKey, privateKey *rsa.PrivateKey, err error) {
	privateKey, err = rsa.GenerateKey(rand.Reader, 2048)
	if err != nil {
		loger.Log.Errorf("GenerateKey: Failed to generate key pair! %s", err)
		return nil, nil, err
	}
	publicKey = &privateKey.PublicKey
	return publicKey, privateKey, nil
}

// RsaEncryptBase64 使用 RSA 公钥加密数据, 返回加密后并编码为 base64 的数据
func RsaEncryptBase64(originalData []byte, publicKey *rsa.PublicKey) (string, error) {
	encryptedData, err := rsa.EncryptPKCS1v15(rand.Reader, publicKey, originalData)
	if err != nil {
		loger.Log.Errorf("RsaEncryptBase64: Encryption failed! %s", err)
		return "", err
	}
	return base64.StdEncoding.EncodeToString(encryptedData), nil
}

// RsaDecryptBase64 使用 RSA 私钥解密数据
func RsaDecryptBase64(encryptedData string, privateKey *rsa.PrivateKey) ([]byte, error) {
	encryptedDecodeBytes, err := base64.StdEncoding.DecodeString(encryptedData)
	if err != nil {
		loger.Log.Errorf("RsaDecryptBase64: Decryption failed! %s", err)
		return nil, err
	}
	originalData, err := rsa.DecryptPKCS1v15(rand.Reader, privateKey, encryptedDecodeBytes)
	return originalData, nil
}
