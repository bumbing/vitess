package main

import (
	"testing"

	"github.com/golang/protobuf/proto"
	querypb "vitess.io/vitess/go/vt/proto/query"
	vschemapb "vitess.io/vitess/go/vt/proto/vschema"
)

var advertisersColumns = []*vschemapb.Column{
	{Name: "id"},
	{Name: "active"},
	{Name: "creation_date"},
	{Name: "name", Type: querypb.Type_VARCHAR},
	{Name: "owner_user_id"},
	{Name: "billing_type", Type: querypb.Type_VARCHAR},
	{Name: "billing_token", Type: querypb.Type_VARCHAR},
	{Name: "billing_profile_id"},
	{Name: "billing_threshold"},
	{Name: "test_account"},
	{Name: "gid"},
	{Name: "is_spam"},
	{Name: "properties"},
	{Name: "deleted"},
	{Name: "country"},
	{Name: "currency"},
	{Name: "business_profile_id"},
	{Name: "updated_time"},
	{Name: "g_billing_profile_id"},
	{Name: "g_business_profile_id"},
	{Name: "daily_spend_cap"},
}

var acceptedTosColumns = []*vschemapb.Column{
	{Name: "advertiser_id"},
	{Name: "g_advertiser_id"},
	{Name: "tos_id"},
	{Name: "accept_date"},
	{Name: "properties"},
	{Name: "id"},
}

var campaignsColumns = []*vschemapb.Column{
	{Name: "id"},
	{Name: "active"},
	{Name: "creation_date"},
	{Name: "campaign_spec_id"},
	{Name: "advertiser_id"},
	{Name: "name", Type: querypb.Type_VARCHAR},
	{Name: "unique_line_count_id"},
	{Name: "action_type"},
	{Name: "gid"},
	{Name: "g_campaign_spec_id"},
	{Name: "g_advertiser_id"},
	{Name: "properties"},
	{Name: "creative_type"},
	{Name: "objective_type"},
	{Name: "updated_time"},
}

var adGroupsColumns = []*vschemapb.Column{
	{Name: "id"},
	{Name: "gid"},
	{Name: "active"},
	{Name: "creation_date"},
	{Name: "spec_id"},
	{Name: "campaign_id"},
	{Name: "properties"},
	{Name: "updated_time"},
	{Name: "g_campaign_id"},
	{Name: "g_spec_id"},
	{Name: "advertiser_id"},
	{Name: "g_advertiser_id"},
}

var targetingAttributeCountsByAdvertiserColumns = []*vschemapb.Column{
	{Name: "advertiser_gid"},
	{Name: "active_keywords_count"},
	{Name: "advertiser_id"},
}

func TestVSchemaOriginal(t *testing.T) {
	ddls, err := parseSchema(ddls)
	if err != nil {
		t.Error(err)
	}

	config := pinschemaConfig{}
	got, err := newVschemaBuilder(ddls, config).ddlsToVSchema()
	if err != nil {
		t.Error(err)
	}
	want := &vschemapb.Keyspace{
		Tables: map[string]*vschemapb.Table{
			"accepted_tos": {},
			"advertisers":  {},
			"campaigns":    {},
			"ad_groups":    {},
			"targeting_attribute_counts_by_advertiser": {},
		},
	}
	if !proto.Equal(got, want) {
		t.Errorf("GetVSchema: %s, want %s", got, want)
	}
}

func TestVSchemaAuthoritative(t *testing.T) {
	ddls, err := parseSchema(ddls)
	if err != nil {
		t.Error(err)
	}

	config := pinschemaConfig{
		colsAuthoritative: true,
		includeCols:       true,
	}
	got, err := newVschemaBuilder(ddls, config).ddlsToVSchema()
	if err != nil {
		t.Error(err)
	}
	want := &vschemapb.Keyspace{
		Tables: map[string]*vschemapb.Table{
			"accepted_tos": {
				Columns:                 acceptedTosColumns,
				ColumnListAuthoritative: true,
			},
			"advertisers": {
				Columns:                 advertisersColumns,
				ColumnListAuthoritative: true,
			},
			"campaigns": {
				Columns:                 campaignsColumns,
				ColumnListAuthoritative: true,
			},
			"ad_groups": {
				Columns:                 adGroupsColumns,
				ColumnListAuthoritative: true,
			},
			"targeting_attribute_counts_by_advertiser": {
				Columns:                 targetingAttributeCountsByAdvertiserColumns,
				ColumnListAuthoritative: true,
			},
		},
	}
	if !proto.Equal(got, want) {
		t.Errorf("GetVSchema: %s, want %s", got, want)
	}
}

func TestVSchemaSequences(t *testing.T) {
	ddls, err := parseSchema(ddls)
	if err != nil {
		t.Error(err)
	}

	config := pinschemaConfig{createSeq: true}
	got, err := newVschemaBuilder(ddls, config).ddlsToVSchema()
	if err != nil {
		t.Error(err)
	}
	want := &vschemapb.Keyspace{
		Tables: map[string]*vschemapb.Table{
			"accepted_tos": {
				AutoIncrement: &vschemapb.AutoIncrement{
					Column:   "id",
					Sequence: "accepted_tos_seq",
				},
			},
			"advertisers": {
				AutoIncrement: &vschemapb.AutoIncrement{
					Column:   "id",
					Sequence: "advertisers_seq",
				},
			},
			"campaigns": {
				AutoIncrement: &vschemapb.AutoIncrement{
					Column:   "id",
					Sequence: "campaigns_seq",
				},
			},
			"ad_groups": {
				AutoIncrement: &vschemapb.AutoIncrement{
					Column:   "id",
					Sequence: "ad_groups_seq",
				},
			},
			"targeting_attribute_counts_by_advertiser": {},
		},
	}
	if !proto.Equal(got, want) {
		t.Errorf("GetVSchema: %s, want %s", got, want)
	}
}

func TestVSchemaPrimaryVindex(t *testing.T) {
	ddls, err := parseSchema(ddls)
	if err != nil {
		t.Error(err)
	}

	config := pinschemaConfig{createPrimary: true}
	got, err := newVschemaBuilder(ddls, config).ddlsToVSchema()
	if err != nil {
		t.Error(err)
	}
	want := &vschemapb.Keyspace{
		Sharded: true,
		Vindexes: map[string]*vschemapb.Vindex{
			"advertiser_id": {
				Type: "hash_offset",
				Params: map[string]string{
					"offset": "549755813888",
				},
			},
			"dark_write_advertiser_id": {
				Type: "hash_offset",
				Params: map[string]string{
					"offset": "549755813888",
				},
			},
			"g_advertiser_id": {
				Type: "hash",
			},
		},
		Tables: map[string]*vschemapb.Table{
			"accepted_tos": {
				ColumnVindexes: []*vschemapb.ColumnVindex{
					{
						Name:    "g_advertiser_id",
						Columns: []string{"g_advertiser_id"},
					},
					{
						Name:    "advertiser_id",
						Columns: []string{"advertiser_id"},
					},
				},
			},
			"advertisers": {
				ColumnVindexes: []*vschemapb.ColumnVindex{
					{
						Name:    "advertiser_id",
						Columns: []string{"id"},
					},
				},
			},
			"campaigns": {
				ColumnVindexes: []*vschemapb.ColumnVindex{
					{
						Name:    "g_advertiser_id",
						Columns: []string{"g_advertiser_id"},
					},
					{
						Name:    "advertiser_id",
						Columns: []string{"advertiser_id"},
					},
				},
			},
			"ad_groups": {
				ColumnVindexes: []*vschemapb.ColumnVindex{
					{
						Name:    "g_advertiser_id",
						Columns: []string{"g_advertiser_id"},
					},
					{
						Name:    "advertiser_id",
						Columns: []string{"advertiser_id"},
					},
				},
			},
			"targeting_attribute_counts_by_advertiser": {
				ColumnVindexes: []*vschemapb.ColumnVindex{
					{
						Name:    "g_advertiser_id",
						Columns: []string{"advertiser_gid"},
					},
					{
						Name:    "advertiser_id",
						Columns: []string{"advertiser_id"},
					},
				},
			},
		},
	}
	if !proto.Equal(got, want) {
		t.Errorf("GetVSchema: %s, want %s", proto.MarshalTextString(got), proto.MarshalTextString(want))
	}
}

func TestVSchemaSecondaryVindex(t *testing.T) {
	ddls, err := parseSchema(ddls)
	if err != nil {
		t.Error(err)
	}

	config := pinschemaConfig{
		createPrimary:               true,
		createSecondary:             true,
		defaultScatterCacheCapacity: 10000,
		tableScatterCacheCapacity: map[string]uint64{
			"campaigns": 20000,
		},
	}
	got, err := newVschemaBuilder(ddls, config).ddlsToVSchema()
	if err != nil {
		t.Error(err)
	}
	want := &vschemapb.Keyspace{
		Sharded: true,
		Vindexes: map[string]*vschemapb.Vindex{
			"g_advertiser_id": {
				Type: "hash",
			},
			"dark_write_advertiser_id": {
				Type: "hash_offset",
				Params: map[string]string{
					"offset": "549755813888",
				},
			},
			"advertiser_id": {
				Type: "hash_offset",
				Params: map[string]string{
					"offset": "549755813888",
				},
			},
			"accepted_tos_id": {
				Type: "scatter_cache",
				Params: map[string]string{
					"capacity": "10000",
					"table":    "accepted_tos",
					"from":     "id",
					"to":       "g_advertiser_id",
				},
			},
			"campaign_id": {
				Type: "scatter_cache",
				Params: map[string]string{
					"capacity": "20000",
					"table":    "campaigns",
					"from":     "id",
					"to":       "g_advertiser_id",
				},
			},
			"ad_group_id": {
				Type: "scatter_cache",
				Params: map[string]string{
					"capacity": "10000",
					"table":    "ad_groups",
					"from":     "id",
					"to":       "g_advertiser_id",
				},
			},
		},
		Tables: map[string]*vschemapb.Table{
			"accepted_tos": {
				ColumnVindexes: []*vschemapb.ColumnVindex{
					{
						Name:    "g_advertiser_id",
						Columns: []string{"g_advertiser_id"},
					},
					{
						Name:    "accepted_tos_id",
						Columns: []string{"id"},
					},
					{
						Name:    "advertiser_id",
						Columns: []string{"advertiser_id"},
					},
				},
			},
			"advertisers": {
				ColumnVindexes: []*vschemapb.ColumnVindex{
					{
						Name:    "advertiser_id",
						Columns: []string{"id"},
					},
				},
			},
			"campaigns": {
				ColumnVindexes: []*vschemapb.ColumnVindex{
					{
						Name:    "g_advertiser_id",
						Columns: []string{"g_advertiser_id"},
					},
					{
						Name:    "advertiser_id",
						Columns: []string{"advertiser_id"},
					},
					{
						Name:    "campaign_id",
						Columns: []string{"id"},
					},
				},
			},
			"ad_groups": {
				ColumnVindexes: []*vschemapb.ColumnVindex{
					{
						Name:    "g_advertiser_id",
						Columns: []string{"g_advertiser_id"},
					},
					{
						Name:    "ad_group_id",
						Columns: []string{"id"},
					},
					{
						Name:    "advertiser_id",
						Columns: []string{"advertiser_id"},
					},
					{
						Name:    "campaign_id",
						Columns: []string{"campaign_id"},
					},
				},
			},
			"targeting_attribute_counts_by_advertiser": {
				ColumnVindexes: []*vschemapb.ColumnVindex{
					{
						Name:    "g_advertiser_id",
						Columns: []string{"advertiser_gid"},
					},
					{
						Name:    "advertiser_id",
						Columns: []string{"advertiser_id"},
					},
				},
			},
		},
	}
	if !proto.Equal(got, want) {
		t.Errorf("GetVSchema: %s, want %s", proto.MarshalTextString(got), proto.MarshalTextString(want))
	}
}
