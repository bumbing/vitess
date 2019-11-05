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
	queryTablePrefix            = flag.String("query-table-prefix", "", "A prefix to add to tables for generated queries. Used to support hive with the sharding integrity check")
	tableResultLimit            = flag.Int("table-result-limit", 0, "max results to show per table when with sharding integrity check. 0 for unlimited")
	summarize                   = flag.Bool("summarize", false, "whether to summarize results")
	colsAuthoritative           = flag.Bool("cols-authoritative", false, "Whether to mark the column list as authoriative")
	defaultScatterCacheCapacity = flag.Uint64("default-scatter-cache-capacity", 100000, "default capacity for a scatter cache vindex")
	tableScatterCacheCapacity   flagutil.StringMapValue
	ignoredTables               flagutil.StringListValue
	sequenceTables              flagutil.StringListValue
	validateKeyspace            = flag.String("validate-keyspace", "patio", "Which keyspace needs to validate the vschema correctness")
	validateShards              = flag.Int("validate-shards", 2, "How many shards is actively serving master for the validate keyspace")
	validateVschemaFile         = flag.String("validate-vschema-file", "", "Where the vschema file is for validation")
	fallbackToScatterCache      = flag.Bool("fall-back-to-scatter-cache", false, "If Lookup Vindex serving wrong data or patiogeneral is not available, VSchema can fall back to ScatterCache is this equals to true.")
)

type pinschemaConfig struct {
	createPrimary               bool
	createSecondary             bool
	createSeq                   bool
	defaultScatterCacheCapacity uint64
	tableScatterCacheCapacity   map[string]uint64
	includeCols                 bool
	colsAuthoritative           bool
	queryTablePrefix            string
	tableResultLimit            int
	summarize                   bool
	sequenceTableWhitelist      []string
	validateVschema             string
	validateKeyspace            string
	validateShards              int
}

var commands = make(map[string]func([]*sqlparser.DDL, pinschemaConfig) (string, error))

func init() {
	flag.Var(&tableScatterCacheCapacity,
		"table-scatter-cache-capacity",
		"comma separated list of table:capacity pairs to override the default capacity")

	flag.Var(&ignoredTables,
		"ignore",
		"comma separated list of tables to ignore")

	flag.Var(&sequenceTables,
		"seq-table-whitelist",
		"comma separated whitelist of tables that should use sequences, for incrementally rolling out sequences to a keyspace table by table")

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
		queryTablePrefix:            *queryTablePrefix,
		tableResultLimit:            *tableResultLimit,
		summarize:                   *summarize,
		sequenceTableWhitelist:      sequenceTables,
		validateKeyspace:            *validateKeyspace,
		validateShards:              *validateShards,
	}

	var ddls []*sqlparser.DDL
	for _, fname := range args {
		fileDdls, err := readAndParseSchema(fname)
		if err != nil {
			return err
		}
		for _, ddl := range fileDdls {
			if !shouldIgnoreTable(ddl, ignoredTables) {
				ddls = append(ddls, ddl)
			}
		}
	}

	// Sort by table name.
	sort.Slice(ddls, func(i int, j int) bool {
		return ddls[i].Table.Name.String() < ddls[j].Table.Name.String()
	})

	// get validate vschema if needed
	if *validateVschemaFile != "" {
		vschema, err := ioutil.ReadFile(*validateVschemaFile)
		if err != nil {
			return err
		}
		config.validateVschema = string(vschema)
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

func shouldIgnoreTable(table *sqlparser.DDL, ignoredTables []string) bool {
	tableName := strings.ToLower(table.Table.Name.String())
	if strings.HasPrefix(tableName, "_") {
		return true
	}
	for _, ignoredTable := range ignoredTables {
		if tableName == ignoredTable {
			return true
		}
	}
	return false
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
