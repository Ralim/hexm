package main

import (
	"fmt"
	"io/ioutil"
	"os"

	"github.com/marcinbor85/gohex"
)

func parseInputFile(path string) (*gohex.Memory, error) {
	mem := gohex.NewMemory()
	isHex, start, path, err := parseFileTypeAndStart(path)
	if err != nil {
		return mem, err
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
		err = mem.AddBinary(start, data)
		if err != nil {
			return mem, err
		}
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

func writeOutput(outputFile string, outputMemory *gohex.Memory) error {
	outputHex, binaryStart, outputFile, err := parseFileTypeAndStart(outputFile)
	if err != nil {
		return err
	}
	file, err := os.Create(outputFile)
	if err != nil {
		return err
	}
	defer file.Close()

	if outputHex {
		err = outputMemory.DumpIntelHex(file, 32)
		if err != nil {
			return err
		}
	} else {
		//We want to write a binary file starting at the specified location, and padding all gaps
		existingSegments := outputMemory.GetDataSegments()
		//Write out each section
		for i, section := range existingSegments {
			start := section.Address - uint32(binaryStart)
			fmt.Printf("Writing %v bytes @ %08X for section %d\r\n", len(section.Data), start, i+1)
			_, err = file.WriteAt(section.Data, int64(start))
			if err != nil {
				return err
			}
		}
	}
	return nil
}

func segmentOverlaps(seg gohex.DataSegment, seg2 gohex.DataSegment) bool {
	if ((seg2.Address >= seg.Address) && (seg2.Address < seg.Address+uint32(len(seg.Data)))) ||
		((seg2.Address < seg.Address) && (seg2.Address+uint32(len(seg2.Data))) > seg.Address) {
		return true
	}
	return false
}
