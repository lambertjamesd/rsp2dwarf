package main

import (
	"errors"
	"sort"
	"strconv"
	"strings"

	"github.com/lambertjamesd/rsp2dwarf/dwarf"
)

func parseMaybeHex(input string, bitSize int) (int64, error) {
	if input[0:2] == "0x" || input[0:2] == "0X" {
		return strconv.ParseInt(input[2:], 16, bitSize)
	} else {
		return strconv.ParseInt(input, 10, bitSize)
	}
}

func parseSymFile(input string) ([]dwarf.InstructionEntry, error) {
	var lines = strings.Split(input, "\n")

	var result []dwarf.InstructionEntry = nil

	for _, line := range lines {
		var parts = strings.Split(line, " ")

		if parts[0] == "line" {
			if len(parts) < 4 {
				return nil, errors.New("Line should have at least 4 arguments")
			}

			addr, err := parseMaybeHex(parts[1], 32)

			if err != nil {
				return nil, err
			}

			lineNumber, err := strconv.ParseInt(parts[3], 10, 32)

			if err != nil {
				return nil, err
			}

			result = append(result, dwarf.CreateInstructionEntry(
				int(addr),
				parts[2],
				int(lineNumber),
				0,
				true,
				false,
			))
		}
	}

	return result, nil
}

type SymbolDef struct {
	Name  string
	Value uint32
	Size  uint32
}

type SortSymbolsByValue []SymbolDef

func (arr SortSymbolsByValue) Len() int {
	return len(arr)
}

func (arr SortSymbolsByValue) Less(i, j int) bool {
	return arr[i].Value < arr[j].Value
}

func (arr SortSymbolsByValue) Swap(i, j int) {
	arr[i], arr[j] = arr[j], arr[i]
}

func assignRanges(instructionSymbols SortSymbolsByValue, size int) []SymbolDef {
	sort.Sort(instructionSymbols)

	for index, _ := range instructionSymbols {
		if index != 0 {
			instructionSymbols[index-1].Size = instructionSymbols[index].Value - instructionSymbols[index-1].Value
		}
	}

	if len(instructionSymbols) > 0 {
		var lastEntry = &instructionSymbols[len(instructionSymbols)-1]
		lastEntry.Size = uint32(size) - lastEntry.Value
	}

	return instructionSymbols
}

func parseDbgFile(input string, textSize int, dataSize int) ([]SymbolDef, []SymbolDef) {
	var instructionSymbols SortSymbolsByValue = nil
	var dataSymbols SortSymbolsByValue = nil

	var lines = strings.Split(input, "\n")

	for _, line := range lines {
		var parts = strings.Split(line, " ")

		if len(parts) == 3 {
			addr, _ := strconv.ParseInt(parts[1], 16, 32)

			if parts[2] == "I" {
				instructionSymbols = append(instructionSymbols, SymbolDef{parts[0], uint32(addr), 0})
			} else if parts[2] == "D" {
				dataSymbols = append(dataSymbols, SymbolDef{parts[0], uint32(addr), 0})
			}
		}
	}

	var finalInstruction SortSymbolsByValue = nil

	for _, check := range instructionSymbols {
		var isDuplicate = false

		for _, dataSymbol := range dataSymbols {
			if dataSymbol.Name == check.Name {
				isDuplicate = true
				break
			}
		}

		if isDuplicate {
			continue
		}

		finalInstruction = append(finalInstruction, check)
	}

	return assignRanges(finalInstruction, textSize), assignRanges(dataSymbols, dataSize)
}
