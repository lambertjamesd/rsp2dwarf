package elf

import (
	"bytes"
	"encoding/binary"
	"io"
)

type RelocationType uint8

const (
	R_MIPS_NONE    RelocationType = 0
	R_MIPS_16      RelocationType = 1
	R_MIPS_32      RelocationType = 2
	R_MIPS_REL32   RelocationType = 3
	R_MIPS_26      RelocationType = 4
	R_MIPS_HI16    RelocationType = 5
	R_MIPS_LO16    RelocationType = 6
	R_MIPS_GPREL16 RelocationType = 7
	R_MIPS_LITERAL RelocationType = 8
	R_MIPS_GOT16   RelocationType = 9
	R_MIPS_PC16    RelocationType = 10
	R_MIPS_CALL16  RelocationType = 11
	R_MIPS_GPREL32 RelocationType = 12
)

type RelocationEntry struct {
	Offset     uint32
	SymbolName string
	Type       RelocationType
}

type RelocationBuilder struct {
	entries []RelocationEntry
}

func NewRelocationBuilder() *RelocationBuilder {
	return &RelocationBuilder{nil}
}

func (builder *RelocationBuilder) AddEntry(offset uint32, symbolName string, relType RelocationType) {
	builder.entries = append(builder.entries, RelocationEntry{offset, symbolName, relType})
}

func (builder *RelocationBuilder) Serialize(writer io.Writer, symbolIndexMapping map[string]uint32, byteOrder binary.ByteOrder) {
	for _, entry := range builder.entries {
		binary.Write(writer, byteOrder, &entry.Offset)
		symbolIndex, _ := symbolIndexMapping[entry.SymbolName]
		var combinedIndexType = (symbolIndex << 8) | (uint32(entry.Type) & 0xFF)
		binary.Write(writer, byteOrder, &combinedIndexType)
	}
}

func (builder *RelocationBuilder) ToElfSection(forSection string, symbolIndexMapping map[string]uint32, byteOrder binary.ByteOrder) ElfSection {
	var buffer bytes.Buffer

	builder.Serialize(&buffer, symbolIndexMapping, byteOrder)

	return BuildElfSection(
		".rel"+forSection,
		SHT_REL,
		0,
		0,
		0,
		0,
		4,
		8,
		buffer.Bytes(),
	)
}

func (builder *RelocationBuilder) AddOffset(offset uint32) {
	for index, _ := range builder.entries {
		builder.entries[index].Offset += offset
	}
}
