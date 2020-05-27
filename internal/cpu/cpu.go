package cpu

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

// 16-bit address words are little endian, lo(w)-byte first, followed by the hi(gh)-byte.

// SR Flags (bit 7 to bit 0):
// N	....	Negative
// V	....	Overflow
// -	....	ignored
// B	....	Break
// D	....	Decimal (use BCD for arithmetics)
// I	....	Interrupt (IRQ disable)
// Z	....	Zero
// C	....	Carry

// Processor Stack:
// LIFO, top down, 8 bit range, 0x0100 - 0x01FF

// addressingMode is a type alias for a string, used below for defining addressing mode types
type addressingMode int

const (
	accumulator addressingMode = iota
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

// Mos6502 TODO: docs
type Mos6502 struct {
	pc uint16 // program counter (16 bit)
	ac uint8  // accumulator (8 bit)
	x  uint8  // X register (8 bit)
	y  uint8  // Y register (8 bit)
	sr uint8  // status register [NV-BDIZC] (8 bit)
	sp uint8  // stack pointer (8 bit)
}

// New TODO: docs
func New() *Mos6502 {
	return &Mos6502{}
}
