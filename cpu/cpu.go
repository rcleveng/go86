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

// HandleGrp18 handles the GRP1 opcode extension for
func (cpu *CPU) HandleGrpOne8() error {
	modrm, err := ParseModRM(cpu)
	if err != nil {
		return err
	}
	// modrm.Reg is an opcode extension
	switch modrm.Reg {
	case 0:
		return cpu.addEbIb(modrm)
	case 1:
		return cpu.orEbIb(modrm)
	case 4:
		return cpu.andEbIb(modrm)
	case 5:
		return cpu.subEbIb(modrm)
	case 6:
		return cpu.xorEbIb(modrm)
	case 7:
		return cpu.cmpEbIb(modrm)
	default:
		return fmt.Errorf("unhandled GRP1 opcode: %x", modrm.Reg)
	}
}

func (cpu *CPU) HandleGrpOne16() error {
	modrm, err := ParseModRM(cpu)
	if err != nil {
		return err
	}
	// modrm.Reg is an opcode extension
	switch modrm.Reg {
	case 0:
		return cpu.addEvIv(modrm)
	case 1:
		return cpu.orEvIv(modrm)
	case 4:
		return cpu.andEvIv(modrm)
	case 5:
		return cpu.subEvIv(modrm)
	case 6:
		return cpu.xorEvIv(modrm)
	case 7:
		return cpu.cmpEvIv(modrm)
	default:
		return fmt.Errorf("unhandled GRP1 opcode: %x", modrm.Reg)
	}
}

// RunOnce executes a single instruction.
func (cpu *CPU) RunOnce() error {
	cs := cpu.Regs.CS()
	ip := uint(cpu.Ip)
	opcode := cpu.Mem.Mem8(cs, ip)

	if cpu.Debugger != nil {
		cpu.Debugger.Step()
	}
	cpu.Ip += 1

	switch opcode {

	// ADD - Add
	case 0x00:
		cpu.addEbGb()
	case 0x01:
		cpu.addEvGv()
	case 0x02:
		cpu.addGbEb()
	case 0x03:
		cpu.addGvEv()
	case 0x04:
		cpu.addALIb()
	case 0x05:
		cpu.addAXIv()

	// PUSH/POP ES
	case 0x06:
		cpu.Regs.Push16(cpu.Mem, uint16(cpu.Regs.ES()))
	case 0x07:
		val := cpu.Regs.Pop16(cpu.Mem)
		cpu.Regs.SetSeg16(ES, uint(val))

	// OR - Logical Inclusive OR
	case 0x08:
		cpu.orEbGb()
	case 0x09:
		cpu.orEvGv()
	case 0x0A:
		cpu.orGbEb()
	case 0x0B:
		cpu.orGvEv()
	case 0x0C:
		cpu.orALIb()
	case 0x0D:
		cpu.orAXIv()
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
	case 0x20:
		cpu.andEbGb()
	case 0x21:
		cpu.andEvGv()
	case 0x22:
		cpu.andGbEb()
	case 0x23:
		cpu.andGvEv()
	case 0x24:
		cpu.andALIb()
	case 0x25:
		cpu.andAXIv()

	// SUB - Subtract
	case 0x28:
		cpu.subEbGb()
	case 0x29:
		cpu.subEvGv()
	case 0x2A:
		cpu.subGbEb()
	case 0x2B:
		cpu.subGvEv()
	case 0x2C:
		cpu.subALIb()
	case 0x2D:
		cpu.subAXIv()

	// XOR - Logical Exclusive OR
	// OR - Logical Inclusive OR
	case 0x30:
		cpu.xorEbGb()
	case 0x31:
		cpu.xorEvGv()
	case 0x32:
		cpu.xorGbEb()
	case 0x33:
		cpu.xorGvEv()
	case 0x34:
		cpu.xorALIb()
	case 0x35:
		cpu.xorAXIv()

	// CMP - Compare
	case 0x38:
		cpu.cmpEbGb()
	case 0x39:
		cpu.cmpEvGv()
	case 0x3A:
		cpu.cmpGbEb()
	case 0x3B:
		cpu.cmpGvEv()
	case 0x3C:
		cpu.cmpALIb()
	case 0x3D:
		cpu.cmpAXIv()

	// Increment/Decrement registers
	case 0x40, 0x41, 0x42, 0x43, 0x44, 0x45, 0x46, 0x47:
		cpu.Regs.Inc16(Reg(opcode-0x40), 1)
	case 0x48, 0x49, 0x4A, 0x4B, 0x4C, 0x4D, 0x4E, 0x4F:
		cpu.Regs.Dec16(Reg(opcode-0x48), 1)

	// Push and pop registers
	case 0x50, 0x51, 0x52, 0x53, 0x54, 0x55, 0x56, 0x57:
		cpu.Regs.PushReg16(Reg(opcode-0x50), cpu.Mem)
	case 0x58, 0x59, 0x5A, 0x5B, 0x5C, 0x5D, 0x5E, 0x5F:
		cpu.Regs.PopReg16(Reg(opcode-0x50), cpu.Mem)

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

	case 0x80, 0x82: // GRP1
		return cpu.HandleGrpOne8()
	case 0x81: // GRP1
		return cpu.HandleGrpOne16()
	case 0x83:
		return fmt.Errorf("unhandled undocumented grp1 OpCode: %x", opcode)

	case 0x90: // NOP
		return nil

	default:
		return fmt.Errorf("unhandled OpCode: %x", opcode)
	}
	return nil
}

// Fetch8 reads an 8-bit value from memory at the current instruction pointer
// and increments the instruction pointer by 1.
func (cpu *CPU) Fetch8() (uint8, error) {
	x := cpu.Mem.Mem8(cpu.Regs.CS(), uint(cpu.Ip))
	cpu.Ip++
	return x, nil
}

// Fetch16 reads a 16-bit value from memory at the current instruction pointer
// and increments the instruction pointer by 2.
func (cpu *CPU) Fetch16() (uint16, error) {
	x := cpu.Mem.Mem16(cpu.Regs.CS(), uint(cpu.Ip))
	cpu.Ip += 2
	return x, nil
}
