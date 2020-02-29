package engine

import (
	"testing"

	"vitess.io/vitess/go/sqltypes"
	querypb "vitess.io/vitess/go/vt/proto/query"
	vschemapb "vitess.io/vitess/go/vt/proto/vschema"
	"vitess.io/vitess/go/vt/vtgate/vindexes"
)

// Pinterest test for unowned Vindex update.
// There are two Column Vindexes. The table is the owner of "ownvindex", while not the
// owner of "unownvindex".
// This test is to verify the owned LookupHashUnique is able to be handled, while the
// unowned one is skipped.
func TestUpdateEqualChangedUnownedVindex(t *testing.T) {
	invschema := &vschemapb.SrvVSchema{
		Keyspaces: map[string]*vschemapb.Keyspace{
			"sharded": {
				Sharded: true,
				Vindexes: map[string]*vschemapb.Vindex{
					"hash": {
						Type: "hash",
					},
					"ownvindex": {
						Type: "pin_lookup_hash_unique",
						Params: map[string]string{
							"table": "lkp2",
							"from":  "from1",
							"to":    "toc",
						},
						Owner: "t1",
					},
					"unownvindex": {
						Type: "pin_lookup_hash_unique",
						Params: map[string]string{
							"table": "lkp1",
							"from":  "from",
							"to":    "toc",
						},
						Owner: "t2",
					},
				},
				Tables: map[string]*vschemapb.Table{
					"t1": {
						ColumnVindexes: []*vschemapb.ColumnVindex{{
							Name:    "hash",
							Columns: []string{"id"},
						}, {
							Name:    "ownvindex",
							Columns: []string{"c1"},
						}, {
							Name:    "unownvindex",
							Columns: []string{"c2"},
						}},
					},
				},
			},
		},
	}
	vs, err := vindexes.BuildVSchema(invschema)
	if err != nil {
		panic(err)
	}
	ks := vs.Keyspaces["sharded"]
	upd := &Update{
		Opcode:   UpdateEqual,
		Keyspace: ks.Keyspace,
		Query:    "dummy_update",
		Vindex:   ks.Vindexes["hash"].(vindexes.SingleColumn),
		Values:   []sqltypes.PlanValue{{Value: sqltypes.NewInt64(1)}},
		ChangedVindexValues: map[string][]sqltypes.PlanValue{
			"ownvindex": {{
				Value: sqltypes.NewInt64(1),
			}},
			"unownvindex": {{
				Value: sqltypes.NewInt64(2),
			}},
		},
		Table:            ks.Tables["t1"],
		OwnedVindexQuery: "dummy_subquery",
	}

	results := []*sqltypes.Result{sqltypes.MakeTestResult(
		sqltypes.MakeTestFields(
			"c1|c2",
			"int64|int64",
		),
		"4|5",
	)}
	vc := &loggingVCursor{
		shards:  []string{"-20", "20-"},
		results: results,
	}

	_, err = upd.Execute(vc, map[string]*querypb.BindVariable{}, false)
	if err != nil {
		t.Fatal(err)
	}
	vc.ExpectLog(t, []string{
		`ResolveDestinations sharded [] Destinations:DestinationKeyspaceID(166b40b44aba4bd6)`,
		// ResolveDestinations is hard-coded to return -20.
		// It gets used to perform the subquery to fetch the changing column values.
		`ExecuteMultiShard sharded.-20: dummy_subquery {} false false`,
		// Those values are returned as 4,5 for ownvindex and 6 for unownvindex.
		// 4,5 have to be replaced by 1,2 (the new values).
		`Execute delete from lkp2 where from1 = :from1 and toc = :toc from1: type:INT64 value:"4" toc: type:UINT64 value:"1"  true`,
		`Execute insert into lkp2(from1, toc) values(:from10, :toc0) from10: type:INT64 value:"1" toc0: type:UINT64 value:"1"  true`,
		// Finally, the actual update, which is also sent to -20, same route as the subquery.
		`ExecuteMultiShard sharded.-20: dummy_update /* vtgate:: keyspace_id:166b40b44aba4bd6 */ {} true true`,
	})

	// No rows changing
	vc = &loggingVCursor{
		shards: []string{"-20", "20-"},
	}
	_, err = upd.Execute(vc, map[string]*querypb.BindVariable{}, false)
	if err != nil {
		t.Fatal(err)
	}
	vc.ExpectLog(t, []string{
		`ResolveDestinations sharded [] Destinations:DestinationKeyspaceID(166b40b44aba4bd6)`,
		// ResolveDestinations is hard-coded to return -20.
		// It gets used to perform the subquery to fetch the changing column values.
		`ExecuteMultiShard sharded.-20: dummy_subquery {} false false`,
		// Subquery returns no rows. So, no vindexes are updated. We still pass-through the original update.
		`ExecuteMultiShard sharded.-20: dummy_update /* vtgate:: keyspace_id:166b40b44aba4bd6 */ {} true true`,
	})
}
