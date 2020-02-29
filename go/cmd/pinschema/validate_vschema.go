package main

import (
	"crypto/md5"
	"encoding/hex"
	"errors"
	"fmt"
	"strings"

	"vitess.io/vitess/go/json2"
	"vitess.io/vitess/go/vt/proto/query"
	vschemapb "vitess.io/vitess/go/vt/proto/vschema"
	"vitess.io/vitess/go/vt/sqlparser"
	"vitess.io/vitess/go/vt/vtexplain"
	"vitess.io/vitess/go/vt/vtgate/vindexes"
)

// ErrPrimaryVindexNotUsed indicates the vschema doesn't have the correct primary vindex
var ErrPrimaryVindexNotUsed = errors.New("primary vindex not used in write")

// ErrVtExplainNoExplains indicates the query isn't understood by vtexplain
var ErrVtExplainNoExplains = errors.New("vtexplain not able to explain on the query")

// ErrUnsupportedPrimaryVindexType indicates the vschema doesn't have
var ErrUnsupportedPrimaryVindexType = errors.New("primary vindex uses a non numeric or string value")

func init() {
	commands["validate-vschema"] = validateVschema
}

func validateVschema(ddls []*sqlparser.DDL, config pinschemaConfig) (string, error) {

	keyspace := config.validateKeyspace
	ksVschema := config.validateVschema
	wrappedStr := fmt.Sprintf(`{"keyspaces": %s}`, ksVschema)
	srvVSchema := &vschemapb.SrvVSchema{}
	if err := json2.Unmarshal([]byte(wrappedStr), srvVSchema); err != nil {
		return "", err
	}

	ks, ok := srvVSchema.Keyspaces[keyspace]
	if !ok {
		return "", fmt.Errorf("no keyspace:%s is found in vschema", keyspace)
	}
	if !ks.Sharded {
		return "", fmt.Errorf("keyspace:%s is not sharded", keyspace)
	}

	sqlBuf := sqlparser.NewTrackedBuffer(nil)
	for _, ddl := range ddls {
		if strings.ToLower(ddl.Action) == "create" {
			ddl.Format(sqlBuf)
			sqlBuf.Myprintf(";\n")
		}
	}
	options := &vtexplain.Options{
		NumShards:       config.validateShards,
		ReplicationMode: "ROW",
	}
	if err := vtexplain.Init(ksVschema, sqlBuf.String(), options); err != nil {
		return "", err
	}

	for _, ddl := range ddls {
		if _, ok := ks.Tables[strings.ToLower(ddl.Table.Name.String())]; !ok {
			// the table isn't part of the validate keyspace, ignore it
			continue
		}
		var err error
		// first validate if the table structure and vschema's table def is a match
		if err = validateColumns(keyspace, ks, ddl); err != nil {
			return "", err
		}
		// then validate the sequence table def is present
		if err = validateSequence(keyspace, ks, ddl, ddls); err != nil {
			return "", err
		}
		// then we check the correctness of vindex columns, each column should be found in the table schema
		if err = validateVindexColumns(keyspace, ks, ddl); err != nil {
			return "", err
		}
		// then validate lookup vindex table def is present
		if err = validatePinLookupVindexes(keyspace, ks, ddls); err != nil {
			return "", err
		}
		// then generates a list of validation SQLs each validates the use of a functional unique vindex
		var sqls []string
		if sqls, err = generateSQLs(keyspace, ks, ddl); err != nil {
			return "", err
		}
		// use vtexplain to acquire the query plan
		var explains []*vtexplain.Explain
		for _, sql := range sqls {
			if explains, err = vtexplain.Run(sql); err != nil {
				return sql, err
			}
			if err = validateSQL(explains); err != nil {
				return sql, err
			}
		}
	}

	// computes the md5 sign of the given vschema
	hash := md5.Sum([]byte(config.validateVschema))
	return hex.EncodeToString(hash[:]), nil
}

func validateColumns(keyspace string, ks *vschemapb.Keyspace, ddl *sqlparser.DDL) (err error) {
	table, ok := ks.Tables[strings.ToLower(ddl.Table.Name.String())]
	if !ok {
		return fmt.Errorf("table %s not found in %s", ddl.Table.Name, keyspace)
	}
	columns := make(map[string]query.Type, len(table.Columns))
	for _, c := range table.Columns {
		columns[strings.ToLower(c.Name)] = c.Type
	}
	if len(ddl.TableSpec.Columns) != len(columns) {
		return fmt.Errorf("table %s has columns list mismatch", ddl.Table.Name)
	}
	for _, c := range ddl.TableSpec.Columns {
		// Type_NULL_TYPE indicates no vschema type is actually provided, tolerate this case.
		if t, ok := columns[c.Name.Lowered()]; !ok || (c.Type.SQLType() != t && t != query.Type_NULL_TYPE) {
			return fmt.Errorf("table %s has column type mismatch of %s", ddl.Table.Name, c.Name.Lowered())
		}
	}
	return
}

func validateSequence(keyspace string, ks *vschemapb.Keyspace, ddl *sqlparser.DDL, ddls []*sqlparser.DDL) (err error) {
	table, ok := ks.Tables[strings.ToLower(ddl.Table.Name.String())]
	if !ok {
		return fmt.Errorf("table %s not found in %s", ddl.Table.Name, keyspace)
	}
	if table.AutoIncrement != nil {
		if seq := strings.ToLower(table.AutoIncrement.GetSequence()); seq != "" {
			for _, seqDDL := range ddls {
				if strings.ToLower(seqDDL.Table.Name.String()) == seq {
					return nil
				}
			}
			return fmt.Errorf("table %s has sequence, but sequence table is not found", ddl.Table.Name)
		}
	}
	return nil
}

func validateVindexColumns(keyspace string, ks *vschemapb.Keyspace, ddl *sqlparser.DDL) (err error) {
	columns := make(map[string]sqlparser.ColumnType, len(ddl.TableSpec.Columns))
	for _, c := range ddl.TableSpec.Columns {
		columns[c.Name.Lowered()] = c.Type
	}
	table, ok := ks.Tables[strings.ToLower(ddl.Table.Name.String())]
	if !ok {
		return fmt.Errorf("table %s not found in %s", ddl.Table.Name, keyspace)
	}
	for _, vi := range table.ColumnVindexes {
		for _, vic := range vi.Columns {
			if _, ok := columns[strings.ToLower(vic)]; !ok {
				return fmt.Errorf("table %s has vindex:%s column: %s not found in schema", ddl.Table.Name, vi.Name, vic)
			}
		}
	}
	return nil
}

func validatePinLookupVindexes(keyspace string, ks *vschemapb.Keyspace, ddls []*sqlparser.DDL) (err error) {
VINDEX:
	for _, vi := range ks.Vindexes {
		if vi.Type == "pin_lookup_hash_unique" {
			lookupTable := vi.GetParams()["table"]
			if offset := strings.IndexByte(lookupTable, '.'); offset > 0 {
				lookupTable = lookupTable[offset+1:]
			}
			for _, lookupTableDDL := range ddls {
				if strings.EqualFold(lookupTableDDL.Table.Name.String(), lookupTable) {
					continue VINDEX
				}
			}
			return fmt.Errorf("table %s has lookup vindex, but the lookup table is not found", lookupTable)
		}
	}
	return nil
}

func generateSQLs(keyspace string, ks *vschemapb.Keyspace, ddl *sqlparser.DDL) (sql []string, err error) {
	table := ddl.Table
	var vindexCols []*sqlparser.ColumnDefinition
	if vindexCols, err = functionalUniqueVindexes(keyspace, ks, ddl); err != nil {
		return nil, err
	}

	sqls := make([]string, len(vindexCols))
	for i := 0; i < len(sqls); i++ {
		vc := vindexCols[i]
		vv, err := someColumnValue(vc)
		if err != nil {
			return nil, err
		}

		buf := sqlparser.NewTrackedBuffer(nil)
		buf.Myprintf("delete from %s.%v where %v=%v", keyspace, table, vc.Name, vv)
		sqls[i] = buf.String()
	}

	return sqls, nil
}

func validateSQL(explains []*vtexplain.Explain) error {
	if len(explains) == 0 || len(explains[0].Plans) == 0 {
		return ErrVtExplainNoExplains
	}
	for _, plan := range explains[0].Plans {
		// lookup vindex route should be `DeleteUnsharded`
		// onwer table route should be `DeleteEqual`
		if plan.Instructions.RouteType() != "DeleteUnsharded" && plan.Instructions.RouteType() != "DeleteEqual" {
			return ErrPrimaryVindexNotUsed
		}
	}
	return nil
}

func functionalUniqueVindexes(keyspace string, ks *vschemapb.Keyspace, ddl *sqlparser.DDL) ([]*sqlparser.ColumnDefinition, error) {
	table, ok := ks.Tables[strings.ToLower(ddl.Table.Name.String())]
	if !ok {
		return nil, fmt.Errorf("no table:%s is found in keyspace:%s", ddl.Table.Name, keyspace)
	}
	if len(table.ColumnVindexes) == 0 {
		return nil, fmt.Errorf("no column vindexes defined for table:%s", ddl.Table.Name)
	}

	// NOTE, there's an assumption that a vindex has a single column in it
	vindexCols := []*sqlparser.ColumnDefinition{}

COLUMN_VINDEXES:
	for _, cvi := range table.ColumnVindexes {
		vis, ok := ks.Vindexes[cvi.Name]
		if !ok {
			return nil, fmt.Errorf("table has column vindex:%s not defined in keyspace", cvi.Name)
		}

		// this additionally checks if the vindex could be successfully created
		vi, err := vindexes.CreateVindex(vis.Type, cvi.Name, vis.Params)
		if err != nil {
			return nil, err
		}
		// primary vindex don't need vcursor to look up in storage and should be uniqne
		if !vi.NeedsVCursor() && vi.IsUnique() {
			vic := cvi.Columns[0]
			for _, col := range ddl.TableSpec.Columns {
				if col.Name.EqualString(vic) {
					vindexCols = append(vindexCols, col)
					continue COLUMN_VINDEXES
				}
			}
			return nil, fmt.Errorf("vindex:%s in table:%s is not found in the column list", vic, ddl.Table.Name)
		}
	}

	if len(vindexCols) == 0 {
		return nil, fmt.Errorf("table:%s has no functional vindex", ddl.Table.Name)
	}

	return vindexCols, nil
}

func someColumnValue(cd *sqlparser.ColumnDefinition) (*sqlparser.SQLVal, error) {
	switch cd.Type.SQLType() {
	case query.Type_FLOAT32, query.Type_FLOAT64, query.Type_DECIMAL:
		return sqlparser.NewFloatVal([]byte("1.1")), nil
	case query.Type_UINT8, query.Type_UINT16, query.Type_UINT32, query.Type_UINT64,
		query.Type_INT8, query.Type_INT16, query.Type_INT32, query.Type_INT64:
		return sqlparser.NewIntVal([]byte("1")), nil
	case query.Type_BINARY, query.Type_BLOB, query.Type_TEXT, query.Type_VARBINARY, query.Type_VARCHAR:
		return sqlparser.NewStrVal([]byte("1")), nil
	default:
		return nil, ErrUnsupportedPrimaryVindexType
	}
}
