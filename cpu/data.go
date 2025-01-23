package go86

// xchg - Exchange

func (cpu *CPU) xchg(leftop, rightop Operand) error {
	left, right, err := ParseTwoOperands(cpu, leftop, rightop)
	if err != nil {
		return err
	}

	leftop.SetByOperand(cpu, right)
	rightop.SetByOperand(cpu, left)
	return nil
}

// xchgRegs() - Exchange
func (cpu *CPU) xchgRegs(leftReg Reg, rightReg Reg) {
	left := cpu.Regs.GetReg16(leftReg)
	right := cpu.Regs.GetReg16(rightReg)
	cpu.Regs.SetReg16(leftReg, right)
	cpu.Regs.SetReg16(rightReg, left)
}

// MOVE

func (cpu *CPU) mov(leftop, rightop Operand) error {
	_, right, err := ParseTwoOperands(cpu, leftop, rightop)
	if err != nil {
		return err
	}

	leftop.SetByOperand(cpu, right)
	return nil
}

// leaGvM - Load Effective Address
func (cpu *CPU) leaGvM() error {
	if err := cpu.FetchModRM(); err != nil {
		return err
	}
	offset := cpu.ModRM.effectiveAddressOffset16(cpu)
	cpu.Regs.SetReg16(Reg(cpu.ModRM.Reg), offset)
	return nil
}

func (cpu *CPU) lesLds(esds SReg) error {
	if err := cpu.FetchModRM(); err != nil {
		return err
	}

	seg, off, err := cpu.ModRM.GetMemoryLocation(cpu)
	if err != nil {
		return err
	}

	newoff := cpu.Mem.GetMem16(seg, off)
	newseg := cpu.Mem.GetMem16(seg, off+2)
	cpu.Regs.SetSeg16(esds, uint(newseg))
	cpu.ModRM.SetR16(cpu, uint(newoff))
	return nil
}

// popEv - Pop
func (cpu *CPU) popEv() error {
	if err := cpu.FetchModRM(); err != nil {
		return err
	}
	value := cpu.Regs.Pop16(cpu.Mem)
	cpu.ModRM.SetRm16(cpu, uint(value))
	return nil
}

// Move - Move AL, Ib and family

// movALIb - Move
func (cpu *CPU) movRegIb(reg Reg8) error {
	value, err := cpu.Fetch8()
	if err != nil {
		return err
	}
	cpu.Regs.SetReg8(reg, uint(value))
	return nil
}

func (cpu *CPU) movRegIv(reg Reg) error {
	value, err := cpu.Fetch16()
	if err != nil {
		return err
	}
	cpu.Regs.SetReg16(reg, uint(value))
	return nil
}

// generate LEA
/* func (cpu *CPU) lea() error {
	if err := cpu.FetchModRM(); err != nil {
		return err
	}
	// ignore seg?
	_, off, err := cpu.ModRM.GetMemoryLocation(cpu)
	if err != nil {
		return err
	}
	cpu.ModRM.SetR16(cpu, off)
	return nil
}
*/
