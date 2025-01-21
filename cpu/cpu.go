package go86

import (
	"fmt"

	log "github.com/golang/glog"
)

// Debugger is an interface for debugging the CPU.
type Debugger interface {
	Step() bool
	Intr() bool
}

// CpuInstructionReader is an interface for reading CPU instructions.
type CpuInstructionReader interface {
	Fetch8() (uint8, error)
	Fetch16() (uint16, error)
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
	Regs Registers
	// Current ModRM byte (if exists)
	ModRM *ModRM
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
	return &CPU{Mem: m, Intrs: make(map[int]func(*CPU, int)), Running: true}
}

func (cpu *CPU) Run() {
	for cpu.Running {
		err := cpu.RunOnce()
		if err != nil {
			log.Warningf("Error running CPU: %v\n", err)
			// should we halt here at some point?
			// cpu.Running = false
		}
	}
}

func (cpu *CPU) HandleGrpOne(left, right Operand) error {
	if err := cpu.FetchModRM(); err != nil {
		return err
	}
	// modrm.Reg is an opcode extension
	switch cpu.ModRM.Reg {
	case 0:
		return cpu.add(left, right)
	case 1:
		return cpu.or(left, right)
	case 4:
		return cpu.and(left, right)
	case 5:
		return cpu.sub(left, right)
	case 6:
		return cpu.xor(left, right)
	case 7:
		return cpu.cmp(left, right)
	default:
		return fmt.Errorf("unhandled GRP1 opcode: %x", cpu.ModRM.Reg)
	}
}

func (cpu *CPU) HandleGrpTwo(left, right Operand) error {
	if err := cpu.FetchModRM(); err != nil {
		return err
	}
	// modrm.Reg is an opcode extension
	switch cpu.ModRM.Reg {
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
		return cpu.shl(left, right)
	case 5:
		return cpu.shr(left, right)
	case 7:
		return cpu.sar(left, right)
	default:
		return fmt.Errorf("unhandled GRP2 opcode: %x", cpu.ModRM.Reg)
	}
}

func (cpu *CPU) HandleGrpThree(left, right Operand) error {
	if err := cpu.FetchModRM(); err != nil {
		return err
	}
	// modrm.Reg is an opcode extension
	switch cpu.ModRM.Reg {
	case 0:
		return cpu.test(left, right)
	case 2:
		return cpu.not(left)
	case 3:
		return cpu.neg(left)
	case 4:
		return cpu.mul(left)
	case 5:
		return cpu.imul(left)
	case 6:
		return cpu.div(left)
	case 7:
		return cpu.idiv(left)
	default:
		return fmt.Errorf("unhandled GRP1 opcode: %x", cpu.ModRM.Reg)
	}
}

// FE only has 0 and 1 and is 8 bit
func (cpu *CPU) HandleFE() error {
	if err := cpu.FetchModRM(); err != nil {
		return err
	}
	// modrm.Reg is an opcode extension
	switch cpu.ModRM.Reg {
	case 0:
		val := cpu.ModRM.GetRm8(cpu)
		cpu.ModRM.SetRm8(cpu, val+1)
		return nil
	case 1:
		val := cpu.ModRM.GetRm8(cpu)
		cpu.ModRM.SetRm8(cpu, val-1)
		return nil
	default:
		return fmt.Errorf("unhandled FE opcode: %x", cpu.ModRM.Reg)
	}
}

// Ev is the only operand here
func (cpu *CPU) HandleFF() error {
	if err := cpu.FetchModRM(); err != nil {
		return err
	}
	// modrm.Reg is an opcode extension
	switch cpu.ModRM.Reg {
	case 0:
		val := cpu.ModRM.GetRm16(cpu)
		cpu.ModRM.SetRm16(cpu, val+1)
		return nil
	case 1:
		val := cpu.ModRM.GetRm16(cpu)
		cpu.ModRM.SetRm16(cpu, val-1)
		return nil
	case 2:
		return cpu.callNearAbsIndirect()
	case 3:
		return cpu.callFarAbsIndirect()
	case 4:
		return cpu.jmpNearAbsIndirect()
	case 5:
		return cpu.jmpFarAbsIndirect()
	case 6:
		val := cpu.ModRM.GetRm16(cpu)
		cpu.Regs.Push16(cpu.Mem, uint16(val))
		return nil
	default:
		return fmt.Errorf("unhandled FF opcode: %x", cpu.ModRM.Reg)
	}
}

func (cpu *CPU) FetchModRM() error {
	// TODO - rename this back to just ParseModRM or somethibng
	// maybe fetchandparse?
	if cpu.ModRM != nil {
		// Initially panic here to clean it up and then just guard against it
		// once we're sure it's safe or to make the api easier
		// to use
		panic("modrm already parsed")
	}
	b, err := cpu.Fetch8()
	if err != nil {
		return err
	}

	modrm, err := NewModRM(cpu, b)
	if err != nil {
		return err
	}
	cpu.ModRM = modrm
	return nil
}

// RunOnce executes a single instruction.
func (cpu *CPU) RunOnce() error {
	// Clean the modrm byte if it was set on the last iteration
	cpu.ModRM = nil
	cpu.Inst = nil

	cs := cpu.Regs.CS()
	ip := uint(cpu.Ip)
	inst, err := Decode(cpu.Mem.At(cs, ip))
	if err != nil {
		return fmt.Errorf("failed to decode instruction at CS:IP %04x:%04x: %v", cs, ip, err)
	}
	cpu.Inst = inst

	if cpu.Debugger != nil {
		cpu.Debugger.Step()
	}
	// Increment IP *after* the debugger
	cpu.Ip += uint16(inst.Len)

	switch inst.OpCode {

	// ADD - Add
	case 0x00, 0x01, 0x02, 0x03, 0x04, 0x05:
		op := StandardOperands[inst.OpCode-0x00]
		return cpu.add(op.Left, op.Right)

	// PUSH/POP ES
	case 0x06:
		cpu.Regs.Push16(cpu.Mem, uint16(cpu.Regs.ES()))
	case 0x07:
		val := cpu.Regs.Pop16(cpu.Mem)
		cpu.Regs.SetSeg16(ES, uint(val))

	// OR - Logical Inclusive OR
	case 0x08, 0x09, 0x0A, 0x0B, 0x0C, 0x0D:
		op := StandardOperands[inst.OpCode-0x08]
		return cpu.or(op.Left, op.Right)

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
		op := StandardOperands[inst.OpCode-0x20]
		return cpu.and(op.Left, op.Right)

	// SUB - Subtract
	case 0x28, 0x29, 0x2A, 0x2B, 0x2C, 0x2D:
		op := StandardOperands[inst.OpCode-0x28]
		return cpu.sub(op.Left, op.Right)

	// XOR - Logical Exclusive OR
	case 0x30, 0x31, 0x32, 0x33, 0x34, 0x35:
		op := StandardOperands[inst.OpCode-0x30]
		return cpu.xor(op.Left, op.Right)

		// CMP - Compare
	case 0x38, 0x39, 0x3A, 0x3B, 0x3C, 0x3D:
		op := StandardOperands[inst.OpCode-0x38]
		return cpu.cmp(op.Left, op.Right)

	// Increment/Decrement registers
	case 0x40, 0x41, 0x42, 0x43, 0x44, 0x45, 0x46, 0x47:
		cpu.Regs.Inc16(Reg(inst.OpCode-0x40), 1)
	case 0x48, 0x49, 0x4A, 0x4B, 0x4C, 0x4D, 0x4E, 0x4F:
		cpu.Regs.Dec16(Reg(inst.OpCode-0x48), 1)

	// Push and pop registers
	case 0x50, 0x51, 0x52, 0x53, 0x54, 0x55, 0x56, 0x57:
		cpu.Regs.PushReg16(Reg(inst.OpCode-0x50), cpu.Mem)
	case 0x58, 0x59, 0x5A, 0x5B, 0x5C, 0x5D, 0x5E, 0x5F:
		cpu.Regs.PopReg16(Reg(inst.OpCode-0x58), cpu.Mem)

	// Jumps
	case 0x70:
		return cpu.jo8()
	case 0x71:
		return cpu.jno8()
	case 0x72:
		return cpu.jb8()
	case 0x73:
		return cpu.jnb8()
	case 0x74:
		return cpu.jz8()
	case 0x75:
		return cpu.jnz8()
	case 0x76:
		return cpu.jbe8()
	case 0x77:
		return cpu.ja8()
	case 0x78:
		return cpu.js8()
	case 0x79:
		return cpu.jns8()
	case 0x7A:
		return cpu.jpe8()
	case 0x7B:
		return cpu.jpo8()
	case 0x7C:
		return cpu.jl8()
	case 0x7D:
		return cpu.jge8()
	case 0x7E:
		return cpu.jle8()
	case 0x7F:
		return cpu.jg8()

	case 0x80: // GRP1
		return cpu.HandleGrpOne(Eb, Ib)
	case 0x81: // GRP1
		return cpu.HandleGrpOne(Ev, Iv)
	case 0x82: // ????
		return fmt.Errorf("unhandled undocumented grp1 OpCode: %x", inst.OpCode)
	case 0x83:
		return cpu.HandleGrpOne(Ev, Ib)

	// Tests
	case 0x84:
		return cpu.test(Eb, Gb)
	case 0x85:
		return cpu.test(Ev, Gv)

	// XCHG - Exchange
	case 0x86:
		return cpu.xchg(Gb, Eb)
	case 0x87:
		return cpu.xchg(Gv, Ev)

	// MOV - Move Gv, Ev and family
	case 0x88:
		return cpu.mov(Eb, Gb)
	case 0x89:
		return cpu.mov(Ev, Gv)
	case 0x8A:
		return cpu.mov(Gb, Eb)
	case 0x8B:
		return cpu.mov(Gv, Ev)
	case 0x8C:
		return cpu.mov(Ew, Sw)
	case 0x8D:
		return cpu.leaGvM()
	case 0x8E:
		return cpu.mov(Sw, Ew)
	case 0x8F:
		return cpu.popEv()

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
		return cpu.callFar()
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
		return cpu.mov(RegAL, Ob)
	case 0xA1:
		return cpu.mov(RegAX, Ov)
	case 0xA2:
		return cpu.mov(Ob, RegAL)
	case 0xA3:
		return cpu.mov(Ov, RegAX)

	// Move to register from immediate value
	case 0xB0, 0xB1, 0xB2, 0xB3, 0xB4, 0xB5, 0xB6, 0xB7:
		return cpu.movRegIb(Reg8(inst.OpCode - 0xB0))
	case 0xB8, 0xB9, 0xBA, 0xBB, 0xBC, 0xBD, 0xBE, 0xBF:
		return cpu.movRegIv(Reg(inst.OpCode - 0xB8))

	case 0xC0: // GRP2
		return cpu.HandleGrpTwo(Eb, Ib)
	case 0xC1: // GRP2
		return cpu.HandleGrpTwo(Ev, Ib)
	case 0xC6:
		if err := cpu.FetchModRM(); err != nil {
			return err
		}
		if cpu.ModRM.Reg != 0 {
			return fmt.Errorf("unexpected reg value for opcode 0xC6: '%x'", cpu.ModRM.Reg)
		}
		return cpu.mov(Eb, Ib)
	case 0xC7:
		if err := cpu.FetchModRM(); err != nil {
			return err
		}
		if cpu.ModRM.Reg != 0 {
			return fmt.Errorf("unexpected reg value for opcode 0xC6: '%x'", cpu.ModRM.Reg)
		}
		return cpu.mov(Ev, Iv)

	case 0xD0: // GRP2
		return cpu.HandleGrpTwo(Eb, ValOne)
	case 0xD1: // GRP2
		return cpu.HandleGrpTwo(Ev, ValOne)
	case 0xD2: // GRP2
		return cpu.HandleGrpTwo(Eb, RegCL)
	case 0xD3: // GRP2
		return cpu.HandleGrpTwo(Ev, RegCL)

	// JMP
	case 0xE3:
		return cpu.jcxz()
	case 0xE9:
		return cpu.jmprel16()
	case 0xEA:
		return cpu.jmpFarAbs()
	case 0xEB:
		return cpu.jmprel8()

	case 0xF6: // GRP3
		return cpu.HandleGrpThree(Eb, Ib)
	case 0xF7: // GRP3
		return cpu.HandleGrpThree(Ev, Iv)

	// CLI - Clear Interrupt Flag
	case 0xFA:
		cpu.Flags.ClearFlag(IF)
		return nil

	// STI - Set Interrupt Flag
	case 0xFB:
		cpu.Flags.SetFlags(IF)
		return nil

	case 0xFE:
		return cpu.HandleFE() // Eb
	case 0xFF:
		return cpu.HandleFF() // Ev

	default:
		return fmt.Errorf("unhandled OpCode: %x", inst.OpCode)
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
