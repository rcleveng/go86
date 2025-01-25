package go86

// xchg - Exchange

func (cpu *CPU) xchg(inst *Inst, leftop, rightop Operand) error {
	left, right, err := ParseTwoOperands(cpu, inst, leftop, rightop)
	if err != nil {
		return err
	}

	leftop.SetByOperand(cpu, inst, cpu.Mem, cpu.Regs, right)
	rightop.SetByOperand(cpu, inst, cpu.Mem, cpu.Regs, left)
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

func (cpu *CPU) mov(inst *Inst, leftop, rightop Operand) error {
	_, right, err := ParseTwoOperands(cpu, inst, leftop, rightop)
	if err != nil {
		return err
	}

	leftop.SetByOperand(cpu, inst, cpu.Mem, cpu.Regs, right)
	return nil
}

// leaGvM - Load Effective Address
func (cpu *CPU) leaGvM(inst *Inst) error {
	if err := inst.FetchModRM(); err != nil {
		return err
	}
	offset := inst.ModRM.effectiveAddressOffset16(cpu)
	cpu.Regs.SetReg16(Reg(inst.ModRM.Reg), offset)
	return nil
}

func (cpu *CPU) lesLds(inst *Inst, esds SReg) error {
	if err := inst.FetchModRM(); err != nil {
		return err
	}

	seg, off, err := inst.ModRM.GetMemoryLocation(cpu, inst)
	if err != nil {
		return err
	}

	newoff := cpu.Mem.GetMem16(seg, off)
	newseg := cpu.Mem.GetMem16(seg, off+2)
	cpu.Regs.SetSeg16(esds, uint(newseg))
	inst.ModRM.SetR16(cpu, uint(newoff))
	return nil
}

// popEv - Pop
func (cpu *CPU) popEv(inst *Inst) error {
	if err := inst.FetchModRM(); err != nil {
		return err
	}
	value := cpu.Regs.Pop16(cpu.Mem)
	inst.ModRM.SetRm16(cpu, inst, uint(value))
	return nil
}

// Move - Move AL, Ib and family

// movALIb - Move
func (cpu *CPU) movRegIb(inst *Inst, reg Reg8) error {
	value, err := inst.Fetch8()
	if err != nil {
		return err
	}
	cpu.Regs.SetReg8(reg, uint(value))
	return nil
}

func (cpu *CPU) movRegIv(inst *Inst, reg Reg) error {
	value, err := inst.Fetch16()
	if err != nil {
		return err
	}
	cpu.Regs.SetReg16(reg, uint(value))
	return nil
}
