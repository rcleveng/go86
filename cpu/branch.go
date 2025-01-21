package go86

// JMP REL8
func (cpu *CPU) jmprel8() error {
	urel, err := cpu.Fetch8()
	if err != nil {
		return err
	}
	return cpu.jump8rel(int8(urel))
}

// JMP REL16 (note, this can be a negative displacement)
func (cpu *CPU) jmprel16() error {
	urel, err := cpu.Fetch16()
	if err != nil {
		return err
	}
	return cpu.jump16rel(int16(urel))
}

// JMP far absolute (ptr16:16)
// Example: 0000000E  EA45230000        jmp 0x0:0x2345
func (cpu *CPU) jmpFarAbs() error {
	offset, err := cpu.Fetch16()
	if err != nil {
		return err
	}
	seg, err := cpu.Fetch16()
	if err != nil {
		return err
	}
	cpu.Regs.SetSeg16(CS, uint(seg))
	cpu.Ip = offset
	return nil
}

func (cpu *CPU) jmpNearAbsIndirect() error {
	// near is always 16 bits
	offset := cpu.ModRM.GetRm16(cpu)
	cpu.Ip = uint16(offset)
	return nil
}

func (cpu *CPU) callNearAbsIndirect() error {
	// near is always 16 bits
	offset := cpu.ModRM.GetRm16(cpu)
	cpu.Regs.PushSeg16(CS, cpu.Mem)
	cpu.Regs.Push16(cpu.Mem, cpu.Ip)
	cpu.Ip = uint16(offset)
	return nil
}

func (cpu *CPU) jmpFarAbsIndirect() error {
	segment, offset, err := cpu.ModRM.GetMemoryLocation(cpu)
	if err != nil {
		return err
	}

	off := cpu.Mem.GetMem16(segment, offset)
	seg := cpu.Mem.GetMem16(segment, offset+2)

	cpu.Regs.SetSeg16(CS, uint(seg))
	cpu.Ip = uint16(off)
	return nil
}

func (cpu *CPU) callFarAbsIndirect() error {
	segment, offset, err := cpu.ModRM.GetMemoryLocation(cpu)
	if err != nil {
		return err
	}
	cpu.Regs.PushSeg16(CS, cpu.Mem)
	cpu.Regs.Push16(cpu.Mem, cpu.Ip)

	off := cpu.Mem.GetMem16(segment, offset)
	seg := cpu.Mem.GetMem16(segment, offset+2)

	cpu.Regs.SetSeg16(CS, uint(seg))
	cpu.Ip = uint16(off)
	return nil
}

// jump to a relative 16 bit offset.
func (cpu *CPU) jump8rel(rel int8) error {
	if rel < 0 {
		cpu.Ip -= uint16(-rel)
	} else {
		cpu.Ip += uint16(rel)
	}
	return nil
}

// jump
func (cpu *CPU) jump16rel(rel int16) error {
	if rel < 0 {
		cpu.Ip -= uint16(-rel)
	} else {
		cpu.Ip += uint16(rel)
	}
	return nil
}

func (cpu *CPU) jumpIfFlag8(flag uint32) error {
	rel, err := cpu.Fetch8()
	if err != nil {
		return err
	}
	if cpu.Flags.IsEnabled(flag) {
		return cpu.jump8rel(int8(rel))
	}
	return nil
}

// jumpIfNoFlag8 - jump if not flag
func (cpu *CPU) jumpIfNoFlag8(flag uint32) error {
	rel, err := cpu.Fetch8()
	if err != nil {
		return err
	}
	if !cpu.Flags.IsEnabled(flag) {
		return cpu.jump8rel(int8(rel))
	}
	return nil
}

// jo8 - jump if overflow
func (cpu *CPU) jo8() error {
	return cpu.jumpIfFlag8(OverflowFlag)
}

// jno8 - jump if not overflow
func (cpu *CPU) jno8() error {
	return cpu.jumpIfNoFlag8(OverflowFlag)
}

// jb8 - jump if below
func (cpu *CPU) jb8() error {
	return cpu.jumpIfFlag8(CarryFlag)
}

// jnob8 - jump if not below
func (cpu *CPU) jnb8() error {
	return cpu.jumpIfNoFlag8(CarryFlag)
}

// jz8 - jump if zero
func (cpu *CPU) jz8() error {
	return cpu.jumpIfFlag8(ZeroFlag)
}

// jnz8 - jump if not zero
func (cpu *CPU) jnz8() error {
	return cpu.jumpIfNoFlag8(ZeroFlag)
}

// jbe8 - jump if below or equal
func (cpu *CPU) jbe8() error {
	return cpu.jumpIfFlag8(CarryFlag | ZeroFlag)
}

// ja8 - jump if above
func (cpu *CPU) ja8() error {
	rel, err := cpu.Fetch8()
	if err != nil {
		return err
	}
	if !cpu.Flags.IsEnabled(CarryFlag) && !cpu.Flags.IsEnabled(ZeroFlag) {
		return cpu.jump8rel(int8(rel))
	}
	return nil
}

// js8 - jump if sign
func (cpu *CPU) js8() error {
	return cpu.jumpIfFlag8(SignFlag)
}

// jnos8 - jump if not sign
func (cpu *CPU) jns8() error {
	return cpu.jumpIfNoFlag8(SignFlag)
}

// jpe8 - jump if parity even
func (cpu *CPU) jpe8() error {
	return cpu.jumpIfFlag8(ParityFlag)
}

// jpo8 - jump if parity odd
func (cpu *CPU) jpo8() error {
	return cpu.jumpIfNoFlag8(ParityFlag)
}

// jl8 - jump if less
func (cpu *CPU) jl8() error {
	rel, err := cpu.Fetch8()
	if err != nil {
		return err
	}
	sf := cpu.Flags.IsEnabled(SignFlag)
	of := cpu.Flags.IsEnabled(OverflowFlag)
	if sf != of {
		return cpu.jump8rel(int8(rel))
	}
	return nil
}

// jle8 - jump if less or equal
func (cpu *CPU) jle8() error {
	rel, err := cpu.Fetch8()
	if err != nil {
		return err
	}
	sf := cpu.Flags.IsEnabled(SignFlag)
	of := cpu.Flags.IsEnabled(OverflowFlag)
	zf := cpu.Flags.IsEnabled(ZeroFlag)
	if zf || sf != of {
		return cpu.jump8rel(int8(rel))
	}
	return nil
}

// jg8 - jump if greater
func (cpu *CPU) jg8() error {
	rel, err := cpu.Fetch8()
	if err != nil {
		return err
	}
	sf := cpu.Flags.IsEnabled(SignFlag)
	of := cpu.Flags.IsEnabled(OverflowFlag)
	zf := cpu.Flags.IsEnabled(ZeroFlag)
	if !zf && sf == of {
		return cpu.jump8rel(int8(rel))
	}
	return nil
}

// jge8 - jump if greater or equal
func (cpu *CPU) jge8() error {
	rel, err := cpu.Fetch8()
	if err != nil {
		return err
	}
	sf := cpu.Flags.IsEnabled(SignFlag)
	of := cpu.Flags.IsEnabled(OverflowFlag)
	if sf == of {
		return cpu.jump8rel(int8(rel))
	}
	return nil
}

// callNear
func (cpu *CPU) callNear() error {
	off, err := cpu.Fetch16()
	if err != nil {
		return err
	}
	cpu.Regs.Push16(cpu.Mem, cpu.Ip)
	return cpu.jump16rel(int16(off))
}

// callFar - call far
func (cpu *CPU) callFar() error {
	off, err := cpu.Fetch16()
	if err != nil {
		return err
	}
	seg, err := cpu.Fetch16()
	if err != nil {
		return err
	}
	cpu.Regs.PushSeg16(CS, cpu.Mem)
	cpu.Regs.Push16(cpu.Mem, cpu.Ip)
	cpu.Regs.SetSeg16(CS, uint(seg))
	cpu.Ip = off & 0xffff // mask to 16 bits
	return nil
}

// jcxz - jump if cx is zero
func (cpu *CPU) jcxz() error {
	// Need to fetch the next value before checking CX
	urel, err := cpu.Fetch8()
	if err != nil {
		return err
	}
	if cpu.Regs.GetReg16(CX) == 0 {
		return cpu.jump8rel(int8(urel))
	}
	return nil
}
