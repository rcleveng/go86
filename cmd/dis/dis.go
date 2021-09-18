package main

import (
	"encoding/hex"
	"flag"
	"fmt"

	dos "go86.org/go86/dos"
	"golang.org/x/arch/x86/x86asm"
)

var help = flag.Bool("help", false, "--help means show help")

func main() {
	flag.Parse()
	if flag.NArg() < 1 {
		fmt.Print("EXPECTED program name\n")
		flag.Usage()
		return
	}
	if *help {
		flag.Usage()
		return
	}
	fmt.Print("Hello: " + flag.Arg(0) + "\n")
	exe, err := dos.ReadExeFromFile(flag.Arg(0))
	if err != nil {
		fmt.Printf("Failed to read file header from: '%s'; error: %s\n", flag.Arg(0), err)
		return
	}
	fmt.Println("Exe Header Information:")
	fmt.Printf("CS:   %X\n", exe.Hdr.CS)
	fmt.Printf("SS:   %X\n", exe.Hdr.SS)
	fmt.Printf("IP:   %X\n", exe.Hdr.IP)
	for i := 0; i < int(exe.Hdr.NumRelos); i++ {
		r := exe.Hdr.Relos[i]
		fmt.Printf("Relo: %04X:%04X\n", r.Segment, r.Offset)

	}

	pos := 0
	raw := exe.Data
	for len(raw) > 0 {
		inst, err := x86asm.Decode(raw, 16)
		if err != nil {
			return
		}
		cur := raw[:inst.Len]
		fmt.Printf("%08X: [%-10s] %v\n", pos, hex.EncodeToString(cur), inst)
		raw = raw[inst.Len:]
		pos += inst.Len

	}
}
