package engine

import (
	"testing"

	"vitess.io/vitess/go/sqltypes"
	querypb "vitess.io/vitess/go/vt/proto/query"
	vschemapb "vitess.io/vitess/go/vt/proto/vschema"
	"vitess.io/vitess/go/vt/vtgate/vindexes"
)

// Pinterest test for NULL-valued Unowned Vindex column inserting.
func TestInsertShardedUnownedNullSuccess(t *testing.T) {
	invschema := &vschemapb.SrvVSchema{
		Keyspaces: map[string]*vschemapb.Keyspace{
			"sharded": {
				Sharded: true,
				Vindexes: map[string]*vschemapb.Vindex{
					"primary": {
						Type: "hash",
					},
					"own": {
						Type: "pin_lookup_hash_unique",
						Params: map[string]string{
							"table": "lkp1",
							"from":  "from",
							"to":    "toc",
						},
						Owner: "t1",
					},
					"unown": {
						Type: "pin_lookup_hash_unique",
						Params: map[string]string{
							"table": "lkp2",
							"from":  "from",
							"to":    "toc",
						},
						Owner: "t2",
					},
				},
				Tables: map[string]*vschemapb.Table{
					"t1": {
						ColumnVindexes: []*vschemapb.ColumnVindex{{
							Name:    "primary",
							Columns: []string{"id"},
						}, {
							Name:    "own",
							Columns: []string{"c3"},
						}, {
							Name:    "unown",
							Columns: []string{"c4"},
						}},
					},
				},
			},
		},
	}
	vs, err := vindexes.BuildVSchema(invschema)
	if err != nil {
		t.Fatal(err)
	}
	ks := vs.Keyspaces["sharded"]

	ins := &Insert{
		Opcode:   InsertSharded,
		Keyspace: ks.Keyspace,
		VindexValues: []sqltypes.PlanValue{{
			// colVindex columns: id
			Values: []sqltypes.PlanValue{{
				// rows for id
				Values: []sqltypes.PlanValue{{
					Value: sqltypes.NewInt64(1),
				}},
			}},
		}, {
			// colVindex columns: c3
			Values: []sqltypes.PlanValue{{
				// rows for c3
				Values: []sqltypes.PlanValue{{
					Value: sqltypes.NewInt64(2),
				}},
			}},
		}, {
			// colVindex columns: c4
			Values: []sqltypes.PlanValue{{
				// rows for c3
				Values: []sqltypes.PlanValue{{
					Value: sqltypes.NULL,
				}},
			}},
		}},
		Table:  ks.Tables["t1"],
		Prefix: "prefix",
		Mid:    []string{" mid1", " mid2", " mid3", " mid4"},
		Suffix: " suffix",
	}

	ksid0 := sqltypes.MakeTestResult(
		sqltypes.MakeTestFields(
			"to",
			"varbinary",
		),
		"\x00",
	)

	vc := &loggingVCursor{
		shards:       []string{"-20", "20-"},
		shardForKsid: []string{"-20", "20-"},
		results: []*sqltypes.Result{
			ksid0,
			ksid0,
			ksid0,
		},
	}
	_, err = ins.Execute(vc, map[string]*querypb.BindVariable{}, false)
	if err != nil {
		t.Fatal(err)
	}
	vc.ExpectLog(t, []string{
		`Execute insert into lkp1(from, toc) values(:from0, :toc0) from0: type:INT64 value:"2" toc0: type:UINT64 ` +
			`value:"1"  true`,
		`ResolveDestinations sharded [value:"0" ] Destinations:DestinationKeyspaceID(166b40b44aba4bd6)`,
		`ExecuteMultiShard sharded.-20: prefix mid1 suffix /* vtgate:: keyspace_id:166b40b44aba4bd6 */ ` +
			`{_c30: type:INT64 value:"2" _c40: _id0: type:INT64 value:"1" } true true`,
	})
}

func TestInsertShardedOwnedSuccess(t *testing.T) {
	invschema := &vschemapb.SrvVSchema{
		Keyspaces: map[string]*vschemapb.Keyspace{
			"sharded": {
				Sharded: true,
				Vindexes: map[string]*vschemapb.Vindex{
					"primary": {
						Type: "hash",
					},
					"own": {
						Type: "pin_lookup_hash_unique",
						Params: map[string]string{
							"table": "lkp1",
							"from":  "from",
							"to":    "toc",
						},
						Owner: "t1",
					},
					"unown": {
						Type: "pin_lookup_hash_unique",
						Params: map[string]string{
							"table": "lkp2",
							"from":  "from",
							"to":    "toc",
						},
						Owner: "t2",
					},
				},
				Tables: map[string]*vschemapb.Table{
					"t1": {
						ColumnVindexes: []*vschemapb.ColumnVindex{{
							Name:    "primary",
							Columns: []string{"id"},
						}, {
							Name:    "own",
							Columns: []string{"id"},
						}, {
							Name:    "unown",
							Columns: []string{"c4"},
						}},
						AutoIncrement: &vschemapb.AutoIncrement{
							Column:   "id",
							Sequence: "lkp1",
						},
					},
				},
			},
			"unsharded": {
				Sharded: false,
				Tables: map[string]*vschemapb.Table{
					"lkp1": {
						Type: "sequence",
					},
				},
			},
		},
	}
	vs, err := vindexes.BuildVSchema(invschema)
	if err != nil {
		t.Fatal(err)
	}
	ks := vs.Keyspaces["sharded"]

	ins := &Insert{
		Opcode:   InsertSharded,
		Keyspace: ks.Keyspace,
		VindexValues: []sqltypes.PlanValue{{
			// colVindex columns: id
			Values: []sqltypes.PlanValue{{
				// rows for id
				Values: []sqltypes.PlanValue{{
					Value: sqltypes.NewInt64(1),
				}},
			}},
		}, {
			// colVindex columns: c3
			Values: []sqltypes.PlanValue{{
				// rows for c3
				Values: []sqltypes.PlanValue{{
					Value: sqltypes.NewInt64(2),
				}},
			}},
		}, {
			// colVindex columns: c4
			Values: []sqltypes.PlanValue{{
				// rows for c3
				Values: []sqltypes.PlanValue{{
					Value: sqltypes.NewInt64(3),
				}},
			}},
		}},
		Table:  ks.Tables["t1"],
		Prefix: "prefix",
		Mid:    []string{" mid1", " mid2", " mid3", " mid4"},
		Suffix: " suffix",
	}

	ins.Generate = &Generate{
		Keyspace: &vindexes.Keyspace{
			Name:    "unsharded",
			Sharded: false,
		},
		Query: "dummy_generate",
		Values: sqltypes.PlanValue{
			Values: []sqltypes.PlanValue{
				{Value: sqltypes.NewInt64(1)},
			},
		},
	}

	ksid0 := sqltypes.MakeTestResult(
		sqltypes.MakeTestFields(
			"to",
			"varbinary",
		),
		"\x00",
	)

	vc := &loggingVCursor{
		shards:       []string{"-20", "20-"},
		shardForKsid: []string{"-20", "20-"},
		results: []*sqltypes.Result{
			ksid0,
			ksid0,
			ksid0,
		},
	}
	_, err = ins.Execute(vc, map[string]*querypb.BindVariable{}, false)
	if err != nil {
		t.Fatal(err)
	}
	vc.ExpectLog(t, []string{
		`Execute insert into lkp1(from, toc) values(:from0, :toc0) from0: type:INT64 value:"2" toc0: type:UINT64 ` +
			`value:"1"  true`,
		`Execute select from from lkp2 where from = :from and toc = :toc from: type:INT64 value:"3" toc: type:UINT64 value:"1"  false`,
		`ResolveDestinations sharded [value:"0" ] Destinations:DestinationKeyspaceID(166b40b44aba4bd6)`,
		`ExecuteMultiShard sharded.-20: prefix mid1 suffix /* vtgate:: keyspace_id:166b40b44aba4bd6 */ ` +
			`{__seq0: type:INT64 value:"1" _c40: type:INT64 value:"3" _id0: type:INT64 value:"2" } true true`,
	})
}
