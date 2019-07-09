package main

import (
	"testing"
)

func TestPinschemaSequenceDDLs(t *testing.T) {
	goldenTest(
		t,
		t.Name(),
		"testdata/patio.sql",
		"create-seq",
		pinschemaConfig{},
	)
}

func TestPinschemaSequenceDDLs_Whitelist(t *testing.T) {
	goldenTest(
		t,
		t.Name(),
		"testdata/patio.sql",
		"create-seq",
		pinschemaConfig{
			sequenceTableWhitelist: []string{"campaigns", "accepted_tos"},
		},
	)
}
