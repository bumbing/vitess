package main

import (
	"testing"
)

func TestPinschemaSequenceDDLs(t *testing.T) {
	goldenTest(
		t,
		t.Name(),
		"testdata/ddls.sql",
		"create-seq",
		pinschemaConfig{},
	)
}
