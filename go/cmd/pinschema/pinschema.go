package main

// To generate a vschema
//   pinschema gen_vschema -keyspace=<name> -add_seqs [ddl.sql] [another_ddl.sql] [...]
//   pinschema gen_seq_ddls [ddl.sql] [another_ddl.sql] [...]

import (
	"bytes"
	"flag"
	"fmt"
	"io/ioutil"
	"os"
	"sort"
	"strconv"
	"strings"

	"vitess.io/vitess/go/exit"
	"vitess.io/vitess/go/flagutil"
	"vitess.io/vitess/go/json2"
	"vitess.io/vitess/go/sqlescape"
	"vitess.io/vitess/go/vt/log"
	"vitess.io/vitess/go/vt/logutil"
	querypb "vitess.io/vitess/go/vt/proto/query"
	vschemapb "vitess.io/vitess/go/vt/proto/vschema"
	"vitess.io/vitess/go/vt/servenv"
	"vitess.io/vitess/go/vt/sqlparser"
)

var (
	createPrimaryVindexes       = flag.Bool("create-primary-vindexes", false, "Whether to make primary vindexes")
	createSecondaryVindexes     = flag.Bool("create-secondary-vindexes", false, "Whether to make secondary vindexes")
	createSequences             = flag.Bool("create-sequences", false, "Whether to make sequences")
	includeCols                 = flag.Bool("include-cols", false, "Whether to include a column list for each table")
	colsAuthoritative           = flag.Bool("cols-authoritative", false, "Whether to mark the column list as authoriative")
	outputDDL                   = flag.String("output-ddl", "", "'create-seq' or 'remove-autoinc' to output DDLs instead of vschema")
	defaultScatterCacheCapacity = flag.Uint64("default-scatter-cache-capacity", 100000, "default capacity for a scatter cache vindex")
	tableScatterCacheCapacity   flagutil.StringMapValue
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

type pinschemaConfig struct {
	createPrimary               bool
	createSecondary             bool
	createSeq                   bool
	defaultScatterCacheCapacity uint64
	tableScatterCacheCapacity   map[string]uint64
	includeCols                 bool
	colsAuthoritative           bool
}

func init() {
	flag.Var(&tableScatterCacheCapacity,
		"table-scatter-cache-capacity",
		"comma separated list of table:capacity pairs to override the default capacity")

	logger := logutil.NewConsoleLogger()
	flag.CommandLine.SetOutput(logutil.NewLoggerWriter(logger))
}

func main() {
	defer exit.RecoverAll()
	defer logutil.Flush()

	args := servenv.ParseFlagsWithArgs("pinschema")

	err := parseAndRun(args)
	if err != nil {
		fmt.Fprintln(os.Stderr, "ERROR:", err)
		exit.Return(1)
	}
}

func readAndParseSchema(fname string) ([]*sqlparser.DDL, error) {
	schemaStr, err := ioutil.ReadFile(fname)
	if err != nil {
		return nil, fmt.Errorf("Cannot read file %v: %v", fname, err)
	}

	ddl, err := parseSchema(string(schemaStr))
	if err != nil {
		return nil, err
	}

	return ddl, nil
}

func tableNameToColName(tableName string) string {
	return singularize(tableName) + "_id"
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

func (vb *vschemaBuilder) createSecondaryVindexes() {
	for _, tableCreate := range vb.ddls {
		tableName := tableCreate.Table.Name.String()
		foreignKeyColName := tableNameToColName(tableName)
		if _, ok := vb.vindexes[foreignKeyColName]; ok {
			continue
		}

		hasIDCol := false
		for _, col := range tableCreate.TableSpec.Columns {
			if "id" == col.Name.String() {
				hasIDCol = true
				break
			}
		}
		if !hasIDCol {
			continue
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
}

func (vb *vschemaBuilder) getVindexName(colName, tableName string) string {
	if colName == "advertiser_gid" {
		return "g_advertiser_id"
	} else if colName == "id" {
		return tableNameToColName(tableName)
	} else if colName == "gid" {
		return "g_" + tableNameToColName(tableName)
	}

	return colName
}

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

			vindexName := vb.getVindexName(colName, tableName)

			// For the advertisers table we use "id" as the primary vindex and we have no
			// secondary vindex on "gid" because it's initially null.
			isPrimaryVindex := false
			if tableName == "advertisers" || tableName == "dark_write_advertisers" {
				isPrimaryVindex = colName == "id"
				if colName == "gid" {
					// Can't set a secondary index on advertisers.gid because it's initially NULL.
					continue
				}
			} else {
				// For every other table, "g_advertiser_id" is the primary vindex.
				isPrimaryVindex = vindexName == "g_advertiser_id"
			}

			// Add the relevant vindex for this column. If it's the primary vindex, it
			// needs to be added to the beginning of the list.
			if _, ok := vb.vindexes[vindexName]; ok {
				tableVindex := &vschemapb.ColumnVindex{
					Name:    vindexName,
					Columns: []string{colName},
				}

				if isPrimaryVindex {
					tblVindexes = append([]*vschemapb.ColumnVindex{tableVindex}, tblVindexes...)
				} else {
					tblVindexes = append(tblVindexes, tableVindex)
				}
			}

			// Sort secondary indexes alphabetically by name to simplify unit testing.
			if len(tblVindexes) > 1 {
				secondaryVindexes := tblVindexes[1:]
				sort.Slice(secondaryVindexes, func(i, j int) bool {
					return secondaryVindexes[i].Name < secondaryVindexes[j].Name
				})
			}

			// A column named "id" which has a primary key will be assigned a sequence.
			// Previously we checked for bool(col.Type.Autoincrement) but that will
			// break once sequences launch and auto-increment is removed.
			if vb.config.createSeq && colShouldBeSequence(col, tableCreate) {
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

func parseAndRun(args []string) error {
	if len(args) == 0 {
		return fmt.Errorf("No .sql ddl files listed")
	}

	var ddls []*sqlparser.DDL
	for _, fname := range args {
		ddl, err := readAndParseSchema(fname)
		if err != nil {
			return err
		}
		ddls = append(ddls, ddl...)
	}

	switch *outputDDL {
	case "create-seq":
		fmt.Print(buildSequenceDDLs(ddls))
		return nil
	case "remove-autoinc":
		fmt.Print(removeAutoInc(ddls))
		return nil
	case "":
		// Fall thhrough
	default:
		return fmt.Errorf("Bad option to -output-ddl: %v. Should be 'create-seq' or 'remove-autoinc'", *outputDDL)
	}

	tableCacheCapacityOverrides := map[string]uint64{}
	for key, value := range tableScatterCacheCapacity {
		parsed, err := strconv.ParseUint(value, 10, 64)
		if err != nil {
			log.Fatalf("Bad -table-scatter-cache-capacity arg: %v", err)
		}
		tableCacheCapacityOverrides[key] = parsed
	}

	vs, err := newVschemaBuilder(ddls, pinschemaConfig{
		createPrimary:               *createPrimaryVindexes,
		createSecondary:             *createSecondaryVindexes,
		createSeq:                   *createSequences,
		includeCols:                 *includeCols,
		colsAuthoritative:           *colsAuthoritative,
		defaultScatterCacheCapacity: *defaultScatterCacheCapacity,
		tableScatterCacheCapacity:   tableCacheCapacityOverrides,
	}).ddlsToVSchema()
	if err != nil {
		return err
	}

	b, err := json2.MarshalIndentPB(vs, "  ")
	if err != nil {
		return err
	}

	fmt.Printf("%s", b)

	return nil
}

func buildSequenceDDLs(ddls []*sqlparser.DDL) string {
	var b bytes.Buffer

	for _, tableCreate := range ddls {
		tableName := tableCreate.Table.Name.String()
		if strings.HasPrefix(tableName, "dark_write") {
			continue
		}

		hasAutoincrement := false
		for _, col := range tableCreate.TableSpec.Columns {
			if colShouldBeSequence(col, tableCreate) {
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
	return b.String()
}

func removeAutoInc(ddls []*sqlparser.DDL) string {
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
	return b.String()
}

// singularize removes the "s" from a table name.
//
// Example: advertisers -> advertiser
func singularize(tableName string) string {
	if strings.HasSuffix(tableName, "s") && tableName != "accepted_tos" {
		return tableName[0 : len(tableName)-1]
	}
	return tableName
}

// parseSchema pulls out the CREATE TABLE ddls from a series of SQL statements.
// This method is copied from vtexplain.go.
func parseSchema(sqlSchema string) ([]*sqlparser.DDL, error) {
	parsedDDLs := make([]*sqlparser.DDL, 0, 16)
	for {
		sql, rem, err := sqlparser.SplitStatement(sqlSchema)
		sqlSchema = rem
		if err != nil {
			return nil, err
		}
		if sql == "" {
			break
		}
		sql = sqlparser.StripComments(sql)
		if sql == "" {
			continue
		}

		stmt, err := sqlparser.Parse(sql)
		if err != nil {
			log.Errorf("ERROR: failed to parse sql: %s, got error: %v", sql, err)
			return nil, err
		}
		ddl, ok := stmt.(*sqlparser.DDL)
		if !ok {
			log.Infof("ignoring non-DDL statement: %s", sql)
			continue
		}
		if ddl.Action != sqlparser.CreateStr {
			log.Infof("ignoring %s table statement", ddl.Action)
			continue
		}
		if ddl.TableSpec == nil {
			log.Errorf("invalid create table statement: %s", sql)
			continue
		}
		parsedDDLs = append(parsedDDLs, ddl)
	}
	return parsedDDLs, nil
}
