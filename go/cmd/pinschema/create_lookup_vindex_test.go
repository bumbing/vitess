package main

import (
	"testing"
)

func TestPinschemaVindexDDLs(t *testing.T) {
	goldenTest(
		t,
		t.Name(),
		"testdata/ddls.sql",
		"create-lookup-vindex",
		pinschemaConfig{},
	)
}

func TestPinschemaVindexDDLs_Whitelist(t *testing.T) {
	goldenTest(
		t,
		t.Name(),
		"testdata/ddls.sql",
		"create-lookup-vindex",
		pinschemaConfig{
			lookupVindexWhitelist: []string{"campaigns"},
		},
	)
}
