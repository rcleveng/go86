package go86

import "fmt"

func getDSSI(cpu *CPU, bit int, allowOverride bool) (uint, error) {
	sreg := DS
	soff := cpu.Regs.GetReg16(SI)
	if allowOverride && cpu.Inst.HasSegmentOverride {
		sreg = cpu.Inst.SegmentOverride
	}

	switch bit {
	case 8:
		return uint(cpu.Mem.GetMem8(cpu.Regs.GetSeg16(sreg), soff)), nil
	case 16:
		return uint(cpu.Mem.GetMem16(cpu.Regs.GetSeg16(sreg), soff)), nil
	default:
		return 0, fmt.Errorf("unknown bits: %d", bit)
	}
}

func getESDI(cpu *CPU, bit int) (uint, error) {
	dreg := ES
	doff := cpu.Regs.GetReg16(DI)

	switch bit {
	case 8:
		return uint(cpu.Mem.GetMem8(cpu.Regs.GetSeg16(dreg), doff)), nil
	case 16:
		return uint(cpu.Mem.GetMem16(cpu.Regs.GetSeg16(dreg), doff)), nil
	default:
		return 0, fmt.Errorf("unknown bits: %d", bit)
	}
}

func incCounter(cpu *CPU, count uint, df bool, reg Reg) {
	switch df {
	case false: // increment
		cpu.Regs.Inc16(reg, count)
	case true: // decrement
		cpu.Regs.Dec16(reg, count)
	}
}

func incCounters(cpu *CPU, count uint, regs ...Reg) {
	df := cpu.Flags.IsEnabled(DirectionFlag) // true == 1
	for _, reg := range regs {
		incCounter(cpu, count, df, reg)
	}
}

func (cpu *CPU) cmps(bit int) error {
	left, err := getDSSI(cpu, bit, true)
	if err != nil {
		return err
	}
	right, err := getESDI(cpu, bit)
	if err != nil {
		return err
	}

	// Update the pointers
	incCounters(cpu, uint(bit/8), SI, DI)
	return cpu.compare(left, right, 8)
}

func (cpu *CPU) scasOne(bit int) error {
	esdi, err := getESDI(cpu, bit)
	if err != nil {
		return err
	}

	// Update the pointers
	var reg, count uint
	switch bit {
	case 8:
		reg = cpu.Regs.GetReg8(AL)
		count = 1
	case 16:
		reg = cpu.Regs.GetReg16(AX)
		count = 2
	}

	df := cpu.Flags.IsEnabled(DirectionFlag) // true == 1
	incCounter(cpu, count, df, DI)
	return cpu.compare(reg, esdi, bit)
}

func repeat(cpu *CPU, usesRepE, usesRepNE bool, inst *Inst, f func(*CPU) error) error {
	cx := cpu.Regs.CX()
	for cx != 0 {
		if err := f(cpu); err != nil {
			return err
		}
		cx--
		if usesRepE && inst.Rep && !cpu.Flags.IsEnabled(ZeroFlag) {
			break
		}
		if usesRepNE && inst.RepNe && cpu.Flags.IsEnabled(ZeroFlag) {
			break
		}
	}
	cpu.Regs.SetReg16(CX, cx)
	return nil
}

func (cpu *CPU) scas(inst *Inst, bit int) error {
	if !inst.Rep && !inst.RepNe {
		return cpu.scasOne(bit)
	}

	return repeat(cpu, true, true, inst, func(cpu *CPU) error {
		return cpu.scasOne(bit)
	})

}

func (cpu *CPU) stosOne(bit int) error {
	// N.B. The ES segment cannot be overridden with a segment override prefix.
	df := cpu.Flags.IsEnabled(DirectionFlag) // true == 1
	switch bit {
	case 8:
		cpu.Mem.SetMem8(cpu.Regs.GetSeg16(ES), cpu.Regs.GetReg16(DI), uint8(cpu.Regs.GetReg8(AL)))
		incCounter(cpu, 1, df, DI)
	case 16:
		cpu.Mem.SetMem16(cpu.Regs.GetSeg16(ES), cpu.Regs.GetReg16(DI), uint16(cpu.Regs.GetReg16(AX)))
		incCounter(cpu, 2, df, DI)
	default:
		panic("wrong bits")
	}

	return nil
}

func (cpu *CPU) stos(inst *Inst, bit int) error {
	if !inst.Rep && !inst.RepNe {
		return cpu.stosOne(bit)
	}
	return repeat(cpu, true, true, inst, func(cpu *CPU) error {
		return cpu.stosOne(bit)
	})
}

func (cpu *CPU) lodsOne(bit int) error {
	dssi, err := getDSSI(cpu, bit, true)
	if err != nil {
		return err
	}

	df := cpu.Flags.IsEnabled(DirectionFlag) // true == 1
	switch bit {
	case 8:
		cpu.Regs.SetReg8(AL, dssi)
		incCounter(cpu, 1, df, SI)
	case 16:
		cpu.Regs.SetReg16(AX, dssi)
		incCounter(cpu, 2, df, SI)
	default:
		panic("wrong bits")
	}
	return nil
}

func (cpu *CPU) lods(inst *Inst, bit int) error {
	if !inst.Rep && !inst.RepNe {
		return cpu.lodsOne(bit)
	}
	return repeat(cpu, true, true, inst, func(cpu *CPU) error {
		return cpu.lodsOne(bit)
	})
}
