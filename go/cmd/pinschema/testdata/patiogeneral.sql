CREATE TABLE IF NOT EXISTS `accepted_tos_id_idx` (
  `id` bigint(20) NOT NULL DEFAULT '0',
  `g_advertiser_id` bigint(20) DEFAULT NULL,
  primary key(`id`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci COMMENT 'vitess_lookup_vindex';

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
