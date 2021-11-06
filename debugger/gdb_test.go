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

func TestCreateChecksumSmoke(t *testing.T) {
	c := CreateChecksum("AAAA")
	if c != 4 {
		t.Errorf("expecting 4, not '%d'", c)
	}
}

func TestGetPacketDataSmoke(t *testing.T) {
	raw := []byte("$AAAA#04")
	s := GdbSession{}
	data, err := s.GetPacketData(raw, len(raw))
	if err != nil {
		t.Error(err)
	}
	if data != "AAAA" {
		t.Errorf("expected 'AAAA', not '%s'", data)
	}
}
