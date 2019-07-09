package main

import (
	"testing"
)

func TestPinschemaRemoveAutoinc(t *testing.T) {
	goldenTest(
		t,
		t.Name(),
		"testdata/patio.sql",
		"remove-autoinc",
		pinschemaConfig{},
	)
}
