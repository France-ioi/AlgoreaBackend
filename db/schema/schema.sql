-- MySQL dump 10.13  Distrib 8.3.0, for macos12.6 (x86_64)
--
-- Host: 127.0.0.1    Database: algorea_db
-- ------------------------------------------------------
-- Server version	8.0.34

/*!40101 SET @OLD_CHARACTER_SET_CLIENT=@@CHARACTER_SET_CLIENT */;
/*!40101 SET @OLD_CHARACTER_SET_RESULTS=@@CHARACTER_SET_RESULTS */;
/*!40101 SET @OLD_COLLATION_CONNECTION=@@COLLATION_CONNECTION */;
/*!50503 SET NAMES utf8mb4 */;
/*!40103 SET @OLD_TIME_ZONE=@@TIME_ZONE */;
/*!40103 SET TIME_ZONE='+00:00' */;
/*!40014 SET @OLD_UNIQUE_CHECKS=@@UNIQUE_CHECKS, UNIQUE_CHECKS=0 */;
/*!40014 SET @OLD_FOREIGN_KEY_CHECKS=@@FOREIGN_KEY_CHECKS, FOREIGN_KEY_CHECKS=0 */;
/*!40101 SET @OLD_SQL_MODE=@@SQL_MODE, SQL_MODE='NO_AUTO_VALUE_ON_ZERO' */;
/*!40111 SET @OLD_SQL_NOTES=@@SQL_NOTES, SQL_NOTES=0 */;

--
-- Table structure for table `access_tokens`
--

DROP TABLE IF EXISTS `access_tokens`;
/*!40101 SET @saved_cs_client     = @@character_set_client */;
/*!50503 SET character_set_client = utf8mb4 */;
CREATE TABLE `access_tokens` (
  `token` varbinary(2000) NOT NULL COMMENT 'The access token.',
  `session_id` bigint NOT NULL,
  `expires_at` datetime NOT NULL COMMENT 'The time the token expires and becomes invalid. It should be deleted after this time.',
  `issued_at` datetime NOT NULL DEFAULT CURRENT_TIMESTAMP COMMENT 'The time the token was issued.',
  KEY `expires_at` (`expires_at`),
  KEY `token` (`token`(767)),
  KEY `fk_access_tokens_sessions_session_id` (`session_id`),
  CONSTRAINT `fk_access_tokens_sessions_session_id` FOREIGN KEY (`session_id`) REFERENCES `sessions` (`session_id`) ON DELETE CASCADE
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_0900_ai_ci COMMENT='Access tokens (short lifetime) distributed to users, to access a specific session.';
/*!40101 SET character_set_client = @saved_cs_client */;

--
-- Dumping data for table `access_tokens`
--

LOCK TABLES `access_tokens` WRITE;
/*!40000 ALTER TABLE `access_tokens` DISABLE KEYS */;
/*!40000 ALTER TABLE `access_tokens` ENABLE KEYS */;
UNLOCK TABLES;

--
-- Table structure for table `answers`
--

DROP TABLE IF EXISTS `answers`;
/*!40101 SET @saved_cs_client     = @@character_set_client */;
/*!50503 SET character_set_client = utf8mb4 */;
CREATE TABLE `answers` (
  `id` bigint NOT NULL AUTO_INCREMENT,
  `participant_id` bigint NOT NULL,
  `attempt_id` bigint NOT NULL,
  `item_id` bigint NOT NULL,
  `author_id` bigint NOT NULL,
  `type` enum('Submission','Saved','Current') NOT NULL COMMENT '''Submission'' for answers submitted for grading, ''Saved'' for manual backups of answers, ''Current'' for automatic snapshots of the latest answers (unique for a user on an attempt)',
  `state` mediumtext COMMENT 'Saved state (sent by the task platform)',
  `answer` mediumtext COMMENT 'Saved answer (sent by the task platform)',
  `created_at` datetime NOT NULL COMMENT 'Submission time',
  PRIMARY KEY (`id`),
  KEY `user_id` (`author_id`),
  KEY `attempt_id` (`attempt_id`),
  KEY `fk_answers_participant_id_attempt_id_item_id_results` (`participant_id`,`attempt_id`,`item_id`),
  KEY `type_created_at_desc_item_id_participant_id_attempt_id_desc` (`type`,`created_at` DESC,`item_id`,`participant_id`,`attempt_id` DESC),
  KEY `type_participant_id_item_id_created_at_desc_attempt_id_desc` (`type`,`participant_id`,`item_id`,`created_at` DESC,`attempt_id` DESC),
  KEY `created_at_d_item_id_participant_id_attempt_id_d_type_d_id_autho` (`created_at` DESC,`item_id`,`participant_id`,`attempt_id` DESC,`type` DESC,`id`,`author_id`),
  CONSTRAINT `fk_answers_author_id_users_group_id` FOREIGN KEY (`author_id`) REFERENCES `users` (`group_id`) ON DELETE CASCADE,
  CONSTRAINT `fk_answers_participant_id_attempt_id_item_id_results` FOREIGN KEY (`participant_id`, `attempt_id`, `item_id`) REFERENCES `results` (`participant_id`, `attempt_id`, `item_id`) ON DELETE CASCADE
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_0900_ai_ci COMMENT='All the submissions made by users on tasks, as well as saved answers and the current answer';
/*!40101 SET character_set_client = @saved_cs_client */;

--
-- Dumping data for table `answers`
--

LOCK TABLES `answers` WRITE;
/*!40000 ALTER TABLE `answers` DISABLE KEYS */;
/*!40000 ALTER TABLE `answers` ENABLE KEYS */;
UNLOCK TABLES;

--
-- Table structure for table `attempts`
--

DROP TABLE IF EXISTS `attempts`;
/*!40101 SET @saved_cs_client     = @@character_set_client */;
/*!50503 SET character_set_client = utf8mb4 */;
CREATE TABLE `attempts` (
  `participant_id` bigint NOT NULL,
  `id` bigint NOT NULL COMMENT 'Identifier of this attempt for this participant, 0 is the default attempt for the participant, the next ones are sequentially assigned.',
  `creator_id` bigint DEFAULT NULL COMMENT 'The user who created this attempt',
  `parent_attempt_id` bigint DEFAULT NULL COMMENT 'The attempt from which this one was forked. NULL for the default attempt.',
  `root_item_id` bigint DEFAULT NULL COMMENT 'The item on which the attempt was created',
  `created_at` datetime NOT NULL DEFAULT CURRENT_TIMESTAMP COMMENT 'Time at which the attempt was manually created or was first marked as started (should be when it is first visited).',
  `ended_at` datetime DEFAULT NULL COMMENT 'Time at which the attempt was (typically manually) ended',
  `allows_submissions_until` datetime NOT NULL DEFAULT '9999-12-31 23:59:59' COMMENT 'Time until which the participant can submit an answer on this attempt',
  PRIMARY KEY (`participant_id`,`id`),
  KEY `fk_attempts_creator_id_users_group_id` (`creator_id`),
  KEY `fk_attempts_root_item_id_items_id` (`root_item_id`),
  KEY `participant_id_parent_attempt_id_root_item_id` (`participant_id`,`parent_attempt_id`,`root_item_id`),
  KEY `participant_id_root_item_id` (`participant_id`,`root_item_id`),
  CONSTRAINT `fk_attempts_creator_id_users_group_id` FOREIGN KEY (`creator_id`) REFERENCES `users` (`group_id`) ON DELETE SET NULL,
  CONSTRAINT `fk_attempts_participant_id_groups_id` FOREIGN KEY (`participant_id`) REFERENCES `groups` (`id`) ON DELETE CASCADE,
  CONSTRAINT `fk_attempts_root_item_id_items_id` FOREIGN KEY (`root_item_id`) REFERENCES `items` (`id`) ON DELETE SET NULL
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_0900_ai_ci COMMENT='Attempts of participants (team or user) to solve a subtree of items. An attempt may have several answers for a same item. Every participant has a default attempt.';
/*!40101 SET character_set_client = @saved_cs_client */;

--
-- Dumping data for table `attempts`
--

LOCK TABLES `attempts` WRITE;
/*!40000 ALTER TABLE `attempts` DISABLE KEYS */;
/*!40000 ALTER TABLE `attempts` ENABLE KEYS */;
UNLOCK TABLES;

--
-- Table structure for table `badges`
--

DROP TABLE IF EXISTS `badges`;
/*!40101 SET @saved_cs_client     = @@character_set_client */;
/*!50503 SET character_set_client = utf8mb4 */;
CREATE TABLE `badges` (
  `id` bigint NOT NULL AUTO_INCREMENT,
  `user_id` bigint NOT NULL,
  `name` text,
  `code` text NOT NULL,
  PRIMARY KEY (`id`),
  KEY `fk_badges_user_id_users_group_id` (`user_id`),
  CONSTRAINT `fk_badges_user_id_users_group_id` FOREIGN KEY (`user_id`) REFERENCES `users` (`group_id`) ON DELETE CASCADE
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_0900_ai_ci;
/*!40101 SET character_set_client = @saved_cs_client */;

--
-- Dumping data for table `badges`
--

LOCK TABLES `badges` WRITE;
/*!40000 ALTER TABLE `badges` DISABLE KEYS */;
/*!40000 ALTER TABLE `badges` ENABLE KEYS */;
UNLOCK TABLES;

--
-- Table structure for table `error_log`
--

DROP TABLE IF EXISTS `error_log`;
/*!40101 SET @saved_cs_client     = @@character_set_client */;
/*!50503 SET character_set_client = utf8mb4 */;
CREATE TABLE `error_log` (
  `id` bigint unsigned NOT NULL AUTO_INCREMENT,
  `date` timestamp NOT NULL DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP,
  `url` text CHARACTER SET utf8mb4 COLLATE utf8mb4_0900_as_ci NOT NULL,
  `browser` text CHARACTER SET utf8mb4 COLLATE utf8mb4_0900_as_ci NOT NULL,
  `details` text CHARACTER SET utf8mb4 COLLATE utf8mb4_0900_as_ci NOT NULL,
  PRIMARY KEY (`id`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_0900_as_ci;
/*!40101 SET character_set_client = @saved_cs_client */;

--
-- Dumping data for table `error_log`
--

LOCK TABLES `error_log` WRITE;
/*!40000 ALTER TABLE `error_log` DISABLE KEYS */;
/*!40000 ALTER TABLE `error_log` ENABLE KEYS */;
UNLOCK TABLES;

--
-- Table structure for table `filters`
--

DROP TABLE IF EXISTS `filters`;
/*!40101 SET @saved_cs_client     = @@character_set_client */;
/*!50503 SET character_set_client = utf8mb4 */;
CREATE TABLE `filters` (
  `id` bigint NOT NULL,
  `user_id` bigint NOT NULL,
  `name` varchar(45) NOT NULL DEFAULT '',
  `selected` tinyint(1) NOT NULL DEFAULT '0',
  `starred` tinyint(1) DEFAULT NULL,
  `start_date` datetime DEFAULT NULL,
  `end_date` datetime DEFAULT NULL,
  `archived` tinyint(1) DEFAULT NULL,
  `participated` tinyint(1) DEFAULT NULL,
  `unread` tinyint(1) DEFAULT NULL,
  `item_id` bigint DEFAULT NULL,
  `group_id` int DEFAULT NULL,
  `older_than` int DEFAULT NULL,
  `newer_than` int DEFAULT NULL,
  `users_search` varchar(200) DEFAULT NULL,
  `body_search` varchar(100) DEFAULT NULL,
  `important` tinyint(1) DEFAULT NULL,
  PRIMARY KEY (`id`),
  KEY `user_idx` (`user_id`),
  CONSTRAINT `fk_filters_user_id_users_group_id` FOREIGN KEY (`user_id`) REFERENCES `users` (`group_id`) ON DELETE CASCADE
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_0900_ai_ci;
/*!40101 SET character_set_client = @saved_cs_client */;

--
-- Dumping data for table `filters`
--

LOCK TABLES `filters` WRITE;
/*!40000 ALTER TABLE `filters` DISABLE KEYS */;
/*!40000 ALTER TABLE `filters` ENABLE KEYS */;
UNLOCK TABLES;
/*!50003 SET @saved_cs_client      = @@character_set_client */ ;
/*!50003 SET @saved_cs_results     = @@character_set_results */ ;
/*!50003 SET @saved_col_connection = @@collation_connection */ ;
/*!50003 SET character_set_client  = utf8mb4 */ ;
/*!50003 SET character_set_results = utf8mb4 */ ;
/*!50003 SET collation_connection  = utf8mb4_general_ci */ ;
/*!50003 SET @saved_sql_mode       = @@sql_mode */ ;
/*!50003 SET sql_mode              = 'ONLY_FULL_GROUP_BY,STRICT_TRANS_TABLES,NO_ZERO_IN_DATE,NO_ZERO_DATE,ERROR_FOR_DIVISION_BY_ZERO,NO_ENGINE_SUBSTITUTION' */ ;
DELIMITER ;;
/*!50003 CREATE TRIGGER `before_insert_filters` BEFORE INSERT ON `filters` FOR EACH ROW BEGIN IF (NEW.id IS NULL OR NEW.id = 0) THEN SET NEW.id = FLOOR(RAND() * 1000000000) + FLOOR(RAND() * 1000000000) * 1000000000; END IF; END */;;
DELIMITER ;
/*!50003 SET sql_mode              = @saved_sql_mode */ ;
/*!50003 SET character_set_client  = @saved_cs_client */ ;
/*!50003 SET character_set_results = @saved_cs_results */ ;
/*!50003 SET collation_connection  = @saved_col_connection */ ;

--
-- Table structure for table `gorp_migrations`
--

DROP TABLE IF EXISTS `gorp_migrations`;
/*!40101 SET @saved_cs_client     = @@character_set_client */;
/*!50503 SET character_set_client = utf8mb4 */;
CREATE TABLE `gorp_migrations` (
  `id` varchar(255) NOT NULL,
  `applied_at` datetime DEFAULT NULL,
  PRIMARY KEY (`id`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_0900_ai_ci;
/*!40101 SET character_set_client = @saved_cs_client */;

--
-- Dumping data for table `gorp_migrations`
--

LOCK TABLES `gorp_migrations` WRITE;
/*!40000 ALTER TABLE `gorp_migrations` DISABLE KEYS */;
INSERT INTO `gorp_migrations` VALUES ('1900000000__fix_db_bugs.sql','2025-09-03 18:25:44'),('1903211315_create_task_children_data_view.sql','2025-09-03 18:25:44'),('1906262031_create_column_groups_attempts_iMinusScore.sql','2025-09-03 18:25:44'),('1906262032_add_index_GroupItemMinusScoreBestAnswerDateID_on_groups_attempts.sql','2025-09-03 18:25:45'),('1907012147_add_index_sType_on_groups.sql','2025-09-03 18:25:45'),('1907041726_add_index_sName_on_groups.sql','2025-09-03 18:25:45'),('1907091841_add_index_sType_sName_on_groups.sql','2025-09-03 18:25:45'),('1907110726_create_table_sessions.sql','2025-09-03 18:25:45'),('1907170545_create_table_login_states.sql','2025-09-03 18:25:45'),('1907250157_create_table_refresh_tokens.sql','2025-09-03 18:25:45'),('1907250330_add_index_ParentOrder_on_groups_groups.sql','2025-09-03 18:25:45'),('1907251939_alter_table_sessions_to_use_long_tokens.sql','2025-09-03 18:25:46'),('1907252019_add_index_idUser_on_sessions.sql','2025-09-03 18:25:46'),('1907310033_alter_table_groups_modify_sType_add_Base.sql','2025-09-03 18:25:46'),('1908020413_add_index_idGroup_on_history_groups_login_prefixes.sql','2025-09-03 18:25:46'),('1908020413_add_index_tempUser_on_users.sql','2025-09-03 18:25:46'),('1908020413_add_indices_idGroupParent_and_idGroupChild_on_history_groups_groups.sql','2025-09-03 18:25:46'),('1908090140_do_not_mark_ancestors_as_todo_in_groups_groups_triggers.sql','2025-09-03 18:25:46'),('1908150302_move_sAdditionalTime_into_groups_items.sql','2025-09-03 18:25:47'),('1908210505_add_unique_index_sPassword_on_groups.sql','2025-09-03 18:25:47'),('1908270419_add_joinedByCode_into_groups_groups_sType.sql','2025-09-03 18:25:47'),('1908270635_rename_groups_passwords_into_groups_codes.sql','2025-09-03 18:25:47'),('1909101811_add_comments_groups_table.sql','2025-09-03 18:25:47'),('1909101822_add_comments_groups_ancestors_table.sql','2025-09-03 18:25:47'),('1909101914_add_comments_groups_attempts_table.sql','2025-09-03 18:25:47'),('1909101934_add_comments_groups_groups_table.sql','2025-09-03 18:25:47'),('1909111146_add_comments_groups_items_table.sql','2025-09-03 18:25:48'),('1909111540_add_comments_propagate_tables.sql','2025-09-03 18:25:48'),('1909111544_add_comments_groups_login_prefixes_table.sql','2025-09-03 18:25:48'),('1909111635_add_comments_items_table.sql','2025-09-03 18:25:48'),('1909111652_add_comments_items_items_ancestors_tables.sql','2025-09-03 18:25:48'),('1909111704_add_comments_items_strings_table.sql','2025-09-03 18:25:48'),('1909111715_add_comments_platforms_table.sql','2025-09-03 18:25:48'),('1909111750_add_comments_users_table.sql','2025-09-03 18:25:48'),('1909111755_add_comments_users_answers_table.sql','2025-09-03 18:25:48'),('1909111810_add_comments_users_items_table.sql','2025-09-03 18:25:49'),('1909150229_add_default_value_for_items_items_iDifficulty.sql','2025-09-03 18:25:49'),('1909150229_add_default_value_for_users_tempUser.sql','2025-09-03 18:25:49'),('1909150229_add_joinedByCode_into_history_groups_groups_sType.sql','2025-09-03 18:25:49'),('1909150229_alter_users_answers_drop_iVersion.sql','2025-09-03 18:25:49'),('1909192009_convert_columns_to_snake_case.sql','2025-09-03 18:25:52'),('1909192011_add_Manual_into_history_items_validation_type.sql','2025-09-03 18:25:52'),('1909200312_replace_items_items_always_visible_and_access_restricted_with_partial_access_propagation.sql','2025-09-03 18:25:52'),('1909220416_add_index_grade_on_groups.sql','2025-09-03 18:25:52'),('1909220416_add_index_parent_type_on_groups_groups.sql','2025-09-03 18:25:52'),('1909240455_remove_groups_items_propagate_access.sql','2025-09-03 18:25:53'),('1909271601_rename_datetime_columns.sql','2025-09-03 18:25:54'),('1909300624_rename_column_code_timer_to_code_lifetime.sql','2025-09-03 18:25:54'),('1910010000_create_tables_groups_contest_items_and_contest_participations.sql','2025-09-03 18:25:54'),('1910010624_drop_column_groups_attempts_contest_started_at.sql','2025-09-03 18:25:54'),('1910010624_drop_column_users_items_contest_started_at.sql','2025-09-03 18:25:55'),('1910010624_rename_team_related_columns_in_items.sql','2025-09-03 18:25:55'),('1910030549_drop_column_groups_items_additional_time.sql','2025-09-03 18:25:55'),('1910032036_drop_stale_tables.sql','2025-09-03 18:25:57'),('1910071834_add_expires_at_into_groups_groups_and_groups_ancestors.sql','2025-09-03 18:25:57'),('1910081805_create_column_items_contest_participants_group_id.sql','2025-09-03 18:25:57'),('1910091719_remove_items_qualified_group_id_contest_opens_at_contest_closes_at.sql','2025-09-03 18:25:57'),('1910140742_merge_users_items_into_groups_attempts.sql','2025-09-03 18:25:57'),('1910152243_make_users_items_active_attempt_id_not_null.sql','2025-09-03 18:25:57'),('1910210559_drop_column_groups_attempts_minus_score.sql','2025-09-03 18:25:58'),('1910250534_get_rid_of_user_id_everywhere.sql','2025-09-03 18:25:59'),('1910280216_move_contests_participations_entered_at_into_groups_attempts.sql','2025-09-03 18:25:59'),('1911050522_rework_items_permissions.sql','2025-09-03 18:25:59'),('1911081410_clean_groups_attempts.sql','2025-09-03 18:25:59'),('1911151755_attempts_computed_validated.sql','2025-09-03 18:25:59'),('1911151834_clean_items.sql','2025-09-03 18:26:00'),('1911152057_drop_column_items_custom_chapter.sql','2025-09-03 18:26:00'),('1911220651_add_table_group_membership_changes.sql','2025-09-03 18:26:00'),('1911220653_add_table_group_pending_requests.sql','2025-09-03 18:26:00'),('1911270643_create_table_group_managers.sql','2025-09-03 18:26:00'),('1911300847_drop_column_users_owned_group_id.sql','2025-09-03 18:26:00'),('1912020442_drop_column_groups_groups_role.sql','2025-09-03 18:26:01'),('1912030546_add_approval_related_columns.sql','2025-09-03 18:26:01'),('1912051435_create_contest_participants_groups.sql','2025-09-03 18:26:01'),('1912052259_create_table_items_unlocking_rules.sql','2025-09-03 18:26:01'),('1912071014_create_columns_items_items_weight.sql','2025-09-03 18:26:01'),('1912160111_set_permissions_granted_origin.sql','2025-09-03 18:26:02'),('1912170617_make_users_answers_attempt_id_not_null.sql','2025-09-03 18:26:02'),('1912200912_change_groups_attempts_ancestors_computation_state.sql','2025-09-03 18:26:02'),('2001090215_rework_scores_in_groups_attempts.sql','2025-09-03 18:26:02'),('2001141753_rename_and_rework_users_answers.sql','2025-09-03 18:26:02'),('2001142018_rename_groups_attempts_to_attempts.sql','2025-09-03 18:26:03'),('2001151623_drop_table_users_items.sql','2025-09-03 18:26:03'),('2001161438_rework_languages.sql','2025-09-03 18:26:03'),('2001161839_drop_column_items_items_id.sql','2025-09-03 18:26:04'),('2001161935_drop_column_items_ancestors_id.sql','2025-09-03 18:26:04'),('2001210530_fix_attempt_dates.sql','2025-09-03 18:26:04'),('2001220301_rename_items_has_attempts_to_allows_multiple_attempts.sql','2025-09-03 18:26:04'),('2001270512_mark_attempts_as_changed_on_permissions_adding.sql','2025-09-03 18:26:04'),('2001281306_mark_attempts_as_changed_on_groups_relations_adding.sql','2025-09-03 18:26:04'),('2001300647_make_groups_created_at_not_null.sql','2025-09-03 18:26:05'),('2002030602_rework_groups_ancestors.sql','2025-09-03 18:26:05'),('2002030647_do_not_delete_groups_ancestors_or_items_ancestors_in_triggers.sql','2025-09-03 18:26:05'),('2002030841_rework_groups_groups.sql','2025-09-03 18:26:05'),('2002030904_add_more_foreign_keys.sql','2025-09-03 18:26:05'),('2002032357_rename_groups_type_UserSelf_to_User.sql','2025-09-03 18:26:06'),('2002040752_clean_platforms.sql','2025-09-03 18:26:06'),('2002042214_rename_groups_free_access_and_opened.sql','2025-09-03 18:26:06'),('2002051347_add_groups_type_session.sql','2025-09-03 18:26:07'),('2002121859_create_tables_for_batches_of_users.sql','2025-09-03 18:26:07'),('2002130543_rename_items_group_code_enter.sql','2025-09-03 18:26:07'),('2002230456_add_column_items_requires_explicit_entry.sql','2025-09-03 18:26:07'),('2002261444_rework_enter_permissions.sql','2025-09-03 18:26:08'),('2002262024_drop_groups_groups_child_order.sql','2025-09-03 18:26:08'),('2003031244_add_new_type_skill_into_items_type.sql','2025-09-03 18:26:08'),('2003110348_extract_results_from_attempts.sql','2025-09-03 18:26:09'),('2003231854_add_column_allows_submissions_until_into_attempts.sql','2025-09-03 18:26:09'),('2004042238_drop_columns_team_item_id_and_team_participating_from_groups.sql','2025-09-03 18:26:09'),('2004082015_change_comment_of_groups_ancestors.sql','2025-09-03 18:26:09'),('2004131131_frozen_teams.sql','2025-09-03 18:26:10'),('2004162312_drop_table_login_states.sql','2025-09-03 18:26:10'),('2004190227_add_column_ended_at_into_attempts_table.sql','2025-09-03 18:26:11'),('2004210235_rename_contest_related_columns.sql','2025-09-03 18:26:11'),('2005190324_rename_groups_activity_id_and_add_groups_root_skill_id.sql','2025-09-03 18:26:12'),('2005231922_add_virtual_column_group_managers_can_manage_value.sql','2025-09-03 18:26:12'),('2008040105_rework_root_group.sql','2025-09-03 18:26:12'),('2008272355_make_groups_code_binary.sql','2025-09-03 18:26:12'),('2009032258_public_activities.sql','2025-09-03 18:26:13'),('2009032342_drop_column_items_is_root.sql','2025-09-03 18:26:13'),('2009040309_add_max_participants_into_groups.sql','2025-09-03 18:26:13'),('2011041640_fix_bugs_of_the_initial_dump.sql','2025-09-03 18:26:14'),('2011072035_add_indexes_needed_by_activity_log_to_answers.sql','2025-09-03 18:26:14'),('2011081833_add_indexes_needed_by_activity_log_to_results.sql','2025-09-03 18:26:14'),('2011222206_mark_results_with_answers_and_hints_as_started.sql','2025-09-03 18:26:14'),('2101181622_mark_all_results_as_to_be_propagated.sql','2025-09-03 18:26:14'),('2101220131_drop_processing_state_from_results_result_propagation_state.sql','2025-09-03 18:26:15'),('2101221642_mark_results_as_to_be_recomputed_on_items_items_update.sql','2025-09-03 18:26:15'),('2102250414_remove_empty_value_from_items_full_screen.sql','2025-09-03 18:26:15'),('2103301810_add_foreign_key_from_items_propagate_id_to_items_id.sql','2025-09-03 18:26:15'),('2104212206_make_items_entry_frozen_teams_not_null.sql','2025-09-03 18:26:16'),('2104222142_drop_column_users_allow_subgroups.sql','2025-09-03 18:26:16'),('2105071401_change_items_items_child_item_id_fk_referential_action.sql','2025-09-03 18:26:16'),('2105071411_change_items_ancestors_child_item_id_fk_referential_action.sql','2025-09-03 18:26:16'),('2105071413_change_items_strings_item_id_fk_referential_action.sql','2025-09-03 18:26:17'),('2105071416_change_results_item_id_fk_referential_action.sql','2025-09-03 18:26:17'),('2105131302_create_table_results_propagate.sql','2025-09-03 18:26:17'),('2105131313_copy_result_propagation_states_into_results_propagate_table.sql','2025-09-03 18:26:18'),('2105131319_drop_column_results_result_propagation_state.sql','2025-09-03 18:26:18'),('2105261404_add_index_type_participant_id_item_id_created_at_desc_attempt_id_desc_for_answers.sql','2025-09-03 18:26:18'),('2105301052_optimize_after_insert_and_after_update_triggers_on_groups_groups.sql','2025-09-03 18:26:18'),('2106150657_add_more_indexes_on_results_to_speed_up_user_progress_calculation.sql','2025-09-03 18:26:19'),('2107162117_add_column_code_lifetime_seconds.sql','2025-09-03 18:26:19'),('2107162125_rename_column_code_lifetime_seconds_to_code_lifetime.sql','2025-09-03 18:26:19'),('2112041553_add_column_results_help_requested.sql','2025-09-03 18:26:20'),('2201092302_make_items_options_nullable.sql','2025-09-03 18:26:20'),('2202011518_make_items_url_longer.sql','2025-09-03 18:26:20'),('2207060651_make_groups_text_id_nullable.sql','2025-09-03 18:26:21'),('2207081134_make_groups_text_id_values_unique.sql','2025-09-03 18:26:21'),('2209060023_add_new_action_joined_by_badge_into_group_membership_changes.sql','2025-09-03 18:26:21'),('2209090721_create_column_users_latest_profile_sync_at.sql','2025-09-03 18:26:22'),('2212081100_add_column_items_layout.sql','2025-09-03 18:26:22'),('2301091000_item_type_course_becomes_task.sql','2025-09-03 18:26:22'),('2301171000_items_strings_image_url_length.sql','2025-09-03 18:26:23'),('2301310900_forum_permissions.sql','2025-09-03 18:26:23'),('2301311000_delete_old_forum_tables.sql','2025-09-03 18:26:23'),('2301311130_new_forum_threads.sql','2025-09-03 18:26:23'),('2302071000_items_text_id_unique.sql','2025-09-03 18:26:24'),('2303281630_fix_before_update_items_trigger.sql','2025-09-03 18:26:24'),('2312121508_add_index_answers_for_items_log_service.sql','2025-09-03 18:26:25'),('2402201446_new_sessions_schema.sql','2025-09-03 18:26:25'),('2403201408_add_new_action_removed_due_to_approval_change_into_group_membership_changes.sql','2025-09-03 18:26:25'),('2405061508_add_index_permissions_propagate.sql','2025-09-03 18:26:26'),('2405071119_results_propagate_mark_item_id.sql','2025-09-03 18:26:26'),('2405071129_results_propagate_insert_item_id_in_table_instead_of_trigger.sql','2025-09-03 18:26:26'),('2406141532_add_full_text_indexes_for_search.sql','2025-09-03 18:26:27'),('2407171717_create_table_user_batches_new.sql','2025-09-03 18:26:27'),('2407201347_fractional_time_in_group_membership_changes.sql','2025-09-03 18:26:27'),('2407201351_fractional_time_in_group_pending_requests.sql','2025-09-03 18:26:27'),('2407291420_remove_index_answers_for_items_log_service.sql','2025-09-03 18:26:27'),('2407292042_add_index_created_at_d_item_id_participant_id_attempt_id_d_type_d_id_autho_on_answers.sql','2025-09-03 18:26:28'),('2408072312_rework_trigger_before_insert_items.sql','2025-09-03 18:26:28'),('2408072318_rework_trigger_before_update_items.sql','2025-09-03 18:26:28'),('2408072320_recalculate_items_platform_id.sql','2025-09-03 18:26:28'),('2408150942_make_items_type_not_null.sql','2025-09-03 18:26:28'),('2409180913_add_enum_value_propagating_into_results_propagate_state.sql','2025-09-03 18:26:29'),('2409191745_add_column_results_propagate_items_is_processing.sql','2025-09-03 18:26:29'),('2409192252_rename_table_results_propagate_items_to_results_recompute_for_items.sql','2025-09-03 18:26:29'),('2409192304_rename_table_results_propagate_items_into_results_recompute_for_items_in_triggers.sql','2025-09-03 18:26:29'),('2409290723_add_column_results_recomputing_state.sql','2025-09-03 18:26:30'),('2409290724_create_trigger_before_update_results.sql','2025-09-03 18:26:30'),('2409300018_add_enum_value_recomputing_into_results_propagate_state.sql','2025-09-03 18:26:30'),('2410170736_modify_trigger_after_insert_groups_to_create_groups_ancestors.sql','2025-09-03 18:26:30'),('2410230158_add_column_groups_groups_is_team_membership.sql','2025-09-03 18:26:30'),('2410230202_initialize_column_groups_groups_is_team_membership.sql','2025-09-03 18:26:30'),('2410260221_modify_trigger_before_insert_groups_groups_to_set_is_team_membership.sql','2025-09-03 18:26:31'),('2410281951_modify_trigger_before_update_groups_groups_to_disallow_modifying_of_is_team_membership.sql','2025-09-03 18:26:31'),('2410282058_modify_trigger_after_insert_groups_groups_to_ignore_team_memberships.sql','2025-09-03 18:26:31'),('2410282314_modify_trigger_after_update_groups_groups_to_ignore_team_memberships.sql','2025-09-03 18:26:31'),('2410282327_modify_trigger_before_delete_groups_groups_to_ignore_team_memberships.sql','2025-09-03 18:26:31'),('2410282334_add_trigger_before_update_groups_to_disallow_modifying_type.sql','2025-09-03 18:26:31'),('2410292333_remove_unnecessary_enum_values_from_groups_propagate_ancestors_computation_state.sql','2025-09-03 18:26:32'),('2410292340_remove_unnecessary_enum_values_from_items_propagate_ancestors_computation_state.sql','2025-09-03 18:26:32'),('2411061620_create_index_child_group_id_is_team_membership_parent_group_id_expires_at_on_groups_groups.sql','2025-09-03 18:26:32'),('2411070615_add_column_is_team_membership_to_groups_groups_active_view.sql','2025-09-03 18:26:32'),('2411070848_drop_trigger_after_insert_items.sql','2025-09-03 18:26:32'),('2411181747_modify_trigger_after_insert_items_items_to_recompute_parent_results.sql','2025-09-03 18:26:33'),('2412070000_rework_triggers_to_allow_sync_permissions_and_results_propagations_when_needed.sql','2025-09-03 18:26:33'),('2501231952_add_column_connection_id_to_permissions_propagate_sync.sql','2025-09-03 18:26:33'),('2501232000_add_column_connection_id_to_results_propagate_sync.sql','2025-09-03 18:26:33'),('2501232006_rework_triggers_to_store_connection_id_for_sync_permissions_and_results_propagations.sql','2025-09-03 18:26:34'),('2501232236_create_view_results_propagate_sync_conn.sql','2025-09-03 18:26:34'),('2501232239_create_view_permissions_propagate_sync_conn.sql','2025-09-03 18:26:34'),('2503081934_set_charset_of_most_tables_to_utf8mb4.sql','2025-09-03 18:26:36'),('2503112139_set_charset_of_answers_table_to_utf8mb4.sql','2025-09-03 18:26:36'),('2503112146_set_charset_of_items_related_tables_to_utf8mb4.sql','2025-09-03 18:26:37'),('2503112150_set_charset_of_batches_related_tables_to_utf8mb4.sql','2025-09-03 18:26:37'),('2503112156_set_default_db_charset_to_utf8mb4.sql','2025-09-03 18:26:37'),('2503141810_rename_table_groups_contest_items_to_group_item_additional_times.sql','2025-09-03 18:26:37'),('2503241255_create_empty_stopwords_table.sql','2025-09-03 18:26:38'),('2503241300_recreate_fulltext_indexes_with_empty_stopwords_table.sql','2025-09-03 18:26:38'),('2506041902_modify_trigger_after_insert_groups_groups_to_ignore_child_group_with_id_4.sql','2025-09-03 18:26:38'),('2506041904_create_non_temp_users_base_group.sql','2025-09-03 18:26:38'),('2506041905_add_non_temp_users_group_into_all_users_group.sql','2025-09-03 18:26:39'),('2506042226_move_non_temp_users_from_all_users_group_into_non_temp_users_group.sql','2025-09-03 18:26:39'),('2507011113_fix_and_rework_marking_permissions_for_recomputing_in_before_delete_items_items_trigger.sql','2025-09-03 18:26:39'),('2507011120_rework_marking_permissions_for_recomputing_in_after_update_items_items_trigger.sql','2025-09-03 18:26:39'),('2507011128_rework_marking_permissions_for_recomputing_in_after_insert_items_items_trigger.sql','2025-09-03 18:26:39'),('2508180018_undo_ignoring_group_with_id_4_in_trigger_after_insert_groups_groups.sql','2025-09-03 18:26:39'),('2508181435_add_support_for_propagating_state_of_results_propagate_in_triggers.sql','2025-09-03 18:26:40'),('2508300002_remove_ignore_from_inserts_in_triggers_where_possible.sql','2025-09-03 18:26:40');
/*!40000 ALTER TABLE `gorp_migrations` ENABLE KEYS */;
UNLOCK TABLES;

--
-- Table structure for table `gradings`
--

DROP TABLE IF EXISTS `gradings`;
/*!40101 SET @saved_cs_client     = @@character_set_client */;
/*!50503 SET character_set_client = utf8mb4 */;
CREATE TABLE `gradings` (
  `answer_id` bigint NOT NULL,
  `score` float NOT NULL COMMENT 'Score obtained',
  `graded_at` datetime NOT NULL COMMENT 'When was it last graded',
  PRIMARY KEY (`answer_id`),
  CONSTRAINT `fk_submissions_answer_id_answers_id` FOREIGN KEY (`answer_id`) REFERENCES `answers` (`id`) ON DELETE CASCADE
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_0900_ai_ci COMMENT='Grading results for answers';
/*!40101 SET character_set_client = @saved_cs_client */;

--
-- Dumping data for table `gradings`
--

LOCK TABLES `gradings` WRITE;
/*!40000 ALTER TABLE `gradings` DISABLE KEYS */;
/*!40000 ALTER TABLE `gradings` ENABLE KEYS */;
UNLOCK TABLES;

--
-- Table structure for table `group_item_additional_times`
--

DROP TABLE IF EXISTS `group_item_additional_times`;
/*!40101 SET @saved_cs_client     = @@character_set_client */;
/*!50503 SET character_set_client = utf8mb4 */;
CREATE TABLE `group_item_additional_times` (
  `group_id` bigint NOT NULL,
  `item_id` bigint NOT NULL,
  `additional_time` time NOT NULL DEFAULT '00:00:00' COMMENT 'Time that was attributed (can be negative) to this group for this time-limited item',
  PRIMARY KEY (`group_id`,`item_id`),
  KEY `fk_group_item_additional_times_item_id_items_id` (`item_id`),
  CONSTRAINT `fk_group_item_additional_times_group_id_groups_id` FOREIGN KEY (`group_id`) REFERENCES `groups` (`id`) ON DELETE CASCADE,
  CONSTRAINT `fk_group_item_additional_times_item_id_items_id` FOREIGN KEY (`item_id`) REFERENCES `items` (`id`) ON DELETE CASCADE
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_0900_ai_ci COMMENT='Additional times of groups on time-limited items';
/*!40101 SET character_set_client = @saved_cs_client */;

--
-- Dumping data for table `group_item_additional_times`
--

LOCK TABLES `group_item_additional_times` WRITE;
/*!40000 ALTER TABLE `group_item_additional_times` DISABLE KEYS */;
/*!40000 ALTER TABLE `group_item_additional_times` ENABLE KEYS */;
UNLOCK TABLES;

--
-- Table structure for table `group_managers`
--

DROP TABLE IF EXISTS `group_managers`;
/*!40101 SET @saved_cs_client     = @@character_set_client */;
/*!50503 SET character_set_client = utf8mb4 */;
CREATE TABLE `group_managers` (
  `group_id` bigint NOT NULL,
  `manager_id` bigint NOT NULL,
  `can_manage` enum('none','memberships','memberships_and_group') NOT NULL DEFAULT 'none',
  `can_grant_group_access` tinyint(1) NOT NULL DEFAULT '0' COMMENT 'Can give members access rights to some items (requires the giver to be allowed to give this permission on the item)',
  `can_watch_members` tinyint(1) NOT NULL DEFAULT '0' COMMENT 'Can watch members’ submissions on items. Requires the watcher to be allowed to watch this item. For members who have agreed',
  `can_edit_personal_info` tinyint(1) NOT NULL DEFAULT '0' COMMENT 'Can change member’s personal info, for those who have agreed (not visible to managers, only for specific uses)',
  `can_manage_value` tinyint unsigned GENERATED ALWAYS AS ((`can_manage` + 0)) VIRTUAL NOT NULL COMMENT 'can_manage as an integer (to use comparison operators)',
  PRIMARY KEY (`group_id`,`manager_id`),
  KEY `fk_group_managers_manager_id_groups_id` (`manager_id`),
  CONSTRAINT `fk_group_managers_group_id_groups_id` FOREIGN KEY (`group_id`) REFERENCES `groups` (`id`) ON DELETE CASCADE,
  CONSTRAINT `fk_group_managers_manager_id_groups_id` FOREIGN KEY (`manager_id`) REFERENCES `groups` (`id`) ON DELETE CASCADE
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_0900_ai_ci COMMENT='Group managers and their permissions';
/*!40101 SET character_set_client = @saved_cs_client */;

--
-- Dumping data for table `group_managers`
--

LOCK TABLES `group_managers` WRITE;
/*!40000 ALTER TABLE `group_managers` DISABLE KEYS */;
/*!40000 ALTER TABLE `group_managers` ENABLE KEYS */;
UNLOCK TABLES;

--
-- Table structure for table `group_membership_changes`
--

DROP TABLE IF EXISTS `group_membership_changes`;
/*!40101 SET @saved_cs_client     = @@character_set_client */;
/*!50503 SET character_set_client = utf8mb4 */;
CREATE TABLE `group_membership_changes` (
  `group_id` bigint NOT NULL,
  `member_id` bigint NOT NULL,
  `at` datetime(3) NOT NULL DEFAULT CURRENT_TIMESTAMP(3) COMMENT 'Time of the action',
  `action` enum('invitation_created','invitation_withdrawn','invitation_refused','invitation_accepted','join_request_created','join_request_withdrawn','join_request_refused','join_request_accepted','leave_request_created','leave_request_withdrawn','leave_request_refused','leave_request_accepted','left','removed','joined_by_code','added_directly','expired','joined_by_badge','removed_due_to_approval_change') DEFAULT NULL,
  `initiator_id` bigint DEFAULT NULL COMMENT 'The user who initiated the action (if any), typically the group owner/manager or the member himself',
  PRIMARY KEY (`group_id`,`member_id`,`at`),
  KEY `group_id_member_id_at_desc` (`group_id`,`member_id`,`at` DESC),
  KEY `group_id_at_desc_member_id` (`group_id`,`at` DESC,`member_id`),
  KEY `member_id_at_desc_group_id` (`member_id`,`at` DESC,`group_id`),
  KEY `group_id_at_member_id` (`group_id`,`at`,`member_id`),
  KEY `member_id_at_group_id` (`member_id`,`at`,`group_id`),
  KEY `fk_group_membership_changes_initiator_id_users_group_id` (`initiator_id`),
  CONSTRAINT `fk_group_membership_changes_group_id_groups_id` FOREIGN KEY (`group_id`) REFERENCES `groups` (`id`) ON DELETE CASCADE,
  CONSTRAINT `fk_group_membership_changes_initiator_id_users_group_id` FOREIGN KEY (`initiator_id`) REFERENCES `users` (`group_id`) ON DELETE SET NULL,
  CONSTRAINT `fk_group_membership_changes_member_id_groups_id` FOREIGN KEY (`member_id`) REFERENCES `groups` (`id`) ON DELETE CASCADE
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_0900_ai_ci COMMENT='Stores the history of group membership changes';
/*!40101 SET character_set_client = @saved_cs_client */;

--
-- Dumping data for table `group_membership_changes`
--

LOCK TABLES `group_membership_changes` WRITE;
/*!40000 ALTER TABLE `group_membership_changes` DISABLE KEYS */;
/*!40000 ALTER TABLE `group_membership_changes` ENABLE KEYS */;
UNLOCK TABLES;

--
-- Table structure for table `group_pending_requests`
--

DROP TABLE IF EXISTS `group_pending_requests`;
/*!40101 SET @saved_cs_client     = @@character_set_client */;
/*!50503 SET character_set_client = utf8mb4 */;
CREATE TABLE `group_pending_requests` (
  `group_id` bigint NOT NULL,
  `member_id` bigint NOT NULL,
  `type` enum('invitation','join_request','leave_request') DEFAULT NULL,
  `at` datetime(3) NOT NULL DEFAULT CURRENT_TIMESTAMP(3),
  `personal_info_view_approved` tinyint(1) NOT NULL DEFAULT '0' COMMENT 'for join requests',
  `lock_membership_approved` tinyint(1) NOT NULL DEFAULT '0' COMMENT 'for join requests',
  `watch_approved` tinyint(1) NOT NULL DEFAULT '0' COMMENT 'for join requests',
  PRIMARY KEY (`group_id`,`member_id`),
  KEY `group_id_member_id_at_desc` (`group_id`,`member_id`,`at` DESC),
  KEY `fk_group_pending_requests_member_id_groups_id` (`member_id`),
  CONSTRAINT `fk_group_pending_requests_group_id_groups_id` FOREIGN KEY (`group_id`) REFERENCES `groups` (`id`) ON DELETE CASCADE,
  CONSTRAINT `fk_group_pending_requests_member_id_groups_id` FOREIGN KEY (`member_id`) REFERENCES `groups` (`id`) ON DELETE CASCADE
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_0900_ai_ci COMMENT='Requests that require an action from a user (group owner/manager or member)';
/*!40101 SET character_set_client = @saved_cs_client */;

--
-- Dumping data for table `group_pending_requests`
--

LOCK TABLES `group_pending_requests` WRITE;
/*!40000 ALTER TABLE `group_pending_requests` DISABLE KEYS */;
/*!40000 ALTER TABLE `group_pending_requests` ENABLE KEYS */;
UNLOCK TABLES;

--
-- Table structure for table `groups`
--

DROP TABLE IF EXISTS `groups`;
/*!40101 SET @saved_cs_client     = @@character_set_client */;
/*!50503 SET character_set_client = utf8mb4 */;
CREATE TABLE `groups` (
  `id` bigint NOT NULL,
  `name` varchar(200) NOT NULL DEFAULT '',
  `type` enum('Class','Team','Club','Friends','Other','User','Session','Base','ContestParticipants') NOT NULL,
  `text_id` varchar(255) DEFAULT NULL COMMENT 'Internal text id for special groups. Used to refer o them and avoid breaking features if an admin renames the group',
  `grade` int NOT NULL DEFAULT '-2' COMMENT 'For some types of groups, indicate which grade the users belong to.',
  `grade_details` varchar(50) DEFAULT NULL COMMENT 'Explanations about the grade',
  `description` text COMMENT 'Purpose of this group. Will be visible by its members. or by the public if the group is public.',
  `created_at` datetime NOT NULL DEFAULT CURRENT_TIMESTAMP,
  `is_open` tinyint(1) NOT NULL DEFAULT '0' COMMENT 'Whether it appears to users as open to new members, i.e. the users can join using the code or create a join request',
  `is_public` tinyint(1) NOT NULL DEFAULT '0' COMMENT 'Whether it is visible to all users (through search) and open to join requests',
  `code` varbinary(50) DEFAULT NULL COMMENT 'Code that can be used to join the group (if it is opened)',
  `code_lifetime` int DEFAULT NULL COMMENT 'How long after the first use of the code it will expire (in seconds), NULL means infinity',
  `code_expires_at` datetime DEFAULT NULL COMMENT 'When the code expires. Set when it is first used.',
  `root_activity_id` bigint DEFAULT NULL COMMENT 'Root activity (chapter, task, or course) associated with this group',
  `root_skill_id` bigint DEFAULT NULL COMMENT 'Root skill associated with this group',
  `open_activity_when_joining` tinyint(1) NOT NULL DEFAULT '0' COMMENT 'Whether the activity should be started for participants as soon as they join the group',
  `send_emails` tinyint(1) NOT NULL DEFAULT '0',
  `is_official_session` tinyint(1) NOT NULL DEFAULT '0' COMMENT 'Whether this session is shown on the activity page (require specific permissions)',
  `require_personal_info_access_approval` enum('none','view','edit') NOT NULL DEFAULT 'none' COMMENT 'If not ''none'', requires (for joining) members to approve that managers may be able to view or edit their personal information',
  `require_lock_membership_approval_until` datetime DEFAULT NULL COMMENT 'If not null and in the future, requires (for joining) members to approve that they will not be able to leave the group without approval until the given date',
  `require_watch_approval` tinyint(1) NOT NULL DEFAULT '0' COMMENT 'Whether it requires (for joining) members to approve that managers may be able to watch their results and answers',
  `require_members_to_join_parent` tinyint(1) NOT NULL DEFAULT '0' COMMENT 'For sessions, whether the user joining this group should join the parent group as well',
  `frozen_membership` tinyint(1) DEFAULT '0' COMMENT 'Whether members can be added/removed to the group (intended for teams)',
  `max_participants` int unsigned DEFAULT NULL COMMENT 'The maximum number of participants (users and teams) in this group (strict limit if `enforce_max_participants`)',
  `enforce_max_participants` tinyint(1) DEFAULT '0' COMMENT 'Whether the number of participants is a strict constraint',
  `organizer` varchar(255) DEFAULT NULL COMMENT 'For sessions, a teacher/animator in charge of the organization',
  `address_line1` varchar(255) DEFAULT NULL COMMENT 'For sessions or schools',
  `address_line2` varchar(255) DEFAULT NULL COMMENT 'For sessions or schools',
  `address_postcode` varchar(25) DEFAULT NULL COMMENT 'For sessions or schools',
  `address_city` varchar(255) DEFAULT NULL COMMENT 'For sessions or schools',
  `address_country` varchar(255) DEFAULT NULL COMMENT 'For sessions or schools',
  `expected_start` datetime DEFAULT NULL COMMENT 'For sessions, time at which the session is expected to start',
  PRIMARY KEY (`id`),
  UNIQUE KEY `password` (`code`),
  UNIQUE KEY `text_id` (`text_id`),
  KEY `type` (`type`),
  KEY `name` (`name`),
  KEY `type_name` (`type`,`name`),
  KEY `grade` (`grade`),
  KEY `fk_groups_root_activity_id_items_id` (`root_activity_id`),
  KEY `fk_groups_root_skill_id_items_id` (`root_skill_id`),
  FULLTEXT KEY `fullTextName` (`name`),
  CONSTRAINT `fk_groups_root_activity_id_items_id` FOREIGN KEY (`root_activity_id`) REFERENCES `items` (`id`) ON DELETE SET NULL,
  CONSTRAINT `fk_groups_root_skill_id_items_id` FOREIGN KEY (`root_skill_id`) REFERENCES `items` (`id`) ON DELETE SET NULL,
  CONSTRAINT `cs_can_enforce_max_participants` CHECK (((0 = `enforce_max_participants`) or (`max_participants` is not null)))
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_0900_ai_ci COMMENT='A group can be either a user, a set of users, or a set of groups.';
/*!40101 SET character_set_client = @saved_cs_client */;

--
-- Dumping data for table `groups`
--

LOCK TABLES `groups` WRITE;
/*!40000 ALTER TABLE `groups` DISABLE KEYS */;
INSERT INTO `groups` VALUES (4,'NonTempUsers','Base','NonTempUsers',-2,NULL,'non-temporary users','2025-09-03 18:26:38',0,0,NULL,NULL,NULL,NULL,NULL,0,0,0,'none',NULL,0,0,0,NULL,0,NULL,NULL,NULL,NULL,NULL,NULL,NULL),(777,'Former task owners','Other',NULL,-2,NULL,'Contains all the task owners from AlgoreaPlatform and has can_edit=children permission on former custom chapters','2025-09-03 18:25:59',0,0,NULL,NULL,NULL,NULL,NULL,0,0,0,'none',NULL,0,0,0,NULL,0,NULL,NULL,NULL,NULL,NULL,NULL,NULL);
/*!40000 ALTER TABLE `groups` ENABLE KEYS */;
UNLOCK TABLES;
/*!50003 SET @saved_cs_client      = @@character_set_client */ ;
/*!50003 SET @saved_cs_results     = @@character_set_results */ ;
/*!50003 SET @saved_col_connection = @@collation_connection */ ;
/*!50003 SET character_set_client  = utf8mb4 */ ;
/*!50003 SET character_set_results = utf8mb4 */ ;
/*!50003 SET collation_connection  = utf8mb4_general_ci */ ;
/*!50003 SET @saved_sql_mode       = @@sql_mode */ ;
/*!50003 SET sql_mode              = 'ONLY_FULL_GROUP_BY,STRICT_TRANS_TABLES,NO_ZERO_IN_DATE,NO_ZERO_DATE,ERROR_FOR_DIVISION_BY_ZERO,NO_ENGINE_SUBSTITUTION' */ ;
DELIMITER ;;
/*!50003 CREATE TRIGGER `before_insert_groups` BEFORE INSERT ON `groups` FOR EACH ROW BEGIN IF (NEW.id IS NULL OR NEW.id = 0) THEN SET NEW.id = FLOOR(RAND() * 1000000000) + FLOOR(RAND() * 1000000000) * 1000000000; END IF; END */;;
DELIMITER ;
/*!50003 SET sql_mode              = @saved_sql_mode */ ;
/*!50003 SET character_set_client  = @saved_cs_client */ ;
/*!50003 SET character_set_results = @saved_cs_results */ ;
/*!50003 SET collation_connection  = @saved_col_connection */ ;
/*!50003 SET @saved_cs_client      = @@character_set_client */ ;
/*!50003 SET @saved_cs_results     = @@character_set_results */ ;
/*!50003 SET @saved_col_connection = @@collation_connection */ ;
/*!50003 SET character_set_client  = utf8mb4 */ ;
/*!50003 SET character_set_results = utf8mb4 */ ;
/*!50003 SET collation_connection  = utf8mb4_general_ci */ ;
/*!50003 SET @saved_sql_mode       = @@sql_mode */ ;
/*!50003 SET sql_mode              = 'ONLY_FULL_GROUP_BY,STRICT_TRANS_TABLES,NO_ZERO_IN_DATE,NO_ZERO_DATE,ERROR_FOR_DIVISION_BY_ZERO,NO_ENGINE_SUBSTITUTION' */ ;
DELIMITER ;;
/*!50003 CREATE TRIGGER `after_insert_groups` AFTER INSERT ON `groups` FOR EACH ROW BEGIN
  INSERT INTO `groups_ancestors` (`ancestor_group_id`, `child_group_id`) VALUES (NEW.`id`, NEW.`id`);
END */;;
DELIMITER ;
/*!50003 SET sql_mode              = @saved_sql_mode */ ;
/*!50003 SET character_set_client  = @saved_cs_client */ ;
/*!50003 SET character_set_results = @saved_cs_results */ ;
/*!50003 SET collation_connection  = @saved_col_connection */ ;
/*!50003 SET @saved_cs_client      = @@character_set_client */ ;
/*!50003 SET @saved_cs_results     = @@character_set_results */ ;
/*!50003 SET @saved_col_connection = @@collation_connection */ ;
/*!50003 SET character_set_client  = utf8mb4 */ ;
/*!50003 SET character_set_results = utf8mb4 */ ;
/*!50003 SET collation_connection  = utf8mb4_general_ci */ ;
/*!50003 SET @saved_sql_mode       = @@sql_mode */ ;
/*!50003 SET sql_mode              = 'ONLY_FULL_GROUP_BY,STRICT_TRANS_TABLES,NO_ZERO_IN_DATE,NO_ZERO_DATE,ERROR_FOR_DIVISION_BY_ZERO,NO_ENGINE_SUBSTITUTION' */ ;
DELIMITER ;;
/*!50003 CREATE TRIGGER `before_update_groups` BEFORE UPDATE ON `groups` FOR EACH ROW BEGIN
  IF OLD.`type` != NEW.`type` AND (OLD.`type` IN ('User', 'Team') OR NEW.`type` IN ('User', 'Team')) THEN
    SIGNAL SQLSTATE '45000' SET MESSAGE_TEXT = 'Unable to change groups.type from/to User/Team';
  END IF;
END */;;
DELIMITER ;
/*!50003 SET sql_mode              = @saved_sql_mode */ ;
/*!50003 SET character_set_client  = @saved_cs_client */ ;
/*!50003 SET character_set_results = @saved_cs_results */ ;
/*!50003 SET collation_connection  = @saved_col_connection */ ;

--
-- Table structure for table `groups_ancestors`
--

DROP TABLE IF EXISTS `groups_ancestors`;
/*!40101 SET @saved_cs_client     = @@character_set_client */;
/*!50503 SET character_set_client = utf8mb4 */;
CREATE TABLE `groups_ancestors` (
  `ancestor_group_id` bigint NOT NULL,
  `child_group_id` bigint NOT NULL,
  `is_self` tinyint(1) GENERATED ALWAYS AS ((`ancestor_group_id` = `child_group_id`)) VIRTUAL COMMENT 'Whether ancestor_group_id = child_group_id (auto-generated)',
  `expires_at` datetime NOT NULL DEFAULT '9999-12-31 23:59:59' COMMENT 'The group relation expires at the specified time',
  PRIMARY KEY (`ancestor_group_id`,`child_group_id`),
  KEY `descendant` (`child_group_id`),
  CONSTRAINT `fk_groups_ancestors_ancestor_group_id_groups_id` FOREIGN KEY (`ancestor_group_id`) REFERENCES `groups` (`id`) ON DELETE CASCADE,
  CONSTRAINT `fk_groups_ancestors_child_group_id_groups_id` FOREIGN KEY (`child_group_id`) REFERENCES `groups` (`id`) ON DELETE CASCADE
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_0900_ai_ci COMMENT='All ancestor relationships for groups, given that a group is its own ancestor and team ancestors are not propagated to their members. It is a cache table that can be recomputed based on the content of groups_groups.';
/*!40101 SET character_set_client = @saved_cs_client */;

--
-- Dumping data for table `groups_ancestors`
--

LOCK TABLES `groups_ancestors` WRITE;
/*!40000 ALTER TABLE `groups_ancestors` DISABLE KEYS */;
INSERT INTO `groups_ancestors` (`ancestor_group_id`, `child_group_id`, `expires_at`) VALUES (4,4,'9999-12-31 23:59:59');
/*!40000 ALTER TABLE `groups_ancestors` ENABLE KEYS */;
UNLOCK TABLES;

--
-- Temporary view structure for view `groups_ancestors_active`
--

DROP TABLE IF EXISTS `groups_ancestors_active`;
/*!50001 DROP VIEW IF EXISTS `groups_ancestors_active`*/;
SET @saved_cs_client     = @@character_set_client;
/*!50503 SET character_set_client = utf8mb4 */;
/*!50001 CREATE VIEW `groups_ancestors_active` AS SELECT
 1 AS `ancestor_group_id`,
 1 AS `child_group_id`,
 1 AS `is_self`,
 1 AS `expires_at`*/;
SET character_set_client = @saved_cs_client;

--
-- Table structure for table `groups_groups`
--

DROP TABLE IF EXISTS `groups_groups`;
/*!40101 SET @saved_cs_client     = @@character_set_client */;
/*!50503 SET character_set_client = utf8mb4 */;
CREATE TABLE `groups_groups` (
  `parent_group_id` bigint NOT NULL,
  `child_group_id` bigint NOT NULL,
  `expires_at` datetime NOT NULL DEFAULT '9999-12-31 23:59:59' COMMENT 'The group membership expires at the specified time',
  `is_team_membership` tinyint(1) NOT NULL DEFAULT '0' COMMENT 'true if the parent group is a team',
  `personal_info_view_approved_at` datetime DEFAULT NULL,
  `personal_info_view_approved` tinyint(1) GENERATED ALWAYS AS ((`personal_info_view_approved_at` is not null)) VIRTUAL NOT NULL COMMENT 'personal_info_view_approved_at as boolean',
  `lock_membership_approved_at` datetime DEFAULT NULL,
  `lock_membership_approved` tinyint(1) GENERATED ALWAYS AS ((`lock_membership_approved_at` is not null)) VIRTUAL NOT NULL COMMENT 'lock_membership_approved_at as boolean',
  `watch_approved_at` datetime DEFAULT NULL,
  `watch_approved` tinyint(1) GENERATED ALWAYS AS ((`watch_approved_at` is not null)) VIRTUAL NOT NULL COMMENT 'watch_approved_at as boolean',
  PRIMARY KEY (`parent_group_id`,`child_group_id`),
  KEY `child_group_id_is_team_membership_parent_group_id_expires_at` (`child_group_id`,`is_team_membership`,`parent_group_id`,`expires_at`),
  CONSTRAINT `fk_groups_groups_child_group_id_groups_id` FOREIGN KEY (`child_group_id`) REFERENCES `groups` (`id`) ON DELETE CASCADE,
  CONSTRAINT `fk_groups_groups_parent_group_id_groups_id` FOREIGN KEY (`parent_group_id`) REFERENCES `groups` (`id`) ON DELETE CASCADE
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_0900_ai_ci COMMENT='Parent-child (N-N) relationships between groups (acyclic graph).';
/*!40101 SET character_set_client = @saved_cs_client */;

--
-- Dumping data for table `groups_groups`
--

LOCK TABLES `groups_groups` WRITE;
/*!40000 ALTER TABLE `groups_groups` DISABLE KEYS */;
/*!40000 ALTER TABLE `groups_groups` ENABLE KEYS */;
UNLOCK TABLES;
/*!50003 SET @saved_cs_client      = @@character_set_client */ ;
/*!50003 SET @saved_cs_results     = @@character_set_results */ ;
/*!50003 SET @saved_col_connection = @@collation_connection */ ;
/*!50003 SET character_set_client  = utf8mb4 */ ;
/*!50003 SET character_set_results = utf8mb4 */ ;
/*!50003 SET collation_connection  = utf8mb4_general_ci */ ;
/*!50003 SET @saved_sql_mode       = @@sql_mode */ ;
/*!50003 SET sql_mode              = 'ONLY_FULL_GROUP_BY,STRICT_TRANS_TABLES,NO_ZERO_IN_DATE,NO_ZERO_DATE,ERROR_FOR_DIVISION_BY_ZERO,NO_ENGINE_SUBSTITUTION' */ ;
DELIMITER ;;
/*!50003 CREATE TRIGGER `before_insert_groups_groups` BEFORE INSERT ON `groups_groups` FOR EACH ROW BEGIN
  SET NEW.is_team_membership = (SELECT type = 'Team' FROM `groups` WHERE id = NEW.parent_group_id FOR SHARE);
  IF NOT NEW.is_team_membership THEN
    INSERT INTO `groups_propagate` (id, ancestors_computation_state) VALUES (NEW.child_group_id, 'todo') ON DUPLICATE KEY UPDATE `ancestors_computation_state` = 'todo';
  END IF;
END */;;
DELIMITER ;
/*!50003 SET sql_mode              = @saved_sql_mode */ ;
/*!50003 SET character_set_client  = @saved_cs_client */ ;
/*!50003 SET character_set_results = @saved_cs_results */ ;
/*!50003 SET collation_connection  = @saved_col_connection */ ;
/*!50003 SET @saved_cs_client      = @@character_set_client */ ;
/*!50003 SET @saved_cs_results     = @@character_set_results */ ;
/*!50003 SET @saved_col_connection = @@collation_connection */ ;
/*!50003 SET character_set_client  = utf8mb4 */ ;
/*!50003 SET character_set_results = utf8mb4 */ ;
/*!50003 SET collation_connection  = utf8mb4_general_ci */ ;
/*!50003 SET @saved_sql_mode       = @@sql_mode */ ;
/*!50003 SET sql_mode              = 'ONLY_FULL_GROUP_BY,STRICT_TRANS_TABLES,NO_ZERO_IN_DATE,NO_ZERO_DATE,ERROR_FOR_DIVISION_BY_ZERO,NO_ENGINE_SUBSTITUTION' */ ;
DELIMITER ;;
/*!50003 CREATE TRIGGER `after_insert_groups_groups` AFTER INSERT ON `groups_groups` FOR EACH ROW BEGIN
  IF NEW.`expires_at` > NOW() AND NOT NEW.`is_team_membership` THEN
    INSERT INTO `results_propagate`
    SELECT `participant_id`, `attempt_id`, `results`.`item_id`, 'to_be_propagated' AS `state`
    FROM (
           SELECT `item_id`
           FROM (
                  SELECT DISTINCT `item_id`
                  FROM `results`
                         JOIN `groups_ancestors_active`
                              ON `groups_ancestors_active`.`child_group_id` = `results`.`participant_id` AND
                                 `groups_ancestors_active`.`ancestor_group_id` = NEW.`child_group_id`
                    FOR SHARE
                ) AS `result_items`
           WHERE EXISTS(
             SELECT 1
             FROM `permissions_generated`
                    JOIN `groups_ancestors_active` AS `grand_ancestors`
                         ON `grand_ancestors`.`child_group_id` = NEW.`parent_group_id` AND
                            `grand_ancestors`.`ancestor_group_id` = `permissions_generated`.`group_id`
                    JOIN `items_ancestors`
                         ON `items_ancestors`.`ancestor_item_id` = `permissions_generated`.`item_id`
             WHERE `items_ancestors`.`child_item_id` = `result_items`.`item_id`
               AND `permissions_generated`.`can_view_generated` != 'none'
               FOR SHARE
           )
             AND NOT EXISTS(
             SELECT 1
             FROM `permissions_generated`
                    JOIN `groups_ancestors_active` AS `child_ancestors`
                         ON `child_ancestors`.`child_group_id` = NEW.`child_group_id` AND
                            `child_ancestors`.`ancestor_group_id` = `permissions_generated`.`group_id`
                    JOIN `items_ancestors`
                         ON `items_ancestors`.`ancestor_item_id` = `permissions_generated`.`item_id`
             WHERE `items_ancestors`.`child_item_id` = `result_items`.`item_id`
               AND `permissions_generated`.`can_view_generated` != 'none'
               FOR SHARE
           )
             FOR SHARE
         ) AS `result_items_filtered`
           JOIN `results` ON `results`.`item_id` = `result_items_filtered`.`item_id`
           JOIN `groups_ancestors_active`
                ON `groups_ancestors_active`.`child_group_id` = `results`.`participant_id` AND
                   `groups_ancestors_active`.`ancestor_group_id` = NEW.`child_group_id`
      FOR SHARE
      ON DUPLICATE KEY UPDATE state = IF(state='propagating', 'to_be_propagated', state);
  END IF;
END */;;
DELIMITER ;
/*!50003 SET sql_mode              = @saved_sql_mode */ ;
/*!50003 SET character_set_client  = @saved_cs_client */ ;
/*!50003 SET character_set_results = @saved_cs_results */ ;
/*!50003 SET collation_connection  = @saved_col_connection */ ;
/*!50003 SET @saved_cs_client      = @@character_set_client */ ;
/*!50003 SET @saved_cs_results     = @@character_set_results */ ;
/*!50003 SET @saved_col_connection = @@collation_connection */ ;
/*!50003 SET character_set_client  = utf8mb4 */ ;
/*!50003 SET character_set_results = utf8mb4 */ ;
/*!50003 SET collation_connection  = utf8mb4_general_ci */ ;
/*!50003 SET @saved_sql_mode       = @@sql_mode */ ;
/*!50003 SET sql_mode              = 'ONLY_FULL_GROUP_BY,STRICT_TRANS_TABLES,NO_ZERO_IN_DATE,NO_ZERO_DATE,ERROR_FOR_DIVISION_BY_ZERO,NO_ENGINE_SUBSTITUTION' */ ;
DELIMITER ;;
/*!50003 CREATE TRIGGER `before_update_groups_groups` BEFORE UPDATE ON `groups_groups` FOR EACH ROW BEGIN
    IF OLD.`parent_group_id` != NEW.`parent_group_id` OR OLD.`child_group_id` != NEW.`child_group_id` OR OLD.`is_team_membership` != NEW.`is_team_membership` THEN
        SIGNAL SQLSTATE '45000' SET MESSAGE_TEXT = 'Unable to change immutable columns of groups_groups (parent_group_id/child_group_id/is_team_membership)';
    END IF;
END */;;
DELIMITER ;
/*!50003 SET sql_mode              = @saved_sql_mode */ ;
/*!50003 SET character_set_client  = @saved_cs_client */ ;
/*!50003 SET character_set_results = @saved_cs_results */ ;
/*!50003 SET collation_connection  = @saved_col_connection */ ;
/*!50003 SET @saved_cs_client      = @@character_set_client */ ;
/*!50003 SET @saved_cs_results     = @@character_set_results */ ;
/*!50003 SET @saved_col_connection = @@collation_connection */ ;
/*!50003 SET character_set_client  = utf8mb4 */ ;
/*!50003 SET character_set_results = utf8mb4 */ ;
/*!50003 SET collation_connection  = utf8mb4_general_ci */ ;
/*!50003 SET @saved_sql_mode       = @@sql_mode */ ;
/*!50003 SET sql_mode              = 'ONLY_FULL_GROUP_BY,STRICT_TRANS_TABLES,NO_ZERO_IN_DATE,NO_ZERO_DATE,ERROR_FOR_DIVISION_BY_ZERO,NO_ENGINE_SUBSTITUTION' */ ;
DELIMITER ;;
/*!50003 CREATE TRIGGER `after_update_groups_groups` AFTER UPDATE ON `groups_groups` FOR EACH ROW BEGIN
    IF OLD.expires_at != NEW.expires_at AND NOT NEW.`is_team_membership` THEN
        IF NEW.`expires_at` > NOW() THEN
            INSERT INTO `results_propagate`
            SELECT `participant_id`, `attempt_id`, `results`.`item_id`, 'to_be_propagated' AS `state`
            FROM (
                     SELECT `item_id`
                     FROM (
                              SELECT DISTINCT `item_id`
                              FROM `results`
                                       JOIN `groups_ancestors_active`
                                            ON `groups_ancestors_active`.`child_group_id` = `results`.`participant_id` AND
                                               `groups_ancestors_active`.`ancestor_group_id` = NEW.`child_group_id`
                              FOR SHARE
                          ) AS `result_items`
                     WHERE EXISTS(
                             SELECT 1
                             FROM `permissions_generated`
                                      JOIN `groups_ancestors_active` AS `grand_ancestors`
                                           ON `grand_ancestors`.`child_group_id` = NEW.`parent_group_id` AND
                                              `grand_ancestors`.`ancestor_group_id` = `permissions_generated`.`group_id`
                                      JOIN `items_ancestors`
                                           ON `items_ancestors`.`ancestor_item_id` = `permissions_generated`.`item_id`
                             WHERE `items_ancestors`.`child_item_id` = `result_items`.`item_id`
                               AND `permissions_generated`.`can_view_generated` != 'none'
                             FOR SHARE
                         )
                       AND NOT EXISTS(
                             SELECT 1
                             FROM `permissions_generated`
                                      JOIN `groups_ancestors_active` AS `child_ancestors`
                                           ON `child_ancestors`.`child_group_id` = NEW.`child_group_id` AND
                                              `child_ancestors`.`ancestor_group_id` = `permissions_generated`.`group_id`
                                      JOIN `items_ancestors`
                                           ON `items_ancestors`.`ancestor_item_id` = `permissions_generated`.`item_id`
                             WHERE `items_ancestors`.`child_item_id` = `result_items`.`item_id`
                               AND `permissions_generated`.`can_view_generated` != 'none'
                             FOR SHARE
                         )
                     FOR SHARE
                 ) AS `result_items_filtered`
            JOIN `results` ON `results`.`item_id` = `result_items_filtered`.`item_id`
            JOIN `groups_ancestors_active`
              ON `groups_ancestors_active`.`child_group_id` = `results`.`participant_id` AND
                 `groups_ancestors_active`.`ancestor_group_id` = NEW.`child_group_id`
            FOR SHARE
            ON DUPLICATE KEY UPDATE state = IF(state='propagating', 'to_be_propagated', state);
        END IF;

        INSERT INTO `groups_propagate` (`id`, `ancestors_computation_state`) VALUES (NEW.child_group_id, 'todo')
        ON DUPLICATE KEY UPDATE `ancestors_computation_state` = 'todo';
    END IF;
END */;;
DELIMITER ;
/*!50003 SET sql_mode              = @saved_sql_mode */ ;
/*!50003 SET character_set_client  = @saved_cs_client */ ;
/*!50003 SET character_set_results = @saved_cs_results */ ;
/*!50003 SET collation_connection  = @saved_col_connection */ ;
/*!50003 SET @saved_cs_client      = @@character_set_client */ ;
/*!50003 SET @saved_cs_results     = @@character_set_results */ ;
/*!50003 SET @saved_col_connection = @@collation_connection */ ;
/*!50003 SET character_set_client  = utf8mb4 */ ;
/*!50003 SET character_set_results = utf8mb4 */ ;
/*!50003 SET collation_connection  = utf8mb4_general_ci */ ;
/*!50003 SET @saved_sql_mode       = @@sql_mode */ ;
/*!50003 SET sql_mode              = 'ONLY_FULL_GROUP_BY,STRICT_TRANS_TABLES,NO_ZERO_IN_DATE,NO_ZERO_DATE,ERROR_FOR_DIVISION_BY_ZERO,NO_ENGINE_SUBSTITUTION' */ ;
DELIMITER ;;
/*!50003 CREATE TRIGGER `before_delete_groups_groups` BEFORE DELETE ON `groups_groups` FOR EACH ROW BEGIN
  IF NOT OLD.`is_team_membership` THEN
    INSERT INTO `groups_propagate` (`id`, `ancestors_computation_state`) VALUES (OLD.child_group_id, 'todo')
    ON DUPLICATE KEY UPDATE `ancestors_computation_state` = 'todo';
  END IF;
END */;;
DELIMITER ;
/*!50003 SET sql_mode              = @saved_sql_mode */ ;
/*!50003 SET character_set_client  = @saved_cs_client */ ;
/*!50003 SET character_set_results = @saved_cs_results */ ;
/*!50003 SET collation_connection  = @saved_col_connection */ ;

--
-- Temporary view structure for view `groups_groups_active`
--

DROP TABLE IF EXISTS `groups_groups_active`;
/*!50001 DROP VIEW IF EXISTS `groups_groups_active`*/;
SET @saved_cs_client     = @@character_set_client;
/*!50503 SET character_set_client = utf8mb4 */;
/*!50001 CREATE VIEW `groups_groups_active` AS SELECT
 1 AS `parent_group_id`,
 1 AS `child_group_id`,
 1 AS `expires_at`,
 1 AS `is_team_membership`,
 1 AS `personal_info_view_approved_at`,
 1 AS `personal_info_view_approved`,
 1 AS `lock_membership_approved_at`,
 1 AS `lock_membership_approved`,
 1 AS `watch_approved_at`,
 1 AS `watch_approved`*/;
SET character_set_client = @saved_cs_client;

--
-- Table structure for table `groups_propagate`
--

DROP TABLE IF EXISTS `groups_propagate`;
/*!40101 SET @saved_cs_client     = @@character_set_client */;
/*!50503 SET character_set_client = utf8mb4 */;
CREATE TABLE `groups_propagate` (
  `id` bigint NOT NULL,
  `ancestors_computation_state` enum('todo','done') NOT NULL,
  PRIMARY KEY (`id`),
  KEY `ancestors_computation_state` (`ancestors_computation_state`),
  CONSTRAINT `fk_groups_propagate_id_groups_id` FOREIGN KEY (`id`) REFERENCES `groups` (`id`) ON DELETE CASCADE
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_0900_ai_ci COMMENT='Used by the algorithm that updates the groups_ancestors table, and keeps track of what groups still need to have their relationship with their descendants / ancestors propagated.';
/*!40101 SET character_set_client = @saved_cs_client */;

--
-- Dumping data for table `groups_propagate`
--

LOCK TABLES `groups_propagate` WRITE;
/*!40000 ALTER TABLE `groups_propagate` DISABLE KEYS */;
INSERT INTO `groups_propagate` VALUES (777,'todo');
/*!40000 ALTER TABLE `groups_propagate` ENABLE KEYS */;
UNLOCK TABLES;

--
-- Table structure for table `item_dependencies`
--

DROP TABLE IF EXISTS `item_dependencies`;
/*!40101 SET @saved_cs_client     = @@character_set_client */;
/*!50503 SET character_set_client = utf8mb4 */;
CREATE TABLE `item_dependencies` (
  `item_id` bigint NOT NULL,
  `dependent_item_id` bigint NOT NULL,
  `score` int NOT NULL DEFAULT '100' COMMENT 'Score of the item from which the dependent item is unlocked (if grant_content_view is true), i.e. can_view:content is given',
  `grant_content_view` tinyint(1) NOT NULL DEFAULT '1' COMMENT 'Whether obtaining the required score at the item grants content view to the dependent item',
  PRIMARY KEY (`item_id`,`dependent_item_id`),
  KEY `fk_item_dependencies_dependent_item_id_items_id` (`dependent_item_id`),
  CONSTRAINT `fk_item_dependencies_dependent_item_id_items_id` FOREIGN KEY (`dependent_item_id`) REFERENCES `items` (`id`) ON DELETE CASCADE,
  CONSTRAINT `fk_item_dependencies_item_id_items_id` FOREIGN KEY (`item_id`) REFERENCES `items` (`id`) ON DELETE CASCADE
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_0900_ai_ci;
/*!40101 SET character_set_client = @saved_cs_client */;

--
-- Dumping data for table `item_dependencies`
--

LOCK TABLES `item_dependencies` WRITE;
/*!40000 ALTER TABLE `item_dependencies` DISABLE KEYS */;
/*!40000 ALTER TABLE `item_dependencies` ENABLE KEYS */;
UNLOCK TABLES;

--
-- Table structure for table `items`
--

DROP TABLE IF EXISTS `items`;
/*!40101 SET @saved_cs_client     = @@character_set_client */;
/*!50503 SET character_set_client = utf8mb4 */;
CREATE TABLE `items` (
  `id` bigint NOT NULL,
  `url` varchar(2048) DEFAULT NULL COMMENT 'Url of the item, as will be loaded in the iframe',
  `options` text COMMENT 'Options passed to the task, formatted as a JSON object',
  `platform_id` int DEFAULT NULL COMMENT 'Platform that hosts the item content. Auto-generated from `url` by triggers.',
  `text_id` varchar(200) DEFAULT NULL COMMENT 'Unique string identifying the item, independently of where it is hosted',
  `repository_path` text,
  `type` enum('Chapter','Task','Skill') NOT NULL,
  `title_bar_visible` tinyint unsigned NOT NULL DEFAULT '1' COMMENT 'Whether the title bar should be visible initially when this item is loaded',
  `display_details_in_parent` tinyint unsigned NOT NULL DEFAULT '0' COMMENT 'If true, display a large icon, the subtitle, and more within the parent chapter',
  `uses_api` tinyint(1) NOT NULL DEFAULT '1' COMMENT 'Whether the item uses the task integration API, at the minimum the load and getHeight functions.',
  `read_only` tinyint(1) NOT NULL DEFAULT '0' COMMENT 'Prevents any modification of the scores for this item (typically, to display a contest item after the end date of the contest)',
  `full_screen` enum('forceYes','forceNo','default') NOT NULL DEFAULT 'default' COMMENT 'Whether the item should be loaded in full screen mode (without the navigation panel and most of the top header). By default, tasks are displayed in full screen, but not chapters.',
  `hints_allowed` tinyint(1) NOT NULL DEFAULT '0' COMMENT 'Whether hints are allowed for tasks accessed through this chapter (currently unused)',
  `fixed_ranks` tinyint(1) NOT NULL DEFAULT '0' COMMENT 'If true, prevents users from changing the order of the children by drag&drop and auto-calculation of the order of children. Allows for manual setting of the order, for instance in cases where we want to have multiple items with the same order (check items_items.child_order).',
  `validation_type` enum('None','All','AllButOne','Categories','One','Manual') NOT NULL DEFAULT 'All' COMMENT 'Criteria for this item to be considered validated, based on the status of the children. Ex: "All" means all children should be validated. Categories means items of the "Validation" category need to be validated.',
  `supported_lang_prog` varchar(200) DEFAULT NULL COMMENT 'Comma-separated list of programming languages that this item can be solved with; not currently used.',
  `default_language_tag` varchar(6) NOT NULL COMMENT 'Default language tag of this task (the reference, used when comparing translations)',
  `requires_explicit_entry` tinyint(1) NOT NULL DEFAULT '0' COMMENT 'Whether this item requires an explicit entry to be started (create an attempt)',
  `entry_participant_type` enum('User','Team') NOT NULL DEFAULT 'User' COMMENT 'For explicit-entry items, the type of participants who can enter',
  `entry_min_admitted_members_ratio` enum('All','Half','One','None') NOT NULL DEFAULT 'None' COMMENT 'The ratio of members in the team (a user alone being considered as a team of one) who needs the “can_enter” permission so that the group can enter',
  `entering_time_min` datetime NOT NULL DEFAULT '1000-01-01 00:00:00' COMMENT 'Lower bound on the entering time. Has the priority over given can_enter_from/until permissions.',
  `entering_time_max` datetime NOT NULL DEFAULT '9999-12-31 23:59:59' COMMENT 'Upper bound on the entering time. Has the priority over given can_enter_from/until permissions.',
  `entry_frozen_teams` tinyint(1) NOT NULL DEFAULT '0' COMMENT 'Whether teams require to have `frozen_membership` for entering',
  `participants_group_id` bigint DEFAULT NULL COMMENT 'Group to which all the entered participants (users or teams) belong. Must not be null for an explicit-entry item.',
  `entry_max_team_size` int NOT NULL DEFAULT '0' COMMENT 'The maximum number of members a team can have to enter',
  `allows_multiple_attempts` tinyint(1) NOT NULL DEFAULT '0' COMMENT 'Whether participants can create multiple attempts when working on this item',
  `duration` time DEFAULT NULL COMMENT 'Not NULL if time-limited item. If so, how long users have to work on it.',
  `show_user_infos` tinyint(1) NOT NULL DEFAULT '0' COMMENT 'Always show user infos in title bar of all descendants. Allows the teacher to see who is working on what (e.g., during an exam).',
  `children_layout` enum('List','Grid') DEFAULT 'List' COMMENT 'How the children list are displayed (for chapters and skills)',
  `no_score` tinyint(1) NOT NULL DEFAULT '0' COMMENT 'Whether this item should not have any score displayed / propagated to the parent.',
  `prompt_to_join_group_by_code` tinyint(1) NOT NULL DEFAULT '0' COMMENT 'Whether the UI should display a form for joining a group by code on the item page',
  PRIMARY KEY (`id`),
  UNIQUE KEY `unique_text_id_unique` (`text_id`),
  KEY `fk_items_id_default_language_tag_items_strings_item_language_tag` (`id`,`default_language_tag`),
  KEY `fk_items_platform_id_platforms_id` (`platform_id`),
  CONSTRAINT `fk_items_id_default_language_tag_items_strings_item_language_tag` FOREIGN KEY (`id`, `default_language_tag`) REFERENCES `items_strings` (`item_id`, `language_tag`),
  CONSTRAINT `fk_items_platform_id_platforms_id` FOREIGN KEY (`platform_id`) REFERENCES `platforms` (`id`) ON DELETE RESTRICT
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_0900_ai_ci;
/*!40101 SET character_set_client = @saved_cs_client */;

--
-- Dumping data for table `items`
--

LOCK TABLES `items` WRITE;
/*!40000 ALTER TABLE `items` DISABLE KEYS */;
INSERT INTO `items` VALUES (694914435881177216,NULL,'{}',NULL,NULL,NULL,'Chapter',1,0,1,0,'default',0,0,'All',NULL,'fr',0,'User','None','1000-01-01 00:00:00','9999-12-31 23:59:59',0,NULL,0,0,NULL,0,'List',0,0);
/*!40000 ALTER TABLE `items` ENABLE KEYS */;
UNLOCK TABLES;
/*!50003 SET @saved_cs_client      = @@character_set_client */ ;
/*!50003 SET @saved_cs_results     = @@character_set_results */ ;
/*!50003 SET @saved_col_connection = @@collation_connection */ ;
/*!50003 SET character_set_client  = utf8mb4 */ ;
/*!50003 SET character_set_results = utf8mb4 */ ;
/*!50003 SET collation_connection  = utf8mb4_general_ci */ ;
/*!50003 SET @saved_sql_mode       = @@sql_mode */ ;
/*!50003 SET sql_mode              = 'ONLY_FULL_GROUP_BY,STRICT_TRANS_TABLES,NO_ZERO_IN_DATE,NO_ZERO_DATE,ERROR_FOR_DIVISION_BY_ZERO,NO_ENGINE_SUBSTITUTION' */ ;
DELIMITER ;;
/*!50003 CREATE TRIGGER `before_insert_items` BEFORE INSERT ON `items` FOR EACH ROW BEGIN
  IF (NEW.id IS NULL OR NEW.id = 0) THEN
    SET NEW.id = FLOOR(RAND() * 1000000000) + FLOOR(RAND() * 1000000000) * 1000000000;
  END IF;

  IF NEW.url IS NOT NULL THEN
    SET NEW.platform_id = (SELECT platforms.id FROM platforms WHERE NEW.url REGEXP platforms.regexp ORDER BY platforms.priority DESC LIMIT 1);
  ELSE
    SET NEW.platform_id = NULL;
  END IF;
END */;;
DELIMITER ;
/*!50003 SET sql_mode              = @saved_sql_mode */ ;
/*!50003 SET character_set_client  = @saved_cs_client */ ;
/*!50003 SET character_set_results = @saved_cs_results */ ;
/*!50003 SET collation_connection  = @saved_col_connection */ ;
/*!50003 SET @saved_cs_client      = @@character_set_client */ ;
/*!50003 SET @saved_cs_results     = @@character_set_results */ ;
/*!50003 SET @saved_col_connection = @@collation_connection */ ;
/*!50003 SET character_set_client  = utf8mb4 */ ;
/*!50003 SET character_set_results = utf8mb4 */ ;
/*!50003 SET collation_connection  = utf8mb4_general_ci */ ;
/*!50003 SET @saved_sql_mode       = @@sql_mode */ ;
/*!50003 SET sql_mode              = 'ONLY_FULL_GROUP_BY,STRICT_TRANS_TABLES,NO_ZERO_IN_DATE,NO_ZERO_DATE,ERROR_FOR_DIVISION_BY_ZERO,NO_ENGINE_SUBSTITUTION' */ ;
DELIMITER ;;
/*!50003 CREATE TRIGGER `before_update_items` BEFORE UPDATE ON `items` FOR EACH ROW BEGIN
  IF NOT OLD.url <=> NEW.url THEN
    IF NEW.url IS NOT NULL THEN
      SET NEW.platform_id = (SELECT platforms.id FROM platforms WHERE NEW.url REGEXP platforms.regexp ORDER BY platforms.priority DESC LIMIT 1);
    ELSE
      SET NEW.platform_id = NULL;
    END IF;
  END IF;
END */;;
DELIMITER ;
/*!50003 SET sql_mode              = @saved_sql_mode */ ;
/*!50003 SET character_set_client  = @saved_cs_client */ ;
/*!50003 SET character_set_results = @saved_cs_results */ ;
/*!50003 SET collation_connection  = @saved_col_connection */ ;

--
-- Table structure for table `items_ancestors`
--

DROP TABLE IF EXISTS `items_ancestors`;
/*!40101 SET @saved_cs_client     = @@character_set_client */;
/*!50503 SET character_set_client = utf8mb4 */;
CREATE TABLE `items_ancestors` (
  `ancestor_item_id` bigint NOT NULL,
  `child_item_id` bigint NOT NULL,
  PRIMARY KEY (`ancestor_item_id`,`child_item_id`),
  KEY `ancestor_item_id` (`ancestor_item_id`),
  KEY `child_item_id` (`child_item_id`),
  CONSTRAINT `fk_items_ancestors_ancestor_item_id_items_id` FOREIGN KEY (`ancestor_item_id`) REFERENCES `items` (`id`),
  CONSTRAINT `fk_items_ancestors_child_item_id_items_id` FOREIGN KEY (`child_item_id`) REFERENCES `items` (`id`) ON DELETE CASCADE
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_0900_ai_ci COMMENT='All child-ancestor relationships (a item is not its own ancestor). Cache table that can be recomputed based on the content of groups_groups.';
/*!40101 SET character_set_client = @saved_cs_client */;

--
-- Dumping data for table `items_ancestors`
--

LOCK TABLES `items_ancestors` WRITE;
/*!40000 ALTER TABLE `items_ancestors` DISABLE KEYS */;
/*!40000 ALTER TABLE `items_ancestors` ENABLE KEYS */;
UNLOCK TABLES;

--
-- Table structure for table `items_items`
--

DROP TABLE IF EXISTS `items_items`;
/*!40101 SET @saved_cs_client     = @@character_set_client */;
/*!50503 SET character_set_client = utf8mb4 */;
CREATE TABLE `items_items` (
  `parent_item_id` bigint NOT NULL,
  `child_item_id` bigint NOT NULL,
  `child_order` int NOT NULL COMMENT 'Position, relative to its siblings, when displaying all the children of the parent. If multiple items have the same child_order, they will be sorted in a random way, specific to each user (a user will always see the items in the same order).',
  `category` enum('Undefined','Discovery','Application','Validation','Challenge') NOT NULL DEFAULT 'Undefined' COMMENT 'Tag that indicates the role of this item, from the point of view of the parent item''s validation criteria. Also gives indication to the user of the role of the item.',
  `score_weight` tinyint unsigned NOT NULL DEFAULT '1' COMMENT 'Weight of this child in his parent''s score computation',
  `content_view_propagation` enum('none','as_info','as_content') NOT NULL DEFAULT 'none' COMMENT 'Defines how a can_view=”content” permission propagates',
  `upper_view_levels_propagation` enum('use_content_view_propagation','as_content_with_descendants','as_is') NOT NULL DEFAULT 'use_content_view_propagation' COMMENT 'Defines how can_view="content_with_descendants"|"solution" permissions propagate',
  `grant_view_propagation` tinyint(1) NOT NULL DEFAULT '0' COMMENT 'Whether can_grant_view propagates (as the same value, with “solution” as the upper limit)',
  `watch_propagation` tinyint(1) NOT NULL DEFAULT '0' COMMENT 'Whether can_watch propagates (as the same value, with “answer” as the upper limit)',
  `edit_propagation` tinyint(1) NOT NULL DEFAULT '0' COMMENT 'Whether can_edit propagates (as the same value, with “all” as the upper limit)',
  `request_help_propagation` tinyint NOT NULL DEFAULT '0' COMMENT 'Whether can_request_help_to propagates',
  `content_view_propagation_value` tinyint unsigned GENERATED ALWAYS AS (`content_view_propagation`) VIRTUAL NOT NULL COMMENT 'content_view_propagation as an integer (to use comparison operators)',
  `upper_view_levels_propagation_value` tinyint unsigned GENERATED ALWAYS AS (`upper_view_levels_propagation`) VIRTUAL NOT NULL COMMENT 'upper_view_levels_propagation as an integer (to use comparison operators)',
  PRIMARY KEY (`parent_item_id`,`child_item_id`),
  KEY `parent_item_id` (`parent_item_id`),
  KEY `child_item_id` (`child_item_id`),
  CONSTRAINT `fk_items_items_child_item_id_items_id` FOREIGN KEY (`child_item_id`) REFERENCES `items` (`id`) ON DELETE CASCADE,
  CONSTRAINT `fk_items_items_parent_item_id_items_id` FOREIGN KEY (`parent_item_id`) REFERENCES `items` (`id`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_0900_ai_ci COMMENT='Parent-child (N-N) relationship between items (acyclic graph)';
/*!40101 SET character_set_client = @saved_cs_client */;

--
-- Dumping data for table `items_items`
--

LOCK TABLES `items_items` WRITE;
/*!40000 ALTER TABLE `items_items` DISABLE KEYS */;
/*!40000 ALTER TABLE `items_items` ENABLE KEYS */;
UNLOCK TABLES;
/*!50003 SET @saved_cs_client      = @@character_set_client */ ;
/*!50003 SET @saved_cs_results     = @@character_set_results */ ;
/*!50003 SET @saved_col_connection = @@collation_connection */ ;
/*!50003 SET character_set_client  = utf8mb4 */ ;
/*!50003 SET character_set_results = utf8mb4 */ ;
/*!50003 SET collation_connection  = utf8mb4_general_ci */ ;
/*!50003 SET @saved_sql_mode       = @@sql_mode */ ;
/*!50003 SET sql_mode              = 'ONLY_FULL_GROUP_BY,STRICT_TRANS_TABLES,NO_ZERO_IN_DATE,NO_ZERO_DATE,ERROR_FOR_DIVISION_BY_ZERO,NO_ENGINE_SUBSTITUTION' */ ;
DELIMITER ;;
/*!50003 CREATE TRIGGER `before_insert_items_items` BEFORE INSERT ON `items_items` FOR EACH ROW BEGIN INSERT INTO `items_propagate` (id, ancestors_computation_state) VALUES (NEW.child_item_id, 'todo') ON DUPLICATE KEY UPDATE `ancestors_computation_state` = 'todo' ; END */;;
DELIMITER ;
/*!50003 SET sql_mode              = @saved_sql_mode */ ;
/*!50003 SET character_set_client  = @saved_cs_client */ ;
/*!50003 SET character_set_results = @saved_cs_results */ ;
/*!50003 SET collation_connection  = @saved_col_connection */ ;
/*!50003 SET @saved_cs_client      = @@character_set_client */ ;
/*!50003 SET @saved_cs_results     = @@character_set_results */ ;
/*!50003 SET @saved_col_connection = @@collation_connection */ ;
/*!50003 SET character_set_client  = utf8mb4 */ ;
/*!50003 SET character_set_results = utf8mb4 */ ;
/*!50003 SET collation_connection  = utf8mb4_general_ci */ ;
/*!50003 SET @saved_sql_mode       = @@sql_mode */ ;
/*!50003 SET sql_mode              = 'ONLY_FULL_GROUP_BY,STRICT_TRANS_TABLES,NO_ZERO_IN_DATE,NO_ZERO_DATE,ERROR_FOR_DIVISION_BY_ZERO,NO_ENGINE_SUBSTITUTION' */ ;
DELIMITER ;;
/*!50003 CREATE TRIGGER `after_insert_items_items` AFTER INSERT ON `items_items` FOR EACH ROW BEGIN
    REPLACE INTO `permissions_propagate` (`group_id`, `item_id`, `propagate_to`)
    SELECT `permissions_generated`.`group_id`, NEW.`child_item_id`, 'self' as `propagate_to`
    FROM `permissions_generated`
    WHERE `permissions_generated`.`item_id` = NEW.`parent_item_id`;

    INSERT INTO `results_propagate`
    SELECT `participant_id`, `attempt_id`, `item_id`, 'to_be_propagated' AS `state`
    FROM `results`
    WHERE `item_id` = NEW.`child_item_id`
    ON DUPLICATE KEY UPDATE state = IF(state='propagating', 'to_be_propagated', state);

    INSERT INTO `results_propagate`
    SELECT `participant_id`, `attempt_id`, `item_id`, 'to_be_recomputed' AS `state`
    FROM `results`
    WHERE `item_id` = NEW.`parent_item_id`
    ON DUPLICATE KEY UPDATE `state` = 'to_be_recomputed';
END */;;
DELIMITER ;
/*!50003 SET sql_mode              = @saved_sql_mode */ ;
/*!50003 SET character_set_client  = @saved_cs_client */ ;
/*!50003 SET character_set_results = @saved_cs_results */ ;
/*!50003 SET collation_connection  = @saved_col_connection */ ;
/*!50003 SET @saved_cs_client      = @@character_set_client */ ;
/*!50003 SET @saved_cs_results     = @@character_set_results */ ;
/*!50003 SET @saved_col_connection = @@collation_connection */ ;
/*!50003 SET character_set_client  = utf8mb4 */ ;
/*!50003 SET character_set_results = utf8mb4 */ ;
/*!50003 SET collation_connection  = utf8mb4_general_ci */ ;
/*!50003 SET @saved_sql_mode       = @@sql_mode */ ;
/*!50003 SET sql_mode              = 'ONLY_FULL_GROUP_BY,STRICT_TRANS_TABLES,NO_ZERO_IN_DATE,NO_ZERO_DATE,ERROR_FOR_DIVISION_BY_ZERO,NO_ENGINE_SUBSTITUTION' */ ;
DELIMITER ;;
/*!50003 CREATE TRIGGER `before_update_items_items` BEFORE UPDATE ON `items_items` FOR EACH ROW BEGIN
    IF (OLD.child_item_id != NEW.child_item_id OR OLD.parent_item_id != NEW.parent_item_id) THEN
        SIGNAL SQLSTATE '45000' SET MESSAGE_TEXT = 'Unable to change immutable items_items.parent_item_id and/or items_items.child_item_id';
    END IF;
END */;;
DELIMITER ;
/*!50003 SET sql_mode              = @saved_sql_mode */ ;
/*!50003 SET character_set_client  = @saved_cs_client */ ;
/*!50003 SET character_set_results = @saved_cs_results */ ;
/*!50003 SET collation_connection  = @saved_col_connection */ ;
/*!50003 SET @saved_cs_client      = @@character_set_client */ ;
/*!50003 SET @saved_cs_results     = @@character_set_results */ ;
/*!50003 SET @saved_col_connection = @@collation_connection */ ;
/*!50003 SET character_set_client  = utf8mb4 */ ;
/*!50003 SET character_set_results = utf8mb4 */ ;
/*!50003 SET collation_connection  = utf8mb4_general_ci */ ;
/*!50003 SET @saved_sql_mode       = @@sql_mode */ ;
/*!50003 SET sql_mode              = 'ONLY_FULL_GROUP_BY,STRICT_TRANS_TABLES,NO_ZERO_IN_DATE,NO_ZERO_DATE,ERROR_FOR_DIVISION_BY_ZERO,NO_ENGINE_SUBSTITUTION' */ ;
DELIMITER ;;
/*!50003 CREATE TRIGGER `after_update_items_items` AFTER UPDATE ON `items_items` FOR EACH ROW BEGIN
    IF (OLD.`content_view_propagation` != NEW.`content_view_propagation` OR
        OLD.`upper_view_levels_propagation` != NEW.`upper_view_levels_propagation` OR
        OLD.`grant_view_propagation` != NEW.`grant_view_propagation` OR
        OLD.`watch_propagation` != NEW.`watch_propagation` OR
        OLD.`edit_propagation` != NEW.`edit_propagation`) THEN
        REPLACE INTO `permissions_propagate` (`group_id`, `item_id`, `propagate_to`)
        SELECT `permissions_generated`.`group_id`, NEW.`child_item_id`, 'self' as `propagate_to`
        FROM `permissions_generated`
        WHERE `permissions_generated`.`item_id` = NEW.`parent_item_id`;
    END IF;
    IF (OLD.`category` != NEW.`category` OR OLD.`score_weight` != NEW.`score_weight`) THEN
        INSERT INTO `results_propagate`
        SELECT `participant_id`, `attempt_id`, `item_id`, 'to_be_recomputed' AS `state`
        FROM `results`
        WHERE `item_id` = NEW.`parent_item_id`
        ON DUPLICATE KEY UPDATE `state` = 'to_be_recomputed';
    END IF;
END */;;
DELIMITER ;
/*!50003 SET sql_mode              = @saved_sql_mode */ ;
/*!50003 SET character_set_client  = @saved_cs_client */ ;
/*!50003 SET character_set_results = @saved_cs_results */ ;
/*!50003 SET collation_connection  = @saved_col_connection */ ;
/*!50003 SET @saved_cs_client      = @@character_set_client */ ;
/*!50003 SET @saved_cs_results     = @@character_set_results */ ;
/*!50003 SET @saved_col_connection = @@collation_connection */ ;
/*!50003 SET character_set_client  = utf8mb4 */ ;
/*!50003 SET character_set_results = utf8mb4 */ ;
/*!50003 SET collation_connection  = utf8mb4_general_ci */ ;
/*!50003 SET @saved_sql_mode       = @@sql_mode */ ;
/*!50003 SET sql_mode              = 'ONLY_FULL_GROUP_BY,STRICT_TRANS_TABLES,NO_ZERO_IN_DATE,NO_ZERO_DATE,ERROR_FOR_DIVISION_BY_ZERO,NO_ENGINE_SUBSTITUTION' */ ;
DELIMITER ;;
/*!50003 CREATE TRIGGER `before_delete_items_items` BEFORE DELETE ON `items_items` FOR EACH ROW BEGIN
  INSERT INTO `items_propagate` (`id`, `ancestors_computation_state`)
  VALUES (OLD.child_item_id, 'todo') ON DUPLICATE KEY UPDATE `ancestors_computation_state` = 'todo';

  REPLACE INTO `permissions_propagate` (`group_id`, `item_id`, `propagate_to`)
  SELECT `permissions_generated`.`group_id`, OLD.`child_item_id`, 'self' as `propagate_to`
  FROM `permissions_generated`
  WHERE `permissions_generated`.`item_id` = OLD.`parent_item_id`;

  -- Some results' ancestors should probably be removed
  -- DELETE FROM `results` WHERE ...

  INSERT INTO `results_recompute_for_items` (`item_id`) VALUES (OLD.`parent_item_id`) ON DUPLICATE KEY UPDATE `item_id`=`item_id`;
END */;;
DELIMITER ;
/*!50003 SET sql_mode              = @saved_sql_mode */ ;
/*!50003 SET character_set_client  = @saved_cs_client */ ;
/*!50003 SET character_set_results = @saved_cs_results */ ;
/*!50003 SET collation_connection  = @saved_col_connection */ ;

--
-- Table structure for table `items_propagate`
--

DROP TABLE IF EXISTS `items_propagate`;
/*!40101 SET @saved_cs_client     = @@character_set_client */;
/*!50503 SET character_set_client = utf8mb4 */;
CREATE TABLE `items_propagate` (
  `id` bigint NOT NULL,
  `ancestors_computation_state` enum('todo','done') NOT NULL,
  PRIMARY KEY (`id`),
  KEY `ancestors_computation_date` (`ancestors_computation_state`),
  CONSTRAINT `fk_id` FOREIGN KEY (`id`) REFERENCES `items` (`id`) ON DELETE CASCADE
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_0900_ai_ci COMMENT='Used by the algorithm that updates the items_ancestors table, and keeps track of what items still need to have their relationship with their descendants / ancestors propagated.';
/*!40101 SET character_set_client = @saved_cs_client */;

--
-- Dumping data for table `items_propagate`
--

LOCK TABLES `items_propagate` WRITE;
/*!40000 ALTER TABLE `items_propagate` DISABLE KEYS */;
INSERT INTO `items_propagate` VALUES (694914435881177216,'todo');
/*!40000 ALTER TABLE `items_propagate` ENABLE KEYS */;
UNLOCK TABLES;

--
-- Table structure for table `items_strings`
--

DROP TABLE IF EXISTS `items_strings`;
/*!40101 SET @saved_cs_client     = @@character_set_client */;
/*!50503 SET character_set_client = utf8mb4 */;
CREATE TABLE `items_strings` (
  `item_id` bigint NOT NULL,
  `language_tag` varchar(6) NOT NULL COMMENT 'Language tag of this content',
  `translator` varchar(100) DEFAULT NULL COMMENT 'Name of the translator(s) of this content',
  `title` varchar(200) DEFAULT NULL COMMENT 'Title of the item, in the specified language',
  `image_url` varchar(2048) DEFAULT NULL COMMENT 'Url of a small image associated with this item.',
  `subtitle` varchar(200) DEFAULT NULL COMMENT 'Subtitle of the item in the specified language',
  `description` text COMMENT 'Description of the item in the specified language',
  `edu_comment` text COMMENT 'Information about what this item teaches, in the specified language.',
  PRIMARY KEY (`item_id`,`language_tag`),
  KEY `item_id` (`item_id`),
  KEY `fk_items_strings_language_tag_languages_tag` (`language_tag`),
  FULLTEXT KEY `fullTextTitle` (`title`),
  CONSTRAINT `fk_items_strings_item_id_items_id` FOREIGN KEY (`item_id`) REFERENCES `items` (`id`) ON DELETE CASCADE,
  CONSTRAINT `fk_items_strings_language_tag_languages_tag` FOREIGN KEY (`language_tag`) REFERENCES `languages` (`tag`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_0900_ai_ci COMMENT='Textual content associated with an item, in a given language.';
/*!40101 SET character_set_client = @saved_cs_client */;

--
-- Dumping data for table `items_strings`
--

LOCK TABLES `items_strings` WRITE;
/*!40000 ALTER TABLE `items_strings` DISABLE KEYS */;
INSERT INTO `items_strings` VALUES (694914435881177216,'en',NULL,'Public activities',NULL,NULL,NULL,NULL),(694914435881177216,'fr',NULL,'Activités publiques',NULL,NULL,NULL,NULL);
/*!40000 ALTER TABLE `items_strings` ENABLE KEYS */;
UNLOCK TABLES;

--
-- Table structure for table `languages`
--

DROP TABLE IF EXISTS `languages`;
/*!40101 SET @saved_cs_client     = @@character_set_client */;
/*!50503 SET character_set_client = utf8mb4 */;
CREATE TABLE `languages` (
  `tag` varchar(6) NOT NULL COMMENT 'Language tag as defined in RFC5646',
  `name` varchar(100) NOT NULL DEFAULT '',
  PRIMARY KEY (`tag`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_0900_ai_ci COMMENT='Languages supported for content';
/*!40101 SET character_set_client = @saved_cs_client */;

--
-- Dumping data for table `languages`
--

LOCK TABLES `languages` WRITE;
/*!40000 ALTER TABLE `languages` DISABLE KEYS */;
/*!40000 ALTER TABLE `languages` ENABLE KEYS */;
UNLOCK TABLES;

--
-- Table structure for table `permissions_generated`
--

DROP TABLE IF EXISTS `permissions_generated`;
/*!40101 SET @saved_cs_client     = @@character_set_client */;
/*!50503 SET character_set_client = utf8mb4 */;
CREATE TABLE `permissions_generated` (
  `group_id` bigint NOT NULL,
  `item_id` bigint NOT NULL,
  `can_view_generated` enum('none','info','content','content_with_descendants','solution') NOT NULL DEFAULT 'none' COMMENT 'The aggregated level of visibility the group has on the item',
  `can_grant_view_generated` enum('none','enter','content','content_with_descendants','solution','solution_with_grant') NOT NULL DEFAULT 'none' COMMENT 'The aggregated level of visibility that the group can give on this item to other groups on which it has the right to',
  `can_watch_generated` enum('none','result','answer','answer_with_grant') NOT NULL DEFAULT 'none' COMMENT 'The aggregated level of observation a group has for an item, on the activity of the users he can watch',
  `can_edit_generated` enum('none','children','all','all_with_grant') NOT NULL DEFAULT 'none' COMMENT 'The aggregated level of edition permissions a group has on an item',
  `is_owner_generated` tinyint(1) NOT NULL DEFAULT '0' COMMENT 'Whether the group is the owner of this item. Implies the maximum level in all of the above permissions. Can delete the item.',
  `can_view_generated_value` tinyint unsigned GENERATED ALWAYS AS ((`can_view_generated` + 0)) VIRTUAL NOT NULL COMMENT 'can_view_generated as an integer (to use comparison operators)',
  `can_grant_view_generated_value` tinyint unsigned GENERATED ALWAYS AS ((`can_grant_view_generated` + 0)) VIRTUAL NOT NULL COMMENT 'can_grant_view_generated as an integer (to use comparison operators)',
  `can_watch_generated_value` tinyint unsigned GENERATED ALWAYS AS ((`can_watch_generated` + 0)) VIRTUAL NOT NULL COMMENT 'can_watch_generated as an integer (to use comparison operators)',
  `can_edit_generated_value` tinyint unsigned GENERATED ALWAYS AS ((`can_edit_generated` + 0)) VIRTUAL NOT NULL COMMENT 'can_edit_generated as an integer (to use comparison operators)',
  PRIMARY KEY (`group_id`,`item_id`),
  KEY `item_id` (`item_id`),
  CONSTRAINT `fk_permissions_generated_group_id_groups_id` FOREIGN KEY (`group_id`) REFERENCES `groups` (`id`) ON DELETE CASCADE,
  CONSTRAINT `fk_permissions_generated_item_id_items_id` FOREIGN KEY (`item_id`) REFERENCES `items` (`id`) ON DELETE CASCADE
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_0900_ai_ci COMMENT='Actual permissions that the group has, considering the aggregation and the propagation';
/*!40101 SET character_set_client = @saved_cs_client */;

--
-- Dumping data for table `permissions_generated`
--

LOCK TABLES `permissions_generated` WRITE;
/*!40000 ALTER TABLE `permissions_generated` DISABLE KEYS */;
/*!40000 ALTER TABLE `permissions_generated` ENABLE KEYS */;
UNLOCK TABLES;
/*!50003 SET @saved_cs_client      = @@character_set_client */ ;
/*!50003 SET @saved_cs_results     = @@character_set_results */ ;
/*!50003 SET @saved_col_connection = @@collation_connection */ ;
/*!50003 SET character_set_client  = utf8mb4 */ ;
/*!50003 SET character_set_results = utf8mb4 */ ;
/*!50003 SET collation_connection  = utf8mb4_general_ci */ ;
/*!50003 SET @saved_sql_mode       = @@sql_mode */ ;
/*!50003 SET sql_mode              = 'ONLY_FULL_GROUP_BY,STRICT_TRANS_TABLES,NO_ZERO_IN_DATE,NO_ZERO_DATE,ERROR_FOR_DIVISION_BY_ZERO,NO_ENGINE_SUBSTITUTION' */ ;
DELIMITER ;;
/*!50003 CREATE TRIGGER `after_insert_permissions_generated` AFTER INSERT ON `permissions_generated` FOR EACH ROW BEGIN
    IF NEW.`can_view_generated` != 'none' THEN
      IF @synchronous_propagations_connection_id > 0 THEN
        INSERT INTO `results_propagate_sync`
        SELECT @synchronous_propagations_connection_id AS `connection_id`, `participant_id`, `attempt_id`, `item_id`, 'to_be_propagated' AS `state`
        FROM `results`
            JOIN `items_ancestors` ON `items_ancestors`.`child_item_id` = `results`.`item_id` AND
                                      `items_ancestors`.`ancestor_item_id` = NEW.`item_id`
            JOIN `groups_ancestors_active` ON `groups_ancestors_active`.`child_group_id` = `results`.`participant_id` AND
                                              `groups_ancestors_active`.`ancestor_group_id` = NEW.`group_id`
        ON DUPLICATE KEY UPDATE state = IF(state='propagating', 'to_be_propagated', state);
      ELSE
        INSERT INTO `results_propagate`
        SELECT `participant_id`, `attempt_id`, `item_id`, 'to_be_propagated' AS `state`
        FROM `results`
            JOIN `items_ancestors` ON `items_ancestors`.`child_item_id` = `results`.`item_id` AND
                                      `items_ancestors`.`ancestor_item_id` = NEW.`item_id`
            JOIN `groups_ancestors_active` ON `groups_ancestors_active`.`child_group_id` = `results`.`participant_id` AND
                                              `groups_ancestors_active`.`ancestor_group_id` = NEW.`group_id`
        ON DUPLICATE KEY UPDATE state = IF(state='propagating', 'to_be_propagated', state);
      END IF;
    END IF;
END */;;
DELIMITER ;
/*!50003 SET sql_mode              = @saved_sql_mode */ ;
/*!50003 SET character_set_client  = @saved_cs_client */ ;
/*!50003 SET character_set_results = @saved_cs_results */ ;
/*!50003 SET collation_connection  = @saved_col_connection */ ;
/*!50003 SET @saved_cs_client      = @@character_set_client */ ;
/*!50003 SET @saved_cs_results     = @@character_set_results */ ;
/*!50003 SET @saved_col_connection = @@collation_connection */ ;
/*!50003 SET character_set_client  = utf8mb4 */ ;
/*!50003 SET character_set_results = utf8mb4 */ ;
/*!50003 SET collation_connection  = utf8mb4_general_ci */ ;
/*!50003 SET @saved_sql_mode       = @@sql_mode */ ;
/*!50003 SET sql_mode              = 'ONLY_FULL_GROUP_BY,STRICT_TRANS_TABLES,NO_ZERO_IN_DATE,NO_ZERO_DATE,ERROR_FOR_DIVISION_BY_ZERO,NO_ENGINE_SUBSTITUTION' */ ;
DELIMITER ;;
/*!50003 CREATE TRIGGER `before_update_permissions_generated` BEFORE UPDATE ON `permissions_generated` FOR EACH ROW BEGIN
    IF OLD.`group_id` != NEW.`group_id` OR OLD.`item_id` != NEW.`item_id` THEN
        SIGNAL SQLSTATE '45000' SET MESSAGE_TEXT = 'Unable to change immutable permissions_generated.group_id and/or permissions_generated.child_item_id';
    END IF;
END */;;
DELIMITER ;
/*!50003 SET sql_mode              = @saved_sql_mode */ ;
/*!50003 SET character_set_client  = @saved_cs_client */ ;
/*!50003 SET character_set_results = @saved_cs_results */ ;
/*!50003 SET collation_connection  = @saved_col_connection */ ;
/*!50003 SET @saved_cs_client      = @@character_set_client */ ;
/*!50003 SET @saved_cs_results     = @@character_set_results */ ;
/*!50003 SET @saved_col_connection = @@collation_connection */ ;
/*!50003 SET character_set_client  = utf8mb4 */ ;
/*!50003 SET character_set_results = utf8mb4 */ ;
/*!50003 SET collation_connection  = utf8mb4_general_ci */ ;
/*!50003 SET @saved_sql_mode       = @@sql_mode */ ;
/*!50003 SET sql_mode              = 'ONLY_FULL_GROUP_BY,STRICT_TRANS_TABLES,NO_ZERO_IN_DATE,NO_ZERO_DATE,ERROR_FOR_DIVISION_BY_ZERO,NO_ENGINE_SUBSTITUTION' */ ;
DELIMITER ;;
/*!50003 CREATE TRIGGER `after_update_permissions_generated` AFTER UPDATE ON `permissions_generated` FOR EACH ROW BEGIN
    IF OLD.`can_view_generated` = 'none' AND NEW.can_view_generated != 'none' THEN
      IF @synchronous_propagations_connection_id > 0 THEN
        INSERT INTO `results_propagate_sync`
        SELECT @synchronous_propagations_connection_id AS `connection_id`, `participant_id`, `attempt_id`, `item_id`, 'to_be_propagated' AS `state`
        FROM `results`
            JOIN `items_ancestors` ON `items_ancestors`.`child_item_id` = `results`.`item_id` AND
                                      `items_ancestors`.`ancestor_item_id` = NEW.`item_id`
            JOIN `groups_ancestors_active` ON `groups_ancestors_active`.`child_group_id` = `results`.`participant_id` AND
                                              `groups_ancestors_active`.`ancestor_group_id` = NEW.`group_id`
        ON DUPLICATE KEY UPDATE state = IF(state='propagating', 'to_be_propagated', state);
      ELSE
        INSERT INTO `results_propagate`
        SELECT `participant_id`, `attempt_id`, `item_id`, 'to_be_propagated' AS `state`
        FROM `results`
            JOIN `items_ancestors` ON `items_ancestors`.`child_item_id` = `results`.`item_id` AND
                                      `items_ancestors`.`ancestor_item_id` = NEW.`item_id`
            JOIN `groups_ancestors_active` ON `groups_ancestors_active`.`child_group_id` = `results`.`participant_id` AND
                                              `groups_ancestors_active`.`ancestor_group_id` = NEW.`group_id`
        ON DUPLICATE KEY UPDATE state = IF(state='propagating', 'to_be_propagated', state);
      END IF;
    END IF;
END */;;
DELIMITER ;
/*!50003 SET sql_mode              = @saved_sql_mode */ ;
/*!50003 SET character_set_client  = @saved_cs_client */ ;
/*!50003 SET character_set_results = @saved_cs_results */ ;
/*!50003 SET collation_connection  = @saved_col_connection */ ;

--
-- Table structure for table `permissions_granted`
--

DROP TABLE IF EXISTS `permissions_granted`;
/*!40101 SET @saved_cs_client     = @@character_set_client */;
/*!50503 SET character_set_client = utf8mb4 */;
CREATE TABLE `permissions_granted` (
  `group_id` bigint NOT NULL,
  `item_id` bigint NOT NULL,
  `source_group_id` bigint NOT NULL,
  `origin` enum('group_membership','item_unlocking','self','other') NOT NULL,
  `latest_update_at` datetime NOT NULL DEFAULT CURRENT_TIMESTAMP COMMENT 'Last time one of the attributes has been modified',
  `can_view` enum('none','info','content','content_with_descendants','solution') NOT NULL DEFAULT 'none' COMMENT 'The level of visibility the group has on the item',
  `can_enter_from` datetime NOT NULL DEFAULT '9999-12-31 23:59:59' COMMENT 'Time from which the group can “enter” this item, superseded by `items.entering_time_min`',
  `can_enter_until` datetime NOT NULL DEFAULT '9999-12-31 23:59:59' COMMENT 'Time until which the group can “enter” this item, superseded by `items.entering_time_max`',
  `can_grant_view` enum('none','enter','content','content_with_descendants','solution','solution_with_grant') NOT NULL DEFAULT 'none' COMMENT 'The level of visibility that the group can give on this item to other groups on which it has the right to',
  `can_watch` enum('none','result','answer','answer_with_grant') NOT NULL DEFAULT 'none' COMMENT 'The level of observation a group has for an item, on the activity of the users he can watch',
  `can_edit` enum('none','children','all','all_with_grant') NOT NULL DEFAULT 'none' COMMENT 'The level of edition permissions a group has on an item',
  `can_make_session_official` tinyint(1) NOT NULL DEFAULT '0' COMMENT 'Whether the group is allowed to associate official sessions to this item',
  `is_owner` tinyint(1) NOT NULL DEFAULT '0' COMMENT 'Whether the group is the owner of this item. Implies the maximum level in all of the above permissions. Can delete the item.',
  `can_view_value` tinyint unsigned GENERATED ALWAYS AS ((`can_view` + 0)) VIRTUAL NOT NULL COMMENT 'can_view as an integer (to use comparison operators)',
  `can_grant_view_value` tinyint unsigned GENERATED ALWAYS AS ((`can_grant_view` + 0)) VIRTUAL NOT NULL COMMENT 'can_grant_view as an integer (to use comparison operators)',
  `can_watch_value` tinyint unsigned GENERATED ALWAYS AS ((`can_watch` + 0)) VIRTUAL NOT NULL COMMENT 'can_watch as an integer (to use comparison operators)',
  `can_edit_value` tinyint unsigned GENERATED ALWAYS AS ((`can_edit` + 0)) VIRTUAL NOT NULL COMMENT 'can_edit as an integer (to use comparison operators)',
  `can_request_help_to` bigint DEFAULT NULL COMMENT 'Whether the group can create a forum thread accessible to the pointed group. NULL = no rights to create.',
  PRIMARY KEY (`group_id`,`item_id`,`source_group_id`,`origin`),
  KEY `group_id_item_id` (`group_id`,`item_id`),
  KEY `fk_permissions_granted_item_id_items_id` (`item_id`),
  KEY `fk_permissions_granted_source_group_id_groups_id` (`source_group_id`),
  KEY `fk_can_request_help_to_groups_id` (`can_request_help_to`),
  CONSTRAINT `fk_can_request_help_to_groups_id` FOREIGN KEY (`can_request_help_to`) REFERENCES `groups` (`id`) ON DELETE SET NULL,
  CONSTRAINT `fk_permissions_granted_group_id_groups_id` FOREIGN KEY (`group_id`) REFERENCES `groups` (`id`) ON DELETE CASCADE,
  CONSTRAINT `fk_permissions_granted_item_id_items_id` FOREIGN KEY (`item_id`) REFERENCES `items` (`id`) ON DELETE CASCADE,
  CONSTRAINT `fk_permissions_granted_source_group_id_groups_id` FOREIGN KEY (`source_group_id`) REFERENCES `groups` (`id`) ON DELETE CASCADE
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_0900_ai_ci COMMENT='Raw permissions given to a group on an item';
/*!40101 SET character_set_client = @saved_cs_client */;

--
-- Dumping data for table `permissions_granted`
--

LOCK TABLES `permissions_granted` WRITE;
/*!40000 ALTER TABLE `permissions_granted` DISABLE KEYS */;
/*!40000 ALTER TABLE `permissions_granted` ENABLE KEYS */;
UNLOCK TABLES;
/*!50003 SET @saved_cs_client      = @@character_set_client */ ;
/*!50003 SET @saved_cs_results     = @@character_set_results */ ;
/*!50003 SET @saved_col_connection = @@collation_connection */ ;
/*!50003 SET character_set_client  = utf8mb4 */ ;
/*!50003 SET character_set_results = utf8mb4 */ ;
/*!50003 SET collation_connection  = utf8mb4_general_ci */ ;
/*!50003 SET @saved_sql_mode       = @@sql_mode */ ;
/*!50003 SET sql_mode              = 'ONLY_FULL_GROUP_BY,STRICT_TRANS_TABLES,NO_ZERO_IN_DATE,NO_ZERO_DATE,ERROR_FOR_DIVISION_BY_ZERO,NO_ENGINE_SUBSTITUTION' */ ;
DELIMITER ;;
/*!50003 CREATE TRIGGER `after_insert_permissions_granted` AFTER INSERT ON `permissions_granted` FOR EACH ROW BEGIN
  IF @synchronous_propagations_connection_id > 0 THEN
    REPLACE INTO `permissions_propagate_sync` (`connection_id`, `group_id`, `item_id`, `propagate_to`)
      VALUE (@synchronous_propagations_connection_id, NEW.`group_id`, NEW.`item_id`, 'self');
  ELSE
    REPLACE INTO `permissions_propagate` (`group_id`, `item_id`, `propagate_to`)
      VALUE (NEW.`group_id`, NEW.`item_id`, 'self');
  END IF;
END */;;
DELIMITER ;
/*!50003 SET sql_mode              = @saved_sql_mode */ ;
/*!50003 SET character_set_client  = @saved_cs_client */ ;
/*!50003 SET character_set_results = @saved_cs_results */ ;
/*!50003 SET collation_connection  = @saved_col_connection */ ;
/*!50003 SET @saved_cs_client      = @@character_set_client */ ;
/*!50003 SET @saved_cs_results     = @@character_set_results */ ;
/*!50003 SET @saved_col_connection = @@collation_connection */ ;
/*!50003 SET character_set_client  = utf8mb4 */ ;
/*!50003 SET character_set_results = utf8mb4 */ ;
/*!50003 SET collation_connection  = utf8mb4_general_ci */ ;
/*!50003 SET @saved_sql_mode       = @@sql_mode */ ;
/*!50003 SET sql_mode              = 'ONLY_FULL_GROUP_BY,STRICT_TRANS_TABLES,NO_ZERO_IN_DATE,NO_ZERO_DATE,ERROR_FOR_DIVISION_BY_ZERO,NO_ENGINE_SUBSTITUTION' */ ;
DELIMITER ;;
/*!50003 CREATE TRIGGER `after_update_permissions_granted` AFTER UPDATE ON `permissions_granted` FOR EACH ROW BEGIN
    IF NOT (NEW.`can_view` <=> OLD.`can_view` AND NEW.`can_grant_view` <=> OLD.`can_grant_view` AND
            NEW.`can_watch` <=> OLD.`can_watch` AND NEW.`can_edit` <=> OLD.`can_edit` AND
            NEW.`is_owner` <=> OLD.`is_owner`) THEN
      IF @synchronous_propagations_connection_id > 0 THEN
        REPLACE INTO `permissions_propagate_sync` (`connection_id`, `group_id`, `item_id`, `propagate_to`)
          VALUE (@synchronous_propagations_connection_id, NEW.`group_id`, NEW.`item_id`, 'self');
      ELSE
        REPLACE INTO `permissions_propagate` (`group_id`, `item_id`, `propagate_to`)
          VALUE (NEW.`group_id`, NEW.`item_id`, 'self');
      END IF;
    END IF;
END */;;
DELIMITER ;
/*!50003 SET sql_mode              = @saved_sql_mode */ ;
/*!50003 SET character_set_client  = @saved_cs_client */ ;
/*!50003 SET character_set_results = @saved_cs_results */ ;
/*!50003 SET collation_connection  = @saved_col_connection */ ;
/*!50003 SET @saved_cs_client      = @@character_set_client */ ;
/*!50003 SET @saved_cs_results     = @@character_set_results */ ;
/*!50003 SET @saved_col_connection = @@collation_connection */ ;
/*!50003 SET character_set_client  = utf8mb4 */ ;
/*!50003 SET character_set_results = utf8mb4 */ ;
/*!50003 SET collation_connection  = utf8mb4_general_ci */ ;
/*!50003 SET @saved_sql_mode       = @@sql_mode */ ;
/*!50003 SET sql_mode              = 'ONLY_FULL_GROUP_BY,STRICT_TRANS_TABLES,NO_ZERO_IN_DATE,NO_ZERO_DATE,ERROR_FOR_DIVISION_BY_ZERO,NO_ENGINE_SUBSTITUTION' */ ;
DELIMITER ;;
/*!50003 CREATE TRIGGER `after_delete_permissions_granted` AFTER DELETE ON `permissions_granted` FOR EACH ROW BEGIN
  IF @synchronous_propagations_connection_id > 0 THEN
    REPLACE INTO `permissions_propagate_sync` (`connection_id`, `group_id`, `item_id`, `propagate_to`)
      VALUE (@synchronous_propagations_connection_id, OLD.`group_id`, OLD.`item_id`, 'self');
  ELSE
    REPLACE INTO `permissions_propagate` (`group_id`, `item_id`, `propagate_to`)
      VALUE (OLD.`group_id`, OLD.`item_id`, 'self');
  END IF;
END */;;
DELIMITER ;
/*!50003 SET sql_mode              = @saved_sql_mode */ ;
/*!50003 SET character_set_client  = @saved_cs_client */ ;
/*!50003 SET character_set_results = @saved_cs_results */ ;
/*!50003 SET collation_connection  = @saved_col_connection */ ;

--
-- Table structure for table `permissions_propagate`
--

DROP TABLE IF EXISTS `permissions_propagate`;
/*!40101 SET @saved_cs_client     = @@character_set_client */;
/*!50503 SET character_set_client = utf8mb4 */;
CREATE TABLE `permissions_propagate` (
  `group_id` bigint NOT NULL,
  `item_id` bigint NOT NULL,
  `propagate_to` enum('self','children') NOT NULL COMMENT 'Which permissions should be recomputed for the group-item pair on the next iteration, either for the pair or for its children (through item hierarchy)',
  PRIMARY KEY (`group_id`,`item_id`),
  KEY `fk_permissions_propagate_item_id_items_id` (`item_id`),
  KEY `propagate_to_group_id_item_id` (`propagate_to`,`group_id`,`item_id`),
  CONSTRAINT `fk_permissions_propagate_group_id_groups_id` FOREIGN KEY (`group_id`) REFERENCES `groups` (`id`) ON DELETE CASCADE,
  CONSTRAINT `fk_permissions_propagate_item_id_items_id` FOREIGN KEY (`item_id`) REFERENCES `items` (`id`) ON DELETE CASCADE
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_0900_ai_ci COMMENT='Used by the access rights propagation algorithm to keep track of the status of the propagation';
/*!40101 SET character_set_client = @saved_cs_client */;

--
-- Dumping data for table `permissions_propagate`
--

LOCK TABLES `permissions_propagate` WRITE;
/*!40000 ALTER TABLE `permissions_propagate` DISABLE KEYS */;
/*!40000 ALTER TABLE `permissions_propagate` ENABLE KEYS */;
UNLOCK TABLES;

--
-- Table structure for table `permissions_propagate_sync`
--

DROP TABLE IF EXISTS `permissions_propagate_sync`;
/*!40101 SET @saved_cs_client     = @@character_set_client */;
/*!50503 SET character_set_client = utf8mb4 */;
CREATE TABLE `permissions_propagate_sync` (
  `connection_id` bigint unsigned NOT NULL,
  `group_id` bigint NOT NULL,
  `item_id` bigint NOT NULL,
  `propagate_to` enum('self','children') NOT NULL COMMENT 'Which permissions should be recomputed for the group-item pair on the next iteration, either for the pair or for its children (through item hierarchy)',
  PRIMARY KEY (`connection_id`,`group_id`,`item_id`),
  KEY `fk_permissions_propagate_item_id_items_id` (`item_id`),
  KEY `connection_id_propagate_to_group_id_item_id` (`connection_id`,`propagate_to`,`group_id`,`item_id`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_0900_ai_ci COMMENT='Used by the access rights propagation algorithm to keep track of the status of the propagation';
/*!40101 SET character_set_client = @saved_cs_client */;

--
-- Dumping data for table `permissions_propagate_sync`
--

LOCK TABLES `permissions_propagate_sync` WRITE;
/*!40000 ALTER TABLE `permissions_propagate_sync` DISABLE KEYS */;
/*!40000 ALTER TABLE `permissions_propagate_sync` ENABLE KEYS */;
UNLOCK TABLES;

--
-- Temporary view structure for view `permissions_propagate_sync_conn`
--

DROP TABLE IF EXISTS `permissions_propagate_sync_conn`;
/*!50001 DROP VIEW IF EXISTS `permissions_propagate_sync_conn`*/;
SET @saved_cs_client     = @@character_set_client;
/*!50503 SET character_set_client = utf8mb4 */;
/*!50001 CREATE VIEW `permissions_propagate_sync_conn` AS SELECT
 1 AS `connection_id`,
 1 AS `group_id`,
 1 AS `item_id`,
 1 AS `propagate_to`*/;
SET character_set_client = @saved_cs_client;

--
-- Table structure for table `platforms`
--

DROP TABLE IF EXISTS `platforms`;
/*!40101 SET @saved_cs_client     = @@character_set_client */;
/*!50503 SET character_set_client = utf8mb4 */;
CREATE TABLE `platforms` (
  `id` int NOT NULL AUTO_INCREMENT,
  `name` varchar(50) NOT NULL DEFAULT '',
  `base_url` varchar(200) DEFAULT NULL COMMENT 'Base URL for calling the API of the platform (for GDPR services)',
  `public_key` varchar(512) DEFAULT NULL COMMENT 'Public key of this platform',
  `regexp` text COMMENT 'Regexp matching the urls, to automatically detect content from this platform. It is the only way to specify which items are from which platform. Recomputation of items.platform_id is triggered when changed.',
  `priority` int NOT NULL DEFAULT '0' COMMENT 'Priority of the regexp compared to others (higher value is tried first). Recomputation of items.platform_id is triggered when changed.',
  PRIMARY KEY (`id`),
  UNIQUE KEY `priority` (`priority` DESC)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_0900_ai_ci COMMENT='Platforms that host content';
/*!40101 SET character_set_client = @saved_cs_client */;

--
-- Dumping data for table `platforms`
--

LOCK TABLES `platforms` WRITE;
/*!40000 ALTER TABLE `platforms` DISABLE KEYS */;
/*!40000 ALTER TABLE `platforms` ENABLE KEYS */;
UNLOCK TABLES;
/*!50003 SET @saved_cs_client      = @@character_set_client */ ;
/*!50003 SET @saved_cs_results     = @@character_set_results */ ;
/*!50003 SET @saved_col_connection = @@collation_connection */ ;
/*!50003 SET character_set_client  = utf8mb4 */ ;
/*!50003 SET character_set_results = utf8mb4 */ ;
/*!50003 SET collation_connection  = utf8mb4_general_ci */ ;
/*!50003 SET @saved_sql_mode       = @@sql_mode */ ;
/*!50003 SET sql_mode              = 'ONLY_FULL_GROUP_BY,STRICT_TRANS_TABLES,NO_ZERO_IN_DATE,NO_ZERO_DATE,ERROR_FOR_DIVISION_BY_ZERO,NO_ENGINE_SUBSTITUTION' */ ;
DELIMITER ;;
/*!50003 CREATE TRIGGER `after_insert_platforms` AFTER INSERT ON `platforms` FOR EACH ROW BEGIN
    UPDATE `items`
        LEFT JOIN `platforms` AS `old_platform` ON `old_platform`.`id` = `items`.`platform_id`
    SET `items`.`platform_id` = (
        SELECT `platforms`.`id` FROM `platforms`
        WHERE `items`.`url` REGEXP `platforms`.`regexp`
        ORDER BY `platforms`.`priority` DESC
        LIMIT 1
    )
    WHERE `old_platform`.`priority` < NEW.`priority` OR (`items`.`url` IS NOT NULL AND `old_platform`.`id` IS NULL);
END */;;
DELIMITER ;
/*!50003 SET sql_mode              = @saved_sql_mode */ ;
/*!50003 SET character_set_client  = @saved_cs_client */ ;
/*!50003 SET character_set_results = @saved_cs_results */ ;
/*!50003 SET collation_connection  = @saved_col_connection */ ;
/*!50003 SET @saved_cs_client      = @@character_set_client */ ;
/*!50003 SET @saved_cs_results     = @@character_set_results */ ;
/*!50003 SET @saved_col_connection = @@collation_connection */ ;
/*!50003 SET character_set_client  = utf8mb4 */ ;
/*!50003 SET character_set_results = utf8mb4 */ ;
/*!50003 SET collation_connection  = utf8mb4_general_ci */ ;
/*!50003 SET @saved_sql_mode       = @@sql_mode */ ;
/*!50003 SET sql_mode              = 'ONLY_FULL_GROUP_BY,STRICT_TRANS_TABLES,NO_ZERO_IN_DATE,NO_ZERO_DATE,ERROR_FOR_DIVISION_BY_ZERO,NO_ENGINE_SUBSTITUTION' */ ;
DELIMITER ;;
/*!50003 CREATE TRIGGER `after_update_platforms` AFTER UPDATE ON `platforms` FOR EACH ROW BEGIN
    IF OLD.`priority` != NEW.`priority` OR NOT OLD.`regexp` <=> NEW.`regexp` THEN
        UPDATE `items`
            LEFT JOIN `platforms` AS `old_platform` ON `old_platform`.`id` = `items`.`platform_id`
        SET `items`.`platform_id` = (
            SELECT `platforms`.`id` FROM `platforms`
            WHERE `items`.`url` REGEXP `platforms`.`regexp`
            ORDER BY `platforms`.`priority` DESC
            LIMIT 1
        )
        WHERE `old_platform`.`priority` < NEW.`priority` OR
              (`items`.`url` IS NOT NULL AND `old_platform`.`id` IS NULL) OR
              `old_platform`.`id` = NEW.`id`;
    END IF;
END */;;
DELIMITER ;
/*!50003 SET sql_mode              = @saved_sql_mode */ ;
/*!50003 SET character_set_client  = @saved_cs_client */ ;
/*!50003 SET character_set_results = @saved_cs_results */ ;
/*!50003 SET collation_connection  = @saved_col_connection */ ;

--
-- Table structure for table `results`
--

DROP TABLE IF EXISTS `results`;
/*!40101 SET @saved_cs_client     = @@character_set_client */;
/*!50503 SET character_set_client = utf8mb4 */;
CREATE TABLE `results` (
  `participant_id` bigint NOT NULL,
  `attempt_id` bigint NOT NULL DEFAULT '0',
  `item_id` bigint NOT NULL,
  `score_computed` float NOT NULL DEFAULT '0' COMMENT 'Score computed from the best answer or by propagation, with score_edit_rule applied',
  `score_edit_rule` enum('set','diff') DEFAULT NULL COMMENT 'Whether the edit value replaces and adds up to the score of the best answer',
  `score_edit_value` float DEFAULT NULL COMMENT 'Score which overrides or adds up (depending on score_edit_rule) to the score obtained from best answer or propagation',
  `score_edit_comment` varchar(200) DEFAULT NULL COMMENT 'Explanation of the value set in score_edit_value',
  `submissions` int NOT NULL DEFAULT '0' COMMENT 'Number of submissions. Only for tasks, not propagated',
  `tasks_tried` int NOT NULL DEFAULT '0' COMMENT 'Number of tasks which have been attempted among this item''s descendants (at least one submission), within this attempt',
  `started` tinyint(1) GENERATED ALWAYS AS ((`started_at` is not null)) VIRTUAL NOT NULL COMMENT 'Auto-generated from `started_at`',
  `validated` tinyint(1) GENERATED ALWAYS AS ((`validated_at` is not null)) VIRTUAL NOT NULL COMMENT 'Auto-generated from `validated_at`',
  `tasks_with_help` int NOT NULL DEFAULT '0' COMMENT 'Number of this item''s descendants tasks within this attempts for which the user asked for hints (or help on the forum - not implemented)',
  `hints_requested` mediumtext COMMENT 'JSON array of the hints that have been requested for this attempt',
  `hints_cached` int NOT NULL DEFAULT '0' COMMENT 'Number of hints which have been requested for this attempt',
  `started_at` datetime DEFAULT NULL COMMENT 'Time at which the attempt was manually created or was first marked as started (should be when it is first visited). Not propagated',
  `validated_at` datetime DEFAULT NULL COMMENT 'Submission time of the first answer that made the attempt validated',
  `latest_activity_at` datetime NOT NULL DEFAULT CURRENT_TIMESTAMP COMMENT 'Time of the latest activity (attempt creation, submission, hint request) of a user on this attempt or its children',
  `score_obtained_at` datetime DEFAULT NULL COMMENT 'Submission time of the first answer which led to the best score',
  `latest_submission_at` datetime DEFAULT NULL COMMENT 'Time of the latest submission. Only for tasks, not propagated',
  `latest_hint_at` datetime DEFAULT NULL COMMENT 'Time of the last request for a hint. Only for tasks, not propagated',
  `help_requested` tinyint(1) NOT NULL DEFAULT '0' COMMENT 'Whether the participant is requesting help on the item in this attempt',
  `recomputing_state` enum('recomputing','modified','unchanged') NOT NULL DEFAULT 'unchanged' COMMENT 'State of the result, used during recomputing',
  PRIMARY KEY (`participant_id`,`attempt_id`,`item_id`),
  KEY `item_id` (`item_id`),
  KEY `participant_id_item_id` (`participant_id`,`item_id`),
  KEY `participant_id_item_id_score_desc_score_obtained_at` (`participant_id`,`item_id`,`score_computed` DESC,`score_obtained_at`),
  KEY `started_at_desc_item_id_participant_id_attempt_id_desc` (`started_at` DESC,`item_id`,`participant_id`,`attempt_id` DESC),
  KEY `validated_at_desc_item_id_participant_id_attempt_id_desc` (`validated_at` DESC,`item_id`,`participant_id`,`attempt_id` DESC),
  KEY `participant_id_item_id_latest_activity_at_desc` (`participant_id`,`item_id`,`latest_activity_at` DESC),
  KEY `participant_id_item_id_started_started_at` (`participant_id`,`item_id`,`started`,`started_at`),
  KEY `participant_id_item_id_validated_validated_at` (`participant_id`,`item_id`,`validated`,`validated_at`),
  CONSTRAINT `fk_results_item_id_items_id` FOREIGN KEY (`item_id`) REFERENCES `items` (`id`) ON DELETE CASCADE,
  CONSTRAINT `fk_results_participant_id_attempt_id_attempts_participant_id_id` FOREIGN KEY (`participant_id`, `attempt_id`) REFERENCES `attempts` (`participant_id`, `id`) ON DELETE CASCADE,
  CONSTRAINT `cs_results_score_computed_is_valid` CHECK ((`score_computed` between 0 and 100)),
  CONSTRAINT `cs_results_score_edit_value_is_valid` CHECK ((ifnull(`score_edit_value`,0) between -(100) and 100))
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_0900_ai_ci COMMENT='Attempts of a group (team or user) to solve a task once or several times with different parameters with the task allows it.';
/*!40101 SET character_set_client = @saved_cs_client */;

--
-- Dumping data for table `results`
--

LOCK TABLES `results` WRITE;
/*!40000 ALTER TABLE `results` DISABLE KEYS */;
/*!40000 ALTER TABLE `results` ENABLE KEYS */;
UNLOCK TABLES;
/*!50003 SET @saved_cs_client      = @@character_set_client */ ;
/*!50003 SET @saved_cs_results     = @@character_set_results */ ;
/*!50003 SET @saved_col_connection = @@collation_connection */ ;
/*!50003 SET character_set_client  = utf8mb4 */ ;
/*!50003 SET character_set_results = utf8mb4 */ ;
/*!50003 SET collation_connection  = utf8mb4_general_ci */ ;
/*!50003 SET @saved_sql_mode       = @@sql_mode */ ;
/*!50003 SET sql_mode              = 'ONLY_FULL_GROUP_BY,STRICT_TRANS_TABLES,NO_ZERO_IN_DATE,NO_ZERO_DATE,ERROR_FOR_DIVISION_BY_ZERO,NO_ENGINE_SUBSTITUTION' */ ;
DELIMITER ;;
/*!50003 CREATE TRIGGER `before_update_results` BEFORE UPDATE ON `results` FOR EACH ROW BEGIN
  IF NEW.recomputing_state = 'recomputing' THEN
    SET NEW.recomputing_state = IF(
      NEW.latest_activity_at <=> OLD.latest_activity_at AND
      NEW.tasks_tried <=> OLD.tasks_tried AND
      NEW.tasks_with_help <=> OLD.tasks_with_help AND
      NEW.validated_at <=> OLD.validated_at AND
      NEW.score_computed <=> OLD.score_computed AND
      -- We always consider results with the default latest_activity_at as changed
      -- because they look like a newly inserted result for a chapter/skill.
      -- This way we make sure that a newly inserted result is propagated.
      NEW.latest_activity_at <> '1000-01-01 00:00:00',
      'unchanged',
      'modified');
  END IF;
END */;;
DELIMITER ;
/*!50003 SET sql_mode              = @saved_sql_mode */ ;
/*!50003 SET character_set_client  = @saved_cs_client */ ;
/*!50003 SET character_set_results = @saved_cs_results */ ;
/*!50003 SET collation_connection  = @saved_col_connection */ ;

--
-- Table structure for table `results_propagate`
--

DROP TABLE IF EXISTS `results_propagate`;
/*!40101 SET @saved_cs_client     = @@character_set_client */;
/*!50503 SET character_set_client = utf8mb4 */;
CREATE TABLE `results_propagate` (
  `participant_id` bigint NOT NULL,
  `attempt_id` bigint NOT NULL DEFAULT '0',
  `item_id` bigint NOT NULL,
  `state` enum('to_be_propagated','to_be_recomputed','propagating','recomputing') NOT NULL COMMENT '"to_be_propagated" means that ancestors should be recomputed',
  PRIMARY KEY (`participant_id`,`attempt_id`,`item_id`),
  KEY `state` (`state`),
  CONSTRAINT `fk_results_propagate_to_results` FOREIGN KEY (`participant_id`, `attempt_id`, `item_id`) REFERENCES `results` (`participant_id`, `attempt_id`, `item_id`) ON DELETE CASCADE
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_0900_ai_ci COMMENT='Used by the algorithm that computes results for items that have children and unlocks items if needed.';
/*!40101 SET character_set_client = @saved_cs_client */;

--
-- Dumping data for table `results_propagate`
--

LOCK TABLES `results_propagate` WRITE;
/*!40000 ALTER TABLE `results_propagate` DISABLE KEYS */;
/*!40000 ALTER TABLE `results_propagate` ENABLE KEYS */;
UNLOCK TABLES;

--
-- Table structure for table `results_propagate_sync`
--

DROP TABLE IF EXISTS `results_propagate_sync`;
/*!40101 SET @saved_cs_client     = @@character_set_client */;
/*!50503 SET character_set_client = utf8mb4 */;
CREATE TABLE `results_propagate_sync` (
  `connection_id` bigint unsigned NOT NULL,
  `participant_id` bigint NOT NULL,
  `attempt_id` bigint NOT NULL DEFAULT '0',
  `item_id` bigint NOT NULL,
  `state` enum('to_be_propagated','to_be_recomputed','propagating','recomputing') NOT NULL COMMENT '"to_be_propagated" means that ancestors should be recomputed',
  PRIMARY KEY (`connection_id`,`participant_id`,`attempt_id`,`item_id`),
  KEY `connection_id_state` (`connection_id`,`state`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_0900_ai_ci COMMENT='Used by the algorithm that computes results for items that have children and unlocks items if needed.';
/*!40101 SET character_set_client = @saved_cs_client */;

--
-- Dumping data for table `results_propagate_sync`
--

LOCK TABLES `results_propagate_sync` WRITE;
/*!40000 ALTER TABLE `results_propagate_sync` DISABLE KEYS */;
/*!40000 ALTER TABLE `results_propagate_sync` ENABLE KEYS */;
UNLOCK TABLES;

--
-- Temporary view structure for view `results_propagate_sync_conn`
--

DROP TABLE IF EXISTS `results_propagate_sync_conn`;
/*!50001 DROP VIEW IF EXISTS `results_propagate_sync_conn`*/;
SET @saved_cs_client     = @@character_set_client;
/*!50503 SET character_set_client = utf8mb4 */;
/*!50001 CREATE VIEW `results_propagate_sync_conn` AS SELECT
 1 AS `connection_id`,
 1 AS `participant_id`,
 1 AS `attempt_id`,
 1 AS `item_id`,
 1 AS `state`*/;
SET character_set_client = @saved_cs_client;

--
-- Table structure for table `results_recompute_for_items`
--

DROP TABLE IF EXISTS `results_recompute_for_items`;
/*!40101 SET @saved_cs_client     = @@character_set_client */;
/*!50503 SET character_set_client = utf8mb4 */;
CREATE TABLE `results_recompute_for_items` (
  `item_id` bigint NOT NULL,
  `is_being_processed` tinyint(1) NOT NULL DEFAULT '0',
  PRIMARY KEY (`item_id`),
  KEY `is_being_processed` (`is_being_processed`),
  CONSTRAINT `fk_results_propagate_items_to_items` FOREIGN KEY (`item_id`) REFERENCES `items` (`id`) ON DELETE CASCADE
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_0900_ai_ci COMMENT='Used by the algorithm that computes results. All results for the item_id have to be recomputed when the item_id is in this table.';
/*!40101 SET character_set_client = @saved_cs_client */;

--
-- Dumping data for table `results_recompute_for_items`
--

LOCK TABLES `results_recompute_for_items` WRITE;
/*!40000 ALTER TABLE `results_recompute_for_items` DISABLE KEYS */;
/*!40000 ALTER TABLE `results_recompute_for_items` ENABLE KEYS */;
UNLOCK TABLES;

--
-- Table structure for table `sessions`
--

DROP TABLE IF EXISTS `sessions`;
/*!40101 SET @saved_cs_client     = @@character_set_client */;
/*!50503 SET character_set_client = utf8mb4 */;
CREATE TABLE `sessions` (
  `session_id` bigint NOT NULL,
  `user_id` bigint NOT NULL,
  `refresh_token` varbinary(2000) DEFAULT NULL COMMENT 'Refresh tokens (unlimited lifetime) used by the backend to request fresh access tokens from the auth module',
  PRIMARY KEY (`session_id`),
  KEY `fk_sessions_users_user_id_group_id` (`user_id`),
  KEY `refresh_token` (`refresh_token`(767)),
  CONSTRAINT `fk_sessions_users_user_id_group_id` FOREIGN KEY (`user_id`) REFERENCES `users` (`group_id`) ON DELETE CASCADE
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_0900_ai_ci COMMENT='Sessions represent a logged in user, on a specific device.';
/*!40101 SET character_set_client = @saved_cs_client */;

--
-- Dumping data for table `sessions`
--

LOCK TABLES `sessions` WRITE;
/*!40000 ALTER TABLE `sessions` DISABLE KEYS */;
/*!40000 ALTER TABLE `sessions` ENABLE KEYS */;
UNLOCK TABLES;

--
-- Table structure for table `stopwords`
--

DROP TABLE IF EXISTS `stopwords`;
/*!40101 SET @saved_cs_client     = @@character_set_client */;
/*!50503 SET character_set_client = utf8mb4 */;
CREATE TABLE `stopwords` (
  `value` varchar(30) DEFAULT NULL
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_0900_ai_ci COMMENT='Stopwords for the fulltext search. The table is empty on purpose. It is used to prevent the fulltext search from using the default stopwords list.\nAll the MySQL fulltext indexes would have to be recreated with innodb_ft_server_stopword_table pointing to this table if we wanted to add some new stopwords.\nAlso the stopwords would have to be filtered out from searched strings in the application code.';
/*!40101 SET character_set_client = @saved_cs_client */;

--
-- Dumping data for table `stopwords`
--

LOCK TABLES `stopwords` WRITE;
/*!40000 ALTER TABLE `stopwords` DISABLE KEYS */;
/*!40000 ALTER TABLE `stopwords` ENABLE KEYS */;
UNLOCK TABLES;

--
-- Table structure for table `threads`
--

DROP TABLE IF EXISTS `threads`;
/*!40101 SET @saved_cs_client     = @@character_set_client */;
/*!50503 SET character_set_client = utf8mb4 */;
CREATE TABLE `threads` (
  `participant_id` bigint NOT NULL,
  `item_id` bigint NOT NULL,
  `status` enum('waiting_for_participant','waiting_for_trainer','closed') NOT NULL,
  `helper_group_id` bigint NOT NULL,
  `latest_update_at` datetime NOT NULL DEFAULT CURRENT_TIMESTAMP COMMENT 'Last time a message was posted or the status was updated.',
  `message_count` int NOT NULL DEFAULT '0' COMMENT 'Approximation of the number of message sent on the thread.',
  PRIMARY KEY (`participant_id`,`item_id`),
  KEY `fk_threads_item_id_items_id` (`item_id`),
  KEY `fk_threads_helper_group_id_groups_id` (`helper_group_id`),
  CONSTRAINT `fk_threads_helper_group_id_groups_id` FOREIGN KEY (`helper_group_id`) REFERENCES `groups` (`id`),
  CONSTRAINT `fk_threads_item_id_items_id` FOREIGN KEY (`item_id`) REFERENCES `items` (`id`) ON DELETE CASCADE,
  CONSTRAINT `fk_threads_participant_id_groups_id` FOREIGN KEY (`participant_id`) REFERENCES `groups` (`id`) ON DELETE CASCADE
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_0900_ai_ci COMMENT='Discussion thread related to participant-item pair.';
/*!40101 SET character_set_client = @saved_cs_client */;

--
-- Dumping data for table `threads`
--

LOCK TABLES `threads` WRITE;
/*!40000 ALTER TABLE `threads` DISABLE KEYS */;
/*!40000 ALTER TABLE `threads` ENABLE KEYS */;
UNLOCK TABLES;

--
-- Table structure for table `user_batch_prefixes`
--

DROP TABLE IF EXISTS `user_batch_prefixes`;
/*!40101 SET @saved_cs_client     = @@character_set_client */;
/*!50503 SET character_set_client = utf8mb4 */;
CREATE TABLE `user_batch_prefixes` (
  `group_prefix` varchar(13) CHARACTER SET utf8mb4 COLLATE utf8mb4_bin NOT NULL COMMENT 'Prefix used in front of all batches',
  `group_id` bigint DEFAULT NULL COMMENT 'Group (and its descendants) in which managers can create users in batch. NULL if the group was deleted, in which case batches should be cleaned manually.',
  `max_users` mediumint unsigned NOT NULL DEFAULT '1000' COMMENT 'Maximum number of users that can be created under this prefix',
  `allow_new` tinyint(1) NOT NULL DEFAULT '1' COMMENT 'Whether this prefix can be used for new user batches',
  PRIMARY KEY (`group_prefix`),
  KEY `fk_user_batch_prefixes_group_id_groups_id` (`group_id`),
  CONSTRAINT `fk_user_batch_prefixes_group_id_groups_id` FOREIGN KEY (`group_id`) REFERENCES `groups` (`id`) ON DELETE SET NULL
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_0900_ai_ci COMMENT='Authorized login prefixes for user batch creation. A prefix cannot be deleted without deleting batches using it.';
/*!40101 SET character_set_client = @saved_cs_client */;

--
-- Dumping data for table `user_batch_prefixes`
--

LOCK TABLES `user_batch_prefixes` WRITE;
/*!40000 ALTER TABLE `user_batch_prefixes` DISABLE KEYS */;
/*!40000 ALTER TABLE `user_batch_prefixes` ENABLE KEYS */;
UNLOCK TABLES;

--
-- Table structure for table `user_batches_v2`
--

DROP TABLE IF EXISTS `user_batches_v2`;
/*!40101 SET @saved_cs_client     = @@character_set_client */;
/*!50503 SET character_set_client = utf8mb4 */;
CREATE TABLE `user_batches_v2` (
  `group_prefix` varchar(13) CHARACTER SET utf8mb4 COLLATE utf8mb4_bin NOT NULL COMMENT 'Authorized (first) part of the full login prefix',
  `custom_prefix` varchar(14) CHARACTER SET utf8mb4 COLLATE utf8mb4_bin NOT NULL COMMENT 'Second part of the full login prefix, given by the user that created the batch',
  `size` mediumint unsigned NOT NULL COMMENT 'Number of users created in this batch',
  `creator_id` bigint DEFAULT NULL,
  `created_at` datetime NOT NULL DEFAULT CURRENT_TIMESTAMP,
  PRIMARY KEY (`group_prefix`,`custom_prefix`),
  KEY `fk_user_batches_v2_creator_id_users_group_id` (`creator_id`),
  CONSTRAINT `fk_user_batches_v2_creator_id_users_group_id` FOREIGN KEY (`creator_id`) REFERENCES `users` (`group_id`) ON DELETE SET NULL,
  CONSTRAINT `fk_user_batches_v2_group_prefix_user_batch_prefixes_group_pref` FOREIGN KEY (`group_prefix`) REFERENCES `user_batch_prefixes` (`group_prefix`) ON DELETE RESTRICT,
  CONSTRAINT `ck_user_batches_v2_custom_prefix` CHECK (regexp_like(`custom_prefix`,_utf8mb4'^[a-z0-9-]+$'))
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_0900_ai_ci COMMENT='Batches of users that were created (replaces user_batches which has been broken by a MySQL update)';
/*!40101 SET character_set_client = @saved_cs_client */;

--
-- Dumping data for table `user_batches_v2`
--

LOCK TABLES `user_batches_v2` WRITE;
/*!40000 ALTER TABLE `user_batches_v2` DISABLE KEYS */;
/*!40000 ALTER TABLE `user_batches_v2` ENABLE KEYS */;
UNLOCK TABLES;

--
-- Table structure for table `users`
--

DROP TABLE IF EXISTS `users`;
/*!40101 SET @saved_cs_client     = @@character_set_client */;
/*!50503 SET character_set_client = utf8mb4 */;
CREATE TABLE `users` (
  `group_id` bigint NOT NULL COMMENT 'Group that represents this user',
  `login_id` bigint DEFAULT NULL COMMENT '"userId" returned by the auth platform',
  `temp_user` tinyint(1) NOT NULL DEFAULT '0' COMMENT 'Whether it is a temporary user. If so, the user will be deleted soon.',
  `login` varchar(100) CHARACTER SET utf8mb4 COLLATE utf8mb4_bin NOT NULL DEFAULT '' COMMENT 'login provided by the auth platform',
  `open_id_identity` varchar(255) DEFAULT NULL COMMENT 'User''s Open Id Identity',
  `password_md5` varchar(100) DEFAULT NULL,
  `salt` varchar(32) DEFAULT NULL,
  `recover` varchar(50) DEFAULT NULL,
  `registered_at` datetime DEFAULT NULL COMMENT 'When the user first connected to this platform',
  `email` varchar(100) DEFAULT NULL COMMENT 'E-mail, provided by auth platform',
  `email_verified` tinyint(1) NOT NULL DEFAULT '0' COMMENT 'Whether email has been verified, provided by auth platform',
  `first_name` varchar(100) DEFAULT NULL COMMENT 'First name, provided by auth platform',
  `last_name` varchar(100) DEFAULT NULL COMMENT 'Last name, provided by auth platform',
  `student_id` text COMMENT 'A student id provided by the school, provided by auth platform',
  `country_code` char(3) NOT NULL DEFAULT '' COMMENT '3-letter country code',
  `time_zone` varchar(100) DEFAULT NULL COMMENT 'Time zone, provided by auth platform',
  `birth_date` date DEFAULT NULL COMMENT 'Date of birth, provided by auth platform',
  `graduation_year` int NOT NULL DEFAULT '0' COMMENT 'High school graduation year',
  `grade` int DEFAULT NULL COMMENT 'School grade, provided by auth platform',
  `sex` enum('Male','Female') DEFAULT NULL COMMENT 'Gender, provided by auth platform',
  `address` mediumtext COMMENT 'Address, provided by auth platform',
  `zipcode` longtext COMMENT 'Zip code, provided by auth platform',
  `city` longtext COMMENT 'City, provided by auth platform',
  `land_line_number` longtext COMMENT 'Phone number, provided by auth platform',
  `cell_phone_number` longtext COMMENT 'Mobile phone number, provided by auth platform',
  `default_language` char(3) NOT NULL DEFAULT 'fr' COMMENT 'Current language used to display content. Initial version provided by auth platform, then can be changed manually.',
  `notify_news` tinyint NOT NULL DEFAULT '0' COMMENT 'Whether the user accepts that we send emails about events related to the platform',
  `notify` enum('Never','Answers','Concerned') NOT NULL DEFAULT 'Answers' COMMENT 'When we should send an email to the user. Answers: when someone posts a message on a thread created by the user. Concerned: when someone post a message on a thread that the user participated in',
  `public_first_name` tinyint NOT NULL DEFAULT '0' COMMENT 'Whether show user''s first name in his public profile',
  `public_last_name` tinyint NOT NULL DEFAULT '0' COMMENT 'Whether show user''s last name in his public profile',
  `free_text` mediumtext COMMENT 'Text provided by the user, to be displayed on his public profile',
  `web_site` varchar(100) DEFAULT NULL COMMENT 'Link to the user''s website, to be displayed on his public profile',
  `photo_autoload` tinyint(1) NOT NULL DEFAULT '0' COMMENT 'Indicates that the user has a picture associated with his profile. Not used yet.',
  `lang_prog` varchar(30) DEFAULT 'Python' COMMENT 'Current programming language selected by the user (to display the corresponding version of tasks)',
  `latest_login_at` datetime DEFAULT NULL COMMENT 'When is the last time this user logged in on the platform',
  `latest_activity_at` datetime DEFAULT NULL COMMENT 'Last activity time on the platform (any action)',
  `latest_profile_sync_at` datetime DEFAULT NULL COMMENT 'Last time when the profile was synced with the login module',
  `last_ip` varchar(16) DEFAULT NULL COMMENT 'Last IP (to detect cheaters).',
  `basic_editor_mode` tinyint NOT NULL DEFAULT '1' COMMENT 'Which editor should be used in programming tasks.',
  `spaces_for_tab` int NOT NULL DEFAULT '3' COMMENT 'How many spaces for a tabulation, in programming tasks.',
  `member_state` tinyint NOT NULL DEFAULT '0' COMMENT 'On old website, indicates if the user is a member of France-ioi',
  `step_level_in_site` int NOT NULL DEFAULT '0' COMMENT 'User''s level',
  `is_admin` tinyint NOT NULL DEFAULT '0' COMMENT 'Is the user an admin? Not used?',
  `no_ranking` tinyint NOT NULL DEFAULT '0' COMMENT 'Whether this user should not be listed when displaying the results of contests, or points obtained on the platform',
  `help_given` int NOT NULL DEFAULT '0' COMMENT 'How many times did the user help others (# of discussions)',
  `access_group_id` bigint DEFAULT NULL,
  `notifications_read_at` datetime DEFAULT NULL COMMENT 'When the user last read notifications',
  `creator_id` bigint DEFAULT NULL COMMENT 'User who created a given login with the login generation tool',
  PRIMARY KEY (`group_id`),
  UNIQUE KEY `login` (`login`),
  KEY `country_code` (`country_code`),
  KEY `lang_prog` (`lang_prog`),
  KEY `login_id` (`login_id`),
  KEY `temp_user` (`temp_user`),
  KEY `fk_users_creator_id_users_group_id` (`creator_id`),
  CONSTRAINT `fk_users_creator_id_users_group_id` FOREIGN KEY (`creator_id`) REFERENCES `users` (`group_id`) ON DELETE SET NULL,
  CONSTRAINT `fk_users_group_id_groups_id` FOREIGN KEY (`group_id`) REFERENCES `groups` (`id`) ON DELETE CASCADE
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_0900_ai_ci COMMENT='Users. A large part is obtained from the auth platform and may not be manually edited';
/*!40101 SET character_set_client = @saved_cs_client */;

--
-- Dumping data for table `users`
--

LOCK TABLES `users` WRITE;
/*!40000 ALTER TABLE `users` DISABLE KEYS */;
/*!40000 ALTER TABLE `users` ENABLE KEYS */;
UNLOCK TABLES;

--
-- Final view structure for view `groups_ancestors_active`
--

/*!50001 DROP VIEW IF EXISTS `groups_ancestors_active`*/;
/*!50001 SET @saved_cs_client          = @@character_set_client */;
/*!50001 SET @saved_cs_results         = @@character_set_results */;
/*!50001 SET @saved_col_connection     = @@collation_connection */;
/*!50001 SET character_set_client      = utf8mb4 */;
/*!50001 SET character_set_results     = utf8mb4 */;
/*!50001 SET collation_connection      = utf8mb4_general_ci */;
/*!50001 CREATE ALGORITHM=UNDEFINED */
/*!50001 VIEW `groups_ancestors_active` AS select `groups_ancestors`.`ancestor_group_id` AS `ancestor_group_id`,`groups_ancestors`.`child_group_id` AS `child_group_id`,`groups_ancestors`.`is_self` AS `is_self`,`groups_ancestors`.`expires_at` AS `expires_at` from `groups_ancestors` where (now() < `groups_ancestors`.`expires_at`) */;
/*!50001 SET character_set_client      = @saved_cs_client */;
/*!50001 SET character_set_results     = @saved_cs_results */;
/*!50001 SET collation_connection      = @saved_col_connection */;

--
-- Final view structure for view `groups_groups_active`
--

/*!50001 DROP VIEW IF EXISTS `groups_groups_active`*/;
/*!50001 SET @saved_cs_client          = @@character_set_client */;
/*!50001 SET @saved_cs_results         = @@character_set_results */;
/*!50001 SET @saved_col_connection     = @@collation_connection */;
/*!50001 SET character_set_client      = utf8mb4 */;
/*!50001 SET character_set_results     = utf8mb4 */;
/*!50001 SET collation_connection      = utf8mb4_general_ci */;
/*!50001 CREATE ALGORITHM=UNDEFINED */
/*!50001 VIEW `groups_groups_active` AS select `groups_groups`.`parent_group_id` AS `parent_group_id`,`groups_groups`.`child_group_id` AS `child_group_id`,`groups_groups`.`expires_at` AS `expires_at`,`groups_groups`.`is_team_membership` AS `is_team_membership`,`groups_groups`.`personal_info_view_approved_at` AS `personal_info_view_approved_at`,`groups_groups`.`personal_info_view_approved` AS `personal_info_view_approved`,`groups_groups`.`lock_membership_approved_at` AS `lock_membership_approved_at`,`groups_groups`.`lock_membership_approved` AS `lock_membership_approved`,`groups_groups`.`watch_approved_at` AS `watch_approved_at`,`groups_groups`.`watch_approved` AS `watch_approved` from `groups_groups` where (now() < `groups_groups`.`expires_at`) */;
/*!50001 SET character_set_client      = @saved_cs_client */;
/*!50001 SET character_set_results     = @saved_cs_results */;
/*!50001 SET collation_connection      = @saved_col_connection */;

--
-- Final view structure for view `permissions_propagate_sync_conn`
--

/*!50001 DROP VIEW IF EXISTS `permissions_propagate_sync_conn`*/;
/*!50001 SET @saved_cs_client          = @@character_set_client */;
/*!50001 SET @saved_cs_results         = @@character_set_results */;
/*!50001 SET @saved_col_connection     = @@collation_connection */;
/*!50001 SET character_set_client      = utf8mb4 */;
/*!50001 SET character_set_results     = utf8mb4 */;
/*!50001 SET collation_connection      = utf8mb4_general_ci */;
/*!50001 CREATE ALGORITHM=UNDEFINED */
/*!50001 VIEW `permissions_propagate_sync_conn` AS select `permissions_propagate_sync`.`connection_id` AS `connection_id`,`permissions_propagate_sync`.`group_id` AS `group_id`,`permissions_propagate_sync`.`item_id` AS `item_id`,`permissions_propagate_sync`.`propagate_to` AS `propagate_to` from `permissions_propagate_sync` where (`permissions_propagate_sync`.`connection_id` = connection_id()) */;
/*!50001 SET character_set_client      = @saved_cs_client */;
/*!50001 SET character_set_results     = @saved_cs_results */;
/*!50001 SET collation_connection      = @saved_col_connection */;

--
-- Final view structure for view `results_propagate_sync_conn`
--

/*!50001 DROP VIEW IF EXISTS `results_propagate_sync_conn`*/;
/*!50001 SET @saved_cs_client          = @@character_set_client */;
/*!50001 SET @saved_cs_results         = @@character_set_results */;
/*!50001 SET @saved_col_connection     = @@collation_connection */;
/*!50001 SET character_set_client      = utf8mb4 */;
/*!50001 SET character_set_results     = utf8mb4 */;
/*!50001 SET collation_connection      = utf8mb4_general_ci */;
/*!50001 CREATE ALGORITHM=UNDEFINED */
/*!50001 VIEW `results_propagate_sync_conn` AS select `results_propagate_sync`.`connection_id` AS `connection_id`,`results_propagate_sync`.`participant_id` AS `participant_id`,`results_propagate_sync`.`attempt_id` AS `attempt_id`,`results_propagate_sync`.`item_id` AS `item_id`,`results_propagate_sync`.`state` AS `state` from `results_propagate_sync` where (`results_propagate_sync`.`connection_id` = connection_id()) */;
/*!50001 SET character_set_client      = @saved_cs_client */;
/*!50001 SET character_set_results     = @saved_cs_results */;
/*!50001 SET collation_connection      = @saved_col_connection */;
/*!40103 SET TIME_ZONE=@OLD_TIME_ZONE */;

--
-- Recreate FULLTEXT indexes with empty stopwords list
--
SET @old_stopword_table := @@innodb_ft_user_stopword_table;
SET SESSION innodb_ft_user_stopword_table = CONCAT(DATABASE(), '/stopwords');

ALTER TABLE `items_strings` DROP INDEX `fullTextTitle`;
CREATE FULLTEXT INDEX `fullTextTitle` ON `items_strings`(`title`);
ALTER TABLE `groups` DROP INDEX `fullTextName`;
CREATE FULLTEXT INDEX `fullTextName` ON `groups`(`name`);

SET SESSION innodb_ft_user_stopword_table = @old_stopword_table;

/*!40101 SET SQL_MODE=@OLD_SQL_MODE */;
/*!40014 SET FOREIGN_KEY_CHECKS=@OLD_FOREIGN_KEY_CHECKS */;
/*!40014 SET UNIQUE_CHECKS=@OLD_UNIQUE_CHECKS */;
/*!40101 SET CHARACTER_SET_CLIENT=@OLD_CHARACTER_SET_CLIENT */;
/*!40101 SET CHARACTER_SET_RESULTS=@OLD_CHARACTER_SET_RESULTS */;
/*!40101 SET COLLATION_CONNECTION=@OLD_COLLATION_CONNECTION */;
/*!40111 SET SQL_NOTES=@OLD_SQL_NOTES */;

-- Dump completed on 2025-09-04 13:02:20
