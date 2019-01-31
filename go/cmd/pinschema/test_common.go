package main

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

--
-- Table structure for table targeting_attribute_counts_by_advertiser
--

DROP TABLE IF EXISTS targeting_attribute_counts_by_advertiser;
/*!40101 SET @saved_cs_client     = @@character_set_client */;
/*!40101 SET character_set_client = utf8 */;
CREATE TABLE targeting_attribute_counts_by_advertiser (
  advertiser_gid bigint(20) NOT NULL,
  active_keywords_count bigint(20) NOT NULL DEFAULT '0',
  advertiser_id bigint(20) NOT NULL,
  PRIMARY KEY (advertiser_gid)
) ENGINE=InnoDB DEFAULT CHARSET=latin1;
/*!40101 SET character_set_client = @saved_cs_client */;
`
