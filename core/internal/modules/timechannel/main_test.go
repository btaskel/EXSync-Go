package timechannel

import (
	"fmt"
	"testing"
)

func TestNewTimeChannel(t *testing.T) {
	timeChannel := NewTimeChannel()
	defer timeChannel.Close()

	err := timeChannel.CreateRecv("abcdefgl")
	if err != nil {
		fmt.Println("1 :", err)
		return
	}
	timeChannel.Set("abcdefgl", []byte{50, 30})
	result, err := timeChannel.GetTimeout("abcdefgl", 3)
	if err != nil {
		fmt.Println("2 :", err)
		return
	}
	fmt.Println("output:", result)
}
