package transport

import (
	"EXSync/core/transport/encrypt"
)

func getCipherLoss(cipher *encrypt.Cipher) int {
	if cipher == nil {
		return 0
	}
	return cipher.Info.GetLossLen()
}

func RegisterCipher(method string, cipherInfo encrypt.CipherInfo) error {
	return encrypt.Register(method, cipherInfo)
}
