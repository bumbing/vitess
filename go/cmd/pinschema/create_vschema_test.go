package main

import (
	"testing"
)

func TestCreateVSchema(t *testing.T) {
	type Tests struct {
		name   string
		config pinschemaConfig
	}

	tests := []Tests{
		{"Unsharded", pinschemaConfig{}},
		{"Authoritative",
			pinschemaConfig{
				colsAuthoritative: true,
				includeCols:       true,
			},
		},
		{"Seqs",
			pinschemaConfig{
				createSeq: true,
			},
		},
		{"SeqsWhitelist",
			pinschemaConfig{
				createSeq:              true,
				sequenceTableWhitelist: []string{"campaigns", "accepted_tos"},
			},
		},
		{"Primary",
			pinschemaConfig{
				createPrimary: true,
			},
		},
		{"PrimaryAndSecondary",
			pinschemaConfig{
				createPrimary:               true,
				createSecondary:             true,
				defaultScatterCacheCapacity: 10000,
				tableScatterCacheCapacity: map[string]uint64{
					"campaigns": 20000,
				},
			},
		},
		{"LookupVindexUnownedVindexWhitelist",
			pinschemaConfig{
				createPrimary:               true,
				createSecondary:             true,
				defaultScatterCacheCapacity: 10000,
				tableScatterCacheCapacity: map[string]uint64{
					"campaigns": 20000,
				},
				lookupVindexWriteOnly:        true,
				lookupVindexWhitelist:        []string{"ad_groups", "ad_group_specs"},
				unownedLookupVindexWhiteList: []string{"ad_group_id_idx"},
				createLookupVindexTables:     true,
			},
		},
		{"LookupVindexWhitelist",
			pinschemaConfig{
				createPrimary:               true,
				createSecondary:             true,
				defaultScatterCacheCapacity: 10000,
				tableScatterCacheCapacity: map[string]uint64{
					"campaigns": 20000,
				},
				lookupVindexWriteOnly:    true,
				lookupVindexWhitelist:    []string{"ad_groups", "ad_group_specs"},
				createLookupVindexTables: true,
			},
		},
		{"LookupVindex",
			pinschemaConfig{
				createPrimary:               true,
				createSecondary:             true,
				defaultScatterCacheCapacity: 10000,
				tableScatterCacheCapacity: map[string]uint64{
					"campaigns": 20000,
				},
				lookupVindexWriteOnly:    true,
				createLookupVindexTables: true,
			},
		},
	}

	for _, test := range tests {
		goldenTest(
			t,
			t.Name()+"_"+test.name,
			"testdata/ddls.sql",
			"create-vschema",
			test.config,
		)
	}
}
