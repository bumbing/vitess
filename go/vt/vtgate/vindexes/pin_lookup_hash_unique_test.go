package vindexes

import (
	"errors"
	"fmt"
	"reflect"
	"strconv"
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
	plhu := createPinLookup(t, "pin_lookup_hash_unique", false, 0)
	pvc := &pinVcursor{numRows: 1}

	got, err := plhu.(SingleColumn).Map(pvc, []sqltypes.Value{sqltypes.NewInt64(1)})
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
	got, err = plhu.(SingleColumn).Map(pvc, []sqltypes.Value{sqltypes.NewInt64(1)})
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
	_, err = plhu.(SingleColumn).Map(pvc, []sqltypes.Value{sqltypes.NewInt64(1)})
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
	got, err = plhu.(SingleColumn).Map(pvc, []sqltypes.Value{sqltypes.NewInt64(1)})
	wantErr = "PinLookupHashUnique.Map: Result key parsing error. Code: INVALID_ARGUMENT\ncould not parse value: 'a'\n"
	if err == nil || err.Error() != wantErr {
		t.Errorf("plhu(parsing fail) err: %v, want %s", err, wantErr)
	}
	if got != nil {
		t.Errorf("Map(): %#v, want %+v", got, want)
	}

	// Test query fail.
	pvc.mustFail = true
	_, err = plhu.(SingleColumn).Map(pvc, []sqltypes.Value{sqltypes.NewInt64(1)})
	wantErr = "PinLookupHashUnique.Map: Select query execution error. execute failed"
	if err == nil || err.Error() != wantErr {
		t.Errorf("plhu(query fail) err: %v, want %s", err, wantErr)
	}
	pvc.mustFail = false
}

func TestPinLookupHashUniqueMapWriteOnly(t *testing.T) {
	plhu := createPinLookup(t, "pin_lookup_hash_unique", true, 0)
	pvc := &pinVcursor{numRows: 0}

	got, err := plhu.(SingleColumn).Map(pvc, []sqltypes.Value{sqltypes.NewInt64(1)})
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
	plhu := createPinLookup(t, "pin_lookup_hash_unique", false, 0)
	pvc := &pinVcursor{numRows: 1}

	// regular test copied from lookup_hash_unique_test.go
	got, err := plhu.(SingleColumn).Verify(pvc,
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
	got, err = plhu.(SingleColumn).Verify(pvc, []sqltypes.Value{sqltypes.NewInt64(1)}, [][]byte{[]byte("\x16k@\xb4J\xbaK\xd6")})
	if err != nil {
		t.Error(err)
	}
	want = []bool{false}
	if !reflect.DeepEqual(got, want) {
		t.Errorf("plhu.Verify(mismatch): %v, want %v", got, want)
	}

	_, err = plhu.(SingleColumn).Verify(pvc, []sqltypes.Value{sqltypes.NewInt64(1)}, [][]byte{[]byte("bogus")})
	wantErr := "PinLookup.Verify: lookup.Verify.vunhash: invalid keyspace id: 626f677573"
	if err == nil || err.Error() != wantErr {
		t.Errorf("plhu.Verify(bogus) err: %v, want %s", err, wantErr)
	}

	// Null value id Verify should pass Verify
	got, err = plhu.(SingleColumn).Verify(pvc, []sqltypes.Value{sqltypes.NULL}, [][]byte{[]byte("\x16k@\xb4J\xbaK\xd6")})
	if err != nil {
		t.Error(err)
	}
	want = []bool{true}
	if !reflect.DeepEqual(got, want) {
		t.Errorf("plhu.Verify(match): %v, want %v", got, want)
	}

	got, err = plhu.(SingleColumn).Verify(pvc, []sqltypes.Value{sqltypes.NewInt64(0)}, [][]byte{[]byte("\x16k@\xb4J\xbaK\xd6")})
	if err != nil {
		t.Error(err)
	}
	want = []bool{true}
	if !reflect.DeepEqual(got, want) {
		t.Errorf("plhu.Verify(match): %v, want %v", got, want)
	}
}

func TestPinLookupHashUniqueCache(t *testing.T) {
	plhu := createPinLookup(t, "pin_lookup_hash_unique", false, 100)
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
	_, err := plhu.(SingleColumn).Map(pvc, []sqltypes.Value{sqltypes.NewInt64(1)})
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

	got, err := plhu.(SingleColumn).Map(pvc, []sqltypes.Value{sqltypes.NewInt64(2)})
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
					Type:   querypb.Type_TUPLE,
					Values: []*querypb.Value{},
				},
			},
		},
	}
	if !reflect.DeepEqual(pvc.queries, wantqueries) {
		t.Errorf("lookup.Create queries:\n%v, want\n%v", pvc.queries, wantqueries)
	}
}

func TestCheckSample(t *testing.T) {
	plhu := createPinLookup(t, "pin_lookup_hash_unique", false, 100)
	pvc := &pinVcursor{numRows: 1}
	input := []sqltypes.Value{sqltypes.NewInt64(1)}
	correctResult := []key.Destination{key.DestinationKeyspaceID([]byte("\x16k@\xb4J\xbaK\xd6"))}
	wrongResult := []key.Destination{key.DestinationKeyspaceID([]byte("\x16k@\xb4J"))}

	plhu.(*PinLookupHashUnique).checkSample(pvc, input, correctResult)
	if vindexServingVerification.Counts()["pin_lookup_hash_unique.result_match"] != 1 {
		t.Errorf("TestCheckSample correct result test wrong")
	}

	vindexServingVerification.ResetAll()
	plhu.(*PinLookupHashUnique).checkSample(pvc, input, wrongResult)
	if vindexServingVerification.Counts()["pin_lookup_hash_unique.result_mismatch"] != 1 {
		t.Errorf("TestCheckSample wrong result test wrong")
	}

	vindexServingVerification.ResetAll()
	wrongSizePvc := &pinVcursor{numRows: 2}
	plhu.(*PinLookupHashUnique).checkSample(wrongSizePvc, input, correctResult)
	if vindexServingVerification.Counts()["pin_lookup_hash_unique.result_match"] != 1 {
		t.Errorf("TestCheckSample wrong size test result wrong")
	}
	if vindexServingVerification.Counts()["pin_lookup_hash_unique.result_mismatch"] != 1 {
		t.Errorf("TestCheckSample wrong size test result wrong")
	}
}

func TestGetSourceTable(t *testing.T) {
	data := map[string]string{
		"accepted_tos_id_idx":                            "accepted_tos",
		"ad_group_id_idx":                                "ad_groups",
		"ad_group_spec_id_idx":                           "ad_group_specs",
		"ad_groups_history_id_idx":                       "ad_groups_history",
		"advertiser_conversion_event_id_idx":             "advertiser_conversion_events",
		"advertiser_discount_id_idx":                     "advertiser_discounts",
		"advertiser_labeling_result_id_idx":              "advertiser_labeling_results",
		"app_event_tracking_config_id_idx":               "app_event_tracking_configs",
		"bill_detail_id_idx":                             "bill_details",
		"bill_id_idx":                                    "bills",
		"billing_action_id_idx":                          "billing_actions",
		"billing_contact_id_idx":                         "billing_contacts",
		"billing_profile_id_idx":                         "billing_profiles",
		"bulk_v2_job_id_idx":                             "bulk_v2_jobs",
		"business_profile_id_idx":                        "business_profiles",
		"campaign_id_idx":                                "campaigns",
		"campaign_spec_id_idx":                           "campaign_specs",
		"campaigns_history_id_idx":                       "campaigns_history",
		"carousel_slot_promotion_id_idx":                 "carousel_slot_promotions",
		"conversion_tag_id_idx":                          "conversion_tags",
		"goal_id_idx":                                    "goals",
		"notification_id_idx":                            "notifications",
		"order_line_id_idx":                              "order_lines",
		"order_line_spec_id_idx":                         "order_line_specs",
		"pin_promotion_id_idx":                           "pin_promotions",
		"pin_promotion_label_id_idx":                     "pin_promotion_labels",
		"pin_promotion_spec_id_idx":                      "pin_promotion_specs",
		"pin_promotions_history_id_idx":                  "pin_promotions_history",
		"pinner_list_id_idx":                             "pinner_lists",
		"pinner_list_spec_id_idx":                        "pinner_list_specs",
		"product_group_id_idx":                           "product_groups",
		"product_group_spec_id_idx":                      "product_group_specs",
		"promoted_catalog_product_group_id_idx":          "promoted_catalog_product_groups",
		"promoted_catalog_product_groups_history_id_idx": "promoted_catalog_product_groups_history",
		"rule_subscription_id_idx":                       "rule_subscriptions",
		"targeting_attribute_history_id_idx":             "targeting_attribute_history",
		"targeting_attribute_id_idx":                     "targeting_attributes",
		"targeting_spec_id_idx":                          "targeting_specs",
		"user_preference_id_idx":                         "user_preferences",
	}
	for key := range data {
		result := getSourceTable(key)
		if strings.Compare(data[key], result) != 0 {
			t.Errorf("Given table %v, expected %v, got %v", key, data[key], result)
		}
	}
}

func createPinLookup(t *testing.T, name string, writeOnly bool, cacheCapacity int) Vindex {
	t.Helper()
	write := "false"
	if writeOnly {
		write = "true"
	}
	l, err := CreateVindex(name, name, map[string]string{
		"table":      "t",
		"from":       "fromc",
		"to":         "toc",
		"capacity":   strconv.Itoa(cacheCapacity),
		"write_only": write,
	})
	if err != nil {
		t.Fatal(err)
	}
	return l
}
