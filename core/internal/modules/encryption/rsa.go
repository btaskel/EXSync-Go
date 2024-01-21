package encryption

import (
	"crypto/rand"
	"crypto/rsa"
	"encoding/base64"
)

// GenerateKey 生成 RSA 密钥对
func GenerateKey() (*rsa.PublicKey, *rsa.PrivateKey, error) {
	privateKey, err := rsa.GenerateKey(rand.Reader, 2048)
	if err != nil {
		return nil, nil, err
	}
	publicKey := &privateKey.PublicKey
	return publicKey, privateKey, nil
}

// RsaEncryptBase64 使用 RSA 公钥加密数据, 返回加密后并编码为 base64 的数据
func RsaEncryptBase64(originalData []byte, publicKey *rsa.PublicKey) (string, error) {
	encryptedData, err := rsa.EncryptPKCS1v15(rand.Reader, publicKey, originalData)
	return base64.StdEncoding.EncodeToString(encryptedData), err
}

// RsaDecryptBase64 使用 RSA 私钥解密数据
func RsaDecryptBase64(encryptedData string, privateKey *rsa.PrivateKey) ([]byte, error) {
	encryptedDecodeBytes, err := base64.StdEncoding.DecodeString(encryptedData)
	if err != nil {
		return nil, err
	}
	originalData, err := rsa.DecryptPKCS1v15(rand.Reader, privateKey, encryptedDecodeBytes)
	return originalData, err
}
