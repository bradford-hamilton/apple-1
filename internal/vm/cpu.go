// Package vm emulates a MOS Technology 6502 Microprocessor which is an 8-bit microprocessor
// that was designed by a small team led by Chuck Peddle for MOS Technology. The design team
// had formerly worked at Motorola on the Motorola 6800 project; the 6502 is essentially
// simplified, less expensive and faster version of that design.
//
// A		....	Accumulator             OPC A           operand is AC (implied single byte instruction)
// abs		....	absolute                OPC $LLHH       operand is address $HHLL *
// abs,X	....	absolute, X-indexed     OPC $LLHH,X     operand is address; effective address is address incremented by X with carry **
// abs,Y	....	absolute, Y-indexed     OPC $LLHH,Y     operand is address; effective address is address incremented by Y with carry **
// #		....	immediate               OPC #$BB        operand is byte BB
// impl		....	implied                 OPC             operand implied
// ind		....	indirect                OPC ($LLHH)     operand is address; effective address is contents of word at address: C.w($HHLL)
// X,ind	....	X-indexed, indirect     OPC ($LL,X)     operand is zeropage address; effective address is word in (LL + X, LL + X + 1), inc. without carry: C.w($00LL + X)
// ind,Y	....	indirect, Y-indexed     OPC ($LL),Y     operand is zeropage address; effective address is word in (LL, LL + 1) incremented by Y with carry: C.w($00LL) + Y
// rel		....	relative                OPC $BB         branch target is PC + signed offset BB ***
// zpg		....	zeropage                OPC $LL         operand is zeropage address (hi-byte is zero, address = $00LL)
// zpg,X	....	zeropage, X-indexed     OPC $LL,X       operand is zeropage address; effective address is address incremented by X without carry **
// zpg,Y	....	zeropage, Y-indexed     OPC $LL,Y       operand is zeropage address; effective address is address incremented by Y without carry **
//
// 16-bit address words are little endian, lo(w)-byte first, followed by the hi(gh)-byte.
//
// Processor Stack:
// LIFO, top down, 8 bit range, 0x0100 - 0x01FF
//
// 6502 instructions have the general form AAABBBCC, where AAA and CC define the opcode, and BBB defines the addressing mode
package vm

// addrMode is a type alias for a string, used below for defining addressing modes
type addrMode int

const (
	accumulator addrMode = iota
	absolute
	absoluteXIndexed
	absoluteYIndexed
	immediate
	implied
	indirect
	indirectXIndexed
	indirectYIndexed
	relative
	zeroPage
	zeroPageXIndexed
	zeroPageYIndexed
)

// Available cpu flags written as binary integer literals
// https://wiki.nesdev.com/w/index.php/Status_flags
// 7     bit     0
// - - - - - - - -
// N V s s D I Z C
// | | | | | | | |
// | | | | | | | +- Carry
// | | | | | | +--- Zero
// | | | | | + ---- Interrupt Disable
// | | | | + ------ Decimal
// | | + +--------- No CPU effect, see: the B flag
// | + ------------ Overflow
// +--------------- Negative
const (
	flagDefault           byte = 0B_00110000
	flagNegative          byte = 0B_10000000
	flagOverflow          byte = 0B_01000000
	flagBreak             byte = 0B_00010000
	flagDecimalMode       byte = 0B_00001000
	flagDisableInterrupts byte = 0B_00000100
	flagZero              byte = 0B_00000010
	flagCarry             byte = 0B_00000001
)

// StackBottom represents the start of the stack
const StackBottom uint16 = 0x0100 // 256

// Mos6502 represents the cpu's registers
type Mos6502 struct {
	sp byte   // register - stack pointer
	pc uint16 // register - program counter
	a  byte   // register - accumulator
	x  byte   // register - x index
	y  byte   // register - y index
	ps byte   // register - processor status
}

// newCPU initializes and returns a new Mos6502 CPU
func newCPU() *Mos6502 {
	return &Mos6502{
		sp: 0xFF,
		pc: 0,
		a:  0,
		x:  0,
		y:  0,
		ps: flagDefault,
	}
}
