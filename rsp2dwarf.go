package main

import (
	"fmt"
	"os"

	"github.com/lambertjamesd/rsp2dwarf/elf"
)

func main() {
	file, err := os.Open(os.Args[1])

	if err != nil {
		fmt.Println(err.Error())
		os.Exit(1)
	}

	elfFile, err := elf.ParseElf(file)

	defer file.Close()

	if err != nil {
		fmt.Println(err.Error())
		os.Exit(1)
	}

	fmt.Println(len(elfFile.Sections))

	outFile, err := os.OpenFile("test.o", os.O_WRONLY|os.O_TRUNC|os.O_CREATE, 0600)

	if err != nil {
		fmt.Println(err.Error())
		os.Exit(1)
	}

	defer outFile.Close()

	elf.Serialize(outFile, elfFile)
}
