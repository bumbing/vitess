package main

import (
	"bytes"
	"fmt"

	"vitess.io/vitess/go/vt/sqlparser"
)

func init() {
	commands["remove-autoinc"] = removeAutoInc
}

func removeAutoInc(ddls []*sqlparser.DDL, config pinschemaConfig) (string, error) {
	var b bytes.Buffer

	for _, tableCreate := range ddls {
		tableName := tableCreate.Table.Name.String()

		for _, col := range tableCreate.TableSpec.Columns {
			if bool(col.Type.Autoincrement) {
				col.Type.Autoincrement = sqlparser.BoolVal(false)

				fmt.Fprintf(
					&b,
					"alter table %v modify %v;\n",
					tableName, sqlparser.String(col))

				break
			}
		}

	}
	return b.String(), nil
}
