package go86

// xchg - Exchange
func (cpu *CPU) xchgGbEb() {
	modrm, err := ParseModRM(cpu)
	if err != nil {
		return
	}
	left := cpu.Regs.GetReg8(Reg8(modrm.Reg))
	right := modrm.GetRm8(cpu)
	cpu.Regs.SetReg8(Reg8(modrm.Reg), right)
	modrm.SetRm8(cpu, left)
}

// xchgGvEv() - Exchange
func (cpu *CPU) xchgGvEv() {
	modrm, err := ParseModRM(cpu)
	if err != nil {
		return
	}
	left := cpu.Regs.GetReg16(Reg(modrm.Reg))
	right := modrm.GetRm16(cpu)
	cpu.Regs.SetReg16(Reg(modrm.Reg), right)
	modrm.SetRm16(cpu, left)
}

// xchgRegs() - Exchange
func (cpu *CPU) xchgRegs(leftReg Reg, rightReg Reg) {
	left := cpu.Regs.GetReg16(leftReg)
	right := cpu.Regs.GetReg16(rightReg)
	cpu.Regs.SetReg16(leftReg, right)
	cpu.Regs.SetReg16(rightReg, left)
}

// MOVE

// movALOb - Move
func (cpu *CPU) movALOb() error {
	offset, err := cpu.Fetch16()
	if err != nil {
		return err
	}
	value := uint(cpu.Mem.GetMem8(cpu.Regs.DS(), uint(offset)))
	cpu.Regs.SetReg8(AL, value)
	return nil
}

// movAXOv - Move
func (cpu *CPU) movAXOv() error {
	offset, err := cpu.Fetch16()
	if err != nil {
		return err
	}
	value := uint(cpu.Mem.GetMem16(cpu.Regs.DS(), uint(offset)))
	cpu.Regs.SetReg16(AX, value)
	return nil
}

// movObAL - Move
func (cpu *CPU) movObAL() error {
	offset, err := cpu.Fetch16()
	if err != nil {
		return err
	}
	value := uint8(cpu.Regs.GetReg8(AL))
	cpu.Mem.SetMem8(cpu.Regs.DS(), uint(offset), value)
	return nil
}

// movOvAX - Move
func (cpu *CPU) movOvAX() error {
	offset, err := cpu.Fetch16()
	if err != nil {
		return err
	}
	value := uint16(cpu.Regs.GetReg16(AX))
	cpu.Mem.SetMem16(cpu.Regs.DS(), uint(offset), value)
	return nil
}

// MOVE - Move Gv, Ev and family

// movEbGb - Move
func (cpu *CPU) movEbGb() error {
	modrm, err := ParseModRM(cpu)
	if err != nil {
		return err
	}
	value := cpu.Regs.GetReg8(Reg8(modrm.Reg))
	modrm.SetRm8(cpu, value)
	return nil
}

// movGbEb - Move
func (cpu *CPU) movGbEb() error {
	modrm, err := ParseModRM(cpu)
	if err != nil {
		return err
	}
	value := modrm.GetRm8(cpu)
	cpu.Regs.SetReg8(Reg8(modrm.Reg), value)
	return nil
}

// movEvGv - Move
func (cpu *CPU) movEvGv() error {
	modrm, err := ParseModRM(cpu)
	if err != nil {
		return err
	}
	value := cpu.Regs.GetReg16(Reg(modrm.Reg))
	modrm.SetRm16(cpu, value)
	return nil
}

// movGvEv - Move
func (cpu *CPU) movGvEv() error {
	modrm, err := ParseModRM(cpu)
	if err != nil {
		return err
	}
	value := modrm.GetRm16(cpu)
	cpu.Regs.SetReg16(Reg(modrm.Reg), value)
	return nil
}

// moveEwSw - Move
func (cpu *CPU) moveEwSw() error {
	modrm, err := ParseModRM(cpu)
	if err != nil {
		return err
	}
	value := cpu.Regs.GetSeg16(SReg(modrm.Reg))
	modrm.SetRm16(cpu, value)
	return nil
}

// movSwEw - Move
func (cpu *CPU) movSwEw() error {
	modrm, err := ParseModRM(cpu)
	if err != nil {
		return err
	}
	value := modrm.GetRm16(cpu)
	cpu.Regs.SetSeg16(SReg(modrm.Reg), value)
	return nil
}

// leaGvM - Load Effective Address
func (cpu *CPU) leaGvM() error {
	modrm, err := ParseModRM(cpu)
	if err != nil {
		return err
	}
	offset := modrm.effectiveAddressOffset16(cpu)
	cpu.Regs.SetReg16(Reg(modrm.Reg), offset)
	return nil
}

// popEv - Pop
func (cpu *CPU) popEv() error {
	modrm, err := ParseModRM(cpu)
	if err != nil {
		return err
	}
	value := cpu.Regs.Pop16(cpu.Mem)
	modrm.SetRm16(cpu, uint(value))
	return nil
}

// Move - Move AL, Ib and family

// movALIb - Move
func (cpu *CPU) movRegIb(reg Reg8) error {
	value, err := cpu.Fetch8()
	if err != nil {
		return err
	}
	cpu.Regs.SetReg8(reg, uint(value))
	return nil
}

func (cpu *CPU) movRegIv(reg Reg) error {
	value, err := cpu.Fetch16()
	if err != nil {
		return err
	}
	cpu.Regs.SetReg16(reg, uint(value))
	return nil
}
