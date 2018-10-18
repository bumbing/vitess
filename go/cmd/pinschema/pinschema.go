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
	"strings"

	"vitess.io/vitess/go/exit"
	"vitess.io/vitess/go/json2"
	"vitess.io/vitess/go/sqlescape"
	"vitess.io/vitess/go/vt/log"
	"vitess.io/vitess/go/vt/logutil"
	vschemapb "vitess.io/vitess/go/vt/proto/vschema"
	"vitess.io/vitess/go/vt/servenv"
	"vitess.io/vitess/go/vt/sqlparser"
)

var (
	createPrimaryVindexes   = flag.Bool("create-primary-vindexes", false, "Whether to make primary vindexes")
	createSecondaryVindexes = flag.Bool("create-secondary-vindexes", false, "Whether to make secondary vindexes")
	createSequences         = flag.Bool("create-sequences", false, "Whether to make sequences")
	sequenceTableDDLs       = flag.Bool("sequence-table-ddls", false, "Whether to output sequence table DDL instead of vschema")
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
	createPrimary   bool
	createSecondary bool
	createSeq       bool
}

func init() {
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

func (vb *vschemaBuilder) createSecondaryVindexes() {
	for _, tableCreate := range vb.ddls {
		tableName := tableCreate.NewName.Name.String()
		foreignKeyColName := tableNameToColName(tableName)
		if _, ok := vb.vindexes[foreignKeyColName]; ok {
			continue
		}
		vb.vindexes[foreignKeyColName] = &vschemapb.Vindex{
			Type: "scatter_cache",
			Params: map[string]string{
				"capacity": "10000",
				"table":    tableName,
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

func (vb *vschemaBuilder) ddlsToVSchema() (*vschemapb.Keyspace, error) {
	if vb.config.createPrimary {
		vb.createPrimaryVindexes()
	}

	if vb.config.createSecondary {
		vb.createSecondaryVindexes()
	}

	for _, tableCreate := range vb.ddls {
		tableName := tableCreate.NewName.Name.String()

		tbl := &vschemapb.Table{}

		tblVindexes := make([]*vschemapb.ColumnVindex, 0)

		// For each column in the current table.
		for _, col := range tableCreate.TableSpec.Columns {
			colName := col.Name.String()
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

			if bool(col.Type.Autoincrement) && vb.config.createSeq && !strings.HasPrefix(tableName, "dark_write") {
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

	if *sequenceTableDDLs {
		fmt.Print(buildSequenceDDLs(ddls))
		return nil
	}

	vs, err := newVschemaBuilder(ddls, pinschemaConfig{
		createPrimary:   *createPrimaryVindexes,
		createSecondary: *createSecondaryVindexes,
		createSeq:       *createSequences,
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
		tableName := tableCreate.NewName.Name.String()
		if strings.HasPrefix(tableName, "dark_write") {
			continue
		}

		seqTableName := sqlescape.EscapeID(tableName + "_seq")
		fmt.Fprintf(
			&b,
			"create table %s(id int, next_id bigint, cache bigint, primary key(id)) comment 'vitess_sequence';\n",
			seqTableName)
		fmt.Fprintf(&b, "insert into %s(id, next_id, cache) values(0, 1, 1);\n\n", seqTableName)
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
