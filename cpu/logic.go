package go86

// OR - Logical Inclusive OR

func (cpu *CPU) or(leftop, rightop Operand) error {
	left, right, err := ParseTwoOperands(cpu, leftop, rightop)
	if err != nil {
		return err
	}

	result := left | right

	leftop.SetByOperand(cpu, result)
	cpu.Flags.SetFlagsZSP(result)
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
	cpu.Flags.SetFlagsZSP(result)
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
	cpu.Flags.SetFlagsZSP(result)
	cpu.Flags.ClearFlagsCO()

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
	cpu.Flags.ClearFlagsCO()
	cpu.Flags.SetFlagsZSP(left & right)
	return nil
}
