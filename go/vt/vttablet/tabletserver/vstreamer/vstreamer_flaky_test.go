package vstreamer

import (
	"testing"

	"golang.org/x/net/context"
	binlogdatapb "vitess.io/vitess/go/vt/proto/binlogdata"
)

func TestREKeyRange(t *testing.T) {
	if testing.Short() {
		t.Skip()
	}

	execStatements(t, []string{
		"create table t1(id1 int, id2 int, val varbinary(128), primary key(id1))",
	})
	defer execStatements(t, []string{
		"drop table t1",
	})
	engine.se.Reload(context.Background())

	if err := env.SetVSchema(shardedVSchema); err != nil {
		t.Fatal(err)
	}
	defer env.SetVSchema("{}")

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	filter := &binlogdatapb.Filter{
		Rules: []*binlogdatapb.Rule{{
			Match:  "/.*/",
			Filter: "-80",
		}},
	}
	ch := startStream(ctx, t, filter, "")

	// 1, 2, 3 and 5 are in shard -80.
	// 4 and 6 are in shard 80-.
	input := []string{
		"begin",
		"insert into t1 values (1, 4, 'aaa')",
		"insert into t1 values (4, 1, 'bbb')",
		// Stay in shard.
		"update t1 set id1 = 2 where id1 = 1",
		// Move from -80 to 80-.
		"update t1 set id1 = 6 where id1 = 2",
		// Move from 80- to -80.
		"update t1 set id1 = 3 where id1 = 4",
		"commit",
	}
	execStatements(t, input)
	expectLog(ctx, t, input, ch, [][]string{{
		`gtid|begin`,
		`gtid|begin`,
		`type:FIELD field_event:<table_name:"t1" fields:<name:"id1" type:INT32 > fields:<name:"id2" type:INT32 > fields:<name:"val" type:VARBINARY > > `,
		`type:ROW row_event:<table_name:"t1" row_changes:<after:<lengths:1 lengths:1 lengths:3 values:"14aaa" > > > `,
		`type:ROW row_event:<table_name:"t1" row_changes:<before:<lengths:1 lengths:1 lengths:3 values:"14aaa" > after:<lengths:1 lengths:1 lengths:3 values:"24aaa" > > > `,
		`type:ROW row_event:<table_name:"t1" row_changes:<before:<lengths:1 lengths:1 lengths:3 values:"24aaa" > > > `,
		`type:ROW row_event:<table_name:"t1" row_changes:<after:<lengths:1 lengths:1 lengths:3 values:"31bbb" > > > `,
		`commit`,
	}})

	// Switch the vschema to make id2 the primary vindex.
	altVSchema := `{
  "sharded": true,
  "vindexes": {
    "hash": {
      "type": "hash"
    }
  },
  "tables": {
    "t1": {
      "column_vindexes": [
        {
          "column": "id2",
          "name": "hash"
        }
      ]
    }
  }
}`
	if err := env.SetVSchema(altVSchema); err != nil {
		t.Fatal(err)
	}

	// Only the first insert should be sent.
	input = []string{
		"begin",
		"insert into t1 values (4, 1, 'aaa')",
		"insert into t1 values (1, 4, 'aaa')",
		"commit",
	}
	execStatements(t, input)
	expectLog(ctx, t, input, ch, [][]string{{
		`gtid|begin`,
		`gtid|begin`,
		`type:ROW row_event:<table_name:"t1" row_changes:<after:<lengths:1 lengths:1 lengths:3 values:"41aaa" > > > `,
		`commit`,
	}})
}
