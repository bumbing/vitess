package main

import (
	"testing"
)

func TestPinschemaVindexDDLs(t *testing.T) {
	goldenTest(
		t,
		t.Name(),
		"testdata/patio.sql",
		"create-lookup-vindex",
		pinschemaConfig{},
	)
}

