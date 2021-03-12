package dwarf

import (
	"bytes"
	"encoding/binary"

	"github.com/lambertjamesd/rsp2dwarf/elf"
)

func GenerateAranges(textSegmentLength int, byteOrder binary.ByteOrder) ([]byte, *elf.RelocationBuilder) {
	var result bytes.Buffer
	var relBuilder = elf.NewRelocationBuilder()

	var size uint32 = 0x1C
	binary.Write(&result, byteOrder, &size)
	var version uint16 = 2
	binary.Write(&result, byteOrder, &version)
	var offset uint32 = 0
	binary.Write(&result, byteOrder, &offset)
	result.WriteByte(4) // size of instruction
	result.WriteByte(0) // segment descriptor size

	// padding
	result.WriteByte(0)
	result.WriteByte(0)
	result.WriteByte(0)
	result.WriteByte(0)

	// single range entry
	relBuilder.AddEntry(uint32(result.Len()), ".text", elf.R_MIPS_32)
	binary.Write(&result, byteOrder, &offset)
	size = uint32(textSegmentLength)
	relBuilder.AddEntry(uint32(result.Len()), ".text", elf.R_MIPS_32)
	binary.Write(&result, byteOrder, &size)

	// null terminator
	binary.Write(&result, byteOrder, &offset)
	binary.Write(&result, byteOrder, &offset)

	return result.Bytes(), relBuilder
}
