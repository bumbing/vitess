package vindexes

// NOTE(dweitzman): This is a modified version of hash.go

import (
	"fmt"
	"strconv"

	"vitess.io/vitess/go/sqltypes"
	"vitess.io/vitess/go/vt/key"
	"vitess.io/vitess/go/vt/vtgate/evalengine"
)

var (
	_ SingleColumn = (*HashOffset)(nil)
	_ Reversible   = (*HashOffset)(nil)
)

// HashOffset defines vindex that adds or subtracts a constant to column values
// before delegating to the Hash vindex. At Pinterest, this will be useful for
// the "advertisers" table in the short term because of advertiser local IDs vs
// GIDs. In time we'll want to replace all the local IDs with their GIDs and
// deprecate the concept of GIDs, but for the short term it's useful to have
// the flexibility to use either a local or a gid column when defining a primary
// vindex, since the "advertisers" table doesn't yet use sequences and doesn't
// yet have the ability to set a gid during row creation (just local IDs).
type HashOffset struct {
	hash   Hash
	offset uint64
}

// NewHashOffset creates a new HashOffset.
func NewHashOffset(name string, m map[string]string) (Vindex, error) {
	offsetStr, ok := m["offset"]
	if !ok || offsetStr == "" {
		return nil, fmt.Errorf("hash_offset.NewHashOffset: 'offset' param missing")
	}

	offset, err := strconv.ParseInt(offsetStr, 10, 64)
	if err != nil {
		return nil, fmt.Errorf(`hash_offset.NewHashOffset: "%v" is not a valid offset`, offsetStr)
	}

	return &HashOffset{hash: Hash{name: name}, offset: uint64(offset)}, nil
}

// String returns the name of the vindex.
func (vind *HashOffset) String() string {
	return vind.hash.String()
}

// Cost returns the cost of this index as 1.
func (vind *HashOffset) Cost() int {
	return vind.hash.Cost()
}

// IsUnique returns true since the Vindex is unique.
func (vind *HashOffset) IsUnique() bool {
	return true
}

// IsFunctional returns true since the Vindex is functional.
func (vind *HashOffset) IsFunctional() bool {
	return true
}

func applyOffset(offset uint64, ids []sqltypes.Value) []sqltypes.Value {
	translated := make([]sqltypes.Value, 0, len(ids))
	for _, id := range ids {
		num, err := evalengine.ToUint64(id)
		if err != nil {
			// Not a unint64, so leave it alone.
			// The underlying Hash implementation will ignore nulls or report bad types.
			translated = append(translated, id)
		} else {
			// Apply offset to numbers.
			adjustedNum := num + offset
			translated = append(translated, sqltypes.NewUint64(adjustedNum))
		}
	}
	return translated
}

// Map can map ids to key.Destination objects.
func (vind *HashOffset) Map(cursor VCursor, ids []sqltypes.Value) ([]key.Destination, error) {
	return vind.hash.Map(cursor, applyOffset(vind.offset, ids))
}

// Verify returns true if ids maps to ksids.
func (vind *HashOffset) Verify(cursor VCursor, ids []sqltypes.Value, ksids [][]byte) ([]bool, error) {
	return vind.hash.Verify(cursor, applyOffset(vind.offset, ids), ksids)
}

// ReverseMap returns the ids from ksids.
func (vind *HashOffset) ReverseMap(cursor VCursor, ksids [][]byte) ([]sqltypes.Value, error) {
	vals, err := vind.hash.ReverseMap(cursor, ksids)
	if err != nil {
		return nil, err
	}

	return applyOffset(-vind.offset, vals), nil
}

// NeedsVCursor satisfies the Vindex interface.
func (vind *HashOffset) NeedsVCursor() bool {
	return false
}

func init() {
	Register("hash_offset", NewHashOffset)
}
