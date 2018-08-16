package main

import (
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
`

func TestPinschemaOriginal(t *testing.T) {
	ddls, err := parseSchema(ddls)
	if err != nil {
		t.Error(err)
	}

	config := pinschemaConfig{}
	got, err := ddlsToVSchema(ddls, config)
	if err != nil {
		t.Error(err)
	}
	want := &vschemapb.Keyspace{
		Tables: map[string]*vschemapb.Table{
			"accepted_tos": {},
			"advertisers":  {},
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
	got, err := ddlsToVSchema(ddls, config)
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
		},
	}
	if !proto.Equal(got, want) {
		t.Errorf("GetVSchema: %s, want %s", got, want)
	}
}

func TestPinschemaPrimaryVindex(t *testing.T) {
	ddls, err := parseSchema(ddls)
	if err != nil {
		t.Error(err)
	}

	config := pinschemaConfig{createPrimary: true}
	got, err := ddlsToVSchema(ddls, config)
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
			"advertiser_gid": {
				Type: "hash",
			},
		},
		Tables: map[string]*vschemapb.Table{
			"accepted_tos": {
				ColumnVindexes: []*vschemapb.ColumnVindex{{
					Name:    "advertiser_gid",
					Columns: []string{"g_advertiser_id"},
				}},
			},
			"advertisers": {
				ColumnVindexes: []*vschemapb.ColumnVindex{{
					Name:    "advertiser_id",
					Columns: []string{"id"},
				}},
			},
		},
	}
	if !proto.Equal(got, want) {
		t.Errorf("GetVSchema: %s, want %s", proto.MarshalTextString(got), proto.MarshalTextString(want))
	}
}
