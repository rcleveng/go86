package go86

import (
	"bytes"
	"errors"
	"io"
	"os"

	log "github.com/golang/glog"
	cpu "go86.org/go86/cpu"
)

type Dos struct {
	// TODO: Make a custom interface that also has this
	Out io.Writer
	In  io.Writer
	Err io.Writer

	Mem *DosMem
	cpu *cpu.CPU
}

func (dos *Dos) Int20(c *cpu.CPU, intnum int) {
	log.V(3).Infof("Dos.int%02x: [AX: %04X]", intnum, c.Regs.GetReg16(cpu.AX))
	c.Halt()
}

func (dos *Dos) Int21(c *cpu.CPU, intnum int) {
	log.V(3).Infof("Dos.int%02x: [AX: %04X]", intnum, c.Regs.GetReg16(cpu.AX))
	switch ah := c.Regs.GetReg8(cpu.AH); ah {
	case 0x02: // Print Char
		dl := c.Regs.GetReg8(cpu.DL)
		s := []byte{byte(dl)}
		dos.Out.Write(s)
	case 0x09: // Print String
		ds := c.Regs.DS()
		dx := c.Regs.GetReg16(cpu.DX)
		b := c.Mem.At(ds, dx)
		end := bytes.IndexByte(b, byte('$'))
		if end != -1 {
			s := b[:end]
			dos.Out.Write(s)
		}
	case 0x30:
		// AH=30h - GET DOS VERSION
		c.Regs.SetReg16(cpu.AX, 0x0203) // 3.2
	case 0x25:
		// AH = 25h - SET INTERRUPT VECTOR
		al := c.Regs.GetReg8(cpu.AL)
		ds := c.Regs.DS()
		dx := c.Regs.GetReg16(cpu.DX)
		c.Mem.SetMem16(0, al, uint16(dx))
		c.Mem.SetMem16(0, al+2, uint16(ds))
	case 0x35:
		// AH=35h - GET INTERRUPT VECTOR
		al := c.Regs.GetReg16(cpu.AX) & 0xFF
		c.Regs.SetSeg16(cpu.ES, uint(c.Mem.GetMem16(0, al)))
		c.Regs.SetReg16(cpu.BX, uint(c.Mem.GetMem16(0, al+2)))
	case 0x40:
		// AH=40h - "WRITE" - WRITE TO FILE OR DEVICE
		bx := c.Regs.GetReg16(cpu.BX)
		cx := c.Regs.GetReg16(cpu.CX)
		dx := c.Regs.GetReg16(cpu.DX)
		s := c.Mem.At(c.Regs.DS(), dx)[:cx]
		switch bx {
		case 1:
			dos.Out.Write(s)
		case 2:
			dos.Err.Write(s)
		default:
			// TODO: look up filehandle from JFT, map to SFT and write.
		}
	case 0x4a:
		// INT 21 - AH = 4Ah DOS 2+ - ADJUST MEMORY BLOCK SIZE (SETBLOCK)
		// ES = Segment address of block to change
		// BX = New size in paragraphs
		es := c.Regs.ES()
		bx := c.Regs.GetReg16(cpu.BX)
		log.V(3).Infof("INT21H: [4A] [BLOCK: 0x%04X, SIZE: %d paragraphs, %d bytes]", es, bx, bx*16)
		newsize, err := dos.Mem.Resize(es, bx)
		if err != nil {
			dos.cpu.Flags.SetFlags(cpu.CF)
			return
		}
		c.Regs.SetReg16(cpu.BX, uint(newsize))
	case 0x4C:
		c.Halt()
	default:
		log.Warningf("Unhandled DOS Interrupt Code: [%02x]\n", ah)
	}
}

func NewDos(cpu *cpu.CPU) *Dos {
	end := uint(0x9FC0)
	dos := &Dos{
		Out: os.Stdout,
		In:  os.Stdin,
		Err: os.Stderr,
		// TODO: If end > c.Mem end, lower it.
		// 0x800
		Mem: NewDosMem(0x0C85, end),
		cpu: cpu,
	}

	cpu.Intrs[0x20] = (*dos).Int20
	cpu.Intrs[0x21] = (*dos).Int21

	return dos
}

func (dos *Dos) createPsp(exe *Executable, seg_base *DosMemBlock, env_seg *DosMemBlock) {
	log.V(1).Infof("Creating PSP at: 0x%04X\n", seg_base.Start)
	log.V(4).Infof("Executable Type: %#v\n", exe.Etype)
	start := seg_base.Start
	dos.cpu.Mem.SetMem8(start, 0, 0xCD)
	dos.cpu.Mem.SetMem8(start, 1, 0x20)
	// First paragraph following this segment.
	dos.cpu.Mem.SetMem16(start, 2, uint16(seg_base.End+1))

	dos.cpu.Mem.SetMem8(start, 10, 0x22) // int22 handler
	dos.cpu.Mem.SetMem8(start, 14, 0x23) // int23 handler
	dos.cpu.Mem.SetMem8(start, 18, 0x24) // int24 handler
	// FFFE means no parent DOS process
	dos.cpu.Mem.SetMem16(start, 22, 0xFFFE)
	dos.cpu.Mem.SetMem16(start, 44, uint16(env_seg.Start))
	// TODO: Create rest
}

func (dos *Dos) LoadCom(exe *Executable, seg_base *DosMemBlock) (seg uint16, err error) {
	// DS is what we allocated, for EXE, CS is 0x100 past it since the PSP goes first
	seg_start := uint16(seg_base.Start)
	image_start := seg_start + 0x10
	dos.cpu.Regs.SetSeg16(cpu.CS, uint(seg_start))
	dos.cpu.Regs.SetSeg16(cpu.DS, uint(seg_start))
	dos.cpu.Regs.SetSeg16(cpu.ES, uint(seg_start))
	dos.cpu.Regs.SetSeg16(cpu.SS, uint(seg_start))
	// default SP
	dos.cpu.Regs.SetReg16(cpu.SP, uint(0xFFFE))
	// skip past PSP
	dos.cpu.Ip = 0x100

	// Copy data into memory at CS from the binary read from disk.
	copy(dos.cpu.Mem.At(uint(image_start), 0), exe.Data)

	return seg_start, nil
}

func (dos *Dos) LoadImage(exe *Executable, seg_base *DosMemBlock) (seg uint16, err error) {
	// DS is what we allocated, for EXE, CS is 0x100 past it since the PSP goes first
	seg_start := uint16(seg_base.Start)
	image_start := seg_start
	dos.cpu.Regs.SetSeg16(cpu.CS, uint(seg_start))
	dos.cpu.Regs.SetSeg16(cpu.DS, uint(seg_start))
	dos.cpu.Regs.SetSeg16(cpu.ES, uint(seg_start))
	dos.cpu.Regs.SetSeg16(cpu.SS, uint(seg_start))
	// default SP
	dos.cpu.Regs.SetReg16(cpu.SP, uint(0xFFFE))
	// skip past PSP
	dos.cpu.Ip = 0

	// Copy data into memory at CS from the binary read from disk.
	copy(dos.cpu.Mem.At(uint(image_start), 0), exe.Data)

	return seg_start, nil
}

func (dos *Dos) LoadExe(exe *Executable, seg_base *DosMemBlock) (seg uint16, err error) {
	// DS is what we allocated, for EXE, CS is 0x100 past it since the PSP goes first
	seg_start := uint(seg_base.Start)
	img_start := uint16(seg_start + 0x0010)
	cs := (img_start + exe.Hdr.CS) & 0xFFFF
	ss := (img_start + exe.Hdr.SS) & 0xFFFF
	dos.cpu.Regs.SetSeg16(cpu.CS, uint(cs))
	dos.cpu.Regs.SetSeg16(cpu.DS, seg_start)
	dos.cpu.Regs.SetSeg16(cpu.ES, seg_start)
	dos.cpu.Regs.SetSeg16(cpu.SS, uint(ss))
	dos.cpu.Regs.SetReg16(cpu.SP, uint(exe.Hdr.SP))
	dos.cpu.Regs.SetReg16(cpu.BP, 0)
	dos.cpu.Ip = exe.Hdr.IP

	// Copy data into memory at CS from the binary read from disk.
	copy(dos.cpu.Mem.At(uint(cs), 0), exe.Data)

	log.V(1).Infof("EXE Values:\nCS: 0x%04X\nDS: 0x%04X\nES: 0x%04X\nSS: 0x%04X\nIP: 0x%04X\n\n",
		cs, seg_start, seg_start, ss, dos.cpu.Ip)
	log.V(1).Infof("SP: 0x%04X\n", dos.cpu.Regs.GetReg16(cpu.SP))

	// Fixup relos
	for _, r := range exe.Hdr.Relos {
		m := dos.cpu.Mem.GetMem16(uint(img_start+r.Segment), uint(r.Offset))
		dos.cpu.Mem.SetMem16(uint(img_start+r.Segment), uint(r.Offset), m+img_start)
		log.V(3).Infof("Relo: [0x%04X:0x%04X] += 0x%04X", r.Segment, r.Offset, img_start)
	}

	return uint16(seg_start), nil
}

func (dos *Dos) Load(exe *Executable) (seg uint16, err error) {
	if !exe.Exists || len(exe.Data) == 0 {
		return 0, errors.New("executable not read")
	}
	// hack really look up environment
	env_seg, err := dos.Mem.Allocate(10)
	if err != nil {
		return 0, err
	}
	env_mem := dos.cpu.Mem.At(env_seg.Start, 0)
	copy(env_mem, []byte("PATH=Z:\\\n"))

	sn := exe.SegmentsNeeded()
	seg_base, err := dos.Mem.Allocate(sn)
	log.V(2).Infof("DOS Allocated [%d segments %d bytes]", sn, sn*0x10)
	if err != nil {
		return 0, err
	}
	// We own our own memory block.

	seg_base.Owner = seg_base.Start
	// TODO: Create PSP

	switch exe.Etype {
	case EXE:
		dos.createPsp(exe, seg_base, env_seg)
		return dos.LoadExe(exe, seg_base)
	case COM:
		dos.createPsp(exe, seg_base, env_seg)
		return dos.LoadCom(exe, seg_base)
	case IMAGE:
		return dos.LoadImage(exe, seg_base)
	default:
		panic("Unhandled Etype")
	}
}
