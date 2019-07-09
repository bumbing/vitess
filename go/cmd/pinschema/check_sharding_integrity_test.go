package main

import (
	"testing"
)

func TestShardingIntegrity(t *testing.T) {
	type Tests struct {
		name   string
		config pinschemaConfig
	}

	tests := []Tests{
		{"Simple", pinschemaConfig{}},
		{"Prefix", pinschemaConfig{
			queryTablePrefix: "db_",
		}},
		{"Summarize", pinschemaConfig{
			summarize: true,
		}},
		{"TableLimit", pinschemaConfig{
			tableResultLimit: 20,
		}},
	}

	for _, test := range tests {
		goldenTest(
			t,
			t.Name()+"_"+test.name,
			"testdata/patio.sql",
			"check-sharding-integrity",
			test.config,
		)
	}
}
