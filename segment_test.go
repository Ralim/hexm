package main

import (
	"crypto/rand"
	"fmt"
	"io/ioutil"
	"os"
	"os/exec"
	"reflect"
	"testing"

	"github.com/marcinbor85/gohex"
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
	t.Parallel()
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
	t.Parallel()
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

func TestMergeSegments(t *testing.T) {
	t.Parallel()
	data := make([]byte, 2048)
	rand.Read(data)

	mem1 := gohex.NewMemory()
	mem2 := gohex.NewMemory()
	mem3 := gohex.NewMemory()
	mem1.AddBinary(0, data[0:1024])
	mem2.AddBinary(1024, data[1024:])
	mergeSegments(mem3, mem1, "")
	mergeSegments(mem3, mem2, "")
	if !reflect.DeepEqual(mem3.GetDataSegments()[0].Data, data) {
		t.Error("Merge should handle simple case")
	}
	//test order is ignored
	mem3 = gohex.NewMemory()
	mergeSegments(mem3, mem2, "")
	mergeSegments(mem3, mem1, "")
	if !reflect.DeepEqual(mem3.GetDataSegments()[0].Data, data) {
		t.Error("Merge should handle simple case")
	}

}

func TestWriteOutputBlobToBin(t *testing.T) {
	t.Parallel()
	data := make([]byte, 1024*256)
	rand.Read(data)
	// Persist this to a hex file
	tmpfile, err := os.CreateTemp("", "*_makebin.bin")
	if err != nil {
		t.Error(err)
	}
	tmpfile.Close()
	defer os.Remove(tmpfile.Name())
	//Write out to them
	mem := gohex.NewMemory()
	mem.AddBinary(0, data)
	writeOutput(tmpfile.Name(), mem) // will have written out a hex file now
	dataread, err := ioutil.ReadFile(tmpfile.Name())
	if err != nil {
		t.Error(err)
	}
	if !reflect.DeepEqual(dataread, data) {
		t.Error("Output hex should convert to flat bin")
	}
}

func TestWriteOutputBlobToHex(t *testing.T) {
	t.Parallel()
	//Test writing out bin and hex
	//By converting hex back to bin too
	data := make([]byte, 1024*256)
	rand.Read(data)
	// Persist this to a hex file
	tmpfile, err := os.CreateTemp("", "*_makehex.hex")
	if err != nil {
		t.Error(err)
	}
	tmpfile.Close()
	defer os.Remove(tmpfile.Name())
	//Write out to them
	mem := gohex.NewMemory()
	mem.AddBinary(0, data)
	writeOutput(tmpfile.Name(), mem) // will have written out a hex file now
	//Convert it to bin via trusted objcopy
	outputName := tmpfile.Name() + ".bin"
	cmd := exec.Command("objcopy", "-O", "binary", "-I", "ihex", tmpfile.Name(), outputName)
	err = cmd.Run()
	if err != nil {
		t.Error(err)
	}
	defer os.Remove(outputName)
	dataread, err := ioutil.ReadFile(outputName)
	if err != nil {
		t.Error(err)
	}
	if !reflect.DeepEqual(dataread, data) {
		t.Error("Output hex should convert to flat bin")
	}
}
