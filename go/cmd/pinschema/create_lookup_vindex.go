package main

import (
	"bytes"
	"fmt"
	"vitess.io/vitess/go/sqlescape"
	"vitess.io/vitess/go/vt/log"

	"vitess.io/vitess/go/vt/sqlparser"
)

func init() {
	commands["create-lookup-vindex"] = buildVindexDDLs
}

func tableContainsIdColumn(tableCreate *sqlparser.DDL) bool {
	for _, col := range tableCreate.TableSpec.Columns {
		if col.Name.Lowered() == "id" {
			return true
		}
	}
	return false
}

func createVindexTable(tableCreate *sqlparser.DDL, b *bytes.Buffer) {
	tableName := singularize(tableCreate.Table.Name.String())

	indexTableName := sqlescape.EscapeID(tableName + "_id" + vindexTableSuffix)
	_, _ = fmt.Fprintf(
		b,
		"CREATE TABLE IF NOT EXISTS %s (\n  `id` bigint(20) NOT NULL DEFAULT '0',\n  `g_advertiser_id` bigint(20) DEFAULT NULL,\n  primary key(`id`)\n) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci COMMENT 'vitess_lookup_vindex';\n",
		indexTableName)
}

func buildVindexDDLs(ddls []*sqlparser.DDL, config pinschemaConfig) (string, error) {
	var b bytes.Buffer

	for _, tableCreate := range ddls {
		if !tableInVindexWhitelist(config, tableCreate) {
			continue
		}

		if !tableContainsIdColumn(tableCreate) {
			log.Warning("Table %s does not contain id column, skip lookup vindex table creation on: ", tableCreate.Table.Name.String())
			continue
		}

		createVindexTable(tableCreate, &b)

	}
	return b.String(), nil
}