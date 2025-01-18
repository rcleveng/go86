package go86

import (
	"fmt"
)

type Inst struct {
	OpCode          uint8
	Lock            bool
	RepNe           bool
	Rep             bool
	SegmentOverride SReg
	Len             int
}

// Decode decodes the leading bytes in src as a single instruction.
// The mode arguments specifies the assumed processor mode:
func Decode(src []byte) (inst Inst, err error) {
	inst.Len = 0
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
			inst.SegmentOverride = CS
			inst.Len++
		case 0x36: // SS segment override
			inst.SegmentOverride = SS
			inst.Len++
		case 0x3E: // DS segment override
			inst.SegmentOverride = DS
			inst.Len++
		case 0x26: // ES segment override
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
