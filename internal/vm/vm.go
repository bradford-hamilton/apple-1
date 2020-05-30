package vm

import (
	"fmt"
)

// Appleone represents the virtual Apple 1 computer
type Appleone struct {
	cpu *Mos6502 // virtual mos6502 cpu
	mem block    // available memory (64kiB)
}

// New returns a pointer to an initialized Appleone with a brand spankin new CPU
func New() *Appleone {
	return &Appleone{
		cpu: newCPU(),
		mem: newBlock(),
	}
}

func (a *Appleone) load(addr uint16, data []uint8) {
	a.mem.load(addr, data)
	a.cpu.pc = addr
}

func (a *Appleone) step() {
	op, err := opByCode(a.mem[a.cpu.pc])
	if err != nil {
		fmt.Println("TODO")
	}

	a.cpu.pc += uint16(op.size)

	if err := op.exec(a, op); err != nil {
		fmt.Println("TODO")
	}
}

func (a *Appleone) littleEndianToUint16(big, little uint8) uint16 {
	return uint16(a.mem[big])<<8 | uint16(a.mem[little])
}

// pushWordToStack pushes the given word (byte) into memory and sets the new stack pointer
func (a *Appleone) pushWordToStack(b byte) {
	a.mem[StackBottom+uint16(a.cpu.sp)] = b
	a.cpu.sp = uint8((uint16(a.cpu.sp) - 1) & 0xFF)
}

// pushWordToStack splits the high and low byte of the data passed in, and pushes them to the stack
func (a *Appleone) pushDWordToStack(data uint16) {
	h := uint8((data >> 8) & 0xFF)
	l := uint8(data & 0xFF)
	a.pushWordToStack(h)
	a.pushWordToStack(l)
}

// popStackWord sets the new stack pointer and returns the appropriate byte in memory
func (a *Appleone) popStackWord() uint8 {
	a.cpu.sp = uint8((uint16(a.cpu.sp) + 1) & 0xFF)
	return a.mem[StackBottom+uint16(a.cpu.sp)]
}

// popStackDWord pops two stack words (a double word - uint16) off the stack
func (a *Appleone) popStackDWord() uint16 {
	l := a.popStackWord()
	h := a.popStackWord()
	return (uint16(h) << 8) | uint16(l)
}

// nextWord returns the next byte in memory
func (a *Appleone) nextWord() uint8 {
	return a.mem[a.cpu.pc-1]
}

// nextDWord returns the next two bytes (double word)
func (a *Appleone) nextDWord() uint16 {
	return a.littleEndianToUint16(a.mem[a.cpu.pc-1], a.mem[a.cpu.pc-2])
}

func (a *Appleone) setZeroIfNeeded(word uint8) {
	a.clearZero()
	if word == 0 {
		a.setZero()
	}
}

func (a *Appleone) setZero() {
	a.cpu.ps |= flagZero
}

func (a *Appleone) clearZero() {
	a.cpu.sp &^= flagZero
}

func (a *Appleone) setNegativeIfOverflow(word uint8) {
	a.clearNegative()
	if word > 127 {
		a.setNegative()
	}
}

func (a *Appleone) setNegative() {
	a.cpu.ps |= flagZero
}

func (a *Appleone) clearNegative() {
	a.cpu.sp &^= flagZero
}
