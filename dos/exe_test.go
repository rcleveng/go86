package go86

import (
	"encoding/binary"
	"encoding/hex"
	"testing"

	"gotest.tools/v3/assert"
)

func TestExeSmoke(t *testing.T) {
	hdr := make([]byte, 500)
	binary.LittleEndian.PutUint16(hdr, 0x5A4D)
	// 1 byte on last page
	binary.LittleEndian.PutUint16(hdr[2:], 0x0028)
	// 2 pages
	binary.LittleEndian.PutUint16(hdr[4:], 0x0002)
	// 1 relo item
	binary.LittleEndian.PutUint16(hdr[6:], 0x0001)
	// header size in paragraphs
	binary.LittleEndian.PutUint16(hdr[8:], 0x0020)
	// min alloc in paragraphs
	binary.LittleEndian.PutUint16(hdr[10:], 0x0011)
	// max alloc in paragraphs
	binary.LittleEndian.PutUint16(hdr[12:], 0xFFFF)
	// SS
	binary.LittleEndian.PutUint16(hdr[14:], 0x0003)
	// SP
	binary.LittleEndian.PutUint16(hdr[16:], 0x0100)
	// Checksum (0)
	binary.LittleEndian.PutUint16(hdr[18:], 0x0000)
	// IP (0)
	binary.LittleEndian.PutUint16(hdr[20:], 0x0000)
	// CS
	binary.LittleEndian.PutUint16(hdr[22:], 0x0100)
	// RELO OFFSET
	binary.LittleEndian.PutUint16(hdr[24:], 0x001E)
	// OVERlay Number
	binary.LittleEndian.PutUint16(hdr[26:], 0x0000)
	// Hello World
	slice, _ := hex.DecodeString("B801008ED88D160A00B409CD21B87F00BA010002C2B44CCD21")
	copy(hdr[0x38:], slice)

	e, err := ReadExe(hdr)
	if err != nil {
		t.Error("Error: ", err)
	}
	assert.Equal(t, e.Hdr.CS, uint16(0x0100))
}
