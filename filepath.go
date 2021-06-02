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

func parseNumberString(data string) (uint32, error) {
	//Parse the prefix of 0x,0b or none
	if len(data) == 0 {
		return 0, fmt.Errorf("no Input")
	}
	base := 10
	number := data
	if len(data) > 1 {
		if data[0:2] == "0x" {
			base = 16
			number = data[2:]
		} else if data[0:2] == "0b" {
			base = 2
			number = data[2:]
		}
	}
	n, err := strconv.ParseUint(number, base, 32)
	return uint32(n), err
}

// parseFileTypeAndStart returns if the path specifies a hex file or not, and if its a binary if it contains a starting address
// This parses a format of test.bin:0x5000 -> binary + start @ 0x5000
func parseFileTypeAndStart(path string) (isHexFile bool, binaryStart uint32, err error) {
	parts := strings.Split(path, ":")
	baseName := path
	if len(parts) == 2 {
		baseName = parts[0]
	}

	extension := filepath.Ext(baseName)
	if extension == ".hex" {
		return true, 0, nil
	}
	if extension == ".bin" && len(parts) == 1 {
		return false, 0, nil
	}
	if len(parts) == 2 {
		if extension == ".bin" {
			n, err := parseNumberString(parts[1])
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
