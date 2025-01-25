package go86

import (
	"encoding/hex"
	"fmt"

	log "github.com/golang/glog"
	"golang.org/x/arch/x86/x86asm"
)

// Debugger is an interface for debugging the CPU.
type Debugger interface {
	Step() bool
	Intr() bool
}

// Represents the state of an 8086 CPU.
type CPU struct {
	// Pointer to the system
	Mem *Memory
	// CPU Flags (overflow, zero, etc)
	Flags Flags
	// Current instruction pointer to execute relative to the code segment (CS).
	Ip uint16
	// CPU registers
	Regs *Registers
	// Currently executing instruction
	Inst *Inst

	Running bool
	// Interrupt map for interrupts which do not exist as 8086 code contained
	// within the CPU's memory.
	Intrs    map[int]func(*CPU, int)
	Debugger Debugger
}

func NewCpu(size int) *CPU {
	m := NewMemory(size)
	log.Infof("NewCPU Memory Size : %d\n", size)
	return &CPU{Mem: m, Intrs: make(map[int]func(*CPU, int)),
		Regs:    &Registers{},
		Running: true}
}

func CpuString(c *CPU) string {
	l1 := fmt.Sprintf("AX=%04X BX=%04X CX=%04X DX=%04X SP=%04X BP=%04X SI=%04X DI=%04X",
		c.Regs.GetReg16(AX), c.Regs.GetReg16(BX),
		c.Regs.GetReg16(CX), c.Regs.GetReg16(DX),
		c.Regs.GetReg16(SP), c.Regs.GetReg16(BP),
		c.Regs.GetReg16(SI), c.Regs.GetReg16(DI))
	l2 := fmt.Sprintf("DS=%04X ES=%04X SS=%04X CS=%04X IP=%04X %s",
		c.Regs.DS(), c.Regs.ES(),
		c.Regs.SS(), c.Regs.CS(), c.Ip, c.Flags.DoxBoxDebugString())

	return fmt.Sprintf("%s\n%s", l1, l2)
}

func (cpu *CPU) verboseLogState(origIp uint) error {
	opcodes := cpu.Mem.At(cpu.Regs.CS(), origIp)[:cpu.Inst.Len]

	inst, err := x86asm.Decode(opcodes, 16)
	if err != nil {
		log.V(4).Infof("Error decoding at offset %v\n", err)
		return err
	}
	disasm := inst.String()

	log.V(4).Infof("[%04X:%04X]: [%-8s] %-20s\n%s\n",
		cpu.Regs.CS(),
		origIp,
		hex.EncodeToString(opcodes), disasm,
		CpuString(cpu))
	return nil
}

func (cpu *CPU) Run() {
	for cpu.Running {
		origIp := uint(cpu.Ip)

		err := cpu.RunOnce()
		if err != nil {
			log.Warningf("Error running CPU: %v\n", err)
			// should we halt here at some point?
			// cpu.Running = false
		}

		if log.V(4) {
			cpu.verboseLogState(origIp)
		}
	}
}

func (cpu *CPU) Halt() {
	log.Warningln("HLT instruction executed")
	cpu.Running = false
}

func (cpu *CPU) HandleGrpOne(inst *Inst, left, right Operand) error {
	if err := inst.FetchModRM(); err != nil {
		return err
	}
	// modrm.Reg is an opcode extension
	switch inst.ModRM.Reg {
	case 0:
		return cpu.add(inst, left, right)
	case 1:
		return cpu.or(inst, left, right)
	case 4:
		return cpu.and(inst, left, right)
	case 5:
		return cpu.sub(inst, left, right)
	case 6:
		return cpu.xor(inst, left, right)
	case 7:
		return cpu.cmp(inst, left, right)
	default:
		return fmt.Errorf("unhandled GRP1 opcode: %x", inst.ModRM.Reg)
	}
}

func (cpu *CPU) HandleGrpTwo(inst *Inst, left, right Operand) error {
	if err := inst.FetchModRM(); err != nil {
		return err
	}
	// modrm.Reg is an opcode extension
	switch inst.ModRM.Reg {
	/*
		case 0:
			return cpu.rol(left, right)
		case 1:
			return cpu.ror(left, right)
		case 2:
			return cpu.rcl(left, right)
		case 3:
			return cpu.rcr(left, right)
	*/
	case 4:
		return cpu.shl(inst, left, right)
	case 5:
		return cpu.shr(inst, left, right)
	case 7:
		return cpu.sar(inst, left, right)
	default:
		return fmt.Errorf("unhandled GRP2 opcode: %x", inst.ModRM.Reg)
	}
}

func (cpu *CPU) HandleGrpThree(inst *Inst, left, right Operand) error {
	if err := inst.FetchModRM(); err != nil {
		return err
	}
	// modrm.Reg is an opcode extension
	switch inst.ModRM.Reg {
	case 0:
		return cpu.test(inst, left, right)
	case 2:
		return cpu.not(inst, left)
	case 3:
		return cpu.neg(inst, left)
	case 4:
		return cpu.mul(inst, left)
	case 5:
		return cpu.imul(inst, left)
	case 6:
		return cpu.div(inst, left)
	case 7:
		return cpu.idiv(inst, left)
	default:
		return fmt.Errorf("unhandled GRP1 opcode: %x", inst.ModRM.Reg)
	}
}

// FE only has 0 and 1 and is 8 bit
func (cpu *CPU) HandleFE(inst *Inst) error {
	if err := inst.FetchModRM(); err != nil {
		return err
	}
	// modrm.Reg is an opcode extension
	switch inst.ModRM.Reg {
	case 0:
		val := inst.ModRM.GetRm8(cpu, inst)
		inst.ModRM.SetRm8(cpu, inst, val+1)
		return nil
	case 1:
		val := inst.ModRM.GetRm8(cpu, inst)
		inst.ModRM.SetRm8(cpu, inst, val-1)
		return nil
	default:
		return fmt.Errorf("unhandled FE opcode: %x", inst.ModRM.Reg)
	}
}

// Ev is the only operand here
func (cpu *CPU) HandleFF(inst *Inst) error {
	if err := inst.FetchModRM(); err != nil {
		return err
	}
	// modrm.Reg is an opcode extension
	switch inst.ModRM.Reg {
	case 0:
		val := inst.ModRM.GetRm16(cpu, inst)
		inst.ModRM.SetRm16(cpu, inst, val+1)
		return nil
	case 1:
		val := inst.ModRM.GetRm16(cpu, inst)
		inst.ModRM.SetRm16(cpu, inst, val-1)
		return nil
	case 2:
		return cpu.callNearAbsIndirect(inst)
	case 3:
		return cpu.callFarAbsIndirect(inst)
	case 4:
		return cpu.jmpNearAbsIndirect(inst)
	case 5:
		return cpu.jmpFarAbsIndirect(inst)
	case 6:
		val := inst.ModRM.GetRm16(cpu, inst)
		cpu.Regs.Push16(cpu.Mem, uint16(val))
		return nil
	default:
		return fmt.Errorf("unhandled FF opcode: %x", inst.ModRM.Reg)
	}
}

// RunOnce executes a single instruction.
func (cpu *CPU) RunOnce() error {
	// Clean the previous instruction (and modrm byte) if it was set on the last iteration
	cpu.Inst = nil

	cs := cpu.Regs.CS()
	ip := uint(cpu.Ip)

	var err error
	if cpu.Inst, err = Decode(cpu.Mem.At(cs, ip), cpu); err != nil {
		return fmt.Errorf("failed to decode instruction at CS:IP %04x:%04x: %v", cs, ip, err)
	}

	if cpu.Debugger != nil {
		cpu.Debugger.Step()
	}
	// Increment IP *after* the debugger
	cpu.Ip += uint16(cpu.Inst.Len)

	switch cpu.Inst.OpCode {

	// ADD - Add
	case 0x00, 0x01, 0x02, 0x03, 0x04, 0x05:
		op := StandardOperands[cpu.Inst.OpCode-0x00]
		return cpu.add(cpu.Inst, op.Left, op.Right)

	// PUSH/POP ES
	case 0x06:
		cpu.Regs.Push16(cpu.Mem, uint16(cpu.Regs.ES()))
	case 0x07:
		val := cpu.Regs.Pop16(cpu.Mem)
		cpu.Regs.SetSeg16(ES, uint(val))

	// OR - Logical Inclusive OR
	case 0x08, 0x09, 0x0A, 0x0B, 0x0C, 0x0D:
		op := StandardOperands[cpu.Inst.OpCode-0x08]
		return cpu.or(cpu.Inst, op.Left, op.Right)

	case 0x0E:
		cpu.Regs.Push16(cpu.Mem, uint16(cpu.Regs.CS()))

	case 0x16:
		cpu.Regs.Push16(cpu.Mem, uint16(cpu.Regs.SS()))
	case 0x17:
		val := cpu.Regs.Pop16(cpu.Mem)
		cpu.Regs.SetSeg16(SS, uint(val))

	case 0x1E:
		cpu.Regs.Push16(cpu.Mem, uint16(cpu.Regs.DS()))
	case 0x1F:
		val := cpu.Regs.Pop16(cpu.Mem)
		cpu.Regs.SetSeg16(DS, uint(val))

	// AND - Logical AND
	case 0x20, 0x21, 0x22, 0x23, 0x24, 0x25:
		op := StandardOperands[cpu.Inst.OpCode-0x20]
		return cpu.and(cpu.Inst, op.Left, op.Right)

	// SUB - Subtract
	case 0x28, 0x29, 0x2A, 0x2B, 0x2C, 0x2D:
		op := StandardOperands[cpu.Inst.OpCode-0x28]
		return cpu.sub(cpu.Inst, op.Left, op.Right)

	// XOR - Logical Exclusive OR
	case 0x30, 0x31, 0x32, 0x33, 0x34, 0x35:
		op := StandardOperands[cpu.Inst.OpCode-0x30]
		return cpu.xor(cpu.Inst, op.Left, op.Right)

		// CMP - Compare
	case 0x38, 0x39, 0x3A, 0x3B, 0x3C, 0x3D:
		op := StandardOperands[cpu.Inst.OpCode-0x38]
		return cpu.cmp(cpu.Inst, op.Left, op.Right)

	// Increment/Decrement registers
	case 0x40, 0x41, 0x42, 0x43, 0x44, 0x45, 0x46, 0x47:
		cpu.Regs.Inc16(Reg(cpu.Inst.OpCode-0x40), 1)
	case 0x48, 0x49, 0x4A, 0x4B, 0x4C, 0x4D, 0x4E, 0x4F:
		cpu.Regs.Dec16(Reg(cpu.Inst.OpCode-0x48), 1)

	// Push and pop registers
	case 0x50, 0x51, 0x52, 0x53, 0x54, 0x55, 0x56, 0x57:
		cpu.Regs.PushReg16(Reg(cpu.Inst.OpCode-0x50), cpu.Mem)
	case 0x58, 0x59, 0x5A, 0x5B, 0x5C, 0x5D, 0x5E, 0x5F:
		cpu.Regs.PopReg16(Reg(cpu.Inst.OpCode-0x58), cpu.Mem)

	case 0x68:
		imm16, err := cpu.Inst.Fetch16()
		if err != nil {
			return err
		}
		cpu.Regs.Push16(cpu.Mem, imm16)
		return nil
	case 0x6A:
		imm8, err := cpu.Inst.Fetch8()
		if err != nil {
			return err
		}
		cpu.Regs.Push16(cpu.Mem, uint16(imm8))
		return nil
	// Jumps
	case 0x70:
		return cpu.jo8(cpu.Inst)
	case 0x71:
		return cpu.jno8(cpu.Inst)
	case 0x72:
		return cpu.jb8(cpu.Inst)
	case 0x73:
		return cpu.jnb8(cpu.Inst)
	case 0x74:
		return cpu.jz8(cpu.Inst)
	case 0x75:
		return cpu.jnz8(cpu.Inst)
	case 0x76:
		return cpu.jbe8(cpu.Inst)
	case 0x77:
		return cpu.ja8(cpu.Inst)
	case 0x78:
		return cpu.js8(cpu.Inst)
	case 0x79:
		return cpu.jns8(cpu.Inst)
	case 0x7A:
		return cpu.jpe8(cpu.Inst)
	case 0x7B:
		return cpu.jpo8(cpu.Inst)
	case 0x7C:
		return cpu.jl8(cpu.Inst)
	case 0x7D:
		return cpu.jge8(cpu.Inst)
	case 0x7E:
		return cpu.jle8(cpu.Inst)
	case 0x7F:
		return cpu.jg8(cpu.Inst)

	case 0x80: // GRP1
		return cpu.HandleGrpOne(cpu.Inst, Eb, Ib)
	case 0x81: // GRP1
		return cpu.HandleGrpOne(cpu.Inst, Ev, Iv)
	case 0x82: // ????
		return fmt.Errorf("unhandled undocumented grp1 OpCode: %x", cpu.Inst.OpCode)
	case 0x83:
		return cpu.HandleGrpOne(cpu.Inst, Ev, Ib)

	// Tests
	case 0x84:
		return cpu.test(cpu.Inst, Eb, Gb)
	case 0x85:
		return cpu.test(cpu.Inst, Ev, Gv)

	// XCHG - Exchange
	case 0x86:
		return cpu.xchg(cpu.Inst, Gb, Eb)
	case 0x87:
		return cpu.xchg(cpu.Inst, Gv, Ev)

	// MOV - Move Gv, Ev and family
	case 0x88:
		return cpu.mov(cpu.Inst, Eb, Gb)
	case 0x89:
		return cpu.mov(cpu.Inst, Ev, Gv)
	case 0x8A:
		return cpu.mov(cpu.Inst, Gb, Eb)
	case 0x8B:
		return cpu.mov(cpu.Inst, Gv, Ev)
	case 0x8C:
		return cpu.mov(cpu.Inst, Ew, Sw)
	case 0x8D:
		return cpu.leaGvM(cpu.Inst)
	case 0x8E:
		return cpu.mov(cpu.Inst, Sw, Ew)
	case 0x8F:
		return cpu.popEv(cpu.Inst)

	case 0x90: // NOP
		return nil

	// XCHG Registers
	case 0x91:
		cpu.xchgRegs(CX, AX)
	case 0x92:
		cpu.xchgRegs(DX, AX)
	case 0x93:
		cpu.xchgRegs(BX, AX)
	case 0x94:
		cpu.xchgRegs(SP, AX)
	case 0x95:
		cpu.xchgRegs(BP, AX)
	case 0x96:
		cpu.xchgRegs(SI, AX)
	case 0x97:
		cpu.xchgRegs(DI, AX)
	case 0x98: // Sign extend AL into AX
		value := uint(int16(int8(cpu.Regs.GetReg8(AL))))
		cpu.Regs.SetReg16(AX, value)
	case 0x99: // Sign extend AX into DX:AX
		value := uint(int32(int16(cpu.Regs.GetReg16(AX))))
		cpu.Regs.SetReg16(DX, value>>16)
	case 0x9A:
		return cpu.callFar(cpu.Inst)
	case 0x9B:
		return fmt.Errorf("WAIT [0x98] not handled")
	case 0x9C: // PUSHF
		cpu.Regs.Push16(cpu.Mem, uint16(cpu.Flags.Value()))
	case 0x9D: // POPF
		cpu.Flags.ReplaceAllFlags(uint32(cpu.Regs.Pop16(cpu.Mem)))
	case 0x9E: // SAHF
		cpu.Regs.SetReg8(AH, uint(cpu.Flags.Value()&0xff))
	case 0x9F: // LAHF
		flags := cpu.Flags.Value() & 0xff00
		flags |= uint32(cpu.Regs.GetReg8(AH) & 0x00ff)
		cpu.Flags.ReplaceAllFlags(flags)

	// MOVE - moffs
	case 0xA0:
		return cpu.mov(cpu.Inst, RegAL, Ob)
	case 0xA1:
		return cpu.mov(cpu.Inst, RegAX, Ov)
	case 0xA2:
		return cpu.mov(cpu.Inst, Ob, RegAL)
	case 0xA3:
		return cpu.mov(cpu.Inst, Ov, RegAX)

	case 0xA6:
		return cpu.cmps(8)
	case 0xA7:
		return cpu.cmps(16)

	case 0xA8:
		return cpu.test(cpu.Inst, RegAL, Ib)
	case 0xA9:
		return cpu.test(cpu.Inst, RegAX, Iv)

	case 0xAA:
		return cpu.stos(cpu.Inst, 8)
	case 0xAB:
		return cpu.stos(cpu.Inst, 16)
	case 0xAC:
		return cpu.lods(cpu.Inst, 8)
	case 0xAD:
		return cpu.lods(cpu.Inst, 16)

	case 0xAE:
		return cpu.scas(cpu.Inst, 8)
	case 0xAF:
		return cpu.scas(cpu.Inst, 16)

	// Move to register from immediate value
	case 0xB0, 0xB1, 0xB2, 0xB3, 0xB4, 0xB5, 0xB6, 0xB7:
		return cpu.movRegIb(cpu.Inst, Reg8(cpu.Inst.OpCode-0xB0))
	case 0xB8, 0xB9, 0xBA, 0xBB, 0xBC, 0xBD, 0xBE, 0xBF:
		return cpu.movRegIv(cpu.Inst, Reg(cpu.Inst.OpCode-0xB8))

	case 0xC0: // GRP2
		return cpu.HandleGrpTwo(cpu.Inst, Eb, Ib)
	case 0xC1: // GRP2
		return cpu.HandleGrpTwo(cpu.Inst, Ev, Ib)
	case 0xC2:
		imm16, err := cpu.Inst.Fetch16()
		if err != nil {
			return err
		}
		return cpu.retNear(uint(imm16))
	case 0xC3:
		return cpu.retNear(0)

	case 0xC4: // LES
		return cpu.lesLds(cpu.Inst, ES)
	case 0xC5: // LDS
		return cpu.lesLds(cpu.Inst, DS)

	case 0xC6:
		if err := cpu.Inst.FetchModRM(); err != nil {
			return err
		}
		if cpu.Inst.ModRM.Reg != 0 {
			return fmt.Errorf("unexpected reg value for opcode 0xC6: '%x'", cpu.Inst.ModRM.Reg)
		}
		return cpu.mov(cpu.Inst, Eb, Ib)
	case 0xC7:
		if err := cpu.Inst.FetchModRM(); err != nil {
			return err
		}
		if cpu.Inst.ModRM.Reg != 0 {
			return fmt.Errorf("unexpected reg value for opcode 0xC6: '%x'", cpu.Inst.ModRM.Reg)
		}
		return cpu.mov(cpu.Inst, Ev, Iv)
	case 0xCA:
		imm16, err := cpu.Inst.Fetch16()
		if err != nil {
			return err
		}
		return cpu.retFar(uint(imm16))
	case 0xCB:
		return cpu.retFar(0)

	case 0xCC:
		return cpu.int(0x03)

	case 0xCD:
		imm8, err := cpu.Inst.Fetch8()
		if err != nil {
			return err
		}
		return cpu.int(int(imm8))

	case 0xD0: // GRP2
		return cpu.HandleGrpTwo(cpu.Inst, Eb, ValOne)
	case 0xD1: // GRP2
		return cpu.HandleGrpTwo(cpu.Inst, Ev, ValOne)
	case 0xD2: // GRP2
		return cpu.HandleGrpTwo(cpu.Inst, Eb, RegCL)
	case 0xD3: // GRP2
		return cpu.HandleGrpTwo(cpu.Inst, Ev, RegCL)

		// LOOP
	case 0xE0:
		return cpu.loopne(cpu.Inst)
	case 0xE1:
		return cpu.loope(cpu.Inst)
	case 0xE2:
		return cpu.loop(cpu.Inst)

	case 0xE8:
		return cpu.callNear(cpu.Inst)
	// JMP
	case 0xE3:
		return cpu.jcxz(cpu.Inst)
	case 0xE9:
		return cpu.jmprel16(cpu.Inst)
	case 0xEA:
		return cpu.jmpFarAbs(cpu.Inst)
	case 0xEB:
		return cpu.jmprel8(cpu.Inst)

	case 0xF6: // GRP3
		return cpu.HandleGrpThree(cpu.Inst, Eb, Ib)
	case 0xF7: // GRP3
		return cpu.HandleGrpThree(cpu.Inst, Ev, Iv)

	case 0xF8: // CLC
		cpu.Flags.ClearFlag(CarryFlag)
		return nil
	case 0xF9: // STC
		cpu.Flags.SetFlags(CarryFlag)
		return nil

		// CLI - Clear Interrupt Flag
	case 0xFA:
		cpu.Flags.ClearFlag(IF)
		return nil

	// STI - Set Interrupt Flag
	case 0xFB:
		cpu.Flags.SetFlags(IF)
		return nil

	case 0xFC: // CLD
		cpu.Flags.ClearFlag(DirectionFlag)
		return nil
	case 0xFD: // STD
		cpu.Flags.SetFlags(DirectionFlag)
		return nil

	case 0xFE:
		return cpu.HandleFE(cpu.Inst) // Eb
	case 0xFF:
		return cpu.HandleFF(cpu.Inst) // Ev

	default:
		return fmt.Errorf("unhandled OpCode: %x", cpu.Inst.OpCode)
	}
	return nil
}

// Fetch8 reads an 8-bit value from memory at the current instruction pointer
// and increments the instruction pointer by 1.
func (cpu *CPU) Fetch8() (uint8, error) {
	x := cpu.Mem.GetMem8(cpu.Regs.CS(), uint(cpu.Ip))
	cpu.Ip++
	return x, nil
}

// Fetch16 reads a 16-bit value from memory at the current instruction pointer
// and increments the instruction pointer by 2.
func (cpu *CPU) Fetch16() (uint16, error) {
	x := cpu.Mem.GetMem16(cpu.Regs.CS(), uint(cpu.Ip))
	cpu.Ip += 2
	return x, nil
}
