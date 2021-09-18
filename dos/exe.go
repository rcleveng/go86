package go86

import (
	"encoding/binary"
	"errors"
	"io"
	"os"
)

/**
 * Go structure represenging the EXE header of a MSDOS binary
 * Please note that this is not memory layout compatible with
 * the actual DOS structure
 */
type exeReloEntry struct {
	Offset  uint16
	Segment uint16
}

type ExeHeader struct {
	Signature          uint16
	BytesInLastBlock   uint16
	BlocksInFile       uint16
	NumRelos           uint16
	HeaderParagraphs   uint16
	MinExtraParagraphs uint16
	MaxExtraParagraphs uint16
	SS                 uint16
	SP                 uint16
	Checksum           uint16
	IP                 uint16
	CS                 uint16
	ReloTableOffset    uint16
	OverlayNumber      uint16
	Relos              []exeReloEntry
}

type ExeType int

const (
	NONE ExeType = iota
	EXE
	COM
	IMAGE
)

type Executable struct {
	Etype  ExeType
	Exists bool
	Hdr    ExeHeader
	Data   []byte
}

// How many segments are needed to load this executable.  For EXEs we look
// at the header, for COM files it's just 64k
func (exe *Executable) SegmentsNeeded() int {
	switch exe.Etype {
	case EXE:
		return int((exe.Hdr.BlocksInFile/32)+exe.Hdr.MinExtraParagraphs) + 1
	case COM, IMAGE:
		return 0x1000 // 64K
	default:
		panic("invalid etype specified")
	}
}

func ReadExeFromFile(filename string) (*Executable, error) {
	file, err := os.Open(filename)
	if err != nil {
		return nil, err
	}
	defer file.Close()

	b, err := io.ReadAll(file)
	if err != nil {
		return nil, err
	}
	if len(b) < 28 {
		return nil, errors.New("File header too small or size unknown for: " + filename)
	}

	return ReadExe(b)
}

func ReadExe(bs []byte) (*Executable, error) {

	exe := &Executable{Etype: EXE}
	exe.Exists = true
	if bs[0] != 'M' && bs[1] != 'Z' {
		exe.Etype = COM
		exe.Data = bs
		return exe, nil
	}
	exe.Hdr.Signature = binary.LittleEndian.Uint16(bs)
	exe.Hdr.BytesInLastBlock = binary.LittleEndian.Uint16(bs[2:])
	exe.Hdr.BlocksInFile = binary.LittleEndian.Uint16(bs[4:])
	exe.Hdr.NumRelos = binary.LittleEndian.Uint16(bs[6:])
	exe.Hdr.HeaderParagraphs = binary.LittleEndian.Uint16(bs[8:])
	exe.Hdr.MinExtraParagraphs = binary.LittleEndian.Uint16(bs[10:])
	exe.Hdr.MaxExtraParagraphs = binary.LittleEndian.Uint16(bs[12:])
	exe.Hdr.SS = binary.LittleEndian.Uint16(bs[14:])
	exe.Hdr.SP = binary.LittleEndian.Uint16(bs[16:])
	exe.Hdr.Checksum = binary.LittleEndian.Uint16(bs[18:])
	exe.Hdr.IP = binary.LittleEndian.Uint16(bs[20:])
	exe.Hdr.CS = binary.LittleEndian.Uint16(bs[22:])
	exe.Hdr.ReloTableOffset = binary.LittleEndian.Uint16(bs[24:])
	exe.Hdr.OverlayNumber = binary.LittleEndian.Uint16(bs[26:])

	// Seek to relo table
	pos := int(exe.Hdr.ReloTableOffset)

	for i := 0; i < int(exe.Hdr.NumRelos); i++ {
		var nr exeReloEntry
		offoff := pos + (i * 4)
		segoff := offoff + 2
		nr.Offset = binary.LittleEndian.Uint16(bs[offoff:])
		nr.Segment = binary.LittleEndian.Uint16(bs[segoff:])
		exe.Hdr.Relos = append(exe.Hdr.Relos, nr)
	}

	bin := bs[(exe.Hdr.HeaderParagraphs * 0x10):]
	exe.Data = bin
	return exe, nil
}
