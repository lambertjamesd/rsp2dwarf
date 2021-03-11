package main

import (
	"encoding/binary"
	"io/ioutil"
	"os"

	"github.com/lambertjamesd/rsp2dwarf/dwarf"
	"github.com/lambertjamesd/rsp2dwarf/elf"
)

func appendDebugSymbols(elfFile *elf.ElfFile, textFilename string, textSectionLength int) error {
	symFile, err := os.Open(textFilename + ".sym")

	if err != nil {
		return err
	}

	defer symFile.Close()

	symData, err := ioutil.ReadAll(symFile)

	instructions, err := parseSymFile(string(symData))

	if err != nil {
		return err
	}

	elfFile.Sections = append(elfFile.Sections, elf.BuildElfSection(
		".debug_line",
		elf.SHT_MIPS_DWARF,
		0,
		0,
		0,
		0,
		1,
		dwarf.GenerateDebugLines(instructions, binary.BigEndian),
	))

	elfFile.Sections = append(elfFile.Sections, elf.BuildElfSection(
		".debug_aranges",
		elf.SHT_MIPS_DWARF,
		0,
		0,
		0,
		0,
		1,
		dwarf.GenerateAranges(textSectionLength, binary.BigEndian),
	))

	return nil
}

func buildElf(textFilename string, linkName string, includeDebug bool) (*elf.ElfFile, error) {
	var result = &elf.ElfFile{
		Header: elf.BuildElfHeader(
			elf.ET_REL,
			elf.EM_MIPS,
			0,
			0x20000101,
		),
		Sections: nil,
	}

	result.Sections = append(result.Sections, elf.BuildElfSection(
		"",
		elf.SHT_NULL,
		0,
		0,
		0,
		0,
		0,
		nil,
	))

	textFile, err := os.Open(textFilename)

	if err != nil {
		return nil, err
	}

	defer textFile.Close()

	textData, err := ioutil.ReadAll(textFile)

	if err != nil {
		return nil, err
	}

	result.Sections = append(result.Sections, elf.BuildElfSection(
		".text",
		elf.SHT_PROGBITS,
		elf.SHF_ALLOC|elf.SHF_EXECINSTR,
		0,
		0,
		0,
		16,
		textData,
	))

	dataFile, err := os.Open(textFilename + ".dat")

	if err != nil {
		return nil, err
	}

	defer dataFile.Close()

	dataData, err := ioutil.ReadAll(dataFile)

	if err != nil {
		return nil, err
	}

	result.Sections = append(result.Sections, elf.BuildElfSection(
		".data",
		elf.SHT_PROGBITS,
		elf.SHF_WRITE|elf.SHF_ALLOC,
		0,
		0,
		0,
		16,
		dataData,
	))

	if includeDebug {
		err = appendDebugSymbols(result, textFilename, len(textData))

		if err != nil {
			return nil, err
		}
	}

	result.AddSymbols([]elf.ElfSymbol{
		elf.BuildSymbol("", 0, 0, elf.STB_LOCAL, elf.STT_NOTYPE, 0, 0),
		elf.BuildSymbol(".text", 0, 0, elf.STB_LOCAL, elf.STT_SECTION, 0, 1),
		elf.BuildSymbol(".data", 0, 0, elf.STB_LOCAL, elf.STT_SECTION, 0, 1),
		elf.BuildSymbol(linkName+"TextStart", 0, uint32(len(textData)), elf.STB_GLOBAL, elf.STT_FUNC, 0, 1),
		elf.BuildSymbol(linkName+"TextEnd", uint32(len(textData)), 0, elf.STB_GLOBAL, elf.STT_FUNC, 0, 1),
		elf.BuildSymbol(linkName+"DataStart", 0, uint32(len(dataData)), elf.STB_GLOBAL, elf.STT_OBJECT, 0, 2),
		elf.BuildSymbol(linkName+"DataEnd", uint32(len(dataData)), 0, elf.STB_GLOBAL, elf.STT_OBJECT, 0, 2),
	}, binary.BigEndian)

	return result, nil
}
