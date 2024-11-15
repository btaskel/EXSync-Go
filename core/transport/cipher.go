package transport

import "EXSync/core/transport/encrypt"

func getCipherLoss(cipher *encrypt.Cipher) int {
	if cipher == nil {
		return 0
	} else {
		return cipher.Info.GetLossLen()
	}
}
