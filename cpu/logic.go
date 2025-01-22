package go86

// OR - Logical Inclusive OR

func (cpu *CPU) or(leftop, rightop Operand) error {
	left, right, err := ParseTwoOperands(cpu, leftop, rightop)
	if err != nil {
		return err
	}

	result := left | right

	leftop.SetByOperand(cpu, result)
	cpu.Flags.SetFlagsZSP(result, leftop.Bits())
	cpu.Flags.ClearFlagsCO()

	return nil
}

// AND - Logical AND

func (cpu *CPU) and(leftop, rightop Operand) error {
	left, right, err := ParseTwoOperands(cpu, leftop, rightop)
	if err != nil {
		return err
	}

	result := left & right

	leftop.SetByOperand(cpu, result)
	cpu.Flags.SetFlagsZSP(result, leftop.Bits())
	cpu.Flags.ClearFlagsCO()

	return nil
}

// XOR - Logical Exclusive OR

func (cpu *CPU) xor(leftop, rightop Operand) error {
	left, right, err := ParseTwoOperands(cpu, leftop, rightop)
	if err != nil {
		return err
	}

	result := left ^ right

	leftop.SetByOperand(cpu, result)
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
func (cpu *CPU) cmp(leftop, rightop Operand) error {
	left, right, err := ParseTwoOperands(cpu, leftop, rightop)
	if err != nil {
		return err
	}
	diff := left - right
	cpu.Flags.SetFlagsSub(diff, left, right, leftop.Bits())
	return nil
}

// TEST - Logical Compare

func (cpu *CPU) test(leftop, rightop Operand) error {
	left, right, err := ParseTwoOperands(cpu, leftop, rightop)
	if err != nil {
		return err
	}
	dest := left & right
	cpu.Flags.ClearFlagsCO()
	cpu.Flags.SetFlagsZSP(dest, leftop.Bits())
	return nil
}

// NOT - Logical NOT

func (cpu *CPU) not(op Operand) error {
	value, err := op.GetByOperand(cpu)
	if err != nil {
		return err
	}

	result := ^value

	op.SetByOperand(cpu, result)
	return nil
}

// NEG - Two's Complement Negation

func (cpu *CPU) neg(op Operand) error {
	value, err := op.GetByOperand(cpu)
	if err != nil {
		return err
	}

	result := -int(value)

	op.SetByOperand(cpu, uint(result))
	cpu.Flags.SetFlagsSub(uint(result), 0, value, op.Bits())
	return nil
}
