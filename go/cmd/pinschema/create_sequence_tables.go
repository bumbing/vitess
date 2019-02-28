package main

import (
	"bytes"
	"fmt"
	"strings"

	"vitess.io/vitess/go/sqlescape"
	"vitess.io/vitess/go/vt/sqlparser"
)

func init() {
	commands["create-seq"] = buildSequenceDDLs
}

func buildSequenceDDLs(ddls []*sqlparser.DDL, config pinschemaConfig) (string, error) {
	var b bytes.Buffer

	for _, tableCreate := range ddls {
		tableName := tableCreate.Table.Name.String()
		if strings.HasPrefix(tableName, "dark_write") {
			continue
		}

		hasAutoincrement := false
		for _, col := range tableCreate.TableSpec.Columns {
			if colShouldBeSequence(config, col, tableCreate) {
				hasAutoincrement = true
				break
			}
		}
		if !hasAutoincrement {
			continue
		}

		seqTableName := sqlescape.EscapeID(tableName + "_seq")
		fmt.Fprintf(
			&b,
			"create table if not exists %s(id int, next_id bigint, cache bigint, primary key(id)) comment 'vitess_sequence';\n",
			seqTableName)
	}
	return b.String(), nil
}
