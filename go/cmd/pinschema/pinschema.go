package main

// To generate a vschema
//   pinschema gen_vschema -keyspace=<name> -add_seqs [ddl.sql] [another_ddl.sql] [...]
//   pinschema gen_seq_ddls [ddl.sql] [another_ddl.sql] [...]

import (
	"flag"
	"fmt"
	"io/ioutil"
	"os"
	"strings"

	"vitess.io/vitess/go/exit"
	"vitess.io/vitess/go/json2"
	"vitess.io/vitess/go/vt/log"
	"vitess.io/vitess/go/vt/logutil"
	"vitess.io/vitess/go/vt/servenv"
	"vitess.io/vitess/go/vt/sqlparser"

	vschemapb "vitess.io/vitess/go/vt/proto/vschema"
)

var (
	createPrimaryVindexes   = flag.Bool("create-primary-vindexes", false, "Whether to make primary vindexes")
	createSecondaryVindexes = flag.Bool("create-secondary-vindexes", false, "Whether to make secondary vindexes")
	createSequences         = flag.Bool("create-sequences", false, "Whether to make sequences")
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

func ddlsToVSchema(ddls []*sqlparser.DDL, config pinschemaConfig) (*vschemapb.Keyspace, error) {
	tables := map[string]*vschemapb.Table{}
	vindexes := map[string]*vschemapb.Vindex{}

	if config.createPrimary {
		vindexes["advertiser_id"] = &vschemapb.Vindex{
			Type: "hash_offset",
			Params: map[string]string{
				"offset": advertiserGIDOffset,
			},
		}
		vindexes["advertiser_gid"] = &vschemapb.Vindex{
			Type: "hash",
		}
	}

	if config.createSecondary {
		for _, tableCreate := range ddls {
			tableName := tableCreate.NewName.Name.String()

			// NOTE(dweitzman): scatter_cache_unique does not exist yet as an
			// vindex, so this won't work in practice yet. The idea is that
			// it would resolve a campaign ID to keyspace ID by doing a scatter
			// query across all shards and then caching the result in memory to
			// avoid repeating the work multiple times on the same vtgate.
			//
			// If this has acceptable performance, it'll be lighter weight to
			// maintain in the short term than a full table-based lookup vindex.
			//
			// In the very long term we'll probably want to use table-backed
			// secondary vindexes because as the number of shards increases the
			// cost of scatter queries and decreased reliability will eventually
			// outweigh the simplicity of not maintaining lookup tables on disk.
			vindexes[singularize(tableName)+"_idx"] = &vschemapb.Vindex{
				Type: "scatter_cache_unique",
			}
		}
	}

	for _, tableCreate := range ddls {
		tableName := tableCreate.NewName.Name.String()

		tbl := &vschemapb.Table{}

		tblVindexes := make([]*vschemapb.ColumnVindex, 0)

		advertiserGIDColName := "g_advertiser_id"
		if tableName == "advertisers" {
			advertiserGIDColName = "gid"
		} else if tableName == "targeting_attribute_counts_by_advertiser" {
			// This one table uses the wrong name.
			advertiserGIDColName = "advertiser_gid"
		}

		if config.createPrimary {
			if tableName == "advertisers" {
				tblVindexes = append(tblVindexes, &vschemapb.ColumnVindex{
					Name:    "advertiser_id",
					Columns: []string{"id"},
				})
			} else {
				tblVindexes = append(tblVindexes, &vschemapb.ColumnVindex{
					Name:    "advertiser_gid",
					Columns: []string{advertiserGIDColName},
				})
			}
		}

		for _, col := range tableCreate.TableSpec.Columns {
			colName := col.Name.String()
			refTable := ""

			// Reference the scatter_unique index for columns like "campaign_id", or like "id"
			// in the campaigns table.
			if strings.HasSuffix(colName, "_id") {
				refTable = singularize(colName[0 : len(colName)-3])
			} else if colName == "id" {
				refTable = singularize(tableName)
			}

			// Only reference secondary vindexes that actually exist. This protects against misinterpretting
			// "pin_id" as a foreign key to a "pin" table with a "pin_idx" vindex. No such table
			// or index exists in patio.
			if _, ok := vindexes[refTable+"_idx"]; ok {
				tblVindexes = append(tblVindexes, &vschemapb.ColumnVindex{
					Name:    refTable + "_idx",
					Columns: []string{colName},
				})
			}

			if bool(col.Type.Autoincrement) && config.createSeq {
				tbl.AutoIncrement = &vschemapb.AutoIncrement{
					Column:   colName,
					Sequence: tableName + "_seq",
				}
			}
		}

		if len(tblVindexes) > 0 {
			tbl.ColumnVindexes = tblVindexes
		}
		tables[tableName] = tbl
	}

	var vs vschemapb.Keyspace
	vs.Tables = tables
	if len(vindexes) > 0 {
		vs.Vindexes = vindexes
	}
	vs.Sharded = config.createPrimary

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

	vs, err := ddlsToVSchema(ddls, pinschemaConfig{
		createPrimary:   *createPrimaryVindexes,
		createSecondary: *createSecondaryVindexes,
		createSeq:       *createSequences,
	})
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
