package vm

// interrupt,                       N Z C I D V
// push PC+2, push SR               - - - 1 - -
func execBRK(vm *VM, o operation) error {
	// set processer status flag to BRK
	vm.cpu.ps = flagBreak

	vm.pushDWordToStack(vm.cpu.pc + 1)
	vm.pushWordToStack(vm.cpu.ps)

	vm.setFlag(flagDisableInterrupts)
	vm.cpu.pc = uint16(vm.mem[0xFFFF])<<8 | uint16(vm.mem[0xFFFE])

	return nil
}

// pull SR, pull PC                 N Z C I D V
// from stack
func execRTI(vm *VM, o operation) error {
	vm.cpu.ps = vm.popStackWord()
	vm.cpu.pc = vm.popStackDWord()
	return nil
}

// M - 1 -> M                       N Z C I D V
//                                  + + - - - -
func execDEC(vm *VM, o operation) error {
	addr, err := vm.getAddr(o)
	if err != nil {
		return err
	}
	b := vm.mem[addr]
	b--
	vm.mem[addr] = b
	vm.maybeSetFlagZero(b)
	vm.maybeSetFlagOverflow(b)
	return nil
}

// M + 1 -> M                       N Z C I D V
//                                  + + - - - -
func execINC(vm *VM, o operation) error {
	addr, err := vm.getAddr(o)
	if err != nil {
		return err
	}
	b := vm.mem[addr]
	b++
	vm.mem[addr] = b
	vm.maybeSetFlagZero(b)
	vm.maybeSetFlagOverflow(b)
	return nil
}

// X + 1 -> X                       N Z C I D V
//                                  + + - - - -
func execINX(vm *VM, o operation) error {
	vm.cpu.x++
	vm.maybeSetFlagZero(vm.cpu.x)
	vm.maybeSetFlagOverflow(vm.cpu.x)
	return nil
}

// Y + 1 -> Y                       N Z C I D V
//                                  + + - - - -
func execINY(vm *VM, o operation) error {
	vm.cpu.y++
	vm.maybeSetFlagZero(vm.cpu.y)
	vm.maybeSetFlagOverflow(vm.cpu.y)
	return nil
}

// A -> X                           N Z C I D V
//                                  + + - - - -
func execTAX(vm *VM, o operation) error {
	vm.cpu.x = vm.cpu.a
	vm.maybeSetFlagZero(vm.cpu.x)
	vm.maybeSetFlagOverflow(vm.cpu.x)
	return nil
}

// A -> Y                           N Z C I D V
//                                  + + - - - -
func execTAY(vm *VM, o operation) error {
	vm.cpu.y = vm.cpu.a
	vm.maybeSetFlagZero(vm.cpu.y)
	vm.maybeSetFlagOverflow(vm.cpu.y)
	return nil
}

// X - 1 -> X                       N Z C I D V
//                                  + + - - - -
func execDEX(vm *VM, o operation) error {
	vm.cpu.x--
	vm.maybeSetFlagZero(vm.cpu.x)
	vm.maybeSetFlagOverflow(vm.cpu.x)
	return nil
}

// Y - 1 -> Y                       N Z C I D V
//                                  + + - - - -
func execDEY(vm *VM, o operation) error {
	vm.cpu.y--
	vm.maybeSetFlagZero(vm.cpu.y)
	vm.maybeSetFlagOverflow(vm.cpu.y)
	return nil
}

// M -> A                           N Z C I D V
//                                  + + - - - -
func execLDA(vm *VM, o operation) error {
	operand, err := vm.getOperand(o)
	if err != nil {
		return err
	}
	vm.cpu.a = operand
	vm.maybeSetFlagZero(vm.cpu.a)
	vm.maybeSetFlagOverflow(vm.cpu.a)
	return nil
}

// M -> X                           N Z C I D V
//                                  + + - - - -
func execLDX(vm *VM, o operation) error {
	operand, err := vm.getOperand(o)
	if err != nil {
		return err
	}
	vm.cpu.x = operand
	vm.maybeSetFlagZero(vm.cpu.x)
	vm.maybeSetFlagOverflow(vm.cpu.x)
	return nil
}

// M -> Y                           N Z C I D V
//                                  + + - - - -
func execLDY(vm *VM, o operation) error {
	operand, err := vm.getOperand(o)
	if err != nil {
		return err
	}
	vm.cpu.y = operand
	vm.maybeSetFlagZero(vm.cpu.y)
	vm.maybeSetFlagOverflow(vm.cpu.y)
	return nil
}

// A + M + C -> A, C                N Z C I D V
//                                  + + + - - +
func execADC(vm *VM, o operation) error {
	b, err := vm.getOperand(o)
	if err != nil {
		return err
	}
	operand := uint16(b)
	regA := uint16(vm.cpu.a)
	sum := regA + operand + uint16(vm.getFlag(flagCarry))
	vm.cpu.a = byte(sum)

	vm.clearFlag(flagCarry)
	if sum > 255 {
		vm.setFlag(flagCarry)
	}

	// http://www.righto.com/2012/12/the-6502-overflow-flag-explained.html
	vm.clearFlag(flagOverflow)
	if (operand^sum)&(regA^sum)&0x80 != 0 {
		vm.setFlag(flagOverflow)
	}

	vm.maybeSetFlagZero(vm.cpu.a)
	vm.maybeSetFlagOverflow(vm.cpu.a)

	return nil
}

// A - M - C -> A                   N Z C I D V
//                                  + + + - - +
func execSBC(vm *VM, o operation) error {
	operand, err := vm.getOperand(o)
	if err != nil {
		return err
	}

	carry := uint16(1 - vm.getFlag(flagCarry))
	regA := vm.cpu.a
	sum := uint16(regA) - carry - uint16(operand)
	vm.cpu.a = byte(sum)
	vm.clearFlag(flagOverflow)

	if byte(regA)>>7 != vm.cpu.a>>7 {
		vm.setFlag(flagOverflow)
	}

	if uint16(sum) < 256 {
		vm.setFlag(flagCarry)
	} else {
		vm.clearFlag(flagCarry)
	}

	vm.clearFlag(flagOverflow)
	if ((255-operand)^vm.cpu.a)&(regA^vm.cpu.a)&0x80 != 0 {
		vm.setFlag(flagOverflow)
	}

	vm.maybeSetFlagZero(vm.cpu.a)
	vm.maybeSetFlagOverflow(vm.cpu.a)
	return nil
}

// X -> M                           N Z C I D V
//                                  - - - - - -
func execSTX(vm *VM, o operation) error {
	addr, err := vm.getAddr(o)
	if err != nil {
		return err
	}
	vm.mem[addr] = vm.cpu.x
	return nil
}

// Y -> M                           N Z C I D V
//                                  - - - - - -
func execSTY(vm *VM, o operation) error {
	addr, err := vm.getAddr(o)
	if err != nil {
		return err
	}
	vm.mem[addr] = vm.cpu.y
	return nil
}

// A -> M                           N Z C I D V
//                                  - - - - - -
func execSTA(vm *VM, o operation) error {
	addr, err := vm.getAddr(o)
	if err != nil {
		return err
	}
	vm.mem[addr] = vm.cpu.a
	return nil
}

// branch on Z = 1                  N Z C I D V
//                                  - - - - - -
func execBEQ(vm *VM, o operation) error {
	if vm.getFlag(flagZero) == flagZero {
		vm.branch(o)
	}
	return nil
}

// branch on Z = 0                  N Z C I D V
//                                  - - - - - -
func execBNE(vm *VM, o operation) error {
	if vm.getFlag(flagZero) != flagZero {
		vm.branch(o)
	}
	return nil
}

// branch on V = 0                  N Z C I D V
//                                  - - - - - -
func execBVC(vm *VM, o operation) error {
	if vm.getFlag(flagOverflow) == 0 {
		vm.branch(o)
	}
	return nil
}

// branch on V = 1                  N Z C I D V
//                                  - - - - - -
func execBVS(vm *VM, o operation) error {
	if vm.getFlag(flagOverflow) != 0 {
		vm.branch(o)
	}
	return nil
}

// bits 7 and 6 of operand are transfered to bit 7 and 6 of SR (N,V);
// the zeroflag is set to the result of operand AND accumulator.
func execBIT(vm *VM, o operation) error {
	operand, err := vm.getOperand(o)
	if err != nil {
		return err
	}
	vm.maybeSetFlagZero(vm.cpu.a & operand)
	vm.clearFlag(flagOverflow)

	if operand&flagOverflow != 0 {
		vm.setFlag(flagOverflow)
	}

	vm.maybeSetFlagOverflow(operand)
	return nil
}

// branch on C = 0                  N Z C I D V
//                                  - - - - - -
func execBCC(vm *VM, o operation) error {
	if vm.getFlag(flagCarry) == 0 {
		vm.branch(o)
	}
	return nil
}

// branch on N = 1                  N Z C I D V
//                                  - - - - - -
func execBMI(vm *VM, o operation) error {
	if vm.getFlag(flagNegative) == flagNegative {
		vm.branch(o)
	}
	return nil
}

// branch on N = 0                  N Z C I D V
//                                  - - - - - -
func execBPL(vm *VM, o operation) error {
	if vm.getFlag(flagNegative) == 0 {
		vm.branch(o)
	}
	return nil
}

// branch on C = 1                  N Z C I D V
//                                  - - - - - -
func execBCS(vm *VM, o operation) error {
	if vm.getFlag(flagCarry) != 0 {
		vm.branch(o)
	}
	return nil
}

// X - M                            N Z C I D V
//                                  + + + - - -
func execCPX(vm *VM, o operation) error {
	operand, err := vm.getOperand(o)
	if err != nil {
		return err
	}
	vm.compare(vm.cpu.x, operand)
	return nil
}

// A EOR M -> A                     N Z C I D V
//                                  + + - - - -
func execEOR(vm *VM, o operation) error {
	operand, err := vm.getOperand(o)
	if err != nil {
		return err
	}
	vm.cpu.a ^= operand
	vm.maybeSetFlagOverflow(vm.cpu.a)
	vm.maybeSetFlagZero(vm.cpu.a)
	return nil
}

// A - M                            N Z C I D V
//                                  + + + - - -
func execCMP(vm *VM, o operation) error {
	operand, err := vm.getOperand(o)
	if err != nil {
		return err
	}
	vm.compare(vm.cpu.a, operand)
	return nil
}

// Y - M                            N Z C I D V
//                                  + + + - - -
func execCPY(vm *VM, o operation) error {
	operand, err := vm.getOperand(o)
	if err != nil {
		return err
	}
	vm.compare(vm.cpu.y, operand)
	return nil
}

// 0 -> C                           N Z C I D V
//                                  - - 0 - - -
func execCLC(vm *VM, o operation) error {
	vm.clearFlag(flagCarry)
	return nil
}

// 0 -> D                           N Z C I D V
//                                  - - - - 0 -
func execCLD(vm *VM, o operation) error {
	vm.clearFlag(flagDecimalMode)
	return nil
}

// 0 -> I                           N Z C I D V
//                                  - - - 0 - -
func execCLI(vm *VM, o operation) error {
	vm.clearFlag(flagDisableInterrupts)
	return nil
}

// 0 -> V                           N Z C I D V
//                                  - - - - - 0
func execCLV(vm *VM, o operation) error {
	vm.clearFlag(flagOverflow)
	return nil
}

// 1 -> C                           N Z C I D V
//                                  - - 1 - - -
func execSEC(vm *VM, o operation) error {
	vm.setFlag(flagCarry)
	return nil
}

// 1 -> D                           N Z C I D V
//                                  - - - - 1 -
func execSED(vm *VM, o operation) error {
	vm.setFlag(flagDecimalMode)
	return nil
}

// 1 -> I                           N Z C I D V
//                                  - - - 1 - -
func execSEI(vm *VM, o operation) error {
	vm.setFlag(flagDisableInterrupts)
	return nil
}

// ---                              N Z C I D V
//                                  - - - - - -
func execNOP(vm *VM, o operation) error {
	return nil
}

// (PC+1) -> PCL                    N Z C I D V
// (PC+2) -> PCH                    - - - - - -
func execJMP(vm *VM, o operation) error {
	if o.addrMode == indirect {
		addr, err := vm.getAddr(o)
		if err != nil {
			return err
		}
		vm.cpu.pc = vm.littleEndianToUint16(vm.mem[addr+1], vm.mem[addr])
		return nil
	}
	addr, err := vm.getAddr(o)
	if err != nil {
		return err
	}
	vm.cpu.pc = addr
	return nil
}

// push A                           N Z C I D V
//                                  - - - - - -
func execPHA(vm *VM, o operation) error {
	vm.pushWordToStack(vm.cpu.a)
	return nil
}

// X -> A                           N Z C I D V
//                                  + + - - - -
func execTXA(vm *VM, o operation) error {
	vm.cpu.a = vm.cpu.x
	vm.maybeSetFlagOverflow(vm.cpu.a)
	vm.maybeSetFlagZero(vm.cpu.a)
	return nil
}

// Y -> A                           N Z C I D V
//                                  + + - - - -
func execTYA(vm *VM, o operation) error {
	vm.cpu.a = vm.cpu.y
	vm.maybeSetFlagOverflow(vm.cpu.a)
	vm.maybeSetFlagZero(vm.cpu.a)
	return nil
}

// SP -> X                          N Z C I D V
//                                  + + - - - -
func execTSX(vm *VM, o operation) error {
	vm.cpu.x = vm.cpu.sp
	vm.maybeSetFlagOverflow(vm.cpu.x)
	vm.maybeSetFlagZero(vm.cpu.x)
	return nil
}

// pull A                           N Z C I D V
//                                  + + - - - -
func execPLA(vm *VM, o operation) error {
	vm.cpu.a = vm.popStackWord()
	vm.maybeSetFlagOverflow(vm.cpu.a)
	vm.maybeSetFlagZero(vm.cpu.a)
	return nil
}

// pull SR from stack                  N Z C I D V
func execPLP(vm *VM, o operation) error {
	vm.cpu.ps = vm.popStackWord() | 0B_00110000
	vm.maybeSetFlagOverflow(vm.cpu.a)
	vm.maybeSetFlagZero(vm.cpu.a)
	return nil
}

// push SR                          N Z C I D V
//                                  - - - - - -
func execPHP(vm *VM, o operation) error {
	vm.pushWordToStack(vm.cpu.ps)
	return nil
}

// push (PC+2),                     N Z C I D V
// (PC+1) -> PCL                    - - - - - -
// (PC+2) -> PCH
func execJSR(vm *VM, o operation) error {
	vm.pushDWordToStack(vm.cpu.pc - 1)
	addr, err := vm.getAddr(o)
	if err != nil {
		return err
	}
	vm.cpu.pc = addr
	return nil
}

// pull PC, PC+1 -> PC              N Z C I D V
//                                  - - - - - -
func execRTS(vm *VM, o operation) error {
	vm.cpu.pc = vm.popStackDWord() + 1
	return nil
}

// 0 -> [76543210] -> C             N Z C I D V
//                                  0 + + - - -
func execLSR(vm *VM, o operation) error {
	operand, err := vm.getOperand(o)
	if err != nil {
		return err
	}

	bit := (operand << 7) > 0
	operand >>= 1

	if bit {
		vm.setFlag(flagCarry)
	} else {
		vm.clearFlag(flagCarry)
	}

	vm.maybeSetFlagZero(operand)
	vm.maybeSetFlagOverflow(operand)

	if o.addrMode == accumulator {
		vm.cpu.a = operand
		return nil
	}

	if err := vm.setMem(o, operand); err != nil {
		return err
	}

	return nil
}

// C <- [76543210] <- C             N Z C I D V
//                                  + + + - - -
func execROL(vm *VM, o operation) error {
	operand, err := vm.getOperand(o)
	if err != nil {
		return err
	}

	var carry bool
	if vm.getFlag(flagCarry) == flagCarry {
		carry = true
	}

	if operand>>7 != 0 {
		vm.setFlag(flagCarry)
	} else {
		vm.clearFlag(flagCarry)
	}

	operand <<= 1

	if carry {
		operand |= flagCarry
	}

	vm.maybeSetFlagZero(operand)
	vm.maybeSetFlagOverflow(operand)

	if o.addrMode == accumulator {
		vm.cpu.a = operand
		return nil
	}

	if err := vm.setMem(o, operand); err != nil {
		return err
	}

	return nil
}

// X -> SP                          N Z C I D V
//                                  - - - - - -
func execTXS(vm *VM, o operation) error {
	vm.cpu.sp = vm.cpu.x
	// TODO: needed?
	// vm.maybeSetFlagZero(vm.cpu.sp)
	// vm.maybeSetFlagOverflow(vm.cpu.sp)
	return nil
}

// C -> [76543210] -> C             N Z C I D V
//                                  + + + - - -
func execROR(vm *VM, o operation) error {
	operand, err := vm.getOperand(o)
	if err != nil {
		return err
	}

	var carry bool
	if vm.getFlag(flagCarry) == flagCarry {
		carry = true
	}

	if operand&0x01 != 0 {
		vm.setFlag(flagCarry)
	} else {
		vm.clearFlag(flagCarry)
	}

	operand >>= 1

	if carry {
		operand |= flagNegative
	}

	vm.maybeSetFlagZero(operand)
	vm.maybeSetFlagOverflow(operand)

	if o.addrMode == accumulator {
		vm.cpu.a = operand
		return nil
	}

	if err := vm.setMem(o, operand); err != nil {
		return err
	}

	return nil
}

// C <- [76543210] <- 0             N Z C I D V
//                                  + + + - - -
func execASL(vm *VM, o operation) error {
	operand, err := vm.getOperand(o)
	if err != nil {
		return err
	}

	if operand>>7 == 1 {
		vm.setFlag(flagCarry)
	} else {
		vm.clearFlag(flagCarry)
	}

	operand <<= 1

	vm.maybeSetFlagZero(operand)
	vm.maybeSetFlagOverflow(operand)

	if o.addrMode == accumulator {
		vm.cpu.a = operand
		return nil
	}

	if err := vm.setMem(o, operand); err != nil {
		return err
	}

	return nil
}

// A AND M -> A                     N Z C I D V
//                                  + + - - - -
func execAND(vm *VM, o operation) error {
	operand, err := vm.getOperand(o)
	if err != nil {
		return err
	}
	vm.cpu.a &= operand
	vm.maybeSetFlagZero(vm.cpu.a)
	vm.maybeSetFlagOverflow(vm.cpu.a)
	return nil
}

// A OR M -> A                      N Z C I D V
//                                  + + - - - -
func execORA(vm *VM, o operation) error {
	operand, err := vm.getOperand(o)
	if err != nil {
		return err
	}
	vm.cpu.a |= operand
	vm.maybeSetFlagZero(vm.cpu.a)
	vm.maybeSetFlagOverflow(vm.cpu.a)
	return nil
}
