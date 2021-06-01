package main

import (
	"fmt"
	"os"
	"path/filepath"
	"strconv"
	"strings"
)

//Splits provided args into input files and the output file
func parseArgs(filePaths []string) ([]string, string, error) {
	if len(filePaths) < 2 {
		return []string{}, "", fmt.Errorf("not enough files specified")
	}
	inputFiles := filePaths[:len(filePaths)-1]
	outputFiles := filePaths[len(filePaths)-1]
	return inputFiles, outputFiles, nil
}

//validateFiles Validate inputs exist and output doesnt exist or confirm overwrite
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

// parseFileTypeAndStart returns if the path specifies a hex file or not, and if its a binary if it contains a starting address
// This parses a format of test.bin:0x5000 -> binary + start @ 0x5000
func parseFileTypeAndStart(path string) (isHexFile bool, binaryStart uint32, err error) {
	extension := filepath.Ext(path)
	if extension == ".hex" {
		return true, 0, nil
	}
	if extension == ".bin" {
		return false, 0, nil
	}
	parts := strings.Split(extension, ":")
	if len(parts) == 2 {
		if parts[0] == ".bin" {
			//We support previx notation of 0x and 0b, otherwise assumed decimal
			base := 10
			number := parts[1]
			if parts[1][0:2] == "0x" {
				base = 16
				number = parts[1][2:]
			} else if parts[1][0:2] == "0b" {
				base = 2
				number = parts[1][2:]
			}
			n, err := strconv.ParseUint(number, base, 32)
			if err == nil {
				return false, uint32(n), nil
			}
		}
	}
	return false, 0, fmt.Errorf("could not parse file type from %s", path)
}
func validateFile(path string, shouldExist bool) error {
	_, _, err := parseFileTypeAndStart(path)
	if err != nil {
		return fmt.Errorf("invalid file format %s => %v", path, err)
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
