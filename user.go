package main

import (
	"bufio"
	"fmt"
	"log"
	"os"
	"strings"

	"github.com/marcinbor85/gohex"
)

func userConfirmOverlap(seg gohex.DataSegment, source string) bool {
	return userConfirm(fmt.Sprintf("Merging segment @ 0x%08X from file %v will overwrite existing data, continue ?", seg.Address, source))
}

func userNumberInput(prompt string, defaultValue uint32) uint32 {
	//Take user input in various formats and parse
	//
	reader := bufio.NewReader(os.Stdin)

	for {
		fmt.Printf("%s [%08X]: ", prompt, defaultValue)

		response, err := reader.ReadString('\n')
		if err != nil {
			log.Fatal(err)
		}

		response = strings.ToLower(strings.TrimSpace(response))
		if len(response) > 0 {
			n, err := parseNumberString(response)
			if err == nil {
				return n
			}
		} else {
			return defaultValue
		}
	}
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
