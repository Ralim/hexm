package main

import (
	"bufio"
	"fmt"
	"log"
	"os"
	"path/filepath"
	"strings"

	"github.com/marcinbor85/gohex"
)

func main() {
	//As we only have trivial command line args, simpler to custom parse
	inputFiles, outputFile, err := parseArgs(os.Args)
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
		fmt.Printf("Loading file %d\r\n", i+1)
		file, err := os.Open(inputFilePath)
		if err != nil {
			panic(err)
		}
		defer file.Close()
		mem := gohex.NewMemory()
		err = mem.ParseIntelHex(file)
		if err != nil {
			panic(err)
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

	// bytes := mem.ToBinary(0xFFF0, 128, 0x00)
	// fmt.Printf("%v\n", bytes)
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

//Splits provided args into input files and the output file
func parseArgs(args []string) ([]string, string, error) {
	filePaths := args[1:]
	if len(filePaths) < 2 {
		return nil, "", fmt.Errorf("not enough files specified")
	}
	inputFiles := filePaths[:len(filePaths)-1]
	outputFiles := filePaths[len(filePaths)-1]
	return inputFiles, outputFiles, nil
}

func validateFiles(inputs []string, output string) error {
	for _, file := range inputs {
		if err := validateFile(file, true); err != nil {
			return err
		}
	}
	if err := validateFile(output, false); err != nil {
		return err
	}
	return nil
}
func validateFile(path string, shouldExist bool) error {
	extension := filepath.Ext(path)
	if !(extension == ".hex" || extension == ".bin") {
		return fmt.Errorf("invalid file format %s", path)
	}
	if _, err := os.Stat(path); err == nil {
		if shouldExist {
			return nil
		} else {
			//Prompt overwrite
			if userConfirm(fmt.Sprintf("Overwrite %s?", path)) {
				return nil
			} else {
				return fmt.Errorf("not overwriting %s", path)
			}
		}

	} else if os.IsNotExist(err) {
		if shouldExist {
			return fmt.Errorf("file does not exist %s", path)
		} else {
			return nil
		}
	} else {
		return fmt.Errorf("file %s raised IO error %v", path, err)
	}
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
