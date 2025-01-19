package go86

import "fmt"

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

// SHR - Logical Shift Right
func (cpu *CPU) shr(leftop, rightop Operand) error {
	left, right, err := ParseTwoOperands(cpu, leftop, rightop)
	if err != nil {
		return err
	}

	cf := (left>>(right-1))&1 == 1
	result := left >> right

	leftop.SetByOperand(cpu, result)
	cpu.Flags.SetFlagsZSP(result)
	cpu.Flags.SetFlagIf(CarryFlag, cf)

	return nil
}

// SAR - Arithmetic Shift Right
func (cpu *CPU) sar(leftop, rightop Operand) error {
	left, right, err := ParseTwoOperands(cpu, leftop, rightop)
	if err != nil {
		return err
	}

	cf := (left>>(right-1))&1 == 1
	// >> preserves signs on int types in golang
	var result uint
	switch leftop.Bits() {
	case 8:
		r := int8(left) >> right
		r1 := uint8(r)
		result = uint(r1)
	case 16:
		r := int16(left) >> right
		result = uint(r)
	default:
		return fmt.Errorf("incorrect bits for left operand: %#v", leftop)
	}

	leftop.SetByOperand(cpu, uint(result))
	cpu.Flags.SetFlagsZSP(uint(result))
	cpu.Flags.SetFlagIf(CarryFlag, cf)

	return nil
}

// SHL - Logical Shift Left
func (cpu *CPU) shl(leftop, rightop Operand) error {
	left, right, err := ParseTwoOperands(cpu, leftop, rightop)
	if err != nil {
		return err
	}

	cf := (left<<(right-1))&(1<<(leftop.Bits()-1)) != 0
	result := left << right

	leftop.SetByOperand(cpu, result)
	cpu.Flags.SetFlagsZSP(result)
	cpu.Flags.SetFlagIf(CarryFlag, cf)

	return nil
}

// MUL - Unsigned Multiply
func (cpu *CPU) mul(leftop, rightop Operand) error {
	left, right, err := ParseTwoOperands(cpu, leftop, rightop)
	if err != nil {
		return err
	}

	result := left * right
	cf := result>>(leftop.Bits()) != 0

	leftop.SetByOperand(cpu, result)
	cpu.Flags.SetFlagsZSP(result)
	// Set CF and OF if the result cannot fit in the destination operand
	cpu.Flags.SetFlagIf(CarryFlag|OverflowFlag, cf)

	return nil
}

// IMUL - Signed Multiply
func (cpu *CPU) imul(leftop, rightop Operand) error {
	left, right, err := ParseTwoOperands(cpu, leftop, rightop)
	if err != nil {
		return err
	}

	result := int(left) * int(right)
	cf := uint(result)>>(leftop.Bits()) != 0

	leftop.SetByOperand(cpu, uint(result))
	cpu.Flags.SetFlagsZSP(uint(result))
	// Set CF and OF if the result cannot fit in the destination operand
	cpu.Flags.SetFlagIf(CarryFlag|OverflowFlag, cf)

	return nil
}

// DIV - Unsigned Divide
func (cpu *CPU) div(leftop, rightop Operand) error {
	left, right, err := ParseTwoOperands(cpu, leftop, rightop)
	if err != nil {
		return err
	}

	if right == 0 {
		return fmt.Errorf("division by zero")
	}

	quotient := left / right
	remainder := left % right
	leftop.SetByOperand(cpu, quotient)

	switch leftop.Bits() {
	case 8: // 8 bit uses AL for quotient, AH for remainder
		cpu.Regs.SetReg8(AH, uint(remainder))
	case 16: // 8 bit uses AX for quotient, DX for remainder
		cpu.Regs.SetReg16(DX, uint(remainder))
	}
	cpu.Flags.SetFlagsZSP(quotient)
	cpu.Flags.ClearFlagsCO()

	return nil
}

// IDIV - Signed Divide
func (cpu *CPU) idiv(leftop, rightop Operand) error {
	left, right, err := ParseTwoOperands(cpu, leftop, rightop)
	if err != nil {
		return err
	}

	if right == 0 {
		// TODO - Handle with processor exception once implemented
		return fmt.Errorf("division by zero")
	}

	quotient := int(left) / int(right)
	remainder := int(left) % int(right)
	leftop.SetByOperand(cpu, uint(quotient))

	switch leftop.Bits() {
	case 8: // 8 bit uses AL for quotient, AH for remainder
		cpu.Regs.SetReg8(AH, uint(remainder))
	case 16: // 16 bit uses AX for quotient, DX for remainder
		cpu.Regs.SetReg16(DX, uint(remainder))
	}
	cpu.Flags.SetFlagsZSP(uint(quotient))
	cpu.Flags.ClearFlagsCO()

	return nil
}
