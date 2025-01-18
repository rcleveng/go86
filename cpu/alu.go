package go86

// ADD - Add
func (cpu *CPU) add(leftop, rightop Operand) error {
	left, right, err := ParseTwoOperands(cpu, leftop, rightop)
	if err != nil {
		return err
	}

	result := left + right

	leftop.SetByOperand(cpu, result)
	cpu.Flags.SetFlagsAdd(result, left, right, leftop.Bits())

	return nil
}

// SUB - Subtract

func (cpu *CPU) sub(leftop, rightop Operand) error {
	left, right, err := ParseTwoOperands(cpu, leftop, rightop)
	if err != nil {
		return err
	}

	result := left - right

	leftop.SetByOperand(cpu, result)
	cpu.Flags.SetFlagsSub(result, left, right, leftop.Bits())

	return nil
}
