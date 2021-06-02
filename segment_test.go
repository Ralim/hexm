package main

import (
	"crypto/rand"
	"fmt"
	"os"
	"os/exec"
	"reflect"
	"testing"
)

func createTestFilePair(t *testing.T, length int, baseAddress int) (hexFile, binFile string) {
	//Create a bin file, then convert that to a hex file for testing with
	tmpfile, err := os.CreateTemp("", "*_testBin.bin")
	if err != nil {
		t.Error(err)
	}
	data := make([]byte, length)
	rand.Read(data)
	tmpfile.Write(data)
	tmpfile.Close()
	//Convert bin to hex
	outputName := tmpfile.Name() + ".hex"
	cmd := exec.Command("objcopy", "--image-base", fmt.Sprintf("%d", baseAddress), "-I", "binary", "-O", "ihex", tmpfile.Name(), outputName)
	err = cmd.Run()
	if err != nil {
		t.Error(err)
	}
	return tmpfile.Name(), outputName
}

func TestParseInputFile(t *testing.T) {
	//create the test files
	hexFile, binFile := createTestFilePair(t, 1024*8, 0)
	defer os.Remove(hexFile)
	defer os.Remove(binFile)
	memhex, err := parseInputFile(hexFile)
	if err != nil {
		t.Error(err)
	}
	membin, err := parseInputFile(binFile)
	if err != nil {
		t.Error(err)
	}
	hexSegs := memhex.GetDataSegments()
	binSegs := membin.GetDataSegments()
	if len(hexSegs) != len(binSegs) {
		t.Errorf("Should return same number of segments")
	}
	for i, _ := range hexSegs {
		if !reflect.DeepEqual(hexSegs[i], binSegs[i]) {
			t.Errorf("Data segments differ")
		}
	}
}

func TestParseInputFileOffsets(t *testing.T) {
	//create the test files

	var tests = []struct {
		offset  int
		offsetS string
		size    int
	}{
		{0, ":0", 1024},
		{1024, ":1024", 1024},
		{4096, ":4096", 4096},
	}

	for _, tt := range tests {

		testname := fmt.Sprintf("%v-%v", tt.offset, tt.size)
		t.Run(testname, func(t *testing.T) {
			hexFile, binFile := createTestFilePair(t, 1024*8, 0)
			defer os.Remove(hexFile)
			defer os.Remove(binFile)
			memhex, err := parseInputFile(hexFile)
			if err != nil {
				t.Error(err)
			}
			membin, err := parseInputFile(binFile + tt.offsetS)
			if err != nil {
				t.Error(err)
			}
			hexSegs := memhex.GetDataSegments()
			binSegs := membin.GetDataSegments()
			if len(hexSegs) != len(binSegs) {
				t.Errorf("Should return same number of segments")
			}
			for i, _ := range hexSegs {
				if !reflect.DeepEqual(hexSegs[i], binSegs[i]) {
					t.Errorf("Data segments differ")
				}
			}
		})
	}
}
