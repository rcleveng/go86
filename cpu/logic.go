package go86

// OR - Logical Inclusive OR

func (cpu *CPU) orEbGb() {
	modrm, err := ParseModRM(cpu)
	if err != nil {
		return
	}
	left := modrm.GetRm8(cpu)
	right := cpu.Regs.GetReg8(Reg8(modrm.Reg))
	modrm.SetRm8(cpu, left|right)
}

func (cpu *CPU) orEvGv() {
	modrm, err := ParseModRM(cpu)
	if err != nil {
		return
	}
	left := modrm.GetRm16(cpu)
	right := cpu.Regs.GetReg16(Reg(modrm.Reg))
	modrm.SetRm16(cpu, left|right)
}

func (cpu *CPU) orGbEb() {
	modrm, err := ParseModRM(cpu)
	if err != nil {
		return
	}
	left := cpu.Regs.GetReg8(Reg8(modrm.Reg))
	right := modrm.GetRm8(cpu)
	modrm.SetR8(cpu, left|right)
}

func (cpu *CPU) orGvEv() {
	modrm, err := ParseModRM(cpu)
	if err != nil {
		return
	}
	left := cpu.Regs.GetReg16(Reg(modrm.Reg))
	right := modrm.GetRm16(cpu)
	modrm.SetR16(cpu, left|right)
}

func (cpu *CPU) orALIb() {
	left := cpu.Regs.GetReg8(AL)
	imm8, err := cpu.Fetch8()
	if err != nil {
		return
	}
	cpu.Regs.SetReg8(AL, uint(uint8(left)|imm8))
}

func (cpu *CPU) orAXIv() {
	left := cpu.Regs.AX()
	imm16, err := cpu.Fetch16()
	if err != nil {
		return
	}
	cpu.Regs.SetReg8(AL, uint(uint16(left)|imm16))
}

func (cpu *CPU) orEvIv(modrm *ModRM) error {
	left := modrm.GetRm16(cpu)
	imm16, err := cpu.Fetch16()
	if err != nil {
		return err
	}
	modrm.SetRm16(cpu, left|uint(imm16))
	return nil
}

func (cpu *CPU) orEbIb(modrm *ModRM) error {
	left := cpu.Regs.GetReg8(Reg8(modrm.Reg))
	imm8, err := cpu.Fetch8()
	if err != nil {
		return err
	}
	modrm.SetR8(cpu, left|uint(imm8))
	return nil
}

// AND - Logical AND

func (cpu *CPU) andEbGb() {
	modrm, err := ParseModRM(cpu)
	if err != nil {
		return
	}
	left := modrm.GetRm8(cpu)
	right := cpu.Regs.GetReg8(Reg8(modrm.Reg))
	modrm.SetRm8(cpu, left&right)
}

func (cpu *CPU) andEvGv() {
	modrm, err := ParseModRM(cpu)
	if err != nil {
		return
	}
	left := modrm.GetRm16(cpu)
	right := cpu.Regs.GetReg16(Reg(modrm.Reg))
	modrm.SetRm16(cpu, left&right)
}

func (cpu *CPU) andGbEb() {
	modrm, err := ParseModRM(cpu)
	if err != nil {
		return
	}
	left := cpu.Regs.GetReg8(Reg8(modrm.Reg))
	right := modrm.GetRm8(cpu)
	modrm.SetR8(cpu, left&right)
}

func (cpu *CPU) andGvEv() {
	modrm, err := ParseModRM(cpu)
	if err != nil {
		return
	}
	left := cpu.Regs.GetReg16(Reg(modrm.Reg))
	right := modrm.GetRm16(cpu)
	modrm.SetR16(cpu, left&right)
}

func (cpu *CPU) andALIb() {
	left := cpu.Regs.GetReg8(AL)
	imm8, err := cpu.Fetch8()
	if err != nil {
		return
	}
	cpu.Regs.SetReg8(AL, uint(uint8(left)&imm8))
}

func (cpu *CPU) andAXIv() {
	left := cpu.Regs.AX()
	imm16, err := cpu.Fetch16()
	if err != nil {
		return
	}
	cpu.Regs.SetReg8(AL, uint(uint16(left)&imm16))
}

func (cpu *CPU) andEvIv(modrm *ModRM) error {
	left := modrm.GetRm16(cpu)
	imm16, err := cpu.Fetch16()
	if err != nil {
		return err
	}
	modrm.SetR16(cpu, left&uint(imm16))
	return nil
}

func (cpu *CPU) andEbIb(modrm *ModRM) error {
	left := modrm.GetRm8(cpu)
	imm8, err := cpu.Fetch8()
	if err != nil {
		return err
	}
	sum := left + uint(imm8)
	cpu.Flags.SetFlagsAdd8(sum, left, uint(imm8))
	modrm.SetRm8(cpu, sum)
	return nil
}

// XOR - Logical Exclusive OR

func (cpu *CPU) xorEbGb() {
	modrm, err := ParseModRM(cpu)
	if err != nil {
		return
	}
	left := modrm.GetRm8(cpu)
	right := cpu.Regs.GetReg8(Reg8(modrm.Reg))
	modrm.SetRm8(cpu, left^right)
}

func (cpu *CPU) xorEvGv() {
	modrm, err := ParseModRM(cpu)
	if err != nil {
		return
	}
	left := modrm.GetRm16(cpu)
	right := cpu.Regs.GetReg16(Reg(modrm.Reg))
	modrm.SetRm16(cpu, left^right)
}

func (cpu *CPU) xorGbEb() {
	modrm, err := ParseModRM(cpu)
	if err != nil {
		return
	}
	left := cpu.Regs.GetReg8(Reg8(modrm.Reg))
	right := modrm.GetRm8(cpu)
	modrm.SetR8(cpu, left^right)
}

func (cpu *CPU) xorGvEv() {
	modrm, err := ParseModRM(cpu)
	if err != nil {
		return
	}
	left := cpu.Regs.GetReg16(Reg(modrm.Reg))
	right := modrm.GetRm16(cpu)
	modrm.SetR16(cpu, left^right)
}

func (cpu *CPU) xorALIb() {
	left := cpu.Regs.GetReg8(AL)
	imm8, err := cpu.Fetch8()
	if err != nil {
		return
	}
	cpu.Regs.SetReg8(AL, uint(uint8(left)^imm8))
}

func (cpu *CPU) xorAXIv() {
	left := cpu.Regs.AX()
	imm16, err := cpu.Fetch16()
	if err != nil {
		return
	}
	cpu.Regs.SetReg8(AL, uint(uint16(left)^imm16))
}

func (cpu *CPU) xorEvIv(modrm *ModRM) error {
	left := modrm.GetRm16(cpu)
	imm16, err := cpu.Fetch16()
	if err != nil {
		return err
	}
	modrm.SetR16(cpu, left^uint(imm16))
	return nil
}

// xorEbIb performs a bitwise XOR operation on an 8-bit register or memory location with an 8-bit immediate value.
func (cpu *CPU) xorEbIb(modrm *ModRM) error {
	left := modrm.GetRm8(cpu)
	imm8, err := cpu.Fetch8()
	if err != nil {
		return err
	}
	modrm.SetRm8(cpu, left^uint(imm8))
	return nil
}

// CMP - Compare

func (cpu *CPU) cmpEbGb() {
	modrm, err := ParseModRM(cpu)
	if err != nil {
		return
	}
	left := modrm.GetRm8(cpu)
	right := cpu.Regs.GetReg8(Reg8(modrm.Reg))
	diff := left - right
	cpu.Flags.SetFlagsSub8(diff, left, right)
}

func (cpu *CPU) cmpEvGv() {
	modrm, err := ParseModRM(cpu)
	if err != nil {
		return
	}
	left := modrm.GetRm16(cpu)
	right := cpu.Regs.GetReg16(Reg(modrm.Reg))
	diff := left - right
	cpu.Flags.SetFlagsSub16(diff, left, right)
}

func (cpu *CPU) cmpGbEb() {
	modrm, err := ParseModRM(cpu)
	if err != nil {
		return
	}
	left := cpu.Regs.GetReg8(Reg8(modrm.Reg))
	right := modrm.GetRm8(cpu)
	diff := left - right
	cpu.Flags.SetFlagsSub8(diff, left, right)
}

func (cpu *CPU) cmpGvEv() {
	modrm, err := ParseModRM(cpu)
	if err != nil {
		return
	}
	left := cpu.Regs.GetReg16(Reg(modrm.Reg))
	right := modrm.GetRm16(cpu)
	diff := left - right
	cpu.Flags.SetFlagsSub16(diff, left, right)
}

func (cpu *CPU) cmpALIb() {
	left := cpu.Regs.GetReg8(AL)
	right, err := cpu.Fetch8()
	if err != nil {
		return
	}
	diff := uint(left) - uint(right)
	cpu.Flags.SetFlagsSub8(diff, left, uint(right))
}

func (cpu *CPU) cmpAXIv() {
	left := cpu.Regs.AX()
	right, err := cpu.Fetch16()
	if err != nil {
		return
	}
	diff := uint(left) - uint(right)
	cpu.Flags.SetFlagsSub16(diff, left, uint(right))
}

func (cpu *CPU) cmpEvIv(modrm *ModRM) error {
	imm16, err := cpu.Fetch16()
	if err != nil {
		return err
	}
	left := modrm.GetRm16(cpu)
	diff := left - uint(imm16)
	cpu.Flags.SetFlagsSub16(diff, left, uint(imm16))
	return nil
}

// generate cmpEbIb
func (cpu *CPU) cmpEbIb(modrm *ModRM) error {
	imm8, err := cpu.Fetch8()
	if err != nil {
		return err
	}
	left := modrm.GetRm8(cpu)
	diff := left - uint(imm8)
	cpu.Flags.SetFlagsSub8(diff, left, uint(imm8))
	return nil
}

// TEST - Logical Compare

func (cpu *CPU) testEbGb() {
	modrm, err := ParseModRM(cpu)
	if err != nil {
		return
	}
	left := modrm.GetRm8(cpu)
	right := cpu.Regs.GetReg8(Reg8(modrm.Reg))
	cpu.Flags.ClearFlag(OverflowFlag)
	cpu.Flags.ClearFlag(CarryFlag)
	cpu.Flags.SetFlagsZSP(left & right)
}

func (cpu *CPU) testEvGv() {
	modrm, err := ParseModRM(cpu)
	if err != nil {
		return
	}
	left := modrm.GetRm16(cpu)
	right := cpu.Regs.GetReg16(Reg(modrm.Reg))
	cpu.Flags.ClearFlag(OverflowFlag)
	cpu.Flags.ClearFlag(CarryFlag)
	cpu.Flags.SetFlagsZSP(left & right)
}
