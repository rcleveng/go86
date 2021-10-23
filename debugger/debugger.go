package go86

import (
	"fmt"
	"sync"

	log "github.com/golang/glog"
	cpu "go86.org/go86/cpu"
)

// Interface type for a breakpoint
type Breaker interface {
	// If this breakpoint should stop at the current s
	ShouldBreak(*cpu.CPU) bool
}

type Breakpoint struct {
	seg uint16
	off uint16
}

func (b Breakpoint) ShouldBreak(c *cpu.CPU) bool {
	cs := c.Sregs[cpu.SREG_CS]
	if cs == b.seg && b.off == c.Ip {
		log.V(1).Infof("Breaking at: [%04X:%04X]", cs, c.Ip)
		return true
	}
	return false
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
	cpu         *cpu.CPU
	breakpoints []Breaker
	request     chan DebuggerRequest
	response    chan DebuggerResponse
	mode        DebuggerMode
	interrupted bool
	mu          sync.RWMutex // for interrupted
}

func NewDebuggerBackend(cpu *cpu.CPU, request chan DebuggerRequest, response chan DebuggerResponse) *DebuggerBackend {
	log.V(4).Infoln("NewDebuggerBackend")
	return &DebuggerBackend{
		cpu:         cpu,
		breakpoints: make([]Breaker, 0, 20),
		request:     request,
		response:    response,
		mode:        STEPPING,
	}
}

func (d *DebuggerBackend) Interrupt() {
	d.mu.Lock()
	d.interrupted = true
	d.mu.Unlock()
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
	d.mu.RLock()
	defer d.mu.RUnlock()
	return d.interrupted
}

func (d *DebuggerBackend) Step() bool {
	log.V(1).Infof("[%04X:%04X] Step", d.cpu.Sregs[cpu.SREG_CS], d.cpu.Ip)
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
			d.cpu.Debugger = nil
			d.mode = RUNNING
			return true
		case STEP:
			d.mode = STEPPING
			return true
		case HALT:
			d.cpu.Running = false
			return false
		case INFO:
			resp := DebuggerResponse{}
			resp.Text = fmt.Sprintf("AX: %x", d.cpu.Regs[cpu.REG_AX])
			d.response <- resp
		case HEARTBEAT:
			resp := DebuggerResponse{}
			resp.Text = "Heartbeat"
			d.response <- resp
		default:
			log.Warningln("Inknown request: ", r)
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

func EnableDebugger(c *cpu.CPU, request chan DebuggerRequest, response chan DebuggerResponse) {
	c.Debugger = NewDebuggerBackend(c, request, response)
}
