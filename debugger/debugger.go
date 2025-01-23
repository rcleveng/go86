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
	cs := uint16(c.Regs.CS())
	if cs == b.seg && b.off == c.Ip {
		log.V(1).Infof("Breaking at: [%04X:%04X]", cs, c.Ip)
		return true
	}
	return false
}

type DebugCommand int

const (
	CONTINUE DebugCommand = iota
	DETACH
	HALT
	INFO
	HEARTBEAT
	// Command to single step
	STEP
	// Stepped automatically
	STEPPED
	MEMORY
	STOP_REASON
	UNKNOWN_COMMAND
)

type DebuggerMode int8

const (
	STEPPING DebuggerMode = iota
	RUNNING
)

type DebugReponseType int

// Information about the memory block requested by the debugger
type DebuggerMemoryRequest struct {
	Seg    uint
	Off    uint
	Length int
}

type DebuggerRequest struct {
	Cmd DebugCommand
	// The raw command text that triggered this debugger request
	RawCmdText string
	Data       string
	Mem        DebuggerMemoryRequest
}

type DebuggerResponse struct {
	Type DebugReponseType
	// The command that necessiated this response
	Cmd DebugCommand
	// The raw command text that necessiated this debugger response
	RawCmdText string

	// Whatever free form text, should not be parsed at all
	Text string

	// CPU State
	Ip    uint16
	Flags uint32

	// returned memory,  the first position of the slice is the start of the range requested.
	Mem []byte

	// registers
	AX uint16
	BX uint16
	CX uint16
	DX uint16
	SP uint16
	BP uint16
	SI uint16
	DI uint16

	// segment registers
	CS uint16
	DS uint16
	ES uint16
	SS uint16

	// Current thread ID (if needed)
	ThreadId int
	// Last signal number that caused the program to halt.
	Signal int
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

func CpuString(c *cpu.CPU) string {
	l1 := fmt.Sprintf("AX=%04X BX=%04X CX=%04X DX=%04X SP=%04X BP=%04X SI=%04X DI=%04X",
		c.Regs.GetReg16(cpu.AX), c.Regs.GetReg16(cpu.BX),
		c.Regs.GetReg16(cpu.CX), c.Regs.GetReg16(cpu.DX),
		c.Regs.GetReg16(cpu.SP), c.Regs.GetReg16(cpu.BP),
		c.Regs.GetReg16(cpu.SI), c.Regs.GetReg16(cpu.DI))
	l2 := fmt.Sprintf("DS=%04X ES=%04X SS=%04X CS=%04X IP=%04X %s",
		c.Regs.DS(), c.Regs.ES(),
		c.Regs.SS(), c.Regs.CS(), c.Ip, c.Flags.ToCodeViewDebugString())

	return fmt.Sprintf("%s\n%s\n", l1, l2)
}

// Returns the next instruction as a disasmembled string
// Example: 0E06:004E BE0010            MOV     SI,1000
func DisasmString(c *cpu.CPU) string {
	prefix := ""
	//	if c.Inst.Prefix[0] != 0 {
	//	prefix = fmt.Sprintf("[%v]", c.Inst.Prefix[0])
	//}
	disam := c.Mem.At(c.Regs.CS(), uint(c.Ip))[:c.Inst.Len]
	return fmt.Sprintf("%04X:%04X %-18s %s %#v\n",
		c.Regs.CS(), c.Ip, hex.EncodeToString(disam), prefix, c.Inst)
}

func (d *DebuggerBackend) Step() bool {
	log.V(1).Infof("[%04X:%04X] Step", d.cpu.Regs.CS(), d.cpu.Ip)
	// Handle any debugger requests first
	if !d.ShouldBreak() {
		return true
	}

	// Send a basic response with just the next set of instructions disasmembled
	resp := DebuggerResponse{}
	resp.Cmd = STEP
	resp.Ip = d.cpu.Ip
	resp.Flags = d.cpu.Flags.Value()
	resp.Text = DisasmString(d.cpu)
	d.response <- resp

	for r := range d.request {
		log.Infoln("Handling Request in Step: ", r)
		resp := DebuggerResponse{}
		resp.Cmd = r.Cmd
		resp.RawCmdText = r.RawCmdText
		resp.Ip = d.cpu.Ip
		resp.Flags = d.cpu.Flags.Value()
		switch r.Cmd {
		case CONTINUE:
			d.mode = RUNNING
			resp.Text = "Continuing"
			d.response <- resp
			resp.Signal = 0
			return true
		case DETACH:
			d.cpu.Debugger = nil
			resp.Text = "Detaching"
			d.response <- resp
			d.mode = RUNNING
			resp.Signal = 0
			return true
		case STEP:
			d.mode = STEPPING
			resp.Signal = 5
			resp.Text = DisasmString(d.cpu)
			d.response <- resp
			return true
		case HALT:
			d.cpu.Running = false
			resp.Text = "Halting"
			resp.Signal = 5
			d.response <- resp
			return false
		case STOP_REASON:
			resp.Text = "Stop Reason"
			resp.ThreadId = 0
			resp.Signal = 5
			d.response <- resp
		case INFO:
			resp.AX = uint16(d.cpu.Regs.GetReg16(cpu.AX))
			resp.BX = uint16(d.cpu.Regs.GetReg16(cpu.BX))
			resp.CX = uint16(d.cpu.Regs.GetReg16(cpu.CX))
			resp.DX = uint16(d.cpu.Regs.GetReg16(cpu.DX))
			resp.SP = uint16(d.cpu.Regs.GetReg16(cpu.SP))
			resp.BP = uint16(d.cpu.Regs.GetReg16(cpu.BP))
			resp.SI = uint16(d.cpu.Regs.GetReg16(cpu.SI))
			resp.DI = uint16(d.cpu.Regs.GetReg16(cpu.DI))

			resp.CS = uint16(d.cpu.Regs.CS())
			resp.DS = uint16(d.cpu.Regs.DS())
			resp.ES = uint16(d.cpu.Regs.ES())
			resp.SS = uint16(d.cpu.Regs.SS())

			resp.Text = CpuString(d.cpu)
			resp.ThreadId = 0
			resp.Signal = 5
			d.response <- resp
		case MEMORY:
			resp.Mem = d.cpu.Mem.At(r.Mem.Seg, r.Mem.Off)[:r.Mem.Length]
			resp.Text = hex.EncodeToString(resp.Mem)
			d.response <- resp
		case HEARTBEAT:
			resp.Text = "Heartbeat"
			d.response <- resp
		default:
			log.Warningln("Unknown request: ", r)
			resp.Text = "Unknown request" + r.RawCmdText
			d.response <- resp
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
	go listen(port, dbgType, request, response)
}
