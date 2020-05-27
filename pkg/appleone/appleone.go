package appleone

import "github.com/bradford-hamilton/apple-1/internal/cpu"

// Appleone represents our virtual Apple 1 computer
type Appleone struct {
	cpu *cpu.Mos6502
}

// New returns a pointer to an initialized Appleone with a brand spankin new CPU
func New() *Appleone {
	return &Appleone{cpu: cpu.New()}
}
