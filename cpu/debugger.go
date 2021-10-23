package go86

import (
	"fmt"

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

type DebugCommand int

const (
	CONTINUE = iota
	DETACH
	HALT
	INFO
	HEARTBEAT
	STEP
	UNKNOWN_COMMAND
)

type DebuggerMode int8

const (
	STEPPING = iota
	RUNNING
)

type DebuggerRequest struct {
	Cmd  DebugCommand
	Data string
}

type DebuggerResponse struct {
	Text string
}

type DebuggerBackend struct {
	cpu         *CPU
	breakpoints []Breaker
	request     chan DebuggerRequest
	response    chan DebuggerResponse
	mode        DebuggerMode
}

func NewDebuggerBackend(cpu *CPU, request chan DebuggerRequest, response chan DebuggerResponse) *DebuggerBackend {
	log.V(4).Infoln("NewDebuger")
	return &DebuggerBackend{
		cpu:         cpu,
		breakpoints: make([]Breaker, 0, 20),
		request:     request,
		response:    response,
		mode:        STEPPING,
	}
}

func (d *DebuggerBackend) ShouldBreak() bool {
	// Are we in single step mode?
	if d.mode == STEPPING {
		return true
	}

	// Now look for breakpoints to see if the debugger should break.
	for _, bp := range d.breakpoints {
		if bp.ShouldBreak(d.cpu) {
			return true
		}
	}
	return false
}

func (d *DebuggerBackend) Step() bool {
	log.V(1).Infof("[%04X:%04X] Step", d.cpu.Sregs[SREG_CS], d.cpu.Ip)

	// Handle any debugger requests first
	if !d.ShouldBreak() {
		return true
	}
	for r := range d.request {
		log.Infoln("Handling Request in Step: ", r)
		switch r.Cmd {
		case CONTINUE:
			d.mode = RUNNING
			return true
		case DETACH:
			d.cpu.debugger = nil
			d.mode = RUNNING
			return true
		case STEP:
			d.mode = STEPPING
			return true
		case HALT:
			return false
		case INFO:
			resp := DebuggerResponse{}
			resp.Text = fmt.Sprintf("AX: %x", d.cpu.Regs[REG_AX])
			d.response <- resp
		case HEARTBEAT:
			resp := DebuggerResponse{}
			resp.Text = "Heartbeat"
			d.response <- resp
		default:
			log.Infoln("Inknown request: ", r)
		}
	}

	return true
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
