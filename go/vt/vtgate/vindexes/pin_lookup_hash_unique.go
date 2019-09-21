package vindexes

import (
	"flag"
	"fmt"
	"strconv"
	"strings"
	"time"
	"vitess.io/vitess/go/cache"
	"vitess.io/vitess/go/decider"
	"vitess.io/vitess/go/sqltypes"
	"vitess.io/vitess/go/stats"
	"vitess.io/vitess/go/sync2"
	"vitess.io/vitess/go/vt/key"
	"vitess.io/vitess/go/vt/log"

	querypb "vitess.io/vitess/go/vt/proto/query"
	topodatapb "vitess.io/vitess/go/vt/proto/topodata"
	vtgatepb "vitess.io/vitess/go/vt/proto/vtgate"
)

const (
	defaultCapacity = 10000
)

var (
	_ Vindex = (*PinLookupHashUnique)(nil)
	_ Lookup = (*PinLookupHashUnique)(nil)

	VindexLookupQuerySize = stats.NewGaugesWithSingleLabel(
		"VindexLookupQuerySize",
		"Number of rows Vindex Lookup need to collect, together from cache and db",
		"TableName")
	VindexLookupCachedSize = stats.NewGaugesWithSingleLabel(
		"VindexLookupCachedSize",
		"Number of rows Vindex Lookup got from cache",
		"TableName")
	failToLookupVindex = stats.NewCountersWithMultiLabels(
		"VindexLookupFailureCount",
		"Count of Map func failure, uses destNone or destAll for all result",
		[]string{"TableName", "FailureReason"})
	batchSize = flag.Int("vindex_lookup_batch_size", 200, "number of rows for each Vindex Lookup query")
	vindexServingVerification = stats.NewCountersWithMultiLabels(
		"VindexServingVerification",
		"The result comparing sampled PinVindex rows with the actual data",
		[]string{"TableName", "Outcome"})
	sampleCheckSize = flag.Int("vindex_sample_check_size", 10, "number of samples to check actual" +
		" data with PinLookupVindex")
)


func RegisterPinVindexCacheStats(getVSchema func() *VSchema) {
	collectStats := func(statFn func(*PinLookupHashUnique) int64) func() map[string]int64 {
		return func() map[string]int64 {
			tstats := make(map[string]int64)

			vschema := getVSchema()
			for name, vindex := range vschema.uniqueVindexes {
				pinVindex, ok := vindex.(*PinLookupHashUnique)
				if !ok {
					continue
				}

				tstats[name] = statFn(pinVindex)
			}

			return tstats
		}
	}

	_ = stats.NewGaugesFuncWithMultiLabels("PinVindexCacheLength", "PinLookupHashUnique cache length", []string{"Vindex"},
		collectStats(func(pinVindex *PinLookupHashUnique) int64 { return pinVindex.keyspaceIDCache.Length() }))
	_ = stats.NewGaugesFuncWithMultiLabels("PinVindexCacheCapacity", "PinLookupHashUnique cache capacity", []string{"Vindex"},
		collectStats(func(pinVindex *PinLookupHashUnique) int64 { return pinVindex.keyspaceIDCache.Capacity() }))
	_ = stats.NewCountersFuncWithMultiLabels("PinVindexCacheEvictions", "PinLookupHashUnique cache evictions", []string{"Vindex"},
		collectStats(func(pinVindex *PinLookupHashUnique) int64 { return pinVindex.keyspaceIDCache.Evictions() }))
	_ = stats.NewCountersFuncWithMultiLabels("PinVindexCacheHits", "PinLookupHashUnique cache hits", []string{"Vindex"},
		collectStats(func(pinVindex *PinLookupHashUnique) int64 { return pinVindex.cacheHits.Get() }))
	_ = stats.NewCountersFuncWithMultiLabels("PinVindexCacheMisses", "PinLookupHashUnique cache missses", []string{"Vindex"},
		collectStats(func(pinVindex *PinLookupHashUnique) int64 { return pinVindex.cacheMisses.Get() }))
}

func init() {
	Register("pin_lookup_hash_unique", NewPinLookupUniqueHash)
}

//====================================================================

// PinLookupHashUnique defines a vindex that uses a lookup table.
// The table is expected to define the id column as unique. It's
// Unique and a Lookup.
type PinLookupHashUnique struct {
	*LookupHashUnique
	keyspaceIDCache *cache.LRUCache
	cacheHits       sync2.AtomicInt64
	cacheMisses     sync2.AtomicInt64
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

	cacheCapacity, err := strconv.ParseUint(m["capacity"], 10, 64)
	if err != nil {
		log.Warning("PinLookupUniqueHash: failed to parse capacity: %v, using the default capacity %v", err, defaultCapacity)
		cacheCapacity = defaultCapacity
	}

	plhu := &PinLookupHashUnique{
		LookupHashUnique: lhu.(*LookupHashUnique),
		keyspaceIDCache:  cache.NewLRUCache(int64(cacheCapacity)),
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

// scatterKeyspaceID is a cache.Value representing a keyspace ID.
type lookupVindexKsID []byte

// Size always returns 1 because we use the cache only to track keyspace IDs.
// This implements the cache.Value interface.
func (lvk lookupVindexKsID) Size() int {
	return 1
}

func (plhu *PinLookupHashUnique) Create(vcursor VCursor, rowsColValues [][]sqltypes.Value, ksids [][]byte, ignoreMode bool) error {
	for idx := range rowsColValues {
		colVals := rowsColValues[idx]
		if len(colVals) != 1 {
			return fmt.Errorf("PinLookupHashUnique.Create: multi-col keys unsupported")
		}
		colVal := colVals[0]
		if !colVal.IsNull() {
			plhu.keyspaceIDCache.Set(colVal.ToString(), lookupVindexKsID(ksids[idx]))
		}
	}

	return plhu.LookupHashUnique.Create(vcursor, rowsColValues, ksids, ignoreMode)
}

func (plhu *PinLookupHashUnique) Update(vcursor VCursor, oldValues []sqltypes.Value, ksid []byte, newValues []sqltypes.Value) error {
	if len(newValues) != 1 {
		return fmt.Errorf("PinLookupHashUnique.Update: multi-col keys unsupported")
	}
	// We skip update cache since in most time, UPDATE DDL are as follow up of an INSERT to update the gid.
	return plhu.LookupHashUnique.Update(vcursor, oldValues, ksid, newValues)
}

// At Pinterest we don't delete data in Patio so far. We will rely on Pepsi script to sync data and Lookup Vindex.
func (plhu *PinLookupHashUnique) Delete(vcursor VCursor, rowsColValues [][]sqltypes.Value, ksid []byte) error {
	return nil
}

// This function will get all mapping by sending one query, to avoid using up all connections like LookupHashUnique's
// implementation.
// If failed to lookup Vindex, it could possible return destAll, which will not work for DMLs.
// For DMLs, only INSERT could possible use Map, but that's for primary vindex, not for secondary vindex.
// The fallback destAll is a protection for SELECTIN and SELECTEQUAL query.
func (plhu *PinLookupHashUnique) Map(cursor VCursor, ids []sqltypes.Value) ([]key.Destination, error) {
	if plhu.writeOnly {
		return plhu.getScatterResultForAll(len(ids)), nil
	}

	statsKey := []string{plhu.name, "map"}
	defer scatterCacheTimings.Record(statsKey, time.Now())

	sel := fmt.Sprintf(
		"select %s, %s from %s where %s in ::%s",
		plhu.lkp.FromColumns[0], plhu.lkp.To, plhu.lkp.Table, plhu.lkp.FromColumns[0], plhu.lkp.FromColumns[0])

	// HashMap to keep the relationship between Ids and KeyspaceId, in case result is returned in different order.
	// Choose both uint64 as key and value, since sqlTypes.Value is not comparable. And by parsing result column, we
	// can also check if data is corrupted.
	// First get cached keys, then get lookup keys.
	m := make(map[uint64]uint64, len(ids))

	// Temporary placeholder for querypb.Value.
	values := make([]querypb.Value, 0, len(ids))

	for _, id := range ids {
		// Directly use cache keys without lookup.
		val, ok := plhu.keyspaceIDCache.Get(id.ToString())
		if ok {
			plhu.cacheHits.Add(1)
			k, err := sqltypes.ToUint64(id)
			if err != nil {
				return nil, fmt.Errorf("PinLookupHashUnique.Map: failed to convert key %v", id.ToString())
			}
			v, err := vunhash(val.(lookupVindexKsID))
			if err != nil {
				return nil, fmt.Errorf("PinLookupHashUnique.Map: failed to unhash keyspaceID %v", err)
			}
			m[k] = v
			continue
		}
		plhu.cacheMisses.Add(1)
		temp := sqltypes.ValueBindVariable(id)
		v := querypb.Value{
			Type:  temp.Type,
			Value: temp.Value,
		}
		values = append(values, v)
	}

	bv := &querypb.BindVariable{
		Type:   querypb.Type_TUPLE,
		Values: make([]*querypb.Value, 0, *batchSize),
	}

	VindexLookupCachedSize.Add(plhu.lkp.Table, int64(len(m)))
	VindexLookupQuerySize.Add(plhu.lkp.Table, int64(len(values)))
	for i := 0; i < len(values); i++ {
		bv.Values = append(bv.Values, &values[i])
		// When reached batch limit or is the last value, execute the query
		if len(bv.Values) == *batchSize || i == len(values)-1 {
			bindVars := map[string]*querypb.BindVariable{
				plhu.lkp.FromColumns[0]: bv,
			}

			// Execution and error handling
			queryResult, err := cursor.Execute(
				"PinLookupHashUnique.Lookup", sel, bindVars, false /* isDML */, vtgatepb.CommitOrder_AUTOCOMMIT)
			if err != nil {
				failToLookupVindex.Add([]string{plhu.lkp.Table, "select_query_failure"}, 1)
				return nil, fmt.Errorf("PinLookupHashUnique.Map: Select query execution error. %v", err)
			}

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
				plhu.keyspaceIDCache.Set(row[0].ToString(), lookupVindexKsID(vhash(toColValue)))
			}
			// Clean bind variables for reuse
			bv.Values = bv.Values[:0]
		}
	}

	out := make([]key.Destination, 0, len(ids))
	for _, id := range ids {
		idToInt, _ := sqltypes.ToUint64(id)
		val, ok := m[idToInt]
		if !ok {
			log.Error("Found mismatch for id: ", id)
			out = append(out, key.DestinationNone{})
		} else {
			out = append(out, key.DestinationKeyspaceID(vhash(val)))
		}
	}

	// Set upperbound to 10 for sampling check
	checkSize := len(ids)
	if checkSize > *sampleCheckSize {
		checkSize = *sampleCheckSize
	}
	if decider.CheckDecider("pinvindex_sample_check", false) {
		plhu.checkSample(cursor, ids[:checkSize], out[:checkSize])
	}
	return out, nil
}

func getSourceTable(name string) string {
	temp := strings.TrimSuffix(name, "_id_idx")
	if strings.HasSuffix(temp, "s") || strings.HasSuffix(temp, "_history") {
		return temp
	} else {
		return temp+"s"
	}
}

func (plhu *PinLookupHashUnique) checkSample(vcursor VCursor, ids []sqltypes.Value, expected []key.Destination) {
	sel := fmt.Sprintf(
		"select /*vt+ FORCE_SCATTER=1 */ %s, %s from %s where %s in ::%s",
		plhu.lkp.FromColumns[0], plhu.lkp.To, getSourceTable(plhu.lkp.Table), plhu.lkp.FromColumns[0], plhu.lkp.FromColumns[0])

	idsInInterface := make([]interface{}, len(ids))
	for i, id := range ids {
		idsInInterface[i] = id
	}
	bv, err := sqltypes.BuildBindVariable(idsInInterface)
	if err != nil {
		vindexServingVerification.Add([]string{plhu.name, "fail_to_bindvar"}, 1)
		return
	}
	bindVars := map[string]*querypb.BindVariable{
		plhu.lkp.FromColumns[0]: bv,
	}
	queryResult, err := vcursor.Execute("PinLookupHashUniqueCheckSample", sel, bindVars, false /* isDML */, vtgatepb.CommitOrder_NORMAL)
	if err != nil {
		vindexServingVerification.Add([]string{plhu.name, "fail_to_lookup"}, 1)
		return
	}

	for i, row := range queryResult.Rows {
		if len(row) != 2 {
			_ = fmt.Sprintf("PinLookupHashUnique.checkSample: Internal error. Expected %v columns. Got %v", 2, len(row))
			vindexServingVerification.Add([]string{plhu.name, "result_length_mismatch"}, 1)
			continue
		}

		toColValue, err := sqltypes.ToUint64(row[1])
		if err != nil {
			vindexServingVerification.Add([]string{plhu.name, "fail_to_parse_result"}, 1)
			continue
		}
		if key.DestinationKeyspaceID(vhash(toColValue)).String() != expected[i].String() {
			vindexServingVerification.Add([]string{plhu.name, "result_mismatch"}, 1)
		} else {
			vindexServingVerification.Add([]string{plhu.name, "result_match"}, 1)
		}
	}
}

// Verify returns true if ids maps to ksids.
func (plhu *PinLookupHashUnique) Verify(vcursor VCursor, ids []sqltypes.Value, ksids [][]byte) ([]bool, error) {
	idsToVerify, ksidsToVerify, preVerifyResults, _ := getThingsToVerify(ids, ksids)
	results, err := plhu.LookupHashUnique.Verify(vcursor, idsToVerify, ksidsToVerify)

	if err != nil {
		return nil, fmt.Errorf("PinLookup.Verify: %v", err)
	}
	return fillInResult(results, preVerifyResults)
}

func getThingsToVerify(ids []sqltypes.Value, ksids [][]byte) ([]sqltypes.Value, [][]byte, []bool, error) {
	idsToVerify := make([]sqltypes.Value, 0, len(ids))
	valuesToVerify := make([][]byte, 0, len(ksids))
	preVerifyResults := make([]bool, len(ids))
	for i, id := range ids {
		if id.IsNull() {
			preVerifyResults[i] = true
			continue
		}

		val, err := sqltypes.ToUint64(id)
		// This failure should most like due to Patio-latest DML query pass in negative value for unowned Vindex.
		// We bypass the verification for to not abort the DML and print a warning instead.
		if err != nil {
			log.Warningf("PinLookupHashUnique.Verify: Failed to parse Id %v, error: %v", id, err)
			preVerifyResults[i] = true
			continue
		}

		// ID should be value bigger than 0, but it is not the case when some patio-latest DML query does not populate
		// their foreign id field, or some potential corrupted data written to Patio-prod (we have not seen it so far).
		// PinVindex.Verify will bypass the abnormal rows. It is only responsible for making sure the correct Id
		// can get correct routing information correspondingly.
		// The data correctness should be taken care of on application side.
		if val == 0 {
			preVerifyResults[i] = true
			continue
		}

		preVerifyResults[i] = false
		idsToVerify = append(idsToVerify, id)
		valuesToVerify = append(valuesToVerify, ksids[i])
	}

	return idsToVerify, valuesToVerify, preVerifyResults, nil
}

func fillInResult(verifiedResults []bool, preVerifyResults []bool) ([]bool, error) {
	var idx = 0
	for i, verified := range preVerifyResults {
		if verified != true {
			preVerifyResults[i] = verifiedResults[idx]
			idx++
		}
	}
	return preVerifyResults, nil
}

// Cost returns the cost of this vindex. It is controlled by a FLAG, 40 for fallback to ScatterCache.
func (plhu *PinLookupHashUnique) Cost() int {
	if decider.CheckDecider("use_pin_lookup_vindex", false) {
		return 10
	} else {
		return 40
	}
}
