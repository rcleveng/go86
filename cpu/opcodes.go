package go86

type OpcodeInfo struct {
	Opcode uint8
	Name   string
}

var OpcodeTable = [256]OpcodeInfo{
	{0x00, "ADD Eb Gb"},
	{0x01, "ADD Ev Gv"},
	{0x02, "ADD Gb Eb"},
	{0x03, "ADD Gv Ev"},
	{0x04, "ADD AL Ib"},
	{0x05, "ADD eAX Iv"},
	{0x06, "PUSH ES"},
	{0x07, "POP ES"},
	{0x08, "OR Eb Gb"},
	{0x09, "OR Ev Gv"},
	{0x0A, "OR Gb Eb"},
	{0x0B, "OR Gv Ev"},
	{0x0C, "OR AL Ib"},
	{0x0D, "OR eAX Iv"},
	{0x0E, "PUSH CS"},
	{0x0F, "--"},
	{0x10, "ADC Eb Gb"},
	{0x11, "ADC Ev Gv"},
	{0x12, "ADC Gb Eb"},
	{0x13, "ADC Gv Ev"},
	{0x14, "ADC AL Ib"},
	{0x15, "ADC eAX Iv"},
	{0x16, "PUSH SS"},
	{0x17, "POP SS"},
	{0x18, "SBB Eb Gb"},
	{0x19, "SBB Ev Gv"},
	{0x1A, "SBB Gb Eb"},
	{0x1B, "SBB Gv Ev"},
	{0x1C, "SBB AL Ib"},
	{0x1D, "SBB eAX Iv"},
	{0x1E, "PUSH DS"},
	{0x1F, "POP DS"},
	{0x20, "AND Eb Gb"},
	{0x21, "AND Ev Gv"},
	{0x22, "AND Gb Eb"},
	{0x23, "AND Gv Ev"},
	{0x24, "AND AL Ib"},
	{0x25, "AND eAX Iv"},
	{0x26, "ES:"},
	{0x27, "DAA"},
	{0x28, "SUB Eb Gb"},
	{0x29, "SUB Ev Gv"},
	{0x2A, "SUB Gb Eb"},
	{0x2B, "SUB Gv Ev"},
	{0x2C, "SUB AL Ib"},
	{0x2D, "SUB eAX Iv"},
	{0x2E, "CS:"},
	{0x2F, "DAS"},
	{0x30, "XOR Eb Gb"},
	{0x31, "XOR Ev Gv"},
	{0x32, "XOR Gb Eb"},
	{0x33, "XOR Gv Ev"},
	{0x34, "XOR AL Ib"},
	{0x35, "XOR eAX Iv"},
	{0x36, "SS:"},
	{0x37, "AAA"},
	{0x38, "CMP Eb Gb"},
	{0x39, "CMP Ev Gv"},
	{0x3A, "CMP Gb Eb"},
	{0x3B, "CMP Gv Ev"},
	{0x3C, "CMP AL Ib"},
	{0x3D, "CMP eAX Iv"},
	{0x3E, "DS:"},
	{0x3F, "AAS"},
	{0x40, "INC eAX"},
	{0x41, "INC eCX"},
	{0x42, "INC eDX"},
	{0x43, "INC eBX"},
	{0x44, "INC eSP"},
	{0x45, "INC eBP"},
	{0x46, "INC eSI"},
	{0x47, "INC eDI"},
	{0x48, "DEC eAX"},
	{0x49, "DEC eCX"},
	{0x4A, "DEC eDX"},
	{0x4B, "DEC eBX"},
	{0x4C, "DEC eSP"},
	{0x4D, "DEC eBP"},
	{0x4E, "DEC eSI"},
	{0x4F, "DEC eDI"},
	{0x50, "PUSH eAX"},
	{0x51, "PUSH eCX"},
	{0x52, "PUSH eDX"},
	{0x53, "PUSH eBX"},
	{0x54, "PUSH eSP"},
	{0x55, "PUSH eBP"},
	{0x56, "PUSH eSI"},
	{0x57, "PUSH eDI"},
	{0x58, "POP eAX"},
	{0x59, "POP eCX"},
	{0x5A, "POP eDX"},
	{0x5B, "POP eBX"},
	{0x5C, "POP eSP"},
	{0x5D, "POP eBP"},
	{0x5E, "POP eSI"},
	{0x5F, "POP eDI"},
	{0x60, "--"},
	{0x61, "--"},
	{0x62, "--"},
	{0x63, "--"},
	{0x64, "--"},
	{0x65, "--"},
	{0x66, "--"},
	{0x67, "--"},
	{0x68, "--"},
	{0x69, "--"},
	{0x6A, "--"},
	{0x6B, "--"},
	{0x6C, "--"},
	{0x6D, "--"},
	{0x6E, "--"},
	{0x6F, "--"},
	{0x70, "JO Jb"},
	{0x71, "JNO Jb"},
	{0x72, "JB Jb"},
	{0x73, "JNB Jb"},
	{0x74, "JZ Jb"},
	{0x75, "JNZ Jb"},
	{0x76, "JBE Jb"},
	{0x77, "JA Jb"},
	{0x78, "JS Jb"},
	{0x79, "JNS Jb"},
	{0x7A, "JPE Jb"},
	{0x7B, "JPO Jb"},
	{0x7C, "JL Jb"},
	{0x7D, "JGE Jb"},
	{0x7E, "JLE Jb"},
	{0x7F, "JG Jb"},
	{0x80, "GRP1 Eb Ib"},
	{0x81, "GRP1 Ev Iv"},
	{0x82, "GRP1 Eb Ib"},
	{0x83, "GRP1 Ev Ib"},
	{0x84, "TEST Gb Eb"},
	{0x85, "TEST Gv Ev"},
	{0x86, "XCHG Gb Eb"},
	{0x87, "XCHG Gv Ev"},
	{0x88, "MOV Eb Gb"},
	{0x89, "MOV Ev Gv"},
	{0x8A, "MOV Gb Eb"},
	{0x8B, "MOV Gv Ev"},
	{0x8C, "MOV Ew Sw"},
	{0x8D, "LEA Gv M"},
	{0x8E, "MOV Sw Ew"},
	{0x8F, "POP Ev"},
	{0x90, "NOP"},
	{0x91, "XCHG eCX eAX"},
	{0x92, "XCHG eDX eAX"},
	{0x93, "XCHG eBX eAX"},
	{0x94, "XCHG eSP eAX"},
	{0x95, "XCHG eBP eAX"},
	{0x96, "XCHG eSI eAX"},
	{0x97, "XCHG eDI eAX"},
	{0x98, "CBW"},
	{0x99, "CWD"},
	{0x9A, "CALL Ap"},
	{0x9B, "WAIT"},
	{0x9C, "PUSHF"},
	{0x9D, "POPF"},
	{0x9E, "SAHF"},
	{0x9F, "LAHF"},
	{0xA0, "MOV AL Ob"},
	{0xA1, "MOV eAX Ov"},
	{0xA2, "MOV Ob AL"},
	{0xA3, "MOV Ov eAX"},
	{0xA4, "MOVSB"},
	{0xA5, "MOVSW"},
	{0xA6, "CMPSB"},
	{0xA7, "CMPSW"},
	{0xA8, "TEST AL Ib"},
	{0xA9, "TEST eAX Iv"},
	{0xAA, "STOSB"},
	{0xAB, "STOSW"},
	{0xAC, "LODSB"},
	{0xAD, "LODSW"},
	{0xAE, "SCASB"},
	{0xAF, "SCASW"},
	{0xB0, "MOV AL Ib"},
	{0xB1, "MOV CL Ib"},
	{0xB2, "MOV DL Ib"},
	{0xB3, "MOV BL Ib"},
	{0xB4, "MOV AH Ib"},
	{0xB5, "MOV CH Ib"},
	{0xB6, "MOV DH Ib"},
	{0xB7, "MOV BH Ib"},
	{0xB8, "MOV eAX Iv"},
	{0xB9, "MOV eCX Iv"},
	{0xBA, "MOV eDX Iv"},
	{0xBB, "MOV eBX Iv"},
	{0xBC, "MOV eSP Iv"},
	{0xBD, "MOV eBP Iv"},
	{0xBE, "MOV eSI Iv"},
	{0xBF, "MOV eDI Iv"},
	{0xC0, "--"},
	{0xC1, "--"},
	{0xC2, "RET Iw"},
	{0xC3, "RET"},
	{0xC4, "LES Gv Mp"},
	{0xC5, "LDS Gv Mp"},
	{0xC6, "MOV Eb Ib"},
	{0xC7, "MOV Ev Iv"},
	{0xC8, "--"},
	{0xC9, "--"},
	{0xCA, "RETF Iw"},
	{0xCB, "RETF"},
	{0xCC, "INT 3"},
	{0xCD, "INT Ib"},
	{0xCE, "INTO"},
	{0xCF, "IRET"},
	{0xD0, "GRP2 Eb 1"},
	{0xD1, "GRP2 Ev 1"},
	{0xD2, "GRP2 Eb CL"},
	{0xD3, "GRP2 Ev CL"},
	{0xD4, "AAM I0"},
	{0xD5, "AAD I0"},
	{0xD6, "--"},
	{0xD7, "XLAT"},
	{0xD8, "--"},
	{0xD9, "--"},
	{0xDA, "--"},
	{0xDB, "--"},
	{0xDC, "--"},
	{0xDD, "--"},
	{0xDE, "--"},
	{0xDF, "--"},
	{0xE0, "LOOPNZ Jb"},
	{0xE1, "LOOPZ Jb"},
	{0xE2, "LOOP Jb"},
	{0xE3, "JCXZ Jb"},
	{0xE4, "IN AL Ib"},
	{0xE5, "IN eAX Ib"},
	{0xE6, "OUT Ib AL"},
	{0xE7, "OUT Ib eAX"},
	{0xE8, "CALL Jv"},
	{0xE9, "JMP Jv"},
	{0xEA, "JMP Ap"},
	{0xEB, "JMP Jb"},
	{0xEC, "IN AL DX"},
	{0xED, "IN eAX DX"},
	{0xEE, "OUT DX AL"},
	{0xEF, "OUT DX eAX"},
	{0xF0, "LOCK"},
	{0xF1, "--"},
	{0xF2, "REPNZ"},
	{0xF3, "REPZ"},
	{0xF4, "HLT"},
	{0xF5, "CMC"},
	{0xF6, "GRP3a Eb"},
	{0xF7, "GRP3b Ev"},
	{0xF8, "CLC"},
	{0xF9, "STC"},
	{0xFA, "CLI"},
	{0xFB, "STI"},
	{0xFC, "CLD"},
	{0xFD, "STD"},
	{0xFE, "GRP4 Eb"},
	{0xFF, "GRP5 Ev"},
}
