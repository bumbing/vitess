package main

import (
	"fmt"
	"sort"
	"strconv"
	"strings"

	"vitess.io/vitess/go/json2"
	querypb "vitess.io/vitess/go/vt/proto/query"
	vschemapb "vitess.io/vitess/go/vt/proto/vschema"
	"vitess.io/vitess/go/vt/sqlparser"
)

// This is the result of running the following query in patio:
// "select gid - id from advertisers limit 1"
//
// It represents the difference between an advertiser "local id"
// and a "global ID", which is just a local ID plus type bits.
//
// The long term plan is to deprecate locate IDs so that only
// gids exist, but until we implement that we'll have both
// floating around.
const advertiserGIDOffset = "549755813888"
const vindexSuffix = "_vdx"

func init() {
	commands["create-vschema"] = createVSchema
}

func createVSchema(ddls []*sqlparser.DDL, config pinschemaConfig) (string, error) {
	vs, err := newVschemaBuilder(ddls, config).ddlsToVSchema()
	if err != nil {
		return "", err
	}

	b, err := json2.MarshalIndentPB(vs, "  ")
	if err != nil {
		return "", err
	}

	return fmt.Sprintf("%s", b), nil
}

type vschemaBuilder struct {
	tables   map[string]*vschemapb.Table
	vindexes map[string]*vschemapb.Vindex
	ddls     []*sqlparser.DDL
	config   pinschemaConfig
}

func newVschemaBuilder(ddls []*sqlparser.DDL, config pinschemaConfig) *vschemaBuilder {
	return &vschemaBuilder{
		tables:   map[string]*vschemapb.Table{},
		vindexes: map[string]*vschemapb.Vindex{},
		ddls:     ddls,
		config:   config,
	}
}

func isPrimaryVindex(vindexName string, tableName string, colName string) bool {
	if tableName == "advertisers" {
		return colName == "id"
	} else if tableName == "targeting_attribute_counts_by_advertiser" {
		return colName == "advertiser_gid"
	} else {
		// For every other table, "g_advertiser_id" is the primary vindex.
		return vindexName == "g_advertiser_id"
	}
}

func shouldSkipColumn(tableName string, colName string) bool {
	if tableName == "advertisers" && colName == "gid" {
		return true
	}

	return false
}

func (vb *vschemaBuilder) addColumnVindex(vindexName string, colName string, isPrimaryVindex bool, tblVindexes *[]*vschemapb.ColumnVindex) {
	// Add the relevant vindex for this column. If it's the primary vindex, it
	// needs to be added to the beginning of the list.
	if _, ok := vb.vindexes[vindexName]; ok {
		tableVindex := &vschemapb.ColumnVindex{
			Name:    vindexName,
			Columns: []string{colName},
		}

		if isPrimaryVindex {
			*tblVindexes = append([]*vschemapb.ColumnVindex{tableVindex}, *tblVindexes...)
		} else {
			*tblVindexes = append(*tblVindexes, tableVindex)
		}
	}
}

func (vb *vschemaBuilder) ddlsToVSchema() (*vschemapb.Keyspace, error) {
	if vb.config.createPrimary {
		vb.createPrimaryVindexes()
	}

	if vb.config.createSecondary {
		vb.createSecondaryVindexes()
	}

	for _, tableCreate := range vb.ddls {
		tableName := tableCreate.Table.Name.String()

		tbl := &vschemapb.Table{}

		if strings.HasSuffix(tableName, "_seq") {
			tbl.Type = "sequence"
		}

		tblVindexes := make([]*vschemapb.ColumnVindex, 0)

		if vb.config.includeCols && vb.config.colsAuthoritative {
			tbl.ColumnListAuthoritative = true
		}

		// For each column in the current table.
		for _, col := range tableCreate.TableSpec.Columns {
			colName := col.Name.String()

			if vb.config.includeCols {
				colSpec := &vschemapb.Column{
					Name: col.Name.String(),
				}
				if strings.ToLower(col.Type.Type) == "varchar" {
					colSpec.Type = querypb.Type_VARCHAR
				}
				tbl.Columns = append(tbl.Columns, colSpec)
			}

			if shouldSkipColumn(tableName, colName) {
				continue
			}

			vindexName, ok := maybeGetVindexName(colName, tableName)
			if ok {
				isPrimaryVindex := isPrimaryVindex(vindexName, tableName, colName)
				isFunctionalVindex := vindexName == "g_advertiser_id" || vindexName == "advertiser_id"
				if isFunctionalVindex {
					vb.addColumnVindex(vindexName, colName, isPrimaryVindex, &tblVindexes)
				}

				addScatterCache := !isFunctionalVindex && *fallbackToScatterCache
				if addScatterCache {
					vb.addColumnVindex(vindexName, colName, false, &tblVindexes)
				}

				lookupVindexName := vindexName + vindexSuffix
				vb.addColumnVindex(lookupVindexName, colName, false, &tblVindexes)

				// Sort secondary indexes alphabetically by name to simplify unit testing.
				if len(tblVindexes) > 1 {
					secondaryVindexes := tblVindexes[1:]
					sort.Slice(secondaryVindexes, func(i, j int) bool {
						return secondaryVindexes[i].Name < secondaryVindexes[j].Name
					})
				}
			}

			// A column named "id" which has a primary key will be assigned a sequence.
			// Previously we checked for bool(col.Type.Autoincrement) but that will
			// break once sequences launch and auto-increment is removed.
			if vb.config.createSeq && colShouldBeSequence(vb.config, col, tableCreate) {
				tbl.AutoIncrement = &vschemapb.AutoIncrement{
					Column:   colName,
					Sequence: tableName + "_seq",
				}
			}
		}

		if len(tblVindexes) > 0 {
			tbl.ColumnVindexes = tblVindexes
		}
		vb.tables[tableName] = tbl
	}

	var vs vschemapb.Keyspace
	vs.Tables = vb.tables
	if len(vb.vindexes) > 0 {
		vs.Vindexes = vb.vindexes
	}
	vs.Sharded = vb.config.createPrimary

	return &vs, nil
}

func tableNameToColName(tableName string) string {
	return singularize(tableName) + "_id"
}

func maybeGetVindexName(colName, tableName string) (string, bool) {
	if colName == "advertiser_gid" {
		return "g_advertiser_id", true
	} else if colName == "id" {
		return tableNameToColName(tableName), true
	} else if colName == "gid" {
		return "g_" + tableNameToColName(tableName), true
	} else if colName == "spec_id" {
		if tableName == "pin_promotion_labels" {
			return "pin_promotion_spec_id", true
		} else if tableName == "pin_promotions" {
			return "", false
		} else {
			return singularize(tableName) + "_spec_id", true
		}
	} else if colName == "conversion_tag_id" {
		// Some tables reference conversion_tag_id to conversion_tag_v3 table, which is not possible to be used as
		// unowned Vindex since the table is in PatioGeneral keyspace. To totally deprecate ScatterCache, skipping it.
		return "", false
	}

	return colName, true
}

func (vb *vschemaBuilder) createPrimaryVindexes() {
	vb.vindexes["advertiser_id"] = &vschemapb.Vindex{
		Type: "hash_offset",
		Params: map[string]string{
			"offset": advertiserGIDOffset,
		},
	}
	vb.vindexes["dark_write_advertiser_id"] = &vschemapb.Vindex{
		Type: "hash_offset",
		Params: map[string]string{
			"offset": advertiserGIDOffset,
		},
	}
	vb.vindexes["g_advertiser_id"] = &vschemapb.Vindex{
		Type: "hash",
	}
}

func (vb *vschemaBuilder) scatterCacheCapacity(tableName string) uint64 {
	tableCapacity, ok := vb.config.tableScatterCacheCapacity[tableName]
	if ok {
		return tableCapacity
	}
	return vb.config.defaultScatterCacheCapacity
}

func (vb *vschemaBuilder) createScatterCache(tableName string) {
	foreignKeyColName := tableNameToColName(tableName)
	if _, ok := vb.vindexes[foreignKeyColName]; ok {
		return
	}

	vb.vindexes[foreignKeyColName] = &vschemapb.Vindex{
		Type: "scatter_cache",
		Params: map[string]string{
			"table":    tableName,
			"capacity": strconv.FormatUint(vb.scatterCacheCapacity(tableName), 10),
			"from":     "id",
			"to":       "g_advertiser_id",
		},
	}
}

func (vb *vschemaBuilder) createPinLookupVindex(tableName string) {
	vindexName := tableNameToColName(tableName) + vindexSuffix
	indexTableName := "patiogeneral." + tableNameToColName(tableName) + vindexTableSuffix
	vb.vindexes[vindexName] = &vschemapb.Vindex{
		Type:  "pin_lookup_hash_unique",
		Owner: tableName,
		Params: map[string]string{
			"table":      indexTableName,
			"capacity":   strconv.FormatUint(vb.scatterCacheCapacity(tableName), 10),
			"from":       "id",
			"to":         "g_advertiser_id",
			"autocommit": "true",
		},
	}
}

func (vb *vschemaBuilder) hasColumn(tableCreate *sqlparser.DDL, columnName string) bool {
	for _, col := range tableCreate.TableSpec.Columns {
		if col.Name.EqualString(columnName) {
			return true
		}
	}
	return false
}

func (vb *vschemaBuilder) createSecondaryVindexes() {
	for _, tableCreate := range vb.ddls {
		tableName := tableCreate.Table.Name.String()

		if shouldCreateSecondaryVindex := vb.hasColumn(tableCreate, "id"); !shouldCreateSecondaryVindex {
			continue //bypass this table, as neither scatter cache or lookup vindex would be needed.
		}
		// create scatter cache
		if *fallbackToScatterCache {
			vb.createScatterCache(tableName)
		}
		// create lookup vindex if needed
		if vb.hasColumn(tableCreate, "g_advertiser_id") {
			if !*fallbackToScatterCache {
				vb.createPinLookupVindex(tableName)
			}
		}
	}
}