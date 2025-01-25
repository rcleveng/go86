package go86

import "fmt"

// ADD - Add
func (cpu *CPU) add(inst *Inst, leftop, rightop Operand) error {
	left, right, err := ParseTwoOperands(cpu, inst, leftop, rightop)
	if err != nil {
		return err
	}

	result := left + right

	leftop.SetByOperand(cpu, inst, cpu.Mem, cpu.Regs, result)
	cpu.Flags.SetFlagsAdd(result, left, right, leftop.Bits())

	return nil
}

// SUB - Subtract
func (cpu *CPU) sub(inst *Inst, leftop, rightop Operand) error {
	left, right, err := ParseTwoOperands(cpu, inst, leftop, rightop)
	if err != nil {
		return err
	}

	result := left - right

	leftop.SetByOperand(cpu, inst, cpu.Mem, cpu.Regs, result)
	cpu.Flags.SetFlagsSub(result, left, right, leftop.Bits())

	return nil
}

// SHR - Logical Shift Right
func (cpu *CPU) shr(inst *Inst, leftop, rightop Operand) error {
	left, right, err := ParseTwoOperands(cpu, inst, leftop, rightop)
	if err != nil {
		return err
	}

	cf := (left>>(right-1))&1 == 1
	result := left >> right

	leftop.SetByOperand(cpu, inst, cpu.Mem, cpu.Regs, result)
	cpu.Flags.SetFlagsZSP(result, leftop.Bits())
	cpu.Flags.SetFlagIf(CarryFlag, cf)

	return nil
}

// SAR - Arithmetic Shift Right
func (cpu *CPU) sar(inst *Inst, leftop, rightop Operand) error {
	left, right, err := ParseTwoOperands(cpu, inst, leftop, rightop)
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

	leftop.SetByOperand(cpu, inst, cpu.Mem, cpu.Regs, result)
	cpu.Flags.SetFlagsZSP(result, leftop.Bits())
	cpu.Flags.SetFlagIf(CarryFlag, cf)

	return nil
}

// SHL - Logical Shift Left
func (cpu *CPU) shl(inst *Inst, leftop, rightop Operand) error {
	left, right, err := ParseTwoOperands(cpu, inst, leftop, rightop)
	if err != nil {
		return err
	}

	cf := (left<<(right-1))&(1<<(leftop.Bits()-1)) != 0
	result := left << right

	leftop.SetByOperand(cpu, inst, cpu.Mem, cpu.Regs, result)
	cpu.Flags.SetFlagsZSP(result, leftop.Bits())
	cpu.Flags.SetFlagIf(CarryFlag, cf)

	return nil
}

// MUL - Unsigned Multiply
// always called for grp3
func (cpu *CPU) mul(inst *Inst, op Operand) error {
	left := cpu.Regs.GetReg16(AX)
	if op.Bits() == 8 {
		left &= 0x00ff
	}
	right, err := op.GetByOperand(cpu, inst, cpu.Mem, cpu.Regs)
	if err != nil {
		return err
	}

	result := left * right

	// if bits is 16, then DX gets the top 16 bits of result.
	cpu.Regs.SetReg16(AX, result)
	var cf bool
	switch op.Bits() {
	case 8:
		cf = (result & 0xff00) != 0
	case 16:
		cf = (result & 0xffff0000) != 0
		dx := (result & 0xffff0000) >> 16
		cpu.Regs.SetReg16(DX, dx)
	}
	// If AH or DX contain values, both the carry flag
	// and overflow flags should be set.  ZSP are undefined.
	cpu.Flags.SetFlagIf(CarryFlag|OverflowFlag, cf)
	return nil
}

// IMUL - Signed Multiply
func (cpu *CPU) imul(inst *Inst, op Operand) error {
	left := cpu.Regs.GetReg16(AX)
	if op.Bits() == 8 {
		left &= 0x00ff
	}
	right, err := op.GetByOperand(cpu, inst, cpu.Mem, cpu.Regs)
	if err != nil {
		return err
	}

	result := int(left) * int(right)

	// if bits is 16, then DX gets the top 16 bits of result.
	// If AH or DX contain values, both the carry flag
	// and overflow flags should be set.  ZSP are undefined.
	cpu.Regs.SetReg16(AX, uint(result))
	switch op.Bits() {
	case 8:
		cf := (result & 0xff00) != 0
		cpu.Flags.SetFlagIf(CarryFlag|OverflowFlag, cf)
	case 16:
		cf := (result & 0xffff0000) != 0
		cpu.Flags.SetFlagIf(CarryFlag|OverflowFlag, cf)
		dx := uint(result&0xffff0000) >> 16
		cpu.Regs.SetReg16(DX, dx)
	}
	return nil
}

// DIV - Unsigned Divide
func (cpu *CPU) div(inst *Inst, op Operand) error {
	left := cpu.Regs.GetReg16(AX)
	if op.Bits() == 8 {
		left &= 0x00ff
	}
	right, err := op.GetByOperand(cpu, inst, cpu.Mem, cpu.Regs)
	if err != nil {
		return err
	}
	if right == 0 {
		return fmt.Errorf("division by zero")
	}

	quotient := left / right
	remainder := left % right

	switch op.Bits() {
	case 8: // 8 bit uses AL for quotient, AH for remainder
		cpu.Regs.SetReg8(AL, quotient)
		cpu.Regs.SetReg8(AH, uint(remainder))
	case 16: // 8 bit uses AX for quotient, DX for remainder
		cpu.Regs.SetReg16(AX, quotient)
		cpu.Regs.SetReg16(DX, uint(remainder))
	}
	return nil
}

// IDIV - Signed Divide
func (cpu *CPU) idiv(inst *Inst, op Operand) error {
	left := cpu.Regs.GetReg16(AX)
	if op.Bits() == 8 {
		left &= 0x00ff
	}
	right, err := op.GetByOperand(cpu, inst, cpu.Mem, cpu.Regs)
	if err != nil {
		return err
	}
	if right == 0 {
		return fmt.Errorf("division by zero")
	}

	quotient := int(left) / int(right)
	remainder := int(left) % int(right)

	switch op.Bits() {
	case 8: // 8 bit uses AL for quotient, AH for remainder
		cpu.Regs.SetReg8(AL, uint(quotient))
		cpu.Regs.SetReg8(AH, uint(remainder))
	case 16: // 8 bit uses AX for quotient, DX for remainder
		cpu.Regs.SetReg16(AX, uint(quotient))
		cpu.Regs.SetReg16(DX, uint(remainder))
	}
	return nil
}
