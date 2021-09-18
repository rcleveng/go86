package go86

import (
	"encoding/hex"
	"fmt"
	"math/bits"
	"strings"

	log "github.com/golang/glog"
	"golang.org/x/arch/x86/x86asm"
)

const (
	REG_AX = iota
	REG_CX
	REG_DX
	REG_BX
	REG_SP
	REG_BP
	REG_SI
	REG_DI
	maxRegs
)

// AL/AH shares space with AX
const (
	REG_AL = iota
	REG_CL
	REG_DL
	REG_BL
)

// AL/AH shares space with AX
const (
	REG_AH = iota
	REG_CH
	REG_DH
	REG_BH
)

const (
	SREG_ES = iota
	SREG_CS
	SREG_SS
	SREG_DS
	maxSegs
)

// Flags
const (
	CF uint16 = 1 << iota
	_  uint16 = 1 << iota
	PF uint16 = 1 << iota
	_  uint16 = 1 << iota
	AF uint16 = 1 << iota
	_  uint16 = 1 << iota
	ZF uint16 = 1 << iota
	SF uint16 = 1 << iota
	TF uint16 = 1 << iota
	IF uint16 = 1 << iota
	DF uint16 = 1 << iota
	OF uint16 = 1 << iota
)

var regnums = map[x86asm.Reg]int{
	x86asm.AH: REG_AH,
	x86asm.BH: REG_BH,
	x86asm.CH: REG_CH,
	x86asm.DH: REG_DH,

	x86asm.AL: REG_AL,
	x86asm.BL: REG_BL,
	x86asm.CL: REG_CL,
	x86asm.DL: REG_DL,

	x86asm.AX: REG_AX,
	x86asm.BX: REG_BX,
	x86asm.CX: REG_CX,
	x86asm.DX: REG_DX,

	x86asm.BP: REG_BP,
	x86asm.SI: REG_SI,
	x86asm.SP: REG_SP,
	x86asm.DI: REG_DI,

	x86asm.CS: SREG_CS,
	x86asm.DS: SREG_DS,
	x86asm.ES: SREG_ES,
	x86asm.SS: SREG_SS,
}

// Map of Op name to function
var opcodes = make(map[x86asm.Op]func(*CPU, x86asm.Inst))

// Good SO posts for flags:
// https://stackoverflow.com/questions/51326423/how-to-calculate-the-auxiliary-flag-status-in-x86-assembly
// https://stackoverflow.com/questions/791991/about-assembly-cfcarry-and-ofoverflow-flag

func (cpu *CPU) ClearFlag(f uint16) {
	// This is cool
	cpu.Flags &^= f
}

func (cpu *CPU) SetFlagIf(f uint16, cond bool) {
	if cond {
		cpu.Flags |= f
	} else {
		cpu.Flags &^= f
	}
}

func (cpu *CPU) WriteFlag(sb *strings.Builder, flag uint16, on string, off string) {
	if (cpu.Flags & flag) != 0 {
		sb.WriteString(on)
	} else {
		sb.WriteString(off)
	}
	sb.WriteString(" ")
}

func (cpu *CPU) FlagsToString() string {
	var sb strings.Builder
	cpu.WriteFlag(&sb, OF, "O", "o")
	cpu.WriteFlag(&sb, DF, "D", "d")
	cpu.WriteFlag(&sb, IF, "I", "i")
	cpu.WriteFlag(&sb, SF, "S", "s")
	cpu.WriteFlag(&sb, ZF, "Z", "z")
	cpu.WriteFlag(&sb, AF, "A", "a")
	cpu.WriteFlag(&sb, PF, "P", "p")
	cpu.WriteFlag(&sb, CF, "C", "c")
	cpu.WriteFlag(&sb, TF, "T", "t")
	return sb.String()
}

func (cpu *CPU) FlagsToDebugString() string {
	var sb strings.Builder
	cpu.WriteFlag(&sb, OF, "OV", "NV")
	cpu.WriteFlag(&sb, DF, "DN", "UP")
	cpu.WriteFlag(&sb, IF, "EI", "DI")
	cpu.WriteFlag(&sb, SF, "NG", "PL")
	cpu.WriteFlag(&sb, ZF, "ZR", "NZ")
	cpu.WriteFlag(&sb, AF, "AC", "NA")
	cpu.WriteFlag(&sb, PF, "PE", "PO")
	cpu.WriteFlag(&sb, CF, "CY", "NC")
	return sb.String()
}

func (cpu *CPU) SetFlagsZSP(result uint, numbits int) {
	if numbits == 8 {
		cpu.SetFlagIf(SF, (result&0x80) != 0)
		cpu.SetFlagIf(ZF, (result&0xFF) == 0)
	} else {
		cpu.SetFlagIf(SF, (result&0x8000) != 0)
		cpu.SetFlagIf(ZF, (result&0xFFFF) == 0)
	}
	// Yes. X86 only ever uses LSB for PF
	count := bits.OnesCount8(uint8(result & 0xFF))
	cpu.SetFlagIf(PF, (count&0x01) == 0)
}

func (cpu *CPU) SetFlagsAdd(res, src, dst uint, bits int) {
	cpu.SetFlagIf(CF, res>>bits != 0)
	if bits == 8 {
		cpu.SetFlagIf(OF, (res^src)&(res^dst)&0x80 != 0)
	} else {
		cpu.SetFlagIf(OF, (res^src)&(res^dst)&0x8000 != 0)
	}
	cpu.SetFlagIf(AF, (res^src^dst)&0x10 != 0)
	cpu.SetFlagsZSP(res, bits)
}

func (cpu *CPU) SetFlagsSub(res, src, dst uint, numbits int) {
	if numbits == 8 {
		cpu.SetFlagIf(CF, res&0x100 == 0x100)
		cpu.SetFlagIf(OF, ((dst^src)&(res^dst)&0x80) != 0)
	} else {
		cpu.SetFlagIf(OF, ((dst^src)&(res^dst)&0x8000) != 0)
		cpu.SetFlagIf(CF, res&0x10000 == 0x10000)
	}
	cpu.SetFlagIf(AF, (res^src^dst)&0x10 != 0)
	cpu.SetFlagsZSP(res, numbits)
}

// Used in CMP, where just PSZ survives
func (cpu *CPU) ClearFlagsCOA() {
	cpu.Flags &^= (AF | CF | OF)
}

// Good SO post on bits and flags
// https://stackoverflow.com/questions/51326423/how-to-calculate-the-auxiliary-flag-status-in-x86-assembly

func (cpu *CPU) add(inst x86asm.Inst) {
	src := cpu.Rmm(inst.Args[1], inst)
	dst := cpu.Rmm(inst.Args[0], inst)
	res := dst + src
	cpu.PutRmm(inst.Args[0], inst, res)
	cpu.SetFlagsAdd(res, src, dst, instbits(inst))
}

func (cpu *CPU) sub(inst x86asm.Inst) {
	src := cpu.Rmm(inst.Args[1], inst)
	dst := cpu.Rmm(inst.Args[0], inst)
	res := dst - src
	cpu.PutRmm(inst.Args[0], inst, res)
	cpu.SetFlagsSub(res, src, dst, instbits(inst))
}

func (cpu *CPU) mov(inst x86asm.Inst) {
	src := cpu.Rmm(inst.Args[1], inst)
	cpu.PutRmm(inst.Args[0], inst, src)
}

// Push an argument onto the stack
func (cpu *CPU) Push(val uint16) {
	cpu.Regs[REG_SP] = cpu.Regs[REG_SP] - 2
	seg := int(cpu.Sregs[SREG_SS])
	off := int(cpu.Regs[REG_SP])
	cpu.Mem.PutMem16(seg, off, val)
	log.V(1).Infof("   PUSH: [%04X:%04X] = %04X", seg, off, val)
}

// Pop an argument off the stack
func (cpu *CPU) Pop() (r uint16) {
	seg := int(cpu.Sregs[SREG_SS])
	off := int(cpu.Regs[REG_SP])
	r = cpu.Mem.Mem16(seg, off)
	log.V(1).Infof("   POP:%04X [%04X:%04X]", r, seg, off)
	cpu.Regs[REG_SP] = cpu.Regs[REG_SP] + 2
	return r
}

func (cpu *CPU) iret(inst x86asm.Inst) {
	cpu.Ip = cpu.Pop()
	cpu.Sregs[SREG_CS] = cpu.Pop()
	cpu.Flags = cpu.Pop()
}

// Call an interrupt
func (cpu *CPU) intr(inst x86asm.Inst) {
	if inst.Args[0] != nil {
		switch a := inst.Args[0].(type) {
		case x86asm.Imm:
			intrno := int(a)
			if cpu.Intrs[intrno] != nil {
				log.V(4).Infof("Call Internal Interrupt # 0x%X", intrno)
				cpu.Intrs[intrno](cpu, intrno)
				return
			} else {
				// Call memory based x86 interrupt code
				cpu.Push(cpu.Flags)
				cpu.Push(cpu.Sregs[SREG_CS])
				cpu.Push(cpu.Ip)

				// Update CS:IP from interrupt table
				cpu.Ip = cpu.Mem.Mem16(0x0000, intrno*4)
				cpu.Sregs[SREG_CS] = cpu.Mem.Mem16(0x0000, 2+(intrno*4))
			}
		}
	}
}

func (cpu *CPU) effectiveAddress(arg x86asm.Arg) uint {
	a := arg.(x86asm.Mem)

	// Offset is effective address
	off := uint(a.Disp & 0xFFFF)
	if uint8(a.Base) != 0 {
		off += cpu.Reg(a.Base)
	}
	if uint8(a.Index) != 0 {
		off += cpu.Reg(a.Index)
	}
	return off
}

func (cpu *CPU) lea(inst x86asm.Inst) {
	off := cpu.effectiveAddress(inst.Args[1])
	cpu.PutRmm(inst.Args[0], inst, off)
}

func (cpu *CPU) jumprel(rel int32) {
	// Go cast fun with uint16
	if rel >= 0 {
		cpu.Ip += uint16(rel)
	} else {
		cpu.Ip -= uint16(-rel)
	}
}

func (cpu *CPU) jmp(inst x86asm.Inst) {
	switch a := inst.Args[0].(type) {
	case x86asm.Reg:
		log.V(3).Infoln("JMP reg")
		r := cpu.Rmm(a, inst)
		cpu.Ip = uint16(r)
	case x86asm.Rel:
		log.V(3).Infoln("JMP rel")
		rel := int32(inst.Args[0].(x86asm.Rel))
		cpu.jumprel(rel)
	case x86asm.Mem:
		log.V(3).Infoln("JMP abs mem")
		seg := int(cpu.segForEA(a, inst))
		off := int(cpu.effectiveAddress(a))
		cpu.Ip = cpu.Mem.Mem16(seg, off)
	default:
		panic(fmt.Sprintf("unknown type for a: %v", a))
	}
}

func (cpu *CPU) ljmp(inst x86asm.Inst) {
	switch a := inst.Args[0].(type) {
	case x86asm.Mem:
		log.V(3).Infoln("LJMP mem")
		seg := int(cpu.segForEA(a, inst))
		off := int(cpu.effectiveAddress(a))
		cpu.Ip = cpu.Mem.Mem16(seg, off)
		cpu.Sregs[SREG_CS] = cpu.Mem.Mem16(seg, off+2)
	default:
		panic(fmt.Sprintf("unknown type for a: %v", a))
	}
}

func (cpu *CPU) jc(inst x86asm.Inst) {
	rel := int32(inst.Args[0].(x86asm.Rel))
	if rel == 0 {
		// No jump, just bail
		return
	}

	cond := false
	switch inst.Op {
	case x86asm.JA:
		cond = cpu.Flags&(CF|ZF) == 0
	case x86asm.JAE:
		cond = cpu.Flags&CF == 0
	case x86asm.JB:
		cond = cpu.Flags&CF != 0
	case x86asm.JBE:
		cond = (cpu.Flags&CF != 0 || cpu.Flags&ZF != 0)
	case x86asm.JCXZ:
		cond = cpu.Regs[REG_CX] == 0
	case x86asm.JE:
		cond = cpu.Flags&ZF != 0
	case x86asm.JG:
		sf := cpu.Flags&SF != 0
		of := cpu.Flags&OF != 0
		zf := cpu.Flags&ZF != 0
		cond = !zf && (sf == of)
	case x86asm.JGE:
		sf := cpu.Flags&SF != 0
		of := cpu.Flags&OF != 0
		cond = sf == of
	case x86asm.JL:
		sf := cpu.Flags&SF != 0
		of := cpu.Flags&OF != 0
		cond = sf != of
	case x86asm.JLE:
		sf := cpu.Flags&SF != 0
		of := cpu.Flags&OF != 0
		zf := cpu.Flags&ZF != 0
		cond = zf || (sf != of)
	case x86asm.JNE:
		cond = cpu.Flags&ZF == 0
	case x86asm.JNO:
		cond = cpu.Flags&OF == 0
	case x86asm.JNP:
		cond = cpu.Flags&PF == 0
	case x86asm.JNS:
		cond = cpu.Flags&SF == 0
	case x86asm.JO:
		cond = cpu.Flags&OF != 0
	case x86asm.JP:
		cond = cpu.Flags&PF != 0
	case x86asm.JS:
		cond = cpu.Flags&SF != 0
	}

	if !cond {
		// Jump condition, just bail
		return
	}
	cpu.jumprel(rel)
}

func (cpu *CPU) loop(inst x86asm.Inst) {
	rel := int32(inst.Args[0].(x86asm.Rel))
	cpu.Regs[REG_CX]--
	if cpu.Regs[REG_CX] != 0 {
		cpu.jumprel(rel)
	}
}

func (cpu *CPU) loope(inst x86asm.Inst) {
	rel := int32(inst.Args[0].(x86asm.Rel))
	cpu.Regs[REG_CX]--
	if cpu.Regs[REG_CX] != 0 && (cpu.Flags&ZF) != 0 {
		cpu.jumprel(rel)
	}
}

func (cpu *CPU) loopne(inst x86asm.Inst) {
	rel := int32(inst.Args[0].(x86asm.Rel))
	cpu.Regs[REG_CX]--
	if cpu.Regs[REG_CX] != 0 && (cpu.Flags&ZF) == 0 {
		cpu.jumprel(rel)
	}
}

func (cpu *CPU) call(inst x86asm.Inst) {
	cpu.Push(cpu.Ip)
	switch a := inst.Args[0].(type) {
	case x86asm.Rel:
		cpu.jumprel(int32(a))
	default:
		v := cpu.Rmm(inst.Args[0], inst)
		cpu.jumprel(int32(v))
	}
}

// FAR Call (9A and FF/3)
func (cpu *CPU) lcall(inst x86asm.Inst) {
	cpu.Push(cpu.Sregs[SREG_CS])
	cpu.Push(cpu.Ip)
	switch a := inst.Args[0].(type) {
	case x86asm.Imm:
		cpu.Sregs[SREG_CS] = uint16(a)
		ip := uint16(inst.Args[1].(x86asm.Imm))
		cpu.Ip = ip
	case x86asm.Mem:
		seg := int(cpu.segForEA(a, inst))
		off := int(cpu.effectiveAddress(a))
		cpu.Ip = cpu.Mem.Mem16(seg, off)
		cpu.Sregs[SREG_CS] = cpu.Mem.Mem16(seg, off+2)
	default:
		panic(fmt.Sprintf("unknown type for a: %v", a))
	}
}

func (cpu *CPU) inc(inst x86asm.Inst) {
	r := cpu.Rmm(inst.Args[0], inst)
	cpu.PutRmm(inst.Args[0], inst, r+1)
}

func (cpu *CPU) dec(inst x86asm.Inst) {
	r := cpu.Rmm(inst.Args[0], inst)
	cpu.PutRmm(inst.Args[0], inst, r-1)
}

func (cpu *CPU) scasbImpl(dst, src uint) {
	// This is 8bit call, so clear the top half.
	src &= 0x00ff
	res := dst - src
	cpu.SetFlagsSub(res, src, dst, 8)
}

func (cpu *CPU) scasb(inst x86asm.Inst) {
	dst := cpu.Rmm(inst.Args[0], inst)

	if inst.Prefix[0] == 0 {
		src := cpu.Rmm(inst.Args[1], inst)
		cpu.scasbImpl(dst, src)
		return
	}
	seg := int(cpu.Sregs[SREG_ES])
	for _, p := range inst.Prefix {
		switch p {
		case x86asm.PrefixREP:
			for cx := cpu.Reg(x86asm.CX); cx > 0; cx-- {
				di := int(cpu.Reg(x86asm.DI))
				src := uint(cpu.Mem.Mem8(seg, di))
				cpu.scasbImpl(dst, src)
				if cpu.Flags&ZF == 0 {
					cpu.Regs[REG_CX] = uint16(cx)
					break
				}
				if cpu.Flags&DF != 0 {
					cpu.PutReg(x86asm.DI, uint(di-1))
				} else {
					cpu.PutReg(x86asm.DI, uint(di+1))
				}
			}
			cpu.Regs[REG_CX] = 0
		case x86asm.PrefixREPN:
			for cx := cpu.Reg(x86asm.CX); cx > 0; cx-- {
				di := int(cpu.Reg(x86asm.DI))
				src := uint(cpu.Mem.Mem8(seg, di))
				cpu.scasbImpl(dst, src)
				if cpu.Flags&ZF != 0 {
					cpu.Regs[REG_CX] = uint16(cx)
					break
				}
				if cpu.Flags&DF != 0 {
					cpu.PutReg(x86asm.DI, uint(di-1))
				} else {
					cpu.PutReg(x86asm.DI, uint(di+1))
				}
			}
			cpu.Regs[REG_CX] = 0
		default:
			return
		}
	}
}

func (cpu *CPU) cmp(inst x86asm.Inst) {
	dst := cpu.Rmm(inst.Args[0], inst)
	src := cpu.Rmm(inst.Args[1], inst)
	bits := instbits(inst)
	log.V(3).Infof("             CMP: [dst: %04x] [src: %04x] [bits: %d]\n", dst, src, bits)
	res := dst - src
	cpu.SetFlagsSub(res, src, dst, bits)
}

func (cpu *CPU) cmpsOne(inst x86asm.Inst, bits int) {
	// TODO add rep/repn support
	dst := cpu.Rmm(inst.Args[0], inst)
	src := cpu.Rmm(inst.Args[1], inst)
	log.V(3).Infof("             cmpsb: [dst: %04x] [src: %04x] [bits: %d]\n", dst, src, bits)
	res := dst - src
	cpu.SetFlagsSub(res, src, dst, bits)
}

func (cpu *CPU) cmpsRep(inst x86asm.Inst, bits int) {
	step := bits / 8
	if cpu.Flags&DF != 0 {
		step = -step
	}
	for _, p := range inst.Prefix {
		switch p {
		case x86asm.PrefixREP:
			for cx := cpu.Reg(x86asm.CX); cx > 0; cx-- {
				di := int(cpu.Reg(x86asm.DI))
				si := int(cpu.Reg(x86asm.SI))
				cpu.cmpsOne(inst, bits)
				if cpu.Flags&ZF == 0 {
					cpu.Regs[REG_CX] = uint16(cx)
					return
				}
				cpu.PutReg(x86asm.DI, uint(di+step))
				cpu.PutReg(x86asm.SI, uint(si+step))
			}
			cpu.Regs[REG_CX] = 0
			return
		case x86asm.PrefixREPN:
			for cx := cpu.Reg(x86asm.CX); cx > 0; cx-- {
				di := int(cpu.Reg(x86asm.DI))
				si := int(cpu.Reg(x86asm.SI))
				cpu.cmpsOne(inst, bits)
				if cpu.Flags&ZF != 0 {
					cpu.Regs[REG_CX] = uint16(cx)
					return
				}
				cpu.PutReg(x86asm.DI, uint(di+step))
				cpu.PutReg(x86asm.SI, uint(si+step))
			}
			cpu.Regs[REG_CX] = 0
			return
		default:
			panic("Unknonw prefix")
		}
	}
}

func (cpu *CPU) cmpsb(inst x86asm.Inst) {
	if inst.Prefix[0] == 0 {
		cpu.cmpsOne(inst, 8)
		return
	}
	cpu.cmpsRep(inst, 8)
}

func (cpu *CPU) cmpsw(inst x86asm.Inst) {
	if inst.Prefix[0] == 0 {
		cpu.cmpsOne(inst, 16)
		return
	}
	cpu.cmpsRep(inst, 16)
}

func (cpu *CPU) cwd(inst x86asm.Inst) {
	dst := cpu.Regs[REG_AX]
	if (dst & 0x8000) != 0 {
		cpu.Regs[REG_DX] = 0xFFFF
	} else {
		cpu.Regs[REG_DX] = 0x0000
	}
}

func (cpu *CPU) ret(inst x86asm.Inst) {
	cpu.Ip = cpu.Pop()
	if cpu.Ip <= 0x100 {
		log.Infof("             IP <= 0x100: [%04X:%04X]\n\n", cpu.Sregs[SREG_CS], cpu.Ip)
	}
	if inst.Args[0] != nil {
		imm := inst.Args[0].(x86asm.Imm)
		cpu.Regs[REG_SP] = cpu.Regs[REG_SP] + uint16(imm)
	}
}

func (cpu *CPU) lret(inst x86asm.Inst) {
	cpu.Ip = cpu.Pop()
	cs := cpu.Pop()
	cpu.Sregs[SREG_CS] = cs
	if inst.Args[0] != nil {
		imm := inst.Args[0].(x86asm.Imm)
		cpu.Regs[REG_SP] = cpu.Regs[REG_SP] + uint16(imm)
	}
}

func (cpu *CPU) push(inst x86asm.Inst) {
	r := cpu.Rmm(inst.Args[0], inst)
	cpu.Push(uint16(r))
}

func (cpu *CPU) pop(inst x86asm.Inst) {
	r := cpu.Pop()
	cpu.PutRmm(inst.Args[0], inst, uint(r))
}

func (cpu *CPU) pushf(inst x86asm.Inst) {
	cpu.Push(cpu.Flags)
}

func (cpu *CPU) popf(inst x86asm.Inst) {
	cpu.Flags = cpu.Pop()
}

func (cpu *CPU) les(inst x86asm.Inst) {
	m := cpu.Rmm(inst.Args[1], inst)
	cpu.PutRmm(inst.Args[0], inst, m)

	a := inst.Args[1].(x86asm.Mem)
	seg := int(cpu.segForEA(a, inst))
	off := int(cpu.effectiveAddress(a)) + 2
	cpu.Sregs[SREG_ES] = cpu.Mem.Mem16(seg, off)
}

func (cpu *CPU) lds(inst x86asm.Inst) {
	m := cpu.Rmm(inst.Args[1], inst)
	cpu.PutRmm(inst.Args[0], inst, m)

	a := inst.Args[1].(x86asm.Mem)
	seg := int(cpu.segForEA(a, inst))
	off := int(cpu.effectiveAddress(a)) + 2
	cpu.Sregs[SREG_DS] = cpu.Mem.Mem16(seg, off)
}

func (cpu *CPU) cld(inst x86asm.Inst) {
	cpu.Flags &^= DF
}

func (cpu *CPU) std(inst x86asm.Inst) {
	cpu.Flags |= DF
}

func (cpu *CPU) clc(inst x86asm.Inst) {
	cpu.Flags &^= CF
}

func (cpu *CPU) stc(inst x86asm.Inst) {
	cpu.Flags |= CF
}

func (cpu *CPU) or(inst x86asm.Inst) {
	bits := instbits(inst)
	dest := cpu.Rmm(inst.Args[0], inst)
	src := cpu.Rmm(inst.Args[1], inst)
	dest |= src
	cpu.PutRmm(inst.Args[0], inst, dest)
	cpu.Flags &^= (OF | CF)
	cpu.SetFlagsZSP(dest, bits)
}

func (cpu *CPU) and(inst x86asm.Inst) {
	bits := instbits(inst)
	dest := cpu.Rmm(inst.Args[0], inst)
	src := cpu.Rmm(inst.Args[1], inst)
	dest &= src
	cpu.PutRmm(inst.Args[0], inst, dest)
	cpu.Flags &^= (OF | CF)
	cpu.SetFlagsZSP(dest, bits)
}

func (cpu *CPU) xchg(inst x86asm.Inst) {
	dest := cpu.Rmm(inst.Args[0], inst)
	src := cpu.Rmm(inst.Args[1], inst)
	cpu.PutRmm(inst.Args[0], inst, src)
	cpu.PutRmm(inst.Args[1], inst, dest)
}

func (cpu *CPU) xor(inst x86asm.Inst) {
	bits := instbits(inst)
	dest := cpu.Rmm(inst.Args[0], inst)
	src := cpu.Rmm(inst.Args[1], inst)
	dest ^= src
	cpu.PutRmm(inst.Args[0], inst, dest)
	cpu.Flags &^= (OF | CF)
	cpu.SetFlagsZSP(dest, bits)
}

func (cpu *CPU) shl(inst x86asm.Inst) {
	cpu.Flags &^= (OF | CF)
	bits := instbits(inst)
	dest := cpu.Rmm(inst.Args[0], inst)
	src := cpu.Rmm(inst.Args[1], inst)
	if src == 0x1 && dest&(1<<bits) != 0 {
		cpu.Flags |= CF
	}
	dest = dest << src
	cpu.PutRmm(inst.Args[0], inst, dest)
	// TODO: Set OF/CF right
	cpu.SetFlagsZSP(dest, bits)
}

func (cpu *CPU) sar(inst x86asm.Inst) {
	cpu.Flags &^= (OF | CF)

	dest := cpu.Rmm(inst.Args[0], inst)
	src := cpu.Rmm(inst.Args[1], inst)
	cf := (dest >> (src - 1) & 0x1) != 0
	bits := instbits(inst)
	sb := dest & (1 << (bits - 1))
	dest = dest>>src | sb
	if cf {
		cpu.Flags |= CF
	}
	cpu.PutRmm(inst.Args[0], inst, dest)
	cpu.SetFlagsZSP(dest, bits)
}

func (cpu *CPU) shr(inst x86asm.Inst) {
	cpu.Flags &^= (OF | CF)

	dest := cpu.Rmm(inst.Args[0], inst)
	src := cpu.Rmm(inst.Args[1], inst)
	cf := (dest >> (src - 1) & 0x1) != 0
	bits := instbits(inst)
	dest = dest >> src
	if cf {
		cpu.Flags |= CF
	}
	cpu.PutRmm(inst.Args[0], inst, dest)
	cpu.SetFlagsZSP(dest, bits)
}

func (cpu *CPU) neg(inst x86asm.Inst) {
	cpu.Flags &^= (AF | OF | CF)

	dest := cpu.Rmm(inst.Args[0], inst)

	bits := instbits(inst)
	sb := dest & (1 << bits)
	highb := sb >> 1
	dest = sb - dest

	if dest != 0 {
		cpu.Flags |= CF
	}
	if dest&highb != 0 {
		cpu.Flags |= OF
	}
	if ((dest ^ (sb - dest)) & 0x10) != 0 {
		cpu.Flags |= AF
	}
	cpu.PutRmm(inst.Args[0], inst, dest)
	cpu.SetFlagsZSP(dest, bits)
}

func (cpu *CPU) cli(inst x86asm.Inst) {
	cpu.Flags &^= IF
}

func (cpu *CPU) sti(inst x86asm.Inst) {
	cpu.Flags |= IF
}

func (cpu *CPU) stosb(inst x86asm.Inst) {
	es := int(cpu.Reg(x86asm.ES))
	v := uint8(cpu.Reg(x86asm.AL))
	if inst.Prefix[0] == x86asm.PrefixREP {
		for cx := cpu.Reg(x86asm.CX); cx > 0; cx-- {
			cpu.Mem.PutMem8(es, int(cpu.Regs[REG_DI]), v)
			if cpu.Flags&DF != 0 {
				cpu.Regs[REG_DI]--
			} else {
				cpu.Regs[REG_DI]++
			}
		}
		cpu.Regs[REG_CX] = 0
	}
	cpu.Mem.PutMem8(es, int(cpu.Regs[REG_DI]), v)
	if cpu.Flags&DF != 0 {
		cpu.Regs[REG_DI]--
	} else {
		cpu.Regs[REG_DI]++
	}
}

func (cpu *CPU) stosw(inst x86asm.Inst) {
	es := int(cpu.Reg(x86asm.ES))
	v := uint16(cpu.Reg(x86asm.AX))
	if inst.Prefix[0] == x86asm.PrefixREP {
		for cx := cpu.Reg(x86asm.CX); cx > 0; cx-- {
			cpu.Mem.PutMem16(es, int(cpu.Regs[REG_DI]), v)
			if cpu.Flags&DF != 0 {
				cpu.Regs[REG_DI] -= 2
			} else {
				cpu.Regs[REG_DI] += 2
			}
		}
		cpu.Regs[REG_CX] = 0
	}
	cpu.Mem.PutMem16(es, int(cpu.Regs[REG_DI]), v)
	if cpu.Flags&DF != 0 {
		cpu.Regs[REG_DI] -= 2
	} else {
		cpu.Regs[REG_DI] += 2
	}
}

func (cpu *CPU) lodsb(inst x86asm.Inst) {
	ds := int(cpu.Reg(x86asm.DS))
	si := int(cpu.Reg(x86asm.SI))
	v := uint(cpu.Mem.Mem8(ds, si))
	cpu.PutReg(x86asm.AL, v)
	if cpu.Flags&DF != 0 {
		cpu.Regs[REG_SI] -= 1
	} else {
		cpu.Regs[REG_SI] += 1
	}
}

func (cpu *CPU) lodsw(inst x86asm.Inst) {
	ds := int(cpu.Reg(x86asm.DS))
	si := int(cpu.Reg(x86asm.SI))
	v := uint(cpu.Mem.Mem16(ds, si))
	cpu.PutReg(x86asm.AX, v)
	if cpu.Flags&DF != 0 {
		cpu.Regs[REG_SI] -= 2
	} else {
		cpu.Regs[REG_SI] += 2
	}
}

func (cpu *CPU) nop(inst x86asm.Inst) {

}

func (cpu *CPU) not(inst x86asm.Inst) {
	dest := cpu.Rmm(inst.Args[0], inst)
	dest = ^dest
	cpu.PutRmm(inst.Args[0], inst, dest)
}

func (cpu *CPU) test(inst x86asm.Inst) {
	bits := instbits(inst)
	dest := cpu.Rmm(inst.Args[0], inst)
	src := cpu.Rmm(inst.Args[1], inst)
	dest &= src
	cpu.Flags &^= (OF | CF)
	cpu.SetFlagsZSP(dest, bits)
}

func init() {
	opcodes[x86asm.ADD] = (*CPU).add
	opcodes[x86asm.AND] = (*CPU).and
	opcodes[x86asm.CALL] = (*CPU).call
	opcodes[x86asm.CLC] = (*CPU).clc
	opcodes[x86asm.CLD] = (*CPU).cld
	opcodes[x86asm.CLI] = (*CPU).cli
	opcodes[x86asm.CMP] = (*CPU).cmp
	opcodes[x86asm.CMPSB] = (*CPU).cmpsb
	opcodes[x86asm.CMPSW] = (*CPU).cmpsw
	opcodes[x86asm.CWD] = (*CPU).cwd
	opcodes[x86asm.DEC] = (*CPU).dec
	opcodes[x86asm.INC] = (*CPU).inc
	opcodes[x86asm.INT] = (*CPU).intr
	opcodes[x86asm.IRET] = (*CPU).iret
	opcodes[x86asm.JA] = (*CPU).jc
	opcodes[x86asm.JAE] = (*CPU).jc
	opcodes[x86asm.JB] = (*CPU).jc
	opcodes[x86asm.JBE] = (*CPU).jc
	opcodes[x86asm.JCXZ] = (*CPU).jc
	opcodes[x86asm.JE] = (*CPU).jc
	opcodes[x86asm.JG] = (*CPU).jc
	opcodes[x86asm.JGE] = (*CPU).jc
	opcodes[x86asm.JL] = (*CPU).jc
	opcodes[x86asm.JLE] = (*CPU).jc
	opcodes[x86asm.JMP] = (*CPU).jmp
	opcodes[x86asm.JNE] = (*CPU).jc
	opcodes[x86asm.JNO] = (*CPU).jc
	opcodes[x86asm.JNP] = (*CPU).jc
	opcodes[x86asm.JNS] = (*CPU).jc
	opcodes[x86asm.JO] = (*CPU).jc
	opcodes[x86asm.JP] = (*CPU).jc
	opcodes[x86asm.JS] = (*CPU).jc
	opcodes[x86asm.LCALL] = (*CPU).lcall
	opcodes[x86asm.LJMP] = (*CPU).ljmp
	opcodes[x86asm.LDS] = (*CPU).lds
	opcodes[x86asm.LEA] = (*CPU).lea
	opcodes[x86asm.LES] = (*CPU).les
	opcodes[x86asm.LODSB] = (*CPU).lodsb
	opcodes[x86asm.LODSW] = (*CPU).lodsw
	opcodes[x86asm.LOOP] = (*CPU).loop
	opcodes[x86asm.LOOPE] = (*CPU).loope
	opcodes[x86asm.LOOPNE] = (*CPU).loopne
	opcodes[x86asm.LRET] = (*CPU).lret
	opcodes[x86asm.MOV] = (*CPU).mov
	opcodes[x86asm.NEG] = (*CPU).neg
	opcodes[x86asm.NOP] = (*CPU).nop
	opcodes[x86asm.NOT] = (*CPU).not
	opcodes[x86asm.OR] = (*CPU).or
	opcodes[x86asm.POP] = (*CPU).pop
	opcodes[x86asm.POPF] = (*CPU).popf
	opcodes[x86asm.PUSH] = (*CPU).push
	opcodes[x86asm.PUSHF] = (*CPU).pushf
	opcodes[x86asm.RET] = (*CPU).ret
	opcodes[x86asm.SCASB] = (*CPU).scasb
	opcodes[x86asm.SHL] = (*CPU).shl
	opcodes[x86asm.SAR] = (*CPU).sar
	opcodes[x86asm.SHR] = (*CPU).shr
	opcodes[x86asm.STC] = (*CPU).stc
	opcodes[x86asm.STD] = (*CPU).std
	opcodes[x86asm.STI] = (*CPU).sti
	opcodes[x86asm.SUB] = (*CPU).sub
	opcodes[x86asm.STOSB] = (*CPU).stosb
	opcodes[x86asm.STOSW] = (*CPU).stosw
	opcodes[x86asm.TEST] = (*CPU).test
	opcodes[x86asm.XCHG] = (*CPU).xchg
	opcodes[x86asm.XOR] = (*CPU).xor
}

type CPU struct {
	Mem   *Memory
	Regs  [maxRegs]uint16
	Sregs [maxSegs]uint16
	Flags uint16
	Ip    uint16

	running  bool
	Intrs    map[int]func(*CPU, int)
	debugger Debugger
}

func NewCpu(size int) *CPU {
	m := NewMemory(size)
	return &CPU{Mem: m, Intrs: make(map[int]func(*CPU, int)), running: true}
}

// returns a register as an unisgned int
// useful when handling operations.
func (cpu *CPU) Reg(reg x86asm.Reg) uint {
	r := regnums[reg]
	switch reg {
	case x86asm.AL, x86asm.BL, x86asm.CL, x86asm.DL:
		return uint(cpu.Regs[r] & 0xFF)
	case x86asm.AH, x86asm.BH, x86asm.CH, x86asm.DH:
		return uint((cpu.Regs[r] >> 8) & 0xFF)
	case x86asm.AX, x86asm.BX, x86asm.CX, x86asm.DX, x86asm.BP, x86asm.SI, x86asm.SP, x86asm.DI:
		return uint(cpu.Regs[r])
	case x86asm.CS, x86asm.SS, x86asm.DS, x86asm.ES:
		return uint(cpu.Sregs[r])
	case x86asm.IP:
		return uint(cpu.Ip)
	default:
		log.Fatalf("Unexpected Reg: %s ", reg)
		return 0
	}
}

// Puts a value into a register, either 8 or 16 bytes
func (cpu *CPU) PutReg(reg x86asm.Reg, val uint) {
	r := regnums[reg]
	switch reg {
	case x86asm.AL, x86asm.BL, x86asm.CL, x86asm.DL:
		cpu.Regs[r] &= 0xFF00
		cpu.Regs[r] |= uint16(val) & 0x00FF
	case x86asm.AH, x86asm.BH, x86asm.CH, x86asm.DH:
		cpu.Regs[r] &= 0x00FF
		cpu.Regs[r] |= uint16(val) << 8
	case x86asm.AX, x86asm.BX, x86asm.CX, x86asm.DX, x86asm.BP, x86asm.SI, x86asm.SP, x86asm.DI:
		cpu.Regs[r] = uint16(val)
	case x86asm.CS, x86asm.SS, x86asm.DS, x86asm.ES:
		cpu.Sregs[r] = uint16(val)
	case x86asm.IP:
		cpu.Ip = uint16(val)
	default:
		log.Fatalf("Unexpected PutReg: %s ", reg)
	}
}

func (cpu *CPU) segForEA(a x86asm.Mem, inst x86asm.Inst) uint16 {
	// Check for predix overrides
	if inst.Prefix[0] != 0 {
		for _, p := range inst.Prefix {
			p &= 0x00FF
			switch p {
			case x86asm.PrefixCS:
				return cpu.Sregs[SREG_CS]
			case x86asm.PrefixDS:
				return cpu.Sregs[SREG_DS]
			case x86asm.PrefixES:
				return cpu.Sregs[SREG_ES]
			case x86asm.PrefixSS:
				return cpu.Sregs[SREG_SS]
			}
		}
	}

	segnum := SREG_DS
	if a.Base == x86asm.BP || a.Index == x86asm.BP {
		segnum = SREG_SS
	}
	if int64(a.Segment) != 0 {
		segnum = regnums[a.Segment]
	}
	return cpu.Sregs[segnum]
}

// size of instruction. 1 (8 bit) or 2 (16-bit)
func instbits(inst x86asm.Inst) int {
	if inst.MemBytes > 0 {
		return inst.MemBytes * 8
	}
	if inst.Args[0] == nil {
		log.Fatal("Instance had no args and no memsize: ", inst)
		return 0
	}
	a := inst.Args[0]
	switch a := a.(type) {
	case x86asm.Reg:
		switch a {
		case x86asm.AX, x86asm.BX, x86asm.CX, x86asm.DX:
			return 16
		case x86asm.SS, x86asm.CS, x86asm.DS, x86asm.ES:
			return 16
		case x86asm.BP, x86asm.SP, x86asm.SI, x86asm.DI:
			return 16
		default:
			return 8
		}
	default:
		log.Fatalf("Unknown type for Bytes: %s", inst)
		return 0
	}
}

func (cpu *CPU) Rmm(a x86asm.Arg, inst x86asm.Inst) uint {
	switch a := a.(type) {
	case x86asm.Reg:
		return cpu.Reg(a)
	case x86asm.Imm:
		return uint(int64(a & 0xFFFF))
	case x86asm.Mem:
		seg := int(cpu.segForEA(a, inst))
		off := int(cpu.effectiveAddress(a))
		if inst.MemBytes == 1 {
			return uint(cpu.Mem.Mem8(seg, off))
		}
		return uint(cpu.Mem.Mem16(seg, off))
	default:
		log.Fatalf("Unknown type for Rmm8: %T", a)
		return 0
	}
}

func (cpu *CPU) PutRmm(a x86asm.Arg, inst x86asm.Inst, val uint) {
	switch a := a.(type) {
	case x86asm.Reg:
		cpu.PutReg(a, val)
	case x86asm.Mem:
		seg := int(cpu.segForEA(a, inst))
		off := int(cpu.effectiveAddress(a))
		if inst.MemBytes == 1 {
			cpu.Mem.PutMem8(seg, int(off), uint8(val&0xFF))
		} else {
			cpu.Mem.PutMem16(seg, int(off), uint16(val&0xFFFF))
		}
	default:
		log.Fatalf("Unknown type for Rmm8: %T", a)
	}
}

func (cpu *CPU) DoInst(inst x86asm.Inst, origIp uint16) bool {
	if log.V(2) {
		prefix := ""
		if inst.Prefix[0] != 0 {
			prefix = fmt.Sprintf("[%v]", inst.Prefix[0])
		}
		disam := cpu.Mem.At(int(cpu.Sregs[SREG_CS]), int(origIp))[:inst.Len]
		// 0E06:004E BE0010            MOV     SI,1000
		log.V(2).Infof("%04X:%04X %-18s %s%s\n",
			cpu.Sregs[SREG_CS], origIp, hex.EncodeToString(disam), prefix, inst)
	}
	log.V(4).Infof("I: %s, 2: %s [DS: %d][MB: %d]\n", inst.Args[0], inst.Args[1],
		inst.DataSize, inst.MemBytes)

	switch inst.Opcode >> 24 {
	case 0x36:
		log.Fatalf("Whut?")
		// TO DO SOMETHING

	default:
		if opcodes[inst.Op] == nil {
			return false
		}
		opcodes[inst.Op](cpu, inst)
	}
	// Log regs after the instruction executes
	// AX=0005 BX=FF00 CX=0000 DX=0000 SP=0800 BP=0000 SI=1000 DI=0F18
	// DS=0DF6 ES=0DF6 SS=0F4C CS=0E06 IP=0051 NV UP EI NG NZ NA PE NC
	log.V(2).Infof("AX=%04X BX=%04X CX=%04X DX=%04X SP=%04X BP=%04X SI=%04X DI=%04X \n",
		cpu.Regs[REG_AX], cpu.Regs[REG_BX], cpu.Regs[REG_CX], cpu.Regs[REG_DX],
		cpu.Regs[REG_SP], cpu.Regs[REG_BP], cpu.Regs[REG_SI], cpu.Regs[REG_DI])
	log.V(2).Infof("DS=%04X ES=%04X SS=%04X CS=%04X IP=%04X %s\n",
		cpu.Sregs[SREG_DS], cpu.Sregs[SREG_ES],
		cpu.Sregs[SREG_SS], cpu.Sregs[SREG_CS], cpu.Ip, cpu.FlagsToDebugString())
	return true
}

func (cpu *CPU) Decode() (x86asm.Inst, error) {
	x := cpu.Mem.At(int(cpu.Sregs[SREG_CS]), int(cpu.Ip))
	inst, err := x86asm.Decode(x, 16)
	return inst, err
}

func (cpu *CPU) Halt() {
	cpu.running = false
}

func (cpu *CPU) Run() {
	for cpu.running {
		cpu.RunOnce()
	}
}

func (cpu *CPU) EnableDebugger() {
	cpu.debugger = NewDebuggerBackend(cpu)
}

func (cpu *CPU) RunOnce() bool {
	inst, err := cpu.Decode()
	cs := int(cpu.Sregs[SREG_CS])
	ip := int(cpu.Ip)

	if err != nil {
		log.Fatalf("Error decoding instruction at %d", cpu.Ip)
		return false
	}

	if cpu.debugger != nil {
		cpu.debugger.Step()
	}
	origIp := cpu.Ip
	cpu.Ip += uint16(inst.Len)
	res := cpu.DoInst(inst, origIp)
	if !res {
		ms := cpu.Mem.At(cs, ip)[:inst.Len]
		opcodestr := hex.EncodeToString(ms)
		log.Warningf("Unhandled OpCode: [%s] %s [%x]: '%s'\n", opcodestr, inst.Op, inst.Opcode, inst)
	}
	return res
}
