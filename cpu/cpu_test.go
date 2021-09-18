package go86

import (
	"encoding/hex"
	"fmt"
	"testing"

	"golang.org/x/arch/x86/x86asm"
	"gotest.tools/v3/assert"
)

func TestCpuSmoke(t *testing.T) {
	cpu := NewCpu(1024 * 1024)
	assert.Equal(t, cpu.Ip, uint16(0))
}

const (
	CS = 0x100
	SS = 0x200
	DS = 0x1000
	ES = 0x2000
)

func SetupCPU(t *testing.T, opcodes string) *CPU {
	inst, err := hex.DecodeString(opcodes)
	if err != nil {
		t.Errorf("failed to parse: %s", opcodes)
	}

	cpu := NewCpu(1024 * 1024)
	cpu.PutReg(x86asm.CS, CS)
	cpu.PutReg(x86asm.SS, SS)
	cpu.PutReg(x86asm.DS, DS)
	cpu.PutReg(x86asm.ES, ES)
	cpu.PutReg(x86asm.SP, 0x00FF)
	copy(cpu.Mem.At(CS, int(cpu.Ip)), inst)
	return cpu
}

func checkFlags(t *testing.T, cpu *CPU, want_flags string) {
	for _, c := range want_flags {
		switch c {
		case 'A':
			assert.Assert(t, cpu.Flags&AF != 0)
		case 'a':
			assert.Assert(t, cpu.Flags&AF == 0)
		case 'C':
			assert.Assert(t, cpu.Flags&CF != 0)
		case 'c':
			assert.Assert(t, cpu.Flags&CF == 0)
		case 'I':
			assert.Assert(t, cpu.Flags&IF != 0)
		case 'i':
			assert.Assert(t, cpu.Flags&IF == 0)
		case 'O':
			assert.Assert(t, cpu.Flags&OF != 0)
		case 'o':
			assert.Assert(t, cpu.Flags&OF == 0)
		case 'P':
			assert.Assert(t, cpu.Flags&PF != 0)
		case 'p':
			assert.Assert(t, cpu.Flags&PF == 0)
		case 'S':
			assert.Assert(t, cpu.Flags&SF != 0)
		case 's':
			assert.Assert(t, cpu.Flags&SF == 0)
		case 'Z':
			assert.Assert(t, cpu.Flags&ZF != 0)
		case 'z':
			assert.Assert(t, cpu.Flags&ZF == 0)
		}
	}
}

type h interface{}
type w interface{}

// a register with value for want/have
type regval struct {
	reg x86asm.Reg
	val uint
}

// Set or clear a flag, only used in have
type flagval struct {
	mask uint16
	set  bool
}

// a memory address with value for want/have
type memval8 struct {
	seg, off int
	val      uint8
}

// a memory address with value for want/have
type memval16 struct {
	seg, off int
	val      uint16
}

// A value is pushed onto the stack
type pushval struct {
	val uint16
}

// A value is popped off the stack
type popval struct {
	val uint16
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
		{"ADD0x81/0", "81C43E03", []h{regval{x86asm.SP, 0x0800}}, []w{regval{x86asm.SP, 0x0B3E}}, "zc"},

		// CLI STI
		{"CLI", "FA", []h{flagval{IF, true}}, []w{}, "i"},
		{"STI", "FB", []h{flagval{IF, false}}, []w{}, "I"},

		// XOR (abbreviated)
		{"XOR/Z", "3401", []h{regval{x86asm.AL, 0x1}}, []w{regval{x86asm.AL, 0x00}}, "Z"},
		{"XOR/ALimm8", "3503", []h{regval{x86asm.AL, 0x1}}, []w{regval{x86asm.AL, 0x2}}, ""},
		{"AND/AXimm16", "350300", []h{regval{x86asm.AX, 0x1}}, []w{regval{x86asm.AX, 0x2}}, ""},
		// SHL (abbreviated)
		{"SHL/ALimm8", "C0E002", []h{regval{x86asm.AL, 0x1}}, []w{regval{x86asm.AL, 0x4}}, ""},
		{"SHL/AXimm16", "C1E002", []h{regval{x86asm.AX, 0x1}}, []w{regval{x86asm.AX, 0x4}}, ""},
		// SAR (abbreviated)
		{"SAR/ALimm8/NEG", "C0F802", []h{regval{x86asm.AX, 0x80}}, []w{regval{x86asm.AX, 0xA0}}, ""},
		{"SAR/ALimm8", "C0F802", []h{regval{x86asm.AL, 0x8}}, []w{regval{x86asm.AL, 0x2}}, ""},
		{"SAR/AXimm16", "C1F80200", []h{regval{x86asm.AX, 0x8}}, []w{regval{x86asm.AX, 0x2}}, ""},
		// SHR (abbreviated)
		{"SHR/ALimm8", "C0E802", []h{regval{x86asm.AL, 0x8}}, []w{regval{x86asm.AL, 0x2}}, ""},
		{"SHR/AXimm16", "C1E802", []h{regval{x86asm.AX, 0x8}}, []w{regval{x86asm.AX, 0x2}}, ""},
		{"SHR/AXimm16/NEG", "C1E80200", []h{regval{x86asm.AX, 0x80}}, []w{regval{x86asm.AX, 0x20}}, ""},
		// NEG (abbreviated)
		{"NEG/AL", "F6D8", []h{regval{x86asm.AL, 0xFF}}, []w{regval{x86asm.AL, 0x01}}, ""},
		{"NEG/AX", "F7D8", []h{regval{x86asm.AX, 0xFFFF}}, []w{regval{x86asm.AL, 0x0001}}, ""},

		// JMP: EBF9: JMP -7
		{"JMP/8+", "EB07", []h{}, []w{regval{x86asm.IP, 9}}, ""},
		{"JMP/8-", "EBF9", []h{}, []w{regval{x86asm.IP, 65531}}, ""},
		{"JMP/16+", "E90701", []h{}, []w{regval{x86asm.IP, 0x010A}}, ""},
		{"JMP/16-", "E9F9FD", []h{}, []w{regval{x86asm.IP, 0xFDFC}}, ""},
		{"JCXZ/Y/8", "E307", []h{regval{x86asm.CX, 0}}, []w{regval{x86asm.IP, 9}}, ""},
		{"JCXZ/N/8", "E307", []h{regval{x86asm.CX, 20}}, []w{regval{x86asm.IP, 2}}, ""},
		// FF /4 - short
		{"JMP/FF/4", "FF26CC20",
			[]h{memval16{DS, 0x20CC, 0x7}, memval16{DS, 0x20CE, DS}},
			[]w{regval{x86asm.IP, 0x7}, regval{x86asm.CS, CS}}, ""},
		// FF /5 - far
		{"JMP/FF/5", "FF2ECC20",
			[]h{memval16{DS, 0x20CC, 0x7}, memval16{DS, 0x20CE, DS}},
			[]w{regval{x86asm.IP, 0x7}, regval{x86asm.CS, DS}}, ""},

		// LOOP
		{"LOOP", "E207", []h{regval{x86asm.CX, 20}}, []w{regval{x86asm.IP, 9}}, ""},
		{"LOOP/CX0", "E207", []h{regval{x86asm.CX, 1}}, []w{regval{x86asm.IP, 2}}, ""},
		{"LOOPE", "E107", []h{regval{x86asm.CX, 20}, flagval{ZF, true}}, []w{regval{x86asm.IP, 9}}, ""},
		{"LOOPE/CX0", "E107", []h{regval{x86asm.CX, 20}, flagval{ZF, false}}, []w{regval{x86asm.IP, 2}}, ""},
		{"LOOPNE", "E007", []h{regval{x86asm.CX, 20}, flagval{ZF, false}}, []w{regval{x86asm.IP, 9}}, ""},
		{"LOOPNE/CX0", "E007", []h{regval{x86asm.CX, 20}, flagval{ZF, true}}, []w{regval{x86asm.IP, 2}}, ""},

		// ADD
		{"ADD0x00", "00C2", []h{regval{x86asm.AL, 3}, regval{x86asm.DL, 2}}, []w{regval{x86asm.DL, 5}}, "Paczsfo"},
		{"ADD0x00/AF", "00C2", []h{regval{x86asm.AL, 0x0F}, regval{x86asm.DX, 0x01}}, []w{regval{x86asm.DL, 0x10}}, "Acozsp"},
		{"ADD0x00/CFSF", "00C2", []h{regval{x86asm.AX, 0x7F}, regval{x86asm.DX, 0x01}}, []w{regval{x86asm.DL, 0x80}}, "zSOpcA"},
		{"ADD0x01", "01C0", []h{regval{x86asm.AX, 0x10}}, []w{regval{x86asm.AX, 0x20}}, "zsoc"},
		{"ADD0x02", "02C0", []h{regval{x86asm.AL, 0x01}}, []w{regval{x86asm.AL, 0x02}}, "zsoc"},
		{"ADD0x03", "03C0", []h{regval{x86asm.AX, 0x10}}, []w{regval{x86asm.AX, 0x20}}, "zsoc"},

		{"ADD0x04/IB", "0411", []h{regval{x86asm.AL, 0x10}}, []w{regval{x86asm.AL, 0x21}}, "zsoc"},
		{"ADD0x05/IW", "041100", []h{regval{x86asm.AX, 0x10}}, []w{regval{x86asm.AX, 0x21}}, "zsoc"},

		{"ADD0x80/0", "80C011", []h{regval{x86asm.AL, 0x10}}, []w{regval{x86asm.AL, 0x21}}, "zsoc"},
		{"ADD0x81/0", "81C01100", []h{regval{x86asm.AL, 0x10}}, []w{regval{x86asm.AL, 0x21}}, "zsoc"},
		{"ADD0x83/0", "83C011", []h{regval{x86asm.AX, 0x10}}, []w{regval{x86asm.AX, 0x21}}, "zsoc"},

		// SUB

		{"SUB/AL/imm8", "2C02", []h{regval{x86asm.AL, 5}}, []w{regval{x86asm.AL, 0x03}}, ""},
		{"SUB/AL/imm8/Z", "2C05", []h{regval{x86asm.AL, 5}}, []w{regval{x86asm.AL, 0x00}}, "Z"},
		{"SUB/AX/imm16", "2D02", []h{regval{x86asm.AX, 5}}, []w{regval{x86asm.AX, 0x03}}, ""},
		{"SUB/rmm8/imm8", "80E802", []h{regval{x86asm.AL, 5}}, []w{regval{x86asm.AL, 0x03}}, ""},
		{"SUB/rmm16/imm16", "80E80200", []h{regval{x86asm.AX, 5}}, []w{regval{x86asm.AX, 0x03}}, ""},
		{"SUB/rmm16/imm8", "80E802", []h{regval{x86asm.AX, 5}}, []w{regval{x86asm.AX, 0x03}}, ""},
		// B is RMM, A is R
		{"SUB/rmm8/r8", "28C3", []h{regval{x86asm.AL, 5}, regval{x86asm.BL, 8}}, []w{regval{x86asm.BL, 0x03}}, ""},
		{"SUB/rmm16/r16", "29C3", []h{regval{x86asm.AX, 5}, regval{x86asm.BX, 8}}, []w{regval{x86asm.BX, 0x03}}, ""},

		{"SUB/r8/rmm8", "2AC3", []h{regval{x86asm.AL, 0xFE}, regval{x86asm.BL, 0x7F}}, []w{regval{x86asm.AL, 0x7F}}, "sO"},
		{"SUB/r16/rmm16", "2BC3", []h{regval{x86asm.AX, 0xFFFE}, regval{x86asm.BX, 0xFF7F}}, []w{regval{x86asm.AX, 0x007F}}, "s0"},

		// MOV

		{"MOV0xB8", "B8FECA", []h{regval{x86asm.AX, 21}}, []w{regval{x86asm.AX, 0xCAFE}}, ""},
		// MOV r/m8,r8
		{"MOV0x88/07", "8807", []h{regval{x86asm.AL, 0xAB}, regval{x86asm.BX, 0x10}},
			[]w{memval8{DS, 0x10, 0xAB}}, ""},
		// MOV r/m16,r16, move [DS:[BX]], AX
		{"MOV0x89/07", "8907", []h{regval{x86asm.AX, 0xABCD}, regval{x86asm.BX, 0x10}},
			[]w{memval16{DS, 0x10, 0xABCD}}, ""},
		// MOV MOV r8,r/m8, MOV AL, [DS:[BL]]
		{"MOV0x8A/07", "8A07", []h{memval8{DS, 0x10, 0xBC}, regval{x86asm.BL, 0x10}},
			[]w{regval{x86asm.AL, 0xBC}}, ""},
		// MOV r16,r/m16 - MOV AX, [DS:[BX]]
		{"MOV0x8B/07", "8B07", []h{memval16{DS, 0x10, 0xABCD}, regval{x86asm.BX, 0x10}},
			[]w{regval{x86asm.AX, 0xABCD}}, ""},
		// MOV r/m16,Sreg** [8C] (MOV AX, DS)
		{"MOV0x8C/07", "8CD8", []h{}, []w{regval{x86asm.AX, DS}}, ""},

		// [A0] MOV AL,moffs8* - Move byte at (seg:offset) to AL.
		{"MOV0xA0", "A00200", []h{memval8{DS, 0x02, 0xAA}},
			[]w{regval{x86asm.AL, 0xAA}}, ""},
		// [A1] MOV AX,moffs16* - Move word at (seg:offset) to AX.
		{"MOV0xA1", "A10200", []h{memval16{DS, 0x02, 0xAABB}},
			[]w{regval{x86asm.AX, 0xAABB}}, ""},

		// [A2] MOV moffs8,AL- Move AL to (seg:offset).
		{"MOV0xA2", "A20200", []h{regval{x86asm.AL, 0xAA}},
			[]w{memval8{DS, 0x02, 0xAA}}, ""},
		// [A3] MOV moffs16*,AX - Move AX to (seg:offset).
		{"MOV0xA3", "A30200", []h{regval{x86asm.AX, 0xAABB}},
			[]w{memval16{DS, 0x02, 0xAABB}}, ""},

		// [C6] MOV r/m8, imm8 - Move imm8 to r/m8. - Move to CL (using C1 ModR/M)
		{"MOV0xC6", "C6C121", []h{},
			[]w{regval{x86asm.CL, 0x21}}, ""},
		// [C7] MOV r/m16, imm16 - Move imm16 to r/m16. - Move to CX (using C1 ModR/M)
		{"MOV0xC7", "C7C1BE21", []h{},
			[]w{regval{x86asm.CX, 0x21BE}}, ""},

		// CD21 Call Interrupt 21
		{"CD/21", "CD21", []h{}, []w{}, ""},

		// LEA: 8D160A00: LEA DX,[0x0A]
		{"LEA/21", "8D160A00", []h{}, []w{regval{x86asm.DX, 0x000A}}, ""},

		// CALL: 9A0010FF00: Address IP:CS in IMM
		{"CALL/FAR/9A", "9AFF000010", []h{},
			[]w{regval{x86asm.IP, 0x000FF}, regval{x86asm.CS, 0x1000}}, ""},
		// CALL: E80001: Address is 0x0100
		{"CALL/E8", "E80001", []h{}, []w{regval{x86asm.IP, 0x0103}}, ""},
		// CALL: FFD3: Address in BX
		{"CALL/NEAR/FF/2", "FFD3",
			[]h{
				regval{x86asm.BX, 0x0100},
			},
			[]w{regval{x86asm.IP, 0x0102}}, ""},
		{"CALL/FAR/FF/3", "FF1F",
			[]h{
				regval{x86asm.BX, 0x100},
				memval16{DS, 0x100, 0xAAAA},
				memval16{DS, 0x102, 0x00BB},
			},
			[]w{
				regval{x86asm.IP, 0xAAAA},
				regval{x86asm.CS, 0x00BB},
			}, ""},
		// INC
		{"INC/RM8", "FEC1", []h{regval{x86asm.CL, 0xFE}},
			[]w{regval{x86asm.CL, 0xFF}}, ""},
		{"INC/RM16", "FFC1", []h{regval{x86asm.CX, 0x00FF}},
			[]w{regval{x86asm.CX, 0x0100}}, ""},
		// DEC
		{"DEC/RM8", "FEC9", []h{regval{x86asm.CL, 0xFF}},
			[]w{regval{x86asm.CL, 0xFE}}, ""},
		{"DEC/RM16", "FFC9", []h{regval{x86asm.CX, 0x0100}},
			[]w{regval{x86asm.CX, 0x00FF}}, ""},

		{"CMP/AL/IMM8/LT", "3C22", []h{regval{x86asm.AL, 0x21}},
			[]w{}, "zoSC"},
		{"CMP/AL/IMM8/EQ", "3C21", []h{regval{x86asm.AL, 0x21}},
			[]w{}, "Zosac"},
		{"CMP/AL/IMM8/GT", "3C20", []h{regval{x86asm.AL, 0x21}},
			[]w{}, "zosac"},
		{"CMP/AX/IMM16", "3D00FF", []h{regval{x86asm.AX, 0x00}},
			[]w{}, "z"},
		{"CMP/AH/AH/OF", "38E4", []h{regval{x86asm.AH, 0xff}},
			[]w{}, "Zco"},

		{"SCASB/REPNE", "F2AE",
			[]h{regval{x86asm.ES, 0x2000},
				regval{x86asm.DI, 0x0100},
				regval{x86asm.AL, 0x0021},
				regval{x86asm.CX, 0x00FF},
				memval8{0x2000, 0x100, 0x11},
				memval8{0x2000, 0x101, 0x21},
			},
			[]w{regval{x86asm.DI, 0x0101}}, ""},
		{"SCASB/NE", "AE",
			[]h{regval{x86asm.ES, 0x2000},
				regval{x86asm.DI, 0x0100},
				regval{x86asm.AL, 0x0010},
				memval8{0x2000, 0x100, 0x21},
			},
			[]w{}, "z"},
		{"SCASB/EQ", "AE",
			[]h{regval{x86asm.ES, 0x2000},
				regval{x86asm.DI, 0x0100},
				regval{x86asm.AL, 0x0021},
				memval8{0x2000, 0x100, 0x21},
			},
			[]w{}, "Z"},
		{"RET/NEAR", "C3", []h{pushval{0x2021}}, []w{regval{x86asm.IP, 0x2021}}, ""},
		{"RET/NEAR/POPBYTES", "C202", []h{pushval{0xBEEF}, pushval{0x2021}},
			[]w{regval{x86asm.IP, 0x2021}, regval{x86asm.SP, 0x00FF}}, ""},
		{"RET/FAR", "CB", []h{pushval{0xCAFE}, pushval{0x2021}},
			[]w{regval{x86asm.IP, 0x2021}, regval{x86asm.CS, 0xCAFE}}, ""},

		{"PUSH/IMM", "68EFBE", []h{}, []w{popval{0xBEEF}}, ""},
		{"POP/ES", "07", []h{pushval{0x2021}}, []w{regval{x86asm.ES, 0x2021}}, ""},
		{"POP/RMM/AX", "8FC0", []h{pushval{0x2021}}, []w{regval{x86asm.AX, 0x2021}}, ""},
		{"LES", "C43E7500", []h{memval16{DS, 0x0075, 0x2021}, memval16{DS, 0x0077, DS}},
			[]w{regval{x86asm.DI, 0x2021}, regval{x86asm.ES, DS}}, ""},
		{"LDS", "C53E7500", []h{memval16{DS, 0x0075, 0x2021}, memval16{DS, 0x0077, 0x2000}},
			[]w{regval{x86asm.DI, 0x2021}, regval{x86asm.DS, 0x2000}}, ""},
		{"CLD", "FC", []h{flagval{DF, true}}, []w{}, "d"},
		{"STD", "FD", []h{flagval{DF, false}}, []w{}, "D"},

		// OR
		{"OR/ALimm8", "0C02", []h{regval{x86asm.AL, 0x1}}, []w{regval{x86asm.AL, 0x3}}, ""},
		{"OR/AXimm16", "0D0200", []h{regval{x86asm.AX, 0x1}}, []w{regval{x86asm.AX, 0x3}}, ""},
		{"OR/80/1/imm8", "80C802", []h{regval{x86asm.AL, 0x1}}, []w{regval{x86asm.AL, 0x3}}, ""},
		{"OR/81/1/imm16", "81C80200", []h{regval{x86asm.AX, 0x1}}, []w{regval{x86asm.AX, 0x3}}, ""},
		{"OR/83/1/imm8", "83C802", []h{regval{x86asm.AX, 0x1}}, []w{regval{x86asm.AX, 0x3}}, ""},
		{"OR/08/rmm8/r8", "08D1", []h{regval{x86asm.DL, 0x1}, regval{x86asm.CL, 0x2}}, []w{regval{x86asm.CL, 0x3}}, ""},
		{"OR/09/rmm16/r16", "09D1", []h{regval{x86asm.DX, 0x1}, regval{x86asm.CX, 0x2}}, []w{regval{x86asm.CX, 0x3}}, ""},
		{"OR/0A/r8/rmm8", "0AD1", []h{regval{x86asm.DL, 0x1}, regval{x86asm.CL, 0x2}}, []w{regval{x86asm.DL, 0x3}}, ""},
		{"OR/0B/r16/rmm16", "0BD1", []h{regval{x86asm.DX, 0x1}, regval{x86asm.CX, 0x2}}, []w{regval{x86asm.DX, 0x3}}, ""},

		// AND
		{"AND/Z", "24F0", []h{regval{x86asm.AL, 0xF}}, []w{regval{x86asm.AL, 0x00}}, "Z"},
		{"AND/ALimm8", "2403", []h{regval{x86asm.AL, 0xF}}, []w{regval{x86asm.AL, 0x3}}, ""},
		{"AND/AXimm16", "250300", []h{regval{x86asm.AX, 0x2}}, []w{regval{x86asm.AX, 0x2}}, ""},
		{"AND/80/1/imm8", "80E003", []h{regval{x86asm.AL, 0xF}}, []w{regval{x86asm.AL, 0x3}}, ""},
		{"AND/81/1/imm16", "81E00300", []h{regval{x86asm.AX, 0xF}}, []w{regval{x86asm.AX, 0x3}}, ""},
		{"AND/83/1/imm8", "83E003", []h{regval{x86asm.AX, 0xF}}, []w{regval{x86asm.AX, 0x3}}, ""},
		{"AND/20/rmm8/r8", "20D1", []h{regval{x86asm.DL, 0xF}, regval{x86asm.CL, 0x3}}, []w{regval{x86asm.CL, 0x3}}, ""},
		{"AND/21/rmm16/r16", "21D1", []h{regval{x86asm.DX, 0xF}, regval{x86asm.CX, 0x3}}, []w{regval{x86asm.CX, 0x3}}, ""},
		{"AND/22/r8/rmm8", "22D1", []h{regval{x86asm.DL, 0xF}, regval{x86asm.CL, 0x3}}, []w{regval{x86asm.DL, 0x3}}, ""},
		{"AND/23/r16/rmm16", "23D1", []h{regval{x86asm.DX, 0xF}, regval{x86asm.CX, 0x3}}, []w{regval{x86asm.DX, 0x3}}, ""},

		// STOSB
		{"STOSB", "AA", []h{regval{x86asm.AL, 0x21}, regval{x86asm.ES, ES}, regval{x86asm.DI, 0x0000}},
			[]w{memval8{ES, 0x0, 0x21}, regval{x86asm.DI, 0x0001}}, ""},
		{"STOSW", "AB", []h{regval{x86asm.AX, 0x1234}, regval{x86asm.ES, ES}, regval{x86asm.DI, 0x0000}},
			[]w{memval16{ES, 0x0, 0x1234}, regval{x86asm.DI, 0x0002}}, ""},
		// STOSB
		{"LODSB", "AC",
			[]h{memval8{DS, 0x0, 0x41}, regval{x86asm.DS, DS}, regval{x86asm.SI, 0x0000}},
			[]w{regval{x86asm.AL, 0x41}, regval{x86asm.SI, 0x01}}, ""},
		{"LODSW", "AD",
			[]h{memval16{DS, 0x0, 0x1941}, regval{x86asm.DS, DS}, regval{x86asm.SI, 0x0000}},
			[]w{regval{x86asm.AX, 0x1941}, regval{x86asm.SI, 0x02}}, ""},
		{"STC", "F9",
			[]h{flagval{CF, false}},
			[]w{}, "C"},
		{"CLC", "F8",
			[]h{flagval{CF, true}},
			[]w{}, "c"},
		{"XCHG", "93", []h{regval{x86asm.AX, 0xA}, regval{x86asm.BX, 0xB}},
			[]w{regval{x86asm.AX, 0xB}, regval{x86asm.BX, 0xA}}, ""},
		// NOT is F7/2. NOT AX
		{"NOT", "F7D0",
			[]h{regval{x86asm.AX, 0x0001}},
			[]w{regval{x86asm.AX, 0xFFFE}}, ""},
		{"TEST/AX/IMM16/z", "A90180",
			[]h{regval{x86asm.AX, 0x0001}},
			[]w{}, "zps"},
		{"TEST/AX/IMM16/z", "A90180",
			[]h{regval{x86asm.AX, 0x0002}},
			[]w{}, "ZPs"},
		{"TEST/AX/IMM16/S", "A90180",
			[]h{regval{x86asm.AX, 0x8000}},
			[]w{}, "zS"},

		{"CWD/POS", "99",
			[]h{regval{x86asm.AX, 0x8000}},
			[]w{regval{x86asm.DX, 0xFFFF}}, ""},
		{"CWD/NEG", "99",
			[]h{regval{x86asm.AX, 0x7FFF}},
			[]w{regval{x86asm.DX, 0x0}}, ""},
		{"CMPSB/EQ", "A6",
			[]h{regval{x86asm.DS, DS}, regval{x86asm.SI, 0},
				regval{x86asm.ES, ES}, regval{x86asm.DI, 0},
				memval8{DS, 0, 42}, memval8{ES, 0, 42}},
			[]w{}, "Z"},

		// TODO
	}

	for r := x86asm.AL; r <= x86asm.BH; r++ {
		// Append B0-B7 for MOV r8, imm8 - Move imm8 to r8.
		tests = append(tests,
			testinfo{fmt.Sprintf("MOV0x%x/%v", uint8(r), r),
				fmt.Sprintf("%02XAA", 0xB0+uint8(r)-uint8(x86asm.AL)),
				[]h{},
				[]w{regval{r, 0xAA}}, ""})

	}
	for r := x86asm.AX; r <= x86asm.BX; r++ {
		// Append B8-BF for MOV r16, imm16- Move imm16 to r16.
		tests = append(tests,
			testinfo{fmt.Sprintf("MOV0x%x/%v", uint8(r), r),
				fmt.Sprintf("%02XEFBE", 0xB8+uint8(r)-uint8(x86asm.AX)),
				[]h{},
				[]w{regval{r, 0xBEEF}}, ""})
	}

	for _, test := range tests {
		descr := fmt.Sprintf("%s:{%s}", test.descr, test.opcodes)
		t.Run(descr, func(t *testing.T) {
			cpu := SetupCPU(t, test.opcodes)
			for _, r := range test.have {
				switch r := r.(type) {
				case flagval:
					if r.set {
						cpu.Flags |= r.mask
					} else {
						cpu.Flags &^= r.mask
					}
				case regval:
					cpu.PutReg(r.reg, r.val)
				case memval8:
					cpu.Mem.PutMem8(r.seg, r.off, r.val)
				case memval16:
					cpu.Mem.PutMem16(r.seg, r.off, r.val)
				case pushval:
					cpu.Push(r.val)
				default:
					t.Fatalf("Unknown have: %#v", r)
				}
			}
			assert.Equal(t, true, cpu.RunOnce(), "RunOnce Failed")

			for _, r := range test.want {
				switch r := r.(type) {
				case regval:
					assert.Equal(t, r.val, cpu.Reg(r.reg), r.reg)
				case memval8:
					actual := cpu.Mem.Mem8(r.seg, r.off)
					assert.Equal(t, actual, r.val, r)
				case memval16:
					actual := cpu.Mem.Mem16(r.seg, r.off)
					assert.Equal(t, actual, r.val, r)
				case popval:
					actual := cpu.Pop()
					assert.Equal(t, actual, r.val)
				default:
					t.Fatalf("Unknown want: %#v", r)
				}
			}
			checkFlags(t, cpu, test.want_flags)
		})
	}
}

func TestCPUJumps8(t *testing.T) {
	type testinfo struct {
		op      x86asm.Op
		b       byte
		yescond uint16
		nocond  uint16
	}

	tests := []testinfo{
		{x86asm.JA, 0x77, ^(CF | ZF), CF},
		{x86asm.JA, 0x77, ^(CF | ZF), ZF},
		{x86asm.JA, 0x77, ^(CF | ZF), CF | ZF},
		// JMP: EBF9: JMP -7
		{x86asm.JAE, 0x73, ^CF, CF},
		{x86asm.JB, 0x73, ^CF, CF},
		{x86asm.JBE, 0x76, CF, 0},
		{x86asm.JBE, 0x76, ZF, 0},
		{x86asm.JE, 0x74, ZF, 0},
		{x86asm.JG, 0x7F, SF | OF, ZF},
		{x86asm.JG, 0x7F, 0, SF},
		{x86asm.JGE, 0x7D, SF | OF, SF},
		{x86asm.JGE, 0x7D, ZF | SF | OF, SF},
		{x86asm.JL, 0x7C, SF, SF | OF},
		{x86asm.JL, 0x7C, OF, SF | OF},
		{x86asm.JLE, 0x7E, OF, SF | OF},
		{x86asm.JLE, 0x7E, SF, SF | OF},
		{x86asm.JLE, 0x7E, ZF, SF | OF},
		{x86asm.JNE, 0x75, 0, ZF},
		{x86asm.JNO, 0x71, 0, OF},
		{x86asm.JNP, 0x7B, 0, PF},
		{x86asm.JNS, 0x79, 0, SF},
		{x86asm.JO, 0x70, OF, 0},
		{x86asm.JP, 0x7A, PF, 0},
		{x86asm.JS, 0x78, SF, 0},
	}

	for _, test := range tests {
		descr := fmt.Sprintf("%v:{%X}", test.op, test.b)
		opcodes := fmt.Sprintf("%02X07", test.b)
		t.Run(fmt.Sprintf("%s/Y", descr), func(t *testing.T) {
			cpu := SetupCPU(t, opcodes)
			cpu.Flags = test.yescond
			assert.Equal(t, true, cpu.RunOnce())
			assert.Equal(t, cpu.Ip, uint16(9))
		})
		t.Run(fmt.Sprintf("%s/N", descr), func(t *testing.T) {
			cpu := SetupCPU(t, opcodes)
			cpu.Flags = test.nocond
			assert.Equal(t, true, cpu.RunOnce())
			assert.Equal(t, cpu.Ip, uint16(2))
		})
	}
}

func TestCpuRegs(t *testing.T) {
	var tests = []struct {
		w x86asm.Reg
		l x86asm.Reg
		h x86asm.Reg
	}{
		{x86asm.AX, x86asm.AL, x86asm.AH},
		{x86asm.BX, x86asm.BL, x86asm.BH},
		{x86asm.CX, x86asm.CL, x86asm.CH},
		{x86asm.DX, x86asm.DL, x86asm.DH},
	}
	for _, test := range tests {
		t.Run(fmt.Sprintf("%v/%v/%v", test.w, test.l, test.h), func(t *testing.T) {
			cpu := SetupCPU(t, "00C2")
			cpu.PutReg(test.w, 0xBEEF)
			assert.Equal(t, cpu.Reg(test.l), uint(0xEF))
			assert.Equal(t, cpu.Reg(test.h), uint(0xBE))

			cpu.PutReg(test.h, 0xCA)
			cpu.PutReg(test.l, 0xFE)
			assert.Equal(t, cpu.Reg(test.w), uint(0xCAFE))
		})
	}
}
