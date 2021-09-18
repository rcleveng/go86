package go86

import (
	log "github.com/golang/glog"
)

// Interface type for a breakpoint
type Breaker interface {
	// If this breakpoint should stop at the current s
	ShouldBreak(cpu *CPU) bool
}

type Breakpoint struct {
	seg uint16
	off uint16
}

func (b Breakpoint) ShouldBreak(cpu *CPU) bool {
	cs := cpu.Sregs[SREG_CS]
	if cs == b.seg && b.off == cpu.Ip {
		log.V(1).Infof("Breaking at: [%04X:%04X]", cs, cpu.Ip)
		return true
	}
	return false
}

type Debugger interface {
	Step() bool
	Intr() bool
}

type DebuggerBackend struct {
	cpu         *CPU
	breakpoints []Breaker
}

func NewDebuggerBackend(cpu *CPU) *DebuggerBackend {
	log.V(4).Infoln("NewDebuger")
	return &DebuggerBackend{
		cpu:         cpu,
		breakpoints: make([]Breaker, 0, 20),
	}
}

func (d *DebuggerBackend) Step() bool {
	log.V(1).Infof("[%04X:%04X] Step", d.cpu.Sregs[SREG_CS], d.cpu.Ip)
	for _, bp := range d.breakpoints {
		if bp.ShouldBreak(d.cpu) {
			return true
		}
	}
	return false
}

func (d *DebuggerBackend) Intr() bool {
	return false
}

func (d *DebuggerBackend) AddBreakpoint(b Breakpoint) {
	d.RemoveBreakpoint(b)
	d.breakpoints = append(d.breakpoints, b)
}

func (d *DebuggerBackend) RemoveBreakpoint(sb Breakpoint) bool {
	for i, rb := range d.breakpoints {
		if rb == sb {
			d.breakpoints = append(d.breakpoints[:i], d.breakpoints[i+1:]...)
			return true
		}
	}
	return false
}
