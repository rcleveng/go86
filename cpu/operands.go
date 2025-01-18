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
	Eb    = Operand{"Eb", 8, true}
	Ev    = Operand{"Ev", 16, true}
	Ew    = Operand{"Ee", 16, true}
	Gb    = Operand{"Gb", 8, true}
	Gv    = Operand{"Gv", 16, true}
	Ib    = Operand{"Ib", 8, false}
	Iv    = Operand{"Iv", 16, false}
	Ob    = Operand{"Ob", 8, false}
	Ov    = Operand{"Ob", 16, false}
	Sw    = Operand{"Sw", 16, true}
	RegAL = Operand{"AL", 8, false}
	RegAX = Operand{"AX", 16, false}
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

func (op Operand) GetByOperand(cpu *CPU) (uint, error) {
	switch op {
	case Eb:
		return cpu.ModRM.GetRm8(cpu), nil
	case Ev, Ew:
		return cpu.ModRM.GetRm16(cpu), nil
	case Gb:
		return cpu.ModRM.R8(cpu), nil
	case Gv:
		return cpu.ModRM.R16(cpu), nil
	case Ib:
		return cpu.Inst.GetImm8(cpu)
	case Iv:
		return cpu.Inst.GetImm16(cpu)
	case Ob:
		imm16, err := cpu.Inst.GetImm16(cpu)
		if err != nil {
			return 0, err
		}

		// TODO - handle segment overrides
		value := uint(cpu.Mem.GetMem8(cpu.Regs.DS(), imm16))
		return value, nil

	case Ov:
		imm16, err := cpu.Inst.GetImm16(cpu)
		if err != nil {
			return 0, err
		}
		// TODO - handle segment overrides
		value := uint(cpu.Mem.GetMem16(cpu.Regs.DS(), imm16))
		return value, nil

	case Sw:
		return cpu.Regs.GetSeg16(SReg(cpu.ModRM.Reg)), nil
	case RegAL:
		return cpu.Regs.GetReg8(AL), nil
	case RegAX:
		return cpu.Regs.GetReg16(AX), nil
	}

	return 0, fmt.Errorf("GetByOperand: unknown operand: %v", op)
}

func (op Operand) SetByOperand(cpu *CPU, val uint) error {
	switch op {
	case Eb:
		cpu.ModRM.SetRm8(cpu, val)
		return nil
	case Ev, Ew:
		cpu.ModRM.SetRm16(cpu, val)
		return nil
	case Gb:
		cpu.ModRM.SetR8(cpu, val)
		return nil
	case Gv:
		cpu.ModRM.SetR16(cpu, val)
		return nil

	case Ob:
		imm16, err := cpu.Inst.GetImm16(cpu)
		if err != nil {
			return err
		}
		cpu.Mem.SetMem8(cpu.Regs.DS(), imm16, uint8(val))

	case Ov:
		imm16, err := cpu.Inst.GetImm16(cpu)
		if err != nil {
			return err
		}
		// TODO - handle segment overrides
		cpu.Mem.SetMem16(cpu.Regs.DS(), imm16, uint16(val))

	case Sw:
		cpu.Regs.SetSeg16(SReg(cpu.ModRM.Reg), val)
		return nil
	case RegAL:
		cpu.Regs.SetReg8(AL, val)
		return nil
	case RegAX:
		cpu.Regs.SetReg16(AX, val)
		return nil
	}

	return fmt.Errorf("SetByOperand: unknown operand")
}

func ParseTwoOperands(cpu *CPU, leftop, rightop Operand) (uint, uint, error) {
	if leftop.HasModRM() || rightop.HasModRM() {
		if cpu.ModRM == nil {
			err := cpu.ParseModRMByte()
			if err != nil {
				return 0, 0, err
			}
		}
	}
	left, err := leftop.GetByOperand(cpu)
	if err != nil {
		return 0, 0, err
	}
	right, err := rightop.GetByOperand(cpu)
	if err != nil {
		return 0, 0, err
	}

	return left, right, nil
}
