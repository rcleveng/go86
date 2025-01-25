package go86

// OR - Logical Inclusive OR

func (cpu *CPU) or(inst *Inst, leftop, rightop Operand) error {
	left, right, err := ParseTwoOperands(cpu, inst, leftop, rightop)
	if err != nil {
		return err
	}

	result := left | right

	leftop.SetByOperand(cpu, inst, cpu.Mem, cpu.Regs, result)
	cpu.Flags.SetFlagsZSP(result, leftop.Bits())
	cpu.Flags.ClearFlagsCO()

	return nil
}

// AND - Logical AND

func (cpu *CPU) and(inst *Inst, leftop, rightop Operand) error {
	left, right, err := ParseTwoOperands(cpu, inst, leftop, rightop)
	if err != nil {
		return err
	}

	result := left & right

	leftop.SetByOperand(cpu, inst, cpu.Mem, cpu.Regs, result)
	cpu.Flags.SetFlagsZSP(result, leftop.Bits())
	cpu.Flags.ClearFlagsCO()

	return nil
}

// XOR - Logical Exclusive OR

func (cpu *CPU) xor(inst *Inst, leftop, rightop Operand) error {
	left, right, err := ParseTwoOperands(cpu, inst, leftop, rightop)
	if err != nil {
		return err
	}

	result := left ^ right

	leftop.SetByOperand(cpu, inst, cpu.Mem, cpu.Regs, result)
	cpu.Flags.SetFlagsZSP(result, leftop.Bits())
	cpu.Flags.ClearFlagsCO()

	return nil
}

func (cpu *CPU) compare(left, right uint, bit int) error {
	diff := left - right
	cpu.Flags.SetFlagsSub(diff, left, right, bit)
	return nil
}

// CMP - Compare
func (cpu *CPU) cmp(inst *Inst, leftop, rightop Operand) error {
	left, right, err := ParseTwoOperands(cpu, inst, leftop, rightop)
	if err != nil {
		return err
	}
	diff := left - right
	cpu.Flags.SetFlagsSub(diff, left, right, leftop.Bits())
	return nil
}

// TEST - Logical Compare

func (cpu *CPU) test(inst *Inst, leftop, rightop Operand) error {
	left, right, err := ParseTwoOperands(cpu, inst, leftop, rightop)
	if err != nil {
		return err
	}
	dest := left & right
	cpu.Flags.ClearFlagsCO()
	cpu.Flags.SetFlagsZSP(dest, leftop.Bits())
	return nil
}

// NOT - Logical NOT

func (cpu *CPU) not(inst *Inst, op Operand) error {
	value, err := op.GetByOperand(cpu, inst, cpu.Mem, cpu.Regs)
	if err != nil {
		return err
	}

	result := ^value

	op.SetByOperand(cpu, inst, cpu.Mem, cpu.Regs, result)
	return nil
}

// NEG - Two's Complement Negation

func (cpu *CPU) neg(inst *Inst, op Operand) error {
	value, err := op.GetByOperand(cpu, inst, cpu.Mem, cpu.Regs)
	if err != nil {
		return err
	}

	result := -int(value)

	op.SetByOperand(cpu, inst, cpu.Mem, cpu.Regs, uint(result))
	cpu.Flags.SetFlagsSub(uint(result), 0, value, op.Bits())
	return nil
}
