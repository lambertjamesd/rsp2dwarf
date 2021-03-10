package main

import (
	"errors"
	"fmt"
	"os"
	"path"

	"github.com/lambertjamesd/rsp2dwarf/elf"
)

type commandLineArgs struct {
	output string
	input  string
	name   string
}

func parseCommandLineArgs() (*commandLineArgs, error) {
	var result commandLineArgs

	if len(os.Args) == 1 {
		return nil, errors.New(`rsp2dwarf [-n name] [-o output] input
	-n    the name to use in the linker
	-o    the output file`)
	}

	for i := 1; i < len(os.Args); i++ {
		var arg = os.Args[i]

		if arg[0] == '-' {
			if arg == "-o" {
				if i+1 >= len(os.Args) {
					return nil, errors.New("-o flag requires a parameter")
				} else {
					result.output = arg
				}
			} else if arg == "-n" {
				if i+1 >= len(os.Args) {
					return nil, errors.New("-n flag requires a parameter")
				} else {
					result.name = arg
				}
			}

		} else {
			if result.input != "" {
				return nil, errors.New("Only one input file is allowed")
			} else {
				result.input = arg
			}
		}
	}

	if result.input == "" {
		return nil, errors.New("An input file is required")
	}

	if result.name == "" {
		result.name = linkNameFromFileName(result.input)
	}

	if result.output == "" {
		result.output = result.input + ".o"
	}

	return &result, nil
}

func linkNameFromFileName(input string) string {
	input = path.Base(input)
	ext := path.Ext(input)

	if len(ext) > 0 {
		input = input[0 : len(input)-len(ext)]
	}

	var output []byte = nil

	if input[0] < 'A' || input[0] > 'Z' && input[0] < 'a' || input[0] > 'z' {
		output = append(output, '_')
	}

	for _, character := range input {
		if character >= 'A' && character <= 'Z' ||
			character >= 'a' && character <= 'z' ||
			character >= '0' && character <= '9' {
			output = append(output, byte(character))
		} else {
			output = append(output, '_')
		}
	}

	return string(output)
}

func main() {
	args, err := parseCommandLineArgs()

	if err != nil {
		fmt.Println(err.Error())
		os.Exit(1)
	}

	elfFile, err := buildElf(args.input, args.name)

	if err != nil {
		fmt.Println(err.Error())
		os.Exit(1)
	}

	outFile, err := os.OpenFile(args.output, os.O_WRONLY|os.O_TRUNC|os.O_CREATE, 0600)

	if err != nil {
		fmt.Println(err.Error())
		os.Exit(1)
	}

	defer outFile.Close()

	elf.Serialize(outFile, elfFile)
}
