package go86

import (
	"encoding/hex"
	"fmt"
	"net"
	"strconv"
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
	Ip   string
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

	// Always send a response with the CS:IP
	resp := DebuggerResponse{}
	resp.Ip = fmt.Sprintf("%04X:%04X", d.cpu.Sregs[cpu.SREG_CS], d.cpu.Ip)

	prefix := ""
	if d.cpu.Inst.Prefix[0] != 0 {
		prefix = fmt.Sprintf("[%v]", d.cpu.Inst.Prefix[0])
	}
	disam := d.cpu.Mem.At(int(d.cpu.Sregs[cpu.SREG_CS]), int(d.cpu.Ip))[:d.cpu.Inst.Len]
	// 0E06:004E BE0010            MOV     SI,1000
	resp.Text = fmt.Sprintf("%04X:%04X %-18s %s%s\n",
		d.cpu.Sregs[cpu.SREG_CS], d.cpu.Ip, hex.EncodeToString(disam), prefix, d.cpu.Inst)

	d.response <- resp

	for r := range d.request {
		log.Infoln("Handling Request in Step: ", r)
		switch r.Cmd {
		case CONTINUE:
			d.mode = RUNNING
			resp.Text = "Continuing"
			d.response <- resp
			return true
		case DETACH:
			d.cpu.Debugger = nil
			resp.Text = "Detaching"
			d.response <- resp
			d.mode = RUNNING
			return true
		case STEP:
			d.mode = STEPPING
			return true
		case HALT:
			d.cpu.Running = false
			resp.Text = "Halting"
			d.response <- resp
			return false
		case INFO:
			resp := DebuggerResponse{}
			l1 := fmt.Sprintf("AX=%04X BX=%04X CX=%04X DX=%04X SP=%04X BP=%04X SI=%04X DI=%04X",
				d.cpu.Regs[cpu.REG_AX], d.cpu.Regs[cpu.REG_BX], d.cpu.Regs[cpu.REG_CX], d.cpu.Regs[cpu.REG_DX],
				d.cpu.Regs[cpu.REG_SP], d.cpu.Regs[cpu.REG_BP], d.cpu.Regs[cpu.REG_SI], d.cpu.Regs[cpu.REG_DI])
			l2 := fmt.Sprintf("DS=%04X ES=%04X SS=%04X CS=%04X IP=%04X %s",
				d.cpu.Sregs[cpu.SREG_DS], d.cpu.Sregs[cpu.SREG_ES],
				d.cpu.Sregs[cpu.SREG_SS], d.cpu.Sregs[cpu.SREG_CS], d.cpu.Ip, d.cpu.FlagsToDebugString())

			resp.Text = fmt.Sprintf("%s\n%s\n", l1, l2)
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

func listen(port int, dbgtype string, request chan DebuggerRequest, response chan DebuggerResponse) error {
	lsock, err := net.Listen("tcp4", ":"+strconv.Itoa(port))
	if err != nil {
		return err
	}
	defer lsock.Close()
	log.Infof("Listening on port %v", port)

	for {
		conn, err := lsock.Accept()
		if err != nil {
			log.Infof("Got error from Accept: %v", err)
			return err
		}

		if dbgtype == "gdb" {
			session := GdbSession{conn: conn}
			go session.HandleRequests(request)
			go session.HandleResponses(response)
		} else if dbgtype == "lame" {
			session := LameSession{conn: conn}
			go session.HandleRequests(request)
			go session.HandleResponses(response)
		} else {
			panic("Unknown dbgtype: " + dbgtype)
		}
	}
}

func EnableDebugger(c *cpu.CPU, port int, dbgType string, request chan DebuggerRequest, response chan DebuggerResponse) {
	c.Debugger = NewDebuggerBackend(c, request, response)
	go listen(1234, dbgType, request, response)
}
