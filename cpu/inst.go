package go86

import (
	"fmt"
)

type Inst struct {
	OpCode             uint8
	Lock               bool
	RepNe              bool
	Rep                bool
	HasSegmentOverride bool
	SegmentOverride    SReg
	Len                int

	// Current Imm8 that didn't come from ModRM
	Imm8 *uint
	// Current Imm16 that didn't come from ModRM
	Imm16 *uint
}

func ImmValue(val uint) *uint {
	result := val
	return &result
}

func (inst *Inst) GetImm16(cpu *CPU) (uint, error) {
	if inst.Imm16 == nil {
		imm16, err := cpu.Fetch16()
		if err != nil {
			return 0, err
		}
		inst.Imm16 = ImmValue(uint(imm16))
	}
	return *inst.Imm16, nil
}

func (inst *Inst) GetImm8(cpu *CPU) (uint, error) {
	if inst.Imm8 == nil {
		imm8, err := cpu.Fetch8()
		if err != nil {
			return 0, err
		}
		inst.Imm8 = ImmValue(uint(imm8))
	}
	return *inst.Imm8, nil
}

// Decode decodes the leading bytes in src as a single instruction.
// The mode arguments specifies the assumed processor mode:
func Decode(src []byte) (*Inst, error) {
	inst := &Inst{}
	for i, b := range src {
		switch b {
		// Prefix group 1
		case 0xF0:
			inst.Lock = true
			inst.Len++
		case 0xF2:
			inst.RepNe = true
			inst.Len++
		case 0xF3:
			inst.Rep = true
			inst.Len++

		// Prefix group 2
		case 0x2E: // CS segment override
			inst.HasSegmentOverride = true
			inst.SegmentOverride = CS
			inst.Len++
		case 0x36: // SS segment override
			inst.HasSegmentOverride = true
			inst.SegmentOverride = SS
			inst.Len++
		case 0x3E: // DS segment override
			inst.HasSegmentOverride = true
			inst.SegmentOverride = DS
			inst.Len++
		case 0x26: // ES segment override
			inst.HasSegmentOverride = true
			inst.SegmentOverride = ES
			inst.Len++
			/*
				Skip till 386 support
				case 0x0F: // Multi byte opcodes
				case 0x64: // FS segment override
					inst.SegmentOverride = FS
				case 0x65: // GS segment override
					inst.SegmentOverride = GS
				case 0x2E: // Branch not taken
				case 0x3E: // Branch taken
			*/
		default:
			inst.Len++
			inst.OpCode = src[i]
			return inst, nil
		}
	}
	return inst, fmt.Errorf("unable to decode instructions")
}
