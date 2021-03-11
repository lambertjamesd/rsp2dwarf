package elf

import (
	"encoding/binary"
	"errors"
	"os"
	"sort"
)

type SeekableWriter interface {
	Write(p []byte) (n int, err error)
	Seek(offset int64, whence int) (ret int64, err error)
}

const sectionHeaderStringName = ".shstrtab"

func FindStringIndex(data []byte, str string) int {
	var current int = 1
	var stringIndex int = 0

	var stringAsBytes = []byte(str)

	if str == "" {
		return 0
	}

	for current < len(data) {
		if stringAsBytes[stringIndex] == data[current+stringIndex] {
			stringIndex++

			if stringIndex == len(stringAsBytes) {
				return current
			}
		} else {
			current++
			stringIndex = 0
		}
	}

	return -1
}

func AddStringToSection(data []byte, value string) ([]byte, int) {
	var index = FindStringIndex(data, value)

	if index == -1 {
		index = len(data)
		data = append(data, []byte(value)...)
		data = append(data, 0)
	}

	return data, index
}

type sortByLength []string

func (arr sortByLength) Len() int {
	return len(arr)
}

func (arr sortByLength) Less(i, j int) bool {
	return len(arr[i]) > len(arr[j])
}

func (arr sortByLength) Swap(i, j int) {
	arr[i], arr[j] = arr[j], arr[i]
}

func BuildStringSection(name string, values []string) ElfSection {
	var data []byte = make([]byte, 1)

	var valuesSorted sortByLength = values

	sort.Sort(valuesSorted)

	for _, value := range valuesSorted {
		data, _ = AddStringToSection(data, value)
	}

	return BuildElfSection(
		name,
		SHT_STRTAB,
		0,
		0,
		0,
		0,
		0,
		1,
		data,
	)
}

func rebuildSectionHeaders(elfFile *ElfFile) {
	var index = -1
	var sectionNames []string = nil

	for i, section := range elfFile.Sections {
		if section.Name == sectionHeaderStringName {
			index = i
		}

		sectionNames = append(sectionNames, section.Name)
	}

	if index == -1 {
		index = len(elfFile.Sections)
		sectionNames = append(sectionNames, sectionHeaderStringName)

		elfFile.Sections = append(elfFile.Sections, BuildStringSection(sectionHeaderStringName, sectionNames))
	} else {
		elfFile.Sections[index] = BuildStringSection(sectionHeaderStringName, sectionNames)
	}

	for i, _ := range elfFile.Sections {
		elfFile.Sections[i].nameOffset = uint32(FindStringIndex(elfFile.Sections[index].Data, elfFile.Sections[i].Name))
	}

	elfFile.Header.eSectionNameEntry = uint16(index)
	elfFile.Header.eSectionHeaderCount = uint16(len(elfFile.Sections))
}

func Serialize(writer SeekableWriter, elfFile *ElfFile) error {
	var byteOrder binary.ByteOrder = binary.BigEndian

	if elfFile.Header.eIdent[EI_DATA] == byte(EI_DATA_BIG_ENDIAN) {
		byteOrder = binary.BigEndian
	} else if elfFile.Header.eIdent[EI_DATA] == byte(EI_DATA_LITTLE_ENDIAN) {
		byteOrder = binary.LittleEndian
	} else {
		return errors.New("Unrecognized data type")
	}

	rebuildElfSymbolsAndStrings(elfFile, byteOrder)
	rebuildSectionHeaders(elfFile)

	writer.Seek(int64(elfFile.Header.eHeaderSize), os.SEEK_SET)

	for index, _ := range elfFile.Sections {
		var section = &elfFile.Sections[index]

		if section.Type == SHT_NULL {
			section.Size = 0
			section.Offset = 0
		} else {
			section.Size = uint32(len(section.Data))
			currentLocation, _ := writer.Seek(0, os.SEEK_CUR)

			if section.AddressAlign != 0 {
				var misalign = currentLocation % int64(section.AddressAlign)
				if misalign != 0 {
					misalign = int64(section.AddressAlign) - misalign
					writer.Write(make([]byte, misalign))
					currentLocation += misalign
				}
			}

			section.Offset = uint32(currentLocation)
			writer.Write(section.Data)
		}
	}

	programHeaderOffset, _ := writer.Seek(0, os.SEEK_CUR)

	elfFile.Header.eSectionHeaderOff = uint32(programHeaderOffset)
	elfFile.Header.eSectionHeaderSize = uint16(0x28)

	for _, section := range elfFile.Sections {
		binary.Write(writer, byteOrder, &section.nameOffset)
		binary.Write(writer, byteOrder, &section.Type)
		binary.Write(writer, byteOrder, &section.Flags)
		binary.Write(writer, byteOrder, &section.Address)
		binary.Write(writer, byteOrder, &section.Offset)
		binary.Write(writer, byteOrder, &section.Size)
		binary.Write(writer, byteOrder, &section.Link)
		binary.Write(writer, byteOrder, &section.Info)
		binary.Write(writer, byteOrder, &section.AddressAlign)
		binary.Write(writer, byteOrder, &section.EntrySize)
	}

	writer.Seek(0, os.SEEK_SET)

	writer.Write(elfFile.Header.eIdent)

	binary.Write(writer, byteOrder, &elfFile.Header.eType)
	binary.Write(writer, byteOrder, &elfFile.Header.eMachine)
	binary.Write(writer, byteOrder, &elfFile.Header.eVersion)
	binary.Write(writer, byteOrder, &elfFile.Header.eEntry)
	binary.Write(writer, byteOrder, &elfFile.Header.eProgramHeaderOff)
	binary.Write(writer, byteOrder, &elfFile.Header.eSectionHeaderOff)
	binary.Write(writer, byteOrder, &elfFile.Header.eFlags)
	binary.Write(writer, byteOrder, &elfFile.Header.eHeaderSize)
	binary.Write(writer, byteOrder, &elfFile.Header.eProgramHeaderSize)
	binary.Write(writer, byteOrder, &elfFile.Header.eProgramHeaderCount)
	binary.Write(writer, byteOrder, &elfFile.Header.eSectionHeaderSize)
	binary.Write(writer, byteOrder, &elfFile.Header.eSectionHeaderCount)
	binary.Write(writer, byteOrder, &elfFile.Header.eSectionNameEntry)

	return nil
}
