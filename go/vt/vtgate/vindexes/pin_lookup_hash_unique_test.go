package vindexes

import (
	"errors"
	"fmt"
	"reflect"
	"strings"
	"testing"
	"vitess.io/vitess/go/sqltypes"
	"vitess.io/vitess/go/vt/key"
	querypb "vitess.io/vitess/go/vt/proto/query"
	topodatapb "vitess.io/vitess/go/vt/proto/topodata"
	vtgatepb "vitess.io/vitess/go/vt/proto/vtgate"
)

// A modified version of Struct vcursor for PinLookupHashUnique Vindex.
type pinVcursor struct {
	mustFail    bool
	numRows     int
	result      *sqltypes.Result
	queries     []*querypb.BoundQuery
	autocommits int
	pre, post   int
}

func (pvc *pinVcursor) Execute(method string, query string, bindvars map[string]*querypb.BindVariable, isDML bool, co vtgatepb.CommitOrder) (*sqltypes.Result, error) {
	switch co {
	case vtgatepb.CommitOrder_PRE:
		pvc.pre++
	case vtgatepb.CommitOrder_POST:
		pvc.post++
	case vtgatepb.CommitOrder_AUTOCOMMIT:
		pvc.autocommits++
	}
	return pvc.execute(method, query, bindvars, isDML)
}

func (pvc *pinVcursor) ExecuteKeyspaceID(keyspace string, ksid []byte, query string, bindVars map[string]*querypb.BindVariable, isDML, autocommit bool) (*sqltypes.Result, error) {
	return pvc.execute("ExecuteKeyspaceID", query, bindVars, isDML)
}

func (pvc *pinVcursor) execute(method string, query string, bindvars map[string]*querypb.BindVariable, isDML bool) (*sqltypes.Result, error) {
	pvc.queries = append(pvc.queries, &querypb.BoundQuery{
		Sql:           query,
		BindVariables: bindvars,
	})
	if pvc.mustFail {
		return nil, errors.New("execute failed")
	}
	switch {
	case strings.HasPrefix(query, "select"):
		if pvc.result != nil {
			return pvc.result, nil
		}
		result := &sqltypes.Result{
			Fields:       sqltypes.MakeTestFields("col", "int32"),
			RowsAffected: uint64(pvc.numRows),
		}
		for i := 0; i < pvc.numRows; i++ {
			result.Rows = append(result.Rows, []sqltypes.Value{
				sqltypes.NewInt64(int64(i + 1)),
				sqltypes.NewInt64(int64(i + 1)),
			})
		}
		return result, nil
	case strings.HasPrefix(query, "insert"):
		return &sqltypes.Result{InsertID: 1}, nil
	case strings.HasPrefix(query, "delete"):
		return &sqltypes.Result{}, nil
	}
	panic("unexpected")
}

func TestPinLookupHashUniqueMap(t *testing.T) {
	plhu := createLookup(t, "pin_lookup_hash_unique", false)
	pvc := &pinVcursor{numRows: 1}

	got, err := plhu.Map(pvc, []sqltypes.Value{sqltypes.NewInt64(1)})
	if err != nil {
		t.Error(err)
	}
	want := []key.Destination{
		key.DestinationKeyspaceID([]byte("\x16k@\xb4J\xbaK\xd6")),
	}
	if !reflect.DeepEqual(got, want) {
		t.Errorf("Map(): %#v, want %+v", got, want)
	}

	pvc.numRows = 0
	got, err = plhu.Map(pvc, []sqltypes.Value{sqltypes.NewInt64(1)})
	if err != nil {
		t.Error(err)
	}
	want = []key.Destination{
		key.DestinationNone{},
	}
	if !reflect.DeepEqual(got, want) {
		t.Errorf("Map(): %#v, want %+v", got, want)
	}

	pvc.numRows = 2
	_, err = plhu.Map(pvc, []sqltypes.Value{sqltypes.NewInt64(1)})
	wantErr := fmt.Sprintf("PinLookupHashUnique.Map: More result than expected. Expected size %v rows. Got %v", 1, pvc.numRows)
	if err == nil || err.Error() != wantErr {
		t.Errorf("plhu(row count mismatch) err: %v, want %s", err, wantErr)
	}

	// Test conversion fail.
	pvc.result = sqltypes.MakeTestResult(
		sqltypes.MakeTestFields(
			"fromc|toc",
			"varbinary|decimal",
		),
		"a|1",
	)
	got, err = plhu.Map(pvc, []sqltypes.Value{sqltypes.NewInt64(1)})
	wantErr = "PinLookupHashUnique.Map: Result key parsing error. Code: INVALID_ARGUMENT\ncould not parse value: 'a'\n"
	if err == nil || err.Error() != wantErr {
		t.Errorf("plhu(parsing fail) err: %v, want %s", err, wantErr)
	}
	if got != nil {
		t.Errorf("Map(): %#v, want %+v", got, want)
	}

	// Test query fail.
	pvc.mustFail = true
	_, err = plhu.Map(pvc, []sqltypes.Value{sqltypes.NewInt64(1)})
	wantErr = "PinLookupHashUnique.Map: Select query execution error. execute failed"
	if err == nil || err.Error() != wantErr {
		t.Errorf("plhu(query fail) err: %v, want %s", err, wantErr)
	}
	pvc.mustFail = false
}

func TestPinLookupHashUniqueMapWriteOnly(t *testing.T) {
	plhu := createLookup(t, "pin_lookup_hash_unique", true)
	pvc := &pinVcursor{numRows: 0}

	got, err := plhu.Map(pvc, []sqltypes.Value{sqltypes.NewInt64(1)})
	if err != nil {
		t.Error(err)
	}
	want := []key.Destination{
		key.DestinationKeyRange{KeyRange: &topodatapb.KeyRange{}},
	}
	if !reflect.DeepEqual(got, want) {
		t.Errorf("Map(): %#v, want %+v", got, want)
	}
}

func TestPinLookupHashUniqueVerify(t *testing.T) {
	plhu := createLookup(t, "pin_lookup_hash_unique", false)
	pvc := &pinVcursor{numRows: 1}

	// regular test copied from lookup_hash_unique_test.go
	got, err := plhu.Verify(pvc,
		[]sqltypes.Value{sqltypes.NewInt64(1), sqltypes.NewInt64(2)},
		[][]byte{[]byte("\x16k@\xb4J\xbaK\xd6"), []byte("\x06\xe7\xea\"Βp\x8f")})
	if err != nil {
		t.Error(err)
	}
	want := []bool{true, true}
	if !reflect.DeepEqual(got, want) {
		t.Errorf("plhu.Verify(match): %v, want %v", got, want)
	}

	pvc.numRows = 0
	got, err = plhu.Verify(pvc, []sqltypes.Value{sqltypes.NewInt64(1)}, [][]byte{[]byte("\x16k@\xb4J\xbaK\xd6")})
	if err != nil {
		t.Error(err)
	}
	want = []bool{false}
	if !reflect.DeepEqual(got, want) {
		t.Errorf("plhu.Verify(mismatch): %v, want %v", got, want)
	}

	_, err = plhu.Verify(pvc, []sqltypes.Value{sqltypes.NewInt64(1)}, [][]byte{[]byte("bogus")})
	wantErr := "lookup.Verify.vunhash: invalid keyspace id: 626f677573"
	if err == nil || err.Error() != wantErr {
		t.Errorf("plhu.Verify(bogus) err: %v, want %s", err, wantErr)
	}

	// Null value id Verify should pass Verify
	got, err = plhu.Verify(pvc, []sqltypes.Value{sqltypes.NULL}, [][]byte{[]byte("\x16k@\xb4J\xbaK\xd6")})
	if err != nil {
		t.Error(err)
	}
	want = []bool{true}
	if !reflect.DeepEqual(got, want) {
		t.Errorf("plhu.Verify(match): %v, want %v", got, want)
	}
}

func TestPinLookupHashUniqueCache(t *testing.T) {
	plhu := createLookup(t, "pin_lookup_hash_unique", false)
	pvc := &pinVcursor{numRows: 1}

	_ = plhu.(Lookup).Create(pvc, [][]sqltypes.Value{{sqltypes.NewInt64(1)}}, [][]byte{[]byte("\x16k@\xb4J\xbaK\xd6")}, false /* ignoreMode */)
	wantqueries := []*querypb.BoundQuery{{
		Sql: "insert into t(fromc, toc) values(:fromc0, :toc0)",
		BindVariables: map[string]*querypb.BindVariable{
			"fromc0": sqltypes.Int64BindVariable(1),
			"toc0":   sqltypes.Uint64BindVariable(1),
		},
	}}
	if !reflect.DeepEqual(pvc.queries, wantqueries) {
		t.Errorf("lookup.Create queries:\n%v, want\n%v", pvc.queries, wantqueries)
	}

	// Cached value should not send a query
	_, err := plhu.Map(pvc, []sqltypes.Value{sqltypes.NewInt64(1)})
	if err != nil {
		t.Error(err)
	}
	if !reflect.DeepEqual(pvc.queries, wantqueries) {
		t.Errorf("lookup.Create queries:\n%v, want\n%v", pvc.queries, wantqueries)
	}

	// Not cached value will have a new query
	pvc.result = sqltypes.MakeTestResult(
		sqltypes.MakeTestFields(
			"fromc|toc",
			"int64|int64",
		),
		"2|1",
	)

	got, err := plhu.Map(pvc, []sqltypes.Value{sqltypes.NewInt64(2)})
	if err != nil {
		t.Error(err)
	}
	want := []key.Destination{
		key.DestinationKeyspaceID([]byte("\x16k@\xb4J\xbaK\xd6")),
	}
	if !reflect.DeepEqual(got, want) {
		t.Errorf("Map(): %#v, want %+v", got, want)
	}
	wantqueries = []*querypb.BoundQuery{
		{
			Sql: "insert into t(fromc, toc) values(:fromc0, :toc0)",
			BindVariables: map[string]*querypb.BindVariable{
				"fromc0": sqltypes.Int64BindVariable(1),
				"toc0":   sqltypes.Uint64BindVariable(1),
			},
		},
		{
			Sql: "select fromc, toc from t where fromc in ::fromc",
			BindVariables: map[string]*querypb.BindVariable{
				"fromc": {
					Type: querypb.Type_TUPLE,
					Values: []*querypb.Value{},
				},
			},
		},
	}
	if !reflect.DeepEqual(pvc.queries, wantqueries) {
		t.Errorf("lookup.Create queries:\n%v, want\n%v", pvc.queries, wantqueries)
	}
}
