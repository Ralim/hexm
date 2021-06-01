package main

import (
	"fmt"
	"os"
	"testing"
)

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
	err = validateFiles([]string{file_exists_bin.Name(), file_exists_hex.Name()}, "nope.bin")
	if err != nil {
		t.Error(err)
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
