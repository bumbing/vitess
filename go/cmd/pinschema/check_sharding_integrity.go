package main

// This tool generates a huge SQL query that joins tables that look related and makes note of
// any inconsistency between the advertiser gid for the rows in related tables. These should not
// happen and won't work in a resharded world. Short term we can fix whatever we find.
// Long term, this query should be regenerated periodically and tested to avoid any regressions.

import (
	"bytes"
	"html/template"
	"sort"
	"strings"

	"vitess.io/vitess/go/vt/sqlparser"
)

func init() {
	commands["check-sharding-integrity"] = checkShardingIntegrity
}

const sqlTemplate = `{{if .JoinLimit}}(
{{end}}{{$tableLimit := .TableLimit}}{{$count := .Count}}{{range $i, $e := .Joins}}{{if $i}}
union
{{end}}  {{if $tableLimit}}({{end}}select
	"{{.Left.Table}}.{{.Left.Column}} with {{.Right.Table}}.{{.Right.Column}}" as tables,{{if $count}}
	count(*){{else}}
	{{.Left.Table}}.id as left_id,
    {{.Right.Table}}.id as right_id,
    {{.Left.Table}}.{{.Left.Advertiser}} as left_advertiser,
    {{.Right.Table}}.{{.Right.Advertiser}} as right_advertiser{{end}}
  from
    {{.Left.Table}}
  inner join
    {{.Right.Table}}
  on
    {{.Left.Table}}.{{.Left.Column}} = {{.Right.Table}}.{{.Right.Column}}
  where
	{{.Left.Table}}.{{.Left.Advertiser}} != {{.Right.Table}}.{{.Right.Advertiser}}{{if .ExtraWhere}}
	{{.ExtraWhere}}{{end}}{{if $tableLimit}}
  limit {{$tableLimit}}){{end}}{{end}}{{if .JoinLimit}})
limit {{.JoinLimit}}{{end}};`

// checkShardingIntegrity builds a SQL script that identifies rows that belong to a different
// advertiser than another row they relate with.
func checkShardingIntegrity(ddls []*sqlparser.DDL, config pinschemaConfig) (string, error) {
	allTableNames := map[string]bool{}
	for _, ddl := range ddls {
		allTableNames[tableNameToColName(ddl.Table.Name.String())] = true
	}

	type QualifiedCol struct {
		Table      string
		Column     string
		Advertiser string
	}

	type Join struct {
		Left       QualifiedCol
		Right      QualifiedCol
		ExtraWhere string
	}

	var b bytes.Buffer

	// Columns that have the same vindex are related, so we group columns by vindex name.
	relatedCols := map[string][]QualifiedCol{}

	for _, tableCreate := range ddls {
		tableName := tableCreate.Table.Name.String()

		for _, col := range tableCreate.TableSpec.Columns {
			colName := col.Name.String()

			vindexName := getVindexName(colName, tableName)

			_, ok := allTableNames[vindexName]
			if !ok {
				continue
			}

			if !strings.HasSuffix(vindexName, "_id") {
				continue
			}

			if strings.HasPrefix(vindexName, "g_") {
				continue
			}

			if vindexName == "advertiser_id" {
				continue
			}

			advertiserCol := "g_advertiser_id"
			if tableName == "advertisers" {
				advertiserCol = "gid"
			}
			qualified := QualifiedCol{tableName, colName, advertiserCol}
			if config.queryTablePrefix != "" {
				qualified.Table = config.queryTablePrefix + qualified.Table
			}
			relatedCols[vindexName] = append(relatedCols[vindexName], qualified)
		}
	}

	tmpl, err := template.New("sql template").Parse(sqlTemplate)
	if err != nil {
		return "", err
	}

	joins := []Join{}
	for _, cols := range relatedCols {
		if len(cols) < 2 {
			continue
		}

		// Sort the table with just "id" to the front.
		sort.Slice(cols, func(i, j int) bool {
			return cols[i].Column == "id" || (cols[j].Column != "id" && cols[i].Table < cols[j].Table)
		})

		rootCol := cols[0]
		for _, childCol := range cols[1:] {
			extraWhere := ""
			if rootCol.Table == "campaigns" && childCol.Table == "bill_details" {
				// NOTE(dweitzman): There are some old bill details in an impossible state from what looks like
				// entities being transfered from to "archived" advertisers to their non-archived new accounts.
				// We are currently allowing bill_details to link to campaigns from a different advertiser in this case.
				// The joins themselves will stop working as resharding separates theses entities into different shards.
				// These bills are relatively old. If these two advertisers reach out directly at some point in the future
				// about an issue loading bill receipts in ads manager we can try to help them make sense of the
				// data or hack ads maanger to lookup campaigns in a separate query instead of joining them to bill details.

				// 549756295881 == CUPSHE (archived)
				// 549756295882 == Sammydress (archived)
				// 549755848277 == CUPSHE
				// 549755917224 == Sammydress
				extraWhere = "AND !(campaigns.g_advertiser_id in (549756295881, 549756295882) and bill_details.g_advertiser_id in (549755848277, 549755917224))"
			}
			joins = append(joins, Join{rootCol, childCol, extraWhere})
		}
	}

	sort.Slice(joins, func(i, j int) bool {
		if joins[i].Left.Table < joins[j].Left.Table {
			return true
		}
		if joins[i].Left.Table > joins[j].Left.Table {
			return false
		}
		if joins[i].Right.Table < joins[j].Right.Table {
			return true
		}
		if joins[i].Right.Table > joins[j].Right.Table {
			return false
		}
		return false
	})

	type IntegrityArgs struct {
		Joins      []Join
		TableLimit int
		JoinLimit  int
		Count      bool
	}
	tmplArgs := IntegrityArgs{
		Joins:      joins,
		TableLimit: config.tableResultLimit,
		JoinLimit:  0,
		Count:      config.summarize,
	}
	tmpl.Execute(&b, tmplArgs)

	return b.String(), nil
}
