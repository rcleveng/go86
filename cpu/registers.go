package go86

import log "github.com/golang/glog"

type Reg int

// Constants to define the non-segment register mapping between its name
// and the correspoding index in the cpu.Regs array.
const (
	AX Reg = iota
	CX
	DX
	BX
	SP
	BP
	SI
	DI
	maxRegs
)

type Reg8 int

// AL/AH shares space with AX
const (
	AL Reg8 = iota
	CL
	DL
	BL
	AH
	CH
	DH
	BH
)

type SReg int

// Constants to define the segment register mapping between its name
// and the corresponding index in the cpu.Sregs array.
const (
	ES SReg = iota
	CS
	SS
	DS
	maxSegs
)

type Registers struct {
	// 16-bit registers
	// Set of registers (AX, BX, CX, etc...)
	regs [maxRegs]uint
	// Set of segment registers (CS, DS, etc...)
	sregs [maxSegs]uint
}

// returns a register as an unisgned int
// useful when handling operations.
func (r Registers) GetReg16(reg Reg) uint {
	return r.regs[reg] & 0xffff
}

func (r *Registers) SetReg16(reg Reg, val uint) {
	r.regs[reg] = val & 0xffff
}

func (r Registers) GetReg8(reg Reg8) uint {
	if reg == AH || reg == CH || reg == DH || reg == BH {
		value := r.regs[reg-4]
		return (value & 0xff00) >> 8
	}
	return r.regs[reg] & 0xff
}

func (r *Registers) SetReg8(reg Reg8, val uint) {
	if reg == AH || reg == CH || reg == DH || reg == BH {
		r.regs[reg-4] &= 0x00ff
		r.regs[reg-4] |= (val & 0xff) << 8
	} else {
		r.regs[reg] &= 0xff00
		r.regs[reg] |= val & 0xff
	}
}

func (r Registers) GetSeg16(sreg SReg) uint {
	return r.sregs[sreg] & 0xffff
}

func (r Registers) GetSeg(sreg SReg) uint {
	return r.sregs[sreg]
}

func (r Registers) CS() uint {
	return r.sregs[CS] & 0xffff
}

func (r Registers) DS() uint {
	return r.sregs[DS] & 0xffff
}

func (r Registers) ES() uint {
	return r.sregs[ES] & 0xffff
}

func (r Registers) SS() uint {
	return r.sregs[SS] & 0xffff
}

func (r Registers) AX() uint {
	return r.regs[AX] & 0xffff
}

func (r Registers) BX() uint {
	return r.regs[BX] & 0xffff
}

func (r Registers) CX() uint {
	return r.regs[CX] & 0xffff
}

func (r Registers) DX() uint {
	return r.regs[DX] & 0xffff
}

func (r Registers) SP() uint {
	return r.regs[SP] & 0xffff
}

func (r Registers) BP() uint {
	return r.regs[BP] & 0xffff
}

func (r Registers) SI() uint {
	return r.regs[SI] & 0xffff
}

func (r Registers) DI() uint {
	return r.regs[DI] & 0xffff
}

func (r *Registers) SetSeg(sreg SReg, val uint) {
	r.sregs[sreg] = val
}

func (r *Registers) SetSeg16(sreg SReg, val uint) {
	r.sregs[sreg] = val & 0xffff
}

// Increment/Decrement

func (r *Registers) Inc16(reg Reg, num uint) {
	r.regs[reg] = r.regs[reg] + num
}

func (r *Registers) Dec16(reg Reg, num uint) {
	r.regs[reg] = r.regs[reg] - num
}

// Push an value onto the stack
func (r *Registers) Push16(mem *Memory, val uint16) {
	r.regs[SP] = r.regs[SP] - 2
	seg := r.sregs[SS]
	off := r.regs[SP]
	mem.PutMem16(seg, off, val)
	log.V(1).Infof("   PUSH: [%04X:%04X] = %04X", seg, off, val)
}

// Pop an value off the stack
func (r *Registers) Pop16(mem *Memory) uint16 {
	seg := r.sregs[SS]
	off := r.regs[SP]
	v := mem.Mem16(seg, off)
	r.regs[SP] = r.regs[SP] + 2
	log.V(1).Infof("   POP:%04X [%04X:%04X]", r, seg, off)
	return v
}

// Push a segment register onto the stack
func (r *Registers) PushReg16(reg Reg, mem *Memory) {
	r.Push16(mem, uint16(r.regs[reg]))
}

// Pop a segment register off the stack
func (r *Registers) PopReg16(reg Reg, mem *Memory) {
	r.regs[reg] = uint(r.Pop16(mem))
}
