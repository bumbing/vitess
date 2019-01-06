package vindexes

import (
	"errors"
	"reflect"
	"strings"
	"testing"

	"vitess.io/vitess/go/sqltypes"
	"vitess.io/vitess/go/vt/key"
	querypb "vitess.io/vitess/go/vt/proto/query"
	topodatapb "vitess.io/vitess/go/vt/proto/topodata"
)

type scatterVcursor struct {
	mustFail    bool
	numRows     int
	result      *sqltypes.Result
	queries     []*querypb.BoundQuery
	autocommits int
}

func (svc *scatterVcursor) Execute(method string, query string, bindvars map[string]*querypb.BindVariable, isDML bool) (*sqltypes.Result, error) {
	return svc.execute(method, query, bindvars, isDML)
}

func (svc *scatterVcursor) ExecuteAutocommit(method string, query string, bindvars map[string]*querypb.BindVariable, isDML bool) (*sqltypes.Result, error) {
	svc.autocommits++
	return svc.execute(method, query, bindvars, isDML)
}

// This method is copied from the lookup vindex tests.
func (svc *scatterVcursor) execute(method string, query string, bindvars map[string]*querypb.BindVariable, isDML bool) (*sqltypes.Result, error) {
	svc.queries = append(svc.queries, &querypb.BoundQuery{
		Sql:           query,
		BindVariables: bindvars,
	})
	if svc.mustFail {
		return nil, errors.New("execute failed")
	}
	switch {
	case strings.HasPrefix(query, "select"):
		if svc.result != nil {
			return svc.result, nil
		}
		result := &sqltypes.Result{
			Fields:       sqltypes.MakeTestFields("col", "int32"),
			RowsAffected: uint64(svc.numRows),
		}
		for i := 0; i < svc.numRows; i++ {
			result.Rows = append(result.Rows, []sqltypes.Value{
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

func TestScatterCacheNew(t *testing.T) {
	l := createScatterCache(t, "2000")
	if want, got := l.(*ScatterCache).keyspaceIDCache.Capacity(), int64(2000); got != want {
		t.Errorf("Create('2000'): %v, want %v", got, want)
	}

	l = createScatterCache(t, "3000")
	if want, got := l.(*ScatterCache).keyspaceIDCache.Capacity(), int64(3000); got != want {
		t.Errorf("Create('3000'): %v, want %v", got, want)
	}

	l = createScatterCache(t, "0")
	if want, got := l.(*ScatterCache).keyspaceIDCache.Capacity(), int64(0); got != want {
		t.Errorf("Create('0'): %v, want %v", got, want)
	}

	l, err := CreateVindex("scatter_cache", "scatter_cache", map[string]string{
		"table":    "t",
		"from":     "fromc",
		"to":       "toc",
		"capacity": "-1",
	})
	want := "scatter_cache: capacity contains illegal characters: -1"
	if err == nil || err.Error() != want {
		t.Errorf("Create(bad_scatter): %v, want %s", err, want)
	}

	l, err = CreateVindex("scatter_cache", "scatter_cache", map[string]string{
		"table": "t",
		"from":  "fromc",
		"to":    "toc",
	})
	want = "scatter_cache: missing required field: capacity"
	if err == nil || err.Error() != want {
		t.Errorf("Create(bad_scatter): %v, want %s", err, want)
	}
}

func TestScatterCacheCost(t *testing.T) {
	scatterCache := createScatterCache(t, "1000")
	if scatterCache.Cost() != 30 {
		t.Errorf("Cost(): %d, want 30", scatterCache.Cost())
	}
}

func TestScatterCacheString(t *testing.T) {
	scatterCache := createScatterCache(t, "1000")
	if strings.Compare("scatter_cache", scatterCache.String()) != 0 {
		t.Errorf("String(): %s, want scatter_cache", scatterCache.String())
	}
}

func TestScatterCacheMap(t *testing.T) {
	scatterCache := createScatterCache(t, "1000")
	svc := &scatterVcursor{numRows: 1}

	got, err := scatterCache.Map(svc, []sqltypes.Value{sqltypes.NewInt64(1), sqltypes.NewInt64(2)})
	if err != nil {
		t.Error(err)
	}
	want := []key.Destination{
		key.DestinationKeyspaceID(vhash(1)),
		key.DestinationKeyspaceID(vhash(1)),
	}
	if !reflect.DeepEqual(got, want) {
		t.Errorf("Map(): %#v, want %+v", got, want)
	}

	wantqueries := []*querypb.BoundQuery{{
		Sql: "select /*vt+ FORCE_SCATTER=1 */ toc from t where fromc = :fromc",
		BindVariables: map[string]*querypb.BindVariable{
			"fromc": sqltypes.Int64BindVariable(1),
		},
	}, {
		Sql: "select /*vt+ FORCE_SCATTER=1 */ toc from t where fromc = :fromc",
		BindVariables: map[string]*querypb.BindVariable{
			"fromc": sqltypes.Int64BindVariable(2),
		},
	}}
	if !reflect.DeepEqual(svc.queries, wantqueries) {
		t.Errorf("lookup.Map queries:\n%v, want\n%v", svc.queries, wantqueries)
	}

	// Result should be cached

	svc.numRows = 0
	got, err = scatterCache.Map(svc, []sqltypes.Value{sqltypes.NewInt64(1), sqltypes.NewInt64(2)})
	if err != nil {
		t.Error(err)
	}
	if !reflect.DeepEqual(got, want) {
		t.Errorf("Map(): %#v, want %+v", got, want)
	}
}

func TestScatterCacheMapNoCapacity(t *testing.T) {
	scatterCache := createScatterCache(t, "0")
	svc := &scatterVcursor{numRows: 1}

	got, err := scatterCache.Map(svc, []sqltypes.Value{sqltypes.NewInt64(1), sqltypes.NewInt64(2)})
	if err != nil {
		t.Error(err)
	}
	want := []key.Destination{
		key.DestinationKeyRange{KeyRange: &topodatapb.KeyRange{}},
		key.DestinationKeyRange{KeyRange: &topodatapb.KeyRange{}},
	}
	if !reflect.DeepEqual(got, want) {
		t.Errorf("Map(): %#v, want %+v", got, want)
	}

	if len(svc.queries) > 0 {
		t.Errorf("lookup.Map unexpected queries:\n%v", svc.queries)
	}
}

func TestScatterCacheMapQueryFail(t *testing.T) {
	scatterCache := createScatterCache(t, "1000")
	svc := &scatterVcursor{numRows: 1}

	// Test query fail.
	svc.mustFail = true
	_, err := scatterCache.Map(svc, []sqltypes.Value{sqltypes.NewInt64(1)})
	wantErr := "ScatterCache.Map: execute failed"
	if err == nil || err.Error() != wantErr {
		t.Errorf("scatterCache(query fail) err: %v, want %s", err, wantErr)
	}
	svc.mustFail = false
}

func TestScatterCacheMapTooManyResults(t *testing.T) {
	scatterCache := createScatterCache(t, "1000")
	svc := &scatterVcursor{numRows: 2}

	_, err := scatterCache.Map(svc, []sqltypes.Value{sqltypes.NewInt64(1)})
	wantErr := "ScatterCache.Map: unexpected multiple results from vindex t, key INT64(1)"
	if err == nil || err.Error() != wantErr {
		t.Errorf("scatterCache(query fail) err: %v, want %s", err, wantErr)
	}
	svc.mustFail = false
}

func TestScatterCacheMapNoResults(t *testing.T) {
	scatterCache := createScatterCache(t, "1000")
	svc := &scatterVcursor{numRows: 0}

	got, err := scatterCache.Map(svc, []sqltypes.Value{sqltypes.NewInt64(1), sqltypes.NewInt64(2)})
	if err != nil {
		t.Error(err)
	}
	want := []key.Destination{
		key.DestinationNone{},
		key.DestinationNone{},
	}
	if !reflect.DeepEqual(got, want) {
		t.Errorf("Map(): %#v, want %+v", got, want)
	}

	wantqueries := []*querypb.BoundQuery{{
		Sql: "select /*vt+ FORCE_SCATTER=1 */ toc from t where fromc = :fromc",
		BindVariables: map[string]*querypb.BindVariable{
			"fromc": sqltypes.Int64BindVariable(1),
		},
	}, {
		Sql: "select /*vt+ FORCE_SCATTER=1 */ toc from t where fromc = :fromc",
		BindVariables: map[string]*querypb.BindVariable{
			"fromc": sqltypes.Int64BindVariable(2),
		},
	}}
	if !reflect.DeepEqual(svc.queries, wantqueries) {
		t.Errorf("lookup.Map queries:\n%v, want\n%v", svc.queries, wantqueries)
	}

	// Results should not be cached
	svc.numRows = 1
	got, err = scatterCache.Map(svc, []sqltypes.Value{sqltypes.NewInt64(1), sqltypes.NewInt64(2)})
	if err != nil {
		t.Error(err)
	}

	want = []key.Destination{
		key.DestinationKeyspaceID(vhash(1)),
		key.DestinationKeyspaceID(vhash(1)),
	}
	if !reflect.DeepEqual(got, want) {
		t.Errorf("Map(): %#v, want %+v", got, want)
	}
}

func createScatterCache(t *testing.T, capacity string) Vindex {
	t.Helper()
	l, err := CreateVindex("scatter_cache", "scatter_cache", map[string]string{
		"table":    "t",
		"from":     "fromc",
		"to":       "toc",
		"capacity": capacity,
	})
	if err != nil {
		t.Fatal(err)
	}
	return l
}
