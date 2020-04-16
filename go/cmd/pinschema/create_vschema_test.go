package main

import (
	"testing"
)

func TestCreateVSchema(t *testing.T) {
	type Tests struct {
		name   string
		ddls   string
		config pinschemaConfig
	}

	tests := []Tests{
		{"Unsharded", "testdata/patio.sql", pinschemaConfig{}},
		{"UnshardedGeneral", "testdata/patiogeneral.sql", pinschemaConfig{}},
		{"Authoritative", "testdata/patio.sql",
			pinschemaConfig{
				colsAuthoritative: true,
				includeCols:       true,
			},
		},
		{"Seqs", "testdata/patio.sql",
			pinschemaConfig{
				createSeq: true,
			},
		},
		{"SeqsWhitelist", "testdata/patio.sql",
			pinschemaConfig{
				createSeq:              true,
				sequenceTableWhitelist: []string{"campaigns", "accepted_tos"},
			},
		},
		{"Primary", "testdata/patio.sql",
			pinschemaConfig{
				createPrimary: true,
			},
		},
		{"PrimaryAndSecondary", "testdata/patio.sql",
			pinschemaConfig{
				createPrimary:               true,
				createSecondary:             true,
				defaultScatterCacheCapacity: 10000,
				tableScatterCacheCapacity: map[string]uint64{
					"campaigns": 20000,
				},
			},
		},
	}

	for _, test := range tests {
		goldenTest(
			t,
			t.Name()+"_"+test.name,
			test.ddls,
			"create-vschema",
			test.config,
		)
	}
}