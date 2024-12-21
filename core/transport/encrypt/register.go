package encrypt

import (
	"errors"
)

func Register(method string, ci CipherInfo) error {
	if ci.keyLen < 0 {
		return errors.New("the Key length cannot be less than 1")
	}
	if ci.ivLen < 0 {
		return errors.New("the IV length cannot be less than 1")
	}
	if ci.lossLen < 0 {
		return errors.New("the Loss length cannot be less than 1")
	}
	cipherMethod[method] = ci
	return nil
}
