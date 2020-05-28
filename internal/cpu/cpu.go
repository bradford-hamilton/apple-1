// Package cpu emulates a Mos6502 cpu
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
// SR Flags (bit 7 to bit 0):
// N	....	Negative
// V	....	Overflow
// -	....	ignored
// B	....	Break
// D	....	Decimal (use BCD for arithmetics)
// I	....	Interrupt (IRQ disable)
// Z	....	Zero
// C	....	Carry
//
// Processor Stack:
// LIFO, top down, 8 bit range, 0x0100 - 0x01FF
//
// 6502 instructions have the general form AAABBBCC, where AAA and CC define the opcode, and BBB defines the addressing mode
package cpu

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
	flagDefault           uint8 = 0B_00110000
	flagNegative          uint8 = 0B_10000000
	flagOverflow          uint8 = 0B_01000000
	flagBreak             uint8 = 0B_00010000
	flagDecimalMode       uint8 = 0B_00001000
	flagDisableInterrupts uint8 = 0B_00000100
	flagZero              uint8 = 0B_00000010
	flagCarry             uint8 = 0B_00000001
)

// StackBottom represents the bottom address
const StackBottom uint16 = 0x0100 // 256

// Mos6502 TODO: docs
type Mos6502 struct {
	sp uint8  // register - stack pointer
	pc uint16 // register - program counter
	a  uint8  // register - accumulator
	x  uint8  // register - x index
	y  uint8  // register - y index
	ps uint8  // register - processor status
}

// New TODO: docs
func New() *Mos6502 {
	return &Mos6502{
		sp: 0xFF,
		pc: 0,
		a:  0,
		x:  0,
		y:  0,
		ps: flagDefault,
	}
}
