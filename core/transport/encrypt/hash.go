package encrypt

import (
	"crypto/sha256"
	"errors"
)

func getPasswordHash(password string, len int) (hash []byte, err error) {
	hasher := sha256.New()
	hasher.Write([]byte(password + "exsync"))
	if len > 256/8 {
		return nil, errors.New("obtain key hash length exceeding 256")
	}
	return hasher.Sum(nil)[:len], nil
}
