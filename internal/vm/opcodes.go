package vm

import (
	"errors"
)

// operation includes the name of the operation, it's 8 bit hexidecimal opcode, how
// many bytes it occupies (it's size), as well as it's addressing mode.
type operation struct {
	name     string
	opcode   byte
	size     byte
	addrMode addrMode
	exec     func(a *VM, op operation) error
}

func newOp(name string, opcode, size byte, addrMode addrMode, exec func(a *VM, op operation) error) operation {
	return operation{
		name:     name,
		opcode:   opcode,
		size:     size,
		addrMode: addrMode,
		exec:     exec,
	}
}

// operationByCode takes an opcode (a single byte/word) and returns the associated operation
func operationByCode(b byte) (operation, error) {
	o, ok := opcodes[b]
	if !ok {
		return operation{}, errors.New("unknown opcode")
	}
	return o, nil
}

// opcodes represent all of the Apple 1 opcodes available. Each 8 bit opcode is mapped to a corresponding
// "op" which is just a struct holding metadata about the operation.
var opcodes = map[byte]operation{
	// BRK Force Break
	// addressing    assembler    opc  bytes  cyles
	// --------------------------------------------
	// implied       BRK          00   1      7
	0x00: newOp("BRK", 0x00, 1, implied, execBRK),

	// RTI Return from Interrupt
	// addressing    assembler    opc  bytes  cyles
	// --------------------------------------------
	// implied       RTI          40   1      6
	0x40: newOp("RTI", 0x40, 1, implied, execRTI),

	// DEC Decrement Memory by One
	// addressing    assembler    opc  bytes  cyles
	// --------------------------------------------
	// zeropage      DEC oper     C6   2      5
	// zeropage,X    DEC oper,X   D6   2      6
	// absolute      DEC oper     CE   3      6
	// absolute,X    DEC oper,X   DE   3      7
	0xC6: newOp("DEC", 0xC6, 2, zeroPage, execDEC),
	0xD6: newOp("DEC", 0xD6, 2, zeroPageXIndexed, execDEC),
	0xCE: newOp("DEC", 0xCE, 3, absolute, execDEC),
	0xDE: newOp("DEC", 0xDE, 3, absoluteXIndexed, execDEC),

	// INC Increment Memory by One
	// addressing    assembler    opc  bytes  cyles
	// --------------------------------------------
	// zeropage      INC oper     E6   2      5
	// zeropage,X    INC oper,X   F6   2      6
	// absolute      INC oper     EE   3      6
	// absolute,X    INC oper,X   FE   3      7
	0xE6: newOp("INC", 0xE6, 2, zeroPage, execINC),
	0xF6: newOp("INC", 0xF6, 2, zeroPageXIndexed, execINC),
	0xEE: newOp("INC", 0xEE, 3, absolute, execINC),
	0xFE: newOp("INC", 0xFE, 3, absoluteXIndexed, execINC),

	// INX Increment Index X by One
	// addressing    assembler    opc  bytes  cyles
	// --------------------------------------------
	// implied       INX          E8   1      2
	0xE8: newOp("INX", 0xE8, 1, implied, execINX),

	// INY  Increment Index Y by One
	// addressing    assembler    opc  bytes  cyles
	// --------------------------------------------
	// implied       INY          C8   1      2
	0xC8: newOp("INY", 0xC8, 1, implied, execINY),

	// TAX Transfer Accumulator to Index X
	// addressing    assembler    opc  bytes  cyles
	// --------------------------------------------
	// implied       TAX          AA   1      2
	0xAA: newOp("TAX", 0xAA, 1, implied, execTAX),

	// TAY Transfer Accumulator to Index Y
	// addressing    assembler    opc  bytes  cyles
	// --------------------------------------------
	// implied       TAY          A8   1      2
	0xA8: newOp("TAY", 0xA8, 1, implied, execTAY),

	// DEX Decrement Index X by One
	// addressing    assembler    opc  bytes  cyles
	// --------------------------------------------
	// implied       DEC          CA   1      2
	0xCA: newOp("DEX", 0xCA, 1, implied, execDEX),

	// DEY Decrement Index Y by One
	// addressing    assembler    opc  bytes  cyles
	// --------------------------------------------
	// implied       DEC          88   1      2
	0x88: newOp("DEY", 0x88, 1, implied, execDEY),

	// LDA Load Accumulator with Memory
	// addressing    assembler    opc  bytes  cyles
	// --------------------------------------------
	// immidiate     LDA #oper    A9   2      2
	// zeropage      LDA oper     A5   2      3
	// zeropage,X    LDA oper,X   B5   2      4
	// absolute      LDA oper     AD   3      4
	// absolute,X    LDA oper,X   BD   3      4*
	// absolute,Y    LDA oper,Y   B9   3      4*
	// (indirect,X)  LDA (oper,X) A1   2      6
	// (indirect),Y  LDA (oper),Y B1   2      5*
	0xA9: newOp("LDA", 0xA9, 2, immediate, execLDA),
	0xA5: newOp("LDA", 0xA5, 2, zeroPage, execLDA),
	0xB5: newOp("LDA", 0xB5, 2, zeroPageXIndexed, execLDA),
	0xAD: newOp("LDA", 0xAD, 3, absolute, execLDA),
	0xBD: newOp("LDA", 0xBD, 3, absoluteXIndexed, execLDA),
	0xB9: newOp("LDA", 0xB9, 3, absoluteYIndexed, execLDA),
	0xA1: newOp("LDA", 0xA1, 2, indirectXIndexed, execLDA),
	0xB1: newOp("LDA", 0xB1, 2, indirectYIndexed, execLDA),

	// LDX Load Index X with Memory
	// addressing    assembler    opc  bytes  cyles
	// --------------------------------------------
	// immidiate     LDX #oper     A2    2     2
	// zeropage      LDX oper      A6    2     3
	// zeropage,Y    LDX oper,Y    B6    2     4
	// absolute      LDX oper      AE    3     4
	// absolute,Y    LDX oper,Y    BE    3     4*
	0xA2: newOp("LDX", 0xA2, 2, immediate, execLDX),
	0xA6: newOp("LDX", 0xA6, 2, zeroPage, execLDX),
	0xB6: newOp("LDX", 0xB6, 2, zeroPageYIndexed, execLDX),
	0xAE: newOp("LDX", 0xAE, 3, absolute, execLDX),
	0xBE: newOp("LDX", 0xBE, 3, absoluteYIndexed, execLDX),

	// LDY Load Index Y with Memory
	// addressing    assembler    opc  bytes  cyles
	// --------------------------------------------
	// immidiate     LDY #oper    A0   2      2
	// zeropage      LDY oper     A4   2      3
	// zeropage,X    LDY oper,X   B4   2      4
	// absolute      LDY oper     AC   3      4
	// absolute,X    LDY oper,X   BC   3      4*
	0xA0: newOp("LDY", 0xA0, 2, immediate, execLDY),
	0xA4: newOp("LDY", 0xA4, 2, zeroPage, execLDY),
	0xB4: newOp("LDY", 0xB4, 2, zeroPageXIndexed, execLDY),
	0xAC: newOp("LDY", 0xAC, 3, absolute, execLDY),
	0xBC: newOp("LDY", 0xBC, 3, absoluteXIndexed, execLDY),

	// ADC Add Memory to Accumulator with Carry
	// addressing    assembler     opc  bytes  cyles
	// --------------------------------------------
	// immidiate     ADC #oper     69   2      2
	// zeropage      ADC oper      65   2      3
	// zeropage,X    ADC oper,X    75   2      4
	// absolute      ADC oper      6D   3      4
	// absolute,X    ADC oper,X    7D   3      4*
	// absolute,Y    ADC oper,Y    79   3      4*
	// (indirect,X)  ADC (oper,X)  61   2      6
	// (indirect),Y  ADC (oper),Y  71   2      5*
	0x69: newOp("ADC", 0x69, 2, immediate, execADC),
	0x65: newOp("ADC", 0x65, 2, zeroPage, execADC),
	0x75: newOp("ADC", 0x75, 2, zeroPageXIndexed, execADC),
	0x6D: newOp("ADC", 0x6D, 3, absolute, execADC),
	0x7D: newOp("ADC", 0x7D, 3, absoluteXIndexed, execADC),
	0x79: newOp("ADC", 0x79, 3, absoluteYIndexed, execADC),
	0x61: newOp("ADC", 0x61, 2, indirectXIndexed, execADC),
	0x71: newOp("ADC", 0x71, 2, indirectYIndexed, execADC),

	// SBC Subtract Memory from Accumulator with Borrow
	// addressing    assembler     opc   bytes cyles
	// --------------------------------------------
	// immidiate     SBC #oper     E9    2     2
	// zeropage      SBC oper      E5    2     3
	// zeropage,X    SBC oper,X    F5    2     4
	// absolute      SBC oper      ED    3     4
	// absolute,X    SBC oper,X    FD    3     4*
	// absolute,Y    SBC oper,Y    F9    3     4*
	// (indirect,X)  SBC (oper,X)  E1    2     6
	// (indirect),Y  SBC (oper),Y  F1    2     5*
	0xE9: newOp("SBC", 0xE9, 2, immediate, execSBC),
	0xE5: newOp("SBC", 0xE5, 2, zeroPage, execSBC),
	0xF5: newOp("SBC", 0xF5, 2, zeroPageXIndexed, execSBC),
	0xED: newOp("SBC", 0xED, 3, absolute, execSBC),
	0xFD: newOp("SBC", 0xFD, 3, absoluteXIndexed, execSBC),
	0xF9: newOp("SBC", 0xF9, 3, absoluteYIndexed, execSBC),
	0xE1: newOp("SBC", 0xE1, 2, indirectXIndexed, execSBC),
	0xF1: newOp("SBC", 0xF1, 2, indirectYIndexed, execSBC),

	// STX Store Index X in Memory
	// addressing    assembler    opc  bytes  cyles
	// --------------------------------------------
	// zeropage      STX oper     86   2      3
	// zeropage,Y    STX oper,Y   96   2      4
	// absolute      STX oper     8E   3      4
	0x86: newOp("STX", 0x86, 2, zeroPage, execSTX),
	0x96: newOp("STX", 0x96, 2, zeroPageYIndexed, execSTX),
	0x8E: newOp("STX", 0x8E, 3, absolute, execSTX),

	// STY Store Index Y in Memory
	// addressing    assembler    opc  bytes  cyles
	// --------------------------------------------
	// zeropage      STY oper     84   2      3
	// zeropage,X    STY oper,X   94   2      4
	// absolute      STY oper     8C   3      4
	0x84: newOp("STY", 0x84, 2, zeroPage, execSTY),
	0x94: newOp("STY", 0x94, 2, zeroPageXIndexed, execSTY),
	0x8C: newOp("STY", 0x8C, 3, absolute, execSTY),

	// STA Store Accumulator in Memory
	// addressing    assembler     opc   bytes  cyles
	// --------------------------------------------
	// zeropage      STA oper      85    2      3
	// zeropage,X    STA oper,X    95    2      4
	// absolute      STA oper      8D    3      4
	// absolute,X    STA oper,X    9D    3      5
	// absolute,Y    STA oper,Y    99    3      5
	// (indirect,X)  STA (oper,X)  81    2      6
	// (indirect),Y  STA (oper),Y  91    2      6
	0x85: newOp("STA", 0x85, 2, zeroPage, execSTA),
	0x95: newOp("STA", 0x95, 2, zeroPageXIndexed, execSTA),
	0x8D: newOp("STA", 0x8D, 3, absolute, execSTA),
	0x9D: newOp("STA", 0x9D, 3, absoluteXIndexed, execSTA),
	0x99: newOp("STA", 0x99, 3, absoluteYIndexed, execSTA),
	0x81: newOp("STA", 0x81, 2, indirectXIndexed, execSTA),
	0x91: newOp("STA", 0x91, 2, indirectYIndexed, execSTA),

	// BEQ  Branch on Result Zero
	// addressing    assembler    opc  bytes  cyles
	// --------------------------------------------
	// relative      BEQ oper      F0    2     2**
	0xF0: newOp("BEQ", 0xF0, 2, relative, execBEQ),

	// BNE Branch on Result not Zero
	// addressing    assembler    opc  bytes  cyles
	// --------------------------------------------
	// relative      BNE oper     D0   2      2**
	0xD0: newOp("BNE", 0xD0, 2, relative, execBNE),

	// BVC Branch on Overflow Clear
	// addressing    assembler    opc  bytes  cyles
	// --------------------------------------------
	// relative      BVC oper      50    2     2**
	0x50: newOp("BVC", 0x50, 2, relative, execBVC),

	// BVS Branch on Overflow Set
	// addressing    assembler    opc  bytes  cyles
	// --------------------------------------------
	// relative      BVC oper      70    2     2**
	0x70: newOp("BVS", 0x70, 2, relative, execBVS),

	// BIT Test Bits in Memory with Accumulator
	// addressing    assembler    opc  bytes  cyles
	// --------------------------------------------
	// zeropage      BIT oper     24   2      3
	// absolute      BIT oper     2C   3      4
	0x24: newOp("BIT", 0x24, 2, zeroPage, execBIT),
	0x2C: newOp("BIT", 0x2C, 3, absolute, execBIT),

	// BCC Branch on Carry Clear
	// addressing    assembler    opc  bytes  cyles
	// --------------------------------------------
	// relative      BCC oper     90   2      2**
	0x90: newOp("BCC", 0x90, 2, relative, execBCC),

	// BMI Branch on Result Minus
	// addressing    assembler    opc  bytes  cyles
	// --------------------------------------------
	// relative      BMI oper     30   2      2**
	0x30: newOp("BMI", 0x30, 2, relative, execBMI),

	// BPL  Branch on Result Plus
	// addressing    assembler    opc  bytes  cyles
	// --------------------------------------------
	// relative      BPL oper     10   2      2**
	0x10: newOp("BPL", 0x10, 2, relative, execBPL),

	// BCS Branch on Carry Set
	// addressing    assembler    opc  bytes  cyles
	// --------------------------------------------
	// relative      BCS oper     B0   2      2**
	0xB0: newOp("BCS", 0xB0, 2, relative, execBCS),

	// CPX Compare Memory and Index X
	// addressing    assembler    opc  bytes  cyles
	// --------------------------------------------
	// immidiate     CPX #oper    E0   2      2
	// zeropage      CPX oper     E4   2      3
	// absolute      CPX oper     EC   3      4
	0xE0: newOp("CPX", 0xE0, 2, immediate, execCPX),
	0xE4: newOp("CPX", 0xE4, 2, zeroPage, execCPX),
	0xEC: newOp("CPX", 0xEC, 3, absolute, execCPX),

	// EOR  Exclusive-OR Memory with Accumulator
	// addressing    assembler     opc   bytes  cyles
	// --------------------------------------------
	// immidiate     EOR #oper     49    2      2
	// zeropage      EOR oper      45    2      3
	// zeropage,X    EOR oper,X    55    2      4
	// absolute      EOR oper      4D    3      4
	// absolute,X    EOR oper,X    5D    3      4*
	// absolute,Y    EOR oper,Y    59    3      4*
	// (indirect,X)  EOR (oper,X)  41    2      6
	// (indirect),Y  EOR (oper),Y  51    2      5*
	0x49: newOp("EOR", 0x49, 2, immediate, execEOR),
	0x45: newOp("EOR", 0x45, 2, zeroPage, execEOR),
	0x55: newOp("EOR", 0x55, 2, zeroPageXIndexed, execEOR),
	0x4D: newOp("EOR", 0x4D, 3, absolute, execEOR),
	0x5D: newOp("EOR", 0x5D, 3, absoluteXIndexed, execEOR),
	0x59: newOp("EOR", 0x59, 3, absoluteYIndexed, execEOR),
	0x41: newOp("EOR", 0x41, 2, indirectXIndexed, execEOR),
	0x51: newOp("EOR", 0x51, 2, indirectYIndexed, execEOR),

	// CMP Compare Memory with Accumulator
	// addressing    assembler     opc   bytes cyles
	// --------------------------------------------
	// immidiate     CMP #oper     C9    2     2
	// zeropage      CMP oper      C5    2     3
	// zeropage,X    CMP oper,X    D5    2     4
	// absolute      CMP oper      CD    3     4
	// absolute,X    CMP oper,X    DD    3     4*
	// absolute,Y    CMP oper,Y    D9    3     4*
	// (indirect,X)  CMP (oper,X)  C1    2     6
	// (indirect),Y  CMP (oper),Y  D1    2     5*
	0xC9: newOp("CMP", 0xC9, 2, immediate, execCMP),
	0xC5: newOp("CMP", 0xC5, 2, zeroPage, execCMP),
	0xD5: newOp("CMP", 0xD5, 2, zeroPageXIndexed, execCMP),
	0xCD: newOp("CMP", 0xCD, 3, absolute, execCMP),
	0xDD: newOp("CMP", 0xDD, 3, absoluteXIndexed, execCMP),
	0xD9: newOp("CMP", 0xD9, 3, absoluteYIndexed, execCMP),
	0xC1: newOp("CMP", 0xC1, 2, indirectXIndexed, execCMP),
	0xD1: newOp("CMP", 0xD1, 2, indirectYIndexed, execCMP),

	// CPY Compare Memory and Index Y
	// addressing    assembler    opc  bytes  cyles
	// --------------------------------------------
	// immidiate     CPY #oper    C0   2      2
	// zeropage      CPY oper     C4   2      3
	// absolute      CPY oper     CC   3      4
	0xC0: newOp("CPY", 0xC0, 2, immediate, execCPY),
	0xC4: newOp("CPY", 0xC4, 2, zeroPage, execCPY),
	0xCC: newOp("CPY", 0xCC, 3, absolute, execCPY),

	// CLC Clear Carry Flag
	// addressing    assembler    opc  bytes  cyles
	// --------------------------------------------
	// implied       CLC          18   1      2
	0x18: newOp("CLC", 0x18, 1, implied, execCLC),

	// CLD Clear Decimal Mode
	// addressing    assembler    opc  bytes  cyles
	// --------------------------------------------
	// implied       CLD          D8   1      2
	0xD8: newOp("CLD", 0xD8, 1, implied, execCLD),

	// CLI Clear Interrupt Disable Bit
	// addressing    assembler    opc  bytes  cyles
	// --------------------------------------------
	// implied       CLI          58   1      2
	0x58: newOp("CLI", 0x58, 1, implied, execCLI),

	// CLV Clear Overflow Flag
	// addressing    assembler    opc  bytes  cyles
	// --------------------------------------------
	// implied       CLV          B8   1      2
	0xB8: newOp("CLV", 0xB8, 1, implied, execCLV),

	// SEC Set Carry Flag
	// addressing    assembler    opc  bytes  cyles
	// --------------------------------------------
	// implied       SEC          38   1      2
	0x38: newOp("SEC", 0x38, 1, implied, execSEC),

	// SED Set Decimal Flag
	// addressing    assembler    opc  bytes  cyles
	// --------------------------------------------
	// implied       SED          F8   1      2
	0xF8: newOp("SED", 0xF8, 1, implied, execSED),

	// SEI Set Interrupt Disable Status
	// addressing    assembler    opc  bytes  cyles
	// --------------------------------------------
	// implied       SEI          78   1      2
	0x78: newOp("SEI", 0x78, 1, implied, execSEI),

	// NOP No Operation
	// addressing    assembler    opc  bytes  cyles
	// --------------------------------------------
	// implied       NOP          EA   1      2
	0xEA: newOp("NOP", 0xEA, 1, implied, execNOP),

	// JMP Jump to New Location
	// addressing    assembler    opc  bytes  cyles
	// --------------------------------------------
	// absolute      JMP oper     4C   3      3
	// indirect      JMP (oper)   6C   3      5
	0x4C: newOp("JMP", 0x4C, 3, absolute, execJMP),
	0x6C: newOp("JMP", 0x6C, 3, indirect, execJMP),

	// PHA Push Accumulator on Stack
	// addressing    assembler    opc  bytes  cyles
	// --------------------------------------------
	// implied       PHA           48    1     3
	0x48: newOp("PHA", 0x48, 1, implied, execPHA),

	// TXA Transfer Index X to Accumulator
	// addressing    assembler    opc  bytes  cyles
	// --------------------------------------------
	// implied       TXA          8A   1      2
	0x8A: newOp("TXA", 0x8A, 1, implied, execTXA),

	// TYA Transfer Index Y to Accumulator
	// addressing    assembler    opc  bytes  cyles
	// --------------------------------------------
	// implied       TYA          98   1      2
	0x98: newOp("TYA", 0x98, 1, implied, execTYA),

	// TSX Transfer Stack Pointer to Index X
	// addressing    assembler    opc  bytes  cyles
	// --------------------------------------------
	// implied       TSX          BA   1      2
	0xBA: newOp("TSX", 0xBA, 1, implied, execTSX),

	// PLA Pull Accumulator from Stack
	// addressing    assembler    opc  bytes  cyles
	// --------------------------------------------
	// implied       PLA          68   1      4
	0x68: newOp("PLA", 0x68, 1, implied, execPLA),

	// PLP Pull Processor Status from Stack
	// addressing    assembler    opc  bytes  cyles
	// --------------------------------------------
	// implied       PLP          28   1      4
	0x28: newOp("PLP", 0x28, 1, implied, execPLP),

	// PHP Push Processor Status on Stack
	// addressing    assembler    opc  bytes  cyles
	// --------------------------------------------
	// implied       PHP          08   1      3
	0x08: newOp("PHP", 0x08, 1, implied, execPHP),

	// JSR Jump to New Location Saving Return Address
	// addressing    assembler    opc  bytes  cyles
	// --------------------------------------------
	// absolute      JSR oper     20   3      6
	0x20: newOp("JSR", 0x20, 3, absolute, execJSR),

	// RTS Return from Subroutine
	// addressing    assembler    opc  bytes  cyles
	// --------------------------------------------
	// implied       RTS          60   1      6
	0x60: newOp("RTS", 0x60, 1, implied, execRTS),

	// LSR Shift One Bit Right (Memory or Accumulator)
	// addressing    assembler    opc  bytes  cyles
	// --------------------------------------------
	// accumulator   LSR A        4A   1      2
	// zeropage      LSR oper     46   2      5
	// zeropage,X    LSR oper,X   56   2      6
	// absolute      LSR oper     4E   3      6
	// absolute,X    LSR oper,X   5E   3      7
	0x4A: newOp("LSR", 0x4A, 1, accumulator, execLSR),
	0x46: newOp("LSR", 0x46, 2, zeroPage, execLSR),
	0x56: newOp("LSR", 0x56, 2, zeroPageXIndexed, execLSR),
	0x4E: newOp("LSR", 0x4E, 3, absolute, execLSR),
	0x5E: newOp("LSR", 0x5E, 3, absoluteXIndexed, execLSR),

	// ROL Rotate One Bit Left (Memory or Accumulator)
	// addressing    assembler    opc  bytes  cyles
	// --------------------------------------------
	// accumulator   ROL A        2A   1      2
	// zeropage      ROL oper     26   2      5
	// zeropage,X    ROL oper,X   36   2      6
	// absolute      ROL oper     2E   3      6
	// absolute,X    ROL oper,X   3E   3      7
	0x2A: newOp("ROL", 0x2A, 1, accumulator, execROL),
	0x26: newOp("ROL", 0x26, 2, zeroPage, execROL),
	0x36: newOp("ROL", 0x36, 2, zeroPageXIndexed, execROL),
	0x2E: newOp("ROL", 0x2E, 3, absolute, execROL),
	0x3E: newOp("ROL", 0x3E, 3, absoluteXIndexed, execROL),

	// TXS Transfer Index X to Stack Register
	// addressing    assembler    opc  bytes  cyles
	// --------------------------------------------
	// implied       TXS          9A   1      2
	0x9A: newOp("TXS", 0x9A, 1, implied, execTXS),

	// ROR Rotate One Bit Right (Memory or Accumulator)
	// addressing    assembler    opc  bytes  cyles
	// --------------------------------------------
	// accumulator   ROR A        6A   1      2
	// zeropage      ROR oper     66   2      5
	// zeropage,X    ROR oper,X   76   2      6
	// absolute      ROR oper     6E   3      6
	// absolute,X    ROR oper,X   7E   3      7
	0x6A: newOp("ROR", 0x6A, 1, accumulator, execROR),
	0x66: newOp("ROR", 0x66, 2, zeroPage, execROR),
	0x76: newOp("ROR", 0x76, 2, zeroPageXIndexed, execROR),
	0x6E: newOp("ROR", 0x6E, 3, absolute, execROR),
	0x7E: newOp("ROR", 0x7E, 3, absoluteXIndexed, execROR),

	// ASL Shift Left One Bit (Memory or Accumulator)
	// addressing    assembler    opc  bytes  cyles
	// --------------------------------------------
	// accumulator   ASL A        0A   1      2
	// zeropage      ASL oper     06   2      5
	// zeropage,X    ASL oper,X   16   2      6
	// absolute      ASL oper     0E   3      6
	// absolute,X    ASL oper,X   1E   3      7
	0x0A: newOp("ASL", 0x0A, 1, accumulator, execASL),
	0x06: newOp("ASL", 0x06, 2, zeroPage, execASL),
	0x16: newOp("ASL", 0x16, 2, zeroPageXIndexed, execASL),
	0x0E: newOp("ASL", 0x0E, 3, absolute, execASL),
	0x1E: newOp("ASL", 0x1E, 3, absoluteXIndexed, execASL),

	// AND AND Memory with Accumulator
	// addressing    assembler     opc   bytes  cyles
	// --------------------------------------------
	// immidiate     AND #oper     29    2      2
	// zeropage      AND oper      25    2      3
	// zeropage,X    AND oper,X    35    2      4
	// absolute      AND oper      2D    3      4
	// absolute,X    AND oper,X    3D    3      4*
	// absolute,Y    AND oper,Y    39    3      4*
	// (indirect,X)  AND (oper,X)  21    2      6
	// (indirect),Y  AND (oper),Y  31    2      5*
	0x29: newOp("AND", 0x29, 2, immediate, execAND),
	0x25: newOp("AND", 0x25, 2, zeroPage, execAND),
	0x35: newOp("AND", 0x35, 2, zeroPageXIndexed, execAND),
	0x2D: newOp("AND", 0x2D, 3, absolute, execAND),
	0x3D: newOp("AND", 0x3D, 3, absoluteXIndexed, execAND),
	0x39: newOp("AND", 0x39, 3, absoluteYIndexed, execAND),
	0x21: newOp("AND", 0x21, 2, indirectXIndexed, execAND),
	0x31: newOp("AND", 0x31, 2, indirectYIndexed, execAND),

	// ORA OR Memory with Accumulator
	// addressing    assembler     opc   bytes  cyles
	// --------------------------------------------
	// immidiate     ORA #oper     09    2      2
	// zeropage      ORA oper      05    2      3
	// zeropage,X    ORA oper,X    15    2      4
	// absolute      ORA oper      0D    3      4
	// absolute,X    ORA oper,X    1D    3      4*
	// absolute,Y    ORA oper,Y    19    3      4*
	// (indirect,X)  ORA (oper,X)  01    2      6
	// (indirect),Y  ORA (oper),Y  11    2      5*
	0x09: newOp("ORA", 0x09, 2, immediate, execORA),
	0x05: newOp("ORA", 0x05, 2, zeroPage, execORA),
	0x15: newOp("ORA", 0x15, 2, zeroPageXIndexed, execORA),
	0x0D: newOp("ORA", 0x0D, 3, absolute, execORA),
	0x1D: newOp("ORA", 0x1D, 3, absoluteXIndexed, execORA),
	0x19: newOp("ORA", 0x19, 3, absoluteYIndexed, execORA),
	0x01: newOp("ORA", 0x01, 2, indirectXIndexed, execORA),
	0x11: newOp("ORA", 0x11, 2, indirectYIndexed, execORA),
}
