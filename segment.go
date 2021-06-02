package main

import (
	"fmt"
	"io/ioutil"
	"os"
	"strings"

	"github.com/marcinbor85/gohex"
)

func parseInputFile(path string) (*gohex.Memory, error) {
	mem := gohex.NewMemory()
	isHex, start, err := parseFileTypeAndStart(path)
	if err != nil {
		return mem, err
	}
	//have to remove colon denoted part
	parts := strings.Split(path, ":")
	if len(parts) > 1 {
		path = parts[0]
	}
	if isHex {

		file, err := os.Open(path)
		if err != nil {
			return mem, err
		}
		defer file.Close()
		err = mem.ParseIntelHex(file)
		if err != nil {
			return mem, err
		}
	} else {
		//This is a binary file, so we can just load it in
		data, err := ioutil.ReadFile(path)
		if err != nil {
			return mem, err
		}
		mem.AddBinary(start, data)
	}
	return mem, nil
}

func mergeSegments(base, addional *gohex.Memory, userPath string) {
	existingSegments := base.GetDataSegments()
	for x, segment := range addional.GetDataSegments() {
		fmt.Printf("Section %d @ 0x%08X ; len %d\n", x+1, segment.Address, len(segment.Data))
		//Check if this segment overlaps the existing segments
		if len(existingSegments) > 0 {
			for _, seg2 := range existingSegments {
				if !segmentOverlaps(segment, seg2) || userConfirmOverlap(segment, userPath) {
					//write this segment into it
					base.SetBinary(segment.Address, segment.Data)
				} else {
					fmt.Printf("Did not merge the segment @ %08X", segment.Address)
				}
			}
		} else {
			base.SetBinary(segment.Address, segment.Data)
		}
	}
}

func writeOutput(outputFile string, outputMemory *gohex.Memory) {
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

func segmentOverlaps(seg gohex.DataSegment, seg2 gohex.DataSegment) bool {
	if ((seg2.Address >= seg.Address) && (seg2.Address < seg.Address+uint32(len(seg.Data)))) ||
		((seg2.Address < seg.Address) && (seg2.Address+uint32(len(seg2.Data))) > seg.Address) {
		return true
	}
	return false
}
