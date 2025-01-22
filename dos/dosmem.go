package go86

import (
	"errors"
	"fmt"
	"strings"
)

// Represents a DOS managed memory block, it may either be free to split
// or assign as-is, or is currently owned by a PSP
type DosMemBlock struct {
	Avail bool
	Start uint
	// Use golang convention that end == last element +1
	End uint
	// The Segment of the PSP of the owning process. This number is also
	// referred to as the DOS PID
	Owner       uint
	ProgramName string
}

// Implement Stringer
func (b DosMemBlock) String() string {
	return fmt.Sprintf("DosMemBlock: paragraphs: [Start:0x%04X][End:0x%04X][Size: 0x%04X][Avail: %t]",
		b.Start, b.End, b.End-b.Start, b.Avail)
}

type FitStrategy int

const (
	First FitStrategy = iota
	Best
	Last
)

const allowedSlackSpace = 512 // 512 * 16 == 8k

type DosMem struct {
	StartSeg uint
	EndSeg   uint
	Fit      FitStrategy
	Blocks   []DosMemBlock
}

func NewDosMem(start uint, end uint) *DosMem {
	d := &DosMem{StartSeg: start, EndSeg: end}
	return d
}

func (m DosMemBlock) Size() uint {
	return m.End - m.Start
}

// Stringer for DosMem
func (m DosMem) String() string {
	var s strings.Builder
	for _, b := range m.Blocks {
		fmt.Fprintf(&s, "%s\n", b.String())
	}
	return s.String()
}

func (m *DosMem) FindBlock(start uint) (blockNum int, found bool) {
	for i, b := range m.Blocks {
		if b.Start == start {
			return i, true
		}
	}
	return -1, false
}

// Resizes an existing allocated block.
func (m *DosMem) Resize(start uint, needed uint) (uint, error) {
	startBlock, found := m.FindBlock(start)
	if !found {
		return 0, errors.New("not found")
	}
	lastBlock := startBlock
	for i := startBlock + 1; i < len(m.Blocks); i++ {
		if !m.Blocks[i].Avail {
			break
		}
		lastBlock = i
		if m.Blocks[i].End-start >= needed {
			break
		}
	}

	if startBlock == lastBlock {
		// nothingn happened!
		return 0, errors.New("unable to resize")
	}
	// assign old start to last block
	m.Blocks[lastBlock].Start = start
	// remove [startBlock, .. )lastBlock
	m.Blocks = append(m.Blocks[:startBlock], m.Blocks[lastBlock:]...)

	// Split last block
	lastBlock, found = m.FindBlock(start)
	if !found {
		panic("huh?? this was found once before")
	}

	if m.Blocks[lastBlock].Size()+allowedSlackSpace > needed {
		// Need to split last block
		nb := DosMemBlock{
			Avail:       true,
			Start:       start + needed,
			End:         m.Blocks[lastBlock].End,
			Owner:       0,
			ProgramName: "",
		}
		m.Blocks[lastBlock].End = start + needed
		m.Blocks = append(m.Blocks, DosMemBlock{})
		copy(m.Blocks[lastBlock+1:], m.Blocks[lastBlock:])
		m.Blocks[lastBlock+1] = nb
	}
	newSize := m.Blocks[lastBlock].Size()
	return newSize, nil
}

func (m *DosMem) AllocateFirst(size uint) (*DosMemBlock, error) {
	for i := 0; i < len(m.Blocks); i++ {
		if !m.Blocks[i].Avail {
			continue
		}
		// We have a free block now
		cursize := m.Blocks[i].End - m.Blocks[i].Start
		if cursize < size {
			continue
		}
		// Allocate whole block if it's close (8k), otherwise we shall split it.
		if cursize+allowedSlackSpace <= size {
			m.Blocks[i].Avail = false
			return &m.Blocks[i], nil
		}
		nb := DosMemBlock{
			Avail:       false,
			Start:       m.Blocks[i].Start,
			End:         m.Blocks[i].Start + size,
			Owner:       0,
			ProgramName: "",
		}
		m.Blocks[i].Start = nb.End
		m.Blocks = append(m.Blocks, DosMemBlock{})
		copy(m.Blocks[i+1:], m.Blocks[i:])
		m.Blocks[i] = nb
		return &m.Blocks[i], nil
	}
	return nil, fmt.Errorf("unable to allocate memory of size: %d", size)
}

func (m *DosMem) AllocateLast(size uint) (*DosMemBlock, error) {
	for i := len(m.Blocks) - 1; i >= 0; i-- {
		if !m.Blocks[i].Avail {
			continue
		}
		// We have a free block now
		cursize := m.Blocks[i].End - m.Blocks[i].Start
		if cursize < size {
			continue
		}
		// Allocate whole block if it's close (8k), otherwise we shall split it.
		if cursize+allowedSlackSpace <= size {
			m.Blocks[i].Avail = false
			return &m.Blocks[i], nil
		}
		m.Blocks = append(m.Blocks, DosMemBlock{})
		copy(m.Blocks[i+1:], m.Blocks[i:])
		// i will be the old free block, shrink end
		m.Blocks[i].End -= size

		// Allocate i+1, so increase the start, leave end alone
		m.Blocks[i+1].Start = m.Blocks[i].End
		m.Blocks[i+1].Avail = false
		return &m.Blocks[i+1], nil
	}
	return nil, errors.New("unable to allocate memory")
}

func (m *DosMem) Allocate(size uint) (*DosMemBlock, error) {
	if len(m.Blocks) == 0 {
		m.Blocks = append(m.Blocks, DosMemBlock{
			Avail: true,
			Start: m.StartSeg,
			End:   m.EndSeg,
		})
	}
	switch m.Fit {
	case Best:
		return nil, errors.New("best fit not implemented")
	case Last:
		return m.AllocateLast(size)
	case First:
		return m.AllocateFirst(size)
	}
	return nil, errors.New("unknown allocation strategy")
}
