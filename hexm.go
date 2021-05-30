package main

import (
	"bufio"
	"fmt"
	"log"
	"os"
	"path/filepath"
	"strings"
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
