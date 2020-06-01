package vm

import (
	"errors"
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

func (a *Appleone) load(addr uint16, data []byte) {
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

func (a *Appleone) getAddr(o op) (uint16, error) {
	switch o.addrMode {
	// TODO: will these ever apply here?
	// case accumulator:
	//
	// case implied:
	//
	case absolute:
		return a.nextDWord(), nil
	case absoluteXIndexed:
		return a.nextDWord() + uint16(a.cpu.x), nil
	case absoluteYIndexed:
		return a.nextDWord() + uint16(a.cpu.y), nil
	case immediate:
		return a.cpu.pc - 1, nil
	case indirect:
		return uint16(a.nextWord()), nil
	case indirectXIndexed:
		addr := (uint16(a.nextWord()) + uint16(a.cpu.x)) & 0xFF
		return a.littleEndianToUint16(a.mem[addr+1], a.mem[addr]), nil
	case indirectYIndexed:
		addr := uint16(a.nextWord())
		val := a.littleEndianToUint16(a.mem[addr+1], a.mem[addr])
		return val + uint16(a.cpu.y), nil
	case relative:
		return a.cpu.pc - 1, nil
	case zeroPage:
		return uint16(a.nextWord()) & 0xFF, nil
	case zeroPageXIndexed:
		return (uint16(a.nextWord()) + uint16(a.cpu.x)) & 0xFF, nil
	case zeroPageYIndexed:
		return (uint16(a.nextWord()) + uint16(a.cpu.y)) & 0xFF, nil
	default:
		return 0, errors.New("unkown addressing mode")
	}
}

func (a *Appleone) getOperand(o op) (uint8, error) {
	if o.addrMode == accumulator {
		return a.cpu.a, nil
	}
	b, err := a.getAddr(o)
	if err != nil {
		return 0, err
	}
	return a.mem[b], nil
}

func (a *Appleone) littleEndianToUint16(big, little byte) uint16 {
	return uint16(a.mem[big])<<8 | uint16(a.mem[little])
}

// pushWordToStack pushes the given word (byte) into memory and sets the new stack pointer
func (a *Appleone) pushWordToStack(b byte) {
	a.mem[StackBottom+uint16(a.cpu.sp)] = b
	a.cpu.sp = byte((uint16(a.cpu.sp) - 1) & 0xFF)
}

// pushWordToStack splits the high and low byte of the data passed in, and pushes them to the stack
func (a *Appleone) pushDWordToStack(data uint16) {
	h := byte((data >> 8) & 0xFF)
	l := byte(data & 0xFF)
	a.pushWordToStack(h)
	a.pushWordToStack(l)
}

// popStackWord sets the new stack pointer and returns the appropriate byte in memory
func (a *Appleone) popStackWord() byte {
	a.cpu.sp = byte((uint16(a.cpu.sp) + 1) & 0xFF)
	return a.mem[StackBottom+uint16(a.cpu.sp)]
}

// popStackDWord pops two stack words (a double word - uint16) off the stack
func (a *Appleone) popStackDWord() uint16 {
	l := a.popStackWord()
	h := a.popStackWord()
	return (uint16(h) << 8) | uint16(l)
}

// nextWord returns the next byte in memory
func (a *Appleone) nextWord() byte {
	return a.mem[a.cpu.pc-1]
}

// nextDWord returns the next two bytes (double word)
func (a *Appleone) nextDWord() uint16 {
	return a.littleEndianToUint16(a.mem[a.cpu.pc-1], a.mem[a.cpu.pc-2])
}

func (a *Appleone) setZeroIfNeeded(word byte) {
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

func (a *Appleone) getNegative() byte {
	return a.cpu.ps & flagNegative
}

func (a *Appleone) setNegativeIfOverflow(word byte) {
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

func (a *Appleone) getCarry() byte {
	return a.cpu.sp & flagCarry
}

func (a *Appleone) setCarry() {
	a.cpu.sp |= flagCarry
}

func (a *Appleone) clearCarry() {
	a.cpu.sp &^= flagCarry
}

func (a *Appleone) getOverflow() byte {
	return a.cpu.ps & flagOverflow
}

func (a *Appleone) setOverflow() {
	a.cpu.sp |= flagOverflow
}

func (a *Appleone) clearOverflow() {
	a.cpu.sp &^= flagOverflow
}

func (a *Appleone) getZero() byte {
	return a.cpu.ps & flagZero
}

func (a *Appleone) branch(o op) error {
	offset, err := a.getOperand(o)
	if err != nil {
		return err
	}
	if offset > 127 {
		a.cpu.pc -= 256 - uint16(offset)
	} else {
		a.cpu.pc += uint16(offset)
	}
	return nil
}

func (a *Appleone) compare(b1, b2 byte) {
	a.clearZero()
	a.clearCarry()
	a.clearNegative()

	if b1 == b2 {
		a.setZero()
		a.setCarry()
	}
	if b1 > b2 {
		a.setCarry()
	}

	sum := byte(uint16(b1) - uint16(b2))
	a.setNegativeIfOverflow(sum)
}

func (a *Appleone) setDec() {
	a.cpu.sp |= flagDecimalMode
}

func (a *Appleone) clearDec() {
	a.cpu.sp &^= flagDecimalMode
}

func (a *Appleone) setInterrupt() {
	a.cpu.sp |= flagDisableInterrupts
}

func (a *Appleone) clearInterrupt() {
	a.cpu.sp &^= flagDisableInterrupts
}
