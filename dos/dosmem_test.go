package go86

import (
	"fmt"
	"testing"

	"gotest.tools/v3/assert"
)

func TestDosMemSmoke(t *testing.T) {
	m := NewDosMem(0x0800, 0x9FC0)
	b, err := m.Allocate(0x3200) // 200k in 16 byte pages
	assert.NilError(t, err)
	if b != nil {
		assert.Equal(t, b.End-b.Start, 0x3200)
		fmt.Println(b)
	} else {
		t.Fatal("b should not be nil")
	}

	b, err = m.Allocate(0x1000) // 64K
	assert.NilError(t, err)
	assert.Check(t, b != nil)

	b, err = m.Allocate(0x100000) // TOO MUCH
	assert.Check(t, err != nil)
	assert.Check(t, b == nil)
}

func TestDosMemDump(t *testing.T) {
	m := NewDosMem(0x0800, 0x9FC0)
	m.Allocate(0x3200) // 200k in 16 byte pages
	fmt.Println(m)
}

func TestDosMemLast(t *testing.T) {
	m := NewDosMem(0x0800, 0x9FC0)
	m.Fit = Last
	b, err := m.Allocate(0x3200) // 200k in 16 byte pages
	assert.NilError(t, err)
	if b != nil {
		assert.Equal(t, b.End-b.Start, 0x3200)
		fmt.Println(b)
	} else {
		t.Fatal("b should not be nil")
	}

	b2, err := m.Allocate(0x1000) // 64K
	assert.NilError(t, err)
	if b2 != nil {
		assert.Check(t, b2.Start < b.Start)
	} else {
		t.Fail()
	}
}

func TestDosMemRealloc(t *testing.T) {
	m := NewDosMem(0x0800, 0x9FC0)
	b, err := m.Allocate(0x3200) // 200k in 16 byte pages
	assert.NilError(t, err)
	if b == nil {
		t.Fatal("b should not be nil")
	}
	assert.Equal(t, b.End-b.Start, 0x3200)
	fmt.Println(b)

	newsize, err := m.Resize(b.Start, 0x6400) // 400k
	assert.NilError(t, err)
	assert.Equal(t, newsize, 0x6400)
	assert.Equal(t, len(m.Blocks), 2)
	fmt.Println(m)
}

func TestDosMemReallocThree(t *testing.T) {
	m := NewDosMem(0x0800, 0x9FC0)
	b, err := m.Allocate(0x3200) // 200k in 16 byte pages
	assert.NilError(t, err)
	if b == nil {
		t.Fatal("b should not be nil")
	}
	b2, err := m.Allocate(0x1000) // 200k in 16 byte pages
	assert.NilError(t, err)
	// mark it avail
	b2.Avail = true

	newsize, err := m.Resize(b.Start, 0x6400) // 400k
	fmt.Println(m)
	assert.NilError(t, err)
	assert.Equal(t, newsize, 0x6400)
	assert.Equal(t, len(m.Blocks), 2)
}
