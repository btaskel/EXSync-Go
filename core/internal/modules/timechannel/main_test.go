package timechannel

import (
	"testing"
)

func TestNewTimeChannel(t *testing.T) {
	timeChannel := NewTimeChannel()
	timeChannel.Close()
}
