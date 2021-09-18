package go86

import (
	"bytes"
	"errors"
	"io"
	"os"

	log "github.com/golang/glog"
	cpucore "go86.org/go86/cpucore"
	"golang.org/x/arch/x86/x86asm"
)

type Dos struct {
	// TODO: Make a custom interface that also has this
	Out io.Writer
	In  io.Writer
	Err io.Writer

	Mem *DosMem
	cpu *cpucore.CPU
}

func (dos *Dos) Int20(cpu *cpucore.CPU, intnum int) {
	log.V(3).Infof("Dos.int%02x: [AX: %04X]", intnum, cpu.Regs[cpucore.REG_AX])
	cpu.Halt()
}

func (dos *Dos) Int21(cpu *cpucore.CPU, intnum int) {
	log.V(3).Infof("Dos.int%02x: [AX: %04X]", intnum, cpu.Regs[cpucore.REG_AX])
	switch ah := cpu.Reg(x86asm.AH); ah {
	case 0x02: // Print Char
		dl := cpu.Reg(x86asm.DL)
		s := []byte{byte(dl)}
		dos.Out.Write(s)
	case 0x09: // Print String
		ds := cpu.Sregs[cpucore.SREG_DS]
		dx := cpu.Reg(x86asm.DX)
		b := cpu.Mem.At(int(ds), int(dx))
		end := bytes.IndexByte(b, byte('$'))
		if end != -1 {
			s := b[:end]
			dos.Out.Write(s)
		}
	case 0x30:
		// AH=30h - GET DOS VERSION
		// 3.2
		cpu.Regs[cpucore.REG_AH] = 2
		cpu.Regs[cpucore.REG_AL] = 3
	case 0x25:
		// AH = 25h - SET INTERRUPT VECTOR
		al := int(cpu.Regs[cpucore.REG_AX] & 0xFF)
		ds := cpu.Sregs[cpucore.SREG_DS]
		dx := cpu.Regs[cpucore.REG_DX]
		cpu.Mem.PutMem16(0, al, dx)
		cpu.Mem.PutMem16(0, al+2, ds)
	case 0x35:
		// AH=35h - GET INTERRUPT VECTOR
		al := int(cpu.Regs[cpucore.REG_AX] & 0xFF)
		cpu.Sregs[cpucore.SREG_ES] = cpu.Mem.Mem16(0, al)
		cpu.Regs[cpucore.REG_BX] = cpu.Mem.Mem16(0, al+2)
	case 0x40:
		// AH=40h - "WRITE" - WRITE TO FILE OR DEVICE
		bx := int(cpu.Regs[cpucore.REG_BX])
		cx := int(cpu.Regs[cpucore.REG_CX])
		ds := int(cpu.Sregs[cpucore.SREG_DS])
		dx := int(cpu.Regs[cpucore.REG_DX])
		s := cpu.Mem.At(ds, dx)[:cx]
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
		es := int(cpu.Sregs[cpucore.SREG_ES])
		bx := int(cpu.Regs[cpucore.REG_BX])
		log.V(3).Infof("INT21H: [4A] [BLOCK: 0x%04X, SIZE: %d paragraphs, %d bytes]", es, bx, bx*16)
		newsize, err := dos.Mem.Resize(es, bx)
		if err != nil {
			dos.cpu.SetFlagIf(cpucore.CF, true)
			return
		}
		cpu.PutReg(x86asm.BX, uint(newsize))
	case 0x4C:
		cpu.Halt()
	default:
		log.Warningf("Unhandled DOS Interrupt Code: [%02x]\n", ah)
	}
}

func NewDos(cpu *cpucore.CPU) *Dos {
	end := 0x9FC0
	dos := &Dos{
		Out: os.Stdout,
		In:  os.Stdin,
		Err: os.Stderr,
		// TODO: If end > cpu.Mem end, lower it.
		// 0x800
		Mem: NewDosMem(0x0C85, end),
		cpu: cpu,
	}

	cpu.Intrs[0x20] = (*dos).Int20
	cpu.Intrs[0x21] = (*dos).Int21

	return dos
}

func (dos *Dos) createPsp(exe *Executable, seg_base *DosMemBlock, env_seg *DosMemBlock) {
	start := seg_base.Start
	dos.cpu.Mem.PutMem8(start, 0, 0xCD)
	dos.cpu.Mem.PutMem8(start, 1, 0x20)
	// First paragraph following this segment.
	dos.cpu.Mem.PutMem16(start, 2, uint16(seg_base.End+1))

	dos.cpu.Mem.PutMem8(start, 10, 0x22) // int22 handler
	dos.cpu.Mem.PutMem8(start, 14, 0x23) // int23 handler
	dos.cpu.Mem.PutMem8(start, 18, 0x24) // int24 handler
	// FFFE means no parent DOS process
	dos.cpu.Mem.PutMem16(start, 22, 0xFFFE)
	dos.cpu.Mem.PutMem16(start, 44, uint16(env_seg.Start))
	// TODO: Create rest
}

func (dos *Dos) LoadCom(exe *Executable, seg_base *DosMemBlock) (seg uint16, err error) {
	// DS is what we allocated, for EXE, CS is 0x100 past it since the PSP goes first
	seg_start := uint16(seg_base.Start)
	image_start := seg_start + 0x10
	dos.cpu.Sregs[cpucore.SREG_CS] = seg_start
	dos.cpu.Sregs[cpucore.SREG_DS] = seg_start
	dos.cpu.Sregs[cpucore.SREG_ES] = seg_start
	dos.cpu.Sregs[cpucore.SREG_SS] = seg_start
	// default SP
	dos.cpu.Regs[cpucore.REG_SP] = 0xFFFE
	// skip past PSP
	dos.cpu.Ip = 0x100

	// Copy data into memory at CS from the binary read from disk.
	copy(dos.cpu.Mem.At(int(image_start), 0), exe.Data)

	return seg_start, nil
}

func (dos *Dos) LoadImage(exe *Executable, seg_base *DosMemBlock) (seg uint16, err error) {
	// DS is what we allocated, for EXE, CS is 0x100 past it since the PSP goes first
	seg_start := uint16(seg_base.Start)
	image_start := seg_start
	dos.cpu.Sregs[cpucore.SREG_CS] = seg_start
	dos.cpu.Sregs[cpucore.SREG_DS] = seg_start
	dos.cpu.Sregs[cpucore.SREG_ES] = seg_start
	dos.cpu.Sregs[cpucore.SREG_SS] = seg_start
	// default SP
	dos.cpu.Regs[cpucore.REG_SP] = 0xFFFE
	// skip past PSP
	dos.cpu.Ip = 0

	// Copy data into memory at CS from the binary read from disk.
	copy(dos.cpu.Mem.At(int(image_start), 0), exe.Data)

	return seg_start, nil
}

func (dos *Dos) LoadExe(exe *Executable, seg_base *DosMemBlock) (seg uint16, err error) {
	// DS is what we allocated, for EXE, CS is 0x100 past it since the PSP goes first
	seg_start := uint16(seg_base.Start)
	img_start := seg_start + 0x0010
	ds := seg_start
	es := seg_start
	cs := (img_start + exe.Hdr.CS) & 0xFFFF
	ss := (img_start + exe.Hdr.SS) & 0xFFFF
	dos.cpu.Sregs[cpucore.SREG_CS] = cs
	dos.cpu.Sregs[cpucore.SREG_DS] = ds
	dos.cpu.Sregs[cpucore.SREG_ES] = es
	dos.cpu.Sregs[cpucore.SREG_SS] = ss
	dos.cpu.Regs[cpucore.REG_SP] = exe.Hdr.SP
	dos.cpu.Regs[cpucore.REG_BP] = 0
	dos.cpu.Ip = exe.Hdr.IP

	// Copy data into memory at CS from the binary read from disk.
	copy(dos.cpu.Mem.At(int(cs), 0), exe.Data)

	log.V(1).Infof("EXE Values:\nCS: 0x%04X\nDS: 0x%04X\nES: 0x%04X\nSS: 0x%04X\nIP: 0x%04X\n\n", cs, ds, es, ss, dos.cpu.Ip)
	log.V(1).Infof("SP: 0x%04X\n", dos.cpu.Regs[cpucore.REG_SP])

	// Fixup relos
	for _, r := range exe.Hdr.Relos {
		m := dos.cpu.Mem.Mem16(int(img_start+r.Segment), int(r.Offset))
		dos.cpu.Mem.PutMem16(int(img_start+r.Segment), int(r.Offset), m+img_start)
		log.V(3).Infof("Relo: [0x%04X:0x%04X] += 0x%04X", r.Segment, r.Offset, img_start)
	}

	return seg_start, nil
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
