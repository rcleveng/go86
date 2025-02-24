package go86

import (
	"encoding/binary"

	"github.com/golang/glog"
)

type Memory struct {
	mem  []byte
	size int
}

func NewMemory(size int) *Memory {
	return &Memory{
		size: size,
		mem:  make([]byte, size),
	}
}

func (m *Memory) GetMem8(seg uint, off uint) uint8 {
	if off >= 0x10000 {
		glog.Warning("GetMem8: off >= 0x100: ", off)
	}
	pos := ((seg * 0x10) + (off & 0xffff))

	return m.mem[pos]
}

func (m *Memory) At(seg uint, off uint) []uint8 {
	if off >= 0x10000 {
		glog.Warning("GetMem8: off >= 0x100: ", off)
	}
	pos := ((seg * 0x10) + (off & 0xffff))
	return m.mem[pos:]
}

func (m *Memory) AtAbs(pos int) []uint8 {
	return m.mem[pos:]
}

func (m *Memory) AbsMem8(pos int) uint8 {
	return m.mem[pos]
}

func (m *Memory) SetMem8(seg uint, off uint, val uint8) {
	if off >= 0x10000 {
		glog.Warning("GetMem8: off >= 0x100: ", off)
	}
	pos := ((seg * 0x10) + (off & 0xffff))
	m.mem[pos] = val
}

func (m *Memory) GetMem16(seg uint, off uint) uint16 {
	if off >= 0x10000 {
		glog.Warning("GetMem8: off >= 0x100: ", off)
	}
	pos := ((seg * 0x10) + (off & 0xffff))
	return binary.LittleEndian.Uint16(m.mem[pos:])
}

func (m *Memory) SetMem16(seg uint, off uint, val uint16) {
	if off >= 0x10000 {
		glog.Warning("GetMem8: off >= 0x100: ", off)
	}
	pos := ((seg * 0x10) + (off & 0xffff))
	binary.LittleEndian.PutUint16(m.mem[pos:], val)
}
