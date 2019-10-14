package main

import (
	"io/ioutil"
	"strings"
	"testing"
)

func TestValidateVschema(t *testing.T) {

	ddls, err := readAndParseSchema("testdata/patio.sql")
	if err != nil {
		t.Fatalf("Failed to parse: %v", err)
	}

	moreddls, err := readAndParseSchema("testdata/patiogeneral.sql")
	if err != nil {
		t.Fatalf("Failed to parse: %v", err)
	}
	for _, more := range moreddls {
		ddls = append(ddls, more)
	}

	vschema, err := ioutil.ReadFile("testdata/TestValidateVSchema_Prod.json")
	if err != nil {
		t.Fatalf("Failed to read vschema: %v", err)
	}

	config := pinschemaConfig{
		validateKeyspace:         "patio",
		validateShards:           2,
		validateVschema:          string(vschema),
	}

	md5, err := validateVschema(ddls, config)
	if err != nil {
		t.Fatalf("Failed to validate: %v", err)
	}
	if len(md5) == 0 {
		t.Fatal("Failed to generate a md5 of vschema")
	}
}

func TestValidateVschemaVSchemaNegatives(t *testing.T) {

	ddls, err := readAndParseSchema("testdata/validate_ddls.sql")
	if err != nil {
		t.Fatalf("Failed to parse: %v", err)
	}

	vschemas := map[string]string{
		"testdata/TestValidateVSchema_Empty.json":              "no keyspace",
		"testdata/TestValidateVSchema_MissingVindex.json":      "table has column vindex:",
		"testdata/TestValidateVSchema_NoFunctionalVindex.json": "has no functional vindex",
	}
	for vschemaPath, errExpected := range vschemas {
		vschema, err := ioutil.ReadFile(vschemaPath)
		config := pinschemaConfig{
			validateKeyspace:         "patio",
			validateShards:           2,
			validateVschema:          string(vschema),
		}

		if err != nil {
			t.Fatalf("Failed to read vschema: %v", err)
		}

		_, err = validateVschema(ddls, config)
		if err == nil {
			t.Fatalf("expected error didn't happen for: %s", vschemaPath)
		}
		if !strings.HasPrefix(err.Error(), errExpected) && !strings.HasSuffix(err.Error(), errExpected) {
			t.Fatalf("expected error didn't match: %s, instead got: %s", errExpected, err.Error())
		}
	}
}
