package vindexes

import (
	"fmt"
	"vitess.io/vitess/go/decider"
	"vitess.io/vitess/go/sqltypes"
)

var (
	_ Vindex = (*PinLookupHashUnique)(nil)
	_ Lookup = (*PinLookupHashUnique)(nil)
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
	plhu := &PinLookupHashUnique{LookupHashUnique: lhu.(*LookupHashUnique)}
	return plhu, nil
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
	if decider.CheckDecider("use_pin_lookup_vindex", false){
		return 10
	} else {
		return 40
	}
}
