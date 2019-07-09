-- MySQL dump 10.13  Distrib 5.6.41-84.1, for debian-linux-gnu (x86_64)
--
-- Host: localhost    Database: patio
-- ------------------------------------------------------
-- Server version	5.6.41-84.1-log

/*!40101 SET @OLD_CHARACTER_SET_CLIENT=@@CHARACTER_SET_CLIENT */;
/*!40101 SET @OLD_CHARACTER_SET_RESULTS=@@CHARACTER_SET_RESULTS */;
/*!40101 SET @OLD_COLLATION_CONNECTION=@@COLLATION_CONNECTION */;
/*!40101 SET NAMES utf8 */;
/*!40103 SET @OLD_TIME_ZONE=@@TIME_ZONE */;
/*!40103 SET TIME_ZONE='+00:00' */;
/*!40014 SET @OLD_UNIQUE_CHECKS=@@UNIQUE_CHECKS, UNIQUE_CHECKS=0 */;
/*!40014 SET @OLD_FOREIGN_KEY_CHECKS=@@FOREIGN_KEY_CHECKS, FOREIGN_KEY_CHECKS=0 */;
/*!40101 SET @OLD_SQL_MODE=@@SQL_MODE, SQL_MODE='NO_AUTO_VALUE_ON_ZERO' */;
/*!40111 SET @OLD_SQL_NOTES=@@SQL_NOTES, SQL_NOTES=0 */;

--
-- Current Database: `patio`
--

CREATE DATABASE /*!32312 IF NOT EXISTS*/ `patio` /*!40100 DEFAULT CHARACTER SET utf8mb4 COLLATE utf8mb4_unicode_ci */;

USE `patio`;

--
-- Table structure for table `accepted_tos`
--

DROP TABLE IF EXISTS `accepted_tos`;
/*!40101 SET @saved_cs_client     = @@character_set_client */;
/*!40101 SET character_set_client = utf8 */;
CREATE TABLE `accepted_tos` (
  `advertiser_id` bigint(20) NOT NULL,
  `g_advertiser_id` bigint(20) NOT NULL,
  `tos_id` smallint(6) NOT NULL,
  `accept_date` timestamp NOT NULL DEFAULT CURRENT_TIMESTAMP,
  `properties` mediumtext COLLATE utf8mb4_unicode_ci,
  `id` bigint(20) NOT NULL AUTO_INCREMENT,
  PRIMARY KEY (`id`),
  KEY `advertiser_id` (`advertiser_id`),
  KEY `g_advertiser_id` (`g_advertiser_id`),
  KEY `tos_id` (`tos_id`),
  CONSTRAINT `accepted_tos_ibfk_1` FOREIGN KEY (`advertiser_id`) REFERENCES `advertisers` (`id`)
) ENGINE=InnoDB AUTO_INCREMENT=726724 DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci;
/*!40101 SET character_set_client = @saved_cs_client */;

--
-- Table structure for table `ad_group_specs`
--

DROP TABLE IF EXISTS `ad_group_specs`;
/*!40101 SET @saved_cs_client     = @@character_set_client */;
/*!40101 SET character_set_client = utf8 */;
CREATE TABLE `ad_group_specs` (
  `id` bigint(20) NOT NULL AUTO_INCREMENT,
  `gid` bigint(20) DEFAULT NULL,
  `ad_group_id` bigint(20) NOT NULL,
  `active` tinyint(1) NOT NULL DEFAULT '1',
  `creation_date` timestamp NOT NULL DEFAULT CURRENT_TIMESTAMP,
  `targeting_spec_id` bigint(20) NOT NULL,
  `properties` longtext COLLATE utf8mb4_unicode_ci,
  `start_time` datetime DEFAULT NULL,
  `end_time` datetime DEFAULT NULL,
  `g_targeting_spec_id` bigint(20) DEFAULT NULL,
  `g_ad_group_id` bigint(20) DEFAULT NULL,
  `status` smallint(6) DEFAULT NULL,
  `billable_metric` int(11) DEFAULT '-1',
  `optimization_goal` int(11) DEFAULT '-1',
  `optimization_goal_metadata` mediumtext COLLATE utf8mb4_unicode_ci,
  `advertiser_id` bigint(20) NOT NULL,
  `g_advertiser_id` bigint(20) NOT NULL,
  `placement_group` int(11) DEFAULT NULL,
  `action_type` int(11) DEFAULT NULL,
  PRIMARY KEY (`id`),
  KEY `ad_group_id` (`ad_group_id`),
  KEY `targeting_spec_id` (`targeting_spec_id`),
  CONSTRAINT `ad_group_specs_ibfk_1` FOREIGN KEY (`ad_group_id`) REFERENCES `ad_groups` (`id`),
  CONSTRAINT `ad_group_specs_ibfk_2` FOREIGN KEY (`targeting_spec_id`) REFERENCES `targeting_specs` (`id`)
) ENGINE=InnoDB AUTO_INCREMENT=18070879 DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci;
/*!40101 SET character_set_client = @saved_cs_client */;

--
-- Table structure for table `ad_groups`
--

DROP TABLE IF EXISTS `ad_groups`;
/*!40101 SET @saved_cs_client     = @@character_set_client */;
/*!40101 SET character_set_client = utf8 */;
CREATE TABLE `ad_groups` (
  `id` bigint(20) NOT NULL AUTO_INCREMENT,
  `gid` bigint(20) DEFAULT NULL,
  `active` tinyint(1) NOT NULL DEFAULT '1',
  `creation_date` timestamp NOT NULL DEFAULT CURRENT_TIMESTAMP,
  `spec_id` bigint(20) DEFAULT NULL,
  `campaign_id` bigint(20) NOT NULL,
  `properties` longtext COLLATE utf8mb4_unicode_ci,
  `updated_time` timestamp NOT NULL DEFAULT CURRENT_TIMESTAMP,
  `g_campaign_id` bigint(20) DEFAULT NULL,
  `g_spec_id` bigint(20) DEFAULT NULL,
  `advertiser_id` bigint(20) NOT NULL,
  `g_advertiser_id` bigint(20) NOT NULL,
  PRIMARY KEY (`id`),
  KEY `spec_id` (`spec_id`),
  KEY `campaign_id` (`campaign_id`),
  KEY `updated_time` (`updated_time`),
  KEY `idx_advertiser_id` (`advertiser_id`),
  CONSTRAINT `ad_groups_ibfk_1` FOREIGN KEY (`spec_id`) REFERENCES `ad_group_specs` (`id`),
  CONSTRAINT `ad_groups_ibfk_2` FOREIGN KEY (`campaign_id`) REFERENCES `campaigns` (`id`)
) ENGINE=InnoDB AUTO_INCREMENT=3840693 DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci;
/*!40101 SET character_set_client = @saved_cs_client */;

--
-- Table structure for table `advertisers`
--

DROP TABLE IF EXISTS `advertisers`;
/*!40101 SET @saved_cs_client     = @@character_set_client */;
/*!40101 SET character_set_client = utf8 */;
CREATE TABLE `advertisers` (
  `id` bigint(20) NOT NULL AUTO_INCREMENT,
  `active` tinyint(1) NOT NULL DEFAULT '1',
  `creation_date` timestamp NOT NULL DEFAULT CURRENT_TIMESTAMP,
  `name` varchar(64) COLLATE utf8mb4_unicode_ci NOT NULL,
  `owner_user_id` bigint(20) NOT NULL,
  `billing_type` varchar(64) COLLATE utf8mb4_unicode_ci DEFAULT NULL,
  `billing_token` varchar(64) COLLATE utf8mb4_unicode_ci DEFAULT NULL,
  `billing_profile_id` bigint(20) DEFAULT NULL,
  `billing_threshold` int(11) DEFAULT NULL,
  `test_account` tinyint(1) NOT NULL DEFAULT '0',
  `gid` bigint(20) DEFAULT NULL,
  `is_spam` tinyint(1) DEFAULT '0',
  `properties` mediumtext COLLATE utf8mb4_unicode_ci,
  `deleted` tinyint(1) NOT NULL DEFAULT '0',
  `country` smallint(6) NOT NULL DEFAULT '1',
  `currency` smallint(6) NOT NULL DEFAULT '1',
  `business_profile_id` bigint(20) DEFAULT NULL,
  `updated_time` timestamp NOT NULL DEFAULT CURRENT_TIMESTAMP,
  `g_billing_profile_id` bigint(20) DEFAULT NULL,
  `g_business_profile_id` bigint(20) DEFAULT NULL,
  `daily_spend_cap` int(11) DEFAULT NULL,
  PRIMARY KEY (`id`),
  KEY `billing_profile_id` (`billing_profile_id`),
  KEY `owner_user_id` (`owner_user_id`),
  KEY `updated_time` (`updated_time`),
  KEY `daily_spend_cap` (`daily_spend_cap`),
  CONSTRAINT `advertisers_ibfk_1` FOREIGN KEY (`billing_profile_id`) REFERENCES `billing_profiles` (`id`)
) ENGINE=InnoDB AUTO_INCREMENT=1271103 DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci;
/*!40101 SET character_set_client = @saved_cs_client */;

--
-- Table structure for table `app_event_tracking_configs`
--

DROP TABLE IF EXISTS `app_event_tracking_configs`;
/*!40101 SET @saved_cs_client     = @@character_set_client */;
/*!40101 SET character_set_client = utf8 */;
CREATE TABLE `app_event_tracking_configs` (
  `id` bigint(20) NOT NULL AUTO_INCREMENT,
  `gid` bigint(20) DEFAULT NULL,
  `advertiser_id` bigint(20) NOT NULL,
  `g_advertiser_id` bigint(20) NOT NULL,
  `view_window_days` smallint(6) NOT NULL,
  `click_window_days` smallint(6) NOT NULL,
  `closeuprepin_window_days` smallint(6) NOT NULL,
  `properties` mediumtext COLLATE utf8mb4_unicode_ci,
  `creation_date` timestamp NOT NULL DEFAULT CURRENT_TIMESTAMP,
  PRIMARY KEY (`id`),
  KEY `advertiser_id` (`advertiser_id`),
  CONSTRAINT `app_event_tracking_configs_ibfk_1` FOREIGN KEY (`advertiser_id`) REFERENCES `advertisers` (`id`)
) ENGINE=InnoDB AUTO_INCREMENT=83 DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci;
/*!40101 SET character_set_client = @saved_cs_client */;

--
-- Table structure for table `bill_details`
--

DROP TABLE IF EXISTS `bill_details`;
/*!40101 SET @saved_cs_client     = @@character_set_client */;
/*!40101 SET character_set_client = utf8 */;
CREATE TABLE `bill_details` (
  `id` bigint(20) NOT NULL AUTO_INCREMENT,
  `bill_id` bigint(20) NOT NULL,
  `campaign_id` bigint(20) DEFAULT NULL,
  `action_type` varchar(64) COLLATE utf8mb4_unicode_ci DEFAULT NULL,
  `quantity` bigint(20) DEFAULT NULL,
  `description` varchar(128) COLLATE utf8mb4_unicode_ci DEFAULT NULL,
  `currency` char(3) COLLATE utf8mb4_unicode_ci DEFAULT 'USD',
  `spend_in_micro_currency` bigint(20) DEFAULT NULL,
  `spend_in_micro_dollar` bigint(20) DEFAULT NULL,
  `unique_line_count_id` bigint(20) DEFAULT NULL,
  `objective_type` int(11) DEFAULT NULL,
  `g_campaign_id` bigint(20) DEFAULT NULL,
  `tax_in_micro_currency` bigint(20) DEFAULT NULL,
  `gid` bigint(20) DEFAULT NULL,
  `g_bill_id` bigint(20) DEFAULT NULL,
  `advertiser_id` bigint(20) NOT NULL,
  `g_advertiser_id` bigint(20) NOT NULL,
  PRIMARY KEY (`id`),
  KEY `bill_id` (`bill_id`),
  KEY `campaign_id` (`campaign_id`),
  CONSTRAINT `bill_details_ibfk_1` FOREIGN KEY (`bill_id`) REFERENCES `bills` (`id`),
  CONSTRAINT `bill_details_ibfk_2` FOREIGN KEY (`campaign_id`) REFERENCES `campaigns` (`id`)
) ENGINE=InnoDB AUTO_INCREMENT=5640177 DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci;
/*!40101 SET character_set_client = @saved_cs_client */;

--
-- Table structure for table `billing_actions`
--

DROP TABLE IF EXISTS `billing_actions`;
/*!40101 SET @saved_cs_client     = @@character_set_client */;
/*!40101 SET character_set_client = utf8 */;
CREATE TABLE `billing_actions` (
  `id` bigint(20) NOT NULL AUTO_INCREMENT,
  `creation_date` timestamp NOT NULL DEFAULT CURRENT_TIMESTAMP,
  `bill_id` bigint(20) NOT NULL,
  `action_type` varchar(64) COLLATE utf8mb4_unicode_ci NOT NULL,
  `action_time` datetime NOT NULL,
  `description` varchar(128) COLLATE utf8mb4_unicode_ci DEFAULT NULL,
  `cybersource_request_id` varchar(32) COLLATE utf8mb4_unicode_ci DEFAULT NULL,
  `processor_reconciliation_id` varchar(64) COLLATE utf8mb4_unicode_ci DEFAULT NULL,
  `encrypted_billing_token` varchar(256) COLLATE utf8mb4_unicode_ci DEFAULT NULL,
  `properties` mediumtext COLLATE utf8mb4_unicode_ci,
  `gid` bigint(20) DEFAULT NULL,
  `g_bill_id` bigint(20) DEFAULT NULL,
  `advertiser_id` bigint(20) NOT NULL,
  `g_advertiser_id` bigint(20) NOT NULL,
  `gateway` varchar(32) COLLATE utf8mb4_unicode_ci DEFAULT NULL,
  PRIMARY KEY (`id`),
  KEY `bill_id` (`bill_id`),
  CONSTRAINT `billing_actions_ibfk_1` FOREIGN KEY (`bill_id`) REFERENCES `bills` (`id`)
) ENGINE=InnoDB AUTO_INCREMENT=1179331 DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci;
/*!40101 SET character_set_client = @saved_cs_client */;

--
-- Table structure for table `billing_contacts`
--

DROP TABLE IF EXISTS `billing_contacts`;
/*!40101 SET @saved_cs_client     = @@character_set_client */;
/*!40101 SET character_set_client = utf8 */;
CREATE TABLE `billing_contacts` (
  `id` bigint(20) NOT NULL AUTO_INCREMENT,
  `gid` bigint(20) DEFAULT NULL,
  `advertiser_id` bigint(20) NOT NULL,
  `g_advertiser_id` bigint(20) NOT NULL,
  `email` varchar(255) COLLATE utf8mb4_unicode_ci NOT NULL,
  `properties` mediumtext COLLATE utf8mb4_unicode_ci,
  `is_loi_sapin` tinyint(1) NOT NULL DEFAULT '0',
  `active` tinyint(1) NOT NULL DEFAULT '0',
  PRIMARY KEY (`id`),
  KEY `advertiser_id` (`advertiser_id`),
  KEY `g_advertiser_id` (`g_advertiser_id`),
  CONSTRAINT `billing_contacts_ibfk_1` FOREIGN KEY (`advertiser_id`) REFERENCES `advertisers` (`id`)
) ENGINE=InnoDB AUTO_INCREMENT=108 DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci;
/*!40101 SET character_set_client = @saved_cs_client */;

--
-- Table structure for table `billing_profiles`
--

DROP TABLE IF EXISTS `billing_profiles`;
/*!40101 SET @saved_cs_client     = @@character_set_client */;
/*!40101 SET character_set_client = utf8 */;
CREATE TABLE `billing_profiles` (
  `id` bigint(20) NOT NULL AUTO_INCREMENT,
  `creation_date` timestamp NOT NULL DEFAULT CURRENT_TIMESTAMP,
  `advertiser_id` bigint(20) NOT NULL,
  `bill_to_email` varchar(128) COLLATE utf8mb4_unicode_ci DEFAULT NULL,
  `card_type` varchar(64) COLLATE utf8mb4_unicode_ci DEFAULT NULL,
  `card_number` varchar(32) COLLATE utf8mb4_unicode_ci DEFAULT NULL,
  `card_expiration` date DEFAULT NULL,
  `billing_type` varchar(64) COLLATE utf8mb4_unicode_ci DEFAULT NULL,
  `invoice_mapping` mediumtext COLLATE utf8mb4_unicode_ci,
  `properties` mediumtext COLLATE utf8mb4_unicode_ci,
  `encrypted_billing_token` varchar(256) COLLATE utf8mb4_unicode_ci DEFAULT NULL,
  `g_advertiser_id` bigint(20) NOT NULL,
  `gid` bigint(20) DEFAULT NULL,
  `gateway` varchar(32) COLLATE utf8mb4_unicode_ci DEFAULT NULL,
  PRIMARY KEY (`id`),
  KEY `advertiser_id` (`advertiser_id`),
  CONSTRAINT `billing_profiles_ibfk_1` FOREIGN KEY (`advertiser_id`) REFERENCES `advertisers` (`id`)
) ENGINE=InnoDB AUTO_INCREMENT=247154 DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci;
/*!40101 SET character_set_client = @saved_cs_client */;

--
-- Table structure for table `bills`
--

DROP TABLE IF EXISTS `bills`;
/*!40101 SET @saved_cs_client     = @@character_set_client */;
/*!40101 SET character_set_client = utf8 */;
CREATE TABLE `bills` (
  `id` bigint(20) NOT NULL AUTO_INCREMENT,
  `creation_date` timestamp NOT NULL DEFAULT CURRENT_TIMESTAMP,
  `advertiser_id` bigint(20) NOT NULL,
  `reconciliation_id` varchar(32) COLLATE utf8mb4_unicode_ci DEFAULT NULL,
  `start_time` datetime NOT NULL,
  `end_time` datetime DEFAULT NULL,
  `currency` char(3) COLLATE utf8mb4_unicode_ci NOT NULL DEFAULT 'USD',
  `spend_in_micro_currency` bigint(20) NOT NULL,
  `spend_in_micro_dollar` bigint(20) NOT NULL,
  `status` varchar(64) COLLATE utf8mb4_unicode_ci DEFAULT NULL,
  `last_billing_action_id` bigint(20) DEFAULT NULL,
  `billing_profile_id` bigint(20) DEFAULT NULL,
  `g_advertiser_id` bigint(20) NOT NULL,
  `tax_in_micro_currency` bigint(20) DEFAULT NULL,
  `properties` mediumtext COLLATE utf8mb4_unicode_ci,
  `original_bill_id` bigint(20) DEFAULT NULL,
  `type` smallint(6) DEFAULT NULL,
  `business_profile_id` bigint(20) DEFAULT NULL,
  `gid` bigint(20) DEFAULT NULL,
  `g_last_billing_action_id` bigint(20) DEFAULT NULL,
  `g_billing_profile_id` bigint(20) DEFAULT NULL,
  `g_original_bill_id` bigint(20) DEFAULT NULL,
  `g_business_profile_id` bigint(20) DEFAULT NULL,
  PRIMARY KEY (`id`),
  UNIQUE KEY `reconciliation_id` (`reconciliation_id`),
  KEY `advertiser_id` (`advertiser_id`),
  KEY `business_profile_id` (`business_profile_id`),
  KEY `g_advertiser_id` (`g_advertiser_id`),
  CONSTRAINT `bills_ibfk_1` FOREIGN KEY (`advertiser_id`) REFERENCES `advertisers` (`id`),
  CONSTRAINT `bills_ibfk_2` FOREIGN KEY (`business_profile_id`) REFERENCES `business_profiles` (`id`)
) ENGINE=InnoDB AUTO_INCREMENT=800054 DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci;
/*!40101 SET character_set_client = @saved_cs_client */;

--
-- Table structure for table `bulk_v2_jobs`
--

DROP TABLE IF EXISTS `bulk_v2_jobs`;
/*!40101 SET @saved_cs_client     = @@character_set_client */;
/*!40101 SET character_set_client = utf8 */;
CREATE TABLE `bulk_v2_jobs` (
  `id` bigint(20) NOT NULL AUTO_INCREMENT,
  `gid` bigint(20) DEFAULT NULL,
  `advertiser_id` bigint(20) NOT NULL,
  `g_advertiser_id` bigint(20) NOT NULL,
  `request_id` varchar(256) COLLATE utf8mb4_unicode_ci DEFAULT NULL,
  `s3_path` varchar(1024) COLLATE utf8mb4_unicode_ci DEFAULT NULL,
  `job_type` smallint(6) NOT NULL,
  `status` smallint(6) NOT NULL,
  `properties` mediumtext COLLATE utf8mb4_unicode_ci,
  `creation_date` timestamp NOT NULL DEFAULT CURRENT_TIMESTAMP,
  `deleted` tinyint(1) NOT NULL DEFAULT '0',
  PRIMARY KEY (`id`),
  KEY `request_id` (`request_id`),
  KEY `job_type` (`job_type`,`status`,`creation_date`),
  KEY `advertiser_id` (`advertiser_id`,`job_type`)
) ENGINE=InnoDB AUTO_INCREMENT=15683827 DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci ROW_FORMAT=DYNAMIC;
/*!40101 SET character_set_client = @saved_cs_client */;

--
-- Table structure for table `business_profiles`
--

DROP TABLE IF EXISTS `business_profiles`;
/*!40101 SET @saved_cs_client     = @@character_set_client */;
/*!40101 SET character_set_client = utf8 */;
CREATE TABLE `business_profiles` (
  `id` bigint(20) NOT NULL AUTO_INCREMENT,
  `creation_date` timestamp NOT NULL DEFAULT CURRENT_TIMESTAMP,
  `advertiser_id` bigint(20) NOT NULL,
  `g_advertiser_id` bigint(20) NOT NULL,
  `properties` mediumtext COLLATE utf8mb4_unicode_ci,
  `gid` bigint(20) DEFAULT NULL,
  PRIMARY KEY (`id`)
) ENGINE=InnoDB AUTO_INCREMENT=256584 DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci;
/*!40101 SET character_set_client = @saved_cs_client */;

--
-- Table structure for table `campaign_specs`
--

DROP TABLE IF EXISTS `campaign_specs`;
/*!40101 SET @saved_cs_client     = @@character_set_client */;
/*!40101 SET character_set_client = utf8 */;
CREATE TABLE `campaign_specs` (
  `id` bigint(20) NOT NULL AUTO_INCREMENT,
  `active` tinyint(1) NOT NULL DEFAULT '1',
  `creation_date` timestamp NOT NULL DEFAULT CURRENT_TIMESTAMP,
  `campaign_id` bigint(20) NOT NULL,
  `start_time` datetime DEFAULT NULL,
  `end_time` datetime DEFAULT NULL,
  `budget` bigint(20) NOT NULL,
  `bonus_budget` bigint(20) DEFAULT NULL,
  `lifetime_paid_budget` bigint(20) NOT NULL DEFAULT '0',
  `lifetime_bonus_budget` bigint(20) NOT NULL DEFAULT '0',
  `gid` bigint(20) DEFAULT NULL,
  `g_campaign_id` bigint(20) DEFAULT NULL,
  `properties` mediumtext COLLATE utf8mb4_unicode_ci,
  `paused` tinyint(1) NOT NULL DEFAULT '0',
  `order_line_id` bigint(20) DEFAULT NULL,
  `status` smallint(6) DEFAULT NULL,
  `g_order_line_id` bigint(20) DEFAULT NULL,
  `is_managed` tinyint(1) DEFAULT NULL,
  `advertiser_id` bigint(20) NOT NULL,
  `g_advertiser_id` bigint(20) NOT NULL,
  `version` smallint(6) NOT NULL DEFAULT '0',
  `objective_type` int(11) DEFAULT NULL,
  PRIMARY KEY (`id`),
  KEY `order_line_id` (`order_line_id`),
  KEY `campaign_id` (`campaign_id`),
  CONSTRAINT `campaign_specs_ibfk_1` FOREIGN KEY (`order_line_id`) REFERENCES `order_lines` (`id`)
) ENGINE=InnoDB AUTO_INCREMENT=6223343 DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci;
/*!40101 SET character_set_client = @saved_cs_client */;

--
-- Table structure for table `campaigns`
--

DROP TABLE IF EXISTS `campaigns`;
/*!40101 SET @saved_cs_client     = @@character_set_client */;
/*!40101 SET character_set_client = utf8 */;
CREATE TABLE `campaigns` (
  `id` bigint(20) NOT NULL AUTO_INCREMENT,
  `active` tinyint(1) NOT NULL DEFAULT '1',
  `creation_date` timestamp NOT NULL DEFAULT CURRENT_TIMESTAMP,
  `campaign_spec_id` bigint(20) DEFAULT NULL,
  `advertiser_id` bigint(20) NOT NULL,
  `name` varchar(255) COLLATE utf8mb4_unicode_ci DEFAULT NULL,
  `unique_line_count_id` bigint(20) DEFAULT NULL,
  `action_type` int(11) DEFAULT NULL,
  `gid` bigint(20) DEFAULT NULL,
  `g_campaign_spec_id` bigint(20) DEFAULT NULL,
  `g_advertiser_id` bigint(20) NOT NULL,
  `properties` mediumtext COLLATE utf8mb4_unicode_ci,
  `creative_type` int(11) NOT NULL DEFAULT '0',
  `objective_type` int(11) DEFAULT NULL,
  `updated_time` timestamp NOT NULL DEFAULT CURRENT_TIMESTAMP,
  PRIMARY KEY (`id`),
  KEY `ix_campaigns_advertiser_id` (`advertiser_id`),
  KEY `campaign_spec_id` (`campaign_spec_id`),
  KEY `active` (`active`),
  KEY `updated_time` (`updated_time`),
  CONSTRAINT `_campaigns_ibfk_1` FOREIGN KEY (`advertiser_id`) REFERENCES `advertisers` (`id`),
  CONSTRAINT `_campaigns_ibfk_2` FOREIGN KEY (`campaign_spec_id`) REFERENCES `campaign_specs` (`id`)
) ENGINE=InnoDB AUTO_INCREMENT=8265068909 DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci;
/*!40101 SET character_set_client = @saved_cs_client */;

--
-- Table structure for table `carousel_slot_promotions`
--

DROP TABLE IF EXISTS `carousel_slot_promotions`;
/*!40101 SET @saved_cs_client     = @@character_set_client */;
/*!40101 SET character_set_client = utf8 */;
CREATE TABLE `carousel_slot_promotions` (
  `id` bigint(20) NOT NULL AUTO_INCREMENT,
  `status` smallint(6) NOT NULL,
  `gid` bigint(20) DEFAULT NULL,
  `carousel_data_id` bigint(20) NOT NULL,
  `carousel_slot_id` bigint(20) NOT NULL,
  `slot_index` smallint(6) NOT NULL,
  `pin_promotion_id` bigint(20) NOT NULL,
  `g_pin_promotion_id` bigint(20) NOT NULL,
  `advertiser_id` bigint(20) NOT NULL,
  `g_advertiser_id` bigint(20) NOT NULL,
  `properties` mediumtext COLLATE utf8mb4_unicode_ci,
  `creation_date` timestamp NOT NULL DEFAULT CURRENT_TIMESTAMP,
  `updated_time` timestamp NOT NULL DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP,
  PRIMARY KEY (`id`),
  KEY `gid` (`gid`),
  KEY `pin_promotion_id` (`pin_promotion_id`),
  CONSTRAINT `carousel_slot_promotions_ibfk_1` FOREIGN KEY (`pin_promotion_id`) REFERENCES `pin_promotions` (`id`)
) ENGINE=InnoDB AUTO_INCREMENT=35597 DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci;
/*!40101 SET character_set_client = @saved_cs_client */;

--
-- Table structure for table `conversion_tags`
--

DROP TABLE IF EXISTS `conversion_tags`;
/*!40101 SET @saved_cs_client     = @@character_set_client */;
/*!40101 SET character_set_client = utf8 */;
CREATE TABLE `conversion_tags` (
  `id` bigint(20) NOT NULL AUTO_INCREMENT,
  `gid` bigint(20) DEFAULT NULL,
  `advertiser_id` bigint(20) NOT NULL,
  `g_advertiser_id` bigint(20) NOT NULL,
  `active` tinyint(1) NOT NULL,
  `creation_date` timestamp NOT NULL DEFAULT CURRENT_TIMESTAMP,
  `name` varchar(64) COLLATE utf8mb4_unicode_ci NOT NULL,
  `id_str` varchar(64) COLLATE utf8mb4_unicode_ci NOT NULL,
  `type` smallint(6) NOT NULL,
  `view_window_days` smallint(6) NOT NULL,
  `click_window_days` smallint(6) NOT NULL,
  `engagement_window_days` smallint(6) NOT NULL,
  `verification_status` smallint(6) NOT NULL,
  `deleted` tinyint(1) NOT NULL,
  `properties` mediumtext COLLATE utf8mb4_unicode_ci,
  `remarketing_enabled` tinyint(1) NOT NULL DEFAULT '0',
  `updated_time` timestamp NOT NULL DEFAULT CURRENT_TIMESTAMP,
  PRIMARY KEY (`id`),
  UNIQUE KEY `id_str` (`id_str`),
  KEY `advertiser_id` (`advertiser_id`),
  CONSTRAINT `conversion_tags_ibfk_1` FOREIGN KEY (`advertiser_id`) REFERENCES `advertisers` (`id`)
) ENGINE=InnoDB AUTO_INCREMENT=18819 DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci;
/*!40101 SET character_set_client = @saved_cs_client */;

--
-- Table structure for table `goals`
--

DROP TABLE IF EXISTS `goals`;
/*!40101 SET @saved_cs_client     = @@character_set_client */;
/*!40101 SET character_set_client = utf8 */;
CREATE TABLE `goals` (
  `id` bigint(20) NOT NULL AUTO_INCREMENT,
  `gid` bigint(20) DEFAULT NULL,
  `advertiser_id` bigint(20) NOT NULL,
  `g_advertiser_id` bigint(20) NOT NULL,
  `entity_id` bigint(20) NOT NULL,
  `entity_type` smallint(6) NOT NULL,
  `goal_type` smallint(6) NOT NULL,
  `goal_value` decimal(15,3) NOT NULL,
  `click_attribution` smallint(6) DEFAULT NULL,
  `engagement_attribution` smallint(6) DEFAULT NULL,
  `view_attribution` smallint(6) DEFAULT NULL,
  `click_attribution_weight` decimal(6,5) NOT NULL DEFAULT '1.00000',
  `engagement_attribution_weight` decimal(6,5) NOT NULL DEFAULT '1.00000',
  `view_attribution_weight` decimal(6,5) NOT NULL DEFAULT '1.00000',
  PRIMARY KEY (`id`),
  KEY `entity` (`entity_id`,`entity_type`),
  KEY `goals_ibfk_1` (`advertiser_id`),
  CONSTRAINT `goals_ibfk_1` FOREIGN KEY (`advertiser_id`) REFERENCES `advertisers` (`id`)
) ENGINE=InnoDB AUTO_INCREMENT=3434 DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci;
/*!40101 SET character_set_client = @saved_cs_client */;

--
-- Table structure for table `notifications`
--

DROP TABLE IF EXISTS `notifications`;
/*!40101 SET @saved_cs_client     = @@character_set_client */;
/*!40101 SET character_set_client = utf8 */;
CREATE TABLE `notifications` (
  `id` bigint(20) NOT NULL AUTO_INCREMENT,
  `gid` bigint(20) DEFAULT NULL,
  `advertiser_id` bigint(20) NOT NULL,
  `g_advertiser_id` bigint(20) NOT NULL,
  `creation_date` timestamp NOT NULL DEFAULT CURRENT_TIMESTAMP,
  `entity_id` bigint(20) NOT NULL,
  `notification_type` tinyint(1) NOT NULL,
  `activation_time` datetime NOT NULL,
  `expiration_time` datetime DEFAULT NULL,
  `dismissal_time` datetime DEFAULT NULL,
  `deleted` tinyint(1) NOT NULL,
  `properties` mediumtext COLLATE utf8mb4_unicode_ci,
  PRIMARY KEY (`id`),
  KEY `advertiser_id` (`advertiser_id`),
  KEY `entity_id` (`entity_id`),
  CONSTRAINT `notifications_ibfk_1` FOREIGN KEY (`advertiser_id`) REFERENCES `advertisers` (`id`)
) ENGINE=InnoDB AUTO_INCREMENT=3815633 DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci;
/*!40101 SET character_set_client = @saved_cs_client */;

--
-- Table structure for table `order_line_specs`
--

DROP TABLE IF EXISTS `order_line_specs`;
/*!40101 SET @saved_cs_client     = @@character_set_client */;
/*!40101 SET character_set_client = utf8 */;
CREATE TABLE `order_line_specs` (
  `id` bigint(20) NOT NULL AUTO_INCREMENT,
  `creation_date` timestamp NOT NULL DEFAULT CURRENT_TIMESTAMP,
  `order_line_id` bigint(20) NOT NULL,
  `properties` mediumtext COLLATE utf8mb4_unicode_ci,
  `gid` bigint(20) DEFAULT NULL,
  `g_order_line_id` bigint(20) DEFAULT NULL,
  `advertiser_id` bigint(20) NOT NULL,
  `g_advertiser_id` bigint(20) NOT NULL,
  PRIMARY KEY (`id`),
  KEY `order_line_id` (`order_line_id`)
) ENGINE=InnoDB AUTO_INCREMENT=95950 DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci;
/*!40101 SET character_set_client = @saved_cs_client */;

--
-- Table structure for table `order_lines`
--

DROP TABLE IF EXISTS `order_lines`;
/*!40101 SET @saved_cs_client     = @@character_set_client */;
/*!40101 SET character_set_client = utf8 */;
CREATE TABLE `order_lines` (
  `id` bigint(20) NOT NULL AUTO_INCREMENT,
  `creation_date` timestamp NOT NULL DEFAULT CURRENT_TIMESTAMP,
  `spec_id` bigint(20) DEFAULT NULL,
  `billing_profile_id` bigint(20) DEFAULT NULL,
  `properties` mediumtext COLLATE utf8mb4_unicode_ci,
  `updated_time` timestamp NOT NULL DEFAULT CURRENT_TIMESTAMP,
  `gid` bigint(20) DEFAULT NULL,
  `g_spec_id` bigint(20) DEFAULT NULL,
  `g_billing_profile_id` bigint(20) DEFAULT NULL,
  `advertiser_id` bigint(20) NOT NULL,
  `g_advertiser_id` bigint(20) NOT NULL,
  PRIMARY KEY (`id`),
  KEY `billing_profile_id` (`billing_profile_id`),
  KEY `spec_id` (`spec_id`),
  KEY `advertiser_id` (`advertiser_id`),
  CONSTRAINT `order_lines_ibfk_1` FOREIGN KEY (`billing_profile_id`) REFERENCES `billing_profiles` (`id`),
  CONSTRAINT `order_lines_ibfk_2` FOREIGN KEY (`spec_id`) REFERENCES `order_line_specs` (`id`),
  CONSTRAINT `order_lines_ibfk_3` FOREIGN KEY (`spec_id`) REFERENCES `order_line_specs` (`id`),
  CONSTRAINT `order_lines_ibfk_4` FOREIGN KEY (`advertiser_id`) REFERENCES `advertisers` (`id`)
) ENGINE=InnoDB AUTO_INCREMENT=28882 DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci;
/*!40101 SET character_set_client = @saved_cs_client */;

--
-- Table structure for table `pin_promotion_labels`
--

DROP TABLE IF EXISTS `pin_promotion_labels`;
/*!40101 SET @saved_cs_client     = @@character_set_client */;
/*!40101 SET character_set_client = utf8 */;
CREATE TABLE `pin_promotion_labels` (
  `id` bigint(20) unsigned NOT NULL AUTO_INCREMENT,
  `g_advertiser_id` bigint(20) NOT NULL,
  `g_spec_id` bigint(20) DEFAULT NULL,
  `label_id` bigint(20) DEFAULT NULL,
  `advertiser_id` bigint(20) DEFAULT NULL,
  `spec_id` bigint(20) DEFAULT NULL,
  `gid` bigint(20) DEFAULT NULL,
  PRIMARY KEY (`id`),
  UNIQUE KEY `pin_promotion_spec_id` (`g_spec_id`,`label_id`),
  KEY `label_id` (`label_id`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci;
/*!40101 SET character_set_client = @saved_cs_client */;

--
-- Table structure for table `pin_promotion_specs`
--

DROP TABLE IF EXISTS `pin_promotion_specs`;
/*!40101 SET @saved_cs_client     = @@character_set_client */;
/*!40101 SET character_set_client = utf8 */;
CREATE TABLE `pin_promotion_specs` (
  `id` bigint(20) NOT NULL AUTO_INCREMENT,
  `active` tinyint(1) NOT NULL DEFAULT '1',
  `creation_date` timestamp NOT NULL DEFAULT CURRENT_TIMESTAMP,
  `pin_promotion_id` bigint(20) NOT NULL,
  `campaign_spec_id` bigint(20) DEFAULT NULL,
  `destination_url` varchar(2048) COLLATE utf8mb4_unicode_ci DEFAULT NULL,
  `targeting_deprecated` longtext COLLATE utf8mb4_unicode_ci,
  `annotations` longtext COLLATE utf8mb4_unicode_ci,
  `frequency_cap` int(11) DEFAULT NULL,
  `bid` bigint(20) DEFAULT NULL,
  `obsolete_auction_bid` bigint(20) DEFAULT '0',
  `silent_promotion` tinyint(1) NOT NULL DEFAULT '0',
  `approved` tinyint(1) NOT NULL DEFAULT '0',
  `reviewed` tinyint(1) NOT NULL DEFAULT '0',
  `exact_match` tinyint(1) NOT NULL DEFAULT '0',
  `user_targeting` tinyint(1) NOT NULL DEFAULT '0',
  `num_total_email_targeted` bigint(20) NOT NULL DEFAULT '0',
  `num_matched_user_emails` bigint(20) NOT NULL DEFAULT '0',
  `gid` bigint(20) DEFAULT NULL,
  `g_campaign_spec_id` bigint(20) DEFAULT NULL,
  `g_pin_promotion_id` bigint(20) DEFAULT NULL,
  `view_tags` mediumtext COLLATE utf8mb4_unicode_ci,
  `properties` mediumtext COLLATE utf8mb4_unicode_ci,
  `targeting_spec_id` bigint(20) DEFAULT NULL,
  `g_targeting_spec_id` bigint(20) DEFAULT NULL,
  `lifetime_frequency_cap` int(11) DEFAULT NULL,
  `user_list_id` bigint(20) DEFAULT NULL,
  `g_user_list_id` bigint(20) DEFAULT '0',
  `num_user_list_versions` bigint(20) DEFAULT '0',
  `has_valid_creative_type` tinyint(1) NOT NULL DEFAULT '1',
  `status` smallint(6) DEFAULT NULL,
  `is_ephemeral` tinyint(1) NOT NULL DEFAULT '0',
  `advertiser_id` bigint(20) NOT NULL,
  `g_advertiser_id` bigint(20) NOT NULL,
  `creative_type` int(11) DEFAULT NULL,
  PRIMARY KEY (`id`),
  KEY `ix_pin_promotion_specs_pin_promotion_id` (`pin_promotion_id`),
  KEY `campaign_spec_id` (`campaign_spec_id`),
  KEY `targeting_spec_id` (`targeting_spec_id`),
  KEY `user_list_id` (`user_list_id`),
  KEY `idx_unreviewed` (`active`,`reviewed`,`creation_date`),
  CONSTRAINT `pin_promotion_specs_ibfk_1` FOREIGN KEY (`pin_promotion_id`) REFERENCES `pin_promotions` (`id`),
  CONSTRAINT `pin_promotion_specs_ibfk_2` FOREIGN KEY (`campaign_spec_id`) REFERENCES `campaign_specs` (`id`),
  CONSTRAINT `pin_promotion_specs_ibfk_3` FOREIGN KEY (`targeting_spec_id`) REFERENCES `targeting_specs` (`id`)
) ENGINE=InnoDB AUTO_INCREMENT=30994933 DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci;
/*!40101 SET character_set_client = @saved_cs_client */;

--
-- Table structure for table `pin_promotions`
--

DROP TABLE IF EXISTS `pin_promotions`;
/*!40101 SET @saved_cs_client     = @@character_set_client */;
/*!40101 SET character_set_client = utf8 */;
CREATE TABLE `pin_promotions` (
  `id` bigint(20) NOT NULL AUTO_INCREMENT,
  `active` tinyint(1) NOT NULL DEFAULT '1',
  `creation_date` timestamp NOT NULL DEFAULT CURRENT_TIMESTAMP,
  `campaign_id` bigint(20) NOT NULL,
  `pin_id` bigint(20) NOT NULL,
  `spec_id` bigint(20) DEFAULT NULL,
  `pin_deleted` tinyint(1) NOT NULL DEFAULT '0',
  `marked_promoted` tinyint(1) NOT NULL DEFAULT '0',
  `gid` bigint(20) DEFAULT NULL,
  `g_campaign_id` bigint(20) DEFAULT NULL,
  `g_spec_id` bigint(20) DEFAULT NULL,
  `properties` mediumtext COLLATE utf8mb4_unicode_ci,
  `name` varchar(255) COLLATE utf8mb4_unicode_ci DEFAULT NULL,
  `updated_time` timestamp NOT NULL DEFAULT CURRENT_TIMESTAMP,
  `ad_group_id` bigint(20) DEFAULT NULL,
  `g_ad_group_id` bigint(20) DEFAULT NULL,
  `advertiser_id` bigint(20) NOT NULL,
  `g_advertiser_id` bigint(20) NOT NULL,
  PRIMARY KEY (`id`),
  KEY `ix_pin_promotions_campaign_id` (`campaign_id`),
  KEY `ix_pin_promotions_pin_id` (`pin_id`),
  KEY `active` (`active`,`pin_deleted`),
  KEY `ad_group_id` (`ad_group_id`),
  KEY `spec_id_2` (`spec_id`,`ad_group_id`,`gid`),
  KEY `updated_time` (`updated_time`),
  KEY `idx_advertiser_id` (`advertiser_id`),
  CONSTRAINT `_pin_promotions_ibfk_1` FOREIGN KEY (`campaign_id`) REFERENCES `campaigns` (`id`),
  CONSTRAINT `_pin_promotions_ibfk_2` FOREIGN KEY (`ad_group_id`) REFERENCES `ad_groups` (`id`)
) ENGINE=InnoDB AUTO_INCREMENT=5668544 DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci;
/*!40101 SET character_set_client = @saved_cs_client */;

--
-- Table structure for table `pin_promotions_history`
--

DROP TABLE IF EXISTS `pin_promotions_history`;
/*!40101 SET @saved_cs_client     = @@character_set_client */;
/*!40101 SET character_set_client = utf8 */;
CREATE TABLE `pin_promotions_history` (
  `id` bigint(20) NOT NULL AUTO_INCREMENT,
  `gid` bigint(20) DEFAULT NULL,
  `user_id` bigint(20) DEFAULT NULL,
  `ldap` varchar(64) COLLATE utf8mb4_unicode_ci DEFAULT NULL,
  `creation_date` timestamp NOT NULL DEFAULT CURRENT_TIMESTAMP,
  `action_time` timestamp NOT NULL DEFAULT CURRENT_TIMESTAMP,
  `crud_type` enum('CREATE','UPDATE','DELETE') COLLATE utf8mb4_unicode_ci NOT NULL,
  `gtid` varchar(64) COLLATE utf8mb4_unicode_ci DEFAULT NULL,
  `campaign_id` bigint(20) NOT NULL,
  `g_campaign_id` bigint(20) NOT NULL,
  `pin_promotion_id` bigint(20) NOT NULL,
  `g_pin_promotion_id` bigint(20) NOT NULL,
  `changes` text COLLATE utf8mb4_unicode_ci,
  `ad_group_id` bigint(20) NOT NULL,
  `g_ad_group_id` bigint(20) NOT NULL,
  `advertiser_id` bigint(20) NOT NULL,
  `g_advertiser_id` bigint(20) NOT NULL,
  PRIMARY KEY (`id`),
  KEY `campaign_id` (`campaign_id`),
  KEY `pin_promotion_id` (`pin_promotion_id`),
  KEY `ad_group_id` (`ad_group_id`),
  KEY `creation_date` (`creation_date`),
  KEY `action_time` (`action_time`),
  KEY `gtid` (`gtid`),
  KEY `advertiser_id_and_time` (`advertiser_id`,`action_time`)
) ENGINE=InnoDB AUTO_INCREMENT=11 DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci;
/*!40101 SET character_set_client = @saved_cs_client */;

--
-- Table structure for table `pinner_list_specs`
--

DROP TABLE IF EXISTS `pinner_list_specs`;
/*!40101 SET @saved_cs_client     = @@character_set_client */;
/*!40101 SET character_set_client = utf8 */;
CREATE TABLE `pinner_list_specs` (
  `id` bigint(20) NOT NULL AUTO_INCREMENT,
  `pinner_list_id` bigint(20) DEFAULT NULL,
  `creation_date` timestamp NOT NULL DEFAULT CURRENT_TIMESTAMP,
  `rule` mediumtext COLLATE utf8mb4_unicode_ci,
  `properties` mediumtext COLLATE utf8mb4_unicode_ci,
  `gid` bigint(20) DEFAULT NULL,
  `advertiser_id` bigint(20) NOT NULL,
  `g_advertiser_id` bigint(20) NOT NULL,
  `data_party_type` tinyint(1) DEFAULT NULL,
  `sharing_type` tinyint(1) DEFAULT NULL,
  `category` varchar(50) COLLATE utf8mb4_unicode_ci DEFAULT NULL,
  `brand_type` varchar(50) COLLATE utf8mb4_unicode_ci DEFAULT NULL,
  PRIMARY KEY (`id`),
  KEY `pinner_list_id` (`pinner_list_id`)
) ENGINE=InnoDB AUTO_INCREMENT=254785 DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci;
/*!40101 SET character_set_client = @saved_cs_client */;

--
-- Table structure for table `pinner_lists`
--

DROP TABLE IF EXISTS `pinner_lists`;
/*!40101 SET @saved_cs_client     = @@character_set_client */;
/*!40101 SET character_set_client = utf8 */;
CREATE TABLE `pinner_lists` (
  `id` bigint(20) NOT NULL AUTO_INCREMENT,
  `gid` bigint(20) DEFAULT NULL,
  `creation_date` timestamp NOT NULL DEFAULT CURRENT_TIMESTAMP,
  `type` tinyint(1) NOT NULL,
  `advertiser_id` bigint(20) NOT NULL,
  `status` tinyint(1) NOT NULL,
  `size` int(11) DEFAULT NULL,
  `spec_id` bigint(20) DEFAULT NULL,
  `properties` mediumtext COLLATE utf8mb4_unicode_ci,
  `deleted` tinyint(1) NOT NULL DEFAULT '0',
  `updated_time` timestamp NOT NULL DEFAULT CURRENT_TIMESTAMP,
  `g_advertiser_id` bigint(20) NOT NULL,
  PRIMARY KEY (`id`),
  KEY `advertiser_id` (`advertiser_id`)
) ENGINE=InnoDB AUTO_INCREMENT=208975 DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci;
/*!40101 SET character_set_client = @saved_cs_client */;

--
-- Table structure for table `product_group_specs`
--

DROP TABLE IF EXISTS `product_group_specs`;
/*!40101 SET @saved_cs_client     = @@character_set_client */;
/*!40101 SET character_set_client = utf8 */;
CREATE TABLE `product_group_specs` (
  `id` bigint(20) NOT NULL AUTO_INCREMENT,
  `gid` bigint(20) DEFAULT NULL,
  `active` tinyint(1) NOT NULL DEFAULT '0',
  `valid` tinyint(1) NOT NULL DEFAULT '0',
  `creation_date` timestamp NOT NULL DEFAULT CURRENT_TIMESTAMP,
  `product_group_id` bigint(20) DEFAULT NULL,
  `g_product_group_id` bigint(20) DEFAULT NULL,
  `bid` bigint(20) DEFAULT NULL,
  `reviewed` tinyint(1) NOT NULL DEFAULT '0',
  `approved` tinyint(1) NOT NULL DEFAULT '0',
  `properties` mediumtext COLLATE utf8mb4_unicode_ci,
  `destination_url` varchar(2048) COLLATE utf8mb4_unicode_ci DEFAULT NULL,
  `is_ephemeral` tinyint(1) NOT NULL DEFAULT '0',
  `advertiser_id` bigint(20) NOT NULL,
  `g_advertiser_id` bigint(20) NOT NULL,
  PRIMARY KEY (`id`),
  KEY `product_group_id` (`product_group_id`),
  CONSTRAINT `product_group_specs_ibfk_1` FOREIGN KEY (`product_group_id`) REFERENCES `product_groups` (`id`)
) ENGINE=InnoDB AUTO_INCREMENT=52524204 DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci;
/*!40101 SET character_set_client = @saved_cs_client */;

--
-- Table structure for table `product_groups`
--

DROP TABLE IF EXISTS `product_groups`;
/*!40101 SET @saved_cs_client     = @@character_set_client */;
/*!40101 SET character_set_client = utf8 */;
CREATE TABLE `product_groups` (
  `id` bigint(20) NOT NULL AUTO_INCREMENT,
  `gid` bigint(20) DEFAULT NULL,
  `active` tinyint(1) NOT NULL DEFAULT '0',
  `valid` tinyint(1) DEFAULT NULL,
  `inclusion` tinyint(1) DEFAULT '0',
  `parent_product_group_id` bigint(20) DEFAULT NULL,
  `g_parent_product_group_id` bigint(20) DEFAULT NULL,
  `ad_group_id` bigint(20) DEFAULT NULL,
  `g_ad_group_id` bigint(20) DEFAULT NULL,
  `spec_id` bigint(20) DEFAULT NULL,
  `g_spec_id` bigint(20) DEFAULT NULL,
  `product_group_type` smallint(6) NOT NULL,
  `product_group_definition` mediumtext COLLATE utf8mb4_unicode_ci,
  `creation_date` timestamp NOT NULL DEFAULT CURRENT_TIMESTAMP,
  `updated_time` timestamp NOT NULL DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP,
  `properties` mediumtext COLLATE utf8mb4_unicode_ci,
  `advertiser_id` bigint(20) NOT NULL,
  `g_advertiser_id` bigint(20) NOT NULL,
  PRIMARY KEY (`id`),
  KEY `ad_group_id` (`ad_group_id`),
  KEY `spec_id` (`spec_id`),
  KEY `ppgid` (`parent_product_group_id`,`valid`),
  KEY `updated_time` (`updated_time`),
  CONSTRAINT `product_groups_ibfk_1` FOREIGN KEY (`ad_group_id`) REFERENCES `ad_groups` (`id`),
  CONSTRAINT `product_groups_ibfk_2` FOREIGN KEY (`spec_id`) REFERENCES `product_group_specs` (`id`)
) ENGINE=InnoDB AUTO_INCREMENT=28121726 DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci;
/*!40101 SET character_set_client = @saved_cs_client */;

--
-- Table structure for table `promoted_catalog_product_groups`
--

DROP TABLE IF EXISTS `promoted_catalog_product_groups`;
/*!40101 SET @saved_cs_client     = @@character_set_client */;
/*!40101 SET character_set_client = utf8 */;
CREATE TABLE `promoted_catalog_product_groups` (
  `id` bigint(20) NOT NULL AUTO_INCREMENT,
  `gid` bigint(20) DEFAULT NULL,
  `status` smallint(6) DEFAULT NULL,
  `ad_group_id` bigint(20) DEFAULT NULL,
  `g_ad_group_id` bigint(20) DEFAULT NULL,
  `catalog_product_group_id` bigint(20) DEFAULT NULL,
  `creation_date` timestamp NOT NULL DEFAULT CURRENT_TIMESTAMP,
  `updated_time` timestamp NOT NULL DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP,
  `bid` bigint(20) DEFAULT NULL,
  `destination_url` varchar(2048) COLLATE utf8mb4_unicode_ci DEFAULT NULL,
  `reviewed` tinyint(1) DEFAULT NULL,
  `approved` tinyint(1) DEFAULT NULL,
  `properties` text COLLATE utf8mb4_unicode_ci,
  `campaign_id` bigint(20) NOT NULL,
  `g_campaign_id` bigint(20) NOT NULL,
  `advertiser_id` bigint(20) NOT NULL,
  `g_advertiser_id` bigint(20) NOT NULL,
  PRIMARY KEY (`id`),
  KEY `ad_group_id` (`ad_group_id`),
  KEY `id` (`gid`),
  KEY `catalog_product_group_id` (`catalog_product_group_id`)
) ENGINE=InnoDB AUTO_INCREMENT=41 DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci;
/*!40101 SET character_set_client = @saved_cs_client */;

--
-- Table structure for table `promoted_catalog_product_groups_history`
--

DROP TABLE IF EXISTS `promoted_catalog_product_groups_history`;
/*!40101 SET @saved_cs_client     = @@character_set_client */;
/*!40101 SET character_set_client = utf8 */;
CREATE TABLE `promoted_catalog_product_groups_history` (
  `id` bigint(20) NOT NULL AUTO_INCREMENT,
  `gid` bigint(20) NOT NULL,
  `promoted_catalog_product_group_id` bigint(20) DEFAULT NULL,
  `promoted_catalog_product_group_gid` bigint(20) DEFAULT NULL,
  `status` tinyint(1) DEFAULT NULL,
  `ad_group_id` bigint(20) DEFAULT NULL,
  `g_ad_group_id` bigint(20) DEFAULT NULL,
  `catalog_product_group_id` bigint(20) DEFAULT NULL,
  `creation_date` timestamp NULL DEFAULT NULL,
  `updated_time` timestamp NULL DEFAULT NULL,
  `bid` bigint(20) DEFAULT NULL,
  `destination_url` varchar(2048) COLLATE utf8mb4_unicode_ci DEFAULT NULL,
  `reviewed` tinyint(1) DEFAULT NULL,
  `approved` tinyint(1) DEFAULT NULL,
  `properties` text COLLATE utf8mb4_unicode_ci,
  `campaign_id` bigint(20) NOT NULL,
  `g_campaign_id` bigint(20) NOT NULL,
  `advertiser_id` bigint(20) NOT NULL,
  `g_advertiser_id` bigint(20) NOT NULL,
  `history_time` timestamp NOT NULL DEFAULT CURRENT_TIMESTAMP,
  PRIMARY KEY (`id`),
  KEY `gid` (`gid`),
  KEY `promoted_catalog_product_group_id` (`promoted_catalog_product_group_id`),
  KEY `promoted_catalog_product_group_gid` (`promoted_catalog_product_group_gid`),
  KEY `catalog_product_group_id` (`catalog_product_group_id`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci;
/*!40101 SET character_set_client = @saved_cs_client */;

--
-- Table structure for table `rule_subscriptions`
--

DROP TABLE IF EXISTS `rule_subscriptions`;
/*!40101 SET @saved_cs_client     = @@character_set_client */;
/*!40101 SET character_set_client = utf8 */;
CREATE TABLE `rule_subscriptions` (
  `id` bigint(20) NOT NULL AUTO_INCREMENT,
  `gid` bigint(20) DEFAULT NULL,
  `g_advertiser_id` bigint(20) NOT NULL,
  `g_entity_id` bigint(20) NOT NULL,
  `entity_type` smallint(6) NOT NULL,
  `rule_type` smallint(6) NOT NULL,
  `subscription_type` smallint(6) NOT NULL,
  `update_frequency_in_seconds` int(11) DEFAULT NULL,
  `properties` mediumtext COLLATE utf8mb4_unicode_ci NOT NULL,
  `active` tinyint(1) NOT NULL DEFAULT '1',
  `creation_date` timestamp NOT NULL DEFAULT CURRENT_TIMESTAMP,
  `advertiser_id` bigint(20) NOT NULL,
  PRIMARY KEY (`id`)
) ENGINE=InnoDB AUTO_INCREMENT=1780 DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci;
/*!40101 SET character_set_client = @saved_cs_client */;

--
-- Table structure for table `targeting_attribute_counts_by_advertiser`
--

DROP TABLE IF EXISTS `targeting_attribute_counts_by_advertiser`;
/*!40101 SET @saved_cs_client     = @@character_set_client */;
/*!40101 SET character_set_client = utf8 */;
CREATE TABLE `targeting_attribute_counts_by_advertiser` (
  `advertiser_gid` bigint(20) NOT NULL,
  `active_keywords_count` bigint(20) NOT NULL DEFAULT '0',
  `advertiser_id` bigint(20) NOT NULL,
  PRIMARY KEY (`advertiser_gid`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci;
/*!40101 SET character_set_client = @saved_cs_client */;

--
-- Table structure for table `targeting_attribute_counts_by_parent`
--

DROP TABLE IF EXISTS `targeting_attribute_counts_by_parent`;
/*!40101 SET @saved_cs_client     = @@character_set_client */;
/*!40101 SET character_set_client = utf8 */;
CREATE TABLE `targeting_attribute_counts_by_parent` (
  `parent_gid` bigint(20) NOT NULL,
  `active_keywords_count` bigint(20) NOT NULL DEFAULT '0',
  `advertiser_id` bigint(20) NOT NULL,
  `g_advertiser_id` bigint(20) NOT NULL,
  PRIMARY KEY (`parent_gid`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci;
/*!40101 SET character_set_client = @saved_cs_client */;

--
-- Table structure for table `targeting_attribute_history`
--

DROP TABLE IF EXISTS `targeting_attribute_history`;
/*!40101 SET @saved_cs_client     = @@character_set_client */;
/*!40101 SET character_set_client = utf8 */;
CREATE TABLE `targeting_attribute_history` (
  `id` bigint(20) NOT NULL AUTO_INCREMENT,
  `creation_date` timestamp NOT NULL DEFAULT CURRENT_TIMESTAMP,
  `targeting_attribute_id` bigint(20) DEFAULT NULL,
  `targeting_attribute_gid` bigint(20) DEFAULT NULL,
  `advertiser_id` bigint(20) NOT NULL,
  `targeting_attribute_creation_date` timestamp NOT NULL DEFAULT '0000-00-00 00:00:00',
  `targeting_attribute_updated_time` timestamp NOT NULL DEFAULT '0000-00-00 00:00:00',
  `type` tinyint(1) NOT NULL,
  `parent_id` bigint(20) DEFAULT NULL,
  `value` varchar(128) COLLATE utf8mb4_unicode_ci DEFAULT NULL,
  `match_type` tinyint(1) NOT NULL,
  `exclude` tinyint(1) NOT NULL DEFAULT '0',
  `bid` bigint(20) DEFAULT NULL,
  `destination_url` varchar(2048) COLLATE utf8mb4_unicode_ci DEFAULT NULL,
  `properties` mediumtext COLLATE utf8mb4_unicode_ci,
  `deleted` tinyint(1) NOT NULL DEFAULT '0',
  `g_advertiser_id` bigint(20) NOT NULL,
  PRIMARY KEY (`id`),
  KEY `targeting_attribute_id` (`targeting_attribute_id`)
) ENGINE=InnoDB AUTO_INCREMENT=269122472 DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci;
/*!40101 SET character_set_client = @saved_cs_client */;

--
-- Table structure for table `targeting_attributes`
--

DROP TABLE IF EXISTS `targeting_attributes`;
/*!40101 SET @saved_cs_client     = @@character_set_client */;
/*!40101 SET character_set_client = utf8 */;
CREATE TABLE `targeting_attributes` (
  `id` bigint(20) NOT NULL AUTO_INCREMENT,
  `gid` bigint(20) DEFAULT NULL,
  `advertiser_id` bigint(20) NOT NULL,
  `creation_date` timestamp NOT NULL DEFAULT CURRENT_TIMESTAMP,
  `updated_time` timestamp NOT NULL DEFAULT CURRENT_TIMESTAMP,
  `type` tinyint(1) NOT NULL,
  `parent_id` bigint(20) DEFAULT NULL,
  `value` varchar(128) COLLATE utf8mb4_unicode_ci DEFAULT NULL,
  `match_type` tinyint(1) NOT NULL,
  `exclude` tinyint(1) NOT NULL DEFAULT '0',
  `bid` bigint(20) DEFAULT NULL,
  `destination_url` varchar(2048) COLLATE utf8mb4_unicode_ci DEFAULT NULL,
  `properties` mediumtext COLLATE utf8mb4_unicode_ci,
  `deleted` tinyint(1) NOT NULL DEFAULT '0',
  `g_advertiser_id` bigint(20) NOT NULL,
  PRIMARY KEY (`id`),
  KEY `adv_del_type_idx` (`advertiser_id`,`deleted`,`type`),
  KEY `parent_id_2` (`parent_id`,`deleted`,`value`,`id`,`match_type`,`exclude`,`bid`)
) ENGINE=InnoDB AUTO_INCREMENT=314630380 DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci;
/*!40101 SET character_set_client = @saved_cs_client */;

--
-- Table structure for table `targeting_specs`
--

DROP TABLE IF EXISTS `targeting_specs`;
/*!40101 SET @saved_cs_client     = @@character_set_client */;
/*!40101 SET character_set_client = utf8 */;
CREATE TABLE `targeting_specs` (
  `id` bigint(20) NOT NULL AUTO_INCREMENT,
  `targeting` longtext COLLATE utf8mb4_unicode_ci,
  `gid` bigint(20) DEFAULT NULL,
  `advertiser_id` bigint(20) NOT NULL,
  `g_advertiser_id` bigint(20) NOT NULL,
  PRIMARY KEY (`id`)
) ENGINE=InnoDB AUTO_INCREMENT=8280456 DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci;
/*!40101 SET character_set_client = @saved_cs_client */;

--
-- Table structure for table `user_preferences`
--

DROP TABLE IF EXISTS `user_preferences`;
/*!40101 SET @saved_cs_client     = @@character_set_client */;
/*!40101 SET character_set_client = utf8 */;
CREATE TABLE `user_preferences` (
  `id` bigint(20) NOT NULL AUTO_INCREMENT,
  `gid` bigint(20) DEFAULT NULL,
  `advertiser_id` bigint(20) NOT NULL,
  `creation_date` timestamp NOT NULL DEFAULT CURRENT_TIMESTAMP,
  `updated_time` timestamp NOT NULL DEFAULT CURRENT_TIMESTAMP,
  `properties` mediumtext COLLATE utf8mb4_unicode_ci,
  `g_advertiser_id` bigint(20) NOT NULL,
  PRIMARY KEY (`id`),
  KEY `advertiser_id` (`advertiser_id`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci;
/*!40101 SET character_set_client = @saved_cs_client */;
/*!40103 SET TIME_ZONE=@OLD_TIME_ZONE */;

/*!40101 SET SQL_MODE=@OLD_SQL_MODE */;
/*!40014 SET FOREIGN_KEY_CHECKS=@OLD_FOREIGN_KEY_CHECKS */;
/*!40014 SET UNIQUE_CHECKS=@OLD_UNIQUE_CHECKS */;
/*!40101 SET CHARACTER_SET_CLIENT=@OLD_CHARACTER_SET_CLIENT */;
/*!40101 SET CHARACTER_SET_RESULTS=@OLD_CHARACTER_SET_RESULTS */;
/*!40101 SET COLLATION_CONNECTION=@OLD_COLLATION_CONNECTION */;
/*!40111 SET SQL_NOTES=@OLD_SQL_NOTES */;

-- Dump completed on 2019-02-06  0:37:23
