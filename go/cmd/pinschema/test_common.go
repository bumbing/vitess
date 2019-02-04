package main

import (
	"flag"
	"fmt"
	"io/ioutil"
	"os"
	"testing"
)

var updateGoldens = flag.Bool("update", false, "update goldens instead of validating them")

func goldenTest(t *testing.T, testName string, ddlsFile string, command string, config pinschemaConfig) {
	goldenFile := fmt.Sprintf("testdata/%s.golden", testName)
	ddls, err := readAndParseSchema(ddlsFile)
	if err != nil {
		t.Errorf("Failed to parse %s: %v", ddlsFile, err)
		return
	}
	output, err := commands[command](ddls, config)
	if err != nil {
		t.Errorf("Unexpected error %v", err)
		return
	}

	if *updateGoldens {
		ioutil.WriteFile(goldenFile, []byte(output), os.ModePerm)
	} else {
		goldenOutput, err := ioutil.ReadFile(goldenFile)
		if err != nil {
			t.Error(err)
			return
		}

		goldenString := string(goldenOutput)
		if goldenString != output {
			t.Errorf("Found differences.\nGot: %v\nWant: %v", output, goldenString)
		}
	}
}
