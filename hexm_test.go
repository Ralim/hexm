package main

import (
	"crypto/rand"
	"fmt"
	"io/ioutil"
	"os"
	"os/exec"
	"reflect"
	"testing"
)

func TestMain(t *testing.T) {
	oldArgs := os.Args
	defer func() { os.Args = oldArgs }()
	//Run through simulated merge of a pair of files into another
	//Create a bin file, then convert that to a hex file for testing with
	length := 1024 * 128
	tmpfile, err := os.CreateTemp("", "*_testMainBin.bin")
	if err != nil {
		t.Fatal(err)
	}
	data := make([]byte, length)
	_, err = rand.Read(data)
	if err != nil {
		t.Fatal(err)
	}
	tmpfile.Write(data)
	tmpfile.Close()
	//Convert bin to hex
	outputName := tmpfile.Name() + ".hex"
	cmd := exec.Command("objcopy", "-I", "binary", "-O", "ihex", tmpfile.Name(), outputName)
	err = cmd.Run()
	if err != nil {
		t.Fatal(err)
	}
	defer os.Remove(tmpfile.Name())
	defer os.Remove(outputName)
	//We now have a bin, hex stacked in address space
	//Merge these together
	mergedOutput := outputName + "_merged.bin"
	os.Args = []string{"hexm", tmpfile.Name() + fmt.Sprintf(":0x%08X", length), outputName, mergedOutput}
	defer os.Remove(mergedOutput)
	main()
	//Read the merged file
	outputBytes, err := ioutil.ReadFile(mergedOutput)
	if err != nil {
		t.Fatal(err)
	}
	expectedBytes := append(data, data...)

	if !reflect.DeepEqual(outputBytes, expectedBytes) {
		t.Error("Failed to merge files seamlessly")
	}
}
