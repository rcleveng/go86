package go86

// ADD - Add

func (cpu *CPU) addEbGb() {
	modrm, err := ParseModRM(cpu)
	if err != nil {
		return
	}
	left := modrm.GetRm8(cpu)
	right := cpu.Regs.GetReg8(Reg8(modrm.Reg))
	sum := left + right
	cpu.Flags.SetFlagsAdd(sum, left, right, 8)
	modrm.SetRm8(cpu, sum)
}

func (cpu *CPU) addEvGv() {
	modrm, err := ParseModRM(cpu)
	if err != nil {
		return
	}
	left := modrm.GetRm16(cpu)
	right := cpu.Regs.GetReg16(Reg(modrm.Reg))
	sum := left + right
	cpu.Flags.SetFlagsAdd(sum, left, right, 16)
	modrm.SetRm16(cpu, sum)
}

func (cpu *CPU) addGbEb() {
	modrm, err := ParseModRM(cpu)
	if err != nil {
		return
	}
	left := cpu.Regs.GetReg8(Reg8(modrm.Reg))
	right := modrm.GetRm8(cpu)
	sum := left + right
	cpu.Flags.SetFlagsAdd(sum, left, right, 8)
	modrm.SetR8(cpu, sum)
}

func (cpu *CPU) addGvEv() {
	modrm, err := ParseModRM(cpu)
	if err != nil {
		return
	}
	left := cpu.Regs.GetReg16(Reg(modrm.Reg))
	right := modrm.GetRm16(cpu)
	sum := left + right
	cpu.Flags.SetFlagsAdd(sum, left, right, 16)
	modrm.SetR16(cpu, sum)
}

func (cpu *CPU) addALIb() {
	left := cpu.Regs.GetReg8(AL)
	right, err := cpu.Fetch8()
	if err != nil {
		return
	}
	sum := uint(left) + uint(right)
	cpu.Flags.SetFlagsAdd(sum, left, uint(right), 8)
	cpu.Regs.SetReg8(AL, sum)
}

func (cpu *CPU) addAXIv() {
	left := cpu.Regs.AX()
	right, err := cpu.Fetch16()
	if err != nil {
		return
	}
	sum := uint(left) + uint(right)
	cpu.Flags.SetFlagsAdd(sum, left, uint(right), 16)
	cpu.Regs.SetReg16(AX, sum)
}

func (cpu *CPU) addEvIv(modrm *ModRM) error {
	left := modrm.GetRm16(cpu)
	imm16, err := cpu.Fetch16()
	if err != nil {
		return err
	}
	sum := left + uint(imm16)
	cpu.Flags.SetFlagsAdd16(sum, left, uint(imm16))
	modrm.SetRm16(cpu, sum)
	return nil
}

func (cpu *CPU) addEbIb(modrm *ModRM) error {
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

// SUB - Subtract

func (cpu *CPU) subEbGb() {
	modrm, err := ParseModRM(cpu)
	if err != nil {
		return
	}
	left := modrm.GetRm8(cpu)
	right := cpu.Regs.GetReg8(Reg8(modrm.Reg))
	diff := left - right
	cpu.Flags.SetFlagsSub8(diff, left, right)
	modrm.SetRm8(cpu, diff)
}

func (cpu *CPU) subEvGv() {
	modrm, err := ParseModRM(cpu)
	if err != nil {
		return
	}
	left := modrm.GetRm16(cpu)
	right := cpu.Regs.GetReg16(Reg(modrm.Reg))
	diff := left - right
	cpu.Flags.SetFlagsSub16(diff, left, right)
	modrm.SetRm16(cpu, diff)
}

func (cpu *CPU) subGbEb() {
	modrm, err := ParseModRM(cpu)
	if err != nil {
		return
	}
	left := cpu.Regs.GetReg8(Reg8(modrm.Reg))
	right := modrm.GetRm8(cpu)
	diff := left - right
	cpu.Flags.SetFlagsSub8(diff, left, right)
	modrm.SetR8(cpu, diff)
}

func (cpu *CPU) subGvEv() {
	modrm, err := ParseModRM(cpu)
	if err != nil {
		return
	}
	left := cpu.Regs.GetReg16(Reg(modrm.Reg))
	right := modrm.GetRm16(cpu)
	diff := left - right
	cpu.Flags.SetFlagsSub16(diff, left, right)
	modrm.SetR16(cpu, diff)
}

func (cpu *CPU) subALIb() {
	left := cpu.Regs.GetReg8(AL)
	right, err := cpu.Fetch8()
	if err != nil {
		return
	}
	diff := uint(left) - uint(right)
	cpu.Flags.SetFlagsSub8(diff, left, uint(right))
	cpu.Regs.SetReg8(AL, diff)
}

func (cpu *CPU) subAXIv() {
	left := cpu.Regs.AX()
	right, err := cpu.Fetch16()
	if err != nil {
		return
	}
	diff := uint(left) - uint(right)
	cpu.Flags.SetFlagsSub16(diff, left, uint(right))
	cpu.Regs.SetReg16(AX, diff)
}

func (cpu *CPU) subEvIv(modrm *ModRM) error {
	left := modrm.GetRm16(cpu)
	imm16, err := cpu.Fetch16()
	if err != nil {
		return err
	}
	diff := left - uint(imm16)
	cpu.Flags.SetFlagsSub16(diff, left, uint(imm16))
	modrm.SetRm16(cpu, diff)
	return nil
}

// subEbIb subtracts an 8-bit immediate value from an 8-bit register or memory location.
func (cpu *CPU) subEbIb(modrm *ModRM) error {
	left := modrm.GetRm8(cpu)
	imm8, err := cpu.Fetch8()
	if err != nil {
		return err
	}
	diff := left - uint(imm8)
	cpu.Flags.SetFlagsSub8(diff, left, uint(imm8))
	modrm.SetRm8(cpu, diff)
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
