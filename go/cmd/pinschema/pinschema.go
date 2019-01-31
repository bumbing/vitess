package main

// To generate a vschema
//   pinschema gen_vschema -keyspace=<name> -add_seqs [ddl.sql] [another_ddl.sql] [...]
//   pinschema gen_seq_ddls [ddl.sql] [another_ddl.sql] [...]

import (
	"flag"
	"fmt"
	"io/ioutil"
	"os"
	"sort"
	"strconv"
	"strings"

	"vitess.io/vitess/go/exit"
	"vitess.io/vitess/go/flagutil"
	"vitess.io/vitess/go/vt/log"
	"vitess.io/vitess/go/vt/logutil"
	"vitess.io/vitess/go/vt/sqlparser"
)

var (
	createPrimaryVindexes       = flag.Bool("create-primary-vindexes", false, "Whether to make primary vindexes")
	createSecondaryVindexes     = flag.Bool("create-secondary-vindexes", false, "Whether to make secondary vindexes")
	createSequences             = flag.Bool("create-sequences", false, "Whether to make sequences")
	includeCols                 = flag.Bool("include-cols", false, "Whether to include a column list for each table")
	colsAuthoritative           = flag.Bool("cols-authoritative", false, "Whether to mark the column list as authoriative")
	defaultScatterCacheCapacity = flag.Uint64("default-scatter-cache-capacity", 100000, "default capacity for a scatter cache vindex")
	tableScatterCacheCapacity   flagutil.StringMapValue
)

type pinschemaConfig struct {
	createPrimary               bool
	createSecondary             bool
	createSeq                   bool
	defaultScatterCacheCapacity uint64
	tableScatterCacheCapacity   map[string]uint64
	includeCols                 bool
	colsAuthoritative           bool
}

var commands = make(map[string]func([]*sqlparser.DDL, pinschemaConfig) (string, error))

func init() {
	flag.Var(&tableScatterCacheCapacity,
		"table-scatter-cache-capacity",
		"comma separated list of table:capacity pairs to override the default capacity")

	logger := logutil.NewConsoleLogger()
	flag.CommandLine.SetOutput(logutil.NewLoggerWriter(logger))
}

func getUsageMsg() error {
	commandsList := []string{}
	for command := range commands {
		commandsList = append(commandsList, command)
	}
	sort.Strings(commandsList)

	return fmt.Errorf("Usage: pinschema <%s> <...sql files with CREATE statements...>", strings.Join(commandsList, "|"))
}

func main() {
	defer exit.RecoverAll()
	defer logutil.Flush()

	if len(os.Args) < 2 {
		fmt.Fprintln(os.Stderr, "ERROR:", getUsageMsg())
		exit.Return(1)
	}

	flag.CommandLine.Parse(os.Args[2:])
	args := flag.Args()

	err := parseAndRun(os.Args[1], args)
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

func parseAndRun(command string, args []string) error {
	if len(args) < 1 {
		return getUsageMsg()
	}

	// Move all the flags into a config structure so logic lower down won't
	// need to read command line flags directly. This makes unit testing cleaner,
	// among other things.
	tableCacheCapacityOverrides := map[string]uint64{}
	for key, value := range tableScatterCacheCapacity {
		parsed, err := strconv.ParseUint(value, 10, 64)
		if err != nil {
			log.Fatalf("Bad -table-scatter-cache-capacity arg: %v", err)
		}
		tableCacheCapacityOverrides[key] = parsed
	}

	config := pinschemaConfig{
		createPrimary:               *createPrimaryVindexes,
		createSecondary:             *createSecondaryVindexes,
		createSeq:                   *createSequences,
		includeCols:                 *includeCols,
		colsAuthoritative:           *colsAuthoritative,
		defaultScatterCacheCapacity: *defaultScatterCacheCapacity,
		tableScatterCacheCapacity:   tableCacheCapacityOverrides,
	}

	var ddls []*sqlparser.DDL
	for _, fname := range args {
		ddl, err := readAndParseSchema(fname)
		if err != nil {
			return err
		}
		ddls = append(ddls, ddl...)
	}

	commandImpl, ok := commands[command]
	if !ok {
		return fmt.Errorf("Unrecognized command: %v. %v", command, getUsageMsg())
	}

	output, err := commandImpl(ddls, config)

	if err != nil {
		return err
	}

	fmt.Printf("%s", output)

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
