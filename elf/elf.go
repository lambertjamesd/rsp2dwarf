package elf

import (
	"bytes"
	"encoding/binary"
)

const (
	EI_MAG0       = 0
	EI_MAG1       = 1
	EI_MAG2       = 2
	EI_MAG3       = 3
	EI_CLASS      = 4
	EI_DATA       = 5
	EI_VERSION    = 6
	EI_OSABI      = 7
	EI_ABIVERSION = 8
)

type EIClass byte

const (
	EI_CLASS_32_BIT EIClass = 1
	EI_CLASS_64_BIT EIClass = 2
)

type EIData byte

const (
	EI_DATA_LITTLE_ENDIAN EIData = 1
	EI_DATA_BIG_ENDIAN    EIData = 2
)

type ElfType uint16

const (
	ET_NONE   = 0
	ET_REL    = 1
	ET_EXEC   = 2
	ET_DYN    = 3
	ET_CORE   = 4
	ET_LOPROC = 0xFF00
	ET_HIPROC = 0xFFFF
)

type ElfMachine uint16

const (
	EM_NONE  = 0
	EM_M32   = 1
	EM_SPARC = 2
	EM_386   = 3
	EM_68K   = 4
	EM_88K   = 5
	EM_860   = 7
	EM_MIPS  = 8
)

type ElfHeader struct {
	eIdent              []byte
	eType               ElfType
	eMachine            ElfMachine
	eVersion            uint32
	eEntry              uint32
	eProgramHeaderOff   uint32
	eSectionHeaderOff   uint32
	eFlags              uint32
	eHeaderSize         uint16
	eProgramHeaderSize  uint16
	eProgramHeaderCount uint16
	eSectionHeaderSize  uint16
	eSectionHeaderCount uint16
	eSectionNameEntry   uint16
}

func BuildElfHeader(
	eType ElfType,
	eMachine ElfMachine,
	entry uint32,
	flags uint32,
) ElfHeader {
	var ident = make([]byte, 16)

	ident[EI_MAG0] = 0x7F
	ident[EI_MAG1] = 0x45
	ident[EI_MAG2] = 0x4C
	ident[EI_MAG3] = 0x46

	ident[EI_CLASS] = byte(EI_CLASS_32_BIT)
	ident[EI_DATA] = byte(EI_DATA_BIG_ENDIAN)
	ident[EI_VERSION] = 1

	return ElfHeader{
		ident,
		eType,
		eMachine,
		1,
		entry,
		0,
		0,
		flags,
		0x34,
		0,
		0,
		0x28,
		0,
		0,
	}
}

type SectionType uint32

const (
	SHT_NULL       SectionType = 0
	SHT_PROGBITS   SectionType = 1
	SHT_SYMTAB     SectionType = 2
	SHT_STRTAB     SectionType = 3
	SHT_RELA       SectionType = 4
	SHT_HASH       SectionType = 5
	SHT_DYNAMIC    SectionType = 6
	SHT_NOTE       SectionType = 7
	SHT_NOBITS     SectionType = 8
	SHT_REL        SectionType = 9
	SHT_SHLIB      SectionType = 10
	SHT_DYNSYM     SectionType = 11
	SHT_INIT_ARRAY SectionType = 12
	SHT_FINI_ARRAY SectionType = 13
	SHT_MIPS_DWARF SectionType = 0x7000001E
)

type SectionHeaderFlags uint32

const (
	SHF_WRITE            SectionHeaderFlags = 0x1
	SHF_ALLOC            SectionHeaderFlags = 0x2
	SHF_EXECINSTR        SectionHeaderFlags = 0x4
	SHF_MERGE            SectionHeaderFlags = 0x10
	SHF_STRINGS          SectionHeaderFlags = 0x20
	SHF_INFO_LINK        SectionHeaderFlags = 0x40
	SHF_LINK_ORDER       SectionHeaderFlags = 0x80
	SHF_OS_NONCONFORMING SectionHeaderFlags = 0x100
	SHF_GROUP            SectionHeaderFlags = 0x200
	SHF_TLS              SectionHeaderFlags = 0x400
	SHF_MASKOS           SectionHeaderFlags = 0x0ff00000
	SHF_MASKPROC         SectionHeaderFlags = 0xf0000000
	SHF_ORDERED          SectionHeaderFlags = 0x4000000
	SHF_EXCLUDE          SectionHeaderFlags = 0x8000000
)

type ElfSection struct {
	nameOffset   uint32
	Name         string
	Type         SectionType
	Flags        SectionHeaderFlags
	Address      uint32
	Offset       uint32
	Size         uint32
	Link         uint32
	Info         uint32
	AddressAlign uint32
	EntrySize    uint32

	Data []byte
}

type ElfFile struct {
	Header   ElfHeader
	Sections []ElfSection
}

func BuildElfSection(
	name string,
	sType SectionType,
	flags SectionHeaderFlags,
	addr uint32,
	link uint32,
	info uint32,
	align uint32,
	data []byte,
) ElfSection {
	var entrySize uint32 = 0

	if sType == SHT_SYMTAB {
		entrySize = 0x10
	} else if sType == SHT_STRTAB {
		entrySize = 1
	}

	return ElfSection{
		0,
		name,
		sType,
		flags,
		addr,
		0,
		uint32(len(data)),
		link,
		info,
		align,
		entrySize,
		data,
	}
}

func GetString(section *ElfSection, offset uint32) string {
	if offset == 0 || offset >= uint32(len(section.Data)) {
		return ""
	} else {
		var endIndex = offset

		for endIndex < uint32(len(section.Data)) && section.Data[endIndex] != 0 {
			endIndex++
		}

		return string(section.Data[offset:endIndex])
	}
}

func (elfFile *ElfFile) FindSectionIndex(name string) int {
	for i, section := range elfFile.Sections {
		if section.Name == sectionHeaderStringName {
			return i
		}
	}

	return -1
}

type ElfSymbol struct {
	Name       string
	nameOffset uint32
	Value      uint32
	Size       uint32
	Info       uint8
	Other      uint8
	SHIndex    uint16
}

type SymbolBinding uint8

const (
	STB_LOCAL  SymbolBinding = 0
	STB_GLOBAL SymbolBinding = 1
	STB_WEAK   SymbolBinding = 2
	STB_LOPROC SymbolBinding = 13
	STB_HIPROC SymbolBinding = 15
)

type SymbolType uint8

const (
	STT_NOTYPE  SymbolType = 0
	STT_OBJECT  SymbolType = 1
	STT_FUNC    SymbolType = 2
	STT_SECTION SymbolType = 3
	STT_FILE    SymbolType = 4
	STT_LOPROC  SymbolType = 13
	STT_HIPROC  SymbolType = 15
)

func BuildSymbol(
	name string,
	value uint32,
	size uint32,
	binding SymbolBinding,
	symbolType SymbolType,
	other uint8,
	shIndex uint16,
) ElfSymbol {
	return ElfSymbol{
		name,
		0,
		value,
		size,
		(uint8(binding) << 4) + (uint8(symbolType) & 0xF),
		other,
		shIndex,
	}
}

func (elfFile *ElfFile) AddSymbols(symbols []ElfSymbol, byteOrder binary.ByteOrder) {
	var stringIndex = elfFile.FindSectionIndex(".strtab")

	if stringIndex == -1 {
		stringIndex = len(elfFile.Sections)

		elfFile.Sections = append(elfFile.Sections, BuildElfSection(
			".strtab",
			SHT_STRTAB,
			0,
			0,
			0,
			0,
			0,
			make([]byte, 1),
		))
	}

	var strTab = &elfFile.Sections[stringIndex]

	var buffer bytes.Buffer

	var info = 0

	for index, symbol := range symbols {
		data, nameOffset := AddStringToSection(strTab.Data, symbol.Name)

		strTab.Data = data
		symbol.nameOffset = uint32(nameOffset)

		binary.Write(&buffer, byteOrder, &symbol.nameOffset)
		binary.Write(&buffer, byteOrder, &symbol.Value)
		binary.Write(&buffer, byteOrder, &symbol.Size)
		binary.Write(&buffer, byteOrder, &symbol.Info)
		binary.Write(&buffer, byteOrder, &symbol.Other)
		binary.Write(&buffer, byteOrder, &symbol.SHIndex)

		if (symbol.Info >> 4) == uint8(STB_LOCAL) {
			info = index + 1
		}
	}

	var symbolIndex = elfFile.FindSectionIndex(".strtab")

	if symbolIndex == -1 {
		elfFile.Sections = append(elfFile.Sections, BuildElfSection(
			".symtab",
			SHT_SYMTAB,
			0,
			0,
			uint32(stringIndex),
			uint32(info),
			0,
			buffer.Bytes(),
		))
	} else {
		var symTab = &elfFile.Sections[symbolIndex]
		symTab.Data = append(symTab.Data, buffer.Bytes()...)
		symTab.Link = uint32(stringIndex)
	}
}
