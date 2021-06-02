package main

import (
	"log"
	"os"
	"testing"

	"github.com/marcinbor85/gohex"
)

func TestUserConfirmOverlap(t *testing.T) {
	//basic test as its just a wrapper
	seg := gohex.DataSegment{Address: 100, Data: []byte{1, 2, 3}}
	tmpfile, err := os.CreateTemp("", "mockstdin")
	if err != nil {
		t.Error(err)
	}
	defer os.Remove(tmpfile.Name()) // clean up
	oldStdin := os.Stdin
	defer func() { os.Stdin = oldStdin }() // Restore original Stdin at end of test
	os.Stdin = tmpfile

	//Run bulk tests

	var tests = []struct {
		user   string
		result bool
	}{
		{"y", true},
		{"yes", true},
		{"n", false},
		{"no", false},
	}

	for _, tt := range tests {

		t.Run(tt.user, func(t *testing.T) {

			content := []byte(tt.user + "\r\n")
			if _, err := tmpfile.Seek(0, 0); err != nil {
				log.Fatal(err)
			}
			if _, err := tmpfile.Write(content); err != nil {
				log.Fatal(err)
			}
			if _, err := tmpfile.Seek(0, 0); err != nil {
				log.Fatal(err)
			}
			confirmed := userConfirmOverlap(seg, "FILENAME")
			if confirmed != tt.result {
				t.Errorf("Should handle user typing %v", tt.user)
			}

		})
	}

	if err := tmpfile.Close(); err != nil {
		log.Fatal(err)
	}
}

func TestUserConfirm(t *testing.T) {
	//Testing parsing y/n
	tmpfile, err := os.CreateTemp("", "mockstdin")
	if err != nil {
		t.Error(err)
	}
	defer os.Remove(tmpfile.Name()) // clean up
	oldStdin := os.Stdin
	defer func() { os.Stdin = oldStdin }() // Restore original Stdin at end of test
	os.Stdin = tmpfile

	//Run bulk tests

	var tests = []struct {
		user   string
		result bool
	}{
		{"y", true},
		{"yes", true},
		{"n", false},
		{"no", false},
	}

	for _, tt := range tests {

		t.Run(tt.user, func(t *testing.T) {

			content := []byte(tt.user + "\r\n")
			if _, err := tmpfile.Seek(0, 0); err != nil {
				log.Fatal(err)
			}
			if _, err := tmpfile.Write(content); err != nil {
				log.Fatal(err)
			}
			if _, err := tmpfile.Seek(0, 0); err != nil {
				log.Fatal(err)
			}
			confirmed := userConfirm("-")
			if confirmed != tt.result {
				t.Errorf("Should handle user typing %v", tt.user)
			}

		})
	}

	if err := tmpfile.Close(); err != nil {
		log.Fatal(err)
	}
}
