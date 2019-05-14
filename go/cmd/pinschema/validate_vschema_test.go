package main

import (
	"io/ioutil"
	"testing"
)

func TestValidateVschema(t *testing.T) {

	ddls, err := readAndParseSchema("testdata/ddls.sql")
	if err != nil {
		t.Fatalf("Failed to parse: %v", err)
	}

	vschema, err := ioutil.ReadFile("testdata/TestValidateVSchema_Prod.golden")
	if err != nil {
		t.Fatalf("Failed to read vschema: %v", err)
	}

	config := pinschemaConfig{
		validateKeyspace: "patio",
		validateShards:   2,
		validateVschema:  string(vschema),
	}

	md5, err := validateVschema(ddls, config)
	if err != nil {
		t.Fatalf("Failed to validate: %v", err)
	}
	if len(md5) == 0 {
		t.Fatal("Failed to generate a md5 of vschema")
	}
}
