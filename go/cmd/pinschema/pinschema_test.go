package main

import (
	"strings"
	"testing"

	"github.com/golang/protobuf/proto"
	vschemapb "vitess.io/vitess/go/vt/proto/vschema"
)

const ddls = `
--
-- Table structure for table advertisers
--

DROP TABLE IF EXISTS advertisers;
/*!40101 SET @saved_cs_client     = @@character_set_client */;
/*!40101 SET character_set_client = utf8 */;
CREATE TABLE advertisers (
  id bigint(20) NOT NULL AUTO_INCREMENT,
  active tinyint(1) NOT NULL DEFAULT '1',
  creation_date timestamp NOT NULL DEFAULT CURRENT_TIMESTAMP,
  name varchar(64) NOT NULL,
  owner_user_id bigint(20) NOT NULL,
  billing_type varchar(64) DEFAULT NULL,
  billing_token varchar(64) DEFAULT NULL,
  billing_profile_id bigint(20) DEFAULT NULL,
  billing_threshold int(11) DEFAULT NULL,
  test_account tinyint(1) NOT NULL DEFAULT '0',
  gid bigint(20) DEFAULT NULL,
  is_spam tinyint(1) DEFAULT '0',
  properties text,
  deleted tinyint(1) NOT NULL DEFAULT '0',
  country smallint(6) NOT NULL DEFAULT '1',
  currency smallint(6) NOT NULL DEFAULT '1',
  business_profile_id bigint(20) DEFAULT NULL,
  updated_time timestamp NOT NULL DEFAULT CURRENT_TIMESTAMP,
  g_billing_profile_id bigint(20) DEFAULT NULL,
  g_business_profile_id bigint(20) DEFAULT NULL,
  daily_spend_cap int(11) DEFAULT NULL,
  PRIMARY KEY (id),
  KEY billing_profile_id (billing_profile_id),
  KEY owner_user_id (owner_user_id),
  CONSTRAINT advertisers_ibfk_1 FOREIGN KEY (billing_profile_id) REFERENCES billing_profiles (id)
) ENGINE=InnoDB AUTO_INCREMENT=1024152 DEFAULT CHARSET=utf8;
/*!40101 SET character_set_client = @saved_cs_client */;

--
-- Table structure for table accepted_tos
--

DROP TABLE IF EXISTS accepted_tos;
/*!40101 SET @saved_cs_client     = @@character_set_client */;
/*!40101 SET character_set_client = utf8 */;
CREATE TABLE accepted_tos (
  advertiser_id bigint(20) NOT NULL,
  g_advertiser_id bigint(20) NOT NULL,
  tos_id smallint(6) NOT NULL,
  accept_date timestamp NOT NULL DEFAULT CURRENT_TIMESTAMP,
  properties text,
  id bigint(20) NOT NULL AUTO_INCREMENT,
  PRIMARY KEY (id),
  KEY advertiser_id (advertiser_id),
  KEY g_advertiser_id (g_advertiser_id),
  KEY tos_id (tos_id),
  CONSTRAINT accepted_tos_ibfk_1 FOREIGN KEY (advertiser_id) REFERENCES advertisers (id)
) ENGINE=InnoDB AUTO_INCREMENT=526723 DEFAULT CHARSET=latin1;
/*!40101 SET character_set_client = @saved_cs_client */;

--
-- Table structure for table campaigns
--

DROP TABLE IF EXISTS campaigns;
/*!40101 SET @saved_cs_client     = @@character_set_client */;
/*!40101 SET character_set_client = utf8 */;
CREATE TABLE campaigns (
  id bigint(20) NOT NULL AUTO_INCREMENT,
  active tinyint(1) NOT NULL DEFAULT '1',
  creation_date timestamp NOT NULL DEFAULT CURRENT_TIMESTAMP,
  campaign_spec_id bigint(20) DEFAULT NULL,
  advertiser_id bigint(20) NOT NULL,
  name varchar(128) DEFAULT NULL,
  unique_line_count_id bigint(20) DEFAULT NULL,
  action_type int(11) DEFAULT NULL,
  gid bigint(20) DEFAULT NULL,
  g_campaign_spec_id bigint(20) DEFAULT NULL,
  g_advertiser_id bigint(20) DEFAULT NULL,
  properties text,
  creative_type int(11) NOT NULL DEFAULT '0',
  objective_type int(11) DEFAULT NULL,
  updated_time timestamp NOT NULL DEFAULT CURRENT_TIMESTAMP,
  PRIMARY KEY (id),
  KEY ix_campaigns_advertiser_id (advertiser_id),
  KEY campaign_spec_id (campaign_spec_id),
  KEY active (active),
  KEY updated_time (updated_time),
  CONSTRAINT campaigns_ibfk_1 FOREIGN KEY (advertiser_id) REFERENCES advertisers (id),
  CONSTRAINT campaigns_ibfk_2 FOREIGN KEY (campaign_spec_id) REFERENCES campaign_specs (id)
) ENGINE=InnoDB AUTO_INCREMENT=8264670178 DEFAULT CHARSET=utf8;
/*!40101 SET character_set_client = @saved_cs_client */;

--
-- Table structure for table ad_groups
--

DROP TABLE IF EXISTS ad_groups;
/*!40101 SET @saved_cs_client     = @@character_set_client */;
/*!40101 SET character_set_client = utf8 */;
CREATE TABLE ad_groups (
  id bigint(20) NOT NULL AUTO_INCREMENT,
  gid bigint(20) DEFAULT NULL,
  active tinyint(1) NOT NULL DEFAULT '1',
  creation_date timestamp NOT NULL DEFAULT CURRENT_TIMESTAMP,
  spec_id bigint(20) DEFAULT NULL,
  campaign_id bigint(20) NOT NULL,
  properties mediumtext,
  updated_time timestamp NOT NULL DEFAULT CURRENT_TIMESTAMP,
  g_campaign_id bigint(20) DEFAULT NULL,
  g_spec_id bigint(20) DEFAULT NULL,
  advertiser_id bigint(20) DEFAULT NULL,
  g_advertiser_id bigint(20) DEFAULT NULL,
  PRIMARY KEY (id),
  KEY spec_id (spec_id),
  KEY campaign_id (campaign_id),
  KEY updated_time (updated_time),
  CONSTRAINT ad_groups_ibfk_1 FOREIGN KEY (spec_id) REFERENCES ad_group_specs (id),
  CONSTRAINT ad_groups_ibfk_2 FOREIGN KEY (campaign_id) REFERENCES campaigns (id)
) ENGINE=InnoDB AUTO_INCREMENT=3228108 DEFAULT CHARSET=utf8;
/*!40101 SET character_set_client = @saved_cs_client */;
`

func TestPinschemaOriginal(t *testing.T) {
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
		},
	}
	if !proto.Equal(got, want) {
		t.Errorf("GetVSchema: %s, want %s", got, want)
	}
}

func TestPinschemaSequences(t *testing.T) {
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
		},
	}
	if !proto.Equal(got, want) {
		t.Errorf("GetVSchema: %s, want %s", got, want)
	}
}

func TestPinschemaSequenceDDLs(t *testing.T) {
	ddls, err := parseSchema(ddls)
	if err != nil {
		t.Error(err)
	}

	got := buildSequenceDDLs(ddls)
	want := strings.Join(
		[]string{
			"create table `advertisers_seq`(id int, next_id bigint, cache bigint, primary key(id)) comment 'vitess_sequence';",
			"insert into `advertisers_seq`(id, next_id, cache) values(0, 1, 1);",
			"",
			"create table `accepted_tos_seq`(id int, next_id bigint, cache bigint, primary key(id)) comment 'vitess_sequence';",
			"insert into `accepted_tos_seq`(id, next_id, cache) values(0, 1, 1);",
			"",
			"create table `campaigns_seq`(id int, next_id bigint, cache bigint, primary key(id)) comment 'vitess_sequence';",
			"insert into `campaigns_seq`(id, next_id, cache) values(0, 1, 1);",
			"",
			"create table `ad_groups_seq`(id int, next_id bigint, cache bigint, primary key(id)) comment 'vitess_sequence';",
			"insert into `ad_groups_seq`(id, next_id, cache) values(0, 1, 1);",
		}, "\n") + "\n\n"
	if got != want {
		t.Errorf("buildSequenceDDLs: \"%s\", want \"%s\"", got, want)
	}
}

func TestPinschemaPrimaryVindex(t *testing.T) {
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
		},
	}
	if !proto.Equal(got, want) {
		t.Errorf("GetVSchema: %s, want %s", proto.MarshalTextString(got), proto.MarshalTextString(want))
	}
}

func TestPinschemaSecondaryVindex(t *testing.T) {
	ddls, err := parseSchema(ddls)
	if err != nil {
		t.Error(err)
	}

	config := pinschemaConfig{createPrimary: true, createSecondary: true}
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
					"capacity": "10000",
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
		},
	}
	if !proto.Equal(got, want) {
		t.Errorf("GetVSchema: %s, want %s", proto.MarshalTextString(got), proto.MarshalTextString(want))
	}
}
