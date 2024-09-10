package go86

import (
	"encoding/hex"
	"fmt"
	"testing"

	"gotest.tools/v3/assert"
)

func TestMem8(t *testing.T) {
	m := NewMemory(1024 * 1024)
	m.SetMem8(1, 100, 0xCA)

	actual := m.GetMem8(1, 100)
	if actual != 0xCA {
		t.Errorf("Expected 0xCA, got: %x", actual)
	}
}

func TestMem16(t *testing.T) {
	m := NewMemory(1024 * 1024)
	m.SetMem16(1, 100, 0xBACE)

	actual := m.GetMem16(1, 100)
	if actual != 0xBACE {
		t.Errorf("Expected 0xBACE, got: %x", actual)
	}
}

func TestMemMixMem8And16(t *testing.T) {
	m := NewMemory(1024 * 1024)
	m.SetMem16(1, 100, 0xCAFE)

	a1 := m.GetMem8(1, 100)
	if a1 != 0xFE {
		t.Errorf("Expected CA, got: %x", a1)
	}
	a2 := m.GetMem8(1, 101)
	if a2 != 0xCA {
		t.Errorf("Expected CA, got: %x", a2)
	}
}

func TestMemAt16(t *testing.T) {
	m := NewMemory(1024 * 1024)
	m.SetMem16(1, 100, 0xCAFE)
	m.SetMem16(1, 102, 0xBEEF)

	s := m.At(1, 100)

	str := fmt.Sprintf("%X/%X/%X/%X", s[0], s[1], s[2], s[3])
	if str != "FE/CA/EF/BE" {
		t.Errorf("Expected FE/CA/EF/BE, got: %s", str)
	}
}

// B801008ED88D160A00B409CD21B87F00BA010002C2B44CCD21
func TestMemAtSet(t *testing.T) {
	m := NewMemory(1024 * 1024)
	d, err := hex.DecodeString("B801008ED88D160A00B409CD21B87F00BA010002C2B44CCD21")
	if err != nil {
		t.Fatalf("Error: %s", err)
	}

	copy(m.At(1, 100), d)
	assert.Equal(t, m.GetMem8(1, 100), uint8(0xB8))
	assert.Equal(t, m.GetMem8(1, 101), uint8(0x01))
	assert.Equal(t, m.GetMem8(1, 102), uint8(0x00))
}

func TestMemAtAll8(t *testing.T) {
	m := NewMemory(1024 * 1024)
	m.SetMem8(1, 100, 0xCA)
	m.SetMem8(1, 101, 0xFE)
	m.SetMem8(1, 102, 0xBE)
	m.SetMem8(1, 103, 0xEF)

	s := m.At(1, 100)

	str := fmt.Sprintf("%X%X%X%X", s[0], s[1], s[2], s[3])
	if str != "CAFEBEEF" {
		t.Errorf("Expected CAFEBEEF, got: %s", str)
	}
}
