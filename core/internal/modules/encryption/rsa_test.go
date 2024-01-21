package encryption

import (
	"fmt"
	"testing"
)

func TestGenerateKey(t *testing.T) {
	publicKey, _, err := GenerateKey()
	if err != nil {
		return
	}
	fmt.Println(publicKey)
}
