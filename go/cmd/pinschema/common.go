package main

import (
	"strings"

	"vitess.io/vitess/go/vt/sqlparser"
)

func colShouldBeSequence(col *sqlparser.ColumnDefinition, tableCreate *sqlparser.DDL) bool {
	tableName := tableCreate.Table.Name.String()

	if strings.HasPrefix(tableName, "_drop_") || strings.HasPrefix(tableName, "dark_write") {
		return false
	}

	colName := col.Name.String()

	// A column named "id" which has a primary key will be assigned a sequence.
	// Previously we checked for bool(col.Type.Autoincrement) but that will
	// break once sequences launch and auto-increment is removed.
	if colName == "id" {
		for _, idx := range tableCreate.TableSpec.Indexes {
			if idx.Info.Primary && len(idx.Columns) == 1 && idx.Columns[0].Column.Equal(col.Name) {
				return true
			}
		}
	}
	return false
}
