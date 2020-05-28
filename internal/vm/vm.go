package vm

import "github.com/bradford-hamilton/apple-1/internal/cpu"

// Appleone represents the virtual Apple 1 computer
type Appleone struct {
	cpu *cpu.Mos6502    // virtual Mos6502 cpu
	mem [64 * 1024]byte // available memory (64kiB)
}

// New returns a pointer to an initialized Appleone with a brand spankin new CPU
func New() *Appleone {
	return &Appleone{
		cpu: cpu.New(),
		mem: [64 * 1024]byte{},
	}
}
