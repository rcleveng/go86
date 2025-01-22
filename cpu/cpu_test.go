package go86

import (
	"encoding/hex"
	"fmt"
	"testing"

	"gotest.tools/v3/assert"
)

func TestCpuSmoke(t *testing.T) {

	cpu := NewCpu(1024 * 1024)
	assert.Equal(t, cpu.Ip, uint16(0))
}

const (
	DEFAULT_CS = 0x100
	DEFAULT_SS = 0x200
	DEFAULT_DS = 0x1000
	DEFAULT_ES = 0x2000
)

func SetupCPU(t *testing.T, opcodes string) *CPU {
	inst, err := hex.DecodeString(opcodes)
	if err != nil {
		t.Errorf("failed to parse: %s", opcodes)
	}

	cpu := NewCpu(1024 * 1024)
	cpu.Regs.SetSeg16(CS, DEFAULT_CS)
	cpu.Regs.SetSeg16(SS, DEFAULT_SS)
	cpu.Regs.SetSeg16(DS, DEFAULT_DS)
	cpu.Regs.SetSeg16(ES, DEFAULT_ES)
	cpu.Regs.SetReg16(SP, 0x00FF)
	copy(cpu.Mem.At(uint(cpu.Regs.GetSeg16(CS)), uint(cpu.Ip)), inst)
	return cpu
}

func checkFlags(t *testing.T, cpu *CPU, want_flags string) {
	for _, c := range want_flags {
		switch c {
		case 'A':
			assert.Assert(t, cpu.Flags.IsEnabled(AF))
		case 'a':
			assert.Assert(t, !cpu.Flags.IsEnabled(AF))
		case 'C':
			assert.Assert(t, cpu.Flags.IsEnabled(CF))
		case 'c':
			assert.Assert(t, !cpu.Flags.IsEnabled(CF))
		case 'D':
			assert.Assert(t, cpu.Flags.IsEnabled(DF))
		case 'd':
			assert.Assert(t, !cpu.Flags.IsEnabled(DF))
		case 'I':
			assert.Assert(t, cpu.Flags.IsEnabled(IF))
		case 'i':
			assert.Assert(t, !cpu.Flags.IsEnabled(IF))
		case 'O':
			assert.Assert(t, cpu.Flags.IsEnabled(OF))
		case 'o':
			assert.Assert(t, !cpu.Flags.IsEnabled(OF))
		case 'P':
			assert.Assert(t, cpu.Flags.IsEnabled(PF))
		case 'p':
			assert.Assert(t, !cpu.Flags.IsEnabled(PF))
		case 'S':
			assert.Assert(t, cpu.Flags.IsEnabled(SF))
		case 's':
			assert.Assert(t, !cpu.Flags.IsEnabled(SF))
		case 'Z':
			assert.Assert(t, cpu.Flags.IsEnabled(ZF))
		case 'z':
			assert.Assert(t, !cpu.Flags.IsEnabled(ZF))
		default:
			t.Logf("unknown want_flag value: '%c'; decimal: '%d'", c, c)
			t.FailNow()
		}
	}
}

type h interface{}
type w interface{}

// a register with value for want/have
type regval8 struct {
	reg Reg8
	val uint
}

type regval16 struct {
	reg Reg
	val uint
}

type sregval struct {
	reg SReg
	val uint
}

// Set or clear a flag, only used in have
type flagval struct {
	mask uint32
	set  bool
}

// a memory address with value for want/have
type memval8 struct {
	seg, off uint
	val      uint8
}

// a memory address with value for want/have
type memval16 struct {
	seg, off uint
	val      uint16
}

// A value is pushed onto the stack
type pushval16 struct {
	val uint16
}

// A value is popped off the stack
type popval struct {
	val uint16
}

// Value wanted for instruction pointer (IP)
type ipval struct {
	val uint
}

// a memory address with value for want/have
// type memslice struct {
// 	seg, off int
// 	val      []byte
// }

func TestCPU(t *testing.T) {
	type testinfo struct {
		descr      string
		opcodes    string
		have       []h
		want       []w
		want_flags string
	}
	tests := []testinfo{
		// 0x33e
		{"ADD0x81/0", "81C43E03", []h{regval16{SP, 0x0800}}, []w{regval16{SP, 0x0B3E}}, "zc"},

		// CLI STI
		{"CLI", "FA", []h{flagval{IF, true}}, []w{}, "i"},
		{"STI", "FB", []h{flagval{IF, false}}, []w{}, "I"},

		// XOR (abbreviated)
		{"XOR/Z", "3401", []h{regval8{AL, 0x1}}, []w{regval8{AL, 0x00}}, "Z"},
		{"XOR/ALimm8", "3503", []h{regval8{AL, 0x1}}, []w{regval8{AL, 0x2}}, ""},
		{"AND/AXimm16", "350300", []h{regval16{AX, 0x1}}, []w{regval16{AX, 0x2}}, ""},
		// SHL (abbreviated)
		{"SHL/ALimm8", "C0E002", []h{regval8{AL, 0x1}}, []w{regval8{AL, 0x4}}, ""},
		{"SHL/AXimm16", "C1E002", []h{regval16{AX, 0x1}}, []w{regval16{AX, 0x4}}, ""},
		// SAR (abbreviated)
		{"SAR/ALimm8/NEG", "C0F802", []h{regval16{AX, 0x80}}, []w{regval16{AX, 0xE0}}, ""},
		{"SAR/ALimm8", "C0F802", []h{regval8{AL, 0x8}}, []w{regval8{AL, 0x2}}, ""},
		{"SAR/AXimm16", "C1F80200", []h{regval16{AX, 0x8}}, []w{regval16{AX, 0x2}}, ""},
		// SHR (abbreviated)
		{"SHR/ALimm8", "C0E802", []h{regval8{AL, 0x8}}, []w{regval8{AL, 0x2}}, ""},
		{"SHR/AXimm16", "C1E802", []h{regval16{AX, 0x8}}, []w{regval16{AX, 0x2}}, ""},
		{"SHR/AXimm16/NEG", "C1E80200", []h{regval16{AX, 0x80}}, []w{regval16{AX, 0x20}}, ""},
		// NEG (abbreviated)
		{"NEG/AL", "F6D8", []h{regval8{AL, 0xFF}}, []w{regval8{AL, 0x01}}, ""},
		{"NEG/AX", "F7D8", []h{regval16{AX, 0xFFFF}}, []w{regval16{AX, 0x0001}}, ""},

		// JMP: EBF9: JMP -7
		{"JMP/8+", "EB07", []h{}, []w{ipval{9}}, ""},
		{"JMP/8-", "EBF9", []h{}, []w{ipval{65531}}, ""},
		{"JMP/16+", "E90701", []h{}, []w{ipval{0x010A}}, ""},
		{"JMP/16-", "E9F9FD", []h{}, []w{ipval{0xFDFC}}, ""},
		{"JCXZ/Y/8", "E307", []h{regval16{CX, 0}}, []w{ipval{9}}, ""},
		{"JCXZ/N/8", "E307", []h{regval16{CX, 20}}, []w{ipval{2}}, ""},
		// FF /4 - short
		{"JMP/FF/4", "FF26CC20",
			[]h{memval16{DEFAULT_DS, 0x20CC, 0x7}, memval16{DEFAULT_DS, 0x20CE, DEFAULT_DS}},
			[]w{ipval{0x7}, sregval{CS, DEFAULT_CS}}, ""},
		// FF /5 - far
		{"JMP/FF/5", "FF2ECC20",
			[]h{memval16{DEFAULT_DS, 0x20CC, 0x7}, memval16{DEFAULT_DS, 0x20CE, DEFAULT_DS}},
			[]w{ipval{0x7}, sregval{CS, DEFAULT_DS}}, ""},

		// LOOP
		{"LOOP", "E207", []h{regval16{CX, 20}}, []w{ipval{9}}, ""},
		{"LOOP/CX0", "E207", []h{regval16{CX, 1}}, []w{ipval{2}}, ""},
		{"LOOPE", "E107", []h{regval16{CX, 20}, flagval{ZF, true}}, []w{ipval{9}}, ""},
		{"LOOPE/CX0", "E107", []h{regval16{CX, 20}, flagval{ZF, false}}, []w{ipval{2}}, ""},
		{"LOOPNE", "E007", []h{regval16{CX, 20}, flagval{ZF, false}}, []w{ipval{9}}, ""},
		{"LOOPNE/CX0", "E007", []h{regval16{CX, 20}, flagval{ZF, true}}, []w{ipval{2}}, ""},

		// ADD
		{"ADD0x00", "00C2", []h{regval8{AL, 3}, regval8{DL, 2}}, []w{regval8{DL, 5}}, "Paczso"},
		{"ADD0x00/AF", "00C2", []h{regval8{AL, 0x0F}, regval16{DX, 0x01}}, []w{regval8{DL, 0x10}}, "Acozsp"},
		{"ADD0x00/CFSF", "00C2", []h{regval16{AX, 0x7F}, regval16{DX, 0x01}}, []w{regval8{DL, 0x80}}, "zSOpcA"},
		{"ADD0x01", "01C0", []h{regval16{AX, 0x10}}, []w{regval16{AX, 0x20}}, "zsoc"},
		{"ADD0x02", "02C0", []h{regval8{AL, 0x01}}, []w{regval8{AL, 0x02}}, "zsoc"},
		{"ADD0x03", "03C0", []h{regval16{AX, 0x10}}, []w{regval16{AX, 0x20}}, "zsoc"},

		{"ADD0x04/IB", "0411", []h{regval8{AL, 0x10}}, []w{regval8{AL, 0x21}}, "zsoc"},
		{"ADD0x05/IW", "041100", []h{regval16{AX, 0x10}}, []w{regval16{AX, 0x21}}, "zsoc"},

		{"ADD0x80/0", "80C011", []h{regval8{AL, 0x10}}, []w{regval8{AL, 0x21}}, "zsoc"},
		{"ADD0x81/0", "81C01100", []h{regval8{AL, 0x10}}, []w{regval8{AL, 0x21}}, "zsoc"},
		{"ADD0x83/0", "83C011", []h{regval16{AX, 0x10}}, []w{regval16{AX, 0x21}}, "zsoc"},

		// SUB

		{"SUB/AL/imm8", "2C02", []h{regval8{AL, 5}}, []w{regval8{AL, 0x03}}, ""},
		{"SUB/AL/imm8/Z", "2C05", []h{regval8{AL, 5}}, []w{regval8{AL, 0x00}}, "Z"},
		{"SUB/AX/imm16", "2D02", []h{regval16{AX, 5}}, []w{regval16{AX, 0x03}}, ""},
		{"SUB/rmm8/imm8", "80E802", []h{regval8{AL, 5}}, []w{regval8{AL, 0x03}}, ""},
		{"SUB/rmm16/imm16", "80E80200", []h{regval16{AX, 5}}, []w{regval16{AX, 0x03}}, ""},
		{"SUB/rmm16/imm8", "80E802", []h{regval16{AX, 5}}, []w{regval16{AX, 0x03}}, ""},
		// B is RMM, A is R
		{"SUB/rmm8/r8", "28C3", []h{regval8{AL, 5}, regval8{BL, 8}}, []w{regval8{BL, 0x03}}, ""},
		{"SUB/rmm16/r16", "29C3", []h{regval16{AX, 5}, regval16{BX, 8}}, []w{regval16{BX, 0x03}}, ""},

		{"SUB/r8/rmm8", "2AC3", []h{regval8{AL, 0xFE}, regval8{BL, 0x7F}}, []w{regval8{AL, 0x7F}}, "so"},
		{"SUB/r16/rmm16", "2BC3", []h{regval16{AX, 0xFFFE}, regval16{BX, 0xFF7F}}, []w{regval16{AX, 0x007F}}, "so"},

		// MOV

		{"MOV0xB8", "B8FECA", []h{regval16{AX, 21}}, []w{regval16{AX, 0xCAFE}}, ""},
		// MOV r/m8,r8
		{"MOV0x88/07", "8807", []h{regval8{AL, 0xAB}, regval16{BX, 0x10}},
			[]w{memval8{DEFAULT_DS, 0x10, 0xAB}}, ""},
		// MOV r/m16,r16, move [DS:[BX]], AX
		{"MOV0x89/07", "8907", []h{regval16{AX, 0xABCD}, regval16{BX, 0x10}},
			[]w{memval16{DEFAULT_DS, 0x10, 0xABCD}}, ""},
		// MOV MOV r8,r/m8, MOV AL, [DS:[BL]]
		{"MOV0x8A/07", "8A07", []h{memval8{DEFAULT_DS, 0x10, 0xBC}, regval8{BL, 0x10}},
			[]w{regval8{AL, 0xBC}}, ""},
		// MOV r16,r/m16 - MOV AX, [DS:[BX]]
		{"MOV0x8B/07", "8B07", []h{memval16{DEFAULT_DS, 0x10, 0xABCD}, regval16{BX, 0x10}},
			[]w{regval16{AX, 0xABCD}}, ""},
		// MOV r/m16,Sreg** [8C] (MOV AX, DS)
		{"MOV0x8C/07", "8CD8", []h{}, []w{regval16{AX, DEFAULT_DS}}, ""},

		// [A0] MOV AL,moffs8* - Move byte at (seg:offset) to AL.
		{"MOV0xA0", "A00200", []h{memval8{DEFAULT_DS, 0x02, 0xAA}},
			[]w{regval8{AL, 0xAA}}, ""},
		// [A1] MOV AX,moffs16* - Move word at (seg:offset) to AX.
		{"MOV0xA1", "A10200", []h{memval16{DEFAULT_DS, 0x02, 0xAABB}},
			[]w{regval16{AX, 0xAABB}}, ""},

		// [A2] MOV moffs8,AL- Move AL to (seg:offset).
		{"MOV0xA2", "A20200", []h{regval8{AL, 0xAA}},
			[]w{memval8{DEFAULT_DS, 0x02, 0xAA}}, ""},
		// [A3] MOV moffs16*,AX - Move AX to (seg:offset).
		{"MOV0xA3", "A30200", []h{regval16{AX, 0xAABB}},
			[]w{memval16{DEFAULT_DS, 0x02, 0xAABB}}, ""},

		// [C6] MOV r/m8, imm8 - Move imm8 to r/m8. - Move to CL (using C1 ModR/M)
		{"MOV0xC6", "C6C121", []h{},
			[]w{regval8{CL, 0x21}}, ""},
		// [C7] MOV r/m16, imm16 - Move imm16 to r/m16. - Move to CX (using C1 ModR/M)
		{"MOV0xC7", "C7C1BE21", []h{},
			[]w{regval16{CX, 0x21BE}}, ""},

		// CD21 Call Interrupt 21
		{"CD/21", "CD21", []h{}, []w{}, ""},

		// LEA: 8D160A00: LEA DX,[0x0A]
		{"LEA/21", "8D160A00", []h{}, []w{regval16{DX, 0x000A}}, ""},

		// CALL: 9A0010FF00: Address IP:CS in IMM
		{"CALL/FAR/9A", "9AFF000010", []h{},
			[]w{ipval{0x000FF}, sregval{CS, 0x1000}}, ""},
		// CALL: E80001: Address is 0x0100
		{"CALL/E8", "E80001", []h{}, []w{ipval{0x0103}}, ""},
		// CALL: FFD3: Address in BX
		{"CALL/NEAR/FF/2", "FFD3",
			[]h{
				regval16{BX, 0x0100},
			},
			// This u	sed to be 0x0102, but I think that was wrong
			// Since this is an absolute and not relative offset
			[]w{ipval{0x0100}}, ""},
		{"CALL/FAR/FF/3", "FF1F",
			[]h{
				regval16{BX, 0x100},
				memval16{DEFAULT_DS, 0x100, 0xAAAA},
				memval16{DEFAULT_DS, 0x102, 0x00BB},
			},
			[]w{
				ipval{0xAAAA},
				sregval{CS, 0x00BB},
			}, ""},
		// INC
		{"INC/RM8", "FEC1", []h{regval8{CL, 0xFE}},
			[]w{regval8{CL, 0xFF}}, ""},
		{"INC/RM16", "FFC1", []h{regval16{CX, 0x00FF}},
			[]w{regval16{CX, 0x0100}}, ""},
		// DEC
		{"DEC/RM8", "FEC9", []h{regval8{CL, 0xFF}},
			[]w{regval8{CL, 0xFE}}, ""},
		{"DEC/RM16", "FFC9", []h{regval16{CX, 0x0100}},
			[]w{regval16{CX, 0x00FF}}, ""},

		{"CMP/AL/IMM8/LT", "3C22", []h{regval8{AL, 0x21}},
			[]w{}, "zoSC"},
		{"CMP/AL/IMM8/EQ", "3C21", []h{regval8{AL, 0x21}},
			[]w{}, "Zosac"},
		{"CMP/AL/IMM8/GT", "3C20", []h{regval8{AL, 0x21}},
			[]w{}, "zosac"},
		{"CMP/AX/IMM16", "3D00FF", []h{regval16{AX, 0x00}},
			[]w{}, "z"},
		{"CMP/AH/AH/OF", "38E4", []h{regval8{AH, 0xff}},
			[]w{}, "Zco"},

		{"SCASB/REPNE", "F2AE",
			[]h{sregval{ES, 0x2000},
				regval16{DI, 0x0100},
				regval8{AL, 0x0021},
				regval16{CX, 0x00FF},
				memval8{0x2000, 0x100, 0x11},
				memval8{0x2000, 0x101, 0x21},
			},
			[]w{regval16{DI, 0x0102}}, ""},
		{"SCASB/NE", "AE",
			[]h{sregval{ES, 0x2000},
				regval16{DI, 0x0100},
				regval8{AL, 0x0010},
				memval8{0x2000, 0x100, 0x21},
			},
			[]w{}, "z"},
		{"SCASB/EQ", "AE",
			[]h{sregval{ES, 0x2000},
				regval16{DI, 0x0100},
				regval8{AL, 0x0021},
				memval8{0x2000, 0x100, 0x21},
			},
			[]w{}, "Z"},
		{"RET/NEAR", "C3", []h{pushval16{0x2021}}, []w{ipval{0x2021}}, ""},
		{"RET/NEAR/POPBYTES", "C202", []h{pushval16{0xBEEF}, pushval16{0x2021}},
			[]w{ipval{0x2021}, regval16{SP, 0x00FF}}, ""},
		{"RET/FAR", "CB", []h{pushval16{0xCAFE}, pushval16{0x2021}},
			[]w{ipval{0x2021}, sregval{CS, 0xCAFE}}, ""},

		{"PUSH/IMM", "68EFBE", []h{}, []w{popval{0xBEEF}}, ""},
		{"POP/ES", "07", []h{pushval16{0x2021}}, []w{sregval{ES, 0x2021}}, ""},
		{"POP/RMM/AX", "8FC0", []h{pushval16{0x2021}}, []w{regval16{AX, 0x2021}}, ""},
		{"LES", "C43E7500", []h{memval16{DEFAULT_DS, 0x0075, 0x2021}, memval16{DEFAULT_DS, 0x0077, DEFAULT_DS}},
			[]w{regval16{DI, 0x2021}, sregval{ES, DEFAULT_DS}}, ""},
		{"LDDS", "C53E7500", []h{memval16{DEFAULT_DS, 0x0075, 0x2021}, memval16{DEFAULT_DS, 0x0077, 0x2000}},
			[]w{regval16{DI, 0x2021}, sregval{DS, 0x2000}}, ""},
		{"CLD", "FC", []h{flagval{DF, true}}, []w{}, "d"},
		{"STD", "FD", []h{flagval{DF, false}}, []w{}, "D"},

		// OR
		{"OR/ALimm8", "0C02", []h{regval8{AL, 0x1}}, []w{regval8{AL, 0x3}}, ""},
		{"OR/AXimm16", "0D0200", []h{regval16{AX, 0x1}}, []w{regval16{AX, 0x3}}, ""},
		{"OR/80/1/imm8", "80C802", []h{regval8{AL, 0x1}}, []w{regval8{AL, 0x3}}, ""},
		{"OR/81/1/imm16", "81C80200", []h{regval16{AX, 0x1}}, []w{regval16{AX, 0x3}}, ""},
		{"OR/83/1/imm8", "83C802", []h{regval16{AX, 0x1}}, []w{regval16{AX, 0x3}}, ""},
		{"OR/08/rmm8/r8", "08D1", []h{regval8{DL, 0x1}, regval8{CL, 0x2}}, []w{regval8{CL, 0x3}}, ""},
		{"OR/09/rmm16/r16", "09D1", []h{regval16{DX, 0x1}, regval16{CX, 0x2}}, []w{regval16{CX, 0x3}}, ""},
		{"OR/0A/r8/rmm8", "0AD1", []h{regval8{DL, 0x1}, regval8{CL, 0x2}}, []w{regval8{DL, 0x3}}, ""},
		{"OR/0B/r16/rmm16", "0BD1", []h{regval16{DX, 0x1}, regval16{CX, 0x2}}, []w{regval16{DX, 0x3}}, ""},

		// AND
		{"AND/Z", "24F0", []h{regval8{AL, 0xF}}, []w{regval8{AL, 0x00}}, "Z"},
		{"AND/ALimm8", "2403", []h{regval8{AL, 0xF}}, []w{regval8{AL, 0x3}}, ""},
		{"AND/AXimm16", "250300", []h{regval16{AX, 0x2}}, []w{regval16{AX, 0x2}}, ""},
		{"AND/80/1/imm8", "80E003", []h{regval8{AL, 0xF}}, []w{regval8{AL, 0x3}}, ""},
		{"AND/81/1/imm16", "81E00300", []h{regval16{AX, 0xF}}, []w{regval16{AX, 0x3}}, ""},
		{"AND/83/1/imm8", "83E003", []h{regval16{AX, 0xF}}, []w{regval16{AX, 0x3}}, ""},
		{"AND/20/rmm8/r8", "20D1", []h{regval8{DL, 0xF}, regval8{CL, 0x3}}, []w{regval8{CL, 0x3}}, ""},
		{"AND/21/rmm16/r16", "21D1", []h{regval16{DX, 0xF}, regval16{CX, 0x3}}, []w{regval16{CX, 0x3}}, ""},
		{"AND/22/r8/rmm8", "22D1", []h{regval8{DL, 0xF}, regval8{CL, 0x3}}, []w{regval8{DL, 0x3}}, ""},
		{"AND/23/r16/rmm16", "23D1", []h{regval16{DX, 0xF}, regval16{CX, 0x3}}, []w{regval16{DX, 0x3}}, ""},

		// STOSB
		{"STOSB", "AA", []h{regval8{AL, 0x21}, sregval{ES, DEFAULT_ES}, regval16{DI, 0x0000}},
			[]w{memval8{DEFAULT_ES, 0x0, 0x21}, regval16{DI, 0x0001}}, ""},
		{"STOSW", "AB", []h{regval16{AX, 0x1234}, sregval{ES, DEFAULT_ES}, regval16{DI, 0x0000}},
			[]w{memval16{DEFAULT_ES, 0x0, 0x1234}, regval16{DI, 0x0002}}, ""},
		// STOSB
		{"LODSB", "AC",
			[]h{memval8{DEFAULT_DS, 0x0, 0x41}, sregval{DS, DEFAULT_DS}, regval16{SI, 0x0000}},
			[]w{regval8{AL, 0x41}, regval16{SI, 0x01}}, ""},
		{"LODSW", "AD",
			[]h{memval16{DEFAULT_DS, 0x0, 0x1941}, sregval{DS, DEFAULT_DS}, regval16{SI, 0x0000}},
			[]w{regval16{AX, 0x1941}, regval16{SI, 0x02}}, ""},
		{"STC", "F9",
			[]h{flagval{CF, false}},
			[]w{}, "C"},
		{"CLC", "F8",
			[]h{flagval{CF, true}},
			[]w{}, "c"},
		{"XCHG", "93", []h{regval16{AX, 0xA}, regval16{BX, 0xB}},
			[]w{regval16{AX, 0xB}, regval16{BX, 0xA}}, ""},
		// NOT is F7/2. NOT AX
		{"NOT", "F7D0",
			[]h{regval16{AX, 0x0001}},
			[]w{regval16{AX, 0xFFFE}}, ""},
		{"TEST/AX/IMM16/z", "A90180",
			[]h{regval16{AX, 0x0001}},
			[]w{}, "zps"},
		{"TEST/AX/IMM16/z", "A90180",
			[]h{regval16{AX, 0x0002}},
			[]w{}, "ZPs"},
		{"TEST/AX/IMM16/S", "A90180",
			[]h{regval16{AX, 0x8000}},
			[]w{}, "zSP"},

		{"CWD/POS", "99",
			[]h{regval16{AX, 0x8000}},
			[]w{regval16{DX, 0xFFFF}}, ""},
		{"CWD/NEG", "99",
			[]h{regval16{AX, 0x7FFF}},
			[]w{regval16{DX, 0x0}}, ""},
		{"CMPSB/EQ", "A6",
			[]h{sregval{DS, DEFAULT_DS}, regval16{SI, 0},
				sregval{ES, DEFAULT_ES}, regval16{DI, 0},
				memval8{DEFAULT_DS, 0, 42}, memval8{DEFAULT_ES, 0, 42}},
			[]w{}, "Z"},

		// TODO
	}

	for r := AL; r <= BH; r++ {
		// Append B0-B7 for MOV r8, imm8 - Move imm8 to r8.
		tests = append(tests,
			testinfo{fmt.Sprintf("MOV0x%x/%v", uint8(r), r),
				fmt.Sprintf("%02XAA", 0xB0+uint8(r)-uint8(AL)),
				[]h{},
				[]w{regval8{r, 0xAA}}, ""})

	}
	for r := AX; r <= BX; r++ {
		// Append B8-BF for MOV r16, imm16- Move imm16 to r16.
		tests = append(tests,
			testinfo{fmt.Sprintf("MOV0x%x/%v", uint8(r), r),
				fmt.Sprintf("%02XEFBE", 0xB8+uint8(r)-uint8(AX)),
				[]h{},
				[]w{regval16{r, 0xBEEF}}, ""})
	}

	for _, test := range tests {
		descr := fmt.Sprintf("%s:{%s}", test.descr, test.opcodes)
		t.Run(descr, func(t *testing.T) {
			cpu := SetupCPU(t, test.opcodes)
			for _, r := range test.have {
				switch r := r.(type) {
				case flagval:
					cpu.Flags.SetFlagIf(r.mask, r.set)
				case regval8:
					cpu.Regs.SetReg8(r.reg, r.val)
				case regval16:
					cpu.Regs.SetReg16(r.reg, r.val)
				case sregval:
					cpu.Regs.SetSeg16(r.reg, r.val)
				case memval8:
					cpu.Mem.SetMem8(r.seg, r.off, r.val)
				case memval16:
					cpu.Mem.SetMem16(r.seg, r.off, r.val)
				case pushval16:
					cpu.Regs.Push16(cpu.Mem, r.val)
				default:
					t.Fatalf("Unknown have: %#v", r)
				}
			}
			assert.NilError(t, cpu.RunOnce(), "RunOnce Failed")

			for _, r := range test.want {
				switch r := r.(type) {
				case regval8:
					assert.Equal(t, r.val, cpu.Regs.GetReg8(r.reg), r.reg)
				case regval16:
					h := cpu.Regs.GetReg16(r.reg)
					assert.Equal(t, r.val, h,
						fmt.Sprintf("want: %#v; have: %#v", r.val, h))
				case sregval:
					assert.Equal(t, r.val, cpu.Regs.GetSeg16(r.reg), r.reg)
				case memval8:
					actual := cpu.Mem.GetMem8(r.seg, r.off)
					assert.Equal(t, actual, r.val, r)
				case memval16:
					actual := cpu.Mem.GetMem16(r.seg, r.off)
					if actual != r.val {
						fmt.Printf("%#v", r)
					}
					assert.Equal(t, actual, r.val, r)
				case popval:
					actual := cpu.Regs.Pop16(cpu.Mem)
					assert.Equal(t, actual, r.val)
				case ipval:
					assert.Equal(t, uint16(r.val), cpu.Ip, r.val)
				default:
					t.Fatalf("Unknown want: %#v", r)
				}
			}
			checkFlags(t, cpu, test.want_flags)
		})
	}
}

/*
func TestCPUJumps8(t *testing.T) {
	type testinfo struct {
		op      Op
		b       byte
		yescond uint32
		nocond  uint32
	}

	tests := []testinfo{
		{JA, 0x77, ^(CF | ZF), CF},
		{JA, 0x77, ^(CF | ZF), ZF},
		{JA, 0x77, ^(CF | ZF), CF | ZF},
		// JMP: EBF9: JMP -7
		{JAE, 0x73, ^CF, CF},
		{JB, 0x73, ^CF, CF},
		{JBE, 0x76, CF, 0},
		{JBE, 0x76, ZF, 0},
		{JE, 0x74, ZF, 0},
		{JG, 0x7F, SF | OF, ZF},
		{JG, 0x7F, 0, SF},
		{JGE, 0x7D, SF | OF, SF},
		{JGE, 0x7D, ZF | SF | OF, SF},
		{JL, 0x7C, SF, SF | OF},
		{JL, 0x7C, OF, SF | OF},
		{JLE, 0x7E, OF, SF | OF},
		{JLE, 0x7E, SF, SF | OF},
		{JLE, 0x7E, ZF, SF | OF},
		{JNE, 0x75, 0, ZF},
		{JNO, 0x71, 0, OF},
		{JNP, 0x7B, 0, PF},
		{JNS, 0x79, 0, SF},
		{JO, 0x70, OF, 0},
		{JP, 0x7A, PF, 0},
		{JS, 0x78, SF, 0},
	}

	for _, test := range tests {
		descr := fmt.Sprintf("%v:{%X}", test.op, test.b)
		opcodes := fmt.Sprintf("%02X07", test.b)
		t.Run(fmt.Sprintf("%s/Y", descr), func(t *testing.T) {
			cpu := SetupCPU(t, opcodes)
			cpu.Flags.ReplaceAllFlags(test.yescond)
			assert.Equal(t, true, cpu.RunOnce())
			assert.Equal(t, cpu.Ip, uint16(9))
		})
		t.Run(fmt.Sprintf("%s/N", descr), func(t *testing.T) {
			cpu := SetupCPU(t, opcodes)
			cpu.Flags.ReplaceAllFlags(test.nocond)
			assert.Equal(t, true, cpu.RunOnce())
			assert.Equal(t, cpu.Ip, uint16(2))
		})
	}
}
*/

func TestCpuRegs(t *testing.T) {
	var tests = []struct {
		w Reg
		l Reg8
		h Reg8
	}{
		{AX, AL, AH},
		{BX, BL, BH},
		{CX, CL, CH},
		{DX, DL, DH},
	}
	for _, test := range tests {
		t.Run(fmt.Sprintf("%v/%v/%v", test.w, test.l, test.h), func(t *testing.T) {
			cpu := SetupCPU(t, "00C2")
			cpu.Regs.SetReg16(test.w, 0xBEEF)
			assert.Equal(t, cpu.Regs.GetReg8(test.l), uint(0xEF))
			assert.Equal(t, cpu.Regs.GetReg8(test.h), uint(0xBE))

			cpu.Regs.SetReg8(test.h, 0xCA)
			cpu.Regs.SetReg8(test.l, 0xFE)
			assert.Equal(t, cpu.Regs.GetReg16(test.w), uint(0xCAFE))
		})
	}
}
