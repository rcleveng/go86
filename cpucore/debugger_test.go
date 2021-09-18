package go86

import (
	"testing"

	"gotest.tools/v3/assert"
)

func TestDebuggerSmoke(t *testing.T) {
	cpu := NewCpu(1024 * 1024)
	d := NewDebuggerBackend(cpu)

	if len(d.breakpoints) != 0 {
		t.Error("Expected breakpoints to be empty")
	}
}

func TestDebuggerAddBreakpoint(t *testing.T) {
	cpu := NewCpu(1024 * 1024)
	d := NewDebuggerBackend(cpu)

	b := Breakpoint{
		seg: 100,
		off: 0,
	}
	d.AddBreakpoint(b)

	if len(d.breakpoints) != 1 {
		t.Error("Expected breakpoints to be 1")
	}
}

func TestDebuggerAddBreakpointDuplicate(t *testing.T) {
	cpu := NewCpu(1024 * 1024)
	d := NewDebuggerBackend(cpu)

	b := Breakpoint{
		seg: 100,
		off: 0,
	}
	d.AddBreakpoint(b)
	d.AddBreakpoint(b)

	if len(d.breakpoints) != 1 {
		t.Error("Expected breakpoints to be 1")
	}
}

func TestDebuggerAddBreakpointRemove(t *testing.T) {
	cpu := NewCpu(1024 * 1024)
	d := NewDebuggerBackend(cpu)

	b1 := Breakpoint{
		seg: 100,
		off: 0,
	}
	b2 := Breakpoint{
		seg: 200,
		off: 0,
	}
	d.AddBreakpoint(b1)
	d.AddBreakpoint(b2)
	assert.Assert(t, d.RemoveBreakpoint(b1) == true)
	assert.Assert(t, d.RemoveBreakpoint(b1) == false)

	if len(d.breakpoints) != 1 {
		t.Error("Expected breakpoints to be 1")
	}
	assert.Equal(t, d.breakpoints[0], b2)
}

func TestDebuggerShouldBreak(t *testing.T) {
	cpu := NewCpu(1024 * 1024)
	cpu.Ip = 0
	cpu.Sregs[SREG_CS] = 100

	b := Breakpoint{
		seg: 100,
		off: 0,
	}
	assert.Assert(t, b.ShouldBreak(cpu) == true)

	cpu.Ip++
	assert.Assert(t, b.ShouldBreak(cpu) == false)
}
