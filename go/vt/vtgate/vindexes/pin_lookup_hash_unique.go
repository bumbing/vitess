package vindexes

import (
	"fmt"
	"vitess.io/vitess/go/decider"
	"vitess.io/vitess/go/sqltypes"
	"vitess.io/vitess/go/stats"
	"vitess.io/vitess/go/vt/key"

	querypb "vitess.io/vitess/go/vt/proto/query"
	topodatapb "vitess.io/vitess/go/vt/proto/topodata"
	vtgatepb "vitess.io/vitess/go/vt/proto/vtgate"
)

var (
	_ Vindex = (*PinLookupHashUnique)(nil)
	_ Lookup = (*PinLookupHashUnique)(nil)

	failToLookupVindex = stats.NewCountersWithMultiLabels(
		"VindexLookupFailureCount",
		"Count of Map func failure, uses destNone or destAll for all result",
		[]string{"TableName", "FailureReason"})
)

func init() {
	Register("pin_lookup_hash_unique", NewPinLookupUniqueHash)
}

//====================================================================

// PinLookupHashUnique defines a vindex that uses a lookup table.
// The table is expected to define the id column as unique. It's
// Unique and a Lookup.
type PinLookupHashUnique struct {
	*LookupHashUnique
}

// NewPinLookupHashUnique creates a PinLookupHashUnique vindex.
// The supplied map has the following required fields:
//   table: name of the backing table. It can be qualified by the keyspace.
//   from: list of columns in the table that have the 'from' values of the lookup vindex.
//   to: The 'to' column name of the table.
//
// The following fields are optional:
//   autocommit: setting this to "true" will cause deletes to be ignored.
//   write_only: in this mode, Map functions return the full keyrange causing a full scatter.
func NewPinLookupUniqueHash(name string, m map[string]string) (Vindex, error) {
	lhu, err := NewLookupHashUnique(name, m)
	if err != nil {
		return nil, err
	}
	plhu := &PinLookupHashUnique{
		LookupHashUnique: lhu.(*LookupHashUnique),
	}
	return plhu, nil
}

// Fallback to use destNone or destAll if Map func failed.
// This should be used when lookupVindex is already rollout and there are some new created table with WriteOnly new
// Vindexes, which should last for very short period of time.
func (plhu *PinLookupHashUnique) getScatterResultForAll(size int) []key.Destination {
	out := make([]key.Destination, 0, size)
	for i := 0; i < size; i++ {
			out = append(out, key.DestinationKeyRange{KeyRange: &topodatapb.KeyRange{}})
	}
	return out
}

// This function will get all mapping by sending one query, to avoid using up all connections like LookupHashUnique's
// implementation.
// If failed to lookup Vindex, it could possible return destAll, which will not work for DMLs.
// For DMLs, only INSERT could possible use Map, but that's for primary vindex, not for secondary vindex.
// The fallback destAll is a protection for SELECTIN and SELECTEQUAL query.
func (plhu *PinLookupHashUnique) Map(vcursor VCursor, ids []sqltypes.Value) ([]key.Destination, error) {
	if plhu.writeOnly {
		return plhu.getScatterResultForAll(len(ids)), nil
	}

	sel := fmt.Sprintf(
		"select %s, %s from %s where %s in ::%s",
		plhu.lkp.FromColumns[0], plhu.lkp.To, plhu.lkp.Table, plhu.lkp.FromColumns[0], plhu.lkp.FromColumns[0])

	bv := &querypb.BindVariable{
		Type:   querypb.Type_TUPLE,
		Values: make([]*querypb.Value, len(ids)),
	}
	values := make([]querypb.Value, len(ids))
	for i, id := range ids {
		temp := sqltypes.ValueBindVariable(id)
		values[i].Type = temp.Type
		values[i].Value = temp.Value
		bv.Values[i] = &values[i]
	}

	bindVars := map[string]*querypb.BindVariable{
		plhu.lkp.FromColumns[0]: bv,
	}

	// Experiment on autocommit Vindex lookup. The normal commit order results to transaction pool usage full.
	// TODO(mingjianliu): clean this after trying autocommit.
	var co vtgatepb.CommitOrder
	if decider.CheckDecider("vindex_lookup_autocommit", false) {
		co = vtgatepb.CommitOrder_AUTOCOMMIT
	} else {
		co = vtgatepb.CommitOrder_NORMAL
	}

	queryResult, err := vcursor.Execute(
		"PinLookupHashUnique.Lookup", sel, bindVars, false /* isDML */, co)
	if err != nil {
		failToLookupVindex.Add([]string{plhu.lkp.Table, "select_query_failure"}, 1)
		return nil, fmt.Errorf("PinLookupHashUnique.Map: Select query execution error. %v", err)
	}

	// HashMap to keep the relationship between Ids and KeyspaceId, in case result is returned in different order.
	// Choose both uint64 as key and value, since sqlTypes.Value is not comparable. And by parsing result column, we
	// can also check if data is corrupted.
	m := make(map[uint64]uint64, len(ids))

	if len(queryResult.Rows) > len(ids) {
		failToLookupVindex.Add([]string{plhu.lkp.Table, "result_row_number_mismatch"}, 1)
		return nil, fmt.Errorf("PinLookupHashUnique.Map: More result than expected. Expected size %v rows. Got %v",
				len(ids), len(queryResult.Rows))
	}
	for _, row := range queryResult.Rows {
		if len(row) != 2 {
			failToLookupVindex.Add([]string{plhu.lkp.Table, "result_column_number_mismatch"}, 1)
			return nil, fmt.Errorf("PinLookupHashUnique.Map: Internal error. Expected %v columns. Got %v", 2, len(row))
		}

		fromColKey, err := sqltypes.ToUint64(row[0])
		if err != nil {
			failToLookupVindex.Add([]string{plhu.lkp.Table, "key_parsing_error"}, 1)
			return nil, fmt.Errorf("PinLookupHashUnique.Map: Result key parsing error. %v", err)
		}

		toColValue, err := sqltypes.ToUint64(row[1])
		if err != nil {
			failToLookupVindex.Add([]string{plhu.lkp.Table, "value_parsing_error"}, 1)
			return nil, fmt.Errorf("PinLookupHashUnique.Map: Result value parsing error. %v", err)
		}

		m[fromColKey] = toColValue
	}

	out := make([]key.Destination, 0, len(ids))
	for _, id := range ids {
		idToInt, _ := sqltypes.ToUint64(id)
		val, ok := m[idToInt]
		if !ok {
				out = append(out, key.DestinationNone{})
		} else {
			out = append(out, key.DestinationKeyspaceID(vhash(val)))
		}
	}
	return out, nil
}

// Verify returns true if ids maps to ksids.
func (plhu *PinLookupHashUnique) Verify(vcursor VCursor, ids []sqltypes.Value, ksids [][]byte) ([]bool, error) {
	if hasNullValue(ids) {
		idsToVerify, ksidsToVerify, _ := getThingsToVerify(ids, ksids)
		results, err := plhu.LookupHashUnique.Verify(vcursor, idsToVerify, ksidsToVerify)

		if err != nil {
			return nil, fmt.Errorf("PinLookup.Verify: %v", err)
		}
		return fillInResult(ids, results)
	} else {
		return plhu.LookupHashUnique.Verify(vcursor, ids, ksids)
	}
}

func hasNullValue(ids []sqltypes.Value) bool {
	for _, id := range ids {
		if id.IsNull() {
			return true
		}
	}
	return false
}

func getThingsToVerify(ids []sqltypes.Value, ksids [][]byte) ([]sqltypes.Value, [][]byte, error) {
	idsToVerify := make([]sqltypes.Value, 0, len(ids))
	valuesToVerify := make([][]byte, 0, len(ksids))
	for i, id := range ids {
		if !id.IsNull() {
			idsToVerify = append(idsToVerify, id)
			valuesToVerify = append(valuesToVerify, ksids[i])
		}
	}

	return idsToVerify, valuesToVerify, nil
}

func fillInResult(ids []sqltypes.Value, verifiedResults []bool) ([]bool, error) {
	out := make([]bool, len(ids))
	var idx = 0
	for i, id := range ids {
		if id.IsNull() {
			out[i] = true
		} else {
			out[i] = verifiedResults[idx]
			idx++
		}
	}
	return out, nil
}

// Cost returns the cost of this vindex. It is controlled by a FLAG, 40 for fallback to ScatterCache.
func (plhu *PinLookupHashUnique) Cost() int {
	if decider.CheckDecider("use_pin_lookup_vindex", false) {
		return 10
	} else {
		return 40
	}
}
