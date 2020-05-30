package vm

// block represents a 64kiB memory block
type block [64 * 1024]byte

func newBlock() [64 * 1024]byte {
	return [64 * 1024]byte{}
}

// load loads a program into memory at the provided address space
func (b block) load(addr uint16, data []uint8) {
	end := int(addr) + len(data)

	for i := int(addr); i < end; i++ {
		b[int(addr)+i] = data[i]
	}
}
