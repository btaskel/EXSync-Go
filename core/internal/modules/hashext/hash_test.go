package hashext

import (
	"fmt"
	"github.com/cespare/xxhash/v2"
	"io"
	"os"
	"testing"
)

//go test .\hash_test.go .\hash.go

func TestGetRandomStr(t *testing.T) {
	t.Logf("TestGetRandomStr passed:%v", GetRandomStr(6))
}

func TestGetXXHash(t *testing.T) {
	hash, err := GetXXHash("test.txt")
	if err != nil {
		return
	}
	fmt.Println(hash)
	t.Logf("hash: %v", hash)
}

func TestGetSha384(t *testing.T) {
	hash := GetSha384("测试文字")
	fmt.Println(hash)
}

func TestGetSha256(t *testing.T) {
	hash := GetSha256("测试文字")
	fmt.Println(hash)
}

func TestUpdateXXHash(t *testing.T) {
	f, err := os.Open("hash.go")
	if err != nil {
		fmt.Println("1")
	}
	hasher := xxhash.New()
	for {
		buf := make([]byte, 8192)
		n, err := f.Read(buf)
		if err != nil && err != io.EOF {
			return
		}
		if n == 0 {
			break
		}
		//fmt.Println(string(buf[:n]))

		_, err = hasher.Write(buf[:n])
		if err != nil {
			return
		}
	}
	fmt.Println(hasher.Sum([]byte{5, 85}))

}
