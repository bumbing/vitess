package vindexes

// !!!! WORK IN PROGRESS !!!!
// This secondary vindex is experimental / incomplete.
//
// The concept is a unique vindex where an individual vtgate does a scatter query to
// find the keyspace ID of a particular value (like a campaign ID) and then caches the
// associated advertiser GID in memory to avoid repeating the scatter in the near feature.
//
// Not yet implemented:
// - LRU cache. For now, the scatter re-executes for every query. This is bad for performance
//   but it lets us proceed with certain types of validation testing work.
// - vtexplain support. vtexplain returns fake results from tablets to simulate a lookup vindex
//   running. Because scatter_cache is a unique vindex, vtexplain's fake query service needs to
//   return results from only one shard. HandleQuery() in go/vt/vtexplain/vtexplain_vttablet.go
//   needs to be updated to recognize if a scatter_cache vindex is evaluating and only return
//   results from one of the shards, by doing something like this (where t.Shard is the
//	 topodatapdb.Tablet.Shard value copied into the explainTable struct):
//
//     rows := [][]sqltypes.Value{values}
//     directives := sqlparser.ExtractCommentDirectives(selStmt.Comments)
//     if directives.IsSet(sqlparser.DirectiveForceScatter) &&
//             !(t.shard == "0" || strings.HasPrefix(t.shard, "-")) {
//             rows = [][]sqltypes.Value{}
//     }
// - Unit tests
//
// Example declaration of a scatter_cache vindex in a vschema:
//   "campaign_id": {
//     "type": "scatter_cache",
//     "owner": "campaigns",
//     "params": {
//       "table": "campaigns",
//       "from": "id",
//       "to": "g_advertiser_id"
//     }
//   }

import (
	"encoding/json"
	"fmt"
	"strings"

	"vitess.io/vitess/go/sqltypes"
	"vitess.io/vitess/go/vt/key"
	querypb "vitess.io/vitess/go/vt/proto/query"
)

var (
	_ Vindex = (*ScatterCache)(nil)
	_ Lookup = (*ScatterCache)(nil)
)

func init() {
	Register("scatter_cache", NewScatterCache)
}

// ScatterCache defines a vindex that does scatter queries and saves the result
// in an LRU cache. The table is expected to define the id column as unique. It's
// Unique and a Lookup.
type ScatterCache struct {
	name    string
	fromCol string
	toCol   string
	table   string
}

// Inspired by scanBindVar in token.go.
func isAllowedCharInBindVar(ch uint16) bool {
	return 'a' <= ch && ch <= 'z' ||
		'A' <= ch && ch <= 'Z' ||
		ch == '_' ||
		ch == '@' ||
		'0' <= ch && ch <= '9'
}

// NewScatterCache creates a ScatterCache vindex.
// The supplied map has the following required fields:
//   table: name of the backing table. It can be qualified by the keyspace.
//   from: columns in the table that has the 'from' value of the scatter_cache vindex.
//   to: The 'to' column name of the table.
func NewScatterCache(name string, m map[string]string) (Vindex, error) {
	isDisallowedCharacter := func(r rune) bool {
		return !isAllowedCharInBindVar(uint16(r))
	}

	requiredFields := []string{"from", "to", "table"}
	for _, field := range requiredFields {
		if m[field] == "" {
			return nil, fmt.Errorf("scatter_cache: missing required field: %v", field)
		}
		if strings.IndexFunc(m[field], isDisallowedCharacter) != -1 {
			return nil, fmt.Errorf("scatter_cache: %s contains illegal characters: %v", field, m[field])
		}
	}

	sc := &ScatterCache{name: name, fromCol: m["from"], toCol: m["to"], table: m["table"]}

	return sc, nil
}

// String returns the name of the vindex.
func (sc *ScatterCache) String() string {
	return sc.name
}

// Cost returns the cost of this vindex as 30. This is more expensive than a typical
// lookup vindex which would be 10 or 20 depending on whether it's unique.
func (sc *ScatterCache) Cost() int {
	return 30
}

// IsUnique returns true since the Vindex is unique.
func (sc *ScatterCache) IsUnique() bool {
	return true
}

// IsFunctional returns false since the Vindex is not functional.
func (sc *ScatterCache) IsFunctional() bool {
	return false
}

// Map can map ids to key.Destination objects.
func (sc *ScatterCache) Map(vcursor VCursor, ids []sqltypes.Value) ([]key.Destination, error) {
	// NOT YET IMPLEMENTED: Consult an LRU cache before querying.

	// TODO: Run a single big "select toCol, fromCol where fromCol in (... all ids ...)" query instead
	// of many small individual queries. It'll be a little more complicated to put the results back into
	// a slice at the end, but the reduced mysql queries and network round-trips will be worth it.

	// Query for the associated primary keys. The FORCE_SCATTER is necessary to avoid
	// infinitely back into this same scatter_cache vindex and hitting a stack overflow panic.
	sel := fmt.Sprintf(
		"select /*vt+ FORCE_SCATTER=1 */ %s from %s where %s = :%s",
		sc.toCol, sc.table, sc.fromCol, sc.fromCol)

	out := make([]key.Destination, 0, len(ids))
	for _, id := range ids {
		bindVars := map[string]*querypb.BindVariable{
			sc.fromCol: sqltypes.ValueBindVariable(id),
		}
		result, err := vcursor.Execute("VindexScatterCacheLookup", sel, bindVars, false /* isDML */)
		if err != nil {
			return nil, fmt.Errorf("ScatterCache.Map: %v", err)
		}
		switch len(result.Rows) {
		case 0:
			out = append(out, key.DestinationNone{})
		case 1:
			unhashedVal, err := sqltypes.ToUint64(result.Rows[0][0])
			if err != nil {
				return nil, err
			}
			out = append(out, key.DestinationKeyspaceID(vhash(unhashedVal)))
		default:
			return nil, fmt.Errorf("ScatterCache.Map: unexpected multiple results from vindex %s: %v, %v", sc.table, id, result)
		}
	}

	return out, nil
}

// Verify always returns true for scatter-cache.
// In a perfect world it would check that the rows exist in the keyspaces, but verifying a vindex
// happens prior to rows being physically inserted so it's too early for us to be able to verify.
func (sc *ScatterCache) Verify(vcursor VCursor, ids []sqltypes.Value, ksids [][]byte) ([]bool, error) {
	out := make([]bool, len(ids))
	for idx := range ids {
		out[idx] = true
	}
	return out, nil
}

// Create doesn't need to do anything for scatter_cache, but could optimistically populate the cache.
func (sc *ScatterCache) Create(vcursor VCursor, rowsColValues [][]sqltypes.Value, ksids [][]byte, ignoreMode bool) error {
	// TODO: add to the cache
	return nil
}

// Update updates the entry in the vindex.
func (sc *ScatterCache) Update(vcursor VCursor, oldValues []sqltypes.Value, ksid []byte, newValues []sqltypes.Value) error {
	// TODO: update the cache
	return nil
}

// Delete deletes the entry from the vindex.
func (sc *ScatterCache) Delete(vcursor VCursor, rowsColValues [][]sqltypes.Value, ksid []byte) error {
	// TODO: clear the cache
	return nil
}

// MarshalJSON returns a JSON representation of ScatterCache.
func (sc *ScatterCache) MarshalJSON() ([]byte, error) {
	return json.Marshal(sc)
}
