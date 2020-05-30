package vm

import "fmt"

func todo(a *Appleone, o op) error {
	fmt.Println("implement me")
	return nil
}

// interrupt,                       N Z C I D V
// push PC+2, push SR               - - - 1 - -
func execBRK(a *Appleone, o op) error {
	// set processer status flag to BRK
	a.cpu.ps = flagBreak

	a.pushDWordToStack(a.cpu.pc + 1)
	a.pushWordToStack(a.cpu.ps)

	a.cpu.ps |= flagDisableInterrupts
	a.cpu.pc = uint16(a.mem[0xFFFF])<<8 | uint16(a.mem[0xFFFE])

	return nil
}

// pull SR, pull PC                 N Z C I D V
// from stack
func execRTI(a *Appleone, o op) error {
	a.cpu.ps = a.popStackWord()
	a.cpu.pc = a.popStackDWord()
	return nil
}

// M - 1 -> M                       N Z C I D V
//                                  + + - - - -
func execDEC(a *Appleone, o op) error {
	addr, err := o.getAddr(a)
	if err != nil {
		return err
	}
	b := a.mem[addr]
	b--
	a.mem[addr] = b
	a.setZeroIfNeeded(b)
	a.setNegativeIfOverflow(b)
	return nil
}

// M + 1 -> M                       N Z C I D V
//                                  + + - - - -
func execINC(a *Appleone, o op) error {
	addr, err := o.getAddr(a)
	if err != nil {
		return err
	}
	b := a.mem[addr]
	b++
	a.mem[addr] = b
	a.setZeroIfNeeded(b)
	a.setNegativeIfOverflow(b)
	return nil
}

// X + 1 -> X                       N Z C I D V
//                                  + + - - - -
func execINX(a *Appleone, o op) error {
	b := a.cpu.x + 1
	a.cpu.x = b
	a.setZeroIfNeeded(b)
	a.setNegativeIfOverflow(b)
	return nil
}
