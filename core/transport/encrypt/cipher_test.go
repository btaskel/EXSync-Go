package encrypt

import (
	"fmt"
	"testing"
)

func TestNewCipher(t *testing.T) {
	cipher, err := NewCipher(Aes256Gcm, "123")
	if err != nil {
		return
	}

	origin := make([]byte, 4096)
	result := make([]byte, 4096)
	fmt.Println("lossLen: ", cipher.Info.lossLen)
	copy(origin[2+cipher.Info.lossLen:], "测试文字")
	fmt.Println(origin)
	// 4*3 = UTF-8 3字节 * 3个字符
	err = cipher.Encrypt(origin[:2+cipher.Info.lossLen+4*3], result)
	if err != nil {
		fmt.Println(err)
		return
	}
	fmt.Println("密文:", result)

	decryptText := make([]byte, 4096)
	err = cipher.Decrypt(result[:2+cipher.Info.lossLen+4*3], decryptText)
	if err != nil {
		fmt.Println(err)
		return
	}
	fmt.Println(decryptText)
}
