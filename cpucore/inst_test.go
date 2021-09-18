package go86

import (
	"encoding/hex"
	"fmt"
	"log"
	"testing"

	"golang.org/x/arch/x86/x86asm"
)

func DumpOpCodeString(t *testing.T, s string) {
	b, err := hex.DecodeString(s)
	if err != nil {
		log.Fatal(err)
	}

	fmt.Println("DecodeOpString: ", s)
	inst, err := x86asm.Decode(b, 16)
	if err != nil {
		log.Fatal(err)
	}
	fmt.Printf("len: %d; text: '%s'; opcode: %x [Ds:%d] [As:%d] [Ms:%d]\n",
		inst.Len, inst.String(), inst.Opcode, inst.DataSize, inst.AddrSize,
		inst.MemBytes)
	for num, a := range inst.Args {
		if a == nil {
			break
		}
		fmt.Printf("Arg #%d: (%T) [Val: %v]", num, a, a)
		switch v := a.(type) {
		case x86asm.Imm:
			av := int64(v)
			fmt.Printf("[imm16: %X]", av&0xFFFF)
		}
		fmt.Println()
	}
}

// C43E7500 - LES DI,[0075]
func TestInstSmoke(t *testing.T) {
	//dumpOpCodeString(t, "C43E7500")
}

// FF 35 0E 20 40 00 (PUSHF?)
func TestInstPushF(t *testing.T) {
	// dumpOpCodeString(t, "FF2EFECA")
	// dumpOpCodeString(t, "EAFECA0000")
	// dumpOpCodeString(t, "EB2EFECA")
	// dumpOpCodeString(t, "0107")
	// dumpOpCodeString(t, "0207")
}
