package vm

// interrupt,                       N Z C I D V
// push PC+2, push SR               - - - 1 - -
func execBRK(a *Appleone, o operation) error {
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
func execRTI(a *Appleone, o operation) error {
	a.cpu.ps = a.popStackWord()
	a.cpu.pc = a.popStackDWord()
	return nil
}

// M - 1 -> M                       N Z C I D V
//                                  + + - - - -
func execDEC(a *Appleone, o operation) error {
	addr, err := a.getAddr(o)
	if err != nil {
		return err
	}
	b := a.mem[addr]
	b--
	a.mem[addr] = b
	a.maybeSetFlagZero(b)
	a.maybeSetFlagOverflow(b)
	return nil
}

// M + 1 -> M                       N Z C I D V
//                                  + + - - - -
func execINC(a *Appleone, o operation) error {
	addr, err := a.getAddr(o)
	if err != nil {
		return err
	}
	b := a.mem[addr]
	b++
	a.mem[addr] = b
	a.maybeSetFlagZero(b)
	a.maybeSetFlagOverflow(b)
	return nil
}

// X + 1 -> X                       N Z C I D V
//                                  + + - - - -
func execINX(a *Appleone, o operation) error {
	a.cpu.x++
	a.maybeSetFlagZero(a.cpu.x)
	a.maybeSetFlagOverflow(a.cpu.x)
	return nil
}

// Y + 1 -> Y                       N Z C I D V
//                                  + + - - - -
func execINY(a *Appleone, o operation) error {
	a.cpu.y++
	a.maybeSetFlagZero(a.cpu.y)
	a.maybeSetFlagOverflow(a.cpu.y)
	return nil
}

// A -> X                           N Z C I D V
//                                  + + - - - -
func execTAX(a *Appleone, o operation) error {
	a.cpu.x = a.cpu.a
	a.maybeSetFlagZero(a.cpu.x)
	a.maybeSetFlagOverflow(a.cpu.x)
	return nil
}

// A -> Y                           N Z C I D V
//                                  + + - - - -
func execTAY(a *Appleone, o operation) error {
	a.cpu.y = a.cpu.a
	a.maybeSetFlagZero(a.cpu.y)
	a.maybeSetFlagOverflow(a.cpu.y)
	return nil
}

// X - 1 -> X                       N Z C I D V
//                                  + + - - - -
func execDEX(a *Appleone, o operation) error {
	a.cpu.x--
	a.maybeSetFlagZero(a.cpu.x)
	a.maybeSetFlagOverflow(a.cpu.x)
	return nil
}

// Y - 1 -> Y                       N Z C I D V
//                                  + + - - - -
func execDEY(a *Appleone, o operation) error {
	a.cpu.y--
	a.maybeSetFlagZero(a.cpu.y)
	a.maybeSetFlagOverflow(a.cpu.y)
	return nil
}

// M -> A                           N Z C I D V
//                                  + + - - - -
func execLDA(a *Appleone, o operation) error {
	operand, err := a.getOperand(o)
	if err != nil {
		return err
	}
	a.cpu.a = operand
	a.maybeSetFlagZero(a.cpu.a)
	a.maybeSetFlagOverflow(a.cpu.a)
	return nil
}

// M -> X                           N Z C I D V
//                                  + + - - - -
func execLDX(a *Appleone, o operation) error {
	operand, err := a.getOperand(o)
	if err != nil {
		return err
	}
	a.cpu.x = operand
	a.maybeSetFlagZero(a.cpu.x)
	a.maybeSetFlagOverflow(a.cpu.x)
	return nil
}

// M -> Y                           N Z C I D V
//                                  + + - - - -
func execLDY(a *Appleone, o operation) error {
	operand, err := a.getOperand(o)
	if err != nil {
		return err
	}
	a.cpu.y = operand
	a.maybeSetFlagZero(a.cpu.y)
	a.maybeSetFlagOverflow(a.cpu.y)
	return nil
}

// A + M + C -> A, C                N Z C I D V
//                                  + + + - - +
func execADC(a *Appleone, o operation) error {
	b, err := a.getOperand(o)
	if err != nil {
		return err
	}
	operand := uint16(b)
	regA := uint16(a.cpu.a)
	sum := regA + operand + uint16(a.getFlag(flagCarry))
	a.cpu.a = byte(sum)

	a.clearFlag(flagCarry)
	if sum > 255 {
		a.setFlag(flagCarry)
	}

	// http://www.righto.com/2012/12/the-6502-overflow-flag-explained.html
	a.clearFlag(flagOverflow)
	if (operand^sum)&(regA^sum)&0x80 != 0 {
		a.setFlag(flagOverflow)
	}

	a.maybeSetFlagZero(a.cpu.a)
	a.maybeSetFlagOverflow(a.cpu.a)

	return nil
}

// A - M - C -> A                   N Z C I D V
//                                  + + + - - +
func execSBC(a *Appleone, o operation) error {
	operand, err := a.getOperand(o)
	if err != nil {
		return err
	}

	carry := uint16(1 - a.getFlag(flagCarry))
	regA := a.cpu.a
	sum := uint16(regA) - carry - uint16(operand)
	a.cpu.a = byte(sum)
	a.clearFlag(flagOverflow)

	if byte(regA)>>7 != a.cpu.a>>7 {
		a.setFlag(flagOverflow)
	}

	if uint16(sum) < 256 {
		a.setFlag(flagCarry)
	} else {
		a.clearFlag(flagCarry)
	}

	a.clearFlag(flagOverflow)
	if ((255-operand)^a.cpu.a)&(regA^a.cpu.a)&0x80 != 0 {
		a.setFlag(flagOverflow)
	}

	a.maybeSetFlagZero(a.cpu.a)
	a.maybeSetFlagOverflow(a.cpu.a)
	return nil
}

// X -> M                           N Z C I D V
//                                  - - - - - -
func execSTX(a *Appleone, o operation) error {
	addr, err := a.getAddr(o)
	if err != nil {
		return err
	}
	a.mem[addr] = a.cpu.x
	return nil
}

// Y -> M                           N Z C I D V
//                                  - - - - - -
func execSTY(a *Appleone, o operation) error {
	addr, err := a.getAddr(o)
	if err != nil {
		return err
	}
	a.mem[addr] = a.cpu.y
	return nil
}

// A -> M                           N Z C I D V
//                                  - - - - - -
func execSTA(a *Appleone, o operation) error {
	addr, err := a.getAddr(o)
	if err != nil {
		return err
	}
	a.mem[addr] = a.cpu.a
	return nil
}

// branch on Z = 1                  N Z C I D V
//                                  - - - - - -
func execBEQ(a *Appleone, o operation) error {
	if a.getFlag(flagZero) == flagZero {
		a.branch(o)
	}
	return nil
}

// branch on Z = 0                  N Z C I D V
//                                  - - - - - -
func execBNE(a *Appleone, o operation) error {
	if a.getFlag(flagZero) != flagZero {
		a.branch(o)
	}
	return nil
}

// branch on V = 0                  N Z C I D V
//                                  - - - - - -
func execBVC(a *Appleone, o operation) error {
	if a.getFlag(flagOverflow) == 0 {
		a.branch(o)
	}
	return nil
}

// branch on V = 1                  N Z C I D V
//                                  - - - - - -
func execBVS(a *Appleone, o operation) error {
	if a.getFlag(flagOverflow) != 0 {
		a.branch(o)
	}
	return nil
}

// bits 7 and 6 of operand are transfered to bit 7 and 6 of SR (N,V);
// the zeroflag is set to the result of operand AND accumulator.
func execBIT(a *Appleone, o operation) error {
	operand, err := a.getOperand(o)
	if err != nil {
		return err
	}
	a.maybeSetFlagZero(a.cpu.a & operand)
	a.clearFlag(flagOverflow)

	if operand&flagOverflow != 0 {
		a.setFlag(flagOverflow)
	}

	a.maybeSetFlagOverflow(operand)
	return nil
}

// branch on C = 0                  N Z C I D V
//                                  - - - - - -
func execBCC(a *Appleone, o operation) error {
	if a.getFlag(flagCarry) == 0 {
		a.branch(o)
	}
	return nil
}

// branch on N = 1                  N Z C I D V
//                                  - - - - - -
func execBMI(a *Appleone, o operation) error {
	if a.getFlag(flagNegative) == flagNegative {
		a.branch(o)
	}
	return nil
}

// branch on N = 0                  N Z C I D V
//                                  - - - - - -
func execBPL(a *Appleone, o operation) error {
	if a.getFlag(flagNegative) == 0 {
		a.branch(o)
	}
	return nil
}

// branch on C = 1                  N Z C I D V
//                                  - - - - - -
func execBCS(a *Appleone, o operation) error {
	if a.getFlag(flagCarry) != 0 {
		a.branch(o)
	}
	return nil
}

// X - M                            N Z C I D V
//                                  + + + - - -
func execCPX(a *Appleone, o operation) error {
	operand, err := a.getOperand(o)
	if err != nil {
		return err
	}
	a.compare(a.cpu.x, operand)
	return nil
}

// A EOR M -> A                     N Z C I D V
//                                  + + - - - -
func execEOR(a *Appleone, o operation) error {
	operand, err := a.getOperand(o)
	if err != nil {
		return err
	}
	a.cpu.a ^= operand
	a.maybeSetFlagOverflow(a.cpu.a)
	a.maybeSetFlagZero(a.cpu.a)
	return nil
}

// A - M                            N Z C I D V
//                                  + + + - - -
func execCMP(a *Appleone, o operation) error {
	operand, err := a.getOperand(o)
	if err != nil {
		return err
	}
	a.compare(a.cpu.a, operand)
	return nil
}

// Y - M                            N Z C I D V
//                                  + + + - - -
func execCPY(a *Appleone, o operation) error {
	operand, err := a.getOperand(o)
	if err != nil {
		return err
	}
	a.compare(a.cpu.y, operand)
	return nil
}

// 0 -> C                           N Z C I D V
//                                  - - 0 - - -
func execCLC(a *Appleone, o operation) error {
	a.clearFlag(flagCarry)
	return nil
}

// 0 -> D                           N Z C I D V
//                                  - - - - 0 -
func execCLD(a *Appleone, o operation) error {
	a.clearFlag(flagDecimalMode)
	return nil
}

// 0 -> I                           N Z C I D V
//                                  - - - 0 - -
func execCLI(a *Appleone, o operation) error {
	a.clearFlag(flagDisableInterrupts)
	return nil
}

// 0 -> V                           N Z C I D V
//                                  - - - - - 0
func execCLV(a *Appleone, o operation) error {
	a.clearFlag(flagOverflow)
	return nil
}

// 1 -> C                           N Z C I D V
//                                  - - 1 - - -
func execSEC(a *Appleone, o operation) error {
	a.setFlag(flagCarry)
	return nil
}

// 1 -> D                           N Z C I D V
//                                  - - - - 1 -
func execSED(a *Appleone, o operation) error {
	a.setFlag(flagDecimalMode)
	return nil
}

// 1 -> I                           N Z C I D V
//                                  - - - 1 - -
func execSEI(a *Appleone, o operation) error {
	a.setFlag(flagDisableInterrupts)
	return nil
}

// ---                              N Z C I D V
//                                  - - - - - -
func execNOP(a *Appleone, o operation) error {
	return nil
}

// (PC+1) -> PCL                    N Z C I D V
// (PC+2) -> PCH                    - - - - - -
func execJMP(a *Appleone, o operation) error {
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
func execPHA(a *Appleone, o operation) error {
	a.pushWordToStack(a.cpu.a)
	return nil
}

// X -> A                           N Z C I D V
//                                  + + - - - -
func execTXA(a *Appleone, o operation) error {
	a.cpu.a = a.cpu.x
	a.maybeSetFlagOverflow(a.cpu.a)
	a.maybeSetFlagZero(a.cpu.a)
	return nil
}

// Y -> A                           N Z C I D V
//                                  + + - - - -
func execTYA(a *Appleone, o operation) error {
	a.cpu.a = a.cpu.y
	a.maybeSetFlagOverflow(a.cpu.a)
	a.maybeSetFlagZero(a.cpu.a)
	return nil
}

// SP -> X                          N Z C I D V
//                                  + + - - - -
func execTSX(a *Appleone, o operation) error {
	a.cpu.x = a.cpu.sp
	a.maybeSetFlagOverflow(a.cpu.x)
	a.maybeSetFlagZero(a.cpu.x)
	return nil
}

// pull A                           N Z C I D V
//                                  + + - - - -
func execPLA(a *Appleone, o operation) error {
	a.cpu.a = a.popStackWord()
	a.maybeSetFlagOverflow(a.cpu.a)
	a.maybeSetFlagZero(a.cpu.a)
	return nil
}

// pull SR from stack                  N Z C I D V
func execPLP(a *Appleone, o operation) error {
	a.cpu.ps = a.popStackWord() | 0B_00110000
	a.maybeSetFlagOverflow(a.cpu.a)
	a.maybeSetFlagZero(a.cpu.a)
	return nil
}

// push SR                          N Z C I D V
//                                  - - - - - -
func execPHP(a *Appleone, o operation) error {
	a.pushWordToStack(a.cpu.ps)
	return nil
}

// push (PC+2),                     N Z C I D V
// (PC+1) -> PCL                    - - - - - -
// (PC+2) -> PCH
func execJSR(a *Appleone, o operation) error {
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
func execRTS(a *Appleone, o operation) error {
	a.cpu.pc = a.popStackDWord() + 1
	return nil
}

// 0 -> [76543210] -> C             N Z C I D V
//                                  0 + + - - -
func execLSR(a *Appleone, o operation) error {
	operand, err := a.getOperand(o)
	if err != nil {
		return err
	}

	bit := (operand << 7) > 0
	operand >>= 1

	if bit {
		a.setFlag(flagCarry)
	} else {
		a.clearFlag(flagCarry)
	}

	a.maybeSetFlagZero(operand)
	a.maybeSetFlagOverflow(operand)

	if o.addrMode == accumulator {
		a.cpu.a = operand
		return nil
	}

	if err := a.setMem(o, operand); err != nil {
		return err
	}

	return nil
}

// C <- [76543210] <- C             N Z C I D V
//                                  + + + - - -
func execROL(a *Appleone, o operation) error {
	operand, err := a.getOperand(o)
	if err != nil {
		return err
	}

	var carry bool
	if a.getFlag(flagCarry) == flagCarry {
		carry = true
	}

	if operand>>7 != 0 {
		a.setFlag(flagCarry)
	} else {
		a.clearFlag(flagCarry)
	}

	operand <<= 1

	if carry {
		operand |= flagCarry
	}

	a.maybeSetFlagZero(operand)
	a.maybeSetFlagOverflow(operand)

	if o.addrMode == accumulator {
		a.cpu.a = operand
		return nil
	}

	if err := a.setMem(o, operand); err != nil {
		return err
	}

	return nil
}

// X -> SP                          N Z C I D V
//                                  - - - - - -
func execTXS(a *Appleone, o operation) error {
	a.cpu.sp = a.cpu.x
	// TODO: needed?
	// a.maybeSetFlagZero(a.cpu.sp)
	// a.maybeSetFlagOverflow(a.cpu.sp)
	return nil
}

// C -> [76543210] -> C             N Z C I D V
//                                  + + + - - -
func execROR(a *Appleone, o operation) error {
	operand, err := a.getOperand(o)
	if err != nil {
		return err
	}

	var carry bool
	if a.getFlag(flagCarry) == flagCarry {
		carry = true
	}

	if operand&0x01 != 0 {
		a.setFlag(flagCarry)
	} else {
		a.clearFlag(flagCarry)
	}

	operand >>= 1

	if carry {
		operand |= flagNegative
	}

	a.maybeSetFlagZero(operand)
	a.maybeSetFlagOverflow(operand)

	if o.addrMode == accumulator {
		a.cpu.a = operand
		return nil
	}

	if err := a.setMem(o, operand); err != nil {
		return err
	}

	return nil
}

// C <- [76543210] <- 0             N Z C I D V
//                                  + + + - - -
func execASL(a *Appleone, o operation) error {
	operand, err := a.getOperand(o)
	if err != nil {
		return err
	}

	if operand>>7 == 1 {
		a.setFlag(flagCarry)
	} else {
		a.clearFlag(flagCarry)
	}

	operand <<= 1

	a.maybeSetFlagZero(operand)
	a.maybeSetFlagOverflow(operand)

	if o.addrMode == accumulator {
		a.cpu.a = operand
		return nil
	}

	if err := a.setMem(o, operand); err != nil {
		return err
	}

	return nil
}

// A AND M -> A                     N Z C I D V
//                                  + + - - - -
func execAND(a *Appleone, o operation) error {
	operand, err := a.getOperand(o)
	if err != nil {
		return err
	}
	a.cpu.a &= operand
	a.maybeSetFlagZero(a.cpu.a)
	a.maybeSetFlagOverflow(a.cpu.a)
	return nil
}

// A OR M -> A                      N Z C I D V
//                                  + + - - - -
func execORA(a *Appleone, o operation) error {
	operand, err := a.getOperand(o)
	if err != nil {
		return err
	}
	a.cpu.a |= operand
	a.maybeSetFlagZero(a.cpu.a)
	a.maybeSetFlagOverflow(a.cpu.a)
	return nil
}
