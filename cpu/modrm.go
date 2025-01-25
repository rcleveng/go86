package go86

import "fmt"

type ModRM struct {
	raw    uint8
	Mod    uint8
	Reg    uint8
	Rm     uint8
	Disp8  int8
	Disp16 int16
	// add these for 32-bit support
	// Sib   uint8
	// Scale uint8
	// Index uint8
	// Base  uint8
}

func (m ModRM) String() string {
	return fmt.Sprintf("ModRM{raw=0x%02X, Mod=%d, Reg=%d, Rm=%d, Disp8=%d, Disp16=%d}",
		m.raw, m.Mod, m.Reg, m.Rm, m.Disp8, m.Disp16)
}

func NewModRM(r CpuInstructionReader, b uint8) (*ModRM, error) {
	m := &ModRM{
		raw: b,
		Mod: (b & 0xc0) >> 6,
		Reg: (b & 0x38) >> 3,
		Rm:  b & 0x07,
	}

	if m.Mod == 2 || (m.Mod == 0 && m.Rm == 6) {
		// special case for mod==0 and rm==6 as well as general
		// case when mod == 2
		disp, err := r.Fetch16()
		if err != nil {
			return nil, err
		}
		m.Disp16 = int16(disp)
		return m, nil
	} else if m.Mod == 1 {
		disp, err := r.Fetch8()
		if err != nil {
			return nil, err
		}
		m.Disp8 = int8(disp)
	}

	return m, nil
}

func (m *ModRM) segmentToUse(cpu *CPU, inst *Inst) uint {
	if inst.HasSegmentOverride {
		// TODO - do we need to make sure this instruction supports it?
		return uint(inst.SegmentOverride)
	}

	if m.Rm == 2 || m.Rm == 3 {
		// addressing modes that use the BP use SS
		return cpu.Regs.SS()
	}
	return cpu.Regs.DS()
}

func (m *ModRM) effectiveAddressOffset8(cpu *CPU) uint {
	if m.Mod == 3 {
		// Only override this case
		return cpu.Regs.GetReg8(Reg8(m.Rm))
	}
	return m.effectiveAddressOffset16(cpu)
}

func (m *ModRM) effectiveAddressOffset16(cpu *CPU) uint {
	if m.Mod == 0 && m.Rm == 6 {
		return uint(m.Disp16)
	}
	if m.Mod == 3 {
		return uint(cpu.Regs.GetReg16(Reg(m.Rm)))
	}

	var address uint = 0
	switch m.Rm {
	case 0:
		address = cpu.Regs.BX() + cpu.Regs.SI()
	case 1:
		address = cpu.Regs.BX() + cpu.Regs.DI()
	case 2:
		address = cpu.Regs.BP() + cpu.Regs.SI()
	case 3:
		address = cpu.Regs.BP() + cpu.Regs.DI()
	case 4:
		address = cpu.Regs.SI()
	case 5:
		address = cpu.Regs.DI()
	case 6:
		address = cpu.Regs.BP()
	case 7:
		address = cpu.Regs.BX()
	}

	if m.Mod == 0 {
		return address
	}
	if m.Mod == 1 {
		return uint(int(address) + int(m.Disp8))
	}
	if m.Mod == 2 {
		return uint(int(address) + int(m.Disp16))
	}

	panic("unknown mod value: " + string(m.Mod))
}

func (m *ModRM) GetRm8(cpu *CPU, inst *Inst) uint {
	if m.Mod == 3 {
		return cpu.Regs.GetReg8(Reg8(m.Rm))
	}
	offset := m.effectiveAddressOffset8(cpu)
	seg := m.segmentToUse(cpu, inst)
	return uint(cpu.Mem.GetMem8(seg, offset))
}

func (m *ModRM) SetRm8(cpu *CPU, inst *Inst, value uint) {
	if m.Mod == 3 {
		// Can't use SetR16 since we want Rm to specify the register and not Reg
		cpu.Regs.SetReg8(Reg8(m.Rm), value)
		return
	}
	offset := m.effectiveAddressOffset8(cpu)
	seg := m.segmentToUse(cpu, inst)
	cpu.Mem.SetMem8(seg, offset, uint8(value))
}

func (m *ModRM) GetRm16(cpu *CPU, inst *Inst) uint {
	if m.Mod == 3 {
		return cpu.Regs.GetReg16(Reg(m.Rm))
	}
	offset := m.effectiveAddressOffset16(cpu)
	seg := m.segmentToUse(cpu, inst)
	return uint(cpu.Mem.GetMem16(seg, offset))
}

// Gets the memory location (segment:offset) to use to lookup indirect values
func (m *ModRM) GetMemoryLocation(cpu *CPU, inst *Inst) (seg uint, offset uint, err error) {
	if m.Mod == 3 {
		return uint(CS), 0, fmt.Errorf("can't get memory location when mod == 3")
	}
	offset = m.effectiveAddressOffset16(cpu)
	seg = m.segmentToUse(cpu, inst)
	return seg, offset, nil
}

func (m *ModRM) SetRm16(cpu *CPU, inst *Inst, value uint) {
	if m.Mod == 3 {
		// Can't use SetR16 since we want Rm to specify the register and not Reg
		cpu.Regs.SetReg16(Reg(m.Rm), value)
		return
	}
	offset := m.effectiveAddressOffset8(cpu)
	seg := m.segmentToUse(cpu, inst)
	cpu.Mem.SetMem16(seg, offset, uint16(value))
}

func (m *ModRM) R8(cpu *CPU) uint {
	return cpu.Regs.GetReg8(Reg8(m.Reg))
}

func (m *ModRM) SetR8(cpu *CPU, value uint) {
	cpu.Regs.SetReg8(Reg8(m.Reg), value)
}

func (m *ModRM) R16(cpu *CPU) uint {
	return cpu.Regs.GetReg16(Reg(m.Reg))
}

func (m *ModRM) SetR16(cpu *CPU, value uint) {
	cpu.Regs.SetReg16(Reg(m.Reg), value)
}
