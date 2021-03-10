package elf

import (
	"encoding/binary"
	"errors"
	"os"
)

type SeekableReader interface {
	Read(p []byte) (n int, err error)
	Seek(offset int64, whence int) (ret int64, err error)
}

func deserializeSection(file SeekableReader, section *ElfSection, byteOrder binary.ByteOrder) error {
	binary.Read(file, byteOrder, &section.nameOffset)
	binary.Read(file, byteOrder, &section.Type)
	binary.Read(file, byteOrder, &section.Flags)
	binary.Read(file, byteOrder, &section.Address)
	binary.Read(file, byteOrder, &section.Offset)
	binary.Read(file, byteOrder, &section.Size)
	binary.Read(file, byteOrder, &section.Link)
	binary.Read(file, byteOrder, &section.Info)
	binary.Read(file, byteOrder, &section.AddressAlign)
	binary.Read(file, byteOrder, &section.EntrySize)

	prevSection, err := file.Seek(0, os.SEEK_CUR)

	if err != nil {
		return err
	}

	file.Seek(int64(section.Offset), os.SEEK_SET)

	section.Data = make([]byte, section.Size)

	file.Read(section.Data)

	file.Seek(prevSection, os.SEEK_SET)

	return nil
}

func ParseElf(file SeekableReader) (*ElfFile, error) {
	var result ElfFile
	var byteOrder binary.ByteOrder = binary.BigEndian

	result.Header.eIdent = make([]byte, 16)

	_, err := file.Read(result.Header.eIdent)

	if err != nil {
		return nil, err
	}

	if result.Header.eIdent[EI_MAG0] != 0x7F ||
		result.Header.eIdent[EI_MAG1] != 0x45 ||
		result.Header.eIdent[EI_MAG2] != 0x4C ||
		result.Header.eIdent[EI_MAG3] != 0x46 {
		return nil, errors.New("Invalid ELF header")
	}

	if result.Header.eIdent[EI_CLASS] != byte(EI_CLASS_32_BIT) {
		return nil, errors.New("Only 32 bit elf file is supported")
	}

	if result.Header.eIdent[EI_DATA] == byte(EI_DATA_BIG_ENDIAN) {
		byteOrder = binary.BigEndian
	} else if result.Header.eIdent[EI_DATA] == byte(EI_DATA_LITTLE_ENDIAN) {
		byteOrder = binary.LittleEndian
	} else {
		return nil, errors.New("Unrecognized data type")
	}

	if result.Header.eIdent[EI_VERSION] != 1 {
		return nil, errors.New("Only version 1 of elf file is supported")
	}

	binary.Read(file, byteOrder, &result.Header.eType)
	binary.Read(file, byteOrder, &result.Header.eMachine)
	binary.Read(file, byteOrder, &result.Header.eVersion)
	binary.Read(file, byteOrder, &result.Header.eEntry)
	binary.Read(file, byteOrder, &result.Header.eProgramHeaderOff)
	binary.Read(file, byteOrder, &result.Header.eSectionHeaderOff)
	binary.Read(file, byteOrder, &result.Header.eFlags)
	binary.Read(file, byteOrder, &result.Header.eHeaderSize)
	binary.Read(file, byteOrder, &result.Header.eProgramHeaderSize)
	binary.Read(file, byteOrder, &result.Header.eProgramHeaderCount)
	binary.Read(file, byteOrder, &result.Header.eSectionHeaderSize)
	binary.Read(file, byteOrder, &result.Header.eSectionHeaderCount)
	binary.Read(file, byteOrder, &result.Header.eSectionNameEntry)

	result.Sections = make([]ElfSection, result.Header.eSectionHeaderCount)

	file.Seek(int64(result.Header.eSectionHeaderOff), os.SEEK_SET)

	for i := 0; i < int(result.Header.eSectionHeaderCount); i++ {
		err = deserializeSection(file, &result.Sections[i], byteOrder)

		if err != nil {
			return nil, err
		}
	}

	if result.Header.eSectionNameEntry < uint16(len(result.Sections)) {
		var sectionNames = &result.Sections[result.Header.eSectionNameEntry]

		for i := 0; i < int(result.Header.eSectionHeaderCount); i++ {
			result.Sections[i].Name = GetString(sectionNames, result.Sections[i].nameOffset)
		}
	}

	return &result, nil
}
