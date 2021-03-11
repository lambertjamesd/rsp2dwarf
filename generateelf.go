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
		0,
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
		0,
		dwarf.GenerateAranges(textSectionLength, binary.BigEndian),
	))

	var attributes = []*dwarf.AbbrevTreeNode{
		{
			Tag: dwarf.DW_TAG_compile_unit,
			Attributes: []dwarf.AbbrevAttr{
				dwarf.CreateConstantAttr(dwarf.DW_AT_stmt_list, 0, 4),
				dwarf.CreateAddrAttr(dwarf.DW_AT_low_pc, 0),
				dwarf.CreateAddrAttr(dwarf.DW_AT_high_pc, 0),
				dwarf.CreateStringAttr(dwarf.DW_AT_name, "", false),
				dwarf.CreateStringAttr(dwarf.DW_AT_comp_dir, "", false),
				dwarf.CreateStringAttr(dwarf.DW_AT_producer, "rspasm", false),
				dwarf.CreateConstantAttr(dwarf.DW_AT_stmt_list, dwarf.DW_LANG_Mips_Assembler, 2),
			},
			Children: nil,
		},
	}

	var infoSections = dwarf.GenerateInfoAndAbbrev(attributes, binary.BigEndian)

	elfFile.Sections = append(elfFile.Sections, elf.BuildElfSection(
		".debug_info",
		elf.SHT_MIPS_DWARF,
		0,
		0,
		0,
		0,
		1,
		0,
		infoSections.Info,
	))

	elfFile.Sections = append(elfFile.Sections, elf.BuildElfSection(
		".debug_abbrev",
		elf.SHT_MIPS_DWARF,
		0,
		0,
		0,
		0,
		1,
		0,
		infoSections.Abbrev,
	))

	elfFile.Sections = append(elfFile.Sections, elf.BuildElfSection(
		".debug_str",
		elf.SHT_MIPS_DWARF,
		0,
		0,
		0,
		0,
		1,
		1,
		infoSections.DebugStr,
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
		0,
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
		0,
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
