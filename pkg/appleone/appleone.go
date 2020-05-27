package appleone

import "github.com/bradford-hamilton/apple-1/internal/cpu"

// Appleone TODO docs
type Appleone struct {
	cpu *cpu.Mos6502
}

// New TODO: docs
func New() *Appleone {
	return &Appleone{cpu: cpu.New()}
}
