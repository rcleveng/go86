package go86

// jump
func (cpu *CPU) jump8(rel int8) {
	if rel < 0 {
		cpu.Ip -= uint16(-rel)
	} else {
		cpu.Ip += uint16(rel)
	}
}

func (cpu *CPU) jumpIfFlag8(flag uint32) error {
	rel, err := cpu.Fetch8()
	if err != nil {
		return err
	}
	if cpu.Flags.IsEnabled(flag) {
		cpu.jump8(int8(rel))
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
		cpu.jump8(int8(rel))
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
		cpu.jump8(int8(rel))
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
		cpu.jump8(int8(rel))
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
		cpu.jump8(int8(rel))
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
		cpu.jump8(int8(rel))
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
		cpu.jump8(int8(rel))
	}
	return nil
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
