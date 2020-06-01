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
	addr, err := a.getAddr(o)
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
	addr, err := a.getAddr(o)
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
	operand, err := a.getOperand(o)
	if err != nil {
		return err
	}
	a.cpu.a = operand
	a.setZeroIfNeeded(a.cpu.a)
	a.setNegativeIfOverflow(a.cpu.a)
	return nil
}

// M -> X                           N Z C I D V
//                                  + + - - - -
func execLDX(a *Appleone, o op) error {
	operand, err := a.getOperand(o)
	if err != nil {
		return err
	}
	a.cpu.x = operand
	a.setZeroIfNeeded(a.cpu.x)
	a.setNegativeIfOverflow(a.cpu.x)
	return nil
}

// M -> Y                           N Z C I D V
//                                  + + - - - -
func execLDY(a *Appleone, o op) error {
	operand, err := a.getOperand(o)
	if err != nil {
		return err
	}
	a.cpu.y = operand
	a.setZeroIfNeeded(a.cpu.y)
	a.setNegativeIfOverflow(a.cpu.y)
	return nil
}

// A + M + C -> A, C                N Z C I D V
//                                  + + + - - +
func execADC(a *Appleone, o op) error {
	b, err := a.getOperand(o)
	if err != nil {
		return err
	}
	operand := uint16(b)
	regA := uint16(a.cpu.a)
	sum := regA + operand + uint16(a.getCarry())
	a.cpu.a = byte(sum)

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

// A - M - C -> A                   N Z C I D V
//                                  + + + - - +
func execSBC(a *Appleone, o op) error {
	operand, err := a.getOperand(o)
	if err != nil {
		return err
	}

	carry := uint16(1 - a.getCarry())
	regA := a.cpu.a
	sum := uint16(regA) - carry - uint16(operand)
	a.cpu.a = byte(sum)
	a.clearOverflow()

	if byte(regA)>>7 != a.cpu.a>>7 {
		a.setOverflow()
	}

	if uint16(sum) < 256 {
		a.setCarry()
	} else {
		a.clearCarry()
	}

	a.clearOverflow()
	if ((255-operand)^a.cpu.a)&(regA^a.cpu.a)&0x80 != 0 {
		a.setOverflow()
	}

	a.setZeroIfNeeded(a.cpu.a)
	a.setNegativeIfOverflow(a.cpu.a)
	return nil
}

// X -> M                           N Z C I D V
//                                  - - - - - -
func execSTX(a *Appleone, o op) error {
	addr, err := a.getAddr(o)
	if err != nil {
		return err
	}
	a.mem[addr] = a.cpu.x
	return nil
}

// Y -> M                           N Z C I D V
//                                  - - - - - -
func execSTY(a *Appleone, o op) error {
	addr, err := a.getAddr(o)
	if err != nil {
		return err
	}
	a.mem[addr] = a.cpu.y
	return nil
}

// A -> M                           N Z C I D V
//                                  - - - - - -
func execSTA(a *Appleone, o op) error {
	addr, err := a.getAddr(o)
	if err != nil {
		return err
	}
	a.mem[addr] = a.cpu.a
	return nil
}

// branch on Z = 1                  N Z C I D V
//                                  - - - - - -
func execBEQ(a *Appleone, o op) error {
	if a.getZero() == flagZero {
		a.branch(o)
	}
	return nil
}

// branch on Z = 0                  N Z C I D V
//                                  - - - - - -
func execBNE(a *Appleone, o op) error {
	if a.getZero() != flagZero {
		a.branch(o)
	}
	return nil
}

// branch on V = 0                  N Z C I D V
//                                  - - - - - -
func execBVC(a *Appleone, o op) error {
	if a.getOverflow() == 0 {
		a.branch(o)
	}
	return nil
}

// branch on V = 1                  N Z C I D V
//                                  - - - - - -
func execBVS(a *Appleone, o op) error {
	if a.getOverflow() != 0 {
		a.branch(o)
	}
	return nil
}

// bits 7 and 6 of operand are transfered to bit 7 and 6 of SR (N,V);
// the zeroflag is set to the result of operand AND accumulator.
func execBIT(a *Appleone, o op) error {
	operand, err := a.getOperand(o)
	if err != nil {
		return err
	}
	a.setZeroIfNeeded(a.cpu.a & operand)
	a.clearOverflow()

	if operand&flagOverflow != 0 {
		a.setOverflow()
	}

	a.setNegativeIfOverflow(operand)
	return nil
}

// branch on C = 0                  N Z C I D V
//                                  - - - - - -
func execBCC(a *Appleone, o op) error {
	if a.getCarry() == 0 {
		a.branch(o)
	}
	return nil
}

// branch on N = 1                  N Z C I D V
//                                  - - - - - -
func execBMI(a *Appleone, o op) error {
	if a.getNegative() == flagNegative {
		a.branch(o)
	}
	return nil
}

// branch on N = 0                  N Z C I D V
//                                  - - - - - -
func execBPL(a *Appleone, o op) error {
	if a.getNegative() == 0 {
		a.branch(o)
	}
	return nil
}

// branch on C = 1                  N Z C I D V
//                                  - - - - - -
func execBCS(a *Appleone, o op) error {
	if a.getCarry() != 0 {
		a.branch(o)
	}
	return nil
}

// X - M                            N Z C I D V
//                                  + + + - - -
func execCPX(a *Appleone, o op) error {
	operand, err := a.getOperand(o)
	if err != nil {
		return err
	}
	a.compare(a.cpu.x, operand)
	return nil
}

// A EOR M -> A                     N Z C I D V
//                                  + + - - - -
func execEOR(a *Appleone, o op) error {
	operand, err := a.getOperand(o)
	if err != nil {
		return err
	}
	a.cpu.a ^= operand
	a.setNegativeIfOverflow(a.cpu.a)
	a.setZeroIfNeeded(a.cpu.a)
	return nil
}

// A - M                            N Z C I D V
//                                  + + + - - -
func execCMP(a *Appleone, o op) error {
	operand, err := a.getOperand(o)
	if err != nil {
		return err
	}
	a.compare(a.cpu.a, operand)
	return nil
}

// Y - M                            N Z C I D V
//                                  + + + - - -
func execCPY(a *Appleone, o op) error {
	operand, err := a.getOperand(o)
	if err != nil {
		return err
	}
	a.compare(a.cpu.y, operand)
	return nil
}

// 0 -> C                           N Z C I D V
//                                  - - 0 - - -
func execCLC(a *Appleone, o op) error {
	a.clearCarry()
	return nil
}

// 0 -> D                           N Z C I D V
//                                  - - - - 0 -
func execCLD(a *Appleone, o op) error {
	a.clearDec()
	return nil
}

// 0 -> I                           N Z C I D V
//                                  - - - 0 - -
func execCLI(a *Appleone, o op) error {
	a.clearInterrupt()
	return nil
}

// 0 -> V                           N Z C I D V
//                                  - - - - - 0
func execCLV(a *Appleone, o op) error {
	a.clearOverflow()
	return nil
}

// 1 -> C                           N Z C I D V
//                                  - - 1 - - -
func execSEC(a *Appleone, o op) error {
	a.setCarry()
	return nil
}

// 1 -> D                           N Z C I D V
//                                  - - - - 1 -
func execSED(a *Appleone, o op) error {
	a.setDec()
	return nil
}

// 1 -> I                           N Z C I D V
//                                  - - - 1 - -
func execSEI(a *Appleone, o op) error {
	a.setInterrupt()
	return nil
}

// ---                              N Z C I D V
//                                  - - - - - -
func execNOP(a *Appleone, o op) error {
	return nil
}

// (PC+1) -> PCL                    N Z C I D V
// (PC+2) -> PCH                    - - - - - -
func execJMP(a *Appleone, o op) error {
	if o.addrMode == indirect {
		addr, err := a.getAddr(o)
		if err != nil {
			return err
		}
		a.cpu.pc = a.littleEndianToUint16(a.mem[addr+1], a.mem[addr])
		return nil
	}
	addr, err := a.getAddr(o)
	if err != nil {
		return err
	}
	a.cpu.pc = addr
	return nil
}

// push A                           N Z C I D V
//                                  - - - - - -
func execPHA(a *Appleone, o op) error {
	a.pushWordToStack(a.cpu.a)
	return nil
}

// X -> A                           N Z C I D V
//                                  + + - - - -
func execTXA(a *Appleone, o op) error {
	a.cpu.a = a.cpu.x
	a.setNegativeIfOverflow(a.cpu.a)
	a.setZeroIfNeeded(a.cpu.a)
	return nil
}

// Y -> A                           N Z C I D V
//                                  + + - - - -
func execTYA(a *Appleone, o op) error {
	a.cpu.a = a.cpu.y
	a.setNegativeIfOverflow(a.cpu.a)
	a.setZeroIfNeeded(a.cpu.a)
	return nil
}

// SP -> X                          N Z C I D V
//                                  + + - - - -
func execTSX(a *Appleone, o op) error {
	a.cpu.x = a.cpu.sp
	a.setNegativeIfOverflow(a.cpu.x)
	a.setZeroIfNeeded(a.cpu.x)
	return nil
}

// pull A                           N Z C I D V
//                                  + + - - - -
func execPLA(a *Appleone, o op) error {
	a.cpu.a = a.popStackWord()
	a.setNegativeIfOverflow(a.cpu.a)
	a.setZeroIfNeeded(a.cpu.a)
	return nil
}

// pull SR from stack                  N Z C I D V
func execPLP(a *Appleone, o op) error {
	a.cpu.ps = a.popStackWord() | 0B_00110000
	a.setNegativeIfOverflow(a.cpu.a)
	a.setZeroIfNeeded(a.cpu.a)
	return nil
}

// push SR                          N Z C I D V
//                                  - - - - - -
func execPHP(a *Appleone, o op) error {
	a.pushWordToStack(a.cpu.ps)
	return nil
}

// push (PC+2),                     N Z C I D V
// (PC+1) -> PCL                    - - - - - -
// (PC+2) -> PCH
func execJSR(a *Appleone, o op) error {
	a.pushDWordToStack(a.cpu.pc - 1)
	addr, err := a.getAddr(o)
	if err != nil {
		return err
	}
	a.cpu.pc = addr
	return nil
}

// pull PC, PC+1 -> PC              N Z C I D V
//                                  - - - - - -
func execRTS(a *Appleone, o op) error {
	a.cpu.pc = a.popStackDWord() + 1
	return nil
}

// 0 -> [76543210] -> C             N Z C I D V
//                                  0 + + - - -
func execLSR(a *Appleone, o op) error {
	operand, err := a.getOperand(o)
	if err != nil {
		return err
	}

	bit := (operand << 7) > 0
	operand >>= 1

	if bit {
		a.setCarry()
	} else {
		a.clearCarry()
	}

	a.setZeroIfNeeded(operand)
	a.setNegativeIfOverflow(operand)

	if o.addrMode == accumulator {
		a.cpu.a = operand
		return nil
	}

	addr, err := a.getAddr(o)
	if err != nil {
		return err
	}
	a.mem[addr] = operand

	return nil
}

// C <- [76543210] <- C             N Z C I D V
//                                  + + + - - -
func execROL(a *Appleone, o op) error {
	operand, err := a.getOperand(o)
	if err != nil {
		return err
	}

	var carry bool
	if a.getCarry() == flagCarry {
		carry = true
	}

	if operand>>7 != 0 {
		a.setCarry()
	} else {
		a.clearCarry()
	}

	operand <<= 1

	if carry {
		operand |= flagCarry
	}

	a.setZeroIfNeeded(operand)
	a.setNegativeIfOverflow(operand)

	if o.addrMode == accumulator {
		a.cpu.a = operand
		return nil
	}

	addr, err := a.getAddr(o)
	if err != nil {
		return err
	}
	a.mem[addr] = operand

	return nil
}

// X -> SP                          N Z C I D V
//                                  - - - - - -
func execTXS(a *Appleone, o op) error {
	a.cpu.sp = a.cpu.x
	// TODO: needed?
	// a.setZeroIfNeeded(a.cpu.sp)
	// a.setNegativeIfOverflow(a.cpu.sp)
	return nil
}

// C -> [76543210] -> C             N Z C I D V
//                                  + + + - - -
func execROR(a *Appleone, o op) error {
	operand, err := a.getOperand(o)
	if err != nil {
		return err
	}

	var carry bool
	if a.getCarry() == flagCarry {
		carry = true
	}

	if operand&0x01 != 0 {
		a.setCarry()
	} else {
		a.clearCarry()
	}

	operand >>= 1

	if carry {
		operand |= flagNegative
	}

	a.setZeroIfNeeded(operand)
	a.setNegativeIfOverflow(operand)

	if o.addrMode == accumulator {
		a.cpu.a = operand
		return nil
	}

	addr, err := a.getAddr(o)
	if err != nil {
		return err
	}
	a.mem[addr] = operand

	return nil
}

// C <- [76543210] <- 0             N Z C I D V
//                                  + + + - - -
func execASL(a *Appleone, o op) error {
	operand, err := a.getOperand(o)
	if err != nil {
		return err
	}

	if operand>>7 == 1 {
		a.setCarry()
	} else {
		a.clearCarry()
	}

	operand <<= 1

	a.setZeroIfNeeded(operand)
	a.setNegativeIfOverflow(operand)

	if o.addrMode == accumulator {
		a.cpu.a = operand
		return nil
	}

	addr, err := a.getAddr(o)
	if err != nil {
		return err
	}
	a.mem[addr] = operand

	return nil
}

// A AND M -> A                     N Z C I D V
//                                  + + - - - -
func execAND(a *Appleone, o op) error {
	operand, err := a.getOperand(o)
	if err != nil {
		return err
	}
	a.cpu.a &= operand
	a.setZeroIfNeeded(a.cpu.a)
	a.setNegativeIfOverflow(a.cpu.a)
	return nil
}

// A OR M -> A                      N Z C I D V
//                                  + + - - - -
func execORA(a *Appleone, o op) error {
	operand, err := a.getOperand(o)
	if err != nil {
		return err
	}
	a.cpu.a |= operand
	a.setZeroIfNeeded(a.cpu.a)
	a.setNegativeIfOverflow(a.cpu.a)
	return nil
}
