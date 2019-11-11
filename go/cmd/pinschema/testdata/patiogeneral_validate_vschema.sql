CREATE TABLE `_drop_me_pin_promotion_specs_labels_xref_id_idx` (
  `id` bigint(20) NOT NULL DEFAULT '0',
  `g_advertiser_id` bigint(20) DEFAULT NULL,
  PRIMARY KEY (`id`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci COMMENT='vitess_lookup_vindex';
CREATE TABLE `accepted_tos_id_idx` (
  `id` bigint(20) NOT NULL DEFAULT '0',
  `g_advertiser_id` bigint(20) DEFAULT NULL,
  PRIMARY KEY (`id`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci COMMENT='vitess_lookup_vindex';
CREATE TABLE `accepted_tos_seq` (
  `id` int(11) NOT NULL DEFAULT '0',
  `next_id` bigint(20) DEFAULT NULL,
  `cache` bigint(20) DEFAULT NULL,
  PRIMARY KEY (`id`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci COMMENT='vitess_sequence';
CREATE TABLE `ad_group_id_idx` (
  `id` bigint(20) NOT NULL DEFAULT '0',
  `g_advertiser_id` bigint(20) DEFAULT NULL,
  PRIMARY KEY (`id`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci COMMENT='vitess_lookup_vindex';
CREATE TABLE `ad_group_spec_id_idx` (
  `id` bigint(20) NOT NULL DEFAULT '0',
  `g_advertiser_id` bigint(20) DEFAULT NULL,
  PRIMARY KEY (`id`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci COMMENT='vitess_lookup_vindex';
CREATE TABLE `ad_group_specs_seq` (
  `id` int(11) NOT NULL DEFAULT '0',
  `next_id` bigint(20) DEFAULT NULL,
  `cache` bigint(20) DEFAULT NULL,
  PRIMARY KEY (`id`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci COMMENT='vitess_sequence';
CREATE TABLE `ad_groups_history_id_idx` (
  `id` bigint(20) NOT NULL DEFAULT '0',
  `g_advertiser_id` bigint(20) DEFAULT NULL,
  PRIMARY KEY (`id`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci COMMENT='vitess_lookup_vindex';
CREATE TABLE `ad_groups_history_seq` (
  `id` int(11) NOT NULL DEFAULT '0',
  `next_id` bigint(20) DEFAULT NULL,
  `cache` bigint(20) DEFAULT NULL,
  PRIMARY KEY (`id`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci COMMENT='vitess_sequence';
CREATE TABLE `ad_groups_seq` (
  `id` int(11) NOT NULL DEFAULT '0',
  `next_id` bigint(20) DEFAULT NULL,
  `cache` bigint(20) DEFAULT NULL,
  PRIMARY KEY (`id`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci COMMENT='vitess_sequence';
CREATE TABLE `ads_manager_preferences` (
  `id` bigint(20) NOT NULL AUTO_INCREMENT,
  `gid` bigint(20) DEFAULT NULL,
  `creation_date` timestamp NOT NULL DEFAULT CURRENT_TIMESTAMP,
  `updated_time` timestamp NOT NULL DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP,
  `advertiser_id` bigint(20) NOT NULL,
  `g_advertiser_id` bigint(20) NOT NULL,
  `user_id` bigint(20) NOT NULL,
  `preference_key` varchar(128) DEFAULT NULL,
  `level` smallint(6) NOT NULL,
  `preferences` text NOT NULL,
  PRIMARY KEY (`id`),
  UNIQUE KEY `unique_preference` (`advertiser_id`,`user_id`,`preference_key`,`level`),
  KEY `gid` (`gid`)
) ENGINE=InnoDB DEFAULT CHARSET=latin1;
CREATE TABLE `advertiser_conversion_event_id_idx` (
  `id` bigint(20) NOT NULL DEFAULT '0',
  `g_advertiser_id` bigint(20) DEFAULT NULL,
  PRIMARY KEY (`id`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci COMMENT='vitess_lookup_vindex';
CREATE TABLE `advertiser_conversion_events_seq` (
  `id` int(11) NOT NULL DEFAULT '0',
  `next_id` bigint(20) DEFAULT NULL,
  `cache` bigint(20) DEFAULT NULL,
  PRIMARY KEY (`id`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci COMMENT='vitess_sequence';
CREATE TABLE `advertiser_discount_id_idx` (
  `id` bigint(20) NOT NULL DEFAULT '0',
  `g_advertiser_id` bigint(20) DEFAULT NULL,
  PRIMARY KEY (`id`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci COMMENT='vitess_lookup_vindex';
CREATE TABLE `advertiser_discounts_seq` (
  `id` int(11) NOT NULL DEFAULT '0',
  `next_id` bigint(20) DEFAULT NULL,
  `cache` bigint(20) DEFAULT NULL,
  PRIMARY KEY (`id`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci COMMENT='vitess_sequence';
CREATE TABLE `advertiser_labeling_result_id_idx` (
  `id` bigint(20) NOT NULL DEFAULT '0',
  `g_advertiser_id` bigint(20) DEFAULT NULL,
  PRIMARY KEY (`id`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci COMMENT='vitess_lookup_vindex';
CREATE TABLE `advertiser_labeling_results_seq` (
  `id` int(11) NOT NULL DEFAULT '0',
  `next_id` bigint(20) DEFAULT NULL,
  `cache` bigint(20) DEFAULT NULL,
  PRIMARY KEY (`id`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci COMMENT='vitess_sequence';
CREATE TABLE `advertiser_labels` (
  `id` bigint(20) unsigned NOT NULL AUTO_INCREMENT,
  `label_name` varchar(64) COLLATE utf8mb4_unicode_ci DEFAULT NULL,
  `status` tinyint(3) unsigned DEFAULT NULL,
  PRIMARY KEY (`id`),
  UNIQUE KEY `label_idx` (`label_name`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci;
CREATE TABLE `advertiser_sales_info_seq` (
  `id` int(11) NOT NULL DEFAULT '0',
  `next_id` bigint(20) DEFAULT NULL,
  `cache` bigint(20) DEFAULT NULL,
  PRIMARY KEY (`id`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci COMMENT='vitess_sequence';
CREATE TABLE `advertisers_history_id_idx` (
  `id` bigint(20) NOT NULL DEFAULT '0',
  `g_advertiser_id` bigint(20) DEFAULT NULL,
  PRIMARY KEY (`id`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci COMMENT='vitess_lookup_vindex';
CREATE TABLE `advertisers_history_seq` (
  `id` int(11) NOT NULL DEFAULT '0',
  `next_id` bigint(20) DEFAULT NULL,
  `cache` bigint(20) DEFAULT NULL,
  PRIMARY KEY (`id`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci COMMENT='vitess_sequence';
CREATE TABLE `advertisers_sales_info_id_idx` (
  `id` bigint(20) NOT NULL DEFAULT '0',
  `g_advertiser_id` bigint(20) DEFAULT NULL,
  PRIMARY KEY (`id`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci COMMENT='vitess_lookup_vindex';
CREATE TABLE `advertisers_seq` (
  `id` int(11) NOT NULL DEFAULT '0',
  `next_id` bigint(20) DEFAULT NULL,
  `cache` bigint(20) DEFAULT NULL,
  PRIMARY KEY (`id`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci COMMENT='vitess_sequence';
CREATE TABLE `app_event_tracking_config_id_idx` (
  `id` bigint(20) NOT NULL DEFAULT '0',
  `g_advertiser_id` bigint(20) DEFAULT NULL,
  PRIMARY KEY (`id`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci COMMENT='vitess_lookup_vindex';
CREATE TABLE `app_event_tracking_configs_seq` (
  `id` int(11) NOT NULL DEFAULT '0',
  `next_id` bigint(20) DEFAULT NULL,
  `cache` bigint(20) DEFAULT NULL,
  PRIMARY KEY (`id`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci COMMENT='vitess_sequence';
CREATE TABLE `bill_detail_id_idx` (
  `id` bigint(20) NOT NULL DEFAULT '0',
  `g_advertiser_id` bigint(20) DEFAULT NULL,
  PRIMARY KEY (`id`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci COMMENT='vitess_lookup_vindex';
CREATE TABLE `bill_details_seq` (
  `id` int(11) NOT NULL DEFAULT '0',
  `next_id` bigint(20) DEFAULT NULL,
  `cache` bigint(20) DEFAULT NULL,
  PRIMARY KEY (`id`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci COMMENT='vitess_sequence';
CREATE TABLE `bill_id_idx` (
  `id` bigint(20) NOT NULL DEFAULT '0',
  `g_advertiser_id` bigint(20) DEFAULT NULL,
  PRIMARY KEY (`id`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci COMMENT='vitess_lookup_vindex';
CREATE TABLE `billing_action_id_idx` (
  `id` bigint(20) NOT NULL DEFAULT '0',
  `g_advertiser_id` bigint(20) DEFAULT NULL,
  PRIMARY KEY (`id`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci COMMENT='vitess_lookup_vindex';
CREATE TABLE `billing_actions_seq` (
  `id` int(11) NOT NULL DEFAULT '0',
  `next_id` bigint(20) DEFAULT NULL,
  `cache` bigint(20) DEFAULT NULL,
  PRIMARY KEY (`id`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci COMMENT='vitess_sequence';
CREATE TABLE `billing_contact_id_idx` (
  `id` bigint(20) NOT NULL DEFAULT '0',
  `g_advertiser_id` bigint(20) DEFAULT NULL,
  PRIMARY KEY (`id`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci COMMENT='vitess_lookup_vindex';
CREATE TABLE `billing_contacts_seq` (
  `id` int(11) NOT NULL DEFAULT '0',
  `next_id` bigint(20) DEFAULT NULL,
  `cache` bigint(20) DEFAULT NULL,
  PRIMARY KEY (`id`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci COMMENT='vitess_sequence';
CREATE TABLE `billing_profile_id_idx` (
  `id` bigint(20) NOT NULL DEFAULT '0',
  `g_advertiser_id` bigint(20) DEFAULT NULL,
  PRIMARY KEY (`id`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci COMMENT='vitess_lookup_vindex';
CREATE TABLE `billing_profiles_seq` (
  `id` int(11) NOT NULL DEFAULT '0',
  `next_id` bigint(20) DEFAULT NULL,
  `cache` bigint(20) DEFAULT NULL,
  PRIMARY KEY (`id`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci COMMENT='vitess_sequence';
CREATE TABLE `bills_seq` (
  `id` int(11) NOT NULL DEFAULT '0',
  `next_id` bigint(20) DEFAULT NULL,
  `cache` bigint(20) DEFAULT NULL,
  PRIMARY KEY (`id`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci COMMENT='vitess_sequence';
CREATE TABLE `bulk_v2_job_id_idx` (
  `id` bigint(20) NOT NULL DEFAULT '0',
  `g_advertiser_id` bigint(20) DEFAULT NULL,
  PRIMARY KEY (`id`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci COMMENT='vitess_lookup_vindex';
CREATE TABLE `bulk_v2_jobs_seq` (
  `id` int(11) NOT NULL DEFAULT '0',
  `next_id` bigint(20) DEFAULT NULL,
  `cache` bigint(20) DEFAULT NULL,
  PRIMARY KEY (`id`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci COMMENT='vitess_sequence';
CREATE TABLE `business_profile_id_idx` (
  `id` bigint(20) NOT NULL DEFAULT '0',
  `g_advertiser_id` bigint(20) DEFAULT NULL,
  PRIMARY KEY (`id`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci COMMENT='vitess_lookup_vindex';
CREATE TABLE `business_profiles_seq` (
  `id` int(11) NOT NULL DEFAULT '0',
  `next_id` bigint(20) DEFAULT NULL,
  `cache` bigint(20) DEFAULT NULL,
  PRIMARY KEY (`id`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci COMMENT='vitess_sequence';
CREATE TABLE `campaign_id_idx` (
  `id` bigint(20) NOT NULL DEFAULT '0',
  `g_advertiser_id` bigint(20) DEFAULT NULL,
  PRIMARY KEY (`id`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci COMMENT='vitess_lookup_vindex';
CREATE TABLE `campaign_spec_id_idx` (
  `id` bigint(20) NOT NULL DEFAULT '0',
  `g_advertiser_id` bigint(20) DEFAULT NULL,
  PRIMARY KEY (`id`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci COMMENT='vitess_lookup_vindex';
CREATE TABLE `campaign_specs_seq` (
  `id` int(11) NOT NULL DEFAULT '0',
  `next_id` bigint(20) DEFAULT NULL,
  `cache` bigint(20) DEFAULT NULL,
  PRIMARY KEY (`id`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci COMMENT='vitess_sequence';
CREATE TABLE `campaigns_history_id_idx` (
  `id` bigint(20) NOT NULL DEFAULT '0',
  `g_advertiser_id` bigint(20) DEFAULT NULL,
  PRIMARY KEY (`id`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci COMMENT='vitess_lookup_vindex';
CREATE TABLE `campaigns_history_seq` (
  `id` int(11) NOT NULL DEFAULT '0',
  `next_id` bigint(20) DEFAULT NULL,
  `cache` bigint(20) DEFAULT NULL,
  PRIMARY KEY (`id`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci COMMENT='vitess_sequence';
CREATE TABLE `campaigns_seq` (
  `id` int(11) NOT NULL DEFAULT '0',
  `next_id` bigint(20) DEFAULT NULL,
  `cache` bigint(20) DEFAULT NULL,
  PRIMARY KEY (`id`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci COMMENT='vitess_sequence';
CREATE TABLE `carousel_slot_promotion_id_idx` (
  `id` bigint(20) NOT NULL DEFAULT '0',
  `g_advertiser_id` bigint(20) DEFAULT NULL,
  PRIMARY KEY (`id`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci COMMENT='vitess_lookup_vindex';
CREATE TABLE `carousel_slot_promotions_seq` (
  `id` int(11) NOT NULL DEFAULT '0',
  `next_id` bigint(20) DEFAULT NULL,
  `cache` bigint(20) DEFAULT NULL,
  PRIMARY KEY (`id`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci COMMENT='vitess_sequence';
CREATE TABLE `completed_report_execution_id_idx` (
  `id` bigint(20) NOT NULL DEFAULT '0',
  `g_advertiser_id` bigint(20) DEFAULT NULL,
  PRIMARY KEY (`id`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci COMMENT='vitess_lookup_vindex';
CREATE TABLE `completed_report_executions_seq` (
  `id` int(11) NOT NULL DEFAULT '0',
  `next_id` bigint(20) DEFAULT NULL,
  `cache` bigint(20) DEFAULT NULL,
  PRIMARY KEY (`id`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci COMMENT='vitess_sequence';
CREATE TABLE `conversion_tag_id_idx` (
  `id` bigint(20) NOT NULL DEFAULT '0',
  `g_advertiser_id` bigint(20) DEFAULT NULL,
  PRIMARY KEY (`id`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci COMMENT='vitess_lookup_vindex';
CREATE TABLE `conversion_tags_seq` (
  `id` int(11) NOT NULL DEFAULT '0',
  `next_id` bigint(20) DEFAULT NULL,
  `cache` bigint(20) DEFAULT NULL,
  PRIMARY KEY (`id`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci COMMENT='vitess_sequence';
CREATE TABLE `conversion_tags_v3` (
  `id` bigint(20) NOT NULL,
  `gid` bigint(20) NOT NULL,
  `owner_user_id` bigint(20) NOT NULL,
  `advertiser_id` bigint(20) DEFAULT NULL,
  `creation_date` timestamp NOT NULL DEFAULT CURRENT_TIMESTAMP,
  `deleted` tinyint(1) NOT NULL,
  `properties` mediumtext COLLATE utf8mb4_unicode_ci,
  `version` varchar(10) COLLATE utf8mb4_unicode_ci NOT NULL,
  `g_advertiser_id` bigint(20) DEFAULT NULL,
  `verification_status` smallint(6) NOT NULL DEFAULT '2',
  PRIMARY KEY (`id`),
  UNIQUE KEY `gid` (`gid`),
  KEY `owner_user_id` (`owner_user_id`),
  KEY `advertiser_id` (`advertiser_id`),
  KEY `g_advertiser_id` (`g_advertiser_id`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci;
CREATE TABLE `dark_write_conversion_tags_v3` (
  `id` bigint(20) NOT NULL,
  `gid` bigint(20) NOT NULL,
  `owner_user_id` bigint(20) NOT NULL,
  `advertiser_id` bigint(20) DEFAULT NULL,
  `creation_date` timestamp NOT NULL DEFAULT CURRENT_TIMESTAMP,
  `deleted` tinyint(1) NOT NULL,
  `properties` mediumtext COLLATE utf8mb4_unicode_ci,
  `version` varchar(10) COLLATE utf8mb4_unicode_ci NOT NULL,
  `g_advertiser_id` bigint(20) DEFAULT NULL,
  `verification_status` smallint(6) NOT NULL DEFAULT '2',
  PRIMARY KEY (`id`),
  UNIQUE KEY `gid` (`gid`),
  KEY `owner_user_id` (`owner_user_id`),
  KEY `advertiser_id` (`advertiser_id`),
  KEY `g_advertiser_id` (`g_advertiser_id`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci;
CREATE TABLE `dark_write_exchange_rates` (
  `exchange_timestamp` bigint(20) NOT NULL,
  `properties` mediumtext COLLATE utf8mb4_unicode_ci,
  PRIMARY KEY (`exchange_timestamp`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci;
CREATE TABLE `dark_write_user_lists` (
  `id` bigint(20) NOT NULL AUTO_INCREMENT,
  `creation_date` timestamp NOT NULL DEFAULT CURRENT_TIMESTAMP,
  `g_advertiser_id` bigint(20) NOT NULL,
  `properties` mediumtext COLLATE utf8mb4_unicode_ci,
  `updated_time` timestamp NOT NULL DEFAULT CURRENT_TIMESTAMP,
  `num_batches` bigint(20) NOT NULL DEFAULT '0',
  `num_processed_batches` bigint(20) NOT NULL DEFAULT '0',
  `num_uploaded_user_emails` bigint(20) NOT NULL DEFAULT '0',
  `num_removed_user_emails` bigint(20) NOT NULL DEFAULT '0',
  `num_matched_user_emails` bigint(20) NOT NULL DEFAULT '0',
  PRIMARY KEY (`id`),
  KEY `g_advertiser_id` (`g_advertiser_id`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci;
CREATE TABLE `exchange_rates` (
  `exchange_timestamp` bigint(20) NOT NULL,
  `properties` mediumtext COLLATE utf8mb4_unicode_ci,
  PRIMARY KEY (`exchange_timestamp`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci;
CREATE TABLE `goal_id_idx` (
  `id` bigint(20) NOT NULL DEFAULT '0',
  `g_advertiser_id` bigint(20) DEFAULT NULL,
  PRIMARY KEY (`id`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci COMMENT='vitess_lookup_vindex';
CREATE TABLE `goals_seq` (
  `id` int(11) NOT NULL DEFAULT '0',
  `next_id` bigint(20) DEFAULT NULL,
  `cache` bigint(20) DEFAULT NULL,
  PRIMARY KEY (`id`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci COMMENT='vitess_sequence';
CREATE TABLE `marketing_offers` (
  `id` bigint(20) NOT NULL AUTO_INCREMENT,
  `creation_date` timestamp NOT NULL DEFAULT CURRENT_TIMESTAMP,
  `active` tinyint(1) NOT NULL DEFAULT '0',
  `title` varchar(128) COLLATE utf8mb4_unicode_ci DEFAULT NULL,
  `discount_in_micro_currency` bigint(20) NOT NULL,
  `redemption_restrictions` mediumtext COLLATE utf8mb4_unicode_ci,
  `discount_properties` mediumtext COLLATE utf8mb4_unicode_ci,
  `attribution` mediumtext COLLATE utf8mb4_unicode_ci,
  `type` varchar(64) COLLATE utf8mb4_unicode_ci NOT NULL,
  `offer_code` varchar(36) CHARACTER SET utf8mb4 COLLATE utf8mb4_bin NOT NULL,
  `currency` smallint(6) NOT NULL DEFAULT '1',
  PRIMARY KEY (`id`),
  UNIQUE KEY `offer_code_currency_key` (`offer_code`,`currency`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci;
CREATE TABLE `metric_report_template_id_idx` (
  `id` bigint(20) NOT NULL DEFAULT '0',
  `g_advertiser_id` bigint(20) DEFAULT NULL,
  PRIMARY KEY (`id`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci COMMENT='vitess_lookup_vindex';
CREATE TABLE `metrics_report_templates_seq` (
  `id` int(11) NOT NULL DEFAULT '0',
  `next_id` bigint(20) DEFAULT NULL,
  `cache` bigint(20) DEFAULT NULL,
  PRIMARY KEY (`id`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci COMMENT='vitess_sequence';
CREATE TABLE `notification_id_idx` (
  `id` bigint(20) NOT NULL DEFAULT '0',
  `g_advertiser_id` bigint(20) DEFAULT NULL,
  PRIMARY KEY (`id`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci COMMENT='vitess_lookup_vindex';
CREATE TABLE `notifications_seq` (
  `id` int(11) NOT NULL DEFAULT '0',
  `next_id` bigint(20) DEFAULT NULL,
  `cache` bigint(20) DEFAULT NULL,
  PRIMARY KEY (`id`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci COMMENT='vitess_sequence';
CREATE TABLE `order_line_id_idx` (
  `id` bigint(20) NOT NULL DEFAULT '0',
  `g_advertiser_id` bigint(20) DEFAULT NULL,
  PRIMARY KEY (`id`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci COMMENT='vitess_lookup_vindex';
CREATE TABLE `order_line_spec_id_idx` (
  `id` bigint(20) NOT NULL DEFAULT '0',
  `g_advertiser_id` bigint(20) DEFAULT NULL,
  PRIMARY KEY (`id`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci COMMENT='vitess_lookup_vindex';
CREATE TABLE `order_line_specs_seq` (
  `id` int(11) NOT NULL DEFAULT '0',
  `next_id` bigint(20) DEFAULT NULL,
  `cache` bigint(20) DEFAULT NULL,
  PRIMARY KEY (`id`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci COMMENT='vitess_sequence';
CREATE TABLE `order_lines_seq` (
  `id` int(11) NOT NULL DEFAULT '0',
  `next_id` bigint(20) DEFAULT NULL,
  `cache` bigint(20) DEFAULT NULL,
  PRIMARY KEY (`id`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci COMMENT='vitess_sequence';
CREATE TABLE `owner_to_advertiser` (
  `owner_user_id` bigint(20) NOT NULL,
  `g_advertiser_id` bigint(20) NOT NULL,
  `creation_date` timestamp NOT NULL DEFAULT CURRENT_TIMESTAMP,
  PRIMARY KEY (`owner_user_id`,`g_advertiser_id`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci;
CREATE TABLE `pin_promotion_id_idx` (
  `id` bigint(20) NOT NULL DEFAULT '0',
  `g_advertiser_id` bigint(20) DEFAULT NULL,
  PRIMARY KEY (`id`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci COMMENT='vitess_lookup_vindex';
CREATE TABLE `pin_promotion_label_id_idx` (
  `id` bigint(20) NOT NULL DEFAULT '0',
  `g_advertiser_id` bigint(20) DEFAULT NULL,
  PRIMARY KEY (`id`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci COMMENT='vitess_lookup_vindex';
CREATE TABLE `pin_promotion_labels_seq` (
  `id` int(11) NOT NULL DEFAULT '0',
  `next_id` bigint(20) DEFAULT NULL,
  `cache` bigint(20) DEFAULT NULL,
  PRIMARY KEY (`id`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci COMMENT='vitess_sequence';
CREATE TABLE `pin_promotion_spec_id_idx` (
  `id` bigint(20) NOT NULL DEFAULT '0',
  `g_advertiser_id` bigint(20) DEFAULT NULL,
  PRIMARY KEY (`id`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci COMMENT='vitess_lookup_vindex';
CREATE TABLE `pin_promotion_specs_seq` (
  `id` int(11) NOT NULL DEFAULT '0',
  `next_id` bigint(20) DEFAULT NULL,
  `cache` bigint(20) DEFAULT NULL,
  PRIMARY KEY (`id`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci COMMENT='vitess_sequence';
CREATE TABLE `pin_promotions_history_id_idx` (
  `id` bigint(20) NOT NULL DEFAULT '0',
  `g_advertiser_id` bigint(20) DEFAULT NULL,
  PRIMARY KEY (`id`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci COMMENT='vitess_lookup_vindex';
CREATE TABLE `pin_promotions_history_seq` (
  `id` int(11) NOT NULL DEFAULT '0',
  `next_id` bigint(20) DEFAULT NULL,
  `cache` bigint(20) DEFAULT NULL,
  PRIMARY KEY (`id`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci COMMENT='vitess_sequence';
CREATE TABLE `pin_promotions_seq` (
  `id` int(11) NOT NULL DEFAULT '0',
  `next_id` bigint(20) DEFAULT NULL,
  `cache` bigint(20) DEFAULT NULL,
  PRIMARY KEY (`id`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci COMMENT='vitess_sequence';
CREATE TABLE `pinner_list_id_idx` (
  `id` bigint(20) NOT NULL DEFAULT '0',
  `g_advertiser_id` bigint(20) DEFAULT NULL,
  PRIMARY KEY (`id`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci COMMENT='vitess_lookup_vindex';
CREATE TABLE `pinner_list_spec_id_idx` (
  `id` bigint(20) NOT NULL DEFAULT '0',
  `g_advertiser_id` bigint(20) DEFAULT NULL,
  PRIMARY KEY (`id`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci COMMENT='vitess_lookup_vindex';
CREATE TABLE `pinner_list_specs_seq` (
  `id` int(11) NOT NULL DEFAULT '0',
  `next_id` bigint(20) DEFAULT NULL,
  `cache` bigint(20) DEFAULT NULL,
  PRIMARY KEY (`id`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci COMMENT='vitess_sequence';
CREATE TABLE `pinner_lists_seq` (
  `id` int(11) NOT NULL DEFAULT '0',
  `next_id` bigint(20) DEFAULT NULL,
  `cache` bigint(20) DEFAULT NULL,
  PRIMARY KEY (`id`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci COMMENT='vitess_sequence';
CREATE TABLE `planning_moments` (
  `id` bigint(20) NOT NULL AUTO_INCREMENT,
  `gid` bigint(20) DEFAULT NULL,
  `name` varchar(50) COLLATE utf8mb4_unicode_ci NOT NULL,
  `description` varchar(200) COLLATE utf8mb4_unicode_ci DEFAULT NULL,
  `creation_date` timestamp NOT NULL DEFAULT CURRENT_TIMESTAMP,
  `start_date` timestamp NOT NULL DEFAULT '2000-01-01 00:00:01',
  `end_date` timestamp NOT NULL DEFAULT '2000-12-31 23:59:59',
  `type` tinyint(1) NOT NULL,
  `status` tinyint(1) NOT NULL DEFAULT '1',
  `user_list_id` bigint(20) NOT NULL,
  `pinner_list_id` bigint(20) NOT NULL,
  `g_pinner_list_id` bigint(20) NOT NULL,
  `properties` mediumtext COLLATE utf8mb4_unicode_ci,
  `updated_time` timestamp NOT NULL DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP,
  PRIMARY KEY (`id`),
  KEY `user_list_id` (`user_list_id`),
  CONSTRAINT `user_list_id` FOREIGN KEY (`user_list_id`) REFERENCES `user_lists` (`id`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci;
CREATE TABLE `product_group_id_idx` (
  `id` bigint(20) NOT NULL DEFAULT '0',
  `g_advertiser_id` bigint(20) DEFAULT NULL,
  PRIMARY KEY (`id`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci COMMENT='vitess_lookup_vindex';
CREATE TABLE `product_group_spec_id_idx` (
  `id` bigint(20) NOT NULL DEFAULT '0',
  `g_advertiser_id` bigint(20) DEFAULT NULL,
  PRIMARY KEY (`id`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci COMMENT='vitess_lookup_vindex';
CREATE TABLE `product_group_specs_seq` (
  `id` int(11) NOT NULL DEFAULT '0',
  `next_id` bigint(20) DEFAULT NULL,
  `cache` bigint(20) DEFAULT NULL,
  PRIMARY KEY (`id`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci COMMENT='vitess_sequence';
CREATE TABLE `product_groups_seq` (
  `id` int(11) NOT NULL DEFAULT '0',
  `next_id` bigint(20) DEFAULT NULL,
  `cache` bigint(20) DEFAULT NULL,
  PRIMARY KEY (`id`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci COMMENT='vitess_sequence';
CREATE TABLE `promoted_catalog_product_group_id_idx` (
  `id` bigint(20) NOT NULL DEFAULT '0',
  `g_advertiser_id` bigint(20) DEFAULT NULL,
  PRIMARY KEY (`id`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci COMMENT='vitess_lookup_vindex';
CREATE TABLE `promoted_catalog_product_groups_history_id_idx` (
  `id` bigint(20) NOT NULL DEFAULT '0',
  `g_advertiser_id` bigint(20) DEFAULT NULL,
  PRIMARY KEY (`id`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci COMMENT='vitess_lookup_vindex';
CREATE TABLE `promoted_catalog_product_groups_history_seq` (
  `id` int(11) NOT NULL DEFAULT '0',
  `next_id` bigint(20) DEFAULT NULL,
  `cache` bigint(20) DEFAULT NULL,
  PRIMARY KEY (`id`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci COMMENT='vitess_sequence';
CREATE TABLE `promoted_catalog_product_groups_seq` (
  `id` int(11) NOT NULL DEFAULT '0',
  `next_id` bigint(20) DEFAULT NULL,
  `cache` bigint(20) DEFAULT NULL,
  PRIMARY KEY (`id`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci COMMENT='vitess_sequence';
CREATE TABLE `review_labels` (
  `id` bigint(20) unsigned NOT NULL AUTO_INCREMENT,
  `label_name` varchar(64) COLLATE utf8mb4_unicode_ci DEFAULT NULL,
  `block_rules` mediumtext COLLATE utf8mb4_unicode_ci,
  `status` tinyint(3) unsigned DEFAULT NULL,
  PRIMARY KEY (`id`),
  UNIQUE KEY `label_idx` (`label_name`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci;
CREATE TABLE `rule_subscription_id_idx` (
  `id` bigint(20) NOT NULL DEFAULT '0',
  `g_advertiser_id` bigint(20) DEFAULT NULL,
  PRIMARY KEY (`id`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci COMMENT='vitess_lookup_vindex';
CREATE TABLE `rule_subscriptions_seq` (
  `id` int(11) NOT NULL DEFAULT '0',
  `next_id` bigint(20) DEFAULT NULL,
  `cache` bigint(20) DEFAULT NULL,
  PRIMARY KEY (`id`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci COMMENT='vitess_sequence';
CREATE TABLE `scheduled_report_chunk_execution_id_idx` (
  `id` bigint(20) NOT NULL DEFAULT '0',
  `g_advertiser_id` bigint(20) DEFAULT NULL,
  PRIMARY KEY (`id`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci COMMENT='vitess_lookup_vindex';
CREATE TABLE `scheduled_report_chunk_executions_seq` (
  `id` int(11) NOT NULL DEFAULT '0',
  `next_id` bigint(20) DEFAULT NULL,
  `cache` bigint(20) DEFAULT NULL,
  PRIMARY KEY (`id`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci COMMENT='vitess_sequence';
CREATE TABLE `scheduled_report_id_idx` (
  `id` bigint(20) NOT NULL DEFAULT '0',
  `g_advertiser_id` bigint(20) DEFAULT NULL,
  PRIMARY KEY (`id`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci COMMENT='vitess_lookup_vindex';
CREATE TABLE `scheduled_reports_seq` (
  `id` bigint(11) NOT NULL DEFAULT '0',
  `next_id` bigint(20) DEFAULT NULL,
  `cache` bigint(20) DEFAULT NULL,
  PRIMARY KEY (`id`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci COMMENT='vitess_sequence';
CREATE TABLE `targeting_attribute_history_id_idx` (
  `id` bigint(20) NOT NULL DEFAULT '0',
  `g_advertiser_id` bigint(20) DEFAULT NULL,
  PRIMARY KEY (`id`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci COMMENT='vitess_lookup_vindex';
CREATE TABLE `targeting_attribute_history_seq` (
  `id` int(11) NOT NULL DEFAULT '0',
  `next_id` bigint(20) DEFAULT NULL,
  `cache` bigint(20) DEFAULT NULL,
  PRIMARY KEY (`id`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci COMMENT='vitess_sequence';
CREATE TABLE `targeting_attribute_id_idx` (
  `id` bigint(20) NOT NULL DEFAULT '0',
  `g_advertiser_id` bigint(20) DEFAULT NULL,
  PRIMARY KEY (`id`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci COMMENT='vitess_lookup_vindex';
CREATE TABLE `targeting_attributes_seq` (
  `id` int(11) NOT NULL DEFAULT '0',
  `next_id` bigint(20) DEFAULT NULL,
  `cache` bigint(20) DEFAULT NULL,
  PRIMARY KEY (`id`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci COMMENT='vitess_sequence';
CREATE TABLE `targeting_spec_id_idx` (
  `id` bigint(20) NOT NULL DEFAULT '0',
  `g_advertiser_id` bigint(20) DEFAULT NULL,
  PRIMARY KEY (`id`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci COMMENT='vitess_lookup_vindex';
CREATE TABLE `targeting_specs_seq` (
  `id` int(11) NOT NULL DEFAULT '0',
  `next_id` bigint(20) DEFAULT NULL,
  `cache` bigint(20) DEFAULT NULL,
  PRIMARY KEY (`id`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci COMMENT='vitess_sequence';
CREATE TABLE `user_lists` (
  `id` bigint(20) NOT NULL AUTO_INCREMENT,
  `creation_date` timestamp NOT NULL DEFAULT CURRENT_TIMESTAMP,
  `g_advertiser_id` bigint(20) NOT NULL,
  `properties` mediumtext COLLATE utf8mb4_unicode_ci,
  `updated_time` timestamp NOT NULL DEFAULT CURRENT_TIMESTAMP,
  `num_batches` bigint(20) NOT NULL DEFAULT '0',
  `num_processed_batches` bigint(20) NOT NULL DEFAULT '0',
  `num_uploaded_user_emails` bigint(20) NOT NULL DEFAULT '0',
  `num_removed_user_emails` bigint(20) NOT NULL DEFAULT '0',
  `num_matched_user_emails` bigint(20) NOT NULL DEFAULT '0',
  PRIMARY KEY (`id`),
  KEY `g_advertiser_id` (`g_advertiser_id`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci;
CREATE TABLE `user_preference_id_idx` (
  `id` bigint(20) NOT NULL DEFAULT '0',
  `g_advertiser_id` bigint(20) DEFAULT NULL,
  PRIMARY KEY (`id`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci COMMENT='vitess_lookup_vindex';
CREATE TABLE `user_preferences_seq` (
  `id` int(11) NOT NULL DEFAULT '0',
  `next_id` bigint(20) DEFAULT NULL,
  `cache` bigint(20) DEFAULT NULL,
  PRIMARY KEY (`id`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci COMMENT='vitess_sequence';
CREATE TABLE `metrics_report_template_id_idx` (
  `id` bigint(20) NOT NULL DEFAULT '0',
  `g_advertiser_id` bigint(20) DEFAULT NULL,
  PRIMARY KEY (`id`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci COMMENT='vitess_lookup_vindex';
CREATE TABLE `advertiser_sales_info_id_idx` (
  `id` bigint(20) NOT NULL DEFAULT '0',
  `g_advertiser_id` bigint(20) DEFAULT NULL,
  PRIMARY KEY (`id`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci COMMENT='vitess_lookup_vindex';
