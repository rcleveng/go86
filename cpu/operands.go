package go86

import "fmt"

type Operand struct {
	code     string
	bits     int
	hasModRM bool
}

type OperandPair struct {
	Left, Right Operand
}

func (r Operand) String() string {
	return r.code
}

func (r Operand) Code() string {
	return r.code
}

func (r Operand) Bits() int {
	return r.bits
}

func (r Operand) HasModRM() bool {
	return r.hasModRM
}

var (
	Eb     = Operand{"Eb", 8, true}
	Ev     = Operand{"Ev", 16, true}
	Ew     = Operand{"Ee", 16, true}
	Gb     = Operand{"Gb", 8, true}
	Gv     = Operand{"Gv", 16, true}
	Ib     = Operand{"Ib", 8, false}
	Iv     = Operand{"Iv", 16, false}
	Ob     = Operand{"Ob", 8, false}
	Ov     = Operand{"Ob", 16, false}
	Sw     = Operand{"Sw", 16, true}
	RegAL  = Operand{"AL", 8, false}
	RegAX  = Operand{"AX", 16, false}
	RegCL  = Operand{"CL", 8, false}
	ValOne = Operand{"1", 8, false}
)

var (
	EbGb = OperandPair{Eb, Gb}
	EvGv = OperandPair{Ev, Gv}
	GbEb = OperandPair{Gb, Eb}
	GvEv = OperandPair{Gv, Ev}
	ALIb = OperandPair{RegAL, Ib}
	AXIv = OperandPair{RegAX, Iv}
)

var StandardOperands = []OperandPair{
	EbGb,
	EvGv,
	GbEb,
	GvEv,
	ALIb,
	AXIv,
}

func (op Operand) GetByOperand(cpu *CPU, inst *Inst, mem *Memory, regs *Registers) (uint, error) {
	switch op {
	case Eb:
		return inst.ModRM.GetRm8(cpu, inst), nil
	case Ev, Ew:
		return inst.ModRM.GetRm16(cpu, inst), nil
	case Gb:
		return inst.ModRM.R8(cpu), nil
	case Gv:
		return inst.ModRM.R16(cpu), nil
	case Ib:
		return inst.GetImm8(cpu)
	case Iv:
		return inst.GetImm16(cpu)
	case Ob:
		imm16, err := inst.GetImm16(cpu)
		if err != nil {
			return 0, err
		}

		seg := DS
		if inst.HasSegmentOverride {
			seg = inst.SegmentOverride
		}
		sval := cpu.Regs.GetSeg16(seg)
		value := uint(cpu.Mem.GetMem8(sval, imm16))
		return value, nil

	case Ov:
		imm16, err := inst.GetImm16(cpu)
		if err != nil {
			return 0, err
		}
		seg := DS
		if inst.HasSegmentOverride {
			seg = inst.SegmentOverride
		}
		sval := cpu.Regs.GetSeg16(seg)
		value := uint(cpu.Mem.GetMem16(sval, imm16))
		return value, nil

	case Sw:
		return cpu.Regs.GetSeg16(SReg(inst.ModRM.Reg)), nil
	case RegAL:
		return cpu.Regs.GetReg8(AL), nil
	case RegAX:
		return cpu.Regs.GetReg16(AX), nil
	case RegCL:
		return cpu.Regs.GetReg8(CL), nil
	case ValOne:
		return uint(1), nil
	}

	return 0, fmt.Errorf("GetByOperand: unknown operand: %v", op)
}

func (op Operand) SetByOperand(cpu *CPU, inst *Inst, mem *Memory, regs *Registers, val uint) error {
	switch op {
	case Eb:
		inst.ModRM.SetRm8(cpu, inst, val)
		return nil
	case Ev, Ew:
		inst.ModRM.SetRm16(cpu, inst, val)
		return nil
	case Gb:
		inst.ModRM.SetR8(cpu, val)
		return nil
	case Gv:
		inst.ModRM.SetR16(cpu, val)
		return nil

	case Ob:
		imm16, err := inst.GetImm16(cpu)
		if err != nil {
			return err
		}
		// hamding override
		seg := DS
		if inst.HasSegmentOverride {
			seg = inst.SegmentOverride
		}
		mem.SetMem8(regs.GetSeg16(seg), imm16, uint8(val))

	case Ov:
		imm16, err := inst.GetImm16(cpu)
		if err != nil {
			return err
		}
		// hamding override
		seg := DS
		if inst.HasSegmentOverride {
			seg = inst.SegmentOverride
		}
		mem.SetMem16(regs.GetSeg16(seg), imm16, uint16(val))

	case Sw:
		regs.SetSeg16(SReg(inst.ModRM.Reg), val)
		return nil
	case RegAL:
		regs.SetReg8(AL, val)
		return nil
	case RegAX:
		regs.SetReg16(AX, val)
		return nil
	}

	return fmt.Errorf("SetByOperand: unknown operand")
}

func ParseTwoOperands(cpu *CPU, inst *Inst, leftop, rightop Operand) (uint, uint, error) {
	if leftop.HasModRM() || rightop.HasModRM() {
		if inst.ModRM == nil {
			err := inst.FetchModRM()
			if err != nil {
				return 0, 0, err
			}
		}
	}
	left, err := leftop.GetByOperand(cpu, inst, cpu.Mem, cpu.Regs)
	if err != nil {
		return 0, 0, err
	}
	right, err := rightop.GetByOperand(cpu, inst, cpu.Mem, cpu.Regs)
	if err != nil {
		return 0, 0, err
	}

	return left, right, nil
}
