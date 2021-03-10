package main

import (
	"errors"
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
