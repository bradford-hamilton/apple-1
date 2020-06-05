package vm

import (
	"errors"
	"fmt"
	"time"
)

const clockSpeed = 1 // 1 MHz

// VM represents the Apple 1 virutal machine
type VM struct {
	cpu       *Mos6502      // virtual mos6502 cpu
	mem       block         // available memory (64kiB)
	clock     *time.Ticker  // the "cpu" clock
	ShutdownC chan struct{} //
}

// New returns a pointer to an initialized VM with a brand spankin new CPU
func New() *VM {
	return &VM{
		cpu:   newCPU(),
		mem:   newBlock(),
		clock: time.NewTicker(time.Second / time.Duration(clockSpeed)),
	}
}

// Run starts the vm and emulates a clock that runs by default at 60MHz
// This can be changed with a flag.
func (vm *VM) Run() {
	for {
		select {
		case <-vm.clock.C:
			vm.emulateCycle()
			continue
		case <-vm.ShutdownC:
			break
		}
		break
	}
	vm.sigTerm("gracefully shutting down...")
}

func (vm *VM) emulateCycle() {
	operation, err := operationByCode(vm.mem[vm.cpu.pc])
	if err != nil {
		fmt.Println("TODO")
	}

	vm.cpu.pc += uint16(operation.size)

	if err := operation.exec(vm, operation); err != nil {
		fmt.Println("TODO")
	}
}

func (vm *VM) sigTerm(msg string) {
	fmt.Println(msg)
	vm.ShutdownC <- struct{}{}
}

// load puts the provided data into the apple1's memory block starting at the provided address
func (vm *VM) load(addr uint16, data []byte) {
	vm.mem.load(addr, data)
	vm.cpu.pc = addr
}

func (vm *VM) getAddr(o operation) (uint16, error) {
	switch o.addrMode {
	// TODO: will these ever apply here?
	// case accumulator:
	//
	// case implied:
	//
	case absolute:
		return vm.nextDWord(), nil
	case absoluteXIndexed:
		return vm.nextDWord() + uint16(vm.cpu.x), nil
	case absoluteYIndexed:
		return vm.nextDWord() + uint16(vm.cpu.y), nil
	case immediate:
		return vm.cpu.pc - 1, nil
	case indirect:
		return uint16(vm.nextWord()), nil
	case indirectXIndexed:
		addr := (uint16(vm.nextWord()) + uint16(vm.cpu.x)) & 0xFF
		return vm.littleEndianToUint16(vm.mem[addr+1], vm.mem[addr]), nil
	case indirectYIndexed:
		addr := uint16(vm.nextWord())
		val := vm.littleEndianToUint16(vm.mem[addr+1], vm.mem[addr])
		return val + uint16(vm.cpu.y), nil
	case relative:
		return vm.cpu.pc - 1, nil
	case zeroPage:
		return uint16(vm.nextWord()) & 0xFF, nil
	case zeroPageXIndexed:
		return (uint16(vm.nextWord()) + uint16(vm.cpu.x)) & 0xFF, nil
	case zeroPageYIndexed:
		return (uint16(vm.nextWord()) + uint16(vm.cpu.y)) & 0xFF, nil
	default:
		return 0, errors.New("unkown addressing mode")
	}
}

func (vm *VM) getOperand(o operation) (byte, error) {
	if o.addrMode == accumulator {
		return vm.cpu.a, nil
	}
	b, err := vm.getAddr(o)
	if err != nil {
		return 0, err
	}
	return vm.mem[b], nil
}

func (vm *VM) littleEndianToUint16(big, little byte) uint16 {
	return uint16(vm.mem[big])<<8 | uint16(vm.mem[little])
}

// pushWordToStack pushes the given word (byte) into memory and sets the new stack pointer
func (vm *VM) pushWordToStack(b byte) {
	vm.mem[StackBottom+uint16(vm.cpu.sp)] = b
	vm.cpu.sp = byte((uint16(vm.cpu.sp) - 1) & 0xFF)
}

// pushWordToStack splits the high and low byte of the data passed in, and pushes them to the stack
func (vm *VM) pushDWordToStack(data uint16) {
	h := byte((data >> 8) & 0xFF)
	l := byte(data & 0xFF)
	vm.pushWordToStack(h)
	vm.pushWordToStack(l)
}

// popStackWord sets the new stack pointer and returns the appropriate byte in memory
func (vm *VM) popStackWord() byte {
	vm.cpu.sp = byte((uint16(vm.cpu.sp) + 1) & 0xFF)
	return vm.mem[StackBottom+uint16(vm.cpu.sp)]
}

// popStackDWord pops two stack words (a double word - uint16) off the stack
func (vm *VM) popStackDWord() uint16 {
	l := vm.popStackWord()
	h := vm.popStackWord()
	return (uint16(h) << 8) | uint16(l)
}

// nextWord returns the next byte in memory
func (vm *VM) nextWord() byte {
	return vm.mem[vm.cpu.pc-1]
}

// nextDWord returns the next two bytes (double word)
func (vm *VM) nextDWord() uint16 {
	return vm.littleEndianToUint16(vm.mem[vm.cpu.pc-1], vm.mem[vm.cpu.pc-2])
}

// maybeSetFlagZero takes a single word (byte), clears flagZero, and sets flagZero if word is 0
func (vm *VM) maybeSetFlagZero(word byte) {
	vm.clearFlag(flagZero)
	if word == 0 {
		vm.setFlag(flagZero)
	}
}

func (vm *VM) getFlag(flag byte) byte {
	return vm.cpu.ps & flag
}

func (vm *VM) setFlag(flag byte) {
	vm.cpu.ps |= flag
}

func (vm *VM) clearFlag(flag byte) {
	vm.cpu.ps &^= flag
}

func (vm *VM) maybeSetFlagOverflow(word byte) {
	vm.clearFlag(flagNegative)
	if word > 127 {
		vm.setFlag(flagNegative)
	}
}

// Branch offsets are signed 8-bit values, -128 ... +127, negative offsets in two's
// complement. Page transitions may occur and add an extra cycle to the exucution
func (vm *VM) branch(o operation) error {
	offset, err := vm.getOperand(o)
	if err != nil {
		return err
	}
	if offset > 127 {
		vm.cpu.pc -= 256 - uint16(offset)
	} else {
		vm.cpu.pc += uint16(offset)
	}
	return nil
}

// compare clears zero, carry, and negative flags, compares the two bytes, and sets the
// appropriate flags based on the comparison between the bytes.
func (vm *VM) compare(b1, b2 byte) {
	vm.clearFlag(flagZero)
	vm.clearFlag(flagCarry)
	vm.clearFlag(flagNegative)

	if b1 == b2 {
		vm.setFlag(flagZero)
		vm.setFlag(flagCarry)
	}
	if b1 > b2 {
		vm.setFlag(flagCarry)
	}

	b := byte(uint16(b1) - uint16(b2))
	vm.maybeSetFlagOverflow(b)
}

func (vm *VM) setMem(o operation, operand byte) error {
	addr, err := vm.getAddr(o)
	if err != nil {
		return err
	}
	vm.mem[addr] = operand
	return nil
}

// func (vm *VM) execBRK(o operation) error {
// 	// set processer status flag to BRK
// 	vm.cpu.ps = flagBreak

// 	vm.pushDWordToStack(vm.cpu.pc + 1)
// 	vm.pushWordToStack(vm.cpu.ps)

// 	vm.setFlag(flagDisableInterrupts)
// 	vm.cpu.pc = uint16(vm.mem[0xFFFF])<<8 | uint16(vm.mem[0xFFFE])

// 	return nil
// }
