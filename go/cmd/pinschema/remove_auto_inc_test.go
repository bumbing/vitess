package main

import (
	"testing"
)

func TestPinschemaRemoveAutoinc(t *testing.T) {
	goldenTest(
		t,
		t.Name(),
		"testdata/ddls.sql",
		"remove-autoinc",
		pinschemaConfig{},
	)
}
