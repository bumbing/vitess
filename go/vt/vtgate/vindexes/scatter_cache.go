package vindexes

// This secondary vindex is experimental / incomplete.
//
// The concept is a unique vindex where an individual vtgate does a scatter query to
// find the keyspace ID of a particular value (like a campaign ID) and then caches the
// associated advertiser GID in memory to avoid repeating the scatter in the near feature.
//
// Not yet implemented:
// - Run a single big scatter query instead of many small queries in serial
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
//       "capacity": "1000",
//       "table": "campaigns",
//       "from": "id",
//       "to": "g_advertiser_id"
//     }
//   }

import (
	"bytes"
	"fmt"
	"strconv"
	"strings"
	"time"

	"vitess.io/vitess/go/cache"
	"vitess.io/vitess/go/sqltypes"
	"vitess.io/vitess/go/stats"
	"vitess.io/vitess/go/sync2"
	"vitess.io/vitess/go/vt/key"
	querypb "vitess.io/vitess/go/vt/proto/query"
	topodatapb "vitess.io/vitess/go/vt/proto/topodata"
	vtgatepb "vitess.io/vitess/go/vt/proto/vtgate"
)

var (
	_ Vindex = (*ScatterCache)(nil)
	_ Lookup = (*ScatterCache)(nil)

	scatterCacheTimings = stats.NewMultiTimings(
		"VindexTimings",
		"Vindex timings",
		[]string{"Name", "Operation"})
)

// RegisterScatterCacheStats arranges for scatter cache stats to be available given a way to fetch the
// current vschema.
func RegisterScatterCacheStats(getVSchema func() *VSchema) {
	collectStats := func(statFn func(*ScatterCache) int64) func() map[string]int64 {
		return func() map[string]int64 {
			tstats := make(map[string]int64)

			vschema := getVSchema()
			for name, vindex := range vschema.uniqueVindexes {
				scatterVindex, ok := vindex.(*ScatterCache)
				if !ok {
					continue
				}

				tstats[name] = statFn(scatterVindex)
			}

			return tstats
		}
	}

	_ = stats.NewGaugesFuncWithMultiLabels("ScatterCacheLength", "scatter cache length", []string{"Vindex"},
		collectStats(func(scatterCache *ScatterCache) int64 { return scatterCache.keyspaceIDCache.Length() }))
	_ = stats.NewGaugesFuncWithMultiLabels("ScatterCacheCapacity", "scatter cache capacity", []string{"Vindex"},
		collectStats(func(scatterCache *ScatterCache) int64 { return scatterCache.keyspaceIDCache.Capacity() }))
	_ = stats.NewCountersFuncWithMultiLabels("ScatterCacheEvictions", "scatter cache evictions", []string{"Vindex"},
		collectStats(func(scatterCache *ScatterCache) int64 { return scatterCache.keyspaceIDCache.Evictions() }))
	_ = stats.NewCountersFuncWithMultiLabels("ScatterCacheHits", "scatter cache hits", []string{"Vindex"},
		collectStats(func(scatterCache *ScatterCache) int64 { return scatterCache.cacheHits.Get() }))
	_ = stats.NewCountersFuncWithMultiLabels("ScatterCacheMisses", "scatter cache missses", []string{"Vindex"},
		collectStats(func(scatterCache *ScatterCache) int64 { return scatterCache.cacheMisses.Get() }))
}

func init() {
	Register("scatter_cache", NewScatterCache)
}

// ScatterCache defines a vindex that does scatter queries and saves the result
// in an LRU cache. The table is expected to define the id column as unique. It's
// Unique and a Lookup.
type ScatterCache struct {
	name            string
	fromCol         string
	toCol           string
	table           string
	keyspaceIDCache *scatterLRUCache
	cacheHits       sync2.AtomicInt64
	cacheMisses     sync2.AtomicInt64
}

// scatterLRU is a thread-safe object for remembering the keyspace ID of recently-searched
// secondary IDs.
type scatterLRUCache struct {
	*cache.LRUCache
}

// scatterKeyspaceID is a cache.Value representing a keyspace ID.
type scatterKeyspaceID []byte

// Size always returns 1 because we use the cache only to track keyspace IDs.
// This implements the cache.Value interface.
func (ski scatterKeyspaceID) Size() int {
	return 1
}

// newScatterLRUCache creates a new cache with the given capacity.
func newScatterLRUCache(capacity int64) *scatterLRUCache {
	return &scatterLRUCache{cache.NewLRUCache(capacity)}
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

	requiredFields := []string{"from", "to", "table", "capacity"}
	for _, field := range requiredFields {
		if m[field] == "" {
			return nil, fmt.Errorf("scatter_cache: missing required field: %v", field)
		}
		if strings.IndexFunc(m[field], isDisallowedCharacter) != -1 {
			return nil, fmt.Errorf("scatter_cache: %s contains illegal characters: %v", field, m[field])
		}
	}

	capacity, err := strconv.ParseUint(m["capacity"], 10, 64)
	if err != nil {
		return nil, fmt.Errorf("scatter_cache: failed to parse capacity: %v", err)
	}

	sc := &ScatterCache{
		name:            name,
		fromCol:         m["from"],
		toCol:           m["to"],
		table:           m["table"],
		keyspaceIDCache: newScatterLRUCache(int64(capacity)),
	}

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

// CanVerifyNull returns true if null values can be verified.
func (sc *ScatterCache) CanVerifyNull() bool {
	return true
}

func (sc *ScatterCache) scatterLookupIds(vcursor VCursor, foundIds map[string]scatterKeyspaceID, missingIds []interface{}) error {
	// Query for the associated primary keys. The FORCE_SCATTER is necessary to avoid
	// infinitely back into this same scatter_cache vindex and hitting a stack overflow panic.
	sel := fmt.Sprintf(
		"select /*vt+ FORCE_SCATTER=1 */ %s, %s from %s where %s in ::%s",
		sc.fromCol, sc.toCol, sc.table, sc.fromCol, sc.fromCol)

	missingIdsBV, err := sqltypes.BuildBindVariable(missingIds)
	if err != nil {
		return err
	}

	bindVars := map[string]*querypb.BindVariable{
		sc.fromCol: missingIdsBV,
	}
	queryResult, err := vcursor.Execute("VindexScatterCacheLookup", sel, bindVars, false /* isDML */, vtgatepb.CommitOrder_NORMAL)
	if err != nil {
		return fmt.Errorf("ScatterCache.Map: %v", err)
	}

	// The simple thing would be to use the query results to prime the cache and
	// then read from cache, but just in case the capacity is super small or something
	// insane is happening with fast cache evictions we'll also read the results into
	// a temporary map to use for serving this request.

	for _, row := range queryResult.Rows {
		if len(row) != 2 {
			return fmt.Errorf("ScatterCache.Map: Internal error. Expected %v columns. Got %v", 2, len(row))
		}

		fromColKey := row[0].ToString()
		toColValue, err := sqltypes.ToUint64(row[1])
		if err != nil {
			return err
		}

		destinationKeyspace := scatterKeyspaceID(vhash(toColValue))

		// Sanity check: if there's an existing value in the foundIds, it must have
		// the same destination keyspace ID. If not, the mapping from column values
		// to keyspace IDs is not globally unique and we have a big problem.
		prevKeyspaceID, ok := foundIds[fromColKey]
		if ok && bytes.Equal(prevKeyspaceID, destinationKeyspace) {
			return fmt.Errorf("ScatterCache.Map: unexpected multiple results from vindex %v, key %v", sc.table, fromColKey)
		}

		// Add to the cache and to the map for this query.
		foundIds[fromColKey] = destinationKeyspace
		sc.keyspaceIDCache.Set(fromColKey, destinationKeyspace)
	}
	return nil
}

func (sc *ScatterCache) findUncachedIds(ids []sqltypes.Value) (foundIds map[string]scatterKeyspaceID, missingIds []interface{}) {
	foundIds = make(map[string]scatterKeyspaceID)
	missingIds = make([]interface{}, 0, len(ids))
	for _, id := range ids {
		if id.IsNull() {
			// null lookup values are nevered mapped to a keyspace ID and never cached.
			continue
		}

		cacheKey := id.ToString()
		cachedKeyspaceID, ok := sc.keyspaceIDCache.Get(cacheKey)
		if ok {
			sc.cacheHits.Add(1)
			foundIds[cacheKey] = cachedKeyspaceID.(scatterKeyspaceID)
			continue
		}

		sc.cacheMisses.Add(1)
		missingIds = append(missingIds, id)
	}
	return
}

// Map can map ids to key.Destination objects.
func (sc *ScatterCache) Map(vcursor VCursor, ids []sqltypes.Value) ([]key.Destination, error) {
	statsKey := []string{sc.name, "map"}
	defer scatterCacheTimings.Record(statsKey, time.Now())

	if sc.keyspaceIDCache.Capacity() == 0 {
		// Degenerate case: just force a scatter
		out := make([]key.Destination, 0, len(ids))
		for range ids {
			out = append(out, key.DestinationKeyRange{KeyRange: &topodatapb.KeyRange{}})
		}
		return out, nil
	}

	foundIds, missingIds := sc.findUncachedIds(ids)
	if len(missingIds) > 0 {
		err := sc.scatterLookupIds(vcursor, foundIds, missingIds)
		if err != nil {
			return nil, err
		}
	}

	out := make([]key.Destination, 0, len(ids))
	for _, id := range ids {
		cachedKeyspaceID, ok := foundIds[id.ToString()]
		if ok {
			out = append(out, key.DestinationKeyspaceID(cachedKeyspaceID))
		} else {
			out = append(out, key.DestinationNone{})
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

// Create for scatter cache populate's the current vtgate's cache value.
// This won't help across vtgates (queries on a different gate will still need a scatter),
// but at Pinterest we often have create transactions that do an insert followed by an
// update and with the mysql protocol they'll go though the same vtgate, so this
// heuristic may help in many cases in practice.
func (sc *ScatterCache) Create(vcursor VCursor, rowsColValues [][]sqltypes.Value, ksids [][]byte, ignoreMode bool) error {
	if sc.keyspaceIDCache.Capacity() == 0 {
		return nil
	}

	if len(rowsColValues) != len(ksids) {
		return fmt.Errorf("ScatterCache.Create: internal error. %v col values != %v keyspace IDs", len(rowsColValues), len(ksids))
	}

	for idx := range rowsColValues {
		colVals := rowsColValues[idx]
		if len(colVals) != 1 {
			return fmt.Errorf("ScatterCache.Create: multi-col create unsupported")
		}
		colVal := colVals[0]
		if !colVal.IsNull() {
			sc.keyspaceIDCache.Set(colVal.ToString(), scatterKeyspaceID(ksids[idx]))
		}
	}
	return nil
}

// Update updates the entry in the vindex.
func (sc *ScatterCache) Update(vcursor VCursor, oldValues []sqltypes.Value, ksid []byte, newValues []sqltypes.Value) error {
	// scatter_cache is fundamentally incompatible with ids being updated to
	// point at a different keyspace. Entity IDs should never be reused and should
	// never transition from one advertiser to another. If there were some emergency
	// need to reassign an entity to a new advertiser (which, again, should not happen)
	// it would be necessary to use ApplyVSchema to re-create the scatter caches on
	// all the vtgates.

	return nil
}

// Delete deletes the entry from the vindex.
func (sc *ScatterCache) Delete(vcursor VCursor, rowsColValues [][]sqltypes.Value, ksid []byte) error {
	// See the comment for Update() above.

	return nil
}

// NeedsVCursor satisfies the Vindex interface.
func (sc *ScatterCache) NeedsVCursor() bool {
	return true
}
