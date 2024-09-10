package go86

import (
	"encoding/binary"
)

type Memory struct {
	mem []byte
}

func NewMemory(size int) *Memory {
	m := Memory{}
	m.mem = make([]byte, size)
	return &m
}

func (m *Memory) Mem8(seg uint, off uint) uint8 {
	pos := (seg * 0x10) + off
	return m.mem[pos]
}

func (m *Memory) At(seg int, off int) []uint8 {
	pos := (seg * 0x10) + off
	return m.mem[pos:]
}

func (m *Memory) AtAbs(pos int) []uint8 {
	return m.mem[pos:]
}

func (m *Memory) AbsMem8(pos int) uint8 {
	return m.mem[pos]
}

func (m *Memory) PutMem8(seg uint, off uint, val uint8) {
	pos := (seg * 0x10) + off
	m.mem[pos] = val
}

func (m *Memory) Mem16(seg uint, off uint) uint16 {
	pos := (seg * 0x10) + off
	return binary.LittleEndian.Uint16(m.mem[pos:])
}

func (m *Memory) PutMem16(seg uint, off uint, val uint16) {
	pos := (seg * 0x10) + off
	binary.LittleEndian.PutUint16(m.mem[pos:], val)
}
