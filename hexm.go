package main

import (
	"fmt"
	"os"

	"github.com/marcinbor85/gohex"
)

func main() {
	//As we only have trivial command line args, simpler to custom parse
	inputFiles, outputFile, err := parseArgs(os.Args[1:])
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error %v\n", err)
		return
	}
	fmt.Printf("Input Files: %v\n", inputFiles)
	fmt.Printf("Output file: %s\n", outputFile)
	err = validateFiles(inputFiles, outputFile)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error %v\n", err)
		return
	}
	outputMemory := gohex.NewMemory()
	//Parse all input files into virtual memory space
	for i, inputFilePath := range inputFiles {
		fmt.Printf("Loading file %d => %s\r\n", i+1, inputFilePath)
		mem, err := parseInputFile(inputFilePath)
		if err != nil {
			panic(err)
		}
		mergeSegments(outputMemory, mem, inputFilePath)
	}
	// Now we want to write out the file, if its hex then we can use the hex writer, otherwise we will want to persist it out to bin
	err = writeOutput(outputFile, outputMemory)
	if err == nil {
		fmt.Println("Output created")
	} else {
		fmt.Printf("Creating output file raised error %v", err)
	}
}
