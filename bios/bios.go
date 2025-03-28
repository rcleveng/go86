package go86

import (
	"io"
	"os"

	log "github.com/golang/glog"
	cpu "go86.org/go86/cpu"
)

type Bios struct {
	// TODO: Make a custom interface that also has this
	Out io.Writer
	In  io.Writer

	cpu *cpu.CPU
}

func (bios *Bios) Int10(c *cpu.CPU, intnum int) {
	log.Infof("Bios.int%02x: [AX: %04X]", intnum, c.Regs.GetReg16(cpu.AX))
	switch ah := c.Regs.GetReg8(cpu.AH); ah {
	default:
		log.Warningf("Unhandled BIOS Interrupt Code: [%02x]\n", ah)
	// HACK
	case 0x09, 0x0A: // Print Char
		// 0x09 shoud set attribute too, we're not doing that yet
		al := c.Regs.GetReg8(cpu.AL)
		s := []byte{byte(al)}
		bios.Out.Write(s)
		if cx := int(c.Regs.GetReg16(cpu.CX)); cx > 1 {
			for i := 1; i < cx; i++ {
				bios.Out.Write(s)
			}
		}
	case 0x0E: // Print Char
		// 0x09 shoud set attribute too, we're not doing that yet
		al := c.Regs.GetReg8(cpu.AL)
		s := []byte{byte(al)}
		bios.Out.Write(s)
	case 0x13: // Print String
		es := c.Regs.GetSeg16(cpu.ES)
		bp := c.Regs.GetReg16(cpu.BP)
		b := c.Mem.At(es, bp)
		end := c.Regs.GetReg16(cpu.CX)
		s := b[:end]
		bios.Out.Write(s)
	}
}

func (dos *Bios) Int13(c *cpu.CPU, intnum int) {
	log.Infof("Bios.int%02x: [AX: %04X]", intnum, c.Regs.GetReg16(cpu.AX))
	ah := c.Regs.GetReg8(cpu.AH)
	switch ah {
	case 0x02:
		// HACK
		// Clear CF == success, AH = status, AL == num read
		c.Regs.SetReg8(cpu.AH, 0)
		c.Flags.ClearFlag(cpu.CarryFlag)
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
