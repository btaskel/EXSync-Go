package encrypt

import (
	"fmt"
	"testing"
)

func TestNewCipher2(t *testing.T) {
	cipher, err := NewCipher(Aes128Gcm, "123")
	if err != nil {
		return
	}
	err = cipher.InitEncrypt()
	if err != nil {
		fmt.Println(err)
		return
	}
	s := []byte(("hello,world"))
	result := cipher.Encrypt(s)
	fmt.Println(result)
	err = cipher.InitDecrypt()
	if err != nil {
		fmt.Println(err)
		return
	}
	err = cipher.Decrypt(result)
	if err != nil {
		fmt.Println(err)
		return
	}
	fmt.Println(string(result))
}
