package main

import (
	"bytes"
	"encoding/hex"
	"fmt"
	"log"

	"golang.org/x/arch/x86/x86asm"
)

func String16(i x86asm.Inst) string {
	var buf bytes.Buffer
	for _, p := range i.Prefix {
		if p == 0 {
			break
		}
		if p&x86asm.PrefixImplicit != 0 {
			continue
		}
		fmt.Fprintf(&buf, "%v ", p)
	}
	fmt.Fprintf(&buf, "%v", i.Op)
	sep := " "
	for _, v := range i.Args {
		if v == nil {
			break
		}
		fmt.Fprintf(&buf, "%s%v", sep, v)
		sep = ", "
	}
	return buf.String()
}

func DumpOpCodeString(s string) {
	b, err := hex.DecodeString(s)
	if err != nil {
		log.Fatal(err)
	}

	inst, _ := x86asm.Decode(b, 16)
	fmt.Printf("len: %d; text: '%s'; opcode: %x\n", inst.Len, inst.String(), inst.Opcode)
	for num, a := range inst.Args {
		if a == nil {
			break
		}
		fmt.Printf("Arg #%d: %v (%T) [Ds:%d] [As:%d] ", num, a, a, inst.DataSize, inst.AddrSize)
		switch v := a.(type) {
		case x86asm.Imm:
			av := int64(v)
			fmt.Printf("[imm16: %X]", av&0xFFFF)
		}
		fmt.Println()
	}

}

func main() {
	// "B8FECA0000"  move ax, 0xcafe
	DumpOpCodeString("B8FECA0000")
}
