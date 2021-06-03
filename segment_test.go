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
		t.Fatal(err)
	}
	data := make([]byte, length)
	_, err = rand.Read(data)
	if err != nil {
		t.Fatal(err)
	}
	_, err = tmpfile.Write(data)
	if err != nil {
		t.Fatal(err)
	}

	tmpfile.Close()
	//Convert bin to hex
	outputName := tmpfile.Name() + ".hex"
	cmd := exec.Command("objcopy", "--image-base", fmt.Sprintf("%d", baseAddress), "-I", "binary", "-O", "ihex", tmpfile.Name(), outputName)
	err = cmd.Run()
	if err != nil {
		t.Fatal(err)
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
		t.Fatal(err)
	}
	membin, err := parseInputFile(binFile)
	if err != nil {
		t.Fatal(err)
	}
	hexSegs := memhex.GetDataSegments()
	binSegs := membin.GetDataSegments()
	if len(hexSegs) != len(binSegs) {
		t.Fatalf("Should return same number of segments")
	}
	for i := range hexSegs {
		if !reflect.DeepEqual(hexSegs[i], binSegs[i]) {
			t.Fatalf("Data segments differ")
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
				t.Fatal(err)
			}
			membin, err := parseInputFile(binFile + tt.offsetS)
			if err != nil {
				t.Fatal(err)
			}
			hexSegs := memhex.GetDataSegments()
			binSegs := membin.GetDataSegments()
			if len(hexSegs) != len(binSegs) {
				t.Fatalf("Should return same number of segments")
			}
			for i := range hexSegs {
				if !reflect.DeepEqual(hexSegs[i], binSegs[i]) {
					t.Fatalf("Data segments differ")
				}
			}
		})
	}
}

func TestMergeSegments(t *testing.T) {
	t.Parallel()
	data := make([]byte, 2048)
	_, err := rand.Read(data)
	if err != nil {
		t.Fatal(err)
	}

	mem1 := gohex.NewMemory()
	mem2 := gohex.NewMemory()
	mem3 := gohex.NewMemory()
	err = mem1.AddBinary(0, data[0:1024])
	if err != nil {
		t.Fatal(err)
	}
	err = mem2.AddBinary(1024, data[1024:])
	if err != nil {
		t.Fatal(err)
	}
	mergeSegments(mem3, mem1, "")
	mergeSegments(mem3, mem2, "")
	if !reflect.DeepEqual(mem3.GetDataSegments()[0].Data, data) {
		t.Fatal("Merge should handle simple case")
	}
	//test order is ignored
	mem3 = gohex.NewMemory()
	mergeSegments(mem3, mem2, "")
	mergeSegments(mem3, mem1, "")
	if !reflect.DeepEqual(mem3.GetDataSegments()[0].Data, data) {
		t.Fatal("Merge should handle simple case")
	}

}

func TestMergeSegmentsOverlap(t *testing.T) {

	tmpfile, err := os.CreateTemp("", "mockstdin")
	if err != nil {
		t.Fatal(err)
	}
	defer os.Remove(tmpfile.Name()) // clean up
	content := []byte("y\r\n")
	if _, err := tmpfile.Write(content); err != nil {
		t.Fatal(err)
	}

	if _, err := tmpfile.Seek(0, 0); err != nil {
		t.Fatal(err)
	}

	oldStdin := os.Stdin
	defer func() { os.Stdin = oldStdin }() // Restore original Stdin at end of test
	os.Stdin = tmpfile

	data := make([]byte, 2048)
	_, err = rand.Read(data)
	if err != nil {
		t.Fatal(err)
	}

	mem1 := gohex.NewMemory()
	mem2 := gohex.NewMemory()
	mem3 := gohex.NewMemory()
	err = mem1.AddBinary(0, data[0:1024])
	if err != nil {
		t.Fatal(err)
	}
	err = mem2.AddBinary(1024, data[1024:])
	if err != nil {
		t.Fatal(err)
	}
	mergeSegments(mem3, mem1, "")
	mergeSegments(mem3, mem2, "")
	if !reflect.DeepEqual(mem3.GetDataSegments()[0].Data, data) {
		t.Fatal("Merge should handle simple case")
	}
	//run again and should overwrite
	mergeSegments(mem3, mem1, "")
	if !reflect.DeepEqual(mem3.GetDataSegments()[0].Data, data) {
		t.Fatal("Merge should handle simple case")
	}
	//Now test that it respects saying no to overwrite
	content = []byte("n\r\n")
	if _, err := tmpfile.Seek(0, 0); err != nil {
		t.Fatal(err)
	}
	if _, err := tmpfile.Write(content); err != nil {
		t.Fatal(err)
	}

	if _, err := tmpfile.Seek(0, 0); err != nil {
		t.Fatal(err)
	}
	mem1 = gohex.NewMemory()
	err = mem1.AddBinary(0, data[256:])
	if err != nil {
		t.Fatal(err)
	}
	mergeSegments(mem3, mem1, "")
	if !reflect.DeepEqual(mem3.GetDataSegments()[0].Data, data) {
		t.Fatal("Merge should reject overwrite i user opts out")
	}

}

func TestWriteOutputBlobToBin(t *testing.T) {
	t.Parallel()
	data := make([]byte, 1024*256)
	_, err := rand.Read(data)
	if err != nil {
		t.Fatal(err)
	}
	// Persist this to a hex file
	tmpfile, err := os.CreateTemp("", "*_makebin.bin")
	if err != nil {
		t.Fatal(err)
	}
	tmpfile.Close()
	defer os.Remove(tmpfile.Name())
	//Write out to them
	mem := gohex.NewMemory()
	err = mem.AddBinary(0, data)
	if err != nil {
		t.Fatal(err)
	}
	err = writeOutput(tmpfile.Name(), mem) // will have written out a hex file now
	if err != nil {
		t.Fatal(err)
	}
	dataread, err := ioutil.ReadFile(tmpfile.Name())
	if err != nil {
		t.Fatal(err)
	}
	if !reflect.DeepEqual(dataread, data) {
		t.Fatal("Output hex should convert to flat bin")
	}
}

func TestWriteOutputFails(t *testing.T) {
	t.Parallel()
	mem := gohex.NewMemory()
	err := writeOutput("badname.bad", mem) // will have written out a hex file now
	if err == nil {
		t.Fatal("Should raise error on bad name format")
	}
	err = writeOutput("/badfolder/test.hex", mem) // will have written out a hex file now
	if err == nil {
		t.Fatal("Should raise error on uncreatable file")
	}

}

func TestWriteOutputBlobToHex(t *testing.T) {
	t.Parallel()
	//Test writing out bin and hex
	//By converting hex back to bin too
	data := make([]byte, 1024*256)
	_, err := rand.Read(data)
	if err != nil {
		t.Fatal(err)
	}
	// Persist this to a hex file
	tmpfile, err := os.CreateTemp("", "*_makehex.hex")
	if err != nil {
		t.Fatal(err)
	}
	tmpfile.Close()
	defer os.Remove(tmpfile.Name())
	//Write out to them
	mem := gohex.NewMemory()
	err = mem.AddBinary(0, data)
	if err != nil {
		t.Fatal(err)
	}
	err = writeOutput(tmpfile.Name(), mem) // will have written out a hex file now
	if err != nil {
		t.Fatal(err)
	}

	//Convert it to bin via trusted objcopy
	outputName := tmpfile.Name() + ".bin"
	cmd := exec.Command("objcopy", "-O", "binary", "-I", "ihex", tmpfile.Name(), outputName)
	err = cmd.Run()
	if err != nil {
		t.Fatal(err)
	}
	defer os.Remove(outputName)
	dataread, err := ioutil.ReadFile(outputName)
	if err != nil {
		t.Fatal(err)
	}
	if !reflect.DeepEqual(dataread, data) {
		t.Fatal("Output hex should convert to flat bin")
	}
}

func TestWriteOutputBlobToBinOffset(t *testing.T) {
	t.Parallel()
	data := make([]byte, 1024*256)
	offset := uint32(1024)
	_, err := rand.Read(data)
	if err != nil {
		t.Fatal(err)
	}
	// Persist this to a hex file
	tmpfile, err := os.CreateTemp("", "*_makebin.bin")
	if err != nil {
		t.Fatal(err)
	}
	tmpfile.Close()
	defer os.Remove(tmpfile.Name())
	//Write out to them
	mem := gohex.NewMemory()
	err = mem.AddBinary(offset, data)
	if err != nil {
		t.Fatal(err)
	}
	err = writeOutput(tmpfile.Name()+fmt.Sprintf(":%d", offset), mem) // will have written out a hex file now
	if err != nil {
		t.Fatal(err)
	}
	dataread, err := ioutil.ReadFile(tmpfile.Name())
	if err != nil {
		t.Fatal(err)
	}
	if !reflect.DeepEqual(dataread, data) {
		t.Fatal("Matching offsets should have no filler")
	}
}

func TestWriteOutputBlobToBinOffsetPadding(t *testing.T) {
	t.Parallel()
	data := make([]byte, 1024*256)
	offset := uint32(1024)
	_, err := rand.Read(data)
	if err != nil {
		t.Fatal(err)
	}
	// Persist this to a bin file
	tmpfile, err := os.CreateTemp("", "*_makebin.bin")
	if err != nil {
		t.Fatal(err)
	}
	tmpfile.Close()
	defer os.Remove(tmpfile.Name())
	//Write out to them
	mem := gohex.NewMemory()
	err = mem.AddBinary(0, data)
	if err != nil {
		t.Fatal(err)
	}
	err = writeOutput(tmpfile.Name()+fmt.Sprintf(":%d", offset), mem) // will have written out a hex file now
	if err != nil {
		t.Fatal(err)
	}
	dataread, err := ioutil.ReadFile(tmpfile.Name())
	if err != nil {
		t.Fatal(err)
	}
	if !reflect.DeepEqual(dataread, data[offset:]) {
		t.Fatal("Rebase should truncate off leading data")
	}
}
