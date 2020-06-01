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

// load puts the provided data into the apple1's memory block starting at the provided address
func (a *Appleone) load(addr uint16, data []byte) {
	a.mem.load(addr, data)
	a.cpu.pc = addr
}

func (a *Appleone) step() {
	operation, err := operationByCode(a.mem[a.cpu.pc])
	if err != nil {
		fmt.Println("TODO")
	}

	a.cpu.pc += uint16(operation.size)

	if err := operation.exec(a, operation); err != nil {
		fmt.Println("TODO")
	}
}

func (a *Appleone) getAddr(o operation) (uint16, error) {
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

func (a *Appleone) getOperand(o operation) (byte, error) {
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

// maybeSetFlagZero takes a single word (byte), clears flagZero, and sets flagZero if word is 0
func (a *Appleone) maybeSetFlagZero(word byte) {
	a.clearFlag(flagZero)
	if word == 0 {
		a.setFlag(flagZero)
	}
}

func (a *Appleone) getFlag(flag byte) byte {
	return a.cpu.ps & flag
}

func (a *Appleone) setFlag(flag byte) {
	a.cpu.ps |= flag
}

func (a *Appleone) clearFlag(flag byte) {
	a.cpu.ps &^= flag
}

func (a *Appleone) maybeSetFlagOverflow(word byte) {
	a.clearFlag(flagNegative)
	if word > 127 {
		a.setFlag(flagNegative)
	}
}

// Branch offsets are signed 8-bit values, -128 ... +127, negative offsets in two's
// complement. Page transitions may occur and add an extra cycle to the exucution
func (a *Appleone) branch(o operation) error {
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

// compare clears zero, carry, and negative flags, compares the two bytes, and sets the
// appropriate flags based on the comparison between the bytes.
func (a *Appleone) compare(b1, b2 byte) {
	a.clearFlag(flagZero)
	a.clearFlag(flagCarry)
	a.clearFlag(flagNegative)

	if b1 == b2 {
		a.setFlag(flagZero)
		a.setFlag(flagCarry)
	}
	if b1 > b2 {
		a.setFlag(flagCarry)
	}

	b := byte(uint16(b1) - uint16(b2))
	a.maybeSetFlagOverflow(b)
}

func (a *Appleone) setMem(o operation, operand byte) error {
	addr, err := a.getAddr(o)
	if err != nil {
		return err
	}
	a.mem[addr] = operand
	return nil
}
