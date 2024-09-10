package go86

import (
	"math/bits"
	"strings"
)

// Constants to define the bit value for each x86 Flags.
// The full set of flags are stored in the f.Flags memeber.
const (
	CF uint32 = 1 << iota
	_  uint32 = 1 << iota
	PF uint32 = 1 << iota
	_  uint32 = 1 << iota
	AF uint32 = 1 << iota
	_  uint32 = 1 << iota
	ZF uint32 = 1 << iota
	SF uint32 = 1 << iota
	TF uint32 = 1 << iota
	IF uint32 = 1 << iota
	DF uint32 = 1 << iota
	OF uint32 = 1 << iota
)

// Friendly names for the flags
const (
	CarryFlag     = CF
	ParityFlag    = PF
	AdjustFlag    = AF
	ZeroFlag      = ZF
	SignFlag      = SF
	TrapFlag      = TF
	InterruptFlag = IF
	DirectionFlag = DF
	OverflowFlag  = OF
)

type Flags struct {
	eflags uint32
}

// Unconditionally clears flag v
func (f *Flags) ClearFlag(v uint32) {
	// This is cool
	f.eflags &^= v
}

// Sets or clears the flag f based on cond.
func (f *Flags) SetFlagIf(v uint32, cond bool) {
	if cond {
		f.eflags |= v
	} else {
		f.eflags &^= v
	}
}

// Sets the flags in f
func (f *Flags) SetFlags(v uint32) {
	f.eflags |= v
}

func (f *Flags) ReplaceAllFlags(v uint32) {
	f.eflags = v
}

func (f *Flags) Value() uint32 {
	return f.eflags
}

func (f *Flags) IsEnabled(v uint32) bool {
	return f.eflags&v != 0
}

// Writes either the on or off value of flag to sb.
func (f *Flags) WriteFlag(sb *strings.Builder, flag uint32, on string, off string) {
	if (f.eflags & flag) != 0 {
		sb.WriteString(on)
	} else {
		sb.WriteString(off)
	}
	sb.WriteString(" ")
}

// Returns a string in the canonical shorthand format for the flags.  The
// format is upper case letters for a flag being on, lower case otherwise.
// Example: "O d i S z a p c t"
// Implements stringer interface
func (f *Flags) String() string {
	var sb strings.Builder
	f.WriteFlag(&sb, OF, "O", "o")
	f.WriteFlag(&sb, DF, "D", "d")
	f.WriteFlag(&sb, IF, "I", "i")
	f.WriteFlag(&sb, SF, "S", "s")
	f.WriteFlag(&sb, ZF, "Z", "z")
	f.WriteFlag(&sb, AF, "A", "a")
	f.WriteFlag(&sb, PF, "P", "p")
	f.WriteFlag(&sb, CF, "C", "c")
	f.WriteFlag(&sb, TF, "T", "t")
	return sb.String()
}

// Update the Zero, Sign, and Parity flags where result is the result of an
// instruction execution and numbits is either 8 or 16 depending on the
// instruction.
func (f *Flags) SetFlagsZSP(result uint) {
	f.SetFlagIf(ZF, result == 0)
	f.SetFlagIf(SF, (result&0x80) != 0)
	// Yes. X86 only ever uses LSB for PF
	count := bits.OnesCount8(uint8(result))
	f.SetFlagIf(PF, count%2 == 0)
}

// Update all of the flags where result is the result of an addition
// instruction using the values for the result, source and destintationm operands
// and numbits is either 8 or 16 depending on the instruction.

// SetFlagsAdd16 sets the flags for a 16-bit addition
func (f *Flags) SetFlagsAdd16(res, src, dst uint) {
	f.SetFlagIf(CF, res>>16 != 0)
	f.SetFlagIf(OF, (res^src)&(res^dst)&0x8000 != 0)
	f.SetFlagIf(AF, (res^src^dst)&0x10 != 0)
	f.SetFlagsZSP(res)
}

// SetFlagsAdd8 sets the flags for an 8-bit addition
func (f *Flags) SetFlagsAdd8(res, src, dst uint) {
	f.SetFlagIf(CF, res>>8 != 0)
	f.SetFlagIf(OF, (res^src)&(res^dst)&0x80 != 0)
	f.SetFlagIf(AF, (res^src^dst)&0x10 != 0)
	f.SetFlagsZSP(res)
}

// Update all of the flags where result is the result of an subtraction
// instruction using the values for the result, source and destintationm operands
// and numbits is either 8 or 16 depending on the instruction.

func (f *Flags) SetFlagsSub8(res, src, dst uint) {
	f.SetFlagIf(CF, res&0x100 == 0x100)
	f.SetFlagIf(OF, ((dst^src)&(res^dst)&0x80) != 0)
	f.SetFlagIf(AF, (res^src^dst)&0x10 != 0)
	f.SetFlagsZSP(res)
}

func (f *Flags) SetFlagsSub16(res, src, dst uint) {
	f.SetFlagIf(OF, ((dst^src)&(res^dst)&0x8000) != 0)
	f.SetFlagIf(CF, res&0x10000 == 0x10000)
	f.SetFlagIf(AF, (res^src^dst)&0x10 != 0)
	f.SetFlagsZSP(res)
}

// Helper function, used in CMP, where just PSZ survives
func (f *Flags) ClearFlagsCOA() {
	f.eflags &^= (AF | CF | OF)
}
