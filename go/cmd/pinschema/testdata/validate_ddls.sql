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


DROP TABLE IF EXISTS `accepted_tos_seq`;
/*!40101 SET @saved_cs_client     = @@character_set_client */;
/*!40101 SET character_set_client = utf8 */;
CREATE TABLE `accepted_tos_seq` (
  `id` int(11) NOT NULL DEFAULT '0',
  `next_id` bigint(20) DEFAULT NULL,
  `cache` bigint(20) DEFAULT NULL,
  PRIMARY KEY (`id`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci COMMENT='vitess_sequence';
/*!40101 SET character_set_client = @saved_cs_client */;

CREATE TABLE `accepted_tos_id_idx` (
  `id` bigint(20) NOT NULL DEFAULT '0',
  `g_advertiser_id` bigint(20) DEFAULT NULL,
  PRIMARY KEY (`id`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci COMMENT='vitess_lookup_vindex';
