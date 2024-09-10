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
	cpu.Flags.SetFlagsAdd8(sum, left, right)
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
	cpu.Flags.SetFlagsAdd16(sum, left, right)
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
	cpu.Flags.SetFlagsAdd8(sum, left, right)
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
	cpu.Flags.SetFlagsAdd16(sum, left, right)
	modrm.SetR16(cpu, sum)
}

func (cpu *CPU) addALIb() {
	left := cpu.Regs.GetReg8(AL)
	right, err := cpu.Fetch8()
	if err != nil {
		return
	}
	sum := uint(left) + uint(right)
	cpu.Flags.SetFlagsAdd8(sum, left, uint(right))
	cpu.Regs.SetReg8(AL, sum)
}

func (cpu *CPU) addAXIv() {
	left := cpu.Regs.AX()
	right, err := cpu.Fetch16()
	if err != nil {
		return
	}
	sum := uint(left) + uint(right)
	cpu.Flags.SetFlagsAdd16(sum, left, uint(right))
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

func (cpu *CPU) addEvIb(modrm *ModRM) error {
	left := modrm.GetRm16(cpu)
	imm8, err := cpu.Fetch8()
	if err != nil {
		return err
	}
	sum := left + uint(int16(imm8))
	cpu.Flags.SetFlagsAdd16(sum, left, uint(imm8))
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

func (cpu *CPU) subEvIb(modrm *ModRM) error {
	left := modrm.GetRm16(cpu)
	imm8, err := cpu.Fetch8()
	if err != nil {
		return err
	}
	right := uint(int16(imm8))
	diff := left - right
	cpu.Flags.SetFlagsSub8(diff, left, right)
	modrm.SetRm16(cpu, diff)
	return nil
}
