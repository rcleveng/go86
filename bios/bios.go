package go86

import (
	"io"
	"os"

	log "github.com/golang/glog"
	cpu "go86.org/go86/cpu"
	"golang.org/x/arch/x86/x86asm"
)

type Bios struct {
	// TODO: Make a custom interface that also has this
	Out io.Writer
	In  io.Writer

	cpu *cpu.CPU
}

func (dos *Bios) Int10(c *cpu.CPU, intnum int) {
	log.Infof("Bios.int%02x: [AX: %04X]", intnum, c.Regs[cpu.REG_AX])
	switch ah := c.Reg(x86asm.AH); ah {
	default:
		log.Warningf("Unhandled BIOS Interrupt Code: [%02x]\n", ah)
	// HACK
	case 0x09, 0x0A: // Print Char
		// 0x09 shoud set attribute too, we're not doing that yet
		al := c.Reg(x86asm.AL)
		s := []byte{byte(al)}
		dos.Out.Write(s)
		if cx := int(c.Reg(x86asm.CX)); cx > 1 {
			for i := 1; i < cx; i++ {
				dos.Out.Write(s)
			}
		}
	case 0x0E: // Print Char
		// 0x09 shoud set attribute too, we're not doing that yet
		al := c.Reg(x86asm.AL)
		s := []byte{byte(al)}
		dos.Out.Write(s)
	case 0x13: // Print String
		es := c.Sregs[cpu.SREG_ES]
		bp := c.Reg(x86asm.BP)
		b := c.Mem.At(int(es), int(bp))
		end := c.Reg(x86asm.CX)
		s := b[:end]
		dos.Out.Write(s)
	}
}

func (dos *Bios) Int13(c *cpu.CPU, intnum int) {
	log.Infof("Bios.int%02x: [AX: %04X]", intnum, c.Regs[cpu.REG_AX])
	ah := c.Reg(x86asm.AH)
	switch ah {
	case 0x02:
		// HACK
		// Clear CF == success, AH = status, AL == num read
		c.PutReg(x86asm.AH, 0)
		c.ClearFlag(cpu.CF)
	default:
	}
}

func NewBios(cpu *cpu.CPU) *Bios {
	bios := &Bios{
		Out: os.Stdout,
		In:  os.Stdin,
		cpu: cpu,
	}

	cpu.Intrs[0x10] = (*bios).Int10
	cpu.Intrs[0x13] = (*bios).Int13

	return bios
}
