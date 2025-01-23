package go86

import (
	"fmt"
	"testing"
)

// MockCpuInstructionReader is a simple mock for CpuInstructionReader.
type MockCpuInstructionReader struct {
	fetch8Called  bool
	fetch16Called bool
	fetch8Val     uint8
	fetch16Val    uint16
}

func (m *MockCpuInstructionReader) Fetch8() (uint8, error) {
	m.fetch8Called = true
	return m.fetch8Val, nil
}

func (m *MockCpuInstructionReader) Fetch16() (uint16, error) {
	m.fetch16Called = true
	return m.fetch16Val, nil
}

// TestNewModRM runs a sub-test for every possible byte (0x00 to 0xff).
func TestModRM(t *testing.T) {
	for i := 0; i < 256; i++ {
		b := uint8(i)

		t.Run(fmt.Sprintf("b=0x%02X", b), func(t *testing.T) {
			// Pre-calculate expected fields based on b
			mod := (b & 0xC0) >> 6
			reg := (b & 0x38) >> 3
			rm := b & 0x07

			// Set up our mock to return specific Fetch8/Fetch16 values
			mock := &MockCpuInstructionReader{
				fetch8Val:  0x7F,   // 127 in decimal
				fetch16Val: 0x1234, // 4660 in decimal
			}

			got, err := NewModRM(mock, b)
			if err != nil {
				t.Fatalf("NewModRM returned unexpected error: %v", err)
			}

			// Check basic decoding
			if got.raw != b {
				t.Errorf("raw = 0x%02X, want 0x%02X", got.raw, b)
			}
			if got.Mod != mod {
				t.Errorf("Mod = %d, want %d", got.Mod, mod)
			}
			if got.Reg != reg {
				t.Errorf("Reg = %d, want %d", got.Reg, reg)
			}
			if got.Rm != rm {
				t.Errorf("Rm = %d, want %d", got.Rm, rm)
			}

			// Check which fetch call(s) should have been made
			expectFetch16 := (mod == 2) || (mod == 0 && rm == 6)
			expectFetch8 := (mod == 1)

			if expectFetch16 && !mock.fetch16Called {
				t.Errorf("Fetch16 was not called, but expected for b=0x%02X (mod=%d, rm=%d)", b, mod, rm)
			}
			if !expectFetch16 && mock.fetch16Called {
				t.Errorf("Fetch16 was called unexpectedly for b=0x%02X (mod=%d, rm=%d)", b, mod, rm)
			}
			if expectFetch8 && !mock.fetch8Called {
				t.Errorf("Fetch8 was not called, but expected for b=0x%02X (mod=%d)", b, mod)
			}
			if !expectFetch8 && mock.fetch8Called {
				t.Errorf("Fetch8 was called unexpectedly for b=0x%02X (mod=%d)", b, mod)
			}

			// If we expected a disp16, check that it matches
			if expectFetch16 {
				wantDisp16 := int16(mock.fetch16Val)
				if got.Disp16 != wantDisp16 {
					t.Errorf("Disp16 = %d, want %d", got.Disp16, wantDisp16)
				}
				// Disp8 should remain default (0) in this case.
				if got.Disp8 != 0 {
					t.Errorf("Disp8 = %d, want 0 (when Disp16 is used)", got.Disp8)
				}
			}

			// If we expected a disp8, check that it matches
			if expectFetch8 {
				wantDisp8 := int8(mock.fetch8Val)
				if got.Disp8 != wantDisp8 {
					t.Errorf("Disp8 = %d, want %d", got.Disp8, wantDisp8)
				}
				// Disp16 should remain default (0) in this case.
				if got.Disp16 != 0 {
					t.Errorf("Disp16 = %d, want 0 (when Disp8 is used)", got.Disp16)
				}
			}

			// If neither disp8 nor disp16 were expected, ensure they're defaulted
			if !expectFetch16 && !expectFetch8 {
				if got.Disp16 != 0 {
					t.Errorf("Disp16 = %d, want 0 (no displacement expected)", got.Disp16)
				}
				if got.Disp8 != 0 {
					t.Errorf("Disp8 = %d, want 0 (no displacement expected)", got.Disp8)
				}
			}
		})
	}
}
