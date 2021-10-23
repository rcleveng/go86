package go86

import (
	"testing"

	cpu "go86.org/go86/cpu"
	"gotest.tools/v3/assert"
)

func TestDebuggerSmoke(t *testing.T) {
	cpu := cpu.NewCpu(1024 * 1024)
	d := NewDebuggerBackend(cpu, nil, nil)

	if len(d.breakpoints) != 0 {
		t.Error("Expected breakpoints to be empty")
	}
}

func TestDebuggerAddBreakpoint(t *testing.T) {
	c := cpu.NewCpu(1024 * 1024)
	d := NewDebuggerBackend(c, nil, nil)

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
	c := cpu.NewCpu(1024 * 1024)
	d := NewDebuggerBackend(c, nil, nil)

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
	c := cpu.NewCpu(1024 * 1024)
	d := NewDebuggerBackend(c, nil, nil)

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
	c := cpu.NewCpu(1024 * 1024)
	c.Ip = 0
	c.Sregs[cpu.SREG_CS] = 100

	b := Breakpoint{
		seg: 100,
		off: 0,
	}
	assert.Assert(t, b.ShouldBreak(c) == true)

	c.Ip++
	assert.Assert(t, b.ShouldBreak(c) == false)
}
