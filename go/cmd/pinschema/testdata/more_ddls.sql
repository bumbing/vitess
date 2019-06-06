CREATE TABLE IF NOT EXISTS `accepted_tos_id_idx` (
  `id` bigint(20) NOT NULL DEFAULT '0',
  `g_advertiser_id` bigint(20) DEFAULT NULL,
  primary key(`id`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci COMMENT 'vitess_lookup_vindex';
