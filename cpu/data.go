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
	modrm, err := ParseModRM(cpu)
	if err != nil {
		return err
	}
	offset := modrm.effectiveAddressOffset16(cpu)
	cpu.Regs.SetReg16(Reg(modrm.Reg), offset)
	return nil
}

// popEv - Pop
func (cpu *CPU) popEv() error {
	modrm, err := ParseModRM(cpu)
	if err != nil {
		return err
	}
	value := cpu.Regs.Pop16(cpu.Mem)
	modrm.SetRm16(cpu, uint(value))
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
