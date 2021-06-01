package main

import (
	"bufio"
	"fmt"
	"log"
	"os"
	"strings"

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
		mem := gohex.NewMemory()
		{
			fmt.Printf("Loading file %d\r\n", i+1)
			file, err := os.Open(inputFilePath)
			if err != nil {
				panic(err)
			}
			defer file.Close()
			err = mem.ParseIntelHex(file)
			if err != nil {
				panic(err)
			}
		}
		existingSegments := outputMemory.GetDataSegments()

		for x, segment := range mem.GetDataSegments() {
			fmt.Printf("Section %d @ 0x%08X ; len %d\n", x+1, segment.Address, len(segment.Data))
			//Check if this segment overlaps the existing segments
			if len(existingSegments) > 0 {
				for _, seg2 := range existingSegments {
					if !segmentOverlaps(segment, seg2) || userConfirmOverlap(segment, inputFilePath) {
						//write this segment into it
						outputMemory.SetBinary(segment.Address, segment.Data)
					}
				}
			} else {
				outputMemory.SetBinary(segment.Address, segment.Data)
			}
		}
	}
	// Now we want to write out the file, if its hex then we can use the hex writer, otherwise we will want to persist it out to bin
	outputHex, binaryStart, err := parseFileTypeAndStart(outputFile)
	if err != nil {
		panic(err)
	}
	file, err := os.Create(outputFile)
	if err != nil {
		panic(err)
	}
	defer file.Close()

	if outputHex {
		outputMemory.DumpIntelHex(file, 32)
	} else {
		//We want to write a binary file starting at the specified location, and padding all gaps
		existingSegments := outputMemory.GetDataSegments()
		//Write out each section
		for i, section := range existingSegments {
			start := section.Address - uint32(binaryStart)
			fmt.Printf("Writing %v bytes @ %08X for section %d\r\n", len(section.Data), start, i+1)
			_, err = file.WriteAt(section.Data, int64(start))
			if err != nil {
				panic(err)
			}
		}
	}
}

func userConfirmOverlap(seg gohex.DataSegment, source string) bool {
	return userConfirm(fmt.Sprintf("Merging segment @ 0x%08X from file %v will overwrite existing data, continue ?", seg.Address, source))
}

func segmentOverlaps(seg gohex.DataSegment, seg2 gohex.DataSegment) bool {
	if ((seg2.Address >= seg.Address) && (seg2.Address < seg.Address+uint32(len(seg.Data)))) ||
		((seg2.Address < seg.Address) && (seg2.Address+uint32(len(seg2.Data))) > seg.Address) {
		return true
	}
	return false
}

func userNumberInput(prompt string) uint64 {
	//Take user input in various formats
	return 0
}

func userConfirm(s string) bool {
	reader := bufio.NewReader(os.Stdin)

	for {
		fmt.Printf("%s [Y/n]: ", s)

		response, err := reader.ReadString('\n')
		if err != nil {
			log.Fatal(err)
		}

		response = strings.ToLower(strings.TrimSpace(response))
		if len(response) > 0 {
			if response[0] == 'y' {
				return true
			} else if response[0] == 'n' {
				return false
			}
		} else {
			return true
		}
	}
}
