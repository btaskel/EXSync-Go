package encryption

import (
	"fmt"
	"testing"
)

func TestGcm_AesGcmEncrypt(t *testing.T) {
	// 10 -> 38 = 28
	gcm, err := NewGCM("123")
	if err != nil {
		fmt.Println("1")
		return
	}
	originData := []byte("0123456789012345678901234567890123456789")
	enData, err2 := gcm.AesGcmEncrypt(originData)
	if err2 != nil {
		fmt.Println("2")
		return
	}
	decrypt, err := gcm.AesGcmDecrypt(enData)
	if err != nil {
		fmt.Println("3")
		return
	}
	// 28

	fmt.Println(decrypt)
	fmt.Println(len(originData))
	fmt.Println(len(enData))
}

func TestGcm_B64GCMDecrypt(t *testing.T) {
	// 9 -> 52 = 43
	gcm, err := NewGCM("123456")
	if err != nil {
		fmt.Println(1)
		return
	}
	originData := []byte("123456789")
	encrypt, err := gcm.B64GCMEncrypt(originData)
	if err != nil {
		fmt.Println(2)
		return
	}
	decrypt, err := gcm.B64GCMDecrypt(encrypt)
	if err != nil {
		fmt.Println(3)
		return
	}
	fmt.Println(len(originData))
	fmt.Println(len(encrypt))
	fmt.Println(string(decrypt))

}
