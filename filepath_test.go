package main

import (
	"fmt"
	"log"
	"os"
	"reflect"
	"testing"
)

func TestParseArgs(t *testing.T) {
	//Simpler list splitter for now

	var tests = []struct {
		args    []string
		inputs  []string
		output  string
		wantErr error
	}{
		{[]string{"1.hex", "2.hex", "3.hex"}, []string{"1.hex", "2.hex"}, "3.hex", nil},
		{[]string{"1.hex"}, []string{}, "", fmt.Errorf("not enough files specified")},
	}

	for _, tt := range tests {

		testname := fmt.Sprintf("%v", tt.args)
		t.Run(testname, func(t *testing.T) {
			inputs, output, err := parseArgs(tt.args)
			if !reflect.DeepEqual(inputs, tt.inputs) {
				t.Errorf("got %v, want %v", inputs, tt.inputs)
			}
			if output != tt.output {
				t.Errorf("got %v, want %v", output, tt.output)
			}
			if err != tt.wantErr {
				if err != nil && tt.wantErr != nil {
					if err.Error() != tt.wantErr.Error() {
						t.Errorf("got %v, want %v", err, tt.wantErr)
					}
				} else {
					t.Errorf("got %v, want %v", err, tt.wantErr)
				}
			}
		})
	}

}
func TestValidateFiles(t *testing.T) {
	//Testing input file that exists, one that doesnt
	//Output file that does, output file that doesnt
	file_exists_hex, err := os.CreateTemp("", "test_*_.hex")
	if err != nil {
		t.Error(err)
	}
	file_exists_hex.Close()
	defer os.Remove(file_exists_hex.Name())
	file_exists_bin, err := os.CreateTemp("", "test_*_.bin")
	if err != nil {
		t.Error(err)
	}
	file_exists_bin.Close()
	defer os.Remove(file_exists_bin.Name())
	file_exists_bad, err := os.CreateTemp("", "test_*_.bad")
	if err != nil {
		t.Error(err)
	}
	file_exists_bad.Close()
	defer os.Remove(file_exists_bad.Name())

	//Basic case, both files exist and should pass
	err = validateFiles([]string{file_exists_bin.Name(), file_exists_hex.Name()}, "nope.bin")
	if err != nil {
		t.Error(err)
	}
	//Test non existing input file
	err = validateFiles([]string{file_exists_bin.Name(), file_exists_hex.Name(), "nothere.bin"}, "nope.bin")
	if err == nil {
		t.Errorf("Should raise error on input file that doesnt exist")
	}
	err = validateFiles([]string{file_exists_bin.Name(), file_exists_hex.Name(), file_exists_bad.Name()}, "nope.bin")
	if err == nil {
		t.Errorf("Should raise error on bad file name even if it exists")
	}
	//Testing bad output files
	err = validateFiles([]string{file_exists_bin.Name(), file_exists_hex.Name()}, "nope.lol")
	if err == nil {
		t.Errorf("Should raise error on output file of unknown type")
	}
	//Testing output file exists case, should prompt asking for confirmation of overwrite
	tmpfile, err := os.CreateTemp("", "mockstdin")
	if err != nil {
		t.Error(err)
	}
	defer os.Remove(tmpfile.Name()) // clean up
	content := []byte("y\r\n")
	if _, err := tmpfile.Write(content); err != nil {
		log.Fatal(err)
	}

	if _, err := tmpfile.Seek(0, 0); err != nil {
		log.Fatal(err)
	}

	oldStdin := os.Stdin
	defer func() { os.Stdin = oldStdin }() // Restore original Stdin at end of test
	os.Stdin = tmpfile
	err = validateFiles([]string{file_exists_hex.Name()}, file_exists_bin.Name())
	if err != nil {
		t.Errorf("Should allow user to confirm overwrite")
	}
	content = []byte("n\r\n")
	if _, err := tmpfile.Seek(0, 0); err != nil {
		log.Fatal(err)
	}

	if _, err := tmpfile.Write(content); err != nil {
		log.Fatal(err)
	}

	if _, err := tmpfile.Seek(0, 0); err != nil {
		log.Fatal(err)
	}

	err = validateFiles([]string{file_exists_hex.Name()}, file_exists_bin.Name())
	if err == nil {
		t.Errorf("Should raise error if user does not acknowledge overwrite")
	}
	if err := tmpfile.Close(); err != nil {
		log.Fatal(err)
	}
}

func TestParseFileTypeAndStart(t *testing.T) {
	//testing that it handles basic bin and hex files correctly

	var tests = []struct {
		path    string
		wantHex bool
		wantN   uint32
		wantErr error
	}{
		{"test.hex", true, 0, nil},
		{"test.bin", false, 0, nil},
		{"test.bin:1024", false, 1024, nil},
		{"test.bin:0x1024", false, 0x1024, nil},
		{"test.bin:0b111", false, 7, nil},
		{"test.bin:0b1011", false, 11, nil},
		{"test.bad:0b1011", false, 0, fmt.Errorf("could not parse file type from test.bad:0b1011")},
		{"test.bad:1024", false, 0, fmt.Errorf("could not parse file type from test.bad:1024")},
		{"test.bad:0x1024", false, 0, fmt.Errorf("could not parse file type from test.bad:0x1024")},
		{"test.bad", false, 0, fmt.Errorf("could not parse file type from test.bad")},
	}

	for _, tt := range tests {

		testname := tt.path
		t.Run(testname, func(t *testing.T) {
			isHex, n, err := parseFileTypeAndStart(tt.path)
			if isHex != tt.wantHex {
				t.Errorf("got %v, want %v", isHex, tt.wantHex)
			}
			if n != tt.wantN {
				t.Errorf("got %v, want %v", n, tt.wantN)
			}
			if err != tt.wantErr {
				if err != nil && tt.wantErr != nil {
					if err.Error() != tt.wantErr.Error() {
						t.Errorf("got %v, want %v", err, tt.wantErr)
					}
				} else {
					t.Errorf("got %v, want %v", err, tt.wantErr)
				}
			}
		})
	}

}
