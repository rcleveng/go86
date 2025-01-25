package go86

import (
	"fmt"

	log "github.com/golang/glog"
)

// JMP REL8
func (cpu *CPU) jmprel8(inst *Inst) error {
	urel, err := inst.Fetch8()
	if err != nil {
		return err
	}
	return cpu.jump8rel(int8(urel))
}

// JMP REL16 (note, this can be a negative displacement)
func (cpu *CPU) jmprel16(inst *Inst) error {
	urel, err := inst.Fetch16()
	if err != nil {
		return err
	}
	return cpu.jump16rel(int16(urel))
}

// JMP far absolute (ptr16:16)
// Example: 0000000E  EA45230000        jmp 0x0:0x2345
func (cpu *CPU) jmpFarAbs(inst *Inst) error {
	offset, err := inst.Fetch16()
	if err != nil {
		return err
	}
	seg, err := inst.Fetch16()
	if err != nil {
		return err
	}
	cpu.Regs.SetSeg16(CS, uint(seg))
	cpu.Ip = offset
	return nil
}

func (cpu *CPU) jmpNearAbsIndirect(inst *Inst) error {
	// near is always 16 bits
	offset := inst.ModRM.GetRm16(cpu, inst)
	cpu.Ip = uint16(offset)
	return nil
}

func (cpu *CPU) callNearAbsIndirect(inst *Inst) error {
	// near is always 16 bits
	offset := inst.ModRM.GetRm16(cpu, inst)
	cpu.Regs.PushSeg16(CS, cpu.Mem)
	cpu.Regs.Push16(cpu.Mem, cpu.Ip)
	cpu.Ip = uint16(offset)
	return nil
}

func (cpu *CPU) jmpFarAbsIndirect(inst *Inst) error {
	segment, offset, err := inst.ModRM.GetMemoryLocation(cpu, inst)
	if err != nil {
		return err
	}

	off := cpu.Mem.GetMem16(segment, offset)
	seg := cpu.Mem.GetMem16(segment, offset+2)

	cpu.Regs.SetSeg16(CS, uint(seg))
	cpu.Ip = uint16(off)
	return nil
}

func (cpu *CPU) callFarAbsIndirect(inst *Inst) error {
	segment, offset, err := inst.ModRM.GetMemoryLocation(cpu, inst)
	if err != nil {
		return err
	}
	cpu.Regs.PushSeg16(CS, cpu.Mem)
	cpu.Regs.Push16(cpu.Mem, cpu.Ip)

	off := cpu.Mem.GetMem16(segment, offset)
	seg := cpu.Mem.GetMem16(segment, offset+2)

	cpu.Regs.SetSeg16(CS, uint(seg))
	cpu.Ip = uint16(off)
	return nil
}

// jump to a relative 16 bit offset.
func (cpu *CPU) jump8rel(rel int8) error {
	if rel < 0 {
		cpu.Ip -= uint16(-rel)
	} else {
		cpu.Ip += uint16(rel)
	}
	return nil
}

// jump
func (cpu *CPU) jump16rel(rel int16) error {
	if rel < 0 {
		cpu.Ip -= uint16(-rel)
	} else {
		cpu.Ip += uint16(rel)
	}
	return nil
}

func (cpu *CPU) jumpIfFlag8(inst *Inst, flag uint32) error {
	rel, err := inst.Fetch8()
	if err != nil {
		return err
	}
	if cpu.Flags.IsEnabled(flag) {
		return cpu.jump8rel(int8(rel))
	}
	return nil
}

// jumpIfNoFlag8 - jump if not flag
func (cpu *CPU) jumpIfNoFlag8(inst *Inst, flag uint32) error {
	rel, err := inst.Fetch8()
	if err != nil {
		return err
	}
	if !cpu.Flags.IsEnabled(flag) {
		return cpu.jump8rel(int8(rel))
	}
	return nil
}

// jo8 - jump if overflow
func (cpu *CPU) jo8(inst *Inst) error {
	return cpu.jumpIfFlag8(inst, OverflowFlag)
}

// jno8 - jump if not overflow
func (cpu *CPU) jno8(inst *Inst) error {
	return cpu.jumpIfNoFlag8(inst, OverflowFlag)
}

// jb8 - jump if below
func (cpu *CPU) jb8(inst *Inst) error {
	return cpu.jumpIfFlag8(inst, CarryFlag)
}

// jnob8 - jump if not below
func (cpu *CPU) jnb8(inst *Inst) error {
	return cpu.jumpIfNoFlag8(inst, CarryFlag)
}

// jz8 - jump if zero
func (cpu *CPU) jz8(inst *Inst) error {
	return cpu.jumpIfFlag8(inst, ZeroFlag)
}

// jnz8 - jump if not zero
func (cpu *CPU) jnz8(inst *Inst) error {
	return cpu.jumpIfNoFlag8(inst, ZeroFlag)
}

// jbe8 - jump if below or equal
func (cpu *CPU) jbe8(inst *Inst) error {
	return cpu.jumpIfFlag8(inst, CarryFlag|ZeroFlag)
}

// ja8 - jump if above
func (cpu *CPU) ja8(inst *Inst) error {
	rel, err := inst.Fetch8()
	if err != nil {
		return err
	}
	if !cpu.Flags.IsEnabled(CarryFlag) && !cpu.Flags.IsEnabled(ZeroFlag) {
		return cpu.jump8rel(int8(rel))
	}
	return nil
}

// js8 - jump if sign
func (cpu *CPU) js8(inst *Inst) error {
	return cpu.jumpIfFlag8(inst, SignFlag)
}

// jnos8 - jump if not sign
func (cpu *CPU) jns8(inst *Inst) error {
	return cpu.jumpIfNoFlag8(inst, SignFlag)
}

// jpe8 - jump if parity even
func (cpu *CPU) jpe8(inst *Inst) error {
	return cpu.jumpIfFlag8(inst, ParityFlag)
}

// jpo8 - jump if parity odd
func (cpu *CPU) jpo8(inst *Inst) error {
	return cpu.jumpIfNoFlag8(inst, ParityFlag)
}

// jl8 - jump if less
func (cpu *CPU) jl8(inst *Inst) error {
	rel, err := inst.Fetch8()
	if err != nil {
		return err
	}
	sf := cpu.Flags.IsEnabled(SignFlag)
	of := cpu.Flags.IsEnabled(OverflowFlag)
	if sf != of {
		return cpu.jump8rel(int8(rel))
	}
	return nil
}

// jle8 - jump if less or equal
func (cpu *CPU) jle8(inst *Inst) error {
	rel, err := inst.Fetch8()
	if err != nil {
		return err
	}
	sf := cpu.Flags.IsEnabled(SignFlag)
	of := cpu.Flags.IsEnabled(OverflowFlag)
	zf := cpu.Flags.IsEnabled(ZeroFlag)
	if zf || sf != of {
		return cpu.jump8rel(int8(rel))
	}
	return nil
}

// jg8 - jump if greater
func (cpu *CPU) jg8(inst *Inst) error {
	rel, err := inst.Fetch8()
	if err != nil {
		return err
	}
	sf := cpu.Flags.IsEnabled(SignFlag)
	of := cpu.Flags.IsEnabled(OverflowFlag)
	zf := cpu.Flags.IsEnabled(ZeroFlag)
	if !zf && sf == of {
		return cpu.jump8rel(int8(rel))
	}
	return nil
}

// jge8 - jump if greater or equal
func (cpu *CPU) jge8(inst *Inst) error {
	rel, err := inst.Fetch8()
	if err != nil {
		return err
	}
	sf := cpu.Flags.IsEnabled(SignFlag)
	of := cpu.Flags.IsEnabled(OverflowFlag)
	if sf == of {
		return cpu.jump8rel(int8(rel))
	}
	return nil
}

// callNear
func (cpu *CPU) callNear(inst *Inst) error {
	off, err := inst.Fetch16()
	if err != nil {
		return err
	}
	cpu.Regs.Push16(cpu.Mem, cpu.Ip)
	return cpu.jump16rel(int16(off))
}

// callFar - call far
func (cpu *CPU) callFar(inst *Inst) error {
	off, err := inst.Fetch16()
	if err != nil {
		return err
	}
	seg, err := inst.Fetch16()
	if err != nil {
		return err
	}
	cpu.Regs.PushSeg16(CS, cpu.Mem)
	cpu.Regs.Push16(cpu.Mem, cpu.Ip)
	cpu.Regs.SetSeg16(CS, uint(seg))
	cpu.Ip = off & 0xffff // mask to 16 bits
	return nil
}

// jcxz - jump if cx is zero
func (cpu *CPU) jcxz(inst *Inst) error {
	// Need to fetch the next value before checking CX
	urel, err := inst.Fetch8()
	if err != nil {
		return err
	}
	if cpu.Regs.GetReg16(CX) == 0 {
		return cpu.jump8rel(int8(urel))
	}
	return nil
}

func (cpu *CPU) loopOnCondition(inst *Inst, cond bool) error {
	urel, err := inst.Fetch8()
	if err != nil {
		return err
	}
	cpu.Regs.Dec16(CX, 1)
	if cpu.Regs.GetReg16(CX) != 0 && cond {
		return cpu.jump8rel(int8(urel))
	}
	return nil
}

func (cpu *CPU) loop(inst *Inst) error {
	return cpu.loopOnCondition(inst, true)
}

func (cpu *CPU) loope(inst *Inst) error {
	zf := cpu.Flags.IsEnabled(ZeroFlag)
	return cpu.loopOnCondition(inst, zf)
}

func (cpu *CPU) loopne(inst *Inst) error {
	zf := cpu.Flags.IsEnabled(ZeroFlag)
	return cpu.loopOnCondition(inst, !zf)
}

// Returns

func (cpu *CPU) retNear(bytesToPop uint) error {
	cpu.Ip = cpu.Regs.Pop16(cpu.Mem)
	if bytesToPop > 0 {
		cpu.Regs.Inc16(SP, bytesToPop)
	}
	return nil
}

func (cpu *CPU) retFar(bytesToPop uint) error {
	cpu.Ip = cpu.Regs.Pop16(cpu.Mem)
	cs := cpu.Regs.Pop16(cpu.Mem)
	cpu.Regs.SetSeg16(CS, uint(cs))
	if bytesToPop > 0 {
		cpu.Regs.Inc16(SP, bytesToPop)
	}
	return nil
}

func (cpu *CPU) int(intrno int) error {
	if cpu.Intrs[intrno] != nil {
		log.V(4).Infof("Call Internal Interrupt # 0x%X", intrno)
		cpu.Intrs[intrno](cpu, intrno)
		return nil
	} else {
		// Call memory based x86 interrupt code
		cpu.Regs.Push16(cpu.Mem, uint16(cpu.Flags.Value()))
		cpu.Regs.PushSeg16(CS, cpu.Mem)
		cpu.Regs.Push16(cpu.Mem, cpu.Ip)

		// Update CS:IP from interrupt table
		cpu.Ip = cpu.Mem.GetMem16(0x0000, uint(intrno*4))
		cs := cpu.Mem.GetMem16(0x0000, 2+uint(intrno*4))
		cpu.Regs.SetSeg16(CS, uint(cs))
	}

	return fmt.Errorf("INT %d: Implement me", intrno)
}
