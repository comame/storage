package sfs

import (
	"strings"
	"testing"
)

func Test_hashFile(t *testing.T) {
	r := strings.NewReader("file contents")
	got, err := hashFile(r)
	if err != nil {
		t.Error(err)
	}

	expected := "7bb6f9f7a47a63e684925af3608c059edcc371eb81188c48c9714896fb1091fd"
	if got != expected {
		t.Errorf("Expect: %s, Got: %s", expected, got)
	}
}

func Test_filenamePrefix(t *testing.T) {
	content := "Hello, world!"
	h, err := hashFile(strings.NewReader(content))
	if err != nil {
		t.Error(err)
	}

	first, second, err := filenamePrefix(h)
	if err != nil {
		t.Error(err)
	}
	if !(first == "31" && second == "5f") {
		t.Errorf("Expect: 31,5f, Got: %s,%s", first, second)
	}
}
