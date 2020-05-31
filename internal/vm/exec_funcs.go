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
	a.cpu.x++
	a.setZeroIfNeeded(a.cpu.x)
	a.setNegativeIfOverflow(a.cpu.x)
	return nil
}

// Y + 1 -> Y                       N Z C I D V
//                                  + + - - - -
func execINY(a *Appleone, o op) error {
	a.cpu.y++
	a.setZeroIfNeeded(a.cpu.y)
	a.setNegativeIfOverflow(a.cpu.y)
	return nil
}

// A -> X                           N Z C I D V
//                                  + + - - - -
func execTAX(a *Appleone, o op) error {
	a.cpu.x = a.cpu.a
	a.setZeroIfNeeded(a.cpu.x)
	a.setNegativeIfOverflow(a.cpu.x)
	return nil
}

// A -> Y                           N Z C I D V
//                                  + + - - - -
func execTAY(a *Appleone, o op) error {
	a.cpu.y = a.cpu.a
	a.setZeroIfNeeded(a.cpu.y)
	a.setNegativeIfOverflow(a.cpu.y)
	return nil
}

// X - 1 -> X                       N Z C I D V
//                                  + + - - - -
func execDEX(a *Appleone, o op) error {
	a.cpu.x--
	a.setZeroIfNeeded(a.cpu.x)
	a.setNegativeIfOverflow(a.cpu.x)
	return nil
}

// Y - 1 -> Y                       N Z C I D V
//                                  + + - - - -
func execDEY(a *Appleone, o op) error {
	a.cpu.y--
	a.setZeroIfNeeded(a.cpu.y)
	a.setNegativeIfOverflow(a.cpu.y)
	return nil
}

// M -> A                           N Z C I D V
//                                  + + - - - -
func execLDA(a *Appleone, o op) error {
	b, err := o.getData(a)
	if err != nil {
		return err
	}
	a.cpu.a = b
	a.setZeroIfNeeded(a.cpu.a)
	a.setNegativeIfOverflow(a.cpu.a)
	return nil
}

// M -> X                           N Z C I D V
//                                  + + - - - -
func execLDX(a *Appleone, o op) error {
	b, err := o.getData(a)
	if err != nil {
		return err
	}
	a.cpu.x = b
	a.setZeroIfNeeded(a.cpu.x)
	a.setNegativeIfOverflow(a.cpu.x)
	return nil
}

// M -> Y                           N Z C I D V
//                                  + + - - - -
func execLDY(a *Appleone, o op) error {
	b, err := o.getData(a)
	if err != nil {
		return err
	}
	a.cpu.y = b
	a.setZeroIfNeeded(a.cpu.y)
	a.setNegativeIfOverflow(a.cpu.y)
	return nil
}

// A + M + C -> A, C                N Z C I D V
//                                  + + + - - +
func execADC(a *Appleone, o op) error {
	b, err := o.getData(a)
	if err != nil {
		return err
	}
	operand := uint16(b)
	regA := uint16(a.cpu.a)
	sum := regA + operand + uint16(a.getCarry())
	a.cpu.a = uint8(sum)

	a.clearCarry()
	if sum > 255 {
		a.setCarry()
	}

	// http://www.righto.com/2012/12/the-6502-overflow-flag-explained.html
	a.clearOverflow()
	if (operand^sum)&(regA^sum)&0x80 != 0 {
		a.setOverflow()
	}

	a.setZeroIfNeeded(a.cpu.a)
	a.setNegativeIfOverflow(a.cpu.a)

	return nil
}
