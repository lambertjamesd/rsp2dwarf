package dwarf

import (
	"bytes"
	"encoding/binary"
	"sort"

	"github.com/lambertjamesd/rsp2dwarf/elf"
)

const lineBase = 0
const lineRange = 16
const opcodeBase = 10
const minInstructionLen = 4

var opcodeLengths = []byte{
	0, 1, 1, 1, 1, 0, 0, 0, 1,
}

const (
	DW_LNS_copy             = 1
	DW_LNS_advance_pc       = 2
	DW_LNS_advance_line     = 3
	DW_LNS_set_file         = 4
	DW_LNS_set_column       = 5
	DW_LNS_negate_stmt      = 6
	DW_LNS_set_basic_block  = 7
	DW_LNS_const_add_pc     = 8
	DW_LNS_fixed_advance_pc = 9
)

const (
	DW_LNE_end_sequence = 1
	DW_LNE_set_address  = 2
	DW_LNE_define_file  = 3
)

type InstructionEntry struct {
	address     int
	filename    string
	line        int
	col         int
	isStatement bool
	isBlock     bool
}

func CreateInstructionEntry(
	address int,
	filename string,
	line int,
	col int,
	isStatement bool,
	isBlock bool,
) InstructionEntry {
	return InstructionEntry{
		address,
		filename,
		line,
		col,
		isStatement,
		isBlock,
	}
}

func (entry *InstructionEntry) Filename() string {
	return entry.filename
}

type instructionEntryByAddress []InstructionEntry

func (arr instructionEntryByAddress) Len() int {
	return len(arr)
}

func (arr instructionEntryByAddress) Less(i, j int) bool {
	return arr[i].address < arr[j].address
}

func (arr instructionEntryByAddress) Swap(i, j int) {
	arr[i], arr[j] = arr[j], arr[i]
}

func sortAndFilter(instructions []InstructionEntry) []InstructionEntry {
	var sorted instructionEntryByAddress = make([]InstructionEntry, len(instructions))

	for index, entry := range instructions {
		sorted[index] = entry
	}

	sort.Sort(sorted)

	var readIndex = 0
	var writeIndex = 0

	for readIndex < len(sorted) {
		if writeIndex != readIndex {
			sorted[writeIndex] = sorted[readIndex]
		}

		var written = writeIndex

		writeIndex++
		readIndex++

		for readIndex < len(sorted) &&
			sorted[readIndex].line == sorted[written].line &&
			sorted[readIndex].col == sorted[written].col &&
			sorted[readIndex].filename == sorted[written].filename {
			// skip duplicate lines
			readIndex++
		}
	}

	return sorted[0:writeIndex]
}

func getSpecialOpcode(lineDelta int, instructionDelta int) int {
	if lineDelta >= lineRange || lineDelta < lineBase {
		return -1
	}
	return (lineDelta - lineBase) + (lineRange * instructionDelta / minInstructionLen) + opcodeBase
}

func findFile(files []string, filename string) int {
	for index, file := range files {
		if file == filename {
			return index + 1
		}
	}

	return 0
}

func generateOpCodes(instructions []InstructionEntry, files []string, isStmt bool) []byte {
	var result bytes.Buffer

	var address = 0
	var file = 1
	var line = 1
	var col = 0
	var basicBlock = false

	result.WriteByte(0) // extended opcode
	result.WriteByte(5) // size of extended operation
	result.WriteByte(DW_LNE_set_address)
	result.WriteByte(0) // address will be modified by the rel table
	result.WriteByte(0)
	result.WriteByte(0)
	result.WriteByte(0)

	for _, inst := range instructions {
		var instFile = findFile(files, inst.filename)

		if instFile != file {
			result.WriteByte(DW_LNS_set_file)
			writeULEB128(&result, uint64(instFile))
			file = instFile
		}

		if col != inst.col {
			result.WriteByte(DW_LNS_set_column)
			writeULEB128(&result, uint64(inst.col))
			col = inst.col
		}

		if !basicBlock && inst.isBlock {
			result.WriteByte(DW_LNS_set_basic_block)
			basicBlock = true
		}

		if inst.isStatement != isStmt {
			result.WriteByte(DW_LNS_negate_stmt)
			isStmt = !isStmt
		}

		var specialOp = getSpecialOpcode(inst.line-line, inst.address-address)

		if specialOp >= opcodeBase && specialOp < 256 {
			result.WriteByte(byte(specialOp))
			line = inst.line
			address = inst.address
		} else {
			if inst.address != address {
				result.WriteByte(DW_LNS_advance_pc)
				writeULEB128(&result, uint64(inst.address-address)/minInstructionLen)
				address = inst.address
			}

			if inst.line != line {
				result.WriteByte(DW_LNS_advance_line)
				writeSLEB128(&result, int64(inst.line-line))
				line = inst.line
			}

			result.WriteByte(DW_LNS_copy)
		}
	}

	result.WriteByte(0) // extended opcode
	result.WriteByte(1) // size of extended operation
	result.WriteByte(DW_LNE_end_sequence)

	return result.Bytes()
}

func GenerateDebugLines(instructions []InstructionEntry, byteOrder binary.ByteOrder) ([]byte, *elf.RelocationBuilder) {
	var sorted = sortAndFilter(instructions)
	var relBuilder = elf.NewRelocationBuilder()

	var files []string = nil
	var filesNameByteLength = 0

	for _, inst := range sorted {
		if findFile(files, inst.filename) == 0 {
			files = append(files, inst.filename)
			filesNameByteLength += len([]byte(inst.filename))
		}
	}

	var generated = generateOpCodes(sorted, files, sorted[0].isStatement)

	var result bytes.Buffer

	var prologueLength = uint32(
		7 + len(opcodeLengths) +
			filesNameByteLength + len(files)*4,
	)

	var totalLength = uint32(len(generated)) + prologueLength + 6
	binary.Write(&result, byteOrder, &totalLength)
	var version = uint16(2)
	binary.Write(&result, byteOrder, &version)
	binary.Write(&result, byteOrder, &prologueLength)
	result.WriteByte(minInstructionLen)
	if sorted[0].isStatement {
		result.WriteByte(1)
	} else {
		result.WriteByte(byte(0))
	}
	result.WriteByte(lineBase)
	result.WriteByte(lineRange)
	result.WriteByte(opcodeBase)
	result.Write(opcodeLengths)

	result.WriteByte(0) // directories

	for _, file := range files {
		result.Write([]byte(file))
		result.WriteByte(0) // null terminated
		result.WriteByte(0) // directory
		result.WriteByte(0) // last modification
		result.WriteByte(0) // size
	}

	result.WriteByte(0) // end of files

	relBuilder.AddEntry(uint32(result.Len())+3, ".text", elf.R_MIPS_32)

	result.Write(generated)

	return result.Bytes(), relBuilder
}
