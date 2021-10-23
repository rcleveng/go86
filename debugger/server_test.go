package go86

import (
	"testing"
)

func TestCreatePacketSmoke(t *testing.T) {
	s := CreatePacket("AAAA")

	if s != "$AAAA#4" {
		t.Errorf("Error '%s' was not expected", s)
	}
}
