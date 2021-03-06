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
			data := section.Data
			start := section.Address - uint32(binaryStart)
			if section.Address < binaryStart {
				offset := int(binaryStart) - int(section.Address)
				data = data[offset:]
				start = 0 // As have no need to pad
			}
			err = checkFileStartPos(file, start)
			if err != nil {
				return err
			}
			fmt.Printf("Writing %v bytes @ %08X for section %d\r\n", len(data), start, i+1)
			_, err = file.WriteAt(data, int64(start))
			if err != nil {
				return err
			}
		}
	}
	return nil
}

//checkFileStartPos Asks user to confirm if we are about to pad more than 128MB of output space
func checkFileStartPos(file *os.File, start uint32) error {
	filestat, err := file.Stat()
	if err != nil {
		return nil
	}
	currentFileSize := uint32(filestat.Size())
	if start > currentFileSize {
		padMBytes := start - currentFileSize
		padMBytes /= 1024 * 1024
		if padMBytes > 128 {
			if !userConfirm(fmt.Sprintf("Output file will contain at least %d Mbytes of padding, are you sure?", padMBytes)) {
				return fmt.Errorf("user aborted write due to padding of %v Mbytes", padMBytes)
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
