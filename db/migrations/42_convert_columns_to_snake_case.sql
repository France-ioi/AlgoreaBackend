-- +migrate Up
ALTER TABLE `groups`
	MODIFY `sTextId` varchar(255) NOT NULL DEFAULT '' COMMENT 'Internal text id for special groups. Used to refer o them and avoid breaking features if an admin renames the group',
	MODIFY `iTeamParticipating` tinyint(1) NOT NULL DEFAULT '0' COMMENT 'Did the team start the item it is associated to (from team_item_id)?',
	MODIFY `bOpenContest` tinyint(1) NOT NULL DEFAULT '0' COMMENT 'If true and the group is associated through redirect_path with an item that is a contest, the contest should be started for this user as soon as he joins the group.';
ALTER TABLE `groups_ancestors`
	MODIFY `bIsSelf` tinyint(1) NOT NULL DEFAULT '0' COMMENT 'Whether ancestor_group_id = child_group_id.';
ALTER TABLE `groups_attempts`
	MODIFY `bKeyObtained` tinyint(1) NOT NULL DEFAULT '0' COMMENT 'Whether the user obtained the key on this item (changed to 1 if the user gets a score >= items.score_min_unlock, will grant access to new items from items.unlocked_item_ids). This information is propagated to users_items.';
ALTER TABLE `items`
	MODIFY `bFixedRanks` tinyint(1) NOT NULL DEFAULT '0' COMMENT 'If true, prevents users from changing the order of the children by drag&drop and auto-calculation of the order of children. Allows for manual setting of the order, for instance in cases where we want to have multiple items with the same order (check items_items.child_order).',
	MODIFY `iScoreMinUnlock` int(11) NOT NULL DEFAULT '100' COMMENT 'Minimum score to obtain so that the item, indicated by "unlocked_item_ids", is actually unlocked',
	MODIFY `sTeamMode` enum('All','Half','One','None') DEFAULT NULL COMMENT 'If qualified_group_id is not NULL, this field specifies how many team members need to belong to that group in order for the whole team to be qualified and able to start the item.',
	MODIFY `idTeamInGroup` bigint(20) DEFAULT NULL COMMENT 'group id in which "qualified" users will belong. team_mode dictates how many of a team''s members must be "qualified" in order to start the item.';
ALTER TABLE `items_items`
	MODIFY `iChildOrder` int(11) NOT NULL COMMENT 'Position, relative to its siblings, when displaying all the children of the parent. If multiple items have the same child_order, they will be sorted in a random way, specific to each user (a user will always see the items in the same order).';
ALTER TABLE `users_items`
	MODIFY `bKeyObtained` tinyint(1) NOT NULL DEFAULT '0' COMMENT 'Whether the user obtained the key on this item. Changed to 1 if the user gets a score >= items.score_min_unlock, will grant access to new item from items.unlocked_item_ids. This information is propagated to users_items.';


ALTER TABLE `badges`
	RENAME COLUMN `ID` TO `id`,
	RENAME COLUMN `idUser` TO `user_id`;
ALTER TABLE `filters`
	RENAME COLUMN `ID` TO `id`,
	RENAME COLUMN `idUser` TO `user_id`,
	RENAME COLUMN `sName` TO `name`,
	RENAME COLUMN `bSelected` TO `selected`,
	RENAME COLUMN `bStarred` TO `starred`,
	RENAME COLUMN `sStartDate` TO `start_date`,
	RENAME COLUMN `sEndDate` TO `end_date`,
	RENAME COLUMN `bArchived` TO `archived`,
	RENAME COLUMN `bParticipated` TO `participated`,
	RENAME COLUMN `bUnread` TO `unread`,
	RENAME COLUMN `idItem` TO `item_id`,
	RENAME COLUMN `idGroup` TO `group_id`,
	RENAME COLUMN `olderThan` TO `older_than`,
	RENAME COLUMN `newerThan` TO `newer_than`,
	RENAME COLUMN `sUsersSearch` TO `users_search`,
	RENAME COLUMN `sBodySearch` TO `body_search`,
	RENAME COLUMN `bImportant` TO `important`,
	RENAME COLUMN `iVersion` TO `version`,
	RENAME INDEX `iVersion` TO `version`;
ALTER TABLE `groups`
	RENAME COLUMN `ID` TO `id`,
	RENAME COLUMN `sName` TO `name`,
	RENAME COLUMN `sTextId` TO `text_id`,
	RENAME COLUMN `iGrade` TO `grade`,
	RENAME COLUMN `sGradeDetails` TO `grade_details`,
	RENAME COLUMN `sDescription` TO `description`,
	RENAME COLUMN `sDateCreated` TO `date_created`,
	RENAME COLUMN `bOpened` TO `opened`,
	RENAME COLUMN `bFreeAccess` TO `free_access`,
	RENAME COLUMN `idTeamItem` TO `team_item_id`,
	RENAME COLUMN `iTeamParticipating` TO `team_participating`,
	RENAME COLUMN `sCode` TO `code`,
	RENAME COLUMN `sCodeTimer` TO `code_timer`,
	RENAME COLUMN `sCodeEnd` TO `code_end`,
	RENAME COLUMN `sRedirectPath` TO `redirect_path`,
	RENAME COLUMN `bOpenContest` TO `open_contest`,
	RENAME COLUMN `sType` TO `type`,
	RENAME COLUMN `bSendEmails` TO `send_emails`,
	RENAME COLUMN `iVersion` TO `version`,
	RENAME COLUMN `lockUserDeletionDate` TO `lock_user_deletion_date`,
	RENAME INDEX `sPassword` TO `password`,
	RENAME INDEX `iVersion` TO `version`,
	RENAME INDEX `sType` TO `type`,
	RENAME INDEX `sName` TO `name`,
	RENAME INDEX `TypeName` TO `type_name`;
ALTER TABLE `groups_ancestors`
	RENAME COLUMN `ID` TO `id`,
	RENAME COLUMN `idGroupAncestor` TO `ancestor_group_id`,
	RENAME COLUMN `idGroupChild` TO `child_group_id`,
	RENAME COLUMN `bIsSelf` TO `is_self`,
	RENAME COLUMN `iVersion` TO `version`,
	RENAME INDEX `idGroupAncestor` TO `ancestor_group_id`;
ALTER TABLE `groups_attempts`
	RENAME COLUMN `ID` TO `id`,
	RENAME COLUMN `idGroup` TO `group_id`,
	RENAME COLUMN `idItem` TO `item_id`,
	RENAME COLUMN `idUserCreator` TO `creator_user_id`,
	DROP CHECK `cs_attempts_order`,
	RENAME COLUMN `iOrder` TO `order`,
	ADD CONSTRAINT `cs_attempts_order` CHECK (`order` > 0),
	RENAME COLUMN `iScore` TO `score`,
	RENAME COLUMN `iScoreComputed` TO `score_computed`,
	RENAME COLUMN `iScoreReeval` TO `score_reeval`,
	RENAME COLUMN `iScoreDiffManual` TO `score_diff_manual`,
	RENAME COLUMN `sScoreDiffComment` TO `score_diff_comment`,
	RENAME COLUMN `nbSubmissionsAttempts` TO `submissions_attempts`,
	RENAME COLUMN `nbTasksTried` TO `tasks_tried`,
	RENAME COLUMN `nbTasksSolved` TO `tasks_solved`,
	RENAME COLUMN `nbChildrenValidated` TO `children_validated`,
	RENAME COLUMN `bValidated` TO `validated`,
	RENAME COLUMN `bFinished` TO `finished`,
	RENAME COLUMN `bKeyObtained` TO `key_obtained`,
	RENAME COLUMN `nbTasksWithHelp` TO `tasks_with_help`,
	RENAME COLUMN `sHintsRequested` TO `hints_requested`,
	RENAME COLUMN `nbHintsCached` TO `hints_cached`,
	RENAME COLUMN `nbCorrectionsRead` TO `corrections_read`,
	RENAME COLUMN `iPrecision` TO `precision`,
	RENAME COLUMN `iAutonomy` TO `autonomy`,
	RENAME COLUMN `sStartDate` TO `start_date`,
	RENAME COLUMN `sValidationDate` TO `validation_date`,
	RENAME COLUMN `sFinishDate` TO `finish_date`,
	RENAME COLUMN `sLastActivityDate` TO `last_activity_date`,
	RENAME COLUMN `sThreadStartDate` TO `thread_start_date`,
	RENAME COLUMN `sBestAnswerDate` TO `best_answer_date`,
	RENAME COLUMN `sLastAnswerDate` TO `last_answer_date`,
	RENAME COLUMN `sLastHintDate` TO `last_hint_date`,
	RENAME COLUMN `sContestStartDate` TO `contest_start_date`,
	RENAME COLUMN `bRanked` TO `ranked`,
	RENAME COLUMN `sAllLangProg` TO `all_lang_prog`,
	RENAME COLUMN `iVersion` TO `version`,
	RENAME COLUMN `sAncestorsComputationState` TO `ancestors_computation_state`,
	RENAME COLUMN `iMinusScore` TO `minus_score`,
	RENAME INDEX `iVersion` TO `version`,
	RENAME INDEX `sAncestorsComputationState` TO `ancestors_computation_state`,
	RENAME INDEX `idItem` TO `item_id`,
	RENAME INDEX `GroupItem` TO `group_item`,
	RENAME INDEX `GroupItemMinusScoreBestAnswerDateID` TO `group_item_minus_score_best_answer_date_id`;
ALTER TABLE `groups_groups`
	RENAME COLUMN `ID` TO `id`,
	RENAME COLUMN `idGroupParent` TO `parent_group_id`,
	RENAME COLUMN `idGroupChild` TO `child_group_id`,
	RENAME COLUMN `iChildOrder` TO `child_order`,
	RENAME COLUMN `sType` TO `type`,
	RENAME COLUMN `sRole` TO `role`,
	RENAME COLUMN `idUserInviting` TO `inviting_user_id`,
	RENAME COLUMN `sStatusDate` TO `status_date`,
	RENAME COLUMN `iVersion` TO `version`,
	RENAME INDEX `iVersion` TO `version`,
	RENAME INDEX `idGroupChild` TO `child_group_id`,
	RENAME INDEX `idGroupParent` TO `parent_group_id`,
	RENAME INDEX `ParentOrder` TO `parent_order`;
ALTER TABLE `groups_items`
	RENAME COLUMN `ID` TO `id`,
	RENAME COLUMN `idGroup` TO `group_id`,
	RENAME COLUMN `idItem` TO `item_id`,
	RENAME COLUMN `idUserCreated` TO `creator_user_id`,
	RENAME COLUMN `sPartialAccessDate` TO `partial_access_date`,
	RENAME COLUMN `sAccessReason` TO `access_reason`,
	RENAME COLUMN `sFullAccessDate` TO `full_access_date`,
	RENAME COLUMN `sAccessSolutionsDate` TO `access_solutions_date`,
	RENAME COLUMN `bOwnerAccess` TO `owner_access`,
	RENAME COLUMN `bManagerAccess` TO `manager_access`,
	RENAME COLUMN `sCachedFullAccessDate` TO `cached_full_access_date`,
	RENAME COLUMN `sCachedPartialAccessDate` TO `cached_partial_access_date`,
	RENAME COLUMN `sCachedAccessSolutionsDate` TO `cached_access_solutions_date`,
	RENAME COLUMN `sCachedGrayedAccessDate` TO `cached_grayed_access_date`,
	RENAME COLUMN `sCachedAccessReason` TO `cached_access_reason`,
	RENAME COLUMN `bCachedFullAccess` TO `cached_full_access`,
	RENAME COLUMN `bCachedPartialAccess` TO `cached_partial_access`,
	RENAME COLUMN `bCachedAccessSolutions` TO `cached_access_solutions`,
	RENAME COLUMN `bCachedGrayedAccess` TO `cached_grayed_access`,
	RENAME COLUMN `bCachedManagerAccess` TO `cached_manager_access`,
	RENAME COLUMN `sPropagateAccess` TO `propagate_access`,
	RENAME COLUMN `sAdditionalTime` TO `additional_time`,
	RENAME COLUMN `iVersion` TO `version`,
	RENAME INDEX `idItem` TO `item_id`,
	RENAME INDEX `iVersion` TO `version`,
	RENAME INDEX `idGroup` TO `group_id`,
	RENAME INDEX `idItemtem` TO `itemtem_id`,
	RENAME INDEX `fullAccess` TO `full_access`,
	RENAME INDEX `accessSolutions` TO `access_solutions`,
	RENAME INDEX `sPropagateAccess` TO `propagate_access`,
	RENAME INDEX `partialAccess` TO `partial_access`;
ALTER TABLE `groups_items_propagate`
	RENAME COLUMN `ID` TO `id`,
	RENAME COLUMN `sPropagateAccess` TO `propagate_access`,
	RENAME INDEX `sPropagateAccess` TO `propagate_access`;
ALTER TABLE `groups_login_prefixes`
	RENAME COLUMN `ID` TO `id`,
	RENAME COLUMN `idGroup` TO `group_id`,
	RENAME COLUMN `iVersion` TO `version`,
	RENAME INDEX `idGroup` TO `group_id`;
ALTER TABLE `groups_propagate`
	RENAME COLUMN `ID` TO `id`,
	RENAME COLUMN `sAncestorsComputationState` TO `ancestors_computation_state`,
	RENAME INDEX `sAncestorsComputationState` TO `ancestors_computation_state`;
ALTER TABLE `history_filters`
	RENAME COLUMN `historyID` TO `history_id`,
	RENAME COLUMN `ID` TO `id`,
	RENAME COLUMN `idUser` TO `user_id`,
	RENAME COLUMN `sName` TO `name`,
	RENAME COLUMN `bSelected` TO `selected`,
	RENAME COLUMN `bStarred` TO `starred`,
	RENAME COLUMN `sStartDate` TO `start_date`,
	RENAME COLUMN `sEndDate` TO `end_date`,
	RENAME COLUMN `bArchived` TO `archived`,
	RENAME COLUMN `bParticipated` TO `participated`,
	RENAME COLUMN `bUnread` TO `unread`,
	RENAME COLUMN `idItem` TO `item_id`,
	RENAME COLUMN `idGroup` TO `group_id`,
	RENAME COLUMN `olderThan` TO `older_than`,
	RENAME COLUMN `newerThan` TO `newer_than`,
	RENAME COLUMN `sUsersSearch` TO `users_search`,
	RENAME COLUMN `sBodySearch` TO `body_search`,
	RENAME COLUMN `bImportant` TO `important`,
	RENAME COLUMN `iVersion` TO `version`,
	RENAME COLUMN `iNextVersion` TO `next_version`,
	RENAME COLUMN `bDeleted` TO `deleted`,
	RENAME INDEX `iVersion` TO `version`,
	RENAME INDEX `iNextVersion` TO `next_version`,
	RENAME INDEX `bDeleted` TO `deleted`,
	RENAME INDEX `ID` TO `id`;
ALTER TABLE `history_groups`
	RENAME COLUMN `historyID` TO `history_id`,
	RENAME COLUMN `ID` TO `id`,
	RENAME COLUMN `sName` TO `name`,
	RENAME COLUMN `iGrade` TO `grade`,
	RENAME COLUMN `sGradeDetails` TO `grade_details`,
	RENAME COLUMN `sDescription` TO `description`,
	RENAME COLUMN `sDateCreated` TO `date_created`,
	RENAME COLUMN `bOpened` TO `opened`,
	RENAME COLUMN `bFreeAccess` TO `free_access`,
	RENAME COLUMN `idTeamItem` TO `team_item_id`,
	RENAME COLUMN `iTeamParticipating` TO `team_participating`,
	RENAME COLUMN `sCode` TO `code`,
	RENAME COLUMN `sCodeTimer` TO `code_timer`,
	RENAME COLUMN `sCodeEnd` TO `code_end`,
	RENAME COLUMN `sRedirectPath` TO `redirect_path`,
	RENAME COLUMN `bOpenContest` TO `open_contest`,
	RENAME COLUMN `sType` TO `type`,
	RENAME COLUMN `bSendEmails` TO `send_emails`,
	RENAME COLUMN `bAncestorsComputed` TO `ancestors_computed`,
	RENAME COLUMN `sAncestorsComputationState` TO `ancestors_computation_state`,
	RENAME COLUMN `iVersion` TO `version`,
	RENAME COLUMN `iNextVersion` TO `next_version`,
	RENAME COLUMN `bDeleted` TO `deleted`,
	RENAME COLUMN `lockUserDeletionDate` TO `lock_user_deletion_date`,
	RENAME INDEX `iVersion` TO `version`,
	RENAME INDEX `ID` TO `id`,
	RENAME INDEX `iNextVersion` TO `next_version`,
	RENAME INDEX `bDeleted` TO `deleted`;
ALTER TABLE `history_groups_ancestors`
	RENAME COLUMN `historyID` TO `history_id`,
	RENAME COLUMN `ID` TO `id`,
	RENAME COLUMN `idGroupAncestor` TO `ancestor_group_id`,
	RENAME COLUMN `idGroupChild` TO `child_group_id`,
	RENAME COLUMN `bIsSelf` TO `is_self`,
	RENAME COLUMN `iVersion` TO `version`,
	RENAME COLUMN `iNextVersion` TO `next_version`,
	RENAME COLUMN `bDeleted` TO `deleted`,
	RENAME INDEX `iVersion` TO `version`,
	RENAME INDEX `iNextVersion` TO `next_version`,
	RENAME INDEX `bDeleted` TO `deleted`,
	RENAME INDEX `idGroupAncestor` TO `ancestor_group_id`,
	RENAME INDEX `ID` TO `id`;
ALTER TABLE `history_groups_attempts`
	RENAME COLUMN `historyID` TO `history_id`,
	RENAME COLUMN `ID` TO `id`,
	RENAME COLUMN `idGroup` TO `group_id`,
	RENAME COLUMN `idItem` TO `item_id`,
	RENAME COLUMN `idUserCreator` TO `creator_user_id`,
	RENAME COLUMN `iOrder` TO `order`,
	RENAME COLUMN `iScore` TO `score`,
	RENAME COLUMN `iScoreComputed` TO `score_computed`,
	RENAME COLUMN `iScoreReeval` TO `score_reeval`,
	RENAME COLUMN `iScoreDiffManual` TO `score_diff_manual`,
	RENAME COLUMN `sScoreDiffComment` TO `score_diff_comment`,
	RENAME COLUMN `nbSubmissionsAttempts` TO `submissions_attempts`,
	RENAME COLUMN `nbTasksTried` TO `tasks_tried`,
	RENAME COLUMN `nbTasksSolved` TO `tasks_solved`,
	RENAME COLUMN `nbChildrenValidated` TO `children_validated`,
	RENAME COLUMN `bValidated` TO `validated`,
	RENAME COLUMN `bFinished` TO `finished`,
	RENAME COLUMN `bKeyObtained` TO `key_obtained`,
	RENAME COLUMN `nbTasksWithHelp` TO `tasks_with_help`,
	RENAME COLUMN `sHintsRequested` TO `hints_requested`,
	RENAME COLUMN `nbHintsCached` TO `hints_cached`,
	RENAME COLUMN `nbCorrectionsRead` TO `corrections_read`,
	RENAME COLUMN `iPrecision` TO `precision`,
	RENAME COLUMN `iAutonomy` TO `autonomy`,
	RENAME COLUMN `sStartDate` TO `start_date`,
	RENAME COLUMN `sValidationDate` TO `validation_date`,
	RENAME COLUMN `sFinishDate` TO `finish_date`,
	RENAME COLUMN `sLastActivityDate` TO `last_activity_date`,
	RENAME COLUMN `sThreadStartDate` TO `thread_start_date`,
	RENAME COLUMN `sBestAnswerDate` TO `best_answer_date`,
	RENAME COLUMN `sLastAnswerDate` TO `last_answer_date`,
	RENAME COLUMN `sLastHintDate` TO `last_hint_date`,
	RENAME COLUMN `sContestStartDate` TO `contest_start_date`,
	RENAME COLUMN `bRanked` TO `ranked`,
	RENAME COLUMN `sAllLangProg` TO `all_lang_prog`,
	RENAME COLUMN `iVersion` TO `version`,
	RENAME COLUMN `iNextVersion` TO `next_version`,
	RENAME COLUMN `bDeleted` TO `deleted`,
	RENAME INDEX `iVersion` TO `version`,
	RENAME INDEX `idItem` TO `item_id`,
	RENAME INDEX `GroupItem` TO `group_item`,
	RENAME INDEX `idGroup` TO `group_id`,
	RENAME INDEX `ID` TO `id`;
ALTER TABLE `history_groups_groups`
	RENAME COLUMN `historyID` TO `history_id`,
	RENAME COLUMN `ID` TO `id`,
	RENAME COLUMN `idGroupParent` TO `parent_group_id`,
	RENAME COLUMN `idGroupChild` TO `child_group_id`,
	RENAME COLUMN `iChildOrder` TO `child_order`,
	RENAME COLUMN `sType` TO `type`,
	RENAME COLUMN `sRole` TO `role`,
	RENAME COLUMN `idUserInviting` TO `inviting_user_id`,
	RENAME COLUMN `sStatusDate` TO `status_date`,
	RENAME COLUMN `iVersion` TO `version`,
	RENAME COLUMN `iNextVersion` TO `next_version`,
	RENAME COLUMN `bDeleted` TO `deleted`,
	RENAME INDEX `iVersion` TO `version`,
	RENAME INDEX `ID` TO `id`,
	RENAME INDEX `iNextVersion` TO `next_version`,
	RENAME INDEX `bDeleted` TO `deleted`,
	RENAME INDEX `idGroupParent` TO `parent_group_id`,
	RENAME INDEX `idGroupChild` TO `child_group_id`;
ALTER TABLE `history_groups_items`
	RENAME COLUMN `historyID` TO `history_id`,
	RENAME COLUMN `ID` TO `id`,
	RENAME COLUMN `idGroup` TO `group_id`,
	RENAME COLUMN `idItem` TO `item_id`,
	RENAME COLUMN `idUserCreated` TO `creator_user_id`,
	RENAME COLUMN `sPartialAccessDate` TO `partial_access_date`,
	RENAME COLUMN `sAccessReason` TO `access_reason`,
	RENAME COLUMN `sFullAccessDate` TO `full_access_date`,
	RENAME COLUMN `sAccessSolutionsDate` TO `access_solutions_date`,
	RENAME COLUMN `bOwnerAccess` TO `owner_access`,
	RENAME COLUMN `bManagerAccess` TO `manager_access`,
	RENAME COLUMN `sCachedFullAccessDate` TO `cached_full_access_date`,
	RENAME COLUMN `sCachedPartialAccessDate` TO `cached_partial_access_date`,
	RENAME COLUMN `sCachedAccessSolutionsDate` TO `cached_access_solutions_date`,
	RENAME COLUMN `sCachedGrayedAccessDate` TO `cached_grayed_access_date`,
	RENAME COLUMN `sCachedAccessReason` TO `cached_access_reason`,
	RENAME COLUMN `bCachedFullAccess` TO `cached_full_access`,
	RENAME COLUMN `bCachedPartialAccess` TO `cached_partial_access`,
	RENAME COLUMN `bCachedAccessSolutions` TO `cached_access_solutions`,
	RENAME COLUMN `bCachedGrayedAccess` TO `cached_grayed_access`,
	RENAME COLUMN `bCachedManagerAccess` TO `cached_manager_access`,
	RENAME COLUMN `sPropagateAccess` TO `propagate_access`,
	RENAME COLUMN `iVersion` TO `version`,
	RENAME COLUMN `iNextVersion` TO `next_version`,
	RENAME COLUMN `bDeleted` TO `deleted`,
	RENAME INDEX `iVersion` TO `version`,
	RENAME INDEX `ID` TO `id`,
	RENAME INDEX `itemGroup` TO `item_group`,
	RENAME INDEX `iNextVersion` TO `next_version`,
	RENAME INDEX `bDeleted` TO `deleted`,
	RENAME INDEX `idItem` TO `item_id`,
	RENAME INDEX `idGroup` TO `group_id`;
ALTER TABLE `history_groups_login_prefixes`
	RENAME COLUMN `historyID` TO `history_id`,
	RENAME COLUMN `ID` TO `id`,
	RENAME COLUMN `idGroup` TO `group_id`,
	RENAME COLUMN `iVersion` TO `version`,
	RENAME COLUMN `iNextVersion` TO `next_version`,
	RENAME COLUMN `bDeleted` TO `deleted`,
	RENAME INDEX `idGroup` TO `group_id`;
ALTER TABLE `history_items`
	RENAME COLUMN `historyID` TO `history_id`,
	RENAME COLUMN `ID` TO `id`,
	RENAME COLUMN `sUrl` TO `url`,
	RENAME COLUMN `idPlatform` TO `platform_id`,
	RENAME COLUMN `sTextId` TO `text_id`,
	RENAME COLUMN `sRepositoryPath` TO `repository_path`,
	RENAME COLUMN `sType` TO `type`,
	RENAME COLUMN `bTitleBarVisible` TO `title_bar_visible`,
	RENAME COLUMN `bTransparentFolder` TO `transparent_folder`,
	RENAME COLUMN `bDisplayDetailsInParent` TO `display_details_in_parent`,
	RENAME COLUMN `bCustomChapter` TO `custom_chapter`,
	RENAME COLUMN `bDisplayChildrenAsTabs` TO `display_children_as_tabs`,
	RENAME COLUMN `bUsesAPI` TO `uses_api`,
	RENAME COLUMN `bReadOnly` TO `read_only`,
	RENAME COLUMN `sFullScreen` TO `full_screen`,
	RENAME COLUMN `bShowDifficulty` TO `show_difficulty`,
	RENAME COLUMN `bShowSource` TO `show_source`,
	RENAME COLUMN `bHintsAllowed` TO `hints_allowed`,
	RENAME COLUMN `bFixedRanks` TO `fixed_ranks`,
	RENAME COLUMN `sValidationType` TO `validation_type`,
	RENAME COLUMN `iValidationMin` TO `validation_min`,
	RENAME COLUMN `sPreparationState` TO `preparation_state`,
	RENAME COLUMN `idItemUnlocked` TO `unlocked_item_ids`,
	RENAME COLUMN `iScoreMinUnlock` TO `score_min_unlock`,
	RENAME COLUMN `sSupportedLangProg` TO `supported_lang_prog`,
	RENAME COLUMN `idDefaultLanguage` TO `default_language_id`,
	RENAME COLUMN `sTeamMode` TO `team_mode`,
	RENAME COLUMN `bTeamsEditable` TO `teams_editable`,
	RENAME COLUMN `idTeamInGroup` TO `qualified_group_id`,
	RENAME COLUMN `iTeamMaxMembers` TO `team_max_members`,
	RENAME COLUMN `bHasAttempts` TO `has_attempts`,
	RENAME COLUMN `sAccessOpenDate` TO `access_open_date`,
	RENAME COLUMN `sDuration` TO `duration`,
	RENAME COLUMN `sEndContestDate` TO `end_contest_date`,
	RENAME COLUMN `bShowUserInfos` TO `show_user_infos`,
	RENAME COLUMN `sContestPhase` TO `contest_phase`,
	RENAME COLUMN `iLevel` TO `level`,
	RENAME COLUMN `bNoScore` TO `no_score`,
	RENAME COLUMN `groupCodeEnter` TO `group_code_enter`,
	RENAME COLUMN `iVersion` TO `version`,
	RENAME COLUMN `iNextVersion` TO `next_version`,
	RENAME COLUMN `bDeleted` TO `deleted`,
	RENAME INDEX `ID` TO `id`,
	RENAME INDEX `iVersion` TO `version`,
	RENAME INDEX `iNextVersion` TO `next_version`,
	RENAME INDEX `bDeleted` TO `deleted`;
ALTER TABLE `history_items_ancestors`
	RENAME COLUMN `historyID` TO `history_id`,
	RENAME COLUMN `ID` TO `id`,
	RENAME COLUMN `idItemAncestor` TO `ancestor_item_id`,
	RENAME COLUMN `idItemChild` TO `child_item_id`,
	RENAME COLUMN `iVersion` TO `version`,
	RENAME COLUMN `iNextVersion` TO `next_version`,
	RENAME COLUMN `bDeleted` TO `deleted`,
	RENAME INDEX `iVersion` TO `version`,
	RENAME INDEX `iNextVersion` TO `next_version`,
	RENAME INDEX `bDeleted` TO `deleted`,
	RENAME INDEX `idItemAncestor` TO `ancestor_item_id_child_item_id`,
	RENAME INDEX `idItemAncestortor` TO `ancestor_item_id`,
	RENAME INDEX `idItemChild` TO `child_item_id`,
	RENAME INDEX `ID` TO `id`;
ALTER TABLE `history_items_items`
	RENAME COLUMN `historyID` TO `history_id`,
	RENAME COLUMN `ID` TO `id`,
	RENAME COLUMN `idItemParent` TO `parent_item_id`,
	RENAME COLUMN `idItemChild` TO `child_item_id`,
	RENAME COLUMN `iChildOrder` TO `child_order`,
	RENAME COLUMN `sCategory` TO `category`,
	RENAME COLUMN `bAlwaysVisible` TO `always_visible`,
	RENAME COLUMN `bAccessRestricted` TO `access_restricted`,
	RENAME COLUMN `iDifficulty` TO `difficulty`,
	RENAME COLUMN `iVersion` TO `version`,
	RENAME COLUMN `iNextVersion` TO `next_version`,
	RENAME COLUMN `bDeleted` TO `deleted`,
	RENAME INDEX `ID` TO `id`,
	RENAME INDEX `iVersion` TO `version`,
	RENAME INDEX `idItemParent` TO `parent_item_id`,
	RENAME INDEX `idItemChild` TO `child_item_id`,
	RENAME INDEX `iNextVersion` TO `next_version`,
	RENAME INDEX `bDeleted` TO `deleted`,
	RENAME INDEX `parentChild` TO `parent_child`;
ALTER TABLE `history_items_strings`
	RENAME COLUMN `historyID` TO `history_id`,
	RENAME COLUMN `ID` TO `id`,
	RENAME COLUMN `idItem` TO `item_id`,
	RENAME COLUMN `idLanguage` TO `language_id`,
	RENAME COLUMN `sTranslator` TO `translator`,
	RENAME COLUMN `sTitle` TO `title`,
	RENAME COLUMN `sImageUrl` TO `image_url`,
	RENAME COLUMN `sSubtitle` TO `subtitle`,
	RENAME COLUMN `sDescription` TO `description`,
	RENAME COLUMN `sEduComment` TO `edu_comment`,
	RENAME COLUMN `sRankingComment` TO `ranking_comment`,
	RENAME COLUMN `iVersion` TO `version`,
	RENAME COLUMN `iNextVersion` TO `next_version`,
	RENAME COLUMN `bDeleted` TO `deleted`,
	RENAME INDEX `ID` TO `id`,
	RENAME INDEX `iVersion` TO `version`,
	RENAME INDEX `itemLanguage` TO `item_language`,
	RENAME INDEX `idItem` TO `item_id`;
ALTER TABLE `history_languages`
	RENAME COLUMN `historyID` TO `history_id`,
	RENAME COLUMN `ID` TO `id`,
	RENAME COLUMN `sName` TO `name`,
	RENAME COLUMN `sCode` TO `code`,
	RENAME COLUMN `iVersion` TO `version`,
	RENAME COLUMN `iNextVersion` TO `next_version`,
	RENAME COLUMN `bDeleted` TO `deleted`,
	RENAME INDEX `ID` TO `id`,
	RENAME INDEX `iVersion` TO `version`,
	RENAME INDEX `sCode` TO `code`;
ALTER TABLE `history_messages`
	RENAME COLUMN `history_ID` TO `history_id`,
	RENAME COLUMN `ID` TO `id`,
	RENAME COLUMN `idThread` TO `thread_id`,
	RENAME COLUMN `idUser` TO `user_id`,
	RENAME COLUMN `sSubmissionDate` TO `submission_date`,
	RENAME COLUMN `bPublished` TO `published`,
	RENAME COLUMN `sTitle` TO `title`,
	RENAME COLUMN `sBody` TO `body`,
	RENAME COLUMN `bTrainersOnly` TO `trainers_only`,
	RENAME COLUMN `bArchived` TO `archived`,
	RENAME COLUMN `bPersistant` TO `persistant`,
	RENAME COLUMN `iVersion` TO `version`,
	RENAME COLUMN `iNextVersion` TO `next_version`,
	RENAME COLUMN `bDeleted` TO `deleted`,
	RENAME INDEX `iVersion` TO `version`,
	RENAME INDEX `iNextVersion` TO `next_version`,
	RENAME INDEX `bDeleted` TO `deleted`,
	RENAME INDEX `ID` TO `id`;
ALTER TABLE `history_threads`
	RENAME COLUMN `history_ID` TO `history_id`,
	RENAME COLUMN `ID` TO `id`,
	RENAME COLUMN `sType` TO `type`,
	RENAME COLUMN `idUserCreated` TO `creator_user_id`,
	RENAME COLUMN `idItem` TO `item_id`,
	RENAME COLUMN `sLastActivityDate` TO `last_activity_date`,
	RENAME COLUMN `sTitle` TO `title`,
	RENAME COLUMN `bAdminHelpAsked` TO `admin_help_asked`,
	RENAME COLUMN `bHidden` TO `hidden`,
	RENAME COLUMN `iVersion` TO `version`,
	RENAME COLUMN `iNextVersion` TO `next_version`,
	RENAME COLUMN `bDeleted` TO `deleted`,
	RENAME INDEX `iVersion` TO `version`,
	RENAME INDEX `iNextVersion` TO `next_version`,
	RENAME INDEX `bDeleted` TO `deleted`,
	RENAME INDEX `ID` TO `id`;
ALTER TABLE `history_users`
	RENAME COLUMN `historyID` TO `history_id`,
	RENAME COLUMN `ID` TO `id`,
	RENAME COLUMN `loginID` TO `login_id`,
	RENAME COLUMN `sLogin` TO `login`,
	RENAME COLUMN `sOpenIdIdentity` TO `open_id_identity`,
	RENAME COLUMN `sPasswordMd5` TO `password_md5`,
	RENAME COLUMN `sSalt` TO `salt`,
	RENAME COLUMN `sRecover` TO `recover`,
	RENAME COLUMN `sRegistrationDate` TO `registration_date`,
	RENAME COLUMN `sEmail` TO `email`,
	RENAME COLUMN `bEmailVerified` TO `email_verified`,
	RENAME COLUMN `sFirstName` TO `first_name`,
	RENAME COLUMN `sLastName` TO `last_name`,
	RENAME COLUMN `sStudentId` TO `student_id`,
	RENAME COLUMN `sCountryCode` TO `country_code`,
	RENAME COLUMN `sTimeZone` TO `time_zone`,
	RENAME COLUMN `sBirthDate` TO `birth_date`,
	RENAME COLUMN `iGraduationYear` TO `graduation_year`,
	RENAME COLUMN `iGrade` TO `grade`,
	RENAME COLUMN `sSex` TO `sex`,
	RENAME COLUMN `sAddress` TO `address`,
	RENAME COLUMN `sZipcode` TO `zipcode`,
	RENAME COLUMN `sCity` TO `city`,
	RENAME COLUMN `sLandLineNumber` TO `land_line_number`,
	RENAME COLUMN `sCellPhoneNumber` TO `cell_phone_number`,
	RENAME COLUMN `sDefaultLanguage` TO `default_language`,
	RENAME COLUMN `bNotifyNews` TO `notify_news`,
	RENAME COLUMN `sNotify` TO `notify`,
	RENAME COLUMN `bPublicFirstName` TO `public_first_name`,
	RENAME COLUMN `bPublicLastName` TO `public_last_name`,
	RENAME COLUMN `sFreeText` TO `free_text`,
	RENAME COLUMN `sWebSite` TO `web_site`,
	RENAME COLUMN `bPhotoAutoload` TO `photo_autoload`,
	RENAME COLUMN `sLangProg` TO `lang_prog`,
	RENAME COLUMN `sLastLoginDate` TO `last_login_date`,
	RENAME COLUMN `sLastActivityDate` TO `last_activity_date`,
	RENAME COLUMN `sLastIP` TO `last_ip`,
	RENAME COLUMN `bBasicEditorMode` TO `basic_editor_mode`,
	RENAME COLUMN `nbSpacesForTab` TO `spaces_for_tab`,
	RENAME COLUMN `iMemberState` TO `member_state`,
	RENAME COLUMN `idUserGodfather` TO `godfather_user_id`,
	RENAME COLUMN `iStepLevelInSite` TO `step_level_in_site`,
	RENAME COLUMN `bIsAdmin` TO `is_admin`,
	RENAME COLUMN `bNoRanking` TO `no_ranking`,
	RENAME COLUMN `nbHelpGiven` TO `help_given`,
	RENAME COLUMN `idGroupSelf` TO `self_group_id`,
	RENAME COLUMN `idGroupOwned` TO `owned_group_id`,
	RENAME COLUMN `idGroupAccess` TO `access_group_id`,
	RENAME COLUMN `sNotificationReadDate` TO `notification_read_date`,
	RENAME COLUMN `iVersion` TO `version`,
	RENAME COLUMN `iNextVersion` TO `next_version`,
	RENAME COLUMN `bDeleted` TO `deleted`,
	RENAME COLUMN `loginModulePrefix` TO `login_module_prefix`,
	RENAME COLUMN `creatorID` TO `creator_id`,
	RENAME COLUMN `allowSubgroups` TO `allow_subgroups`,
	RENAME INDEX `ID` TO `id`,
	RENAME INDEX `iVersion` TO `version`,
	RENAME INDEX `sCountryCode` TO `country_code`,
	RENAME INDEX `idUserGodfather` TO `godfather_user_id`,
	RENAME INDEX `sLangProg` TO `lang_prog`,
	RENAME INDEX `iNextVersion` TO `next_version`,
	RENAME INDEX `bDeleted` TO `deleted`,
	RENAME INDEX `idGroupSelf` TO `self_group_id`,
	RENAME INDEX `idGroupOwned` TO `owned_group_id`;
ALTER TABLE `history_users_items`
	RENAME COLUMN `historyID` TO `history_id`,
	RENAME COLUMN `ID` TO `id`,
	RENAME COLUMN `idUser` TO `user_id`,
	RENAME COLUMN `idItem` TO `item_id`,
	RENAME COLUMN `idAttemptActive` TO `active_attempt_id`,
	RENAME COLUMN `iScore` TO `score`,
	RENAME COLUMN `iScoreComputed` TO `score_computed`,
	RENAME COLUMN `iScoreReeval` TO `score_reeval`,
	RENAME COLUMN `iScoreDiffManual` TO `score_diff_manual`,
	RENAME COLUMN `sScoreDiffComment` TO `score_diff_comment`,
	RENAME COLUMN `nbSubmissionsAttempts` TO `submissions_attempts`,
	RENAME COLUMN `nbTasksTried` TO `tasks_tried`,
	RENAME COLUMN `nbTasksSolved` TO `tasks_solved`,
	RENAME COLUMN `nbChildrenValidated` TO `children_validated`,
	RENAME COLUMN `bValidated` TO `validated`,
	RENAME COLUMN `bFinished` TO `finished`,
	RENAME COLUMN `bKeyObtained` TO `key_obtained`,
	RENAME COLUMN `nbTasksWithHelp` TO `tasks_with_help`,
	RENAME COLUMN `sHintsRequested` TO `hints_requested`,
	RENAME COLUMN `nbHintsCached` TO `hints_cached`,
	RENAME COLUMN `nbCorrectionsRead` TO `corrections_read`,
	RENAME COLUMN `iPrecision` TO `precision`,
	RENAME COLUMN `iAutonomy` TO `autonomy`,
	RENAME COLUMN `sStartDate` TO `start_date`,
	RENAME COLUMN `sValidationDate` TO `validation_date`,
	RENAME COLUMN `sFinishDate` TO `finish_date`,
	RENAME COLUMN `sLastActivityDate` TO `last_activity_date`,
	RENAME COLUMN `sThreadStartDate` TO `thread_start_date`,
	RENAME COLUMN `sBestAnswerDate` TO `best_answer_date`,
	RENAME COLUMN `sLastAnswerDate` TO `last_answer_date`,
	RENAME COLUMN `sLastHintDate` TO `last_hint_date`,
	RENAME COLUMN `sContestStartDate` TO `contest_start_date`,
	RENAME COLUMN `bRanked` TO `ranked`,
	RENAME COLUMN `sAllLangProg` TO `all_lang_prog`,
	RENAME COLUMN `sState` TO `state`,
	RENAME COLUMN `sAnswer` TO `answer`,
	RENAME COLUMN `iVersion` TO `version`,
	RENAME COLUMN `iNextVersion` TO `next_version`,
	RENAME COLUMN `bDeleted` TO `deleted`,
	RENAME COLUMN `bPlatformDataRemoved` TO `platform_data_removed`,
	RENAME INDEX `ID` TO `id`,
	RENAME INDEX `iVersion` TO `version`,
	RENAME INDEX `itemUser` TO `item_user`,
	RENAME INDEX `iNextVersion` TO `next_version`,
	RENAME INDEX `bDeleted` TO `deleted`,
	RENAME INDEX `idItem` TO `item_id`,
	RENAME INDEX `idUser` TO `user_id`;
ALTER TABLE `history_users_threads`
	RENAME COLUMN `history_ID` TO `history_id`,
	RENAME COLUMN `ID` TO `id`,
	RENAME COLUMN `idUser` TO `user_id`,
	RENAME COLUMN `idThread` TO `thread_id`,
	RENAME COLUMN `sLastReadDate` TO `last_read_date`,
	RENAME COLUMN `bParticipated` TO `participated`,
	RENAME COLUMN `sLastWriteDate` TO `last_write_date`,
	RENAME COLUMN `bStarred` TO `starred`,
	RENAME COLUMN `iVersion` TO `version`,
	RENAME COLUMN `iNextVersion` TO `next_version`,
	RENAME COLUMN `bDeleted` TO `deleted`,
	RENAME INDEX `userThread` TO `user_thread`,
	RENAME INDEX `iVersion` TO `version`,
	RENAME INDEX `iNextVersion` TO `next_version`,
	RENAME INDEX `bDeleted` TO `deleted`,
	RENAME INDEX `ID` TO `id`;
ALTER TABLE `items`
	RENAME COLUMN `ID` TO `id`,
	RENAME COLUMN `sUrl` TO `url`,
	RENAME COLUMN `idPlatform` TO `platform_id`,
	RENAME COLUMN `sTextId` TO `text_id`,
	RENAME COLUMN `sRepositoryPath` TO `repository_path`,
	RENAME COLUMN `sType` TO `type`,
	RENAME COLUMN `bTitleBarVisible` TO `title_bar_visible`,
	RENAME COLUMN `bTransparentFolder` TO `transparent_folder`,
	RENAME COLUMN `bDisplayDetailsInParent` TO `display_details_in_parent`,
	RENAME COLUMN `bCustomChapter` TO `custom_chapter`,
	RENAME COLUMN `bDisplayChildrenAsTabs` TO `display_children_as_tabs`,
	RENAME COLUMN `bUsesAPI` TO `uses_api`,
	RENAME COLUMN `bReadOnly` TO `read_only`,
	RENAME COLUMN `sFullScreen` TO `full_screen`,
	RENAME COLUMN `bShowDifficulty` TO `show_difficulty`,
	RENAME COLUMN `bShowSource` TO `show_source`,
	RENAME COLUMN `bHintsAllowed` TO `hints_allowed`,
	RENAME COLUMN `bFixedRanks` TO `fixed_ranks`,
	RENAME COLUMN `sValidationType` TO `validation_type`,
	RENAME COLUMN `iValidationMin` TO `validation_min`,
	RENAME COLUMN `sPreparationState` TO `preparation_state`,
	RENAME COLUMN `idItemUnlocked` TO `unlocked_item_ids`,
	RENAME COLUMN `iScoreMinUnlock` TO `score_min_unlock`,
	RENAME COLUMN `sSupportedLangProg` TO `supported_lang_prog`,
	RENAME COLUMN `idDefaultLanguage` TO `default_language_id`,
	RENAME COLUMN `sTeamMode` TO `team_mode`,
	RENAME COLUMN `bTeamsEditable` TO `teams_editable`,
	RENAME COLUMN `idTeamInGroup` TO `qualified_group_id`,
	RENAME COLUMN `iTeamMaxMembers` TO `team_max_members`,
	RENAME COLUMN `bHasAttempts` TO `has_attempts`,
	RENAME COLUMN `sAccessOpenDate` TO `access_open_date`,
	RENAME COLUMN `sDuration` TO `duration`,
	RENAME COLUMN `sEndContestDate` TO `end_contest_date`,
	RENAME COLUMN `bShowUserInfos` TO `show_user_infos`,
	RENAME COLUMN `sContestPhase` TO `contest_phase`,
	RENAME COLUMN `iLevel` TO `level`,
	RENAME COLUMN `bNoScore` TO `no_score`,
	RENAME COLUMN `groupCodeEnter` TO `group_code_enter`,
	RENAME COLUMN `iVersion` TO `version`,
	RENAME INDEX `iVersion` TO `version`;
ALTER TABLE `items_ancestors`
	RENAME COLUMN `ID` TO `id`,
	RENAME COLUMN `idItemAncestor` TO `ancestor_item_id`,
	RENAME COLUMN `idItemChild` TO `child_item_id`,
	RENAME COLUMN `iVersion` TO `version`,
	RENAME INDEX `idItemAncestor` TO `ancestor_item_id_child_item_id`,
	RENAME INDEX `idItemAncestortor` TO `ancestor_item_id`,
	RENAME INDEX `idItemChild` TO `child_item_id`;
ALTER TABLE `items_items`
	RENAME COLUMN `ID` TO `id`,
	RENAME COLUMN `idItemParent` TO `parent_item_id`,
	RENAME COLUMN `idItemChild` TO `child_item_id`,
	RENAME COLUMN `iChildOrder` TO `child_order`,
	RENAME COLUMN `sCategory` TO `category`,
	RENAME COLUMN `bAlwaysVisible` TO `always_visible`,
	RENAME COLUMN `bAccessRestricted` TO `access_restricted`,
	RENAME COLUMN `iDifficulty` TO `difficulty`,
	RENAME COLUMN `iVersion` TO `version`,
	RENAME INDEX `idItemParent` TO `parent_item_id`,
	RENAME INDEX `idItemChild` TO `child_item_id`,
	RENAME INDEX `iVersion` TO `version`,
	RENAME INDEX `parentChild` TO `parent_child`,
	RENAME INDEX `parentVersion` TO `parent_version`;
ALTER TABLE `items_propagate`
	RENAME COLUMN `ID` TO `id`,
	RENAME COLUMN `sAncestorsComputationState` TO `ancestors_computation_state`,
	RENAME INDEX `sAncestorsComputationDate` TO `ancestors_computation_date`;
ALTER TABLE `items_strings`
	RENAME COLUMN `ID` TO `id`,
	RENAME COLUMN `idItem` TO `item_id`,
	RENAME COLUMN `idLanguage` TO `language_id`,
	RENAME COLUMN `sTranslator` TO `translator`,
	RENAME COLUMN `sTitle` TO `title`,
	RENAME COLUMN `sImageUrl` TO `image_url`,
	RENAME COLUMN `sSubtitle` TO `subtitle`,
	RENAME COLUMN `sDescription` TO `description`,
	RENAME COLUMN `sEduComment` TO `edu_comment`,
	RENAME COLUMN `sRankingComment` TO `ranking_comment`,
	RENAME COLUMN `iVersion` TO `version`,
	RENAME INDEX `idItem` TO `item_id_language_id`,
	RENAME INDEX `iVersion` TO `version`,
	RENAME INDEX `idItemAlone` TO `item_id`;
ALTER TABLE `languages`
	RENAME COLUMN `ID` TO `id`,
	RENAME COLUMN `sName` TO `name`,
	RENAME COLUMN `sCode` TO `code`,
	RENAME COLUMN `iVersion` TO `version`,
	RENAME INDEX `iVersion` TO `version`,
	RENAME INDEX `sCode` TO `code`;
ALTER TABLE `login_states`
	RENAME COLUMN `sCookie` TO `cookie`,
	RENAME COLUMN `sState` TO `state`,
	RENAME COLUMN `sExpirationDate` TO `expiration_date`,
	RENAME INDEX `sExpirationDate` TO `expiration_date`;
ALTER TABLE `messages`
	RENAME COLUMN `ID` TO `id`,
	RENAME COLUMN `idThread` TO `thread_id`,
	RENAME COLUMN `idUser` TO `user_id`,
	RENAME COLUMN `sSubmissionDate` TO `submission_date`,
	RENAME COLUMN `bPublished` TO `published`,
	RENAME COLUMN `sTitle` TO `title`,
	RENAME COLUMN `sBody` TO `body`,
	RENAME COLUMN `bTrainersOnly` TO `trainers_only`,
	RENAME COLUMN `bArchived` TO `archived`,
	RENAME COLUMN `bPersistant` TO `persistant`,
	RENAME COLUMN `iVersion` TO `version`,
	RENAME INDEX `iVersion` TO `version`,
	RENAME INDEX `idThread` TO `thread_id`;
ALTER TABLE `platforms`
	RENAME COLUMN `ID` TO `id`,
	RENAME COLUMN `sName` TO `name`,
	RENAME COLUMN `sBaseUrl` TO `base_url`,
	RENAME COLUMN `sPublicKey` TO `public_key`,
	RENAME COLUMN `bUsesTokens` TO `uses_tokens`,
	RENAME COLUMN `sRegexp` TO `regexp`,
	RENAME COLUMN `iPriority` TO `priority`;
ALTER TABLE `refresh_tokens`
	RENAME COLUMN `idUser` TO `user_id`,
	RENAME COLUMN `sRefreshToken` TO `refresh_token`,
	RENAME INDEX `sRefreshTokenPrefix` TO `refresh_token_prefix`;
ALTER TABLE `sessions`
	RENAME COLUMN `sAccessToken` TO `access_token`,
	RENAME COLUMN `idUser` TO `user_id`,
	RENAME COLUMN `sExpirationDate` TO `expiration_date`,
	RENAME COLUMN `sIssuedAtDate` TO `issued_at_date`,
	RENAME COLUMN `sIssuer` TO `issuer`,
	RENAME INDEX `sExpirationDate` TO `expiration_date`,
	RENAME INDEX `sAccessTokenPrefix` TO `access_token_prefix`,
	RENAME INDEX `idUser` TO `user_id`;
ALTER TABLE `synchro_version`
	RENAME COLUMN `ID` TO `id`,
	RENAME COLUMN `iVersion` TO `version`,
	RENAME COLUMN `iLastServerVersion` TO `last_server_version`,
	RENAME COLUMN `iLastClientVersion` TO `last_client_version`,
	RENAME INDEX `iVersion` TO `version`;
ALTER TABLE `threads`
	RENAME COLUMN `ID` TO `id`,
	RENAME COLUMN `sType` TO `type`,
	RENAME COLUMN `sLastActivityDate` TO `last_activity_date`,
	RENAME COLUMN `idUserCreated` TO `creator_user_id`,
	RENAME COLUMN `idItem` TO `item_id`,
	RENAME COLUMN `sTitle` TO `title`,
	RENAME COLUMN `bAdminHelpAsked` TO `admin_help_asked`,
	RENAME COLUMN `bHidden` TO `hidden`,
	RENAME COLUMN `iVersion` TO `version`,
	RENAME INDEX `iVersion` TO `version`;
ALTER TABLE `users`
	RENAME COLUMN `ID` TO `id`,
	RENAME COLUMN `loginID` TO `login_id`,
	RENAME COLUMN `tempUser` TO `temp_user`,
	RENAME COLUMN `sLogin` TO `login`,
	RENAME COLUMN `sOpenIdIdentity` TO `open_id_identity`,
	RENAME COLUMN `sPasswordMd5` TO `password_md5`,
	RENAME COLUMN `sSalt` TO `salt`,
	RENAME COLUMN `sRecover` TO `recover`,
	RENAME COLUMN `sRegistrationDate` TO `registration_date`,
	RENAME COLUMN `sEmail` TO `email`,
	RENAME COLUMN `bEmailVerified` TO `email_verified`,
	RENAME COLUMN `sFirstName` TO `first_name`,
	RENAME COLUMN `sLastName` TO `last_name`,
	RENAME COLUMN `sStudentId` TO `student_id`,
	RENAME COLUMN `sCountryCode` TO `country_code`,
	RENAME COLUMN `sTimeZone` TO `time_zone`,
	RENAME COLUMN `sBirthDate` TO `birth_date`,
	RENAME COLUMN `iGraduationYear` TO `graduation_year`,
	RENAME COLUMN `iGrade` TO `grade`,
	RENAME COLUMN `sSex` TO `sex`,
	RENAME COLUMN `sAddress` TO `address`,
	RENAME COLUMN `sZipcode` TO `zipcode`,
	RENAME COLUMN `sCity` TO `city`,
	RENAME COLUMN `sLandLineNumber` TO `land_line_number`,
	RENAME COLUMN `sCellPhoneNumber` TO `cell_phone_number`,
	RENAME COLUMN `sDefaultLanguage` TO `default_language`,
	RENAME COLUMN `bNotifyNews` TO `notify_news`,
	RENAME COLUMN `sNotify` TO `notify`,
	RENAME COLUMN `bPublicFirstName` TO `public_first_name`,
	RENAME COLUMN `bPublicLastName` TO `public_last_name`,
	RENAME COLUMN `sFreeText` TO `free_text`,
	RENAME COLUMN `sWebSite` TO `web_site`,
	RENAME COLUMN `bPhotoAutoload` TO `photo_autoload`,
	RENAME COLUMN `sLangProg` TO `lang_prog`,
	RENAME COLUMN `sLastLoginDate` TO `last_login_date`,
	RENAME COLUMN `sLastActivityDate` TO `last_activity_date`,
	RENAME COLUMN `sLastIP` TO `last_ip`,
	RENAME COLUMN `bBasicEditorMode` TO `basic_editor_mode`,
	RENAME COLUMN `nbSpacesForTab` TO `spaces_for_tab`,
	RENAME COLUMN `iMemberState` TO `member_state`,
	RENAME COLUMN `idUserGodfather` TO `godfather_user_id`,
	RENAME COLUMN `iStepLevelInSite` TO `step_level_in_site`,
	RENAME COLUMN `bIsAdmin` TO `is_admin`,
	RENAME COLUMN `bNoRanking` TO `no_ranking`,
	RENAME COLUMN `nbHelpGiven` TO `help_given`,
	RENAME COLUMN `idGroupSelf` TO `self_group_id`,
	RENAME COLUMN `idGroupOwned` TO `owned_group_id`,
	RENAME COLUMN `idGroupAccess` TO `access_group_id`,
	RENAME COLUMN `sNotificationReadDate` TO `notification_read_date`,
	RENAME COLUMN `iVersion` TO `version`,
	RENAME COLUMN `loginModulePrefix` TO `login_module_prefix`,
	RENAME COLUMN `creatorID` TO `creator_id`,
	RENAME COLUMN `allowSubgroups` TO `allow_subgroups`,
	RENAME INDEX `sLogin` TO `login`,
	RENAME INDEX `idGroupSelf` TO `self_group_id`,
	RENAME INDEX `idGroupOwned` TO `owned_group_id`,
	RENAME INDEX `iVersion` TO `version`,
	RENAME INDEX `sCountryCode` TO `country_code`,
	RENAME INDEX `idUserGodfather` TO `godfather_user_id`,
	RENAME INDEX `sLangProg` TO `lang_prog`,
	RENAME INDEX `loginID` TO `login_id`,
	RENAME INDEX `tempUser` TO `temp_user`;
ALTER TABLE `users_answers`
	RENAME COLUMN `ID` TO `id`,
	RENAME COLUMN `idUser` TO `user_id`,
	RENAME COLUMN `idItem` TO `item_id`,
	RENAME COLUMN `idAttempt` TO `attempt_id`,
	RENAME COLUMN `sName` TO `name`,
	RENAME COLUMN `sType` TO `type`,
	RENAME COLUMN `sState` TO `state`,
	RENAME COLUMN `sAnswer` TO `answer`,
	RENAME COLUMN `sLangProg` TO `lang_prog`,
	RENAME COLUMN `sSubmissionDate` TO `submission_date`,
	RENAME COLUMN `iScore` TO `score`,
	RENAME COLUMN `bValidated` TO `validated`,
	RENAME COLUMN `sGradingDate` TO `grading_date`,
	RENAME COLUMN `idUserGrader` TO `grader_user_id`,
	RENAME INDEX `idUser` TO `user_id`,
	RENAME INDEX `idItem` TO `item_id`,
	RENAME INDEX `idAttempt` TO `attempt_id`;
ALTER TABLE `users_items`
	RENAME COLUMN `ID` TO `id`,
	RENAME COLUMN `idUser` TO `user_id`,
	RENAME COLUMN `idItem` TO `item_id`,
	RENAME COLUMN `idAttemptActive` TO `active_attempt_id`,
	RENAME COLUMN `iScore` TO `score`,
	RENAME COLUMN `iScoreComputed` TO `score_computed`,
	RENAME COLUMN `iScoreReeval` TO `score_reeval`,
	RENAME COLUMN `iScoreDiffManual` TO `score_diff_manual`,
	RENAME COLUMN `sScoreDiffComment` TO `score_diff_comment`,
	RENAME COLUMN `nbSubmissionsAttempts` TO `submissions_attempts`,
	RENAME COLUMN `nbTasksTried` TO `tasks_tried`,
	RENAME COLUMN `nbTasksSolved` TO `tasks_solved`,
	RENAME COLUMN `nbChildrenValidated` TO `children_validated`,
	RENAME COLUMN `bValidated` TO `validated`,
	RENAME COLUMN `bFinished` TO `finished`,
	RENAME COLUMN `bKeyObtained` TO `key_obtained`,
	RENAME COLUMN `nbTasksWithHelp` TO `tasks_with_help`,
	RENAME COLUMN `sHintsRequested` TO `hints_requested`,
	RENAME COLUMN `nbHintsCached` TO `hints_cached`,
	RENAME COLUMN `nbCorrectionsRead` TO `corrections_read`,
	RENAME COLUMN `iPrecision` TO `precision`,
	RENAME COLUMN `iAutonomy` TO `autonomy`,
	RENAME COLUMN `sStartDate` TO `start_date`,
	RENAME COLUMN `sValidationDate` TO `validation_date`,
	RENAME COLUMN `sFinishDate` TO `finish_date`,
	RENAME COLUMN `sLastActivityDate` TO `last_activity_date`,
	RENAME COLUMN `sThreadStartDate` TO `thread_start_date`,
	RENAME COLUMN `sBestAnswerDate` TO `best_answer_date`,
	RENAME COLUMN `sLastAnswerDate` TO `last_answer_date`,
	RENAME COLUMN `sLastHintDate` TO `last_hint_date`,
	RENAME COLUMN `sContestStartDate` TO `contest_start_date`,
	RENAME COLUMN `bRanked` TO `ranked`,
	RENAME COLUMN `sAllLangProg` TO `all_lang_prog`,
	RENAME COLUMN `iVersion` TO `version`,
	RENAME COLUMN `sAncestorsComputationState` TO `ancestors_computation_state`,
	RENAME COLUMN `sState` TO `state`,
	RENAME COLUMN `sAnswer` TO `answer`,
	RENAME COLUMN `bPlatformDataRemoved` TO `platform_data_removed`,
	RENAME INDEX `UserItem` TO `user_item`,
	RENAME INDEX `iVersion` TO `version`,
	RENAME INDEX `sAncestorsComputationState` TO `ancestors_computation_state`,
	RENAME INDEX `idItem` TO `item_id`,
	RENAME INDEX `idUser` TO `user_id`,
	RENAME INDEX `idAttemptActive` TO `active_attempt_id`;
ALTER TABLE `users_threads`
	RENAME COLUMN `ID` TO `id`,
	RENAME COLUMN `idUser` TO `user_id`,
	RENAME COLUMN `idThread` TO `thread_id`,
	RENAME COLUMN `sLastReadDate` TO `last_read_date`,
	RENAME COLUMN `bParticipated` TO `participated`,
	RENAME COLUMN `sLastWriteDate` TO `last_write_date`,
	RENAME COLUMN `bStarred` TO `starred`,
	RENAME COLUMN `iVersion` TO `version`,
	RENAME INDEX `userThread` TO `user_thread`,
	RENAME INDEX `iVersion` TO `version`;


DROP TRIGGER `before_insert_filters`;
-- +migrate StatementBegin
CREATE TRIGGER `before_insert_filters` BEFORE INSERT ON `filters` FOR EACH ROW BEGIN IF (NEW.id IS NULL OR NEW.id = 0) THEN SET NEW.id = FLOOR(RAND() * 1000000000) + FLOOR(RAND() * 1000000000) * 1000000000; END IF ; SELECT ROUND(UNIX_TIMESTAMP(CURTIME(2)) * 10) INTO @curVersion;SET NEW.version = @curVersion; END
-- +migrate StatementEnd
DROP TRIGGER `after_insert_filters`;
-- +migrate StatementBegin
CREATE TRIGGER `after_insert_filters` AFTER INSERT ON `filters` FOR EACH ROW BEGIN INSERT INTO `history_filters` (`id`,`version`,`user_id`,`name`,`selected`,`starred`,`start_date`,`end_date`,`archived`,`participated`,`unread`,`item_id`,`group_id`,`older_than`,`newer_than`,`users_search`,`body_search`,`important`) VALUES (NEW.`id`,@curVersion,NEW.`user_id`,NEW.`name`,NEW.`selected`,NEW.`starred`,NEW.`start_date`,NEW.`end_date`,NEW.`archived`,NEW.`participated`,NEW.`unread`,NEW.`item_id`,NEW.`group_id`,NEW.`older_than`,NEW.`newer_than`,NEW.`users_search`,NEW.`body_search`,NEW.`important`); END
-- +migrate StatementEnd
DROP TRIGGER `before_update_filters`;
-- +migrate StatementBegin
CREATE TRIGGER `before_update_filters` BEFORE UPDATE ON `filters` FOR EACH ROW BEGIN IF NEW.version <> OLD.version THEN SET @curVersion = NEW.version; ELSE SELECT ROUND(UNIX_TIMESTAMP(CURTIME(2)) * 10) INTO @curVersion; END IF; IF NOT (OLD.`id` = NEW.`id` AND OLD.`user_id` <=> NEW.`user_id` AND OLD.`name` <=> NEW.`name` AND OLD.`starred` <=> NEW.`starred` AND OLD.`start_date` <=> NEW.`start_date` AND OLD.`end_date` <=> NEW.`end_date` AND OLD.`archived` <=> NEW.`archived` AND OLD.`participated` <=> NEW.`participated` AND OLD.`unread` <=> NEW.`unread` AND OLD.`item_id` <=> NEW.`item_id` AND OLD.`group_id` <=> NEW.`group_id` AND OLD.`older_than` <=> NEW.`older_than` AND OLD.`newer_than` <=> NEW.`newer_than` AND OLD.`users_search` <=> NEW.`users_search` AND OLD.`body_search` <=> NEW.`body_search` AND OLD.`important` <=> NEW.`important`) THEN   SET NEW.version = @curVersion;   UPDATE `history_filters` SET `next_version` = @curVersion WHERE `id` = OLD.`id` AND `next_version` IS NULL;   INSERT INTO `history_filters` (`id`,`version`,`user_id`,`name`,`selected`,`starred`,`start_date`,`end_date`,`archived`,`participated`,`unread`,`item_id`,`group_id`,`older_than`,`newer_than`,`users_search`,`body_search`,`important`)       VALUES (NEW.`id`,@curVersion,NEW.`user_id`,NEW.`name`,NEW.`selected`,NEW.`starred`,NEW.`start_date`,NEW.`end_date`,NEW.`archived`,NEW.`participated`,NEW.`unread`,NEW.`item_id`,NEW.`group_id`,NEW.`older_than`,NEW.`newer_than`,NEW.`users_search`,NEW.`body_search`,NEW.`important`) ; END IF; END
-- +migrate StatementEnd
DROP TRIGGER `before_delete_filters`;
-- +migrate StatementBegin
CREATE TRIGGER `before_delete_filters` BEFORE DELETE ON `filters` FOR EACH ROW BEGIN SELECT ROUND(UNIX_TIMESTAMP(CURTIME(2)) * 10) INTO @curVersion; UPDATE `history_filters` SET `next_version` = @curVersion WHERE `id` = OLD.`id` AND `next_version` IS NULL; INSERT INTO `history_filters` (`id`,`version`,`user_id`,`name`,`selected`,`starred`,`start_date`,`end_date`,`archived`,`participated`,`unread`,`item_id`,`group_id`,`older_than`,`newer_than`,`users_search`,`body_search`,`important`, `deleted`) VALUES (OLD.`id`,@curVersion,OLD.`user_id`,OLD.`name`,OLD.`selected`,OLD.`starred`,OLD.`start_date`,OLD.`end_date`,OLD.`archived`,OLD.`participated`,OLD.`unread`,OLD.`item_id`,OLD.`group_id`,OLD.`older_than`,OLD.`newer_than`,OLD.`users_search`,OLD.`body_search`,OLD.`important`, 1); END
-- +migrate StatementEnd
DROP TRIGGER `before_insert_groups`;
-- +migrate StatementBegin
CREATE TRIGGER `before_insert_groups` BEFORE INSERT ON `groups` FOR EACH ROW BEGIN IF (NEW.id IS NULL OR NEW.id = 0) THEN SET NEW.id = FLOOR(RAND() * 1000000000) + FLOOR(RAND() * 1000000000) * 1000000000; END IF ; SELECT ROUND(UNIX_TIMESTAMP(CURTIME(2)) * 10) INTO @curVersion;SET NEW.version = @curVersion; END
-- +migrate StatementEnd
DROP TRIGGER `after_insert_groups`;
-- +migrate StatementBegin
CREATE TRIGGER `after_insert_groups` AFTER INSERT ON `groups` FOR EACH ROW BEGIN INSERT INTO `history_groups` (`id`,`version`,`name`,`grade`,`grade_details`,`description`,`date_created`,`opened`,`free_access`,`team_item_id`,`team_participating`,`code`,`code_timer`,`code_end`,`redirect_path`,`open_contest`,`type`,`send_emails`) VALUES (NEW.`id`,@curVersion,NEW.`name`,NEW.`grade`,NEW.`grade_details`,NEW.`description`,NEW.`date_created`,NEW.`opened`,NEW.`free_access`,NEW.`team_item_id`,NEW.`team_participating`,NEW.`code`,NEW.`code_timer`,NEW.`code_end`,NEW.`redirect_path`,NEW.`open_contest`,NEW.`type`,NEW.`send_emails`); INSERT IGNORE INTO `groups_propagate` (`id`, `ancestors_computation_state`) VALUES (NEW.`id`, 'todo') ; END
-- +migrate StatementEnd
DROP TRIGGER `before_update_groups`;
-- +migrate StatementBegin
CREATE TRIGGER `before_update_groups` BEFORE UPDATE ON `groups` FOR EACH ROW BEGIN IF NEW.version <> OLD.version THEN SET @curVersion = NEW.version; ELSE SELECT ROUND(UNIX_TIMESTAMP(CURTIME(2)) * 10) INTO @curVersion; END IF; IF NOT (OLD.`id` = NEW.`id` AND OLD.`name` <=> NEW.`name` AND OLD.`grade` <=> NEW.`grade` AND OLD.`grade_details` <=> NEW.`grade_details` AND OLD.`description` <=> NEW.`description` AND OLD.`date_created` <=> NEW.`date_created` AND OLD.`opened` <=> NEW.`opened` AND OLD.`free_access` <=> NEW.`free_access` AND OLD.`team_item_id` <=> NEW.`team_item_id` AND OLD.`team_participating` <=> NEW.`team_participating` AND OLD.`code` <=> NEW.`code` AND OLD.`code_timer` <=> NEW.`code_timer` AND OLD.`code_end` <=> NEW.`code_end` AND OLD.`redirect_path` <=> NEW.`redirect_path` AND OLD.`open_contest` <=> NEW.`open_contest` AND OLD.`type` <=> NEW.`type` AND OLD.`send_emails` <=> NEW.`send_emails`) THEN   SET NEW.version = @curVersion;   UPDATE `history_groups` SET `next_version` = @curVersion WHERE `id` = OLD.`id` AND `next_version` IS NULL;   INSERT INTO `history_groups` (`id`,`version`,`name`,`grade`,`grade_details`,`description`,`date_created`,`opened`,`free_access`,`team_item_id`,`team_participating`,`code`,`code_timer`,`code_end`,`redirect_path`,`open_contest`,`type`,`send_emails`)       VALUES (NEW.`id`,@curVersion,NEW.`name`,NEW.`grade`,NEW.`grade_details`,NEW.`description`,NEW.`date_created`,NEW.`opened`,NEW.`free_access`,NEW.`team_item_id`,NEW.`team_participating`,NEW.`code`,NEW.`code_timer`,NEW.`code_end`,NEW.`redirect_path`,NEW.`open_contest`,NEW.`type`,NEW.`send_emails`) ; END IF; END
-- +migrate StatementEnd
DROP TRIGGER `before_delete_groups`;
-- +migrate StatementBegin
CREATE TRIGGER `before_delete_groups` BEFORE DELETE ON `groups` FOR EACH ROW BEGIN SELECT ROUND(UNIX_TIMESTAMP(CURTIME(2)) * 10) INTO @curVersion; UPDATE `history_groups` SET `next_version` = @curVersion WHERE `id` = OLD.`id` AND `next_version` IS NULL; INSERT INTO `history_groups` (`id`,`version`,`name`,`grade`,`grade_details`,`description`,`date_created`,`opened`,`free_access`,`team_item_id`,`team_participating`,`code`,`code_timer`,`code_end`,`redirect_path`,`open_contest`,`type`,`send_emails`, `deleted`) VALUES (OLD.`id`,@curVersion,OLD.`name`,OLD.`grade`,OLD.`grade_details`,OLD.`description`,OLD.`date_created`,OLD.`opened`,OLD.`free_access`,OLD.`team_item_id`,OLD.`team_participating`,OLD.`code`,OLD.`code_timer`,OLD.`code_end`,OLD.`redirect_path`,OLD.`open_contest`,OLD.`type`,OLD.`send_emails`, 1); END
-- +migrate StatementEnd
DROP TRIGGER `after_delete_groups`;
-- +migrate StatementBegin
CREATE TRIGGER `after_delete_groups` AFTER DELETE ON `groups` FOR EACH ROW BEGIN DELETE FROM groups_propagate where id = OLD.id ; END
-- +migrate StatementEnd
DROP TRIGGER `before_insert_groups_ancestors`;
-- +migrate StatementBegin
CREATE TRIGGER `before_insert_groups_ancestors` BEFORE INSERT ON `groups_ancestors` FOR EACH ROW BEGIN IF (NEW.id IS NULL OR NEW.id = 0) THEN SET NEW.id = FLOOR(RAND() * 1000000000) + FLOOR(RAND() * 1000000000) * 1000000000; END IF ; SELECT ROUND(UNIX_TIMESTAMP(CURTIME(2)) * 10) INTO @curVersion;SET NEW.version = @curVersion; END
-- +migrate StatementEnd
DROP TRIGGER `after_insert_groups_ancestors`;
-- +migrate StatementBegin
CREATE TRIGGER `after_insert_groups_ancestors` AFTER INSERT ON `groups_ancestors` FOR EACH ROW BEGIN INSERT INTO `history_groups_ancestors` (`id`,`version`,`ancestor_group_id`,`child_group_id`,`is_self`) VALUES (NEW.`id`,@curVersion,NEW.`ancestor_group_id`,NEW.`child_group_id`,NEW.`is_self`); END
-- +migrate StatementEnd
DROP TRIGGER `before_update_groups_ancestors`;
-- +migrate StatementBegin
CREATE TRIGGER `before_update_groups_ancestors` BEFORE UPDATE ON `groups_ancestors` FOR EACH ROW BEGIN IF NEW.version <> OLD.version THEN SET @curVersion = NEW.version; ELSE SELECT ROUND(UNIX_TIMESTAMP(CURTIME(2)) * 10) INTO @curVersion; END IF; IF NOT (OLD.`id` = NEW.`id` AND OLD.`ancestor_group_id` <=> NEW.`ancestor_group_id` AND OLD.`child_group_id` <=> NEW.`child_group_id` AND OLD.`is_self` <=> NEW.`is_self`) THEN   SET NEW.version = @curVersion;   UPDATE `history_groups_ancestors` SET `next_version` = @curVersion WHERE `id` = OLD.`id` AND `next_version` IS NULL;   INSERT INTO `history_groups_ancestors` (`id`,`version`,`ancestor_group_id`,`child_group_id`,`is_self`)       VALUES (NEW.`id`,@curVersion,NEW.`ancestor_group_id`,NEW.`child_group_id`,NEW.`is_self`) ; END IF; END
-- +migrate StatementEnd
DROP TRIGGER `before_delete_groups_ancestors`;
-- +migrate StatementBegin
CREATE TRIGGER `before_delete_groups_ancestors` BEFORE DELETE ON `groups_ancestors` FOR EACH ROW BEGIN SELECT ROUND(UNIX_TIMESTAMP(CURTIME(2)) * 10) INTO @curVersion; UPDATE `history_groups_ancestors` SET `next_version` = @curVersion WHERE `id` = OLD.`id` AND `next_version` IS NULL; INSERT INTO `history_groups_ancestors` (`id`,`version`,`ancestor_group_id`,`child_group_id`,`is_self`, `deleted`) VALUES (OLD.`id`,@curVersion,OLD.`ancestor_group_id`,OLD.`child_group_id`,OLD.`is_self`, 1); END
-- +migrate StatementEnd
DROP TRIGGER `before_insert_groups_attempts`;
-- +migrate StatementBegin
CREATE TRIGGER `before_insert_groups_attempts` BEFORE INSERT ON `groups_attempts` FOR EACH ROW BEGIN IF (NEW.id IS NULL OR NEW.id = 0) THEN SET NEW.id = FLOOR(RAND() * 1000000000) + FLOOR(RAND() * 1000000000) * 1000000000; END IF ; SELECT ROUND(UNIX_TIMESTAMP(CURTIME(2)) * 10) INTO @curVersion;SET NEW.version = @curVersion; SET NEW.minus_score = -NEW.score; END
-- +migrate StatementEnd
DROP TRIGGER `after_insert_groups_attempts`;
-- +migrate StatementBegin
CREATE TRIGGER `after_insert_groups_attempts` AFTER INSERT ON `groups_attempts` FOR EACH ROW BEGIN INSERT INTO `history_groups_attempts` (`id`,`version`,`group_id`,`item_id`,`creator_user_id`,`order`,`score`,`score_computed`,`score_reeval`,`score_diff_manual`,`score_diff_comment`,`submissions_attempts`,`tasks_tried`,`children_validated`,`validated`,`finished`,`key_obtained`,`tasks_with_help`,`hints_requested`,`hints_cached`,`corrections_read`,`precision`,`autonomy`,`start_date`,`validation_date`,`best_answer_date`,`last_answer_date`,`thread_start_date`,`last_hint_date`,`finish_date`,`last_activity_date`,`contest_start_date`,`ranked`,`all_lang_prog`) VALUES (NEW.`id`,@curVersion,NEW.`group_id`,NEW.`item_id`,NEW.`creator_user_id`,NEW.`order`,NEW.`score`,NEW.`score_computed`,NEW.`score_reeval`,NEW.`score_diff_manual`,NEW.`score_diff_comment`,NEW.`submissions_attempts`,NEW.`tasks_tried`,NEW.`children_validated`,NEW.`validated`,NEW.`finished`,NEW.`key_obtained`,NEW.`tasks_with_help`,NEW.`hints_requested`,NEW.`hints_cached`,NEW.`corrections_read`,NEW.`precision`,NEW.`autonomy`,NEW.`start_date`,NEW.`validation_date`,NEW.`best_answer_date`,NEW.`last_answer_date`,NEW.`thread_start_date`,NEW.`last_hint_date`,NEW.`finish_date`,NEW.`last_activity_date`,NEW.`contest_start_date`,NEW.`ranked`,NEW.`all_lang_prog`); END
-- +migrate StatementEnd
DROP TRIGGER `before_update_groups_attempts`;
-- +migrate StatementBegin
CREATE TRIGGER `before_update_groups_attempts` BEFORE UPDATE ON `groups_attempts` FOR EACH ROW BEGIN IF NEW.version <> OLD.version THEN SET @curVersion = NEW.version; ELSE SELECT ROUND(UNIX_TIMESTAMP(CURTIME(2)) * 10) INTO @curVersion; END IF; IF NOT (OLD.`id` = NEW.`id` AND OLD.`group_id` <=> NEW.`group_id` AND OLD.`item_id` <=> NEW.`item_id` AND OLD.`creator_user_id` <=> NEW.`creator_user_id` AND OLD.`order` <=> NEW.`order` AND OLD.`score` <=> NEW.`score` AND OLD.`score_computed` <=> NEW.`score_computed` AND OLD.`score_reeval` <=> NEW.`score_reeval` AND OLD.`score_diff_manual` <=> NEW.`score_diff_manual` AND OLD.`score_diff_comment` <=> NEW.`score_diff_comment` AND OLD.`submissions_attempts` <=> NEW.`submissions_attempts` AND OLD.`tasks_tried` <=> NEW.`tasks_tried` AND OLD.`children_validated` <=> NEW.`children_validated` AND OLD.`validated` <=> NEW.`validated` AND OLD.`finished` <=> NEW.`finished` AND OLD.`key_obtained` <=> NEW.`key_obtained` AND OLD.`tasks_with_help` <=> NEW.`tasks_with_help` AND OLD.`hints_requested` <=> NEW.`hints_requested` AND OLD.`hints_cached` <=> NEW.`hints_cached` AND OLD.`corrections_read` <=> NEW.`corrections_read` AND OLD.`precision` <=> NEW.`precision` AND OLD.`autonomy` <=> NEW.`autonomy` AND OLD.`start_date` <=> NEW.`start_date` AND OLD.`validation_date` <=> NEW.`validation_date` AND OLD.`best_answer_date` <=> NEW.`best_answer_date` AND OLD.`last_answer_date` <=> NEW.`last_answer_date` AND OLD.`thread_start_date` <=> NEW.`thread_start_date` AND OLD.`last_hint_date` <=> NEW.`last_hint_date` AND OLD.`finish_date` <=> NEW.`finish_date` AND OLD.`contest_start_date` <=> NEW.`contest_start_date` AND OLD.`ranked` <=> NEW.`ranked` AND OLD.`all_lang_prog` <=> NEW.`all_lang_prog`) THEN   SET NEW.version = @curVersion;   UPDATE `history_groups_attempts` SET `next_version` = @curVersion WHERE `id` = OLD.`id` AND `next_version` IS NULL;   INSERT INTO `history_groups_attempts` (`id`,`version`,`group_id`,`item_id`,`creator_user_id`,`order`,`score`,`score_computed`,`score_reeval`,`score_diff_manual`,`score_diff_comment`,`submissions_attempts`,`tasks_tried`,`children_validated`,`validated`,`finished`,`key_obtained`,`tasks_with_help`,`hints_requested`,`hints_cached`,`corrections_read`,`precision`,`autonomy`,`start_date`,`validation_date`,`best_answer_date`,`last_answer_date`,`thread_start_date`,`last_hint_date`,`finish_date`,`last_activity_date`,`contest_start_date`,`ranked`,`all_lang_prog`)       VALUES (NEW.`id`,@curVersion,NEW.`group_id`,NEW.`item_id`,NEW.`creator_user_id`,NEW.`order`,NEW.`score`,NEW.`score_computed`,NEW.`score_reeval`,NEW.`score_diff_manual`,NEW.`score_diff_comment`,NEW.`submissions_attempts`,NEW.`tasks_tried`,NEW.`children_validated`,NEW.`validated`,NEW.`finished`,NEW.`key_obtained`,NEW.`tasks_with_help`,NEW.`hints_requested`,NEW.`hints_cached`,NEW.`corrections_read`,NEW.`precision`,NEW.`autonomy`,NEW.`start_date`,NEW.`validation_date`,NEW.`best_answer_date`,NEW.`last_answer_date`,NEW.`thread_start_date`,NEW.`last_hint_date`,NEW.`finish_date`,NEW.`last_activity_date`,NEW.`contest_start_date`,NEW.`ranked`,NEW.`all_lang_prog`) ; SET NEW.minus_score = -NEW.score; END IF; END
-- +migrate StatementEnd
DROP TRIGGER `before_delete_groups_attempts`;
-- +migrate StatementBegin
CREATE TRIGGER `before_delete_groups_attempts` BEFORE DELETE ON `groups_attempts` FOR EACH ROW BEGIN SELECT ROUND(UNIX_TIMESTAMP(CURTIME(2)) * 10) INTO @curVersion; UPDATE `history_groups_attempts` SET `next_version` = @curVersion WHERE `id` = OLD.`id` AND `next_version` IS NULL; INSERT INTO `history_groups_attempts` (`id`,`version`,`group_id`,`item_id`,`creator_user_id`,`order`,`score`,`score_computed`,`score_reeval`,`score_diff_manual`,`score_diff_comment`,`submissions_attempts`,`tasks_tried`,`children_validated`,`validated`,`finished`,`key_obtained`,`tasks_with_help`,`hints_requested`,`hints_cached`,`corrections_read`,`precision`,`autonomy`,`start_date`,`validation_date`,`best_answer_date`,`last_answer_date`,`thread_start_date`,`last_hint_date`,`finish_date`,`last_activity_date`,`contest_start_date`,`ranked`,`all_lang_prog`, `deleted`) VALUES (OLD.`id`,@curVersion,OLD.`group_id`,OLD.`item_id`,OLD.`creator_user_id`,OLD.`order`,OLD.`score`,OLD.`score_computed`,OLD.`score_reeval`,OLD.`score_diff_manual`,OLD.`score_diff_comment`,OLD.`submissions_attempts`,OLD.`tasks_tried`,OLD.`children_validated`,OLD.`validated`,OLD.`finished`,OLD.`key_obtained`,OLD.`tasks_with_help`,OLD.`hints_requested`,OLD.`hints_cached`,OLD.`corrections_read`,OLD.`precision`,OLD.`autonomy`,OLD.`start_date`,OLD.`validation_date`,OLD.`best_answer_date`,OLD.`last_answer_date`,OLD.`thread_start_date`,OLD.`last_hint_date`,OLD.`finish_date`,OLD.`last_activity_date`,OLD.`contest_start_date`,OLD.`ranked`,OLD.`all_lang_prog`, 1); END
-- +migrate StatementEnd
DROP TRIGGER `before_insert_groups_groups`;
-- +migrate StatementBegin
CREATE TRIGGER `before_insert_groups_groups` BEFORE INSERT ON `groups_groups` FOR EACH ROW BEGIN IF (NEW.id IS NULL OR NEW.id = 0) THEN SET NEW.id = FLOOR(RAND() * 1000000000) + FLOOR(RAND() * 1000000000) * 1000000000; END IF ; SELECT ROUND(UNIX_TIMESTAMP(CURTIME(2)) * 10) INTO @curVersion;SET NEW.version = @curVersion; INSERT IGNORE INTO `groups_propagate` (id, ancestors_computation_state) VALUES (NEW.child_group_id, 'todo') ON DUPLICATE KEY UPDATE `ancestors_computation_state` = 'todo' ; END
-- +migrate StatementEnd
DROP TRIGGER `after_insert_groups_groups`;
-- +migrate StatementBegin
CREATE TRIGGER `after_insert_groups_groups` AFTER INSERT ON `groups_groups` FOR EACH ROW BEGIN INSERT INTO `history_groups_groups` (`id`,`version`,`parent_group_id`,`child_group_id`,`child_order`,`type`,`role`,`status_date`,`inviting_user_id`) VALUES (NEW.`id`,@curVersion,NEW.`parent_group_id`,NEW.`child_group_id`,NEW.`child_order`,NEW.`type`,NEW.`role`,NEW.`status_date`,NEW.`inviting_user_id`); END
-- +migrate StatementEnd
DROP TRIGGER `before_update_groups_groups`;
-- +migrate StatementBegin
CREATE TRIGGER `before_update_groups_groups` BEFORE UPDATE ON `groups_groups` FOR EACH ROW BEGIN
  IF NEW.version <> OLD.version THEN
    SET @curVersion = NEW.version;
  ELSE
    SELECT ROUND(UNIX_TIMESTAMP(CURTIME(2)) * 10) INTO @curVersion;
  END IF;
  IF NOT (OLD.`id` = NEW.`id` AND OLD.`parent_group_id` <=> NEW.`parent_group_id` AND
          OLD.`child_group_id` <=> NEW.`child_group_id` AND OLD.`child_order` <=> NEW.`child_order`AND
          OLD.`type` <=> NEW.`type` AND OLD.`role` <=> NEW.`role` AND OLD.`status_date` <=> NEW.`status_date` AND
          OLD.`inviting_user_id` <=> NEW.`inviting_user_id`) THEN
    SET NEW.version = @curVersion;
    UPDATE `history_groups_groups` SET `next_version` = @curVersion WHERE `id` = OLD.`id` AND `next_version` IS NULL;
    INSERT INTO `history_groups_groups` (
      `id`,`version`,`parent_group_id`,`child_group_id`,`child_order`,`type`,`role`,`status_date`,`inviting_user_id`
    ) VALUES (
      NEW.`id`,@curVersion,NEW.`parent_group_id`,NEW.`child_group_id`,NEW.`child_order`,NEW.`type`,NEW.`role`,
      NEW.`status_date`,NEW.`inviting_user_id`
    );
  END IF;
  IF (OLD.child_group_id != NEW.child_group_id OR OLD.parent_group_id != NEW.parent_group_id OR OLD.type != NEW.type) THEN
    INSERT IGNORE INTO `groups_propagate` (`id`, `ancestors_computation_state`) VALUES (OLD.child_group_id, 'todo')
      ON DUPLICATE KEY UPDATE `ancestors_computation_state` = 'todo';
    INSERT IGNORE INTO `groups_propagate` (`id`, `ancestors_computation_state`) (
      SELECT `groups_ancestors`.`child_group_id`, 'todo'
        FROM `groups_ancestors`
        WHERE `groups_ancestors`.`ancestor_group_id` = OLD.`child_group_id`
    ) ON DUPLICATE KEY UPDATE `ancestors_computation_state` = 'todo';
    DELETE `groups_ancestors` FROM `groups_ancestors`
      WHERE `groups_ancestors`.`child_group_id` = OLD.`child_group_id` AND
            `groups_ancestors`.`ancestor_group_id` = OLD.`parent_group_id`;
    DELETE `bridges` FROM `groups_ancestors` `child_descendants`
      JOIN `groups_ancestors` `parent_ancestors`
      JOIN `groups_ancestors` `bridges`
        ON (`bridges`.`ancestor_group_id` = `parent_ancestors`.`ancestor_group_id` AND
            `bridges`.`child_group_id` = `child_descendants`.`child_group_id`)
      WHERE `parent_ancestors`.`child_group_id` = OLD.`parent_group_id` AND
            `child_descendants`.`ancestor_group_id` = OLD.`child_group_id`;
    DELETE `child_ancestors` FROM `groups_ancestors` `child_ancestors`
      JOIN `groups_ancestors` `parent_ancestors`
        ON (`child_ancestors`.`child_group_id` = OLD.`child_group_id` AND
            `child_ancestors`.`ancestor_group_id` = `parent_ancestors`.`ancestor_group_id`)
      WHERE `parent_ancestors`.`child_group_id` = OLD.`parent_group_id`;
    DELETE `parent_ancestors` FROM `groups_ancestors` `parent_ancestors`
      JOIN  `groups_ancestors` `child_ancestors`
        ON (`parent_ancestors`.`ancestor_group_id` = OLD.`parent_group_id` AND
            `child_ancestors`.`child_group_id` = `parent_ancestors`.`child_group_id`)
      WHERE `child_ancestors`.`ancestor_group_id` = OLD.`child_group_id`;
  END IF;
  IF (OLD.child_group_id != NEW.child_group_id OR OLD.parent_group_id != NEW.parent_group_id OR OLD.type != NEW.type) THEN
    INSERT IGNORE INTO `groups_propagate` (id, ancestors_computation_state) VALUES (NEW.child_group_id, 'todo')
      ON DUPLICATE KEY UPDATE `ancestors_computation_state` = 'todo';
  END IF;
END
-- +migrate StatementEnd
DROP TRIGGER `before_delete_groups_groups`;
-- +migrate StatementBegin
CREATE TRIGGER `before_delete_groups_groups` BEFORE DELETE ON `groups_groups` FOR EACH ROW BEGIN
  SELECT ROUND(UNIX_TIMESTAMP(CURTIME(2)) * 10) INTO @curVersion;
  UPDATE `history_groups_groups` SET `next_version` = @curVersion WHERE `id` = OLD.`id` AND `next_version` IS NULL;
  INSERT INTO `history_groups_groups` (
    `id`,`version`,`parent_group_id`,`child_group_id`,`child_order`,`type`,`role`,`status_date`,`inviting_user_id`,`deleted`
  ) VALUES (
    OLD.`id`,@curVersion,OLD.`parent_group_id`,OLD.`child_group_id`,OLD.`child_order`,OLD.`type`,OLD.`role`,
    OLD.`status_date`,OLD.`inviting_user_id`, 1
  );
  INSERT IGNORE INTO `groups_propagate` (`id`, `ancestors_computation_state`) VALUES (OLD.child_group_id, 'todo')
    ON DUPLICATE KEY UPDATE `ancestors_computation_state` = 'todo';
  INSERT IGNORE INTO `groups_propagate` (`id`, `ancestors_computation_state`) (
    SELECT `groups_ancestors`.`child_group_id`, 'todo'
      FROM `groups_ancestors`
      WHERE `groups_ancestors`.`ancestor_group_id` = OLD.`child_group_id`
  ) ON DUPLICATE KEY UPDATE `ancestors_computation_state` = 'todo';
  DELETE `groups_ancestors` FROM `groups_ancestors`
    WHERE `groups_ancestors`.`child_group_id` = OLD.`child_group_id` AND
          `groups_ancestors`.`ancestor_group_id` = OLD.`parent_group_id`;
  DELETE `bridges`
    FROM `groups_ancestors` `child_descendants`
    JOIN `groups_ancestors` `parent_ancestors`
    JOIN `groups_ancestors` `bridges`
      ON (`bridges`.`ancestor_group_id` = `parent_ancestors`.`ancestor_group_id` AND
          `bridges`.`child_group_id` = `child_descendants`.`child_group_id`)
    WHERE `parent_ancestors`.`child_group_id` = OLD.`parent_group_id` AND
          `child_descendants`.`ancestor_group_id` = OLD.`child_group_id`;
  DELETE `child_ancestors`
    FROM `groups_ancestors` `child_ancestors`
    JOIN  `groups_ancestors` `parent_ancestors`
      ON (`child_ancestors`.`child_group_id` = OLD.`child_group_id` AND
          `child_ancestors`.`ancestor_group_id` = `parent_ancestors`.`ancestor_group_id`)
    WHERE `parent_ancestors`.`child_group_id` = OLD.`parent_group_id`;
  DELETE `parent_ancestors`
    FROM `groups_ancestors` `parent_ancestors`
    JOIN  `groups_ancestors` `child_ancestors`
      ON (`parent_ancestors`.`ancestor_group_id` = OLD.`parent_group_id` AND
          `child_ancestors`.`child_group_id` = `parent_ancestors`.`child_group_id`)
    WHERE `child_ancestors`.`ancestor_group_id` = OLD.`child_group_id`;
END
-- +migrate StatementEnd
DROP TRIGGER `before_insert_groups_items`;
-- +migrate StatementBegin
CREATE TRIGGER `before_insert_groups_items` BEFORE INSERT ON `groups_items` FOR EACH ROW BEGIN IF (NEW.id IS NULL OR NEW.id = 0) THEN SET NEW.id = FLOOR(RAND() * 1000000000) + FLOOR(RAND() * 1000000000) * 1000000000; END IF ; SELECT ROUND(UNIX_TIMESTAMP(CURTIME(2)) * 10) INTO @curVersion;SET NEW.version = @curVersion; SET NEW.`propagate_access`='self' ; END
-- +migrate StatementEnd
DROP TRIGGER `after_insert_groups_items`;
-- +migrate StatementBegin
CREATE TRIGGER `after_insert_groups_items` AFTER INSERT ON `groups_items` FOR EACH ROW BEGIN INSERT INTO `history_groups_items` (`id`,`version`,`group_id`,`item_id`,`creator_user_id`,`partial_access_date`,`full_access_date`,`access_reason`,`access_solutions_date`,`owner_access`,`manager_access`,`cached_partial_access_date`,`cached_full_access_date`,`cached_access_solutions_date`,`cached_grayed_access_date`,`cached_full_access`,`cached_partial_access`,`cached_access_solutions`,`cached_grayed_access`,`cached_manager_access`,`propagate_access`) VALUES (NEW.`id`,@curVersion,NEW.`group_id`,NEW.`item_id`,NEW.`creator_user_id`,NEW.`partial_access_date`,NEW.`full_access_date`,NEW.`access_reason`,NEW.`access_solutions_date`,NEW.`owner_access`,NEW.`manager_access`,NEW.`cached_partial_access_date`,NEW.`cached_full_access_date`,NEW.`cached_access_solutions_date`,NEW.`cached_grayed_access_date`,NEW.`cached_full_access`,NEW.`cached_partial_access`,NEW.`cached_access_solutions`,NEW.`cached_grayed_access`,NEW.`cached_manager_access`,NEW.`propagate_access`); END
-- +migrate StatementEnd
DROP TRIGGER `before_update_groups_items`;
-- +migrate StatementBegin
CREATE TRIGGER `before_update_groups_items` BEFORE UPDATE ON `groups_items` FOR EACH ROW BEGIN IF NEW.version <> OLD.version THEN SET @curVersion = NEW.version; ELSE SELECT ROUND(UNIX_TIMESTAMP(CURTIME(2)) * 10) INTO @curVersion; END IF; IF NOT (OLD.`id` = NEW.`id` AND OLD.`group_id` <=> NEW.`group_id` AND OLD.`item_id` <=> NEW.`item_id` AND OLD.`creator_user_id` <=> NEW.`creator_user_id` AND OLD.`partial_access_date` <=> NEW.`partial_access_date` AND OLD.`full_access_date` <=> NEW.`full_access_date` AND OLD.`access_reason` <=> NEW.`access_reason` AND OLD.`access_solutions_date` <=> NEW.`access_solutions_date` AND OLD.`owner_access` <=> NEW.`owner_access` AND OLD.`manager_access` <=> NEW.`manager_access` AND OLD.`cached_partial_access_date` <=> NEW.`cached_partial_access_date` AND OLD.`cached_full_access_date` <=> NEW.`cached_full_access_date` AND OLD.`cached_access_solutions_date` <=> NEW.`cached_access_solutions_date` AND OLD.`cached_grayed_access_date` <=> NEW.`cached_grayed_access_date` AND OLD.`cached_full_access` <=> NEW.`cached_full_access` AND OLD.`cached_partial_access` <=> NEW.`cached_partial_access` AND OLD.`cached_access_solutions` <=> NEW.`cached_access_solutions` AND OLD.`cached_grayed_access` <=> NEW.`cached_grayed_access` AND OLD.`cached_manager_access` <=> NEW.`cached_manager_access`) THEN   SET NEW.version = @curVersion;   UPDATE `history_groups_items` SET `next_version` = @curVersion WHERE `id` = OLD.`id` AND `next_version` IS NULL;   INSERT INTO `history_groups_items` (`id`,`version`,`group_id`,`item_id`,`creator_user_id`,`partial_access_date`,`full_access_date`,`access_reason`,`access_solutions_date`,`owner_access`,`manager_access`,`cached_partial_access_date`,`cached_full_access_date`,`cached_access_solutions_date`,`cached_grayed_access_date`,`cached_full_access`,`cached_partial_access`,`cached_access_solutions`,`cached_grayed_access`,`cached_manager_access`,`propagate_access`)       VALUES (NEW.`id`,@curVersion,NEW.`group_id`,NEW.`item_id`,NEW.`creator_user_id`,NEW.`partial_access_date`,NEW.`full_access_date`,NEW.`access_reason`,NEW.`access_solutions_date`,NEW.`owner_access`,NEW.`manager_access`,NEW.`cached_partial_access_date`,NEW.`cached_full_access_date`,NEW.`cached_access_solutions_date`,NEW.`cached_grayed_access_date`,NEW.`cached_full_access`,NEW.`cached_partial_access`,NEW.`cached_access_solutions`,NEW.`cached_grayed_access`,NEW.`cached_manager_access`,NEW.`propagate_access`) ; END IF; IF NOT (NEW.`full_access_date` <=> OLD.`full_access_date`AND NEW.`partial_access_date` <=> OLD.`partial_access_date`AND NEW.`access_solutions_date` <=> OLD.`access_solutions_date`AND NEW.`manager_access` <=> OLD.`manager_access`AND NEW.`access_reason` <=> OLD.`access_reason`)THEN SET NEW.`propagate_access` = 'self'; END IF; END
-- +migrate StatementEnd
DROP TRIGGER `before_delete_groups_items`;
-- +migrate StatementBegin
CREATE TRIGGER `before_delete_groups_items` BEFORE DELETE ON `groups_items` FOR EACH ROW BEGIN SELECT ROUND(UNIX_TIMESTAMP(CURTIME(2)) * 10) INTO @curVersion; UPDATE `history_groups_items` SET `next_version` = @curVersion WHERE `id` = OLD.`id` AND `next_version` IS NULL; INSERT INTO `history_groups_items` (`id`,`version`,`group_id`,`item_id`,`creator_user_id`,`partial_access_date`,`full_access_date`,`access_reason`,`access_solutions_date`,`owner_access`,`manager_access`,`cached_partial_access_date`,`cached_full_access_date`,`cached_access_solutions_date`,`cached_grayed_access_date`,`cached_full_access`,`cached_partial_access`,`cached_access_solutions`,`cached_grayed_access`,`cached_manager_access`,`propagate_access`, `deleted`) VALUES (OLD.`id`,@curVersion,OLD.`group_id`,OLD.`item_id`,OLD.`creator_user_id`,OLD.`partial_access_date`,OLD.`full_access_date`,OLD.`access_reason`,OLD.`access_solutions_date`,OLD.`owner_access`,OLD.`manager_access`,OLD.`cached_partial_access_date`,OLD.`cached_full_access_date`,OLD.`cached_access_solutions_date`,OLD.`cached_grayed_access_date`,OLD.`cached_full_access`,OLD.`cached_partial_access`,OLD.`cached_access_solutions`,OLD.`cached_grayed_access`,OLD.`cached_manager_access`,OLD.`propagate_access`, 1); END
-- +migrate StatementEnd
DROP TRIGGER `after_delete_groups_items`;
-- +migrate StatementBegin
CREATE TRIGGER `after_delete_groups_items` AFTER DELETE ON `groups_items` FOR EACH ROW BEGIN DELETE FROM groups_items_propagate where id = OLD.id ; END
-- +migrate StatementEnd
DROP TRIGGER `before_insert_groups_login_prefixes`;
-- +migrate StatementBegin
CREATE TRIGGER `before_insert_groups_login_prefixes` BEFORE INSERT ON `groups_login_prefixes` FOR EACH ROW BEGIN IF (NEW.id IS NULL OR NEW.id = 0) THEN SET NEW.id = FLOOR(RAND() * 1000000000) + FLOOR(RAND() * 1000000000) * 1000000000; END IF ; SELECT ROUND(UNIX_TIMESTAMP(CURTIME(2)) * 10) INTO @curVersion;SET NEW.version = @curVersion; END
-- +migrate StatementEnd
DROP TRIGGER `after_insert_groups_login_prefixes`;
-- +migrate StatementBegin
CREATE TRIGGER `after_insert_groups_login_prefixes` AFTER INSERT ON `groups_login_prefixes` FOR EACH ROW BEGIN INSERT INTO `history_groups_login_prefixes` (`id`,`version`,`group_id`,`prefix`) VALUES (NEW.`id`,@curVersion,NEW.`group_id`,NEW.`prefix`); END
-- +migrate StatementEnd
DROP TRIGGER `before_update_groups_login_prefixes`;
-- +migrate StatementBegin
CREATE TRIGGER `before_update_groups_login_prefixes` BEFORE UPDATE ON `groups_login_prefixes` FOR EACH ROW BEGIN IF NEW.version <> OLD.version THEN SET @curVersion = NEW.version; ELSE SELECT ROUND(UNIX_TIMESTAMP(CURTIME(2)) * 10) INTO @curVersion; END IF; IF NOT (OLD.`id` = NEW.`id` AND OLD.`group_id` <=> NEW.`group_id` AND OLD.`prefix` <=> NEW.`prefix`) THEN   SET NEW.version = @curVersion;   UPDATE `history_groups_login_prefixes` SET `next_version` = @curVersion WHERE `id` = OLD.`id` AND `next_version` IS NULL;   INSERT INTO `history_groups_login_prefixes` (`id`,`version`,`group_id`,`prefix`)       VALUES (NEW.`id`,@curVersion,NEW.`group_id`,NEW.`prefix`) ; END IF; END
-- +migrate StatementEnd
DROP TRIGGER `before_delete_groups_login_prefixes`;
-- +migrate StatementBegin
CREATE TRIGGER `before_delete_groups_login_prefixes` BEFORE DELETE ON `groups_login_prefixes` FOR EACH ROW BEGIN SELECT ROUND(UNIX_TIMESTAMP(CURTIME(2)) * 10) INTO @curVersion; UPDATE `history_groups_login_prefixes` SET `next_version` = @curVersion WHERE `id` = OLD.`id` AND `next_version` IS NULL; INSERT INTO `history_groups_login_prefixes` (`id`,`version`,`group_id`,`prefix`, `deleted`) VALUES (OLD.`id`,@curVersion,OLD.`group_id`,OLD.`prefix`, 1); END
-- +migrate StatementEnd
DROP TRIGGER `before_insert_items`;
-- +migrate StatementBegin
CREATE TRIGGER `before_insert_items` BEFORE INSERT ON `items` FOR EACH ROW BEGIN IF (NEW.id IS NULL OR NEW.id = 0) THEN SET NEW.id = FLOOR(RAND() * 1000000000) + FLOOR(RAND() * 1000000000) * 1000000000; END IF ; SELECT ROUND(UNIX_TIMESTAMP(CURTIME(2)) * 10) INTO @curVersion;SET NEW.version = @curVersion; SELECT platforms.id INTO @platformID FROM platforms WHERE NEW.url REGEXP platforms.regexp ORDER BY platforms.priority DESC LIMIT 1 ; SET NEW.platform_id=@platformID ; END
-- +migrate StatementEnd
DROP TRIGGER `after_insert_items`;
-- +migrate StatementBegin
CREATE TRIGGER `after_insert_items` AFTER INSERT ON `items` FOR EACH ROW BEGIN INSERT INTO `history_items` (`id`,`version`,`url`,`platform_id`,`text_id`,`repository_path`,`type`,`uses_api`,`read_only`,`full_screen`,`show_difficulty`,`show_source`,`hints_allowed`,`fixed_ranks`,`validation_type`,`validation_min`,`preparation_state`,`unlocked_item_ids`,`score_min_unlock`,`supported_lang_prog`,`default_language_id`,`team_mode`,`teams_editable`,`qualified_group_id`,`team_max_members`,`has_attempts`,`access_open_date`,`duration`,`end_contest_date`,`show_user_infos`,`contest_phase`,`level`,`no_score`,`title_bar_visible`,`transparent_folder`,`display_details_in_parent`,`display_children_as_tabs`,`custom_chapter`,`group_code_enter`) VALUES (NEW.`id`,@curVersion,NEW.`url`,NEW.`platform_id`,NEW.`text_id`,NEW.`repository_path`,NEW.`type`,NEW.`uses_api`,NEW.`read_only`,NEW.`full_screen`,NEW.`show_difficulty`,NEW.`show_source`,NEW.`hints_allowed`,NEW.`fixed_ranks`,NEW.`validation_type`,NEW.`validation_min`,NEW.`preparation_state`,NEW.`unlocked_item_ids`,NEW.`score_min_unlock`,NEW.`supported_lang_prog`,NEW.`default_language_id`,NEW.`team_mode`,NEW.`teams_editable`,NEW.`qualified_group_id`,NEW.`team_max_members`,NEW.`has_attempts`,NEW.`access_open_date`,NEW.`duration`,NEW.`end_contest_date`,NEW.`show_user_infos`,NEW.`contest_phase`,NEW.`level`,NEW.`no_score`,NEW.`title_bar_visible`,NEW.`transparent_folder`,NEW.`display_details_in_parent`,NEW.`display_children_as_tabs`,NEW.`custom_chapter`,NEW.`group_code_enter`); INSERT IGNORE INTO `items_propagate` (`id`, `ancestors_computation_state`) VALUES (NEW.`id`, 'todo') ; END
-- +migrate StatementEnd
DROP TRIGGER `before_update_items`;
-- +migrate StatementBegin
CREATE TRIGGER `before_update_items` BEFORE UPDATE ON `items` FOR EACH ROW BEGIN IF NEW.version <> OLD.version THEN SET @curVersion = NEW.version; ELSE SELECT ROUND(UNIX_TIMESTAMP(CURTIME(2)) * 10) INTO @curVersion; END IF; IF NOT (OLD.`id` = NEW.`id` AND OLD.`url` <=> NEW.`url` AND OLD.`platform_id` <=> NEW.`platform_id` AND OLD.`text_id` <=> NEW.`text_id` AND OLD.`repository_path` <=> NEW.`repository_path` AND OLD.`type` <=> NEW.`type` AND OLD.`uses_api` <=> NEW.`uses_api` AND OLD.`read_only` <=> NEW.`read_only` AND OLD.`full_screen` <=> NEW.`full_screen` AND OLD.`show_difficulty` <=> NEW.`show_difficulty` AND OLD.`show_source` <=> NEW.`show_source` AND OLD.`hints_allowed` <=> NEW.`hints_allowed` AND OLD.`fixed_ranks` <=> NEW.`fixed_ranks` AND OLD.`validation_type` <=> NEW.`validation_type` AND OLD.`validation_min` <=> NEW.`validation_min` AND OLD.`preparation_state` <=> NEW.`preparation_state` AND OLD.`unlocked_item_ids` <=> NEW.`unlocked_item_ids` AND OLD.`score_min_unlock` <=> NEW.`score_min_unlock` AND OLD.`supported_lang_prog` <=> NEW.`supported_lang_prog` AND OLD.`default_language_id` <=> NEW.`default_language_id` AND OLD.`team_mode` <=> NEW.`team_mode` AND OLD.`teams_editable` <=> NEW.`teams_editable` AND OLD.`qualified_group_id` <=> NEW.`qualified_group_id` AND OLD.`team_max_members` <=> NEW.`team_max_members` AND OLD.`has_attempts` <=> NEW.`has_attempts` AND OLD.`access_open_date` <=> NEW.`access_open_date` AND OLD.`duration` <=> NEW.`duration` AND OLD.`end_contest_date` <=> NEW.`end_contest_date` AND OLD.`show_user_infos` <=> NEW.`show_user_infos` AND OLD.`contest_phase` <=> NEW.`contest_phase` AND OLD.`level` <=> NEW.`level` AND OLD.`no_score` <=> NEW.`no_score` AND OLD.`title_bar_visible` <=> NEW.`title_bar_visible` AND OLD.`transparent_folder` <=> NEW.`transparent_folder` AND OLD.`display_details_in_parent` <=> NEW.`display_details_in_parent` AND OLD.`display_children_as_tabs` <=> NEW.`display_children_as_tabs` AND OLD.`custom_chapter` <=> NEW.`custom_chapter` AND OLD.`group_code_enter` <=> NEW.`group_code_enter`) THEN   SET NEW.version = @curVersion;   UPDATE `history_items` SET `next_version` = @curVersion WHERE `id` = OLD.`id` AND `next_version` IS NULL;   INSERT INTO `history_items` (`id`,`version`,`url`,`platform_id`,`text_id`,`repository_path`,`type`,`uses_api`,`read_only`,`full_screen`,`show_difficulty`,`show_source`,`hints_allowed`,`fixed_ranks`,`validation_type`,`validation_min`,`preparation_state`,`unlocked_item_ids`,`score_min_unlock`,`supported_lang_prog`,`default_language_id`,`team_mode`,`teams_editable`,`qualified_group_id`,`team_max_members`,`has_attempts`,`access_open_date`,`duration`,`end_contest_date`,`show_user_infos`,`contest_phase`,`level`,`no_score`,`title_bar_visible`,`transparent_folder`,`display_details_in_parent`,`display_children_as_tabs`,`custom_chapter`,`group_code_enter`)       VALUES (NEW.`id`,@curVersion,NEW.`url`,NEW.`platform_id`,NEW.`text_id`,NEW.`repository_path`,NEW.`type`,NEW.`uses_api`,NEW.`read_only`,NEW.`full_screen`,NEW.`show_difficulty`,NEW.`show_source`,NEW.`hints_allowed`,NEW.`fixed_ranks`,NEW.`validation_type`,NEW.`validation_min`,NEW.`preparation_state`,NEW.`unlocked_item_ids`,NEW.`score_min_unlock`,NEW.`supported_lang_prog`,NEW.`default_language_id`,NEW.`team_mode`,NEW.`teams_editable`,NEW.`qualified_group_id`,NEW.`team_max_members`,NEW.`has_attempts`,NEW.`access_open_date`,NEW.`duration`,NEW.`end_contest_date`,NEW.`show_user_infos`,NEW.`contest_phase`,NEW.`level`,NEW.`no_score`,NEW.`title_bar_visible`,NEW.`transparent_folder`,NEW.`display_details_in_parent`,NEW.`display_children_as_tabs`,NEW.`custom_chapter`,NEW.`group_code_enter`) ; END IF; SELECT platforms.id INTO @platformID FROM platforms WHERE NEW.url REGEXP platforms.regexp ORDER BY platforms.priority DESC LIMIT 1 ; SET NEW.platform_id=@platformID ; END
-- +migrate StatementEnd
DROP TRIGGER `before_delete_items`;
-- +migrate StatementBegin
CREATE TRIGGER `before_delete_items` BEFORE DELETE ON `items` FOR EACH ROW BEGIN SELECT ROUND(UNIX_TIMESTAMP(CURTIME(2)) * 10) INTO @curVersion; UPDATE `history_items` SET `next_version` = @curVersion WHERE `id` = OLD.`id` AND `next_version` IS NULL; INSERT INTO `history_items` (`id`,`version`,`url`,`platform_id`,`text_id`,`repository_path`,`type`,`uses_api`,`read_only`,`full_screen`,`show_difficulty`,`show_source`,`hints_allowed`,`fixed_ranks`,`validation_type`,`validation_min`,`preparation_state`,`unlocked_item_ids`,`score_min_unlock`,`supported_lang_prog`,`default_language_id`,`team_mode`,`teams_editable`,`qualified_group_id`,`team_max_members`,`has_attempts`,`access_open_date`,`duration`,`end_contest_date`,`show_user_infos`,`contest_phase`,`level`,`no_score`,`title_bar_visible`,`transparent_folder`,`display_details_in_parent`,`display_children_as_tabs`,`custom_chapter`,`group_code_enter`, `deleted`) VALUES (OLD.`id`,@curVersion,OLD.`url`,OLD.`platform_id`,OLD.`text_id`,OLD.`repository_path`,OLD.`type`,OLD.`uses_api`,OLD.`read_only`,OLD.`full_screen`,OLD.`show_difficulty`,OLD.`show_source`,OLD.`hints_allowed`,OLD.`fixed_ranks`,OLD.`validation_type`,OLD.`validation_min`,OLD.`preparation_state`,OLD.`unlocked_item_ids`,OLD.`score_min_unlock`,OLD.`supported_lang_prog`,OLD.`default_language_id`,OLD.`team_mode`,OLD.`teams_editable`,OLD.`qualified_group_id`,OLD.`team_max_members`,OLD.`has_attempts`,OLD.`access_open_date`,OLD.`duration`,OLD.`end_contest_date`,OLD.`show_user_infos`,OLD.`contest_phase`,OLD.`level`,OLD.`no_score`,OLD.`title_bar_visible`,OLD.`transparent_folder`,OLD.`display_details_in_parent`,OLD.`display_children_as_tabs`,OLD.`custom_chapter`,OLD.`group_code_enter`, 1); END
-- +migrate StatementEnd
DROP TRIGGER `after_delete_items`;
-- +migrate StatementBegin
CREATE TRIGGER `after_delete_items` AFTER DELETE ON `items` FOR EACH ROW BEGIN DELETE FROM items_propagate where id = OLD.id ; END
-- +migrate StatementEnd
DROP TRIGGER `before_insert_items_ancestors`;
-- +migrate StatementBegin
CREATE TRIGGER `before_insert_items_ancestors` BEFORE INSERT ON `items_ancestors` FOR EACH ROW BEGIN IF (NEW.id IS NULL OR NEW.id = 0) THEN SET NEW.id = FLOOR(RAND() * 1000000000) + FLOOR(RAND() * 1000000000) * 1000000000; END IF ; SELECT ROUND(UNIX_TIMESTAMP(CURTIME(2)) * 10) INTO @curVersion;SET NEW.version = @curVersion; END
-- +migrate StatementEnd
DROP TRIGGER `after_insert_items_ancestors`;
-- +migrate StatementBegin
CREATE TRIGGER `after_insert_items_ancestors` AFTER INSERT ON `items_ancestors` FOR EACH ROW BEGIN INSERT INTO `history_items_ancestors` (`id`,`version`,`ancestor_item_id`,`child_item_id`) VALUES (NEW.`id`,@curVersion,NEW.`ancestor_item_id`,NEW.`child_item_id`); END
-- +migrate StatementEnd
DROP TRIGGER `before_update_items_ancestors`;
-- +migrate StatementBegin
CREATE TRIGGER `before_update_items_ancestors` BEFORE UPDATE ON `items_ancestors` FOR EACH ROW BEGIN IF NEW.version <> OLD.version THEN SET @curVersion = NEW.version; ELSE SELECT ROUND(UNIX_TIMESTAMP(CURTIME(2)) * 10) INTO @curVersion; END IF; IF NOT (OLD.`id` = NEW.`id` AND OLD.`ancestor_item_id` <=> NEW.`ancestor_item_id` AND OLD.`child_item_id` <=> NEW.`child_item_id`) THEN   SET NEW.version = @curVersion;   UPDATE `history_items_ancestors` SET `next_version` = @curVersion WHERE `id` = OLD.`id` AND `next_version` IS NULL;   INSERT INTO `history_items_ancestors` (`id`,`version`,`ancestor_item_id`,`child_item_id`)       VALUES (NEW.`id`,@curVersion,NEW.`ancestor_item_id`,NEW.`child_item_id`) ; END IF; END
-- +migrate StatementEnd
DROP TRIGGER `before_delete_items_ancestors`;
-- +migrate StatementBegin
CREATE TRIGGER `before_delete_items_ancestors` BEFORE DELETE ON `items_ancestors` FOR EACH ROW BEGIN SELECT ROUND(UNIX_TIMESTAMP(CURTIME(2)) * 10) INTO @curVersion; UPDATE `history_items_ancestors` SET `next_version` = @curVersion WHERE `id` = OLD.`id` AND `next_version` IS NULL; INSERT INTO `history_items_ancestors` (`id`,`version`,`ancestor_item_id`,`child_item_id`, `deleted`) VALUES (OLD.`id`,@curVersion,OLD.`ancestor_item_id`,OLD.`child_item_id`, 1); END
-- +migrate StatementEnd
DROP TRIGGER `before_insert_items_items`;
-- +migrate StatementBegin
CREATE TRIGGER `before_insert_items_items` BEFORE INSERT ON `items_items` FOR EACH ROW BEGIN IF (NEW.id IS NULL OR NEW.id = 0) THEN SET NEW.id = FLOOR(RAND() * 1000000000) + FLOOR(RAND() * 1000000000) * 1000000000; END IF ; SELECT ROUND(UNIX_TIMESTAMP(CURTIME(2)) * 10) INTO @curVersion;SET NEW.version = @curVersion; INSERT IGNORE INTO `items_propagate` (id, ancestors_computation_state) VALUES (NEW.child_item_id, 'todo') ON DUPLICATE KEY UPDATE `ancestors_computation_state` = 'todo' ; END
-- +migrate StatementEnd
DROP TRIGGER `after_insert_items_items`;
-- +migrate StatementBegin
CREATE TRIGGER `after_insert_items_items` AFTER INSERT ON `items_items` FOR EACH ROW BEGIN INSERT INTO `history_items_items` (`id`,`version`,`parent_item_id`,`child_item_id`,`child_order`,`category`,`access_restricted`,`always_visible`,`difficulty`) VALUES (NEW.`id`,@curVersion,NEW.`parent_item_id`,NEW.`child_item_id`,NEW.`child_order`,NEW.`category`,NEW.`access_restricted`,NEW.`always_visible`,NEW.`difficulty`); INSERT IGNORE INTO `groups_items_propagate` SELECT `id`, 'children' as `propagate_access` FROM `groups_items` WHERE `groups_items`.`item_id` = NEW.`parent_item_id` ON DUPLICATE KEY UPDATE propagate_access='children' ; END
-- +migrate StatementEnd
DROP TRIGGER `before_update_items_items`;
-- +migrate StatementBegin
CREATE TRIGGER `before_update_items_items` BEFORE UPDATE ON `items_items` FOR EACH ROW BEGIN IF NEW.version <> OLD.version THEN SET @curVersion = NEW.version; ELSE SELECT ROUND(UNIX_TIMESTAMP(CURTIME(2)) * 10) INTO @curVersion; END IF; IF NOT (OLD.`id` = NEW.`id` AND OLD.`parent_item_id` <=> NEW.`parent_item_id` AND OLD.`child_item_id` <=> NEW.`child_item_id` AND OLD.`child_order` <=> NEW.`child_order` AND OLD.`category` <=> NEW.`category` AND OLD.`access_restricted` <=> NEW.`access_restricted` AND OLD.`always_visible` <=> NEW.`always_visible` AND OLD.`difficulty` <=> NEW.`difficulty`) THEN   SET NEW.version = @curVersion;   UPDATE `history_items_items` SET `next_version` = @curVersion WHERE `id` = OLD.`id` AND `next_version` IS NULL;   INSERT INTO `history_items_items` (`id`,`version`,`parent_item_id`,`child_item_id`,`child_order`,`category`,`access_restricted`,`always_visible`,`difficulty`)       VALUES (NEW.`id`,@curVersion,NEW.`parent_item_id`,NEW.`child_item_id`,NEW.`child_order`,NEW.`category`,NEW.`access_restricted`,NEW.`always_visible`,NEW.`difficulty`) ; END IF; IF (OLD.child_item_id != NEW.child_item_id OR OLD.parent_item_id != NEW.parent_item_id) THEN INSERT IGNORE INTO `items_propagate` (`id`, `ancestors_computation_state`) VALUES (OLD.child_item_id, 'todo') ON DUPLICATE KEY UPDATE `ancestors_computation_state` = 'todo'; INSERT IGNORE INTO `items_propagate` (`id`, `ancestors_computation_state`) VALUES (OLD.parent_item_id, 'todo') ON DUPLICATE KEY UPDATE `ancestors_computation_state` = 'todo'; INSERT IGNORE INTO `items_propagate` (`id`, `ancestors_computation_state`) (SELECT `items_ancestors`.`child_item_id`, 'todo' FROM `items_ancestors` WHERE `items_ancestors`.`ancestor_item_id` = OLD.`child_item_id`) ON DUPLICATE KEY UPDATE `ancestors_computation_state` = 'todo'; DELETE `items_ancestors` from `items_ancestors` WHERE `items_ancestors`.`child_item_id` = OLD.`child_item_id` and `items_ancestors`.`ancestor_item_id` = OLD.`parent_item_id`;DELETE `bridges` FROM `items_ancestors` `child_descendants` JOIN `items_ancestors` `parent_ancestors` JOIN `items_ancestors` `bridges` ON (`bridges`.`ancestor_item_id` = `parent_ancestors`.`ancestor_item_id` AND `bridges`.`child_item_id` = `child_descendants`.`child_item_id`) WHERE `parent_ancestors`.`child_item_id` = OLD.`parent_item_id` AND `child_descendants`.`ancestor_item_id` = OLD.`child_item_id`; DELETE `child_ancestors` FROM `items_ancestors` `child_ancestors` JOIN  `items_ancestors` `parent_ancestors` ON (`child_ancestors`.`child_item_id` = OLD.`child_item_id` AND `child_ancestors`.`ancestor_item_id` = `parent_ancestors`.`ancestor_item_id`) WHERE `parent_ancestors`.`child_item_id` = OLD.`parent_item_id`; DELETE `parent_ancestors` FROM `items_ancestors` `parent_ancestors` JOIN  `items_ancestors` `child_ancestors` ON (`parent_ancestors`.`ancestor_item_id` = OLD.`parent_item_id` AND `child_ancestors`.`child_item_id` = `parent_ancestors`.`child_item_id`) WHERE `child_ancestors`.`ancestor_item_id` = OLD.`child_item_id`  ; END IF; IF (OLD.child_item_id != NEW.child_item_id OR OLD.parent_item_id != NEW.parent_item_id) THEN INSERT IGNORE INTO `items_propagate` (id, ancestors_computation_state) VALUES (NEW.child_item_id, 'todo') ON DUPLICATE KEY UPDATE `ancestors_computation_state` = 'todo'  ; END IF; END
-- +migrate StatementEnd
DROP TRIGGER `after_update_items_items`;
-- +migrate StatementBegin
CREATE TRIGGER `after_update_items_items` AFTER UPDATE ON `items_items` FOR EACH ROW BEGIN INSERT IGNORE INTO `groups_items_propagate` SELECT `id`, 'children' as `propagate_access` FROM `groups_items` WHERE `groups_items`.`item_id` = NEW.`parent_item_id` OR `groups_items`.`item_id` = OLD.`parent_item_id` ON DUPLICATE KEY UPDATE propagate_access='children' ; END
-- +migrate StatementEnd
DROP TRIGGER `before_delete_items_items`;
-- +migrate StatementBegin
CREATE TRIGGER `before_delete_items_items` BEFORE DELETE ON `items_items` FOR EACH ROW BEGIN SELECT ROUND(UNIX_TIMESTAMP(CURTIME(2)) * 10) INTO @curVersion; UPDATE `history_items_items` SET `next_version` = @curVersion WHERE `id` = OLD.`id` AND `next_version` IS NULL; INSERT INTO `history_items_items` (`id`,`version`,`parent_item_id`,`child_item_id`,`child_order`,`category`,`access_restricted`,`always_visible`,`difficulty`, `deleted`) VALUES (OLD.`id`,@curVersion,OLD.`parent_item_id`,OLD.`child_item_id`,OLD.`child_order`,OLD.`category`,OLD.`access_restricted`,OLD.`always_visible`,OLD.`difficulty`, 1); INSERT IGNORE INTO `items_propagate` (`id`, `ancestors_computation_state`) VALUES (OLD.child_item_id, 'todo') ON DUPLICATE KEY UPDATE `ancestors_computation_state` = 'todo'; INSERT IGNORE INTO `items_propagate` (`id`, `ancestors_computation_state`) VALUES (OLD.parent_item_id, 'todo') ON DUPLICATE KEY UPDATE `ancestors_computation_state` = 'todo'; INSERT IGNORE INTO `items_propagate` (`id`, `ancestors_computation_state`) (SELECT `items_ancestors`.`child_item_id`, 'todo' FROM `items_ancestors` WHERE `items_ancestors`.`ancestor_item_id` = OLD.`child_item_id`) ON DUPLICATE KEY UPDATE `ancestors_computation_state` = 'todo'; DELETE `items_ancestors` from `items_ancestors` WHERE `items_ancestors`.`child_item_id` = OLD.`child_item_id` and `items_ancestors`.`ancestor_item_id` = OLD.`parent_item_id`;DELETE `bridges` FROM `items_ancestors` `child_descendants` JOIN `items_ancestors` `parent_ancestors` JOIN `items_ancestors` `bridges` ON (`bridges`.`ancestor_item_id` = `parent_ancestors`.`ancestor_item_id` AND `bridges`.`child_item_id` = `child_descendants`.`child_item_id`) WHERE `parent_ancestors`.`child_item_id` = OLD.`parent_item_id` AND `child_descendants`.`ancestor_item_id` = OLD.`child_item_id`; DELETE `child_ancestors` FROM `items_ancestors` `child_ancestors` JOIN  `items_ancestors` `parent_ancestors` ON (`child_ancestors`.`child_item_id` = OLD.`child_item_id` AND `child_ancestors`.`ancestor_item_id` = `parent_ancestors`.`ancestor_item_id`) WHERE `parent_ancestors`.`child_item_id` = OLD.`parent_item_id`; DELETE `parent_ancestors` FROM `items_ancestors` `parent_ancestors` JOIN  `items_ancestors` `child_ancestors` ON (`parent_ancestors`.`ancestor_item_id` = OLD.`parent_item_id` AND `child_ancestors`.`child_item_id` = `parent_ancestors`.`child_item_id`) WHERE `child_ancestors`.`ancestor_item_id` = OLD.`child_item_id` ; INSERT IGNORE INTO `groups_items_propagate` SELECT `id`, 'children' as `propagate_access` FROM `groups_items` WHERE `groups_items`.`item_id` = OLD.`parent_item_id` ON DUPLICATE KEY UPDATE propagate_access='children' ; END
-- +migrate StatementEnd
DROP TRIGGER `before_insert_items_strings`;
-- +migrate StatementBegin
CREATE TRIGGER `before_insert_items_strings` BEFORE INSERT ON `items_strings` FOR EACH ROW BEGIN IF (NEW.id IS NULL OR NEW.id = 0) THEN SET NEW.id = FLOOR(RAND() * 1000000000) + FLOOR(RAND() * 1000000000) * 1000000000; END IF ; SELECT ROUND(UNIX_TIMESTAMP(CURTIME(2)) * 10) INTO @curVersion;SET NEW.version = @curVersion; END
-- +migrate StatementEnd
DROP TRIGGER `after_insert_items_strings`;
-- +migrate StatementBegin
CREATE TRIGGER `after_insert_items_strings` AFTER INSERT ON `items_strings` FOR EACH ROW BEGIN INSERT INTO `history_items_strings` (`id`,`version`,`item_id`,`language_id`,`translator`,`title`,`image_url`,`subtitle`,`description`,`edu_comment`,`ranking_comment`) VALUES (NEW.`id`,@curVersion,NEW.`item_id`,NEW.`language_id`,NEW.`translator`,NEW.`title`,NEW.`image_url`,NEW.`subtitle`,NEW.`description`,NEW.`edu_comment`,NEW.`ranking_comment`); END
-- +migrate StatementEnd
DROP TRIGGER `before_update_items_strings`;
-- +migrate StatementBegin
CREATE TRIGGER `before_update_items_strings` BEFORE UPDATE ON `items_strings` FOR EACH ROW BEGIN IF NEW.version <> OLD.version THEN SET @curVersion = NEW.version; ELSE SELECT ROUND(UNIX_TIMESTAMP(CURTIME(2)) * 10) INTO @curVersion; END IF; IF NOT (OLD.`id` = NEW.`id` AND OLD.`item_id` <=> NEW.`item_id` AND OLD.`language_id` <=> NEW.`language_id` AND OLD.`translator` <=> NEW.`translator` AND OLD.`title` <=> NEW.`title` AND OLD.`image_url` <=> NEW.`image_url` AND OLD.`subtitle` <=> NEW.`subtitle` AND OLD.`description` <=> NEW.`description` AND OLD.`edu_comment` <=> NEW.`edu_comment` AND OLD.`ranking_comment` <=> NEW.`ranking_comment`) THEN   SET NEW.version = @curVersion;   UPDATE `history_items_strings` SET `next_version` = @curVersion WHERE `id` = OLD.`id` AND `next_version` IS NULL;   INSERT INTO `history_items_strings` (`id`,`version`,`item_id`,`language_id`,`translator`,`title`,`image_url`,`subtitle`,`description`,`edu_comment`,`ranking_comment`)       VALUES (NEW.`id`,@curVersion,NEW.`item_id`,NEW.`language_id`,NEW.`translator`,NEW.`title`,NEW.`image_url`,NEW.`subtitle`,NEW.`description`,NEW.`edu_comment`,NEW.`ranking_comment`) ; END IF; END
-- +migrate StatementEnd
DROP TRIGGER `before_delete_items_strings`;
-- +migrate StatementBegin
CREATE TRIGGER `before_delete_items_strings` BEFORE DELETE ON `items_strings` FOR EACH ROW BEGIN SELECT ROUND(UNIX_TIMESTAMP(CURTIME(2)) * 10) INTO @curVersion; UPDATE `history_items_strings` SET `next_version` = @curVersion WHERE `id` = OLD.`id` AND `next_version` IS NULL; INSERT INTO `history_items_strings` (`id`,`version`,`item_id`,`language_id`,`translator`,`title`,`image_url`,`subtitle`,`description`,`edu_comment`,`ranking_comment`, `deleted`) VALUES (OLD.`id`,@curVersion,OLD.`item_id`,OLD.`language_id`,OLD.`translator`,OLD.`title`,OLD.`image_url`,OLD.`subtitle`,OLD.`description`,OLD.`edu_comment`,OLD.`ranking_comment`, 1); END
-- +migrate StatementEnd
DROP TRIGGER `before_insert_languages`;
-- +migrate StatementBegin
CREATE TRIGGER `before_insert_languages` BEFORE INSERT ON `languages` FOR EACH ROW BEGIN IF (NEW.id IS NULL OR NEW.id = 0) THEN SET NEW.id = FLOOR(RAND() * 1000000000) + FLOOR(RAND() * 1000000000) * 1000000000; END IF ; SELECT ROUND(UNIX_TIMESTAMP(CURTIME(2)) * 10) INTO @curVersion;SET NEW.version = @curVersion; END
-- +migrate StatementEnd
DROP TRIGGER `after_insert_languages`;
-- +migrate StatementBegin
CREATE TRIGGER `after_insert_languages` AFTER INSERT ON `languages` FOR EACH ROW BEGIN INSERT INTO `history_languages` (`id`,`version`,`name`,`code`) VALUES (NEW.`id`,@curVersion,NEW.`name`,NEW.`code`); END
-- +migrate StatementEnd
DROP TRIGGER `before_update_languages`;
-- +migrate StatementBegin
CREATE TRIGGER `before_update_languages` BEFORE UPDATE ON `languages` FOR EACH ROW BEGIN IF NEW.version <> OLD.version THEN SET @curVersion = NEW.version; ELSE SELECT ROUND(UNIX_TIMESTAMP(CURTIME(2)) * 10) INTO @curVersion; END IF; IF NOT (OLD.`id` = NEW.`id` AND OLD.`name` <=> NEW.`name` AND OLD.`code` <=> NEW.`code`) THEN   SET NEW.version = @curVersion;   UPDATE `history_languages` SET `next_version` = @curVersion WHERE `id` = OLD.`id` AND `next_version` IS NULL;   INSERT INTO `history_languages` (`id`,`version`,`name`,`code`)       VALUES (NEW.`id`,@curVersion,NEW.`name`,NEW.`code`) ; END IF; END
-- +migrate StatementEnd
DROP TRIGGER `before_delete_languages`;
-- +migrate StatementBegin
CREATE TRIGGER `before_delete_languages` BEFORE DELETE ON `languages` FOR EACH ROW BEGIN SELECT ROUND(UNIX_TIMESTAMP(CURTIME(2)) * 10) INTO @curVersion; UPDATE `history_languages` SET `next_version` = @curVersion WHERE `id` = OLD.`id` AND `next_version` IS NULL; INSERT INTO `history_languages` (`id`,`version`,`name`,`code`, `deleted`) VALUES (OLD.`id`,@curVersion,OLD.`name`,OLD.`code`, 1); END
-- +migrate StatementEnd
DROP TRIGGER `before_insert_messages`;
-- +migrate StatementBegin
CREATE TRIGGER `before_insert_messages` BEFORE INSERT ON `messages` FOR EACH ROW BEGIN IF (NEW.id IS NULL OR NEW.id = 0) THEN SET NEW.id = FLOOR(RAND() * 1000000000) + FLOOR(RAND() * 1000000000) * 1000000000; END IF ; SELECT ROUND(UNIX_TIMESTAMP(CURTIME(2)) * 10) INTO @curVersion;SET NEW.version = @curVersion; END
-- +migrate StatementEnd
DROP TRIGGER `after_insert_messages`;
-- +migrate StatementBegin
CREATE TRIGGER `after_insert_messages` AFTER INSERT ON `messages` FOR EACH ROW BEGIN INSERT INTO `history_messages` (`id`,`version`,`thread_id`,`user_id`,`submission_date`,`published`,`title`,`body`,`trainers_only`,`archived`,`persistant`) VALUES (NEW.`id`,@curVersion,NEW.`thread_id`,NEW.`user_id`,NEW.`submission_date`,NEW.`published`,NEW.`title`,NEW.`body`,NEW.`trainers_only`,NEW.`archived`,NEW.`persistant`); END
-- +migrate StatementEnd
DROP TRIGGER `before_update_messages`;
-- +migrate StatementBegin
CREATE TRIGGER `before_update_messages` BEFORE UPDATE ON `messages` FOR EACH ROW BEGIN IF NEW.version <> OLD.version THEN SET @curVersion = NEW.version; ELSE SELECT ROUND(UNIX_TIMESTAMP(CURTIME(2)) * 10) INTO @curVersion; END IF; IF NOT (OLD.`id` = NEW.`id` AND OLD.`thread_id` <=> NEW.`thread_id` AND OLD.`user_id` <=> NEW.`user_id` AND OLD.`submission_date` <=> NEW.`submission_date` AND OLD.`published` <=> NEW.`published` AND OLD.`title` <=> NEW.`title` AND OLD.`body` <=> NEW.`body` AND OLD.`trainers_only` <=> NEW.`trainers_only` AND OLD.`archived` <=> NEW.`archived` AND OLD.`persistant` <=> NEW.`persistant`) THEN   SET NEW.version = @curVersion;   UPDATE `history_messages` SET `next_version` = @curVersion WHERE `id` = OLD.`id` AND `next_version` IS NULL;   INSERT INTO `history_messages` (`id`,`version`,`thread_id`,`user_id`,`submission_date`,`published`,`title`,`body`,`trainers_only`,`archived`,`persistant`)       VALUES (NEW.`id`,@curVersion,NEW.`thread_id`,NEW.`user_id`,NEW.`submission_date`,NEW.`published`,NEW.`title`,NEW.`body`,NEW.`trainers_only`,NEW.`archived`,NEW.`persistant`) ; END IF; END
-- +migrate StatementEnd
DROP TRIGGER `before_delete_messages`;
-- +migrate StatementBegin
CREATE TRIGGER `before_delete_messages` BEFORE DELETE ON `messages` FOR EACH ROW BEGIN SELECT ROUND(UNIX_TIMESTAMP(CURTIME(2)) * 10) INTO @curVersion; UPDATE `history_messages` SET `next_version` = @curVersion WHERE `id` = OLD.`id` AND `next_version` IS NULL; INSERT INTO `history_messages` (`id`,`version`,`thread_id`,`user_id`,`submission_date`,`published`,`title`,`body`,`trainers_only`,`archived`,`persistant`, `deleted`) VALUES (OLD.`id`,@curVersion,OLD.`thread_id`,OLD.`user_id`,OLD.`submission_date`,OLD.`published`,OLD.`title`,OLD.`body`,OLD.`trainers_only`,OLD.`archived`,OLD.`persistant`, 1); END
-- +migrate StatementEnd
DROP TRIGGER `before_insert_threads`;
-- +migrate StatementBegin
CREATE TRIGGER `before_insert_threads` BEFORE INSERT ON `threads` FOR EACH ROW BEGIN IF (NEW.id IS NULL OR NEW.id = 0) THEN SET NEW.id = FLOOR(RAND() * 1000000000) + FLOOR(RAND() * 1000000000) * 1000000000; END IF ; SELECT ROUND(UNIX_TIMESTAMP(CURTIME(2)) * 10) INTO @curVersion;SET NEW.version = @curVersion; END
-- +migrate StatementEnd
DROP TRIGGER `after_insert_threads`;
-- +migrate StatementBegin
CREATE TRIGGER `after_insert_threads` AFTER INSERT ON `threads` FOR EACH ROW BEGIN INSERT INTO `history_threads` (`id`,`version`,`type`,`creator_user_id`,`item_id`,`title`,`admin_help_asked`,`hidden`,`last_activity_date`) VALUES (NEW.`id`,@curVersion,NEW.`type`,NEW.`creator_user_id`,NEW.`item_id`,NEW.`title`,NEW.`admin_help_asked`,NEW.`hidden`,NEW.`last_activity_date`); END
-- +migrate StatementEnd
DROP TRIGGER `before_update_threads`;
-- +migrate StatementBegin
CREATE TRIGGER `before_update_threads` BEFORE UPDATE ON `threads` FOR EACH ROW BEGIN IF NEW.version <> OLD.version THEN SET @curVersion = NEW.version; ELSE SELECT ROUND(UNIX_TIMESTAMP(CURTIME(2)) * 10) INTO @curVersion; END IF; IF NOT (OLD.`id` = NEW.`id` AND OLD.`type` <=> NEW.`type` AND OLD.`creator_user_id` <=> NEW.`creator_user_id` AND OLD.`item_id` <=> NEW.`item_id` AND OLD.`title` <=> NEW.`title` AND OLD.`admin_help_asked` <=> NEW.`admin_help_asked` AND OLD.`hidden` <=> NEW.`hidden` AND OLD.`last_activity_date` <=> NEW.`last_activity_date`) THEN   SET NEW.version = @curVersion;   UPDATE `history_threads` SET `next_version` = @curVersion WHERE `id` = OLD.`id` AND `next_version` IS NULL;   INSERT INTO `history_threads` (`id`,`version`,`type`,`creator_user_id`,`item_id`,`title`,`admin_help_asked`,`hidden`,`last_activity_date`)       VALUES (NEW.`id`,@curVersion,NEW.`type`,NEW.`creator_user_id`,NEW.`item_id`,NEW.`title`,NEW.`admin_help_asked`,NEW.`hidden`,NEW.`last_activity_date`) ; END IF; END
-- +migrate StatementEnd
DROP TRIGGER `before_delete_threads`;
-- +migrate StatementBegin
CREATE TRIGGER `before_delete_threads` BEFORE DELETE ON `threads` FOR EACH ROW BEGIN SELECT ROUND(UNIX_TIMESTAMP(CURTIME(2)) * 10) INTO @curVersion; UPDATE `history_threads` SET `next_version` = @curVersion WHERE `id` = OLD.`id` AND `next_version` IS NULL; INSERT INTO `history_threads` (`id`,`version`,`type`,`creator_user_id`,`item_id`,`title`,`admin_help_asked`,`hidden`,`last_activity_date`, `deleted`) VALUES (OLD.`id`,@curVersion,OLD.`type`,OLD.`creator_user_id`,OLD.`item_id`,OLD.`title`,OLD.`admin_help_asked`,OLD.`hidden`,OLD.`last_activity_date`, 1); END
-- +migrate StatementEnd
DROP TRIGGER `before_insert_users`;
-- +migrate StatementBegin
CREATE TRIGGER `before_insert_users` BEFORE INSERT ON `users` FOR EACH ROW BEGIN IF (NEW.id IS NULL OR NEW.id = 0) THEN SET NEW.id = FLOOR(RAND() * 1000000000) + FLOOR(RAND() * 1000000000) * 1000000000; END IF ; SELECT ROUND(UNIX_TIMESTAMP(CURTIME(2)) * 10) INTO @curVersion;SET NEW.version = @curVersion; END
-- +migrate StatementEnd
DROP TRIGGER `after_insert_users`;
-- +migrate StatementBegin
CREATE TRIGGER `after_insert_users` AFTER INSERT ON `users` FOR EACH ROW BEGIN INSERT INTO `history_users` (`id`,`version`,`login`,`open_id_identity`,`password_md5`,`salt`,`recover`,`registration_date`,`email`,`email_verified`,`first_name`,`last_name`,`country_code`,`time_zone`,`birth_date`,`graduation_year`,`grade`,`sex`,`student_id`,`address`,`zipcode`,`city`,`land_line_number`,`cell_phone_number`,`default_language`,`notify_news`,`notify`,`public_first_name`,`public_last_name`,`free_text`,`web_site`,`photo_autoload`,`lang_prog`,`last_login_date`,`last_activity_date`,`last_ip`,`basic_editor_mode`,`spaces_for_tab`,`member_state`,`godfather_user_id`,`step_level_in_site`,`is_admin`,`no_ranking`,`help_given`,`self_group_id`,`owned_group_id`,`access_group_id`,`notification_read_date`,`login_module_prefix`,`allow_subgroups`) VALUES (NEW.`id`,@curVersion,NEW.`login`,NEW.`open_id_identity`,NEW.`password_md5`,NEW.`salt`,NEW.`recover`,NEW.`registration_date`,NEW.`email`,NEW.`email_verified`,NEW.`first_name`,NEW.`last_name`,NEW.`country_code`,NEW.`time_zone`,NEW.`birth_date`,NEW.`graduation_year`,NEW.`grade`,NEW.`sex`,NEW.`student_id`,NEW.`address`,NEW.`zipcode`,NEW.`city`,NEW.`land_line_number`,NEW.`cell_phone_number`,NEW.`default_language`,NEW.`notify_news`,NEW.`notify`,NEW.`public_first_name`,NEW.`public_last_name`,NEW.`free_text`,NEW.`web_site`,NEW.`photo_autoload`,NEW.`lang_prog`,NEW.`last_login_date`,NEW.`last_activity_date`,NEW.`last_ip`,NEW.`basic_editor_mode`,NEW.`spaces_for_tab`,NEW.`member_state`,NEW.`godfather_user_id`,NEW.`step_level_in_site`,NEW.`is_admin`,NEW.`no_ranking`,NEW.`help_given`,NEW.`self_group_id`,NEW.`owned_group_id`,NEW.`access_group_id`,NEW.`notification_read_date`,NEW.`login_module_prefix`,NEW.`allow_subgroups`); END
-- +migrate StatementEnd
DROP TRIGGER `before_update_users`;
-- +migrate StatementBegin
CREATE TRIGGER `before_update_users` BEFORE UPDATE ON `users` FOR EACH ROW BEGIN IF NEW.version <> OLD.version THEN SET @curVersion = NEW.version; ELSE SELECT ROUND(UNIX_TIMESTAMP(CURTIME(2)) * 10) INTO @curVersion; END IF; IF NOT (OLD.`id` = NEW.`id` AND OLD.`login` <=> NEW.`login` AND OLD.`open_id_identity` <=> NEW.`open_id_identity` AND OLD.`password_md5` <=> NEW.`password_md5` AND OLD.`salt` <=> NEW.`salt` AND OLD.`recover` <=> NEW.`recover` AND OLD.`registration_date` <=> NEW.`registration_date` AND OLD.`email` <=> NEW.`email` AND OLD.`email_verified` <=> NEW.`email_verified` AND OLD.`first_name` <=> NEW.`first_name` AND OLD.`last_name` <=> NEW.`last_name` AND OLD.`country_code` <=> NEW.`country_code` AND OLD.`time_zone` <=> NEW.`time_zone` AND OLD.`birth_date` <=> NEW.`birth_date` AND OLD.`graduation_year` <=> NEW.`graduation_year` AND OLD.`grade` <=> NEW.`grade` AND OLD.`sex` <=> NEW.`sex` AND OLD.`student_id` <=> NEW.`student_id` AND OLD.`address` <=> NEW.`address` AND OLD.`zipcode` <=> NEW.`zipcode` AND OLD.`city` <=> NEW.`city` AND OLD.`land_line_number` <=> NEW.`land_line_number` AND OLD.`cell_phone_number` <=> NEW.`cell_phone_number` AND OLD.`default_language` <=> NEW.`default_language` AND OLD.`notify_news` <=> NEW.`notify_news` AND OLD.`notify` <=> NEW.`notify` AND OLD.`public_first_name` <=> NEW.`public_first_name` AND OLD.`public_last_name` <=> NEW.`public_last_name` AND OLD.`free_text` <=> NEW.`free_text` AND OLD.`web_site` <=> NEW.`web_site` AND OLD.`photo_autoload` <=> NEW.`photo_autoload` AND OLD.`lang_prog` <=> NEW.`lang_prog` AND OLD.`last_login_date` <=> NEW.`last_login_date` AND OLD.`last_activity_date` <=> NEW.`last_activity_date` AND OLD.`last_ip` <=> NEW.`last_ip` AND OLD.`basic_editor_mode` <=> NEW.`basic_editor_mode` AND OLD.`spaces_for_tab` <=> NEW.`spaces_for_tab` AND OLD.`member_state` <=> NEW.`member_state` AND OLD.`godfather_user_id` <=> NEW.`godfather_user_id` AND OLD.`step_level_in_site` <=> NEW.`step_level_in_site` AND OLD.`is_admin` <=> NEW.`is_admin` AND OLD.`no_ranking` <=> NEW.`no_ranking` AND OLD.`help_given` <=> NEW.`help_given` AND OLD.`self_group_id` <=> NEW.`self_group_id` AND OLD.`owned_group_id` <=> NEW.`owned_group_id` AND OLD.`access_group_id` <=> NEW.`access_group_id` AND OLD.`notification_read_date` <=> NEW.`notification_read_date` AND OLD.`login_module_prefix` <=> NEW.`login_module_prefix` AND OLD.`allow_subgroups` <=> NEW.`allow_subgroups`) THEN   SET NEW.version = @curVersion;   UPDATE `history_users` SET `next_version` = @curVersion WHERE `id` = OLD.`id` AND `next_version` IS NULL;   INSERT INTO `history_users` (`id`,`version`,`login`,`open_id_identity`,`password_md5`,`salt`,`recover`,`registration_date`,`email`,`email_verified`,`first_name`,`last_name`,`country_code`,`time_zone`,`birth_date`,`graduation_year`,`grade`,`sex`,`student_id`,`address`,`zipcode`,`city`,`land_line_number`,`cell_phone_number`,`default_language`,`notify_news`,`notify`,`public_first_name`,`public_last_name`,`free_text`,`web_site`,`photo_autoload`,`lang_prog`,`last_login_date`,`last_activity_date`,`last_ip`,`basic_editor_mode`,`spaces_for_tab`,`member_state`,`godfather_user_id`,`step_level_in_site`,`is_admin`,`no_ranking`,`help_given`,`self_group_id`,`owned_group_id`,`access_group_id`,`notification_read_date`,`login_module_prefix`,`allow_subgroups`)       VALUES (NEW.`id`,@curVersion,NEW.`login`,NEW.`open_id_identity`,NEW.`password_md5`,NEW.`salt`,NEW.`recover`,NEW.`registration_date`,NEW.`email`,NEW.`email_verified`,NEW.`first_name`,NEW.`last_name`,NEW.`country_code`,NEW.`time_zone`,NEW.`birth_date`,NEW.`graduation_year`,NEW.`grade`,NEW.`sex`,NEW.`student_id`,NEW.`address`,NEW.`zipcode`,NEW.`city`,NEW.`land_line_number`,NEW.`cell_phone_number`,NEW.`default_language`,NEW.`notify_news`,NEW.`notify`,NEW.`public_first_name`,NEW.`public_last_name`,NEW.`free_text`,NEW.`web_site`,NEW.`photo_autoload`,NEW.`lang_prog`,NEW.`last_login_date`,NEW.`last_activity_date`,NEW.`last_ip`,NEW.`basic_editor_mode`,NEW.`spaces_for_tab`,NEW.`member_state`,NEW.`godfather_user_id`,NEW.`step_level_in_site`,NEW.`is_admin`,NEW.`no_ranking`,NEW.`help_given`,NEW.`self_group_id`,NEW.`owned_group_id`,NEW.`access_group_id`,NEW.`notification_read_date`,NEW.`login_module_prefix`,NEW.`allow_subgroups`) ; END IF; END
-- +migrate StatementEnd
DROP TRIGGER `before_delete_users`;
-- +migrate StatementBegin
CREATE TRIGGER `before_delete_users` BEFORE DELETE ON `users` FOR EACH ROW BEGIN SELECT ROUND(UNIX_TIMESTAMP(CURTIME(2)) * 10) INTO @curVersion; UPDATE `history_users` SET `next_version` = @curVersion WHERE `id` = OLD.`id` AND `next_version` IS NULL; INSERT INTO `history_users` (`id`,`version`,`login`,`open_id_identity`,`password_md5`,`salt`,`recover`,`registration_date`,`email`,`email_verified`,`first_name`,`last_name`,`country_code`,`time_zone`,`birth_date`,`graduation_year`,`grade`,`sex`,`student_id`,`address`,`zipcode`,`city`,`land_line_number`,`cell_phone_number`,`default_language`,`notify_news`,`notify`,`public_first_name`,`public_last_name`,`free_text`,`web_site`,`photo_autoload`,`lang_prog`,`last_login_date`,`last_activity_date`,`last_ip`,`basic_editor_mode`,`spaces_for_tab`,`member_state`,`godfather_user_id`,`step_level_in_site`,`is_admin`,`no_ranking`,`help_given`,`self_group_id`,`owned_group_id`,`access_group_id`,`notification_read_date`,`login_module_prefix`,`allow_subgroups`, `deleted`) VALUES (OLD.`id`,@curVersion,OLD.`login`,OLD.`open_id_identity`,OLD.`password_md5`,OLD.`salt`,OLD.`recover`,OLD.`registration_date`,OLD.`email`,OLD.`email_verified`,OLD.`first_name`,OLD.`last_name`,OLD.`country_code`,OLD.`time_zone`,OLD.`birth_date`,OLD.`graduation_year`,OLD.`grade`,OLD.`sex`,OLD.`student_id`,OLD.`address`,OLD.`zipcode`,OLD.`city`,OLD.`land_line_number`,OLD.`cell_phone_number`,OLD.`default_language`,OLD.`notify_news`,OLD.`notify`,OLD.`public_first_name`,OLD.`public_last_name`,OLD.`free_text`,OLD.`web_site`,OLD.`photo_autoload`,OLD.`lang_prog`,OLD.`last_login_date`,OLD.`last_activity_date`,OLD.`last_ip`,OLD.`basic_editor_mode`,OLD.`spaces_for_tab`,OLD.`member_state`,OLD.`godfather_user_id`,OLD.`step_level_in_site`,OLD.`is_admin`,OLD.`no_ranking`,OLD.`help_given`,OLD.`self_group_id`,OLD.`owned_group_id`,OLD.`access_group_id`,OLD.`notification_read_date`,OLD.`login_module_prefix`,OLD.`allow_subgroups`, 1); END
-- +migrate StatementEnd
DROP TRIGGER `before_insert_users_items`;
-- +migrate StatementBegin
CREATE TRIGGER `before_insert_users_items` BEFORE INSERT ON `users_items` FOR EACH ROW BEGIN IF (NEW.id IS NULL OR NEW.id = 0) THEN SET NEW.id = FLOOR(RAND() * 1000000000) + FLOOR(RAND() * 1000000000) * 1000000000; END IF ; SELECT ROUND(UNIX_TIMESTAMP(CURTIME(2)) * 10) INTO @curVersion;SET NEW.version = @curVersion; END
-- +migrate StatementEnd
DROP TRIGGER `after_insert_users_items`;
-- +migrate StatementBegin
CREATE TRIGGER `after_insert_users_items` AFTER INSERT ON `users_items` FOR EACH ROW BEGIN INSERT INTO `history_users_items` (`id`,`version`,`user_id`,`item_id`,`active_attempt_id`,`score`,`score_computed`,`score_reeval`,`score_diff_manual`,`score_diff_comment`,`submissions_attempts`,`tasks_tried`,`children_validated`,`validated`,`finished`,`key_obtained`,`tasks_with_help`,`hints_requested`,`hints_cached`,`corrections_read`,`precision`,`autonomy`,`start_date`,`validation_date`,`best_answer_date`,`last_answer_date`,`thread_start_date`,`last_hint_date`,`finish_date`,`last_activity_date`,`contest_start_date`,`ranked`,`all_lang_prog`,`state`,`answer`) VALUES (NEW.`id`,@curVersion,NEW.`user_id`,NEW.`item_id`,NEW.`active_attempt_id`,NEW.`score`,NEW.`score_computed`,NEW.`score_reeval`,NEW.`score_diff_manual`,NEW.`score_diff_comment`,NEW.`submissions_attempts`,NEW.`tasks_tried`,NEW.`children_validated`,NEW.`validated`,NEW.`finished`,NEW.`key_obtained`,NEW.`tasks_with_help`,NEW.`hints_requested`,NEW.`hints_cached`,NEW.`corrections_read`,NEW.`precision`,NEW.`autonomy`,NEW.`start_date`,NEW.`validation_date`,NEW.`best_answer_date`,NEW.`last_answer_date`,NEW.`thread_start_date`,NEW.`last_hint_date`,NEW.`finish_date`,NEW.`last_activity_date`,NEW.`contest_start_date`,NEW.`ranked`,NEW.`all_lang_prog`,NEW.`state`,NEW.`answer`); END
-- +migrate StatementEnd
DROP TRIGGER `before_update_users_items`;
-- +migrate StatementBegin
CREATE TRIGGER `before_update_users_items` BEFORE UPDATE ON `users_items` FOR EACH ROW BEGIN IF NEW.version <> OLD.version THEN SET @curVersion = NEW.version; ELSE SELECT ROUND(UNIX_TIMESTAMP(CURTIME(2)) * 10) INTO @curVersion; END IF; IF NOT (OLD.`id` = NEW.`id` AND OLD.`user_id` <=> NEW.`user_id` AND OLD.`item_id` <=> NEW.`item_id` AND OLD.`active_attempt_id` <=> NEW.`active_attempt_id` AND OLD.`score` <=> NEW.`score` AND OLD.`score_computed` <=> NEW.`score_computed` AND OLD.`score_reeval` <=> NEW.`score_reeval` AND OLD.`score_diff_manual` <=> NEW.`score_diff_manual` AND OLD.`score_diff_comment` <=> NEW.`score_diff_comment` AND OLD.`tasks_tried` <=> NEW.`tasks_tried` AND OLD.`children_validated` <=> NEW.`children_validated` AND OLD.`validated` <=> NEW.`validated` AND OLD.`finished` <=> NEW.`finished` AND OLD.`key_obtained` <=> NEW.`key_obtained` AND OLD.`tasks_with_help` <=> NEW.`tasks_with_help` AND OLD.`hints_requested` <=> NEW.`hints_requested` AND OLD.`hints_cached` <=> NEW.`hints_cached` AND OLD.`corrections_read` <=> NEW.`corrections_read` AND OLD.`precision` <=> NEW.`precision` AND OLD.`autonomy` <=> NEW.`autonomy` AND OLD.`start_date` <=> NEW.`start_date` AND OLD.`validation_date` <=> NEW.`validation_date` AND OLD.`best_answer_date` <=> NEW.`best_answer_date` AND OLD.`last_answer_date` <=> NEW.`last_answer_date` AND OLD.`thread_start_date` <=> NEW.`thread_start_date` AND OLD.`last_hint_date` <=> NEW.`last_hint_date` AND OLD.`finish_date` <=> NEW.`finish_date` AND OLD.`contest_start_date` <=> NEW.`contest_start_date` AND OLD.`ranked` <=> NEW.`ranked` AND OLD.`all_lang_prog` <=> NEW.`all_lang_prog` AND OLD.`state` <=> NEW.`state` AND OLD.`answer` <=> NEW.`answer`) THEN   SET NEW.version = @curVersion;   UPDATE `history_users_items` SET `next_version` = @curVersion WHERE `id` = OLD.`id` AND `next_version` IS NULL;   INSERT INTO `history_users_items` (`id`,`version`,`user_id`,`item_id`,`active_attempt_id`,`score`,`score_computed`,`score_reeval`,`score_diff_manual`,`score_diff_comment`,`submissions_attempts`,`tasks_tried`,`children_validated`,`validated`,`finished`,`key_obtained`,`tasks_with_help`,`hints_requested`,`hints_cached`,`corrections_read`,`precision`,`autonomy`,`start_date`,`validation_date`,`best_answer_date`,`last_answer_date`,`thread_start_date`,`last_hint_date`,`finish_date`,`last_activity_date`,`contest_start_date`,`ranked`,`all_lang_prog`,`state`,`answer`)       VALUES (NEW.`id`,@curVersion,NEW.`user_id`,NEW.`item_id`,NEW.`active_attempt_id`,NEW.`score`,NEW.`score_computed`,NEW.`score_reeval`,NEW.`score_diff_manual`,NEW.`score_diff_comment`,NEW.`submissions_attempts`,NEW.`tasks_tried`,NEW.`children_validated`,NEW.`validated`,NEW.`finished`,NEW.`key_obtained`,NEW.`tasks_with_help`,NEW.`hints_requested`,NEW.`hints_cached`,NEW.`corrections_read`,NEW.`precision`,NEW.`autonomy`,NEW.`start_date`,NEW.`validation_date`,NEW.`best_answer_date`,NEW.`last_answer_date`,NEW.`thread_start_date`,NEW.`last_hint_date`,NEW.`finish_date`,NEW.`last_activity_date`,NEW.`contest_start_date`,NEW.`ranked`,NEW.`all_lang_prog`,NEW.`state`,NEW.`answer`) ; END IF; END
-- +migrate StatementEnd
DROP TRIGGER `before_delete_users_items`;
-- +migrate StatementBegin
CREATE TRIGGER `before_delete_users_items` BEFORE DELETE ON `users_items` FOR EACH ROW BEGIN SELECT ROUND(UNIX_TIMESTAMP(CURTIME(2)) * 10) INTO @curVersion; UPDATE `history_users_items` SET `next_version` = @curVersion WHERE `id` = OLD.`id` AND `next_version` IS NULL; INSERT INTO `history_users_items` (`id`,`version`,`user_id`,`item_id`,`active_attempt_id`,`score`,`score_computed`,`score_reeval`,`score_diff_manual`,`score_diff_comment`,`submissions_attempts`,`tasks_tried`,`children_validated`,`validated`,`finished`,`key_obtained`,`tasks_with_help`,`hints_requested`,`hints_cached`,`corrections_read`,`precision`,`autonomy`,`start_date`,`validation_date`,`best_answer_date`,`last_answer_date`,`thread_start_date`,`last_hint_date`,`finish_date`,`last_activity_date`,`contest_start_date`,`ranked`,`all_lang_prog`,`state`,`answer`, `deleted`) VALUES (OLD.`id`,@curVersion,OLD.`user_id`,OLD.`item_id`,OLD.`active_attempt_id`,OLD.`score`,OLD.`score_computed`,OLD.`score_reeval`,OLD.`score_diff_manual`,OLD.`score_diff_comment`,OLD.`submissions_attempts`,OLD.`tasks_tried`,OLD.`children_validated`,OLD.`validated`,OLD.`finished`,OLD.`key_obtained`,OLD.`tasks_with_help`,OLD.`hints_requested`,OLD.`hints_cached`,OLD.`corrections_read`,OLD.`precision`,OLD.`autonomy`,OLD.`start_date`,OLD.`validation_date`,OLD.`best_answer_date`,OLD.`last_answer_date`,OLD.`thread_start_date`,OLD.`last_hint_date`,OLD.`finish_date`,OLD.`last_activity_date`,OLD.`contest_start_date`,OLD.`ranked`,OLD.`all_lang_prog`,OLD.`state`,OLD.`answer`, 1); END
-- +migrate StatementEnd
DROP TRIGGER `before_insert_users_threads`;
-- +migrate StatementBegin
CREATE TRIGGER `before_insert_users_threads` BEFORE INSERT ON `users_threads` FOR EACH ROW BEGIN IF (NEW.id IS NULL OR NEW.id = 0) THEN SET NEW.id = FLOOR(RAND() * 1000000000) + FLOOR(RAND() * 1000000000) * 1000000000; END IF ; SELECT ROUND(UNIX_TIMESTAMP(CURTIME(2)) * 10) INTO @curVersion;SET NEW.version = @curVersion; END
-- +migrate StatementEnd
DROP TRIGGER `after_insert_users_threads`;
-- +migrate StatementBegin
CREATE TRIGGER `after_insert_users_threads` AFTER INSERT ON `users_threads` FOR EACH ROW BEGIN INSERT INTO `history_users_threads` (`id`,`version`,`user_id`,`thread_id`,`last_read_date`,`last_write_date`,`starred`) VALUES (NEW.`id`,@curVersion,NEW.`user_id`,NEW.`thread_id`,NEW.`last_read_date`,NEW.`last_write_date`,NEW.`starred`); END
-- +migrate StatementEnd
DROP TRIGGER `before_update_users_threads`;
-- +migrate StatementBegin
CREATE TRIGGER `before_update_users_threads` BEFORE UPDATE ON `users_threads` FOR EACH ROW BEGIN IF NEW.version <> OLD.version THEN SET @curVersion = NEW.version; ELSE SELECT ROUND(UNIX_TIMESTAMP(CURTIME(2)) * 10) INTO @curVersion; END IF; IF NOT (OLD.`id` = NEW.`id` AND OLD.`user_id` <=> NEW.`user_id` AND OLD.`thread_id` <=> NEW.`thread_id` AND OLD.`last_read_date` <=> NEW.`last_read_date` AND OLD.`last_write_date` <=> NEW.`last_write_date` AND OLD.`starred` <=> NEW.`starred`) THEN   SET NEW.version = @curVersion;   UPDATE `history_users_threads` SET `next_version` = @curVersion WHERE `id` = OLD.`id` AND `next_version` IS NULL;   INSERT INTO `history_users_threads` (`id`,`version`,`user_id`,`thread_id`,`last_read_date`,`last_write_date`,`starred`)       VALUES (NEW.`id`,@curVersion,NEW.`user_id`,NEW.`thread_id`,NEW.`last_read_date`,NEW.`last_write_date`,NEW.`starred`) ; END IF; END
-- +migrate StatementEnd
DROP TRIGGER `before_delete_users_threads`;
-- +migrate StatementBegin
CREATE TRIGGER `before_delete_users_threads` BEFORE DELETE ON `users_threads` FOR EACH ROW BEGIN SELECT ROUND(UNIX_TIMESTAMP(CURTIME(2)) * 10) INTO @curVersion; UPDATE `history_users_threads` SET `next_version` = @curVersion WHERE `id` = OLD.`id` AND `next_version` IS NULL; INSERT INTO `history_users_threads` (`id`,`version`,`user_id`,`thread_id`,`last_read_date`,`last_write_date`,`starred`, `deleted`) VALUES (OLD.`id`,@curVersion,OLD.`user_id`,OLD.`thread_id`,OLD.`last_read_date`,OLD.`last_write_date`,OLD.`starred`, 1); END
-- +migrate StatementEnd

ALTER ALGORITHM=UNDEFINED
  SQL SECURITY DEFINER
  VIEW `task_children_data_view` AS
SELECT
    `parent_users_items`.`id` AS `user_item_id`,
    SUM(IF(`task_children`.`id` IS NOT NULL AND `task_children`.`validated`, 1, 0)) AS `children_validated`,
    SUM(IF(`task_children`.`id` IS NOT NULL AND `task_children`.`validated`, 0, 1)) AS `children_non_validated`,
    SUM(IF(`items_items`.`category` = 'Validation' AND
           (ISNULL(`task_children`.`id`) OR `task_children`.`validated` != 1), 1, 0)) AS `children_category`,
    MAX(`task_children`.`validation_date`) AS `max_validation_date`,
    MAX(IF(`items_items`.`category` = 'Validation', `task_children`.`validation_date`, NULL)) AS `max_validation_date_categories`
FROM `users_items` AS `parent_users_items`
         JOIN `items_items` ON(
        `parent_users_items`.`item_id` = `items_items`.`parent_item_id`
    )
         LEFT JOIN `users_items` AS `task_children` ON(
            `items_items`.`child_item_id` = `task_children`.`item_id` AND
            `task_children`.`user_id` = `parent_users_items`.`user_id`
    )
         JOIN `items` ON(
        `items`.`ID` = `items_items`.`child_item_id`
    )
WHERE `items`.`type` <> 'Course' AND `items`.`no_score` = 0
GROUP BY `user_item_id`;

-- +migrate Down
ALTER TABLE `badges`
	RENAME COLUMN `id` TO `ID`,
	RENAME COLUMN `user_id` TO `idUser`;
ALTER TABLE `filters`
	RENAME COLUMN `id` TO `ID`,
	RENAME COLUMN `user_id` TO `idUser`,
	RENAME COLUMN `name` TO `sName`,
	RENAME COLUMN `selected` TO `bSelected`,
	RENAME COLUMN `starred` TO `bStarred`,
	RENAME COLUMN `start_date` TO `sStartDate`,
	RENAME COLUMN `end_date` TO `sEndDate`,
	RENAME COLUMN `archived` TO `bArchived`,
	RENAME COLUMN `participated` TO `bParticipated`,
	RENAME COLUMN `unread` TO `bUnread`,
	RENAME COLUMN `item_id` TO `idItem`,
	RENAME COLUMN `group_id` TO `idGroup`,
	RENAME COLUMN `older_than` TO `olderThan`,
	RENAME COLUMN `newer_than` TO `newerThan`,
	RENAME COLUMN `users_search` TO `sUsersSearch`,
	RENAME COLUMN `body_search` TO `sBodySearch`,
	RENAME COLUMN `important` TO `bImportant`,
	RENAME COLUMN `version` TO `iVersion`,
	RENAME INDEX `version` TO `iVersion`;
ALTER TABLE `groups`
	RENAME COLUMN `id` TO `ID`,
	RENAME COLUMN `name` TO `sName`,
	RENAME COLUMN `text_id` TO `sTextId`,
	RENAME COLUMN `grade` TO `iGrade`,
	RENAME COLUMN `grade_details` TO `sGradeDetails`,
	RENAME COLUMN `description` TO `sDescription`,
	RENAME COLUMN `date_created` TO `sDateCreated`,
	RENAME COLUMN `opened` TO `bOpened`,
	RENAME COLUMN `free_access` TO `bFreeAccess`,
	RENAME COLUMN `team_item_id` TO `idTeamItem`,
	RENAME COLUMN `team_participating` TO `iTeamParticipating`,
	RENAME COLUMN `code` TO `sCode`,
	RENAME COLUMN `code_timer` TO `sCodeTimer`,
	RENAME COLUMN `code_end` TO `sCodeEnd`,
	RENAME COLUMN `redirect_path` TO `sRedirectPath`,
	RENAME COLUMN `open_contest` TO `bOpenContest`,
	RENAME COLUMN `type` TO `sType`,
	RENAME COLUMN `send_emails` TO `bSendEmails`,
	RENAME COLUMN `version` TO `iVersion`,
	RENAME COLUMN `lock_user_deletion_date` TO `lockUserDeletionDate`,
	RENAME INDEX `password` TO `sPassword`,
	RENAME INDEX `version` TO `iVersion`,
	RENAME INDEX `type` TO `sType`,
	RENAME INDEX `name` TO `sName`,
	RENAME INDEX `type_name` TO `TypeName`;
ALTER TABLE `groups_ancestors`
	RENAME COLUMN `id` TO `ID`,
	RENAME COLUMN `ancestor_group_id` TO `idGroupAncestor`,
	RENAME COLUMN `child_group_id` TO `idGroupChild`,
	RENAME COLUMN `is_self` TO `bIsSelf`,
	RENAME COLUMN `version` TO `iVersion`,
	RENAME INDEX `ancestor_group_id` TO `idGroupAncestor`;
ALTER TABLE `groups_attempts`
	RENAME COLUMN `id` TO `ID`,
	RENAME COLUMN `group_id` TO `idGroup`,
	RENAME COLUMN `item_id` TO `idItem`,
	RENAME COLUMN `creator_user_id` TO `idUserCreator`,
	DROP CHECK `cs_attempts_order`,
	RENAME COLUMN `order` TO `iOrder`,
	ADD CONSTRAINT `cs_attempts_order` CHECK (`iOrder` > 0),
	RENAME COLUMN `score` TO `iScore`,
	RENAME COLUMN `score_computed` TO `iScoreComputed`,
	RENAME COLUMN `score_reeval` TO `iScoreReeval`,
	RENAME COLUMN `score_diff_manual` TO `iScoreDiffManual`,
	RENAME COLUMN `score_diff_comment` TO `sScoreDiffComment`,
	RENAME COLUMN `submissions_attempts` TO `nbSubmissionsAttempts`,
	RENAME COLUMN `tasks_tried` TO `nbTasksTried`,
	RENAME COLUMN `tasks_solved` TO `nbTasksSolved`,
	RENAME COLUMN `children_validated` TO `nbChildrenValidated`,
	RENAME COLUMN `validated` TO `bValidated`,
	RENAME COLUMN `finished` TO `bFinished`,
	RENAME COLUMN `key_obtained` TO `bKeyObtained`,
	RENAME COLUMN `tasks_with_help` TO `nbTasksWithHelp`,
	RENAME COLUMN `hints_requested` TO `sHintsRequested`,
	RENAME COLUMN `hints_cached` TO `nbHintsCached`,
	RENAME COLUMN `corrections_read` TO `nbCorrectionsRead`,
	RENAME COLUMN `precision` TO `iPrecision`,
	RENAME COLUMN `autonomy` TO `iAutonomy`,
	RENAME COLUMN `start_date` TO `sStartDate`,
	RENAME COLUMN `validation_date` TO `sValidationDate`,
	RENAME COLUMN `finish_date` TO `sFinishDate`,
	RENAME COLUMN `last_activity_date` TO `sLastActivityDate`,
	RENAME COLUMN `thread_start_date` TO `sThreadStartDate`,
	RENAME COLUMN `best_answer_date` TO `sBestAnswerDate`,
	RENAME COLUMN `last_answer_date` TO `sLastAnswerDate`,
	RENAME COLUMN `last_hint_date` TO `sLastHintDate`,
	RENAME COLUMN `contest_start_date` TO `sContestStartDate`,
	RENAME COLUMN `ranked` TO `bRanked`,
	RENAME COLUMN `all_lang_prog` TO `sAllLangProg`,
	RENAME COLUMN `version` TO `iVersion`,
	RENAME COLUMN `ancestors_computation_state` TO `sAncestorsComputationState`,
	RENAME COLUMN `minus_score` TO `iMinusScore`,
	RENAME INDEX `version` TO `iVersion`,
	RENAME INDEX `ancestors_computation_state` TO `sAncestorsComputationState`,
	RENAME INDEX `item_id` TO `idItem`,
	RENAME INDEX `group_item` TO `GroupItem`,
	RENAME INDEX `group_item_minus_score_best_answer_date_id` TO `GroupItemMinusScoreBestAnswerDateID`;
ALTER TABLE `groups_groups`
	RENAME COLUMN `id` TO `ID`,
	RENAME COLUMN `parent_group_id` TO `idGroupParent`,
	RENAME COLUMN `child_group_id` TO `idGroupChild`,
	RENAME COLUMN `child_order` TO `iChildOrder`,
	RENAME COLUMN `type` TO `sType`,
	RENAME COLUMN `role` TO `sRole`,
	RENAME COLUMN `inviting_user_id` TO `idUserInviting`,
	RENAME COLUMN `status_date` TO `sStatusDate`,
	RENAME COLUMN `version` TO `iVersion`,
	RENAME INDEX `version` TO `iVersion`,
	RENAME INDEX `child_group_id` TO `idGroupChild`,
	RENAME INDEX `parent_group_id` TO `idGroupParent`,
	RENAME INDEX `parent_order` TO `ParentOrder`;
ALTER TABLE `groups_items`
	RENAME COLUMN `id` TO `ID`,
	RENAME COLUMN `group_id` TO `idGroup`,
	RENAME COLUMN `item_id` TO `idItem`,
	RENAME COLUMN `creator_user_id` TO `idUserCreated`,
	RENAME COLUMN `partial_access_date` TO `sPartialAccessDate`,
	RENAME COLUMN `access_reason` TO `sAccessReason`,
	RENAME COLUMN `full_access_date` TO `sFullAccessDate`,
	RENAME COLUMN `access_solutions_date` TO `sAccessSolutionsDate`,
	RENAME COLUMN `owner_access` TO `bOwnerAccess`,
	RENAME COLUMN `manager_access` TO `bManagerAccess`,
	RENAME COLUMN `cached_full_access_date` TO `sCachedFullAccessDate`,
	RENAME COLUMN `cached_partial_access_date` TO `sCachedPartialAccessDate`,
	RENAME COLUMN `cached_access_solutions_date` TO `sCachedAccessSolutionsDate`,
	RENAME COLUMN `cached_grayed_access_date` TO `sCachedGrayedAccessDate`,
	RENAME COLUMN `cached_access_reason` TO `sCachedAccessReason`,
	RENAME COLUMN `cached_full_access` TO `bCachedFullAccess`,
	RENAME COLUMN `cached_partial_access` TO `bCachedPartialAccess`,
	RENAME COLUMN `cached_access_solutions` TO `bCachedAccessSolutions`,
	RENAME COLUMN `cached_grayed_access` TO `bCachedGrayedAccess`,
	RENAME COLUMN `cached_manager_access` TO `bCachedManagerAccess`,
	RENAME COLUMN `propagate_access` TO `sPropagateAccess`,
	RENAME COLUMN `additional_time` TO `sAdditionalTime`,
	RENAME COLUMN `version` TO `iVersion`,
	RENAME INDEX `item_id` TO `idItem`,
	RENAME INDEX `version` TO `iVersion`,
	RENAME INDEX `group_id` TO `idGroup`,
	RENAME INDEX `itemtem_id` TO `idItemtem`,
	RENAME INDEX `full_access` TO `fullAccess`,
	RENAME INDEX `access_solutions` TO `accessSolutions`,
	RENAME INDEX `propagate_access` TO `sPropagateAccess`,
	RENAME INDEX `partial_access` TO `partialAccess`;
ALTER TABLE `groups_items_propagate`
	RENAME COLUMN `id` TO `ID`,
	RENAME COLUMN `propagate_access` TO `sPropagateAccess`,
	RENAME INDEX `propagate_access` TO `sPropagateAccess`;
ALTER TABLE `groups_login_prefixes`
	RENAME COLUMN `id` TO `ID`,
	RENAME COLUMN `group_id` TO `idGroup`,
	RENAME COLUMN `version` TO `iVersion`,
	RENAME INDEX `group_id` TO `idGroup`;
ALTER TABLE `groups_propagate`
	RENAME COLUMN `id` TO `ID`,
	RENAME COLUMN `ancestors_computation_state` TO `sAncestorsComputationState`,
	RENAME INDEX `ancestors_computation_state` TO `sAncestorsComputationState`;
ALTER TABLE `history_filters`
	RENAME COLUMN `history_id` TO `historyID`,
	RENAME COLUMN `id` TO `ID`,
	RENAME COLUMN `user_id` TO `idUser`,
	RENAME COLUMN `name` TO `sName`,
	RENAME COLUMN `selected` TO `bSelected`,
	RENAME COLUMN `starred` TO `bStarred`,
	RENAME COLUMN `start_date` TO `sStartDate`,
	RENAME COLUMN `end_date` TO `sEndDate`,
	RENAME COLUMN `archived` TO `bArchived`,
	RENAME COLUMN `participated` TO `bParticipated`,
	RENAME COLUMN `unread` TO `bUnread`,
	RENAME COLUMN `item_id` TO `idItem`,
	RENAME COLUMN `group_id` TO `idGroup`,
	RENAME COLUMN `older_than` TO `olderThan`,
	RENAME COLUMN `newer_than` TO `newerThan`,
	RENAME COLUMN `users_search` TO `sUsersSearch`,
	RENAME COLUMN `body_search` TO `sBodySearch`,
	RENAME COLUMN `important` TO `bImportant`,
	RENAME COLUMN `version` TO `iVersion`,
	RENAME COLUMN `next_version` TO `iNextVersion`,
	RENAME COLUMN `deleted` TO `bDeleted`,
	RENAME INDEX `version` TO `iVersion`,
	RENAME INDEX `next_version` TO `iNextVersion`,
	RENAME INDEX `deleted` TO `bDeleted`,
	RENAME INDEX `id` TO `ID`;
ALTER TABLE `history_groups`
	RENAME COLUMN `history_id` TO `historyID`,
	RENAME COLUMN `id` TO `ID`,
	RENAME COLUMN `name` TO `sName`,
	RENAME COLUMN `grade` TO `iGrade`,
	RENAME COLUMN `grade_details` TO `sGradeDetails`,
	RENAME COLUMN `description` TO `sDescription`,
	RENAME COLUMN `date_created` TO `sDateCreated`,
	RENAME COLUMN `opened` TO `bOpened`,
	RENAME COLUMN `free_access` TO `bFreeAccess`,
	RENAME COLUMN `team_item_id` TO `idTeamItem`,
	RENAME COLUMN `team_participating` TO `iTeamParticipating`,
	RENAME COLUMN `code` TO `sCode`,
	RENAME COLUMN `code_timer` TO `sCodeTimer`,
	RENAME COLUMN `code_end` TO `sCodeEnd`,
	RENAME COLUMN `redirect_path` TO `sRedirectPath`,
	RENAME COLUMN `open_contest` TO `bOpenContest`,
	RENAME COLUMN `type` TO `sType`,
	RENAME COLUMN `send_emails` TO `bSendEmails`,
	RENAME COLUMN `ancestors_computed` TO `bAncestorsComputed`,
	RENAME COLUMN `ancestors_computation_state` TO `sAncestorsComputationState`,
	RENAME COLUMN `version` TO `iVersion`,
	RENAME COLUMN `next_version` TO `iNextVersion`,
	RENAME COLUMN `deleted` TO `bDeleted`,
	RENAME COLUMN `lock_user_deletion_date` TO `lockUserDeletionDate`,
	RENAME INDEX `version` TO `iVersion`,
	RENAME INDEX `id` TO `ID`,
	RENAME INDEX `next_version` TO `iNextVersion`,
	RENAME INDEX `deleted` TO `bDeleted`;
ALTER TABLE `history_groups_ancestors`
	RENAME COLUMN `history_id` TO `historyID`,
	RENAME COLUMN `id` TO `ID`,
	RENAME COLUMN `ancestor_group_id` TO `idGroupAncestor`,
	RENAME COLUMN `child_group_id` TO `idGroupChild`,
	RENAME COLUMN `is_self` TO `bIsSelf`,
	RENAME COLUMN `version` TO `iVersion`,
	RENAME COLUMN `next_version` TO `iNextVersion`,
	RENAME COLUMN `deleted` TO `bDeleted`,
	RENAME INDEX `version` TO `iVersion`,
	RENAME INDEX `next_version` TO `iNextVersion`,
	RENAME INDEX `deleted` TO `bDeleted`,
	RENAME INDEX `ancestor_group_id` TO `idGroupAncestor`,
	RENAME INDEX `id` TO `ID`;
ALTER TABLE `history_groups_attempts`
	RENAME COLUMN `history_id` TO `historyID`,
	RENAME COLUMN `id` TO `ID`,
	RENAME COLUMN `group_id` TO `idGroup`,
	RENAME COLUMN `item_id` TO `idItem`,
	RENAME COLUMN `creator_user_id` TO `idUserCreator`,
	RENAME COLUMN `order` TO `iOrder`,
	RENAME COLUMN `score` TO `iScore`,
	RENAME COLUMN `score_computed` TO `iScoreComputed`,
	RENAME COLUMN `score_reeval` TO `iScoreReeval`,
	RENAME COLUMN `score_diff_manual` TO `iScoreDiffManual`,
	RENAME COLUMN `score_diff_comment` TO `sScoreDiffComment`,
	RENAME COLUMN `submissions_attempts` TO `nbSubmissionsAttempts`,
	RENAME COLUMN `tasks_tried` TO `nbTasksTried`,
	RENAME COLUMN `tasks_solved` TO `nbTasksSolved`,
	RENAME COLUMN `children_validated` TO `nbChildrenValidated`,
	RENAME COLUMN `validated` TO `bValidated`,
	RENAME COLUMN `finished` TO `bFinished`,
	RENAME COLUMN `key_obtained` TO `bKeyObtained`,
	RENAME COLUMN `tasks_with_help` TO `nbTasksWithHelp`,
	RENAME COLUMN `hints_requested` TO `sHintsRequested`,
	RENAME COLUMN `hints_cached` TO `nbHintsCached`,
	RENAME COLUMN `corrections_read` TO `nbCorrectionsRead`,
	RENAME COLUMN `precision` TO `iPrecision`,
	RENAME COLUMN `autonomy` TO `iAutonomy`,
	RENAME COLUMN `start_date` TO `sStartDate`,
	RENAME COLUMN `validation_date` TO `sValidationDate`,
	RENAME COLUMN `finish_date` TO `sFinishDate`,
	RENAME COLUMN `last_activity_date` TO `sLastActivityDate`,
	RENAME COLUMN `thread_start_date` TO `sThreadStartDate`,
	RENAME COLUMN `best_answer_date` TO `sBestAnswerDate`,
	RENAME COLUMN `last_answer_date` TO `sLastAnswerDate`,
	RENAME COLUMN `last_hint_date` TO `sLastHintDate`,
	RENAME COLUMN `contest_start_date` TO `sContestStartDate`,
	RENAME COLUMN `ranked` TO `bRanked`,
	RENAME COLUMN `all_lang_prog` TO `sAllLangProg`,
	RENAME COLUMN `version` TO `iVersion`,
	RENAME COLUMN `next_version` TO `iNextVersion`,
	RENAME COLUMN `deleted` TO `bDeleted`,
	RENAME INDEX `version` TO `iVersion`,
	RENAME INDEX `item_id` TO `idItem`,
	RENAME INDEX `group_item` TO `GroupItem`,
	RENAME INDEX `group_id` TO `idGroup`,
	RENAME INDEX `id` TO `ID`;
ALTER TABLE `history_groups_groups`
	RENAME COLUMN `history_id` TO `historyID`,
	RENAME COLUMN `id` TO `ID`,
	RENAME COLUMN `parent_group_id` TO `idGroupParent`,
	RENAME COLUMN `child_group_id` TO `idGroupChild`,
	RENAME COLUMN `child_order` TO `iChildOrder`,
	RENAME COLUMN `type` TO `sType`,
	RENAME COLUMN `role` TO `sRole`,
	RENAME COLUMN `inviting_user_id` TO `idUserInviting`,
	RENAME COLUMN `status_date` TO `sStatusDate`,
	RENAME COLUMN `version` TO `iVersion`,
	RENAME COLUMN `next_version` TO `iNextVersion`,
	RENAME COLUMN `deleted` TO `bDeleted`,
	RENAME INDEX `version` TO `iVersion`,
	RENAME INDEX `id` TO `ID`,
	RENAME INDEX `next_version` TO `iNextVersion`,
	RENAME INDEX `deleted` TO `bDeleted`,
	RENAME INDEX `parent_group_id` TO `idGroupParent`,
	RENAME INDEX `child_group_id` TO `idGroupChild`;
ALTER TABLE `history_groups_items`
	RENAME COLUMN `history_id` TO `historyID`,
	RENAME COLUMN `id` TO `ID`,
	RENAME COLUMN `group_id` TO `idGroup`,
	RENAME COLUMN `item_id` TO `idItem`,
	RENAME COLUMN `creator_user_id` TO `idUserCreated`,
	RENAME COLUMN `partial_access_date` TO `sPartialAccessDate`,
	RENAME COLUMN `access_reason` TO `sAccessReason`,
	RENAME COLUMN `full_access_date` TO `sFullAccessDate`,
	RENAME COLUMN `access_solutions_date` TO `sAccessSolutionsDate`,
	RENAME COLUMN `owner_access` TO `bOwnerAccess`,
	RENAME COLUMN `manager_access` TO `bManagerAccess`,
	RENAME COLUMN `cached_full_access_date` TO `sCachedFullAccessDate`,
	RENAME COLUMN `cached_partial_access_date` TO `sCachedPartialAccessDate`,
	RENAME COLUMN `cached_access_solutions_date` TO `sCachedAccessSolutionsDate`,
	RENAME COLUMN `cached_grayed_access_date` TO `sCachedGrayedAccessDate`,
	RENAME COLUMN `cached_access_reason` TO `sCachedAccessReason`,
	RENAME COLUMN `cached_full_access` TO `bCachedFullAccess`,
	RENAME COLUMN `cached_partial_access` TO `bCachedPartialAccess`,
	RENAME COLUMN `cached_access_solutions` TO `bCachedAccessSolutions`,
	RENAME COLUMN `cached_grayed_access` TO `bCachedGrayedAccess`,
	RENAME COLUMN `cached_manager_access` TO `bCachedManagerAccess`,
	RENAME COLUMN `propagate_access` TO `sPropagateAccess`,
	RENAME COLUMN `version` TO `iVersion`,
	RENAME COLUMN `next_version` TO `iNextVersion`,
	RENAME COLUMN `deleted` TO `bDeleted`,
	RENAME INDEX `version` TO `iVersion`,
	RENAME INDEX `id` TO `ID`,
	RENAME INDEX `item_group` TO `itemGroup`,
	RENAME INDEX `next_version` TO `iNextVersion`,
	RENAME INDEX `deleted` TO `bDeleted`,
	RENAME INDEX `item_id` TO `idItem`,
	RENAME INDEX `group_id` TO `idGroup`;
ALTER TABLE `history_groups_login_prefixes`
	RENAME COLUMN `history_id` TO `historyID`,
	RENAME COLUMN `id` TO `ID`,
	RENAME COLUMN `group_id` TO `idGroup`,
	RENAME COLUMN `version` TO `iVersion`,
	RENAME COLUMN `next_version` TO `iNextVersion`,
	RENAME COLUMN `deleted` TO `bDeleted`,
	RENAME INDEX `group_id` TO `idGroup`;
ALTER TABLE `history_items`
	RENAME COLUMN `history_id` TO `historyID`,
	RENAME COLUMN `id` TO `ID`,
	RENAME COLUMN `url` TO `sUrl`,
	RENAME COLUMN `platform_id` TO `idPlatform`,
	RENAME COLUMN `text_id` TO `sTextId`,
	RENAME COLUMN `repository_path` TO `sRepositoryPath`,
	RENAME COLUMN `type` TO `sType`,
	RENAME COLUMN `title_bar_visible` TO `bTitleBarVisible`,
	RENAME COLUMN `transparent_folder` TO `bTransparentFolder`,
	RENAME COLUMN `display_details_in_parent` TO `bDisplayDetailsInParent`,
	RENAME COLUMN `custom_chapter` TO `bCustomChapter`,
	RENAME COLUMN `display_children_as_tabs` TO `bDisplayChildrenAsTabs`,
	RENAME COLUMN `uses_api` TO `bUsesAPI`,
	RENAME COLUMN `read_only` TO `bReadOnly`,
	RENAME COLUMN `full_screen` TO `sFullScreen`,
	RENAME COLUMN `show_difficulty` TO `bShowDifficulty`,
	RENAME COLUMN `show_source` TO `bShowSource`,
	RENAME COLUMN `hints_allowed` TO `bHintsAllowed`,
	RENAME COLUMN `fixed_ranks` TO `bFixedRanks`,
	RENAME COLUMN `validation_type` TO `sValidationType`,
	RENAME COLUMN `validation_min` TO `iValidationMin`,
	RENAME COLUMN `preparation_state` TO `sPreparationState`,
	RENAME COLUMN `unlocked_item_ids` TO `idItemUnlocked`,
	RENAME COLUMN `score_min_unlock` TO `iScoreMinUnlock`,
	RENAME COLUMN `supported_lang_prog` TO `sSupportedLangProg`,
	RENAME COLUMN `default_language_id` TO `idDefaultLanguage`,
	RENAME COLUMN `team_mode` TO `sTeamMode`,
	RENAME COLUMN `teams_editable` TO `bTeamsEditable`,
	RENAME COLUMN `qualified_group_id` TO `idTeamInGroup`,
	RENAME COLUMN `team_max_members` TO `iTeamMaxMembers`,
	RENAME COLUMN `has_attempts` TO `bHasAttempts`,
	RENAME COLUMN `access_open_date` TO `sAccessOpenDate`,
	RENAME COLUMN `duration` TO `sDuration`,
	RENAME COLUMN `end_contest_date` TO `sEndContestDate`,
	RENAME COLUMN `show_user_infos` TO `bShowUserInfos`,
	RENAME COLUMN `contest_phase` TO `sContestPhase`,
	RENAME COLUMN `level` TO `iLevel`,
	RENAME COLUMN `no_score` TO `bNoScore`,
	RENAME COLUMN `group_code_enter` TO `groupCodeEnter`,
	RENAME COLUMN `version` TO `iVersion`,
	RENAME COLUMN `next_version` TO `iNextVersion`,
	RENAME COLUMN `deleted` TO `bDeleted`,
	RENAME INDEX `id` TO `ID`,
	RENAME INDEX `version` TO `iVersion`,
	RENAME INDEX `next_version` TO `iNextVersion`,
	RENAME INDEX `deleted` TO `bDeleted`;
ALTER TABLE `history_items_ancestors`
	RENAME COLUMN `history_id` TO `historyID`,
	RENAME COLUMN `id` TO `ID`,
	RENAME COLUMN `ancestor_item_id` TO `idItemAncestor`,
	RENAME COLUMN `child_item_id` TO `idItemChild`,
	RENAME COLUMN `version` TO `iVersion`,
	RENAME COLUMN `next_version` TO `iNextVersion`,
	RENAME COLUMN `deleted` TO `bDeleted`,
	RENAME INDEX `version` TO `iVersion`,
	RENAME INDEX `next_version` TO `iNextVersion`,
	RENAME INDEX `deleted` TO `bDeleted`,
	RENAME INDEX `ancestor_item_id_child_item_id` TO `idItemAncestor`,
	RENAME INDEX `ancestor_item_id` TO `idItemAncestortor`,
	RENAME INDEX `child_item_id` TO `idItemChild`,
	RENAME INDEX `id` TO `ID`;
ALTER TABLE `history_items_items`
	RENAME COLUMN `history_id` TO `historyID`,
	RENAME COLUMN `id` TO `ID`,
	RENAME COLUMN `parent_item_id` TO `idItemParent`,
	RENAME COLUMN `child_item_id` TO `idItemChild`,
	RENAME COLUMN `child_order` TO `iChildOrder`,
	RENAME COLUMN `category` TO `sCategory`,
	RENAME COLUMN `always_visible` TO `bAlwaysVisible`,
	RENAME COLUMN `access_restricted` TO `bAccessRestricted`,
	RENAME COLUMN `difficulty` TO `iDifficulty`,
	RENAME COLUMN `version` TO `iVersion`,
	RENAME COLUMN `next_version` TO `iNextVersion`,
	RENAME COLUMN `deleted` TO `bDeleted`,
	RENAME INDEX `id` TO `ID`,
	RENAME INDEX `version` TO `iVersion`,
	RENAME INDEX `parent_item_id` TO `idItemParent`,
	RENAME INDEX `child_item_id` TO `idItemChild`,
	RENAME INDEX `next_version` TO `iNextVersion`,
	RENAME INDEX `deleted` TO `bDeleted`,
	RENAME INDEX `parent_child` TO `parentChild`;
ALTER TABLE `history_items_strings`
	RENAME COLUMN `history_id` TO `historyID`,
	RENAME COLUMN `id` TO `ID`,
	RENAME COLUMN `item_id` TO `idItem`,
	RENAME COLUMN `language_id` TO `idLanguage`,
	RENAME COLUMN `translator` TO `sTranslator`,
	RENAME COLUMN `title` TO `sTitle`,
	RENAME COLUMN `image_url` TO `sImageUrl`,
	RENAME COLUMN `subtitle` TO `sSubtitle`,
	RENAME COLUMN `description` TO `sDescription`,
	RENAME COLUMN `edu_comment` TO `sEduComment`,
	RENAME COLUMN `ranking_comment` TO `sRankingComment`,
	RENAME COLUMN `version` TO `iVersion`,
	RENAME COLUMN `next_version` TO `iNextVersion`,
	RENAME COLUMN `deleted` TO `bDeleted`,
	RENAME INDEX `id` TO `ID`,
	RENAME INDEX `version` TO `iVersion`,
	RENAME INDEX `item_language` TO `itemLanguage`,
	RENAME INDEX `item_id` TO `idItem`;
ALTER TABLE `history_languages`
	RENAME COLUMN `history_id` TO `historyID`,
	RENAME COLUMN `id` TO `ID`,
	RENAME COLUMN `name` TO `sName`,
	RENAME COLUMN `code` TO `sCode`,
	RENAME COLUMN `version` TO `iVersion`,
	RENAME COLUMN `next_version` TO `iNextVersion`,
	RENAME COLUMN `deleted` TO `bDeleted`,
	RENAME INDEX `id` TO `ID`,
	RENAME INDEX `version` TO `iVersion`,
	RENAME INDEX `code` TO `sCode`;
ALTER TABLE `history_messages`
	RENAME COLUMN `history_id` TO `history_ID`,
	RENAME COLUMN `id` TO `ID`,
	RENAME COLUMN `thread_id` TO `idThread`,
	RENAME COLUMN `user_id` TO `idUser`,
	RENAME COLUMN `submission_date` TO `sSubmissionDate`,
	RENAME COLUMN `published` TO `bPublished`,
	RENAME COLUMN `title` TO `sTitle`,
	RENAME COLUMN `body` TO `sBody`,
	RENAME COLUMN `trainers_only` TO `bTrainersOnly`,
	RENAME COLUMN `archived` TO `bArchived`,
	RENAME COLUMN `persistant` TO `bPersistant`,
	RENAME COLUMN `version` TO `iVersion`,
	RENAME COLUMN `next_version` TO `iNextVersion`,
	RENAME COLUMN `deleted` TO `bDeleted`,
	RENAME INDEX `version` TO `iVersion`,
	RENAME INDEX `next_version` TO `iNextVersion`,
	RENAME INDEX `deleted` TO `bDeleted`,
	RENAME INDEX `id` TO `ID`;
ALTER TABLE `history_threads`
	RENAME COLUMN `history_id` TO `history_ID`,
	RENAME COLUMN `id` TO `ID`,
	RENAME COLUMN `type` TO `sType`,
	RENAME COLUMN `creator_user_id` TO `idUserCreated`,
	RENAME COLUMN `item_id` TO `idItem`,
	RENAME COLUMN `last_activity_date` TO `sLastActivityDate`,
	RENAME COLUMN `title` TO `sTitle`,
	RENAME COLUMN `admin_help_asked` TO `bAdminHelpAsked`,
	RENAME COLUMN `hidden` TO `bHidden`,
	RENAME COLUMN `version` TO `iVersion`,
	RENAME COLUMN `next_version` TO `iNextVersion`,
	RENAME COLUMN `deleted` TO `bDeleted`,
	RENAME INDEX `version` TO `iVersion`,
	RENAME INDEX `next_version` TO `iNextVersion`,
	RENAME INDEX `deleted` TO `bDeleted`,
	RENAME INDEX `id` TO `ID`;
ALTER TABLE `history_users`
	RENAME COLUMN `history_id` TO `historyID`,
	RENAME COLUMN `id` TO `ID`,
	RENAME COLUMN `login_id` TO `loginID`,
	RENAME COLUMN `login` TO `sLogin`,
	RENAME COLUMN `open_id_identity` TO `sOpenIdIdentity`,
	RENAME COLUMN `password_md5` TO `sPasswordMd5`,
	RENAME COLUMN `salt` TO `sSalt`,
	RENAME COLUMN `recover` TO `sRecover`,
	RENAME COLUMN `registration_date` TO `sRegistrationDate`,
	RENAME COLUMN `email` TO `sEmail`,
	RENAME COLUMN `email_verified` TO `bEmailVerified`,
	RENAME COLUMN `first_name` TO `sFirstName`,
	RENAME COLUMN `last_name` TO `sLastName`,
	RENAME COLUMN `student_id` TO `sStudentId`,
	RENAME COLUMN `country_code` TO `sCountryCode`,
	RENAME COLUMN `time_zone` TO `sTimeZone`,
	RENAME COLUMN `birth_date` TO `sBirthDate`,
	RENAME COLUMN `graduation_year` TO `iGraduationYear`,
	RENAME COLUMN `grade` TO `iGrade`,
	RENAME COLUMN `sex` TO `sSex`,
	RENAME COLUMN `address` TO `sAddress`,
	RENAME COLUMN `zipcode` TO `sZipcode`,
	RENAME COLUMN `city` TO `sCity`,
	RENAME COLUMN `land_line_number` TO `sLandLineNumber`,
	RENAME COLUMN `cell_phone_number` TO `sCellPhoneNumber`,
	RENAME COLUMN `default_language` TO `sDefaultLanguage`,
	RENAME COLUMN `notify_news` TO `bNotifyNews`,
	RENAME COLUMN `notify` TO `sNotify`,
	RENAME COLUMN `public_first_name` TO `bPublicFirstName`,
	RENAME COLUMN `public_last_name` TO `bPublicLastName`,
	RENAME COLUMN `free_text` TO `sFreeText`,
	RENAME COLUMN `web_site` TO `sWebSite`,
	RENAME COLUMN `photo_autoload` TO `bPhotoAutoload`,
	RENAME COLUMN `lang_prog` TO `sLangProg`,
	RENAME COLUMN `last_login_date` TO `sLastLoginDate`,
	RENAME COLUMN `last_activity_date` TO `sLastActivityDate`,
	RENAME COLUMN `last_ip` TO `sLastIP`,
	RENAME COLUMN `basic_editor_mode` TO `bBasicEditorMode`,
	RENAME COLUMN `spaces_for_tab` TO `nbSpacesForTab`,
	RENAME COLUMN `member_state` TO `iMemberState`,
	RENAME COLUMN `godfather_user_id` TO `idUserGodfather`,
	RENAME COLUMN `step_level_in_site` TO `iStepLevelInSite`,
	RENAME COLUMN `is_admin` TO `bIsAdmin`,
	RENAME COLUMN `no_ranking` TO `bNoRanking`,
	RENAME COLUMN `help_given` TO `nbHelpGiven`,
	RENAME COLUMN `self_group_id` TO `idGroupSelf`,
	RENAME COLUMN `owned_group_id` TO `idGroupOwned`,
	RENAME COLUMN `access_group_id` TO `idGroupAccess`,
	RENAME COLUMN `notification_read_date` TO `sNotificationReadDate`,
	RENAME COLUMN `version` TO `iVersion`,
	RENAME COLUMN `next_version` TO `iNextVersion`,
	RENAME COLUMN `deleted` TO `bDeleted`,
	RENAME COLUMN `login_module_prefix` TO `loginModulePrefix`,
	RENAME COLUMN `creator_id` TO `creatorID`,
	RENAME COLUMN `allow_subgroups` TO `allowSubgroups`,
	RENAME INDEX `id` TO `ID`,
	RENAME INDEX `version` TO `iVersion`,
	RENAME INDEX `country_code` TO `sCountryCode`,
	RENAME INDEX `godfather_user_id` TO `idUserGodfather`,
	RENAME INDEX `lang_prog` TO `sLangProg`,
	RENAME INDEX `next_version` TO `iNextVersion`,
	RENAME INDEX `deleted` TO `bDeleted`,
	RENAME INDEX `self_group_id` TO `idGroupSelf`,
	RENAME INDEX `owned_group_id` TO `idGroupOwned`;
ALTER TABLE `history_users_items`
	RENAME COLUMN `history_id` TO `historyID`,
	RENAME COLUMN `id` TO `ID`,
	RENAME COLUMN `user_id` TO `idUser`,
	RENAME COLUMN `item_id` TO `idItem`,
	RENAME COLUMN `active_attempt_id` TO `idAttemptActive`,
	RENAME COLUMN `score` TO `iScore`,
	RENAME COLUMN `score_computed` TO `iScoreComputed`,
	RENAME COLUMN `score_reeval` TO `iScoreReeval`,
	RENAME COLUMN `score_diff_manual` TO `iScoreDiffManual`,
	RENAME COLUMN `score_diff_comment` TO `sScoreDiffComment`,
	RENAME COLUMN `submissions_attempts` TO `nbSubmissionsAttempts`,
	RENAME COLUMN `tasks_tried` TO `nbTasksTried`,
	RENAME COLUMN `tasks_solved` TO `nbTasksSolved`,
	RENAME COLUMN `children_validated` TO `nbChildrenValidated`,
	RENAME COLUMN `validated` TO `bValidated`,
	RENAME COLUMN `finished` TO `bFinished`,
	RENAME COLUMN `key_obtained` TO `bKeyObtained`,
	RENAME COLUMN `tasks_with_help` TO `nbTasksWithHelp`,
	RENAME COLUMN `hints_requested` TO `sHintsRequested`,
	RENAME COLUMN `hints_cached` TO `nbHintsCached`,
	RENAME COLUMN `corrections_read` TO `nbCorrectionsRead`,
	RENAME COLUMN `precision` TO `iPrecision`,
	RENAME COLUMN `autonomy` TO `iAutonomy`,
	RENAME COLUMN `start_date` TO `sStartDate`,
	RENAME COLUMN `validation_date` TO `sValidationDate`,
	RENAME COLUMN `finish_date` TO `sFinishDate`,
	RENAME COLUMN `last_activity_date` TO `sLastActivityDate`,
	RENAME COLUMN `thread_start_date` TO `sThreadStartDate`,
	RENAME COLUMN `best_answer_date` TO `sBestAnswerDate`,
	RENAME COLUMN `last_answer_date` TO `sLastAnswerDate`,
	RENAME COLUMN `last_hint_date` TO `sLastHintDate`,
	RENAME COLUMN `contest_start_date` TO `sContestStartDate`,
	RENAME COLUMN `ranked` TO `bRanked`,
	RENAME COLUMN `all_lang_prog` TO `sAllLangProg`,
	RENAME COLUMN `state` TO `sState`,
	RENAME COLUMN `answer` TO `sAnswer`,
	RENAME COLUMN `version` TO `iVersion`,
	RENAME COLUMN `next_version` TO `iNextVersion`,
	RENAME COLUMN `deleted` TO `bDeleted`,
	RENAME COLUMN `platform_data_removed` TO `bPlatformDataRemoved`,
	RENAME INDEX `id` TO `ID`,
	RENAME INDEX `version` TO `iVersion`,
	RENAME INDEX `item_user` TO `itemUser`,
	RENAME INDEX `next_version` TO `iNextVersion`,
	RENAME INDEX `deleted` TO `bDeleted`,
	RENAME INDEX `item_id` TO `idItem`,
	RENAME INDEX `user_id` TO `idUser`;
ALTER TABLE `history_users_threads`
	RENAME COLUMN `history_id` TO `history_ID`,
	RENAME COLUMN `id` TO `ID`,
	RENAME COLUMN `user_id` TO `idUser`,
	RENAME COLUMN `thread_id` TO `idThread`,
	RENAME COLUMN `last_read_date` TO `sLastReadDate`,
	RENAME COLUMN `participated` TO `bParticipated`,
	RENAME COLUMN `last_write_date` TO `sLastWriteDate`,
	RENAME COLUMN `starred` TO `bStarred`,
	RENAME COLUMN `version` TO `iVersion`,
	RENAME COLUMN `next_version` TO `iNextVersion`,
	RENAME COLUMN `deleted` TO `bDeleted`,
	RENAME INDEX `user_thread` TO `userThread`,
	RENAME INDEX `version` TO `iVersion`,
	RENAME INDEX `next_version` TO `iNextVersion`,
	RENAME INDEX `deleted` TO `bDeleted`,
	RENAME INDEX `id` TO `ID`;
ALTER TABLE `items`
	RENAME COLUMN `id` TO `ID`,
	RENAME COLUMN `url` TO `sUrl`,
	RENAME COLUMN `platform_id` TO `idPlatform`,
	RENAME COLUMN `text_id` TO `sTextId`,
	RENAME COLUMN `repository_path` TO `sRepositoryPath`,
	RENAME COLUMN `type` TO `sType`,
	RENAME COLUMN `title_bar_visible` TO `bTitleBarVisible`,
	RENAME COLUMN `transparent_folder` TO `bTransparentFolder`,
	RENAME COLUMN `display_details_in_parent` TO `bDisplayDetailsInParent`,
	RENAME COLUMN `custom_chapter` TO `bCustomChapter`,
	RENAME COLUMN `display_children_as_tabs` TO `bDisplayChildrenAsTabs`,
	RENAME COLUMN `uses_api` TO `bUsesAPI`,
	RENAME COLUMN `read_only` TO `bReadOnly`,
	RENAME COLUMN `full_screen` TO `sFullScreen`,
	RENAME COLUMN `show_difficulty` TO `bShowDifficulty`,
	RENAME COLUMN `show_source` TO `bShowSource`,
	RENAME COLUMN `hints_allowed` TO `bHintsAllowed`,
	RENAME COLUMN `fixed_ranks` TO `bFixedRanks`,
	RENAME COLUMN `validation_type` TO `sValidationType`,
	RENAME COLUMN `validation_min` TO `iValidationMin`,
	RENAME COLUMN `preparation_state` TO `sPreparationState`,
	RENAME COLUMN `unlocked_item_ids` TO `idItemUnlocked`,
	RENAME COLUMN `score_min_unlock` TO `iScoreMinUnlock`,
	RENAME COLUMN `supported_lang_prog` TO `sSupportedLangProg`,
	RENAME COLUMN `default_language_id` TO `idDefaultLanguage`,
	RENAME COLUMN `team_mode` TO `sTeamMode`,
	RENAME COLUMN `teams_editable` TO `bTeamsEditable`,
	RENAME COLUMN `qualified_group_id` TO `idTeamInGroup`,
	RENAME COLUMN `team_max_members` TO `iTeamMaxMembers`,
	RENAME COLUMN `has_attempts` TO `bHasAttempts`,
	RENAME COLUMN `access_open_date` TO `sAccessOpenDate`,
	RENAME COLUMN `duration` TO `sDuration`,
	RENAME COLUMN `end_contest_date` TO `sEndContestDate`,
	RENAME COLUMN `show_user_infos` TO `bShowUserInfos`,
	RENAME COLUMN `contest_phase` TO `sContestPhase`,
	RENAME COLUMN `level` TO `iLevel`,
	RENAME COLUMN `no_score` TO `bNoScore`,
	RENAME COLUMN `group_code_enter` TO `groupCodeEnter`,
	RENAME COLUMN `version` TO `iVersion`,
	RENAME INDEX `version` TO `iVersion`;
ALTER TABLE `items_ancestors`
	RENAME COLUMN `id` TO `ID`,
	RENAME COLUMN `ancestor_item_id` TO `idItemAncestor`,
	RENAME COLUMN `child_item_id` TO `idItemChild`,
	RENAME COLUMN `version` TO `iVersion`,
	RENAME INDEX `ancestor_item_id_child_item_id` TO `idItemAncestor`,
	RENAME INDEX `ancestor_item_id` TO `idItemAncestortor`,
	RENAME INDEX `child_item_id` TO `idItemChild`;
ALTER TABLE `items_items`
	RENAME COLUMN `id` TO `ID`,
	RENAME COLUMN `parent_item_id` TO `idItemParent`,
	RENAME COLUMN `child_item_id` TO `idItemChild`,
	RENAME COLUMN `child_order` TO `iChildOrder`,
	RENAME COLUMN `category` TO `sCategory`,
	RENAME COLUMN `always_visible` TO `bAlwaysVisible`,
	RENAME COLUMN `access_restricted` TO `bAccessRestricted`,
	RENAME COLUMN `difficulty` TO `iDifficulty`,
	RENAME COLUMN `version` TO `iVersion`,
	RENAME INDEX `parent_item_id` TO `idItemParent`,
	RENAME INDEX `child_item_id` TO `idItemChild`,
	RENAME INDEX `version` TO `iVersion`,
	RENAME INDEX `parent_child` TO `parentChild`,
	RENAME INDEX `parent_version` TO `parentVersion`;
ALTER TABLE `items_propagate`
	RENAME COLUMN `id` TO `ID`,
	RENAME COLUMN `ancestors_computation_state` TO `sAncestorsComputationState`,
	RENAME INDEX `ancestors_computation_date` TO `sAncestorsComputationDate`;
ALTER TABLE `items_strings`
	RENAME COLUMN `id` TO `ID`,
	RENAME COLUMN `item_id` TO `idItem`,
	RENAME COLUMN `language_id` TO `idLanguage`,
	RENAME COLUMN `translator` TO `sTranslator`,
	RENAME COLUMN `title` TO `sTitle`,
	RENAME COLUMN `image_url` TO `sImageUrl`,
	RENAME COLUMN `subtitle` TO `sSubtitle`,
	RENAME COLUMN `description` TO `sDescription`,
	RENAME COLUMN `edu_comment` TO `sEduComment`,
	RENAME COLUMN `ranking_comment` TO `sRankingComment`,
	RENAME COLUMN `version` TO `iVersion`,
	RENAME INDEX `item_id_language_id` TO `idItem`,
	RENAME INDEX `version` TO `iVersion`,
	RENAME INDEX `item_id` TO `idItemAlone`;
ALTER TABLE `languages`
	RENAME COLUMN `id` TO `ID`,
	RENAME COLUMN `name` TO `sName`,
	RENAME COLUMN `code` TO `sCode`,
	RENAME COLUMN `version` TO `iVersion`,
	RENAME INDEX `version` TO `iVersion`,
	RENAME INDEX `code` TO `sCode`;
ALTER TABLE `login_states`
	RENAME COLUMN `cookie` TO `sCookie`,
	RENAME COLUMN `state` TO `sState`,
	RENAME COLUMN `expiration_date` TO `sExpirationDate`,
	RENAME INDEX `expiration_date` TO `sExpirationDate`;
ALTER TABLE `messages`
	RENAME COLUMN `id` TO `ID`,
	RENAME COLUMN `thread_id` TO `idThread`,
	RENAME COLUMN `user_id` TO `idUser`,
	RENAME COLUMN `submission_date` TO `sSubmissionDate`,
	RENAME COLUMN `published` TO `bPublished`,
	RENAME COLUMN `title` TO `sTitle`,
	RENAME COLUMN `body` TO `sBody`,
	RENAME COLUMN `trainers_only` TO `bTrainersOnly`,
	RENAME COLUMN `archived` TO `bArchived`,
	RENAME COLUMN `persistant` TO `bPersistant`,
	RENAME COLUMN `version` TO `iVersion`,
	RENAME INDEX `version` TO `iVersion`,
	RENAME INDEX `thread_id` TO `idThread`;
ALTER TABLE `platforms`
	RENAME COLUMN `id` TO `ID`,
	RENAME COLUMN `name` TO `sName`,
	RENAME COLUMN `base_url` TO `sBaseUrl`,
	RENAME COLUMN `public_key` TO `sPublicKey`,
	RENAME COLUMN `uses_tokens` TO `bUsesTokens`,
	RENAME COLUMN `regexp` TO `sRegexp`,
	RENAME COLUMN `priority` TO `iPriority`;
ALTER TABLE `refresh_tokens`
	RENAME COLUMN `user_id` TO `idUser`,
	RENAME COLUMN `refresh_token` TO `sRefreshToken`,
	RENAME INDEX `refresh_token_prefix` TO `sRefreshTokenPrefix`;
ALTER TABLE `sessions`
	RENAME COLUMN `access_token` TO `sAccessToken`,
	RENAME COLUMN `user_id` TO `idUser`,
	RENAME COLUMN `expiration_date` TO `sExpirationDate`,
	RENAME COLUMN `issued_at_date` TO `sIssuedAtDate`,
	RENAME COLUMN `issuer` TO `sIssuer`,
	RENAME INDEX `expiration_date` TO `sExpirationDate`,
	RENAME INDEX `access_token_prefix` TO `sAccessTokenPrefix`,
	RENAME INDEX `user_id` TO `idUser`;
ALTER TABLE `synchro_version`
	RENAME COLUMN `id` TO `ID`,
	RENAME COLUMN `version` TO `iVersion`,
	RENAME COLUMN `last_server_version` TO `iLastServerVersion`,
	RENAME COLUMN `last_client_version` TO `iLastClientVersion`,
	RENAME INDEX `version` TO `iVersion`;
ALTER TABLE `threads`
	RENAME COLUMN `id` TO `ID`,
	RENAME COLUMN `type` TO `sType`,
	RENAME COLUMN `last_activity_date` TO `sLastActivityDate`,
	RENAME COLUMN `creator_user_id` TO `idUserCreated`,
	RENAME COLUMN `item_id` TO `idItem`,
	RENAME COLUMN `title` TO `sTitle`,
	RENAME COLUMN `admin_help_asked` TO `bAdminHelpAsked`,
	RENAME COLUMN `hidden` TO `bHidden`,
	RENAME COLUMN `version` TO `iVersion`,
	RENAME INDEX `version` TO `iVersion`;
ALTER TABLE `users`
	RENAME COLUMN `id` TO `ID`,
	RENAME COLUMN `login_id` TO `loginID`,
	RENAME COLUMN `temp_user` TO `tempUser`,
	RENAME COLUMN `login` TO `sLogin`,
	RENAME COLUMN `open_id_identity` TO `sOpenIdIdentity`,
	RENAME COLUMN `password_md5` TO `sPasswordMd5`,
	RENAME COLUMN `salt` TO `sSalt`,
	RENAME COLUMN `recover` TO `sRecover`,
	RENAME COLUMN `registration_date` TO `sRegistrationDate`,
	RENAME COLUMN `email` TO `sEmail`,
	RENAME COLUMN `email_verified` TO `bEmailVerified`,
	RENAME COLUMN `first_name` TO `sFirstName`,
	RENAME COLUMN `last_name` TO `sLastName`,
	RENAME COLUMN `student_id` TO `sStudentId`,
	RENAME COLUMN `country_code` TO `sCountryCode`,
	RENAME COLUMN `time_zone` TO `sTimeZone`,
	RENAME COLUMN `birth_date` TO `sBirthDate`,
	RENAME COLUMN `graduation_year` TO `iGraduationYear`,
	RENAME COLUMN `grade` TO `iGrade`,
	RENAME COLUMN `sex` TO `sSex`,
	RENAME COLUMN `address` TO `sAddress`,
	RENAME COLUMN `zipcode` TO `sZipcode`,
	RENAME COLUMN `city` TO `sCity`,
	RENAME COLUMN `land_line_number` TO `sLandLineNumber`,
	RENAME COLUMN `cell_phone_number` TO `sCellPhoneNumber`,
	RENAME COLUMN `default_language` TO `sDefaultLanguage`,
	RENAME COLUMN `notify_news` TO `bNotifyNews`,
	RENAME COLUMN `notify` TO `sNotify`,
	RENAME COLUMN `public_first_name` TO `bPublicFirstName`,
	RENAME COLUMN `public_last_name` TO `bPublicLastName`,
	RENAME COLUMN `free_text` TO `sFreeText`,
	RENAME COLUMN `web_site` TO `sWebSite`,
	RENAME COLUMN `photo_autoload` TO `bPhotoAutoload`,
	RENAME COLUMN `lang_prog` TO `sLangProg`,
	RENAME COLUMN `last_login_date` TO `sLastLoginDate`,
	RENAME COLUMN `last_activity_date` TO `sLastActivityDate`,
	RENAME COLUMN `last_ip` TO `sLastIP`,
	RENAME COLUMN `basic_editor_mode` TO `bBasicEditorMode`,
	RENAME COLUMN `spaces_for_tab` TO `nbSpacesForTab`,
	RENAME COLUMN `member_state` TO `iMemberState`,
	RENAME COLUMN `godfather_user_id` TO `idUserGodfather`,
	RENAME COLUMN `step_level_in_site` TO `iStepLevelInSite`,
	RENAME COLUMN `is_admin` TO `bIsAdmin`,
	RENAME COLUMN `no_ranking` TO `bNoRanking`,
	RENAME COLUMN `help_given` TO `nbHelpGiven`,
	RENAME COLUMN `self_group_id` TO `idGroupSelf`,
	RENAME COLUMN `owned_group_id` TO `idGroupOwned`,
	RENAME COLUMN `access_group_id` TO `idGroupAccess`,
	RENAME COLUMN `notification_read_date` TO `sNotificationReadDate`,
	RENAME COLUMN `version` TO `iVersion`,
	RENAME COLUMN `login_module_prefix` TO `loginModulePrefix`,
	RENAME COLUMN `creator_id` TO `creatorID`,
	RENAME COLUMN `allow_subgroups` TO `allowSubgroups`,
	RENAME INDEX `login` TO `sLogin`,
	RENAME INDEX `self_group_id` TO `idGroupSelf`,
	RENAME INDEX `owned_group_id` TO `idGroupOwned`,
	RENAME INDEX `version` TO `iVersion`,
	RENAME INDEX `country_code` TO `sCountryCode`,
	RENAME INDEX `godfather_user_id` TO `idUserGodfather`,
	RENAME INDEX `lang_prog` TO `sLangProg`,
	RENAME INDEX `login_id` TO `loginID`,
	RENAME INDEX `temp_user` TO `tempUser`;
ALTER TABLE `users_answers`
	RENAME COLUMN `id` TO `ID`,
	RENAME COLUMN `user_id` TO `idUser`,
	RENAME COLUMN `item_id` TO `idItem`,
	RENAME COLUMN `attempt_id` TO `idAttempt`,
	RENAME COLUMN `name` TO `sName`,
	RENAME COLUMN `type` TO `sType`,
	RENAME COLUMN `state` TO `sState`,
	RENAME COLUMN `answer` TO `sAnswer`,
	RENAME COLUMN `lang_prog` TO `sLangProg`,
	RENAME COLUMN `submission_date` TO `sSubmissionDate`,
	RENAME COLUMN `score` TO `iScore`,
	RENAME COLUMN `validated` TO `bValidated`,
	RENAME COLUMN `grading_date` TO `sGradingDate`,
	RENAME COLUMN `grader_user_id` TO `idUserGrader`,
	RENAME INDEX `user_id` TO `idUser`,
	RENAME INDEX `item_id` TO `idItem`,
	RENAME INDEX `attempt_id` TO `idAttempt`;
ALTER TABLE `users_items`
	RENAME COLUMN `id` TO `ID`,
	RENAME COLUMN `user_id` TO `idUser`,
	RENAME COLUMN `item_id` TO `idItem`,
	RENAME COLUMN `active_attempt_id` TO `idAttemptActive`,
	RENAME COLUMN `score` TO `iScore`,
	RENAME COLUMN `score_computed` TO `iScoreComputed`,
	RENAME COLUMN `score_reeval` TO `iScoreReeval`,
	RENAME COLUMN `score_diff_manual` TO `iScoreDiffManual`,
	RENAME COLUMN `score_diff_comment` TO `sScoreDiffComment`,
	RENAME COLUMN `submissions_attempts` TO `nbSubmissionsAttempts`,
	RENAME COLUMN `tasks_tried` TO `nbTasksTried`,
	RENAME COLUMN `tasks_solved` TO `nbTasksSolved`,
	RENAME COLUMN `children_validated` TO `nbChildrenValidated`,
	RENAME COLUMN `validated` TO `bValidated`,
	RENAME COLUMN `finished` TO `bFinished`,
	RENAME COLUMN `key_obtained` TO `bKeyObtained`,
	RENAME COLUMN `tasks_with_help` TO `nbTasksWithHelp`,
	RENAME COLUMN `hints_requested` TO `sHintsRequested`,
	RENAME COLUMN `hints_cached` TO `nbHintsCached`,
	RENAME COLUMN `corrections_read` TO `nbCorrectionsRead`,
	RENAME COLUMN `precision` TO `iPrecision`,
	RENAME COLUMN `autonomy` TO `iAutonomy`,
	RENAME COLUMN `start_date` TO `sStartDate`,
	RENAME COLUMN `validation_date` TO `sValidationDate`,
	RENAME COLUMN `finish_date` TO `sFinishDate`,
	RENAME COLUMN `last_activity_date` TO `sLastActivityDate`,
	RENAME COLUMN `thread_start_date` TO `sThreadStartDate`,
	RENAME COLUMN `best_answer_date` TO `sBestAnswerDate`,
	RENAME COLUMN `last_answer_date` TO `sLastAnswerDate`,
	RENAME COLUMN `last_hint_date` TO `sLastHintDate`,
	RENAME COLUMN `contest_start_date` TO `sContestStartDate`,
	RENAME COLUMN `ranked` TO `bRanked`,
	RENAME COLUMN `all_lang_prog` TO `sAllLangProg`,
	RENAME COLUMN `version` TO `iVersion`,
	RENAME COLUMN `ancestors_computation_state` TO `sAncestorsComputationState`,
	RENAME COLUMN `state` TO `sState`,
	RENAME COLUMN `answer` TO `sAnswer`,
	RENAME COLUMN `platform_data_removed` TO `bPlatformDataRemoved`,
	RENAME INDEX `user_item` TO `UserItem`,
	RENAME INDEX `version` TO `iVersion`,
	RENAME INDEX `ancestors_computation_state` TO `sAncestorsComputationState`,
	RENAME INDEX `item_id` TO `idItem`,
	RENAME INDEX `user_id` TO `idUser`,
	RENAME INDEX `active_attempt_id` TO `idAttemptActive`;
ALTER TABLE `users_threads`
	RENAME COLUMN `id` TO `ID`,
	RENAME COLUMN `user_id` TO `idUser`,
	RENAME COLUMN `thread_id` TO `idThread`,
	RENAME COLUMN `last_read_date` TO `sLastReadDate`,
	RENAME COLUMN `participated` TO `bParticipated`,
	RENAME COLUMN `last_write_date` TO `sLastWriteDate`,
	RENAME COLUMN `starred` TO `bStarred`,
	RENAME COLUMN `version` TO `iVersion`,
	RENAME INDEX `user_thread` TO `userThread`,
	RENAME INDEX `version` TO `iVersion`;


ALTER TABLE `groups`
	MODIFY `sTextId` varchar(255) NOT NULL DEFAULT '' COMMENT 'Internal text ID for special groups. Used to refer o them and avoid breaking features if an admin renames the group',
	MODIFY `iTeamParticipating` tinyint(1) NOT NULL DEFAULT '0' COMMENT 'Did the team start the item it is associated to (from idTeamItem)?',
	MODIFY `bOpenContest` tinyint(1) NOT NULL DEFAULT '0' COMMENT 'If true and the group is associated through sRedirectPath with an item that is a contest, the contest should be started for this user as soon as he joins the group.';
ALTER TABLE `groups_ancestors`
	MODIFY `bIsSelf` tinyint(1) NOT NULL DEFAULT '0' COMMENT 'Whether idGroupAncestor = idGroupChild.';
ALTER TABLE `groups_attempts`
	MODIFY `bKeyObtained` tinyint(1) NOT NULL DEFAULT '0' COMMENT 'Whether the user obtained the key on this item (changed to 1 if the user gets a score >= items.iScoreMinUnlock, will grant access to new items from items.idItemUnlocked). This information is propagated to users_items.';
ALTER TABLE `items`
	MODIFY `bFixedRanks` tinyint(1) NOT NULL DEFAULT '0' COMMENT 'If true, prevents users from changing the order of the children by drag&drop and auto-calculation of the iOrder of children. Allows for manual setting of the iOrder, for instance in cases where we want to have multiple items with the same iOrder (check items_items.iChildOrder).',
	MODIFY `iScoreMinUnlock` int(11) NOT NULL DEFAULT '100' COMMENT 'Minimum score to obtain so that the item, indicated by "idItemUnlocked", is actually unlocked',
	MODIFY `sTeamMode` enum('All','Half','One','None') DEFAULT NULL COMMENT 'If idTeamInGroup is not NULL, this field specifies how many team members need to belong to that group in order for the whole team to be qualified and able to start the item.',
	MODIFY `idTeamInGroup` bigint(20) DEFAULT NULL COMMENT 'group ID in which "qualified" users will belong. sTeamMode dictates how many of a team''s members must be "qualified" in order to start the item.';
ALTER TABLE `items_items`
	MODIFY `iChildOrder` int(11) NOT NULL COMMENT 'Position, relative to its siblings, when displaying all the children of the parent. If multiple items have the same iChildOrder, they will be sorted in a random way, specific to each user (a user will always see the items in the same order).';
ALTER TABLE `users_items`
	MODIFY `bKeyObtained` tinyint(1) NOT NULL DEFAULT '0' COMMENT 'Whether the user obtained the key on this item. Changed to 1 if the user gets a score >= items.iScoreMinUnlock, will grant access to new item from items.idItemUnlocked. This information is propagated to users_items.';


DROP TRIGGER `before_insert_filters`;
-- +migrate StatementBegin
CREATE TRIGGER `before_insert_filters` BEFORE INSERT ON `filters` FOR EACH ROW BEGIN IF (NEW.ID IS NULL OR NEW.ID = 0) THEN SET NEW.ID = FLOOR(RAND() * 1000000000) + FLOOR(RAND() * 1000000000) * 1000000000; END IF ; SELECT ROUND(UNIX_TIMESTAMP(CURTIME(2)) * 10) INTO @curVersion;SET NEW.iVersion = @curVersion; END
-- +migrate StatementEnd
DROP TRIGGER `after_insert_filters`;
-- +migrate StatementBegin
CREATE TRIGGER `after_insert_filters` AFTER INSERT ON `filters` FOR EACH ROW BEGIN INSERT INTO `history_filters` (`ID`,`iVersion`,`idUser`,`sName`,`bSelected`,`bStarred`,`sStartDate`,`sEndDate`,`bArchived`,`bParticipated`,`bUnread`,`idItem`,`idGroup`,`olderThan`,`newerThan`,`sUsersSearch`,`sBodySearch`,`bImportant`) VALUES (NEW.`ID`,@curVersion,NEW.`idUser`,NEW.`sName`,NEW.`bSelected`,NEW.`bStarred`,NEW.`sStartDate`,NEW.`sEndDate`,NEW.`bArchived`,NEW.`bParticipated`,NEW.`bUnread`,NEW.`idItem`,NEW.`idGroup`,NEW.`olderThan`,NEW.`newerThan`,NEW.`sUsersSearch`,NEW.`sBodySearch`,NEW.`bImportant`); END
-- +migrate StatementEnd
DROP TRIGGER `before_update_filters`;
-- +migrate StatementBegin
CREATE TRIGGER `before_update_filters` BEFORE UPDATE ON `filters` FOR EACH ROW BEGIN IF NEW.iVersion <> OLD.iVersion THEN SET @curVersion = NEW.iVersion; ELSE SELECT ROUND(UNIX_TIMESTAMP(CURTIME(2)) * 10) INTO @curVersion; END IF; IF NOT (OLD.`ID` = NEW.`ID` AND OLD.`idUser` <=> NEW.`idUser` AND OLD.`sName` <=> NEW.`sName` AND OLD.`bStarred` <=> NEW.`bStarred` AND OLD.`sStartDate` <=> NEW.`sStartDate` AND OLD.`sEndDate` <=> NEW.`sEndDate` AND OLD.`bArchived` <=> NEW.`bArchived` AND OLD.`bParticipated` <=> NEW.`bParticipated` AND OLD.`bUnread` <=> NEW.`bUnread` AND OLD.`idItem` <=> NEW.`idItem` AND OLD.`idGroup` <=> NEW.`idGroup` AND OLD.`olderThan` <=> NEW.`olderThan` AND OLD.`newerThan` <=> NEW.`newerThan` AND OLD.`sUsersSearch` <=> NEW.`sUsersSearch` AND OLD.`sBodySearch` <=> NEW.`sBodySearch` AND OLD.`bImportant` <=> NEW.`bImportant`) THEN   SET NEW.iVersion = @curVersion;   UPDATE `history_filters` SET `iNextVersion` = @curVersion WHERE `ID` = OLD.`ID` AND `iNextVersion` IS NULL;   INSERT INTO `history_filters` (`ID`,`iVersion`,`idUser`,`sName`,`bSelected`,`bStarred`,`sStartDate`,`sEndDate`,`bArchived`,`bParticipated`,`bUnread`,`idItem`,`idGroup`,`olderThan`,`newerThan`,`sUsersSearch`,`sBodySearch`,`bImportant`)       VALUES (NEW.`ID`,@curVersion,NEW.`idUser`,NEW.`sName`,NEW.`bSelected`,NEW.`bStarred`,NEW.`sStartDate`,NEW.`sEndDate`,NEW.`bArchived`,NEW.`bParticipated`,NEW.`bUnread`,NEW.`idItem`,NEW.`idGroup`,NEW.`olderThan`,NEW.`newerThan`,NEW.`sUsersSearch`,NEW.`sBodySearch`,NEW.`bImportant`) ; END IF; END
-- +migrate StatementEnd
DROP TRIGGER `before_delete_filters`;
-- +migrate StatementBegin
CREATE TRIGGER `before_delete_filters` BEFORE DELETE ON `filters` FOR EACH ROW BEGIN SELECT ROUND(UNIX_TIMESTAMP(CURTIME(2)) * 10) INTO @curVersion; UPDATE `history_filters` SET `iNextVersion` = @curVersion WHERE `ID` = OLD.`ID` AND `iNextVersion` IS NULL; INSERT INTO `history_filters` (`ID`,`iVersion`,`idUser`,`sName`,`bSelected`,`bStarred`,`sStartDate`,`sEndDate`,`bArchived`,`bParticipated`,`bUnread`,`idItem`,`idGroup`,`olderThan`,`newerThan`,`sUsersSearch`,`sBodySearch`,`bImportant`, `bDeleted`) VALUES (OLD.`ID`,@curVersion,OLD.`idUser`,OLD.`sName`,OLD.`bSelected`,OLD.`bStarred`,OLD.`sStartDate`,OLD.`sEndDate`,OLD.`bArchived`,OLD.`bParticipated`,OLD.`bUnread`,OLD.`idItem`,OLD.`idGroup`,OLD.`olderThan`,OLD.`newerThan`,OLD.`sUsersSearch`,OLD.`sBodySearch`,OLD.`bImportant`, 1); END
-- +migrate StatementEnd
DROP TRIGGER `before_insert_groups`;
-- +migrate StatementBegin
CREATE TRIGGER `before_insert_groups` BEFORE INSERT ON `groups` FOR EACH ROW BEGIN IF (NEW.ID IS NULL OR NEW.ID = 0) THEN SET NEW.ID = FLOOR(RAND() * 1000000000) + FLOOR(RAND() * 1000000000) * 1000000000; END IF ; SELECT ROUND(UNIX_TIMESTAMP(CURTIME(2)) * 10) INTO @curVersion;SET NEW.iVersion = @curVersion; END
-- +migrate StatementEnd
DROP TRIGGER `after_insert_groups`;
-- +migrate StatementBegin
CREATE TRIGGER `after_insert_groups` AFTER INSERT ON `groups` FOR EACH ROW BEGIN INSERT INTO `history_groups` (`ID`,`iVersion`,`sName`,`iGrade`,`sGradeDetails`,`sDescription`,`sDateCreated`,`bOpened`,`bFreeAccess`,`idTeamItem`,`iTeamParticipating`,`sCode`,`sCodeTimer`,`sCodeEnd`,`sRedirectPath`,`bOpenContest`,`sType`,`bSendEmails`) VALUES (NEW.`ID`,@curVersion,NEW.`sName`,NEW.`iGrade`,NEW.`sGradeDetails`,NEW.`sDescription`,NEW.`sDateCreated`,NEW.`bOpened`,NEW.`bFreeAccess`,NEW.`idTeamItem`,NEW.`iTeamParticipating`,NEW.`sCode`,NEW.`sCodeTimer`,NEW.`sCodeEnd`,NEW.`sRedirectPath`,NEW.`bOpenContest`,NEW.`sType`,NEW.`bSendEmails`); INSERT IGNORE INTO `groups_propagate` (`ID`, `sAncestorsComputationState`) VALUES (NEW.`ID`, 'todo') ; END
-- +migrate StatementEnd
DROP TRIGGER `before_update_groups`;
-- +migrate StatementBegin
CREATE TRIGGER `before_update_groups` BEFORE UPDATE ON `groups` FOR EACH ROW BEGIN IF NEW.iVersion <> OLD.iVersion THEN SET @curVersion = NEW.iVersion; ELSE SELECT ROUND(UNIX_TIMESTAMP(CURTIME(2)) * 10) INTO @curVersion; END IF; IF NOT (OLD.`ID` = NEW.`ID` AND OLD.`sName` <=> NEW.`sName` AND OLD.`iGrade` <=> NEW.`iGrade` AND OLD.`sGradeDetails` <=> NEW.`sGradeDetails` AND OLD.`sDescription` <=> NEW.`sDescription` AND OLD.`sDateCreated` <=> NEW.`sDateCreated` AND OLD.`bOpened` <=> NEW.`bOpened` AND OLD.`bFreeAccess` <=> NEW.`bFreeAccess` AND OLD.`idTeamItem` <=> NEW.`idTeamItem` AND OLD.`iTeamParticipating` <=> NEW.`iTeamParticipating` AND OLD.`sCode` <=> NEW.`sCode` AND OLD.`sCodeTimer` <=> NEW.`sCodeTimer` AND OLD.`sCodeEnd` <=> NEW.`sCodeEnd` AND OLD.`sRedirectPath` <=> NEW.`sRedirectPath` AND OLD.`bOpenContest` <=> NEW.`bOpenContest` AND OLD.`sType` <=> NEW.`sType` AND OLD.`bSendEmails` <=> NEW.`bSendEmails`) THEN   SET NEW.iVersion = @curVersion;   UPDATE `history_groups` SET `iNextVersion` = @curVersion WHERE `ID` = OLD.`ID` AND `iNextVersion` IS NULL;   INSERT INTO `history_groups` (`ID`,`iVersion`,`sName`,`iGrade`,`sGradeDetails`,`sDescription`,`sDateCreated`,`bOpened`,`bFreeAccess`,`idTeamItem`,`iTeamParticipating`,`sCode`,`sCodeTimer`,`sCodeEnd`,`sRedirectPath`,`bOpenContest`,`sType`,`bSendEmails`)       VALUES (NEW.`ID`,@curVersion,NEW.`sName`,NEW.`iGrade`,NEW.`sGradeDetails`,NEW.`sDescription`,NEW.`sDateCreated`,NEW.`bOpened`,NEW.`bFreeAccess`,NEW.`idTeamItem`,NEW.`iTeamParticipating`,NEW.`sCode`,NEW.`sCodeTimer`,NEW.`sCodeEnd`,NEW.`sRedirectPath`,NEW.`bOpenContest`,NEW.`sType`,NEW.`bSendEmails`) ; END IF; END
-- +migrate StatementEnd
DROP TRIGGER `before_delete_groups`;
-- +migrate StatementBegin
CREATE TRIGGER `before_delete_groups` BEFORE DELETE ON `groups` FOR EACH ROW BEGIN SELECT ROUND(UNIX_TIMESTAMP(CURTIME(2)) * 10) INTO @curVersion; UPDATE `history_groups` SET `iNextVersion` = @curVersion WHERE `ID` = OLD.`ID` AND `iNextVersion` IS NULL; INSERT INTO `history_groups` (`ID`,`iVersion`,`sName`,`iGrade`,`sGradeDetails`,`sDescription`,`sDateCreated`,`bOpened`,`bFreeAccess`,`idTeamItem`,`iTeamParticipating`,`sCode`,`sCodeTimer`,`sCodeEnd`,`sRedirectPath`,`bOpenContest`,`sType`,`bSendEmails`, `bDeleted`) VALUES (OLD.`ID`,@curVersion,OLD.`sName`,OLD.`iGrade`,OLD.`sGradeDetails`,OLD.`sDescription`,OLD.`sDateCreated`,OLD.`bOpened`,OLD.`bFreeAccess`,OLD.`idTeamItem`,OLD.`iTeamParticipating`,OLD.`sCode`,OLD.`sCodeTimer`,OLD.`sCodeEnd`,OLD.`sRedirectPath`,OLD.`bOpenContest`,OLD.`sType`,OLD.`bSendEmails`, 1); END
-- +migrate StatementEnd
DROP TRIGGER `after_delete_groups`;
-- +migrate StatementBegin
CREATE TRIGGER `after_delete_groups` AFTER DELETE ON `groups` FOR EACH ROW BEGIN DELETE FROM groups_propagate where ID = OLD.ID ; END
-- +migrate StatementEnd
DROP TRIGGER `before_insert_groups_ancestors`;
-- +migrate StatementBegin
CREATE TRIGGER `before_insert_groups_ancestors` BEFORE INSERT ON `groups_ancestors` FOR EACH ROW BEGIN IF (NEW.ID IS NULL OR NEW.ID = 0) THEN SET NEW.ID = FLOOR(RAND() * 1000000000) + FLOOR(RAND() * 1000000000) * 1000000000; END IF ; SELECT ROUND(UNIX_TIMESTAMP(CURTIME(2)) * 10) INTO @curVersion;SET NEW.iVersion = @curVersion; END
-- +migrate StatementEnd
DROP TRIGGER `after_insert_groups_ancestors`;
-- +migrate StatementBegin
CREATE TRIGGER `after_insert_groups_ancestors` AFTER INSERT ON `groups_ancestors` FOR EACH ROW BEGIN INSERT INTO `history_groups_ancestors` (`ID`,`iVersion`,`idGroupAncestor`,`idGroupChild`,`bIsSelf`) VALUES (NEW.`ID`,@curVersion,NEW.`idGroupAncestor`,NEW.`idGroupChild`,NEW.`bIsSelf`); END
-- +migrate StatementEnd
DROP TRIGGER `before_update_groups_ancestors`;
-- +migrate StatementBegin
CREATE TRIGGER `before_update_groups_ancestors` BEFORE UPDATE ON `groups_ancestors` FOR EACH ROW BEGIN IF NEW.iVersion <> OLD.iVersion THEN SET @curVersion = NEW.iVersion; ELSE SELECT ROUND(UNIX_TIMESTAMP(CURTIME(2)) * 10) INTO @curVersion; END IF; IF NOT (OLD.`ID` = NEW.`ID` AND OLD.`idGroupAncestor` <=> NEW.`idGroupAncestor` AND OLD.`idGroupChild` <=> NEW.`idGroupChild` AND OLD.`bIsSelf` <=> NEW.`bIsSelf`) THEN   SET NEW.iVersion = @curVersion;   UPDATE `history_groups_ancestors` SET `iNextVersion` = @curVersion WHERE `ID` = OLD.`ID` AND `iNextVersion` IS NULL;   INSERT INTO `history_groups_ancestors` (`ID`,`iVersion`,`idGroupAncestor`,`idGroupChild`,`bIsSelf`)       VALUES (NEW.`ID`,@curVersion,NEW.`idGroupAncestor`,NEW.`idGroupChild`,NEW.`bIsSelf`) ; END IF; END
-- +migrate StatementEnd
DROP TRIGGER `before_delete_groups_ancestors`;
-- +migrate StatementBegin
CREATE TRIGGER `before_delete_groups_ancestors` BEFORE DELETE ON `groups_ancestors` FOR EACH ROW BEGIN SELECT ROUND(UNIX_TIMESTAMP(CURTIME(2)) * 10) INTO @curVersion; UPDATE `history_groups_ancestors` SET `iNextVersion` = @curVersion WHERE `ID` = OLD.`ID` AND `iNextVersion` IS NULL; INSERT INTO `history_groups_ancestors` (`ID`,`iVersion`,`idGroupAncestor`,`idGroupChild`,`bIsSelf`, `bDeleted`) VALUES (OLD.`ID`,@curVersion,OLD.`idGroupAncestor`,OLD.`idGroupChild`,OLD.`bIsSelf`, 1); END
-- +migrate StatementEnd
DROP TRIGGER `before_insert_groups_attempts`;
-- +migrate StatementBegin
CREATE TRIGGER `before_insert_groups_attempts` BEFORE INSERT ON `groups_attempts` FOR EACH ROW BEGIN IF (NEW.ID IS NULL OR NEW.ID = 0) THEN SET NEW.ID = FLOOR(RAND() * 1000000000) + FLOOR(RAND() * 1000000000) * 1000000000; END IF ; SELECT ROUND(UNIX_TIMESTAMP(CURTIME(2)) * 10) INTO @curVersion;SET NEW.iVersion = @curVersion; SET NEW.iMinusScore = -NEW.iScore; END
-- +migrate StatementEnd
DROP TRIGGER `after_insert_groups_attempts`;
-- +migrate StatementBegin
CREATE TRIGGER `after_insert_groups_attempts` AFTER INSERT ON `groups_attempts` FOR EACH ROW BEGIN INSERT INTO `history_groups_attempts` (`ID`,`iVersion`,`idGroup`,`idItem`,`idUserCreator`,`iOrder`,`iScore`,`iScoreComputed`,`iScoreReeval`,`iScoreDiffManual`,`sScoreDiffComment`,`nbSubmissionsAttempts`,`nbTasksTried`,`nbChildrenValidated`,`bValidated`,`bFinished`,`bKeyObtained`,`nbTasksWithHelp`,`sHintsRequested`,`nbHintsCached`,`nbCorrectionsRead`,`iPrecision`,`iAutonomy`,`sStartDate`,`sValidationDate`,`sBestAnswerDate`,`sLastAnswerDate`,`sThreadStartDate`,`sLastHintDate`,`sFinishDate`,`sLastActivityDate`,`sContestStartDate`,`bRanked`,`sAllLangProg`) VALUES (NEW.`ID`,@curVersion,NEW.`idGroup`,NEW.`idItem`,NEW.`idUserCreator`,NEW.`iOrder`,NEW.`iScore`,NEW.`iScoreComputed`,NEW.`iScoreReeval`,NEW.`iScoreDiffManual`,NEW.`sScoreDiffComment`,NEW.`nbSubmissionsAttempts`,NEW.`nbTasksTried`,NEW.`nbChildrenValidated`,NEW.`bValidated`,NEW.`bFinished`,NEW.`bKeyObtained`,NEW.`nbTasksWithHelp`,NEW.`sHintsRequested`,NEW.`nbHintsCached`,NEW.`nbCorrectionsRead`,NEW.`iPrecision`,NEW.`iAutonomy`,NEW.`sStartDate`,NEW.`sValidationDate`,NEW.`sBestAnswerDate`,NEW.`sLastAnswerDate`,NEW.`sThreadStartDate`,NEW.`sLastHintDate`,NEW.`sFinishDate`,NEW.`sLastActivityDate`,NEW.`sContestStartDate`,NEW.`bRanked`,NEW.`sAllLangProg`); END
-- +migrate StatementEnd
DROP TRIGGER `before_update_groups_attempts`;
-- +migrate StatementBegin
CREATE TRIGGER `before_update_groups_attempts` BEFORE UPDATE ON `groups_attempts` FOR EACH ROW BEGIN IF NEW.iVersion <> OLD.iVersion THEN SET @curVersion = NEW.iVersion; ELSE SELECT ROUND(UNIX_TIMESTAMP(CURTIME(2)) * 10) INTO @curVersion; END IF; IF NOT (OLD.`ID` = NEW.`ID` AND OLD.`idGroup` <=> NEW.`idGroup` AND OLD.`idItem` <=> NEW.`idItem` AND OLD.`idUserCreator` <=> NEW.`idUserCreator` AND OLD.`iOrder` <=> NEW.`iOrder` AND OLD.`iScore` <=> NEW.`iScore` AND OLD.`iScoreComputed` <=> NEW.`iScoreComputed` AND OLD.`iScoreReeval` <=> NEW.`iScoreReeval` AND OLD.`iScoreDiffManual` <=> NEW.`iScoreDiffManual` AND OLD.`sScoreDiffComment` <=> NEW.`sScoreDiffComment` AND OLD.`nbSubmissionsAttempts` <=> NEW.`nbSubmissionsAttempts` AND OLD.`nbTasksTried` <=> NEW.`nbTasksTried` AND OLD.`nbChildrenValidated` <=> NEW.`nbChildrenValidated` AND OLD.`bValidated` <=> NEW.`bValidated` AND OLD.`bFinished` <=> NEW.`bFinished` AND OLD.`bKeyObtained` <=> NEW.`bKeyObtained` AND OLD.`nbTasksWithHelp` <=> NEW.`nbTasksWithHelp` AND OLD.`sHintsRequested` <=> NEW.`sHintsRequested` AND OLD.`nbHintsCached` <=> NEW.`nbHintsCached` AND OLD.`nbCorrectionsRead` <=> NEW.`nbCorrectionsRead` AND OLD.`iPrecision` <=> NEW.`iPrecision` AND OLD.`iAutonomy` <=> NEW.`iAutonomy` AND OLD.`sStartDate` <=> NEW.`sStartDate` AND OLD.`sValidationDate` <=> NEW.`sValidationDate` AND OLD.`sBestAnswerDate` <=> NEW.`sBestAnswerDate` AND OLD.`sLastAnswerDate` <=> NEW.`sLastAnswerDate` AND OLD.`sThreadStartDate` <=> NEW.`sThreadStartDate` AND OLD.`sLastHintDate` <=> NEW.`sLastHintDate` AND OLD.`sFinishDate` <=> NEW.`sFinishDate` AND OLD.`sContestStartDate` <=> NEW.`sContestStartDate` AND OLD.`bRanked` <=> NEW.`bRanked` AND OLD.`sAllLangProg` <=> NEW.`sAllLangProg`) THEN   SET NEW.iVersion = @curVersion;   UPDATE `history_groups_attempts` SET `iNextVersion` = @curVersion WHERE `ID` = OLD.`ID` AND `iNextVersion` IS NULL;   INSERT INTO `history_groups_attempts` (`ID`,`iVersion`,`idGroup`,`idItem`,`idUserCreator`,`iOrder`,`iScore`,`iScoreComputed`,`iScoreReeval`,`iScoreDiffManual`,`sScoreDiffComment`,`nbSubmissionsAttempts`,`nbTasksTried`,`nbChildrenValidated`,`bValidated`,`bFinished`,`bKeyObtained`,`nbTasksWithHelp`,`sHintsRequested`,`nbHintsCached`,`nbCorrectionsRead`,`iPrecision`,`iAutonomy`,`sStartDate`,`sValidationDate`,`sBestAnswerDate`,`sLastAnswerDate`,`sThreadStartDate`,`sLastHintDate`,`sFinishDate`,`sLastActivityDate`,`sContestStartDate`,`bRanked`,`sAllLangProg`)       VALUES (NEW.`ID`,@curVersion,NEW.`idGroup`,NEW.`idItem`,NEW.`idUserCreator`,NEW.`iOrder`,NEW.`iScore`,NEW.`iScoreComputed`,NEW.`iScoreReeval`,NEW.`iScoreDiffManual`,NEW.`sScoreDiffComment`,NEW.`nbSubmissionsAttempts`,NEW.`nbTasksTried`,NEW.`nbChildrenValidated`,NEW.`bValidated`,NEW.`bFinished`,NEW.`bKeyObtained`,NEW.`nbTasksWithHelp`,NEW.`sHintsRequested`,NEW.`nbHintsCached`,NEW.`nbCorrectionsRead`,NEW.`iPrecision`,NEW.`iAutonomy`,NEW.`sStartDate`,NEW.`sValidationDate`,NEW.`sBestAnswerDate`,NEW.`sLastAnswerDate`,NEW.`sThreadStartDate`,NEW.`sLastHintDate`,NEW.`sFinishDate`,NEW.`sLastActivityDate`,NEW.`sContestStartDate`,NEW.`bRanked`,NEW.`sAllLangProg`) ; SET NEW.iMinusScore = -NEW.iScore; END IF; END
-- +migrate StatementEnd
DROP TRIGGER `before_delete_groups_attempts`;
-- +migrate StatementBegin
CREATE TRIGGER `before_delete_groups_attempts` BEFORE DELETE ON `groups_attempts` FOR EACH ROW BEGIN SELECT ROUND(UNIX_TIMESTAMP(CURTIME(2)) * 10) INTO @curVersion; UPDATE `history_groups_attempts` SET `iNextVersion` = @curVersion WHERE `ID` = OLD.`ID` AND `iNextVersion` IS NULL; INSERT INTO `history_groups_attempts` (`ID`,`iVersion`,`idGroup`,`idItem`,`idUserCreator`,`iOrder`,`iScore`,`iScoreComputed`,`iScoreReeval`,`iScoreDiffManual`,`sScoreDiffComment`,`nbSubmissionsAttempts`,`nbTasksTried`,`nbChildrenValidated`,`bValidated`,`bFinished`,`bKeyObtained`,`nbTasksWithHelp`,`sHintsRequested`,`nbHintsCached`,`nbCorrectionsRead`,`iPrecision`,`iAutonomy`,`sStartDate`,`sValidationDate`,`sBestAnswerDate`,`sLastAnswerDate`,`sThreadStartDate`,`sLastHintDate`,`sFinishDate`,`sLastActivityDate`,`sContestStartDate`,`bRanked`,`sAllLangProg`, `bDeleted`) VALUES (OLD.`ID`,@curVersion,OLD.`idGroup`,OLD.`idItem`,OLD.`idUserCreator`,OLD.`iOrder`,OLD.`iScore`,OLD.`iScoreComputed`,OLD.`iScoreReeval`,OLD.`iScoreDiffManual`,OLD.`sScoreDiffComment`,OLD.`nbSubmissionsAttempts`,OLD.`nbTasksTried`,OLD.`nbChildrenValidated`,OLD.`bValidated`,OLD.`bFinished`,OLD.`bKeyObtained`,OLD.`nbTasksWithHelp`,OLD.`sHintsRequested`,OLD.`nbHintsCached`,OLD.`nbCorrectionsRead`,OLD.`iPrecision`,OLD.`iAutonomy`,OLD.`sStartDate`,OLD.`sValidationDate`,OLD.`sBestAnswerDate`,OLD.`sLastAnswerDate`,OLD.`sThreadStartDate`,OLD.`sLastHintDate`,OLD.`sFinishDate`,OLD.`sLastActivityDate`,OLD.`sContestStartDate`,OLD.`bRanked`,OLD.`sAllLangProg`, 1); END
-- +migrate StatementEnd
DROP TRIGGER `before_insert_groups_groups`;
-- +migrate StatementBegin
CREATE TRIGGER `before_insert_groups_groups` BEFORE INSERT ON `groups_groups` FOR EACH ROW BEGIN IF (NEW.ID IS NULL OR NEW.ID = 0) THEN SET NEW.ID = FLOOR(RAND() * 1000000000) + FLOOR(RAND() * 1000000000) * 1000000000; END IF ; SELECT ROUND(UNIX_TIMESTAMP(CURTIME(2)) * 10) INTO @curVersion;SET NEW.iVersion = @curVersion; INSERT IGNORE INTO `groups_propagate` (ID, sAncestorsComputationState) VALUES (NEW.idGroupChild, 'todo') ON DUPLICATE KEY UPDATE `sAncestorsComputationState` = 'todo' ; END
-- +migrate StatementEnd
DROP TRIGGER `after_insert_groups_groups`;
-- +migrate StatementBegin
CREATE TRIGGER `after_insert_groups_groups` AFTER INSERT ON `groups_groups` FOR EACH ROW BEGIN INSERT INTO `history_groups_groups` (`ID`,`iVersion`,`idGroupParent`,`idGroupChild`,`iChildOrder`,`sType`,`sRole`,`sStatusDate`,`idUserInviting`) VALUES (NEW.`ID`,@curVersion,NEW.`idGroupParent`,NEW.`idGroupChild`,NEW.`iChildOrder`,NEW.`sType`,NEW.`sRole`,NEW.`sStatusDate`,NEW.`idUserInviting`); END
-- +migrate StatementEnd
DROP TRIGGER `before_update_groups_groups`;
-- +migrate StatementBegin
CREATE TRIGGER `before_update_groups_groups` BEFORE UPDATE ON `groups_groups` FOR EACH ROW BEGIN
  IF NEW.iVersion <> OLD.iVersion THEN
    SET @curVersion = NEW.iVersion;
  ELSE
    SELECT ROUND(UNIX_TIMESTAMP(CURTIME(2)) * 10) INTO @curVersion;
  END IF;
  IF NOT (OLD.`ID` = NEW.`ID` AND OLD.`idGroupParent` <=> NEW.`idGroupParent` AND
          OLD.`idGroupChild` <=> NEW.`idGroupChild` AND OLD.`iChildOrder` <=> NEW.`iChildOrder`AND
          OLD.`sType` <=> NEW.`sType` AND OLD.`sRole` <=> NEW.`sRole` AND OLD.`sStatusDate` <=> NEW.`sStatusDate` AND
          OLD.`idUserInviting` <=> NEW.`idUserInviting`) THEN
    SET NEW.iVersion = @curVersion;
    UPDATE `history_groups_groups` SET `iNextVersion` = @curVersion WHERE `ID` = OLD.`ID` AND `iNextVersion` IS NULL;
    INSERT INTO `history_groups_groups` (
      `ID`,`iVersion`,`idGroupParent`,`idGroupChild`,`iChildOrder`,`sType`,`sRole`,`sStatusDate`,`idUserInviting`
    ) VALUES (
      NEW.`ID`,@curVersion,NEW.`idGroupParent`,NEW.`idGroupChild`,NEW.`iChildOrder`,NEW.`sType`,NEW.`sRole`,
      NEW.`sStatusDate`,NEW.`idUserInviting`
    );
  END IF;
  IF (OLD.idGroupChild != NEW.idGroupChild OR OLD.idGroupParent != NEW.idGroupParent OR OLD.sType != NEW.sType) THEN
    INSERT IGNORE INTO `groups_propagate` (`ID`, `sAncestorsComputationState`) VALUES (OLD.idGroupChild, 'todo')
      ON DUPLICATE KEY UPDATE `sAncestorsComputationState` = 'todo';
    INSERT IGNORE INTO `groups_propagate` (`ID`, `sAncestorsComputationState`) (
      SELECT `groups_ancestors`.`idGroupChild`, 'todo'
        FROM `groups_ancestors`
        WHERE `groups_ancestors`.`idGroupAncestor` = OLD.`idGroupChild`
    ) ON DUPLICATE KEY UPDATE `sAncestorsComputationState` = 'todo';
    DELETE `groups_ancestors` FROM `groups_ancestors`
      WHERE `groups_ancestors`.`idGroupChild` = OLD.`idGroupChild` AND
            `groups_ancestors`.`idGroupAncestor` = OLD.`idGroupParent`;
    DELETE `bridges` FROM `groups_ancestors` `child_descendants`
      JOIN `groups_ancestors` `parent_ancestors`
      JOIN `groups_ancestors` `bridges`
        ON (`bridges`.`idGroupAncestor` = `parent_ancestors`.`idGroupAncestor` AND
            `bridges`.`idGroupChild` = `child_descendants`.`idGroupChild`)
      WHERE `parent_ancestors`.`idGroupChild` = OLD.`idGroupParent` AND
            `child_descendants`.`idGroupAncestor` = OLD.`idGroupChild`;
    DELETE `child_ancestors` FROM `groups_ancestors` `child_ancestors`
      JOIN `groups_ancestors` `parent_ancestors`
        ON (`child_ancestors`.`idGroupChild` = OLD.`idGroupChild` AND
            `child_ancestors`.`idGroupAncestor` = `parent_ancestors`.`idGroupAncestor`)
      WHERE `parent_ancestors`.`idGroupChild` = OLD.`idGroupParent`;
    DELETE `parent_ancestors` FROM `groups_ancestors` `parent_ancestors`
      JOIN  `groups_ancestors` `child_ancestors`
        ON (`parent_ancestors`.`idGroupAncestor` = OLD.`idGroupParent` AND
            `child_ancestors`.`idGroupChild` = `parent_ancestors`.`idGroupChild`)
      WHERE `child_ancestors`.`idGroupAncestor` = OLD.`idGroupChild`;
  END IF;
  IF (OLD.idGroupChild != NEW.idGroupChild OR OLD.idGroupParent != NEW.idGroupParent OR OLD.sType != NEW.sType) THEN
    INSERT IGNORE INTO `groups_propagate` (ID, sAncestorsComputationState) VALUES (NEW.idGroupChild, 'todo')
      ON DUPLICATE KEY UPDATE `sAncestorsComputationState` = 'todo';
  END IF;
END
-- +migrate StatementEnd
DROP TRIGGER `before_delete_groups_groups`;
-- +migrate StatementBegin
CREATE TRIGGER `before_delete_groups_groups` BEFORE DELETE ON `groups_groups` FOR EACH ROW BEGIN
  SELECT ROUND(UNIX_TIMESTAMP(CURTIME(2)) * 10) INTO @curVersion;
  UPDATE `history_groups_groups` SET `iNextVersion` = @curVersion WHERE `ID` = OLD.`ID` AND `iNextVersion` IS NULL;
  INSERT INTO `history_groups_groups` (
    `ID`,`iVersion`,`idGroupParent`,`idGroupChild`,`iChildOrder`,`sType`,`sRole`,`sStatusDate`,`idUserInviting`,`bDeleted`
  ) VALUES (
    OLD.`ID`,@curVersion,OLD.`idGroupParent`,OLD.`idGroupChild`,OLD.`iChildOrder`,OLD.`sType`,OLD.`sRole`,
    OLD.`sStatusDate`,OLD.`idUserInviting`, 1
  );
  INSERT IGNORE INTO `groups_propagate` (`ID`, `sAncestorsComputationState`) VALUES (OLD.idGroupChild, 'todo')
    ON DUPLICATE KEY UPDATE `sAncestorsComputationState` = 'todo';
  INSERT IGNORE INTO `groups_propagate` (`ID`, `sAncestorsComputationState`) (
    SELECT `groups_ancestors`.`idGroupChild`, 'todo'
      FROM `groups_ancestors`
      WHERE `groups_ancestors`.`idGroupAncestor` = OLD.`idGroupChild`
  ) ON DUPLICATE KEY UPDATE `sAncestorsComputationState` = 'todo';
  DELETE `groups_ancestors` FROM `groups_ancestors`
    WHERE `groups_ancestors`.`idGroupChild` = OLD.`idGroupChild` AND
          `groups_ancestors`.`idGroupAncestor` = OLD.`idGroupParent`;
  DELETE `bridges`
    FROM `groups_ancestors` `child_descendants`
    JOIN `groups_ancestors` `parent_ancestors`
    JOIN `groups_ancestors` `bridges`
      ON (`bridges`.`idGroupAncestor` = `parent_ancestors`.`idGroupAncestor` AND
          `bridges`.`idGroupChild` = `child_descendants`.`idGroupChild`)
    WHERE `parent_ancestors`.`idGroupChild` = OLD.`idGroupParent` AND
          `child_descendants`.`idGroupAncestor` = OLD.`idGroupChild`;
  DELETE `child_ancestors`
    FROM `groups_ancestors` `child_ancestors`
    JOIN  `groups_ancestors` `parent_ancestors`
      ON (`child_ancestors`.`idGroupChild` = OLD.`idGroupChild` AND
          `child_ancestors`.`idGroupAncestor` = `parent_ancestors`.`idGroupAncestor`)
    WHERE `parent_ancestors`.`idGroupChild` = OLD.`idGroupParent`;
  DELETE `parent_ancestors`
    FROM `groups_ancestors` `parent_ancestors`
    JOIN  `groups_ancestors` `child_ancestors`
      ON (`parent_ancestors`.`idGroupAncestor` = OLD.`idGroupParent` AND
          `child_ancestors`.`idGroupChild` = `parent_ancestors`.`idGroupChild`)
    WHERE `child_ancestors`.`idGroupAncestor` = OLD.`idGroupChild`;
END
-- +migrate StatementEnd
DROP TRIGGER `before_insert_groups_items`;
-- +migrate StatementBegin
CREATE TRIGGER `before_insert_groups_items` BEFORE INSERT ON `groups_items` FOR EACH ROW BEGIN IF (NEW.ID IS NULL OR NEW.ID = 0) THEN SET NEW.ID = FLOOR(RAND() * 1000000000) + FLOOR(RAND() * 1000000000) * 1000000000; END IF ; SELECT ROUND(UNIX_TIMESTAMP(CURTIME(2)) * 10) INTO @curVersion;SET NEW.iVersion = @curVersion; SET NEW.`sPropagateAccess`='self' ; END
-- +migrate StatementEnd
DROP TRIGGER `after_insert_groups_items`;
-- +migrate StatementBegin
CREATE TRIGGER `after_insert_groups_items` AFTER INSERT ON `groups_items` FOR EACH ROW BEGIN INSERT INTO `history_groups_items` (`ID`,`iVersion`,`idGroup`,`idItem`,`idUserCreated`,`sPartialAccessDate`,`sFullAccessDate`,`sAccessReason`,`sAccessSolutionsDate`,`bOwnerAccess`,`bManagerAccess`,`sCachedPartialAccessDate`,`sCachedFullAccessDate`,`sCachedAccessSolutionsDate`,`sCachedGrayedAccessDate`,`bCachedFullAccess`,`bCachedPartialAccess`,`bCachedAccessSolutions`,`bCachedGrayedAccess`,`bCachedManagerAccess`,`sPropagateAccess`) VALUES (NEW.`ID`,@curVersion,NEW.`idGroup`,NEW.`idItem`,NEW.`idUserCreated`,NEW.`sPartialAccessDate`,NEW.`sFullAccessDate`,NEW.`sAccessReason`,NEW.`sAccessSolutionsDate`,NEW.`bOwnerAccess`,NEW.`bManagerAccess`,NEW.`sCachedPartialAccessDate`,NEW.`sCachedFullAccessDate`,NEW.`sCachedAccessSolutionsDate`,NEW.`sCachedGrayedAccessDate`,NEW.`bCachedFullAccess`,NEW.`bCachedPartialAccess`,NEW.`bCachedAccessSolutions`,NEW.`bCachedGrayedAccess`,NEW.`bCachedManagerAccess`,NEW.`sPropagateAccess`); END
-- +migrate StatementEnd
DROP TRIGGER `before_update_groups_items`;
-- +migrate StatementBegin
CREATE TRIGGER `before_update_groups_items` BEFORE UPDATE ON `groups_items` FOR EACH ROW BEGIN IF NEW.iVersion <> OLD.iVersion THEN SET @curVersion = NEW.iVersion; ELSE SELECT ROUND(UNIX_TIMESTAMP(CURTIME(2)) * 10) INTO @curVersion; END IF; IF NOT (OLD.`ID` = NEW.`ID` AND OLD.`idGroup` <=> NEW.`idGroup` AND OLD.`idItem` <=> NEW.`idItem` AND OLD.`idUserCreated` <=> NEW.`idUserCreated` AND OLD.`sPartialAccessDate` <=> NEW.`sPartialAccessDate` AND OLD.`sFullAccessDate` <=> NEW.`sFullAccessDate` AND OLD.`sAccessReason` <=> NEW.`sAccessReason` AND OLD.`sAccessSolutionsDate` <=> NEW.`sAccessSolutionsDate` AND OLD.`bOwnerAccess` <=> NEW.`bOwnerAccess` AND OLD.`bManagerAccess` <=> NEW.`bManagerAccess` AND OLD.`sCachedPartialAccessDate` <=> NEW.`sCachedPartialAccessDate` AND OLD.`sCachedFullAccessDate` <=> NEW.`sCachedFullAccessDate` AND OLD.`sCachedAccessSolutionsDate` <=> NEW.`sCachedAccessSolutionsDate` AND OLD.`sCachedGrayedAccessDate` <=> NEW.`sCachedGrayedAccessDate` AND OLD.`bCachedFullAccess` <=> NEW.`bCachedFullAccess` AND OLD.`bCachedPartialAccess` <=> NEW.`bCachedPartialAccess` AND OLD.`bCachedAccessSolutions` <=> NEW.`bCachedAccessSolutions` AND OLD.`bCachedGrayedAccess` <=> NEW.`bCachedGrayedAccess` AND OLD.`bCachedManagerAccess` <=> NEW.`bCachedManagerAccess`) THEN   SET NEW.iVersion = @curVersion;   UPDATE `history_groups_items` SET `iNextVersion` = @curVersion WHERE `ID` = OLD.`ID` AND `iNextVersion` IS NULL;   INSERT INTO `history_groups_items` (`ID`,`iVersion`,`idGroup`,`idItem`,`idUserCreated`,`sPartialAccessDate`,`sFullAccessDate`,`sAccessReason`,`sAccessSolutionsDate`,`bOwnerAccess`,`bManagerAccess`,`sCachedPartialAccessDate`,`sCachedFullAccessDate`,`sCachedAccessSolutionsDate`,`sCachedGrayedAccessDate`,`bCachedFullAccess`,`bCachedPartialAccess`,`bCachedAccessSolutions`,`bCachedGrayedAccess`,`bCachedManagerAccess`,`sPropagateAccess`)       VALUES (NEW.`ID`,@curVersion,NEW.`idGroup`,NEW.`idItem`,NEW.`idUserCreated`,NEW.`sPartialAccessDate`,NEW.`sFullAccessDate`,NEW.`sAccessReason`,NEW.`sAccessSolutionsDate`,NEW.`bOwnerAccess`,NEW.`bManagerAccess`,NEW.`sCachedPartialAccessDate`,NEW.`sCachedFullAccessDate`,NEW.`sCachedAccessSolutionsDate`,NEW.`sCachedGrayedAccessDate`,NEW.`bCachedFullAccess`,NEW.`bCachedPartialAccess`,NEW.`bCachedAccessSolutions`,NEW.`bCachedGrayedAccess`,NEW.`bCachedManagerAccess`,NEW.`sPropagateAccess`) ; END IF; IF NOT (NEW.`sFullAccessDate` <=> OLD.`sFullAccessDate`AND NEW.`sPartialAccessDate` <=> OLD.`sPartialAccessDate`AND NEW.`sAccessSolutionsDate` <=> OLD.`sAccessSolutionsDate`AND NEW.`bManagerAccess` <=> OLD.`bManagerAccess`AND NEW.`sAccessReason` <=> OLD.`sAccessReason`)THEN SET NEW.`sPropagateAccess` = 'self'; END IF; END
-- +migrate StatementEnd
DROP TRIGGER `before_delete_groups_items`;
-- +migrate StatementBegin
CREATE TRIGGER `before_delete_groups_items` BEFORE DELETE ON `groups_items` FOR EACH ROW BEGIN SELECT ROUND(UNIX_TIMESTAMP(CURTIME(2)) * 10) INTO @curVersion; UPDATE `history_groups_items` SET `iNextVersion` = @curVersion WHERE `ID` = OLD.`ID` AND `iNextVersion` IS NULL; INSERT INTO `history_groups_items` (`ID`,`iVersion`,`idGroup`,`idItem`,`idUserCreated`,`sPartialAccessDate`,`sFullAccessDate`,`sAccessReason`,`sAccessSolutionsDate`,`bOwnerAccess`,`bManagerAccess`,`sCachedPartialAccessDate`,`sCachedFullAccessDate`,`sCachedAccessSolutionsDate`,`sCachedGrayedAccessDate`,`bCachedFullAccess`,`bCachedPartialAccess`,`bCachedAccessSolutions`,`bCachedGrayedAccess`,`bCachedManagerAccess`,`sPropagateAccess`, `bDeleted`) VALUES (OLD.`ID`,@curVersion,OLD.`idGroup`,OLD.`idItem`,OLD.`idUserCreated`,OLD.`sPartialAccessDate`,OLD.`sFullAccessDate`,OLD.`sAccessReason`,OLD.`sAccessSolutionsDate`,OLD.`bOwnerAccess`,OLD.`bManagerAccess`,OLD.`sCachedPartialAccessDate`,OLD.`sCachedFullAccessDate`,OLD.`sCachedAccessSolutionsDate`,OLD.`sCachedGrayedAccessDate`,OLD.`bCachedFullAccess`,OLD.`bCachedPartialAccess`,OLD.`bCachedAccessSolutions`,OLD.`bCachedGrayedAccess`,OLD.`bCachedManagerAccess`,OLD.`sPropagateAccess`, 1); END
-- +migrate StatementEnd
DROP TRIGGER `after_delete_groups_items`;
-- +migrate StatementBegin
CREATE TRIGGER `after_delete_groups_items` AFTER DELETE ON `groups_items` FOR EACH ROW BEGIN DELETE FROM groups_items_propagate where ID = OLD.ID ; END
-- +migrate StatementEnd
DROP TRIGGER `before_insert_groups_login_prefixes`;
-- +migrate StatementBegin
CREATE TRIGGER `before_insert_groups_login_prefixes` BEFORE INSERT ON `groups_login_prefixes` FOR EACH ROW BEGIN IF (NEW.ID IS NULL OR NEW.ID = 0) THEN SET NEW.ID = FLOOR(RAND() * 1000000000) + FLOOR(RAND() * 1000000000) * 1000000000; END IF ; SELECT ROUND(UNIX_TIMESTAMP(CURTIME(2)) * 10) INTO @curVersion;SET NEW.iVersion = @curVersion; END
-- +migrate StatementEnd
DROP TRIGGER `after_insert_groups_login_prefixes`;
-- +migrate StatementBegin
CREATE TRIGGER `after_insert_groups_login_prefixes` AFTER INSERT ON `groups_login_prefixes` FOR EACH ROW BEGIN INSERT INTO `history_groups_login_prefixes` (`ID`,`iVersion`,`idGroup`,`prefix`) VALUES (NEW.`ID`,@curVersion,NEW.`idGroup`,NEW.`prefix`); END
-- +migrate StatementEnd
DROP TRIGGER `before_update_groups_login_prefixes`;
-- +migrate StatementBegin
CREATE TRIGGER `before_update_groups_login_prefixes` BEFORE UPDATE ON `groups_login_prefixes` FOR EACH ROW BEGIN IF NEW.iVersion <> OLD.iVersion THEN SET @curVersion = NEW.iVersion; ELSE SELECT ROUND(UNIX_TIMESTAMP(CURTIME(2)) * 10) INTO @curVersion; END IF; IF NOT (OLD.`ID` = NEW.`ID` AND OLD.`idGroup` <=> NEW.`idGroup` AND OLD.`prefix` <=> NEW.`prefix`) THEN   SET NEW.iVersion = @curVersion;   UPDATE `history_groups_login_prefixes` SET `iNextVersion` = @curVersion WHERE `ID` = OLD.`ID` AND `iNextVersion` IS NULL;   INSERT INTO `history_groups_login_prefixes` (`ID`,`iVersion`,`idGroup`,`prefix`)       VALUES (NEW.`ID`,@curVersion,NEW.`idGroup`,NEW.`prefix`) ; END IF; END
-- +migrate StatementEnd
DROP TRIGGER `before_delete_groups_login_prefixes`;
-- +migrate StatementBegin
CREATE TRIGGER `before_delete_groups_login_prefixes` BEFORE DELETE ON `groups_login_prefixes` FOR EACH ROW BEGIN SELECT ROUND(UNIX_TIMESTAMP(CURTIME(2)) * 10) INTO @curVersion; UPDATE `history_groups_login_prefixes` SET `iNextVersion` = @curVersion WHERE `ID` = OLD.`ID` AND `iNextVersion` IS NULL; INSERT INTO `history_groups_login_prefixes` (`ID`,`iVersion`,`idGroup`,`prefix`, `bDeleted`) VALUES (OLD.`ID`,@curVersion,OLD.`idGroup`,OLD.`prefix`, 1); END
-- +migrate StatementEnd
DROP TRIGGER `before_insert_items`;
-- +migrate StatementBegin
CREATE TRIGGER `before_insert_items` BEFORE INSERT ON `items` FOR EACH ROW BEGIN IF (NEW.ID IS NULL OR NEW.ID = 0) THEN SET NEW.ID = FLOOR(RAND() * 1000000000) + FLOOR(RAND() * 1000000000) * 1000000000; END IF ; SELECT ROUND(UNIX_TIMESTAMP(CURTIME(2)) * 10) INTO @curVersion;SET NEW.iVersion = @curVersion; SELECT platforms.ID INTO @platformID FROM platforms WHERE NEW.sUrl REGEXP platforms.sRegexp ORDER BY platforms.iPriority DESC LIMIT 1 ; SET NEW.idPlatform=@platformID ; END
-- +migrate StatementEnd
DROP TRIGGER `after_insert_items`;
-- +migrate StatementBegin
CREATE TRIGGER `after_insert_items` AFTER INSERT ON `items` FOR EACH ROW BEGIN INSERT INTO `history_items` (`ID`,`iVersion`,`sUrl`,`idPlatform`,`sTextId`,`sRepositoryPath`,`sType`,`bUsesAPI`,`bReadOnly`,`sFullScreen`,`bShowDifficulty`,`bShowSource`,`bHintsAllowed`,`bFixedRanks`,`sValidationType`,`iValidationMin`,`sPreparationState`,`idItemUnlocked`,`iScoreMinUnlock`,`sSupportedLangProg`,`idDefaultLanguage`,`sTeamMode`,`bTeamsEditable`,`idTeamInGroup`,`iTeamMaxMembers`,`bHasAttempts`,`sAccessOpenDate`,`sDuration`,`sEndContestDate`,`bShowUserInfos`,`sContestPhase`,`iLevel`,`bNoScore`,`bTitleBarVisible`,`bTransparentFolder`,`bDisplayDetailsInParent`,`bDisplayChildrenAsTabs`,`bCustomChapter`,`groupCodeEnter`) VALUES (NEW.`ID`,@curVersion,NEW.`sUrl`,NEW.`idPlatform`,NEW.`sTextId`,NEW.`sRepositoryPath`,NEW.`sType`,NEW.`bUsesAPI`,NEW.`bReadOnly`,NEW.`sFullScreen`,NEW.`bShowDifficulty`,NEW.`bShowSource`,NEW.`bHintsAllowed`,NEW.`bFixedRanks`,NEW.`sValidationType`,NEW.`iValidationMin`,NEW.`sPreparationState`,NEW.`idItemUnlocked`,NEW.`iScoreMinUnlock`,NEW.`sSupportedLangProg`,NEW.`idDefaultLanguage`,NEW.`sTeamMode`,NEW.`bTeamsEditable`,NEW.`idTeamInGroup`,NEW.`iTeamMaxMembers`,NEW.`bHasAttempts`,NEW.`sAccessOpenDate`,NEW.`sDuration`,NEW.`sEndContestDate`,NEW.`bShowUserInfos`,NEW.`sContestPhase`,NEW.`iLevel`,NEW.`bNoScore`,NEW.`bTitleBarVisible`,NEW.`bTransparentFolder`,NEW.`bDisplayDetailsInParent`,NEW.`bDisplayChildrenAsTabs`,NEW.`bCustomChapter`,NEW.`groupCodeEnter`); INSERT IGNORE INTO `items_propagate` (`ID`, `sAncestorsComputationState`) VALUES (NEW.`ID`, 'todo') ; END
-- +migrate StatementEnd
DROP TRIGGER `before_update_items`;
-- +migrate StatementBegin
CREATE TRIGGER `before_update_items` BEFORE UPDATE ON `items` FOR EACH ROW BEGIN IF NEW.iVersion <> OLD.iVersion THEN SET @curVersion = NEW.iVersion; ELSE SELECT ROUND(UNIX_TIMESTAMP(CURTIME(2)) * 10) INTO @curVersion; END IF; IF NOT (OLD.`ID` = NEW.`ID` AND OLD.`sUrl` <=> NEW.`sUrl` AND OLD.`idPlatform` <=> NEW.`idPlatform` AND OLD.`sTextId` <=> NEW.`sTextId` AND OLD.`sRepositoryPath` <=> NEW.`sRepositoryPath` AND OLD.`sType` <=> NEW.`sType` AND OLD.`bUsesAPI` <=> NEW.`bUsesAPI` AND OLD.`bReadOnly` <=> NEW.`bReadOnly` AND OLD.`sFullScreen` <=> NEW.`sFullScreen` AND OLD.`bShowDifficulty` <=> NEW.`bShowDifficulty` AND OLD.`bShowSource` <=> NEW.`bShowSource` AND OLD.`bHintsAllowed` <=> NEW.`bHintsAllowed` AND OLD.`bFixedRanks` <=> NEW.`bFixedRanks` AND OLD.`sValidationType` <=> NEW.`sValidationType` AND OLD.`iValidationMin` <=> NEW.`iValidationMin` AND OLD.`sPreparationState` <=> NEW.`sPreparationState` AND OLD.`idItemUnlocked` <=> NEW.`idItemUnlocked` AND OLD.`iScoreMinUnlock` <=> NEW.`iScoreMinUnlock` AND OLD.`sSupportedLangProg` <=> NEW.`sSupportedLangProg` AND OLD.`idDefaultLanguage` <=> NEW.`idDefaultLanguage` AND OLD.`sTeamMode` <=> NEW.`sTeamMode` AND OLD.`bTeamsEditable` <=> NEW.`bTeamsEditable` AND OLD.`idTeamInGroup` <=> NEW.`idTeamInGroup` AND OLD.`iTeamMaxMembers` <=> NEW.`iTeamMaxMembers` AND OLD.`bHasAttempts` <=> NEW.`bHasAttempts` AND OLD.`sAccessOpenDate` <=> NEW.`sAccessOpenDate` AND OLD.`sDuration` <=> NEW.`sDuration` AND OLD.`sEndContestDate` <=> NEW.`sEndContestDate` AND OLD.`bShowUserInfos` <=> NEW.`bShowUserInfos` AND OLD.`sContestPhase` <=> NEW.`sContestPhase` AND OLD.`iLevel` <=> NEW.`iLevel` AND OLD.`bNoScore` <=> NEW.`bNoScore` AND OLD.`bTitleBarVisible` <=> NEW.`bTitleBarVisible` AND OLD.`bTransparentFolder` <=> NEW.`bTransparentFolder` AND OLD.`bDisplayDetailsInParent` <=> NEW.`bDisplayDetailsInParent` AND OLD.`bDisplayChildrenAsTabs` <=> NEW.`bDisplayChildrenAsTabs` AND OLD.`bCustomChapter` <=> NEW.`bCustomChapter` AND OLD.`groupCodeEnter` <=> NEW.`groupCodeEnter`) THEN   SET NEW.iVersion = @curVersion;   UPDATE `history_items` SET `iNextVersion` = @curVersion WHERE `ID` = OLD.`ID` AND `iNextVersion` IS NULL;   INSERT INTO `history_items` (`ID`,`iVersion`,`sUrl`,`idPlatform`,`sTextId`,`sRepositoryPath`,`sType`,`bUsesAPI`,`bReadOnly`,`sFullScreen`,`bShowDifficulty`,`bShowSource`,`bHintsAllowed`,`bFixedRanks`,`sValidationType`,`iValidationMin`,`sPreparationState`,`idItemUnlocked`,`iScoreMinUnlock`,`sSupportedLangProg`,`idDefaultLanguage`,`sTeamMode`,`bTeamsEditable`,`idTeamInGroup`,`iTeamMaxMembers`,`bHasAttempts`,`sAccessOpenDate`,`sDuration`,`sEndContestDate`,`bShowUserInfos`,`sContestPhase`,`iLevel`,`bNoScore`,`bTitleBarVisible`,`bTransparentFolder`,`bDisplayDetailsInParent`,`bDisplayChildrenAsTabs`,`bCustomChapter`,`groupCodeEnter`)       VALUES (NEW.`ID`,@curVersion,NEW.`sUrl`,NEW.`idPlatform`,NEW.`sTextId`,NEW.`sRepositoryPath`,NEW.`sType`,NEW.`bUsesAPI`,NEW.`bReadOnly`,NEW.`sFullScreen`,NEW.`bShowDifficulty`,NEW.`bShowSource`,NEW.`bHintsAllowed`,NEW.`bFixedRanks`,NEW.`sValidationType`,NEW.`iValidationMin`,NEW.`sPreparationState`,NEW.`idItemUnlocked`,NEW.`iScoreMinUnlock`,NEW.`sSupportedLangProg`,NEW.`idDefaultLanguage`,NEW.`sTeamMode`,NEW.`bTeamsEditable`,NEW.`idTeamInGroup`,NEW.`iTeamMaxMembers`,NEW.`bHasAttempts`,NEW.`sAccessOpenDate`,NEW.`sDuration`,NEW.`sEndContestDate`,NEW.`bShowUserInfos`,NEW.`sContestPhase`,NEW.`iLevel`,NEW.`bNoScore`,NEW.`bTitleBarVisible`,NEW.`bTransparentFolder`,NEW.`bDisplayDetailsInParent`,NEW.`bDisplayChildrenAsTabs`,NEW.`bCustomChapter`,NEW.`groupCodeEnter`) ; END IF; SELECT platforms.ID INTO @platformID FROM platforms WHERE NEW.sUrl REGEXP platforms.sRegexp ORDER BY platforms.iPriority DESC LIMIT 1 ; SET NEW.idPlatform=@platformID ; END
-- +migrate StatementEnd
DROP TRIGGER `before_delete_items`;
-- +migrate StatementBegin
CREATE TRIGGER `before_delete_items` BEFORE DELETE ON `items` FOR EACH ROW BEGIN SELECT ROUND(UNIX_TIMESTAMP(CURTIME(2)) * 10) INTO @curVersion; UPDATE `history_items` SET `iNextVersion` = @curVersion WHERE `ID` = OLD.`ID` AND `iNextVersion` IS NULL; INSERT INTO `history_items` (`ID`,`iVersion`,`sUrl`,`idPlatform`,`sTextId`,`sRepositoryPath`,`sType`,`bUsesAPI`,`bReadOnly`,`sFullScreen`,`bShowDifficulty`,`bShowSource`,`bHintsAllowed`,`bFixedRanks`,`sValidationType`,`iValidationMin`,`sPreparationState`,`idItemUnlocked`,`iScoreMinUnlock`,`sSupportedLangProg`,`idDefaultLanguage`,`sTeamMode`,`bTeamsEditable`,`idTeamInGroup`,`iTeamMaxMembers`,`bHasAttempts`,`sAccessOpenDate`,`sDuration`,`sEndContestDate`,`bShowUserInfos`,`sContestPhase`,`iLevel`,`bNoScore`,`bTitleBarVisible`,`bTransparentFolder`,`bDisplayDetailsInParent`,`bDisplayChildrenAsTabs`,`bCustomChapter`,`groupCodeEnter`, `bDeleted`) VALUES (OLD.`ID`,@curVersion,OLD.`sUrl`,OLD.`idPlatform`,OLD.`sTextId`,OLD.`sRepositoryPath`,OLD.`sType`,OLD.`bUsesAPI`,OLD.`bReadOnly`,OLD.`sFullScreen`,OLD.`bShowDifficulty`,OLD.`bShowSource`,OLD.`bHintsAllowed`,OLD.`bFixedRanks`,OLD.`sValidationType`,OLD.`iValidationMin`,OLD.`sPreparationState`,OLD.`idItemUnlocked`,OLD.`iScoreMinUnlock`,OLD.`sSupportedLangProg`,OLD.`idDefaultLanguage`,OLD.`sTeamMode`,OLD.`bTeamsEditable`,OLD.`idTeamInGroup`,OLD.`iTeamMaxMembers`,OLD.`bHasAttempts`,OLD.`sAccessOpenDate`,OLD.`sDuration`,OLD.`sEndContestDate`,OLD.`bShowUserInfos`,OLD.`sContestPhase`,OLD.`iLevel`,OLD.`bNoScore`,OLD.`bTitleBarVisible`,OLD.`bTransparentFolder`,OLD.`bDisplayDetailsInParent`,OLD.`bDisplayChildrenAsTabs`,OLD.`bCustomChapter`,OLD.`groupCodeEnter`, 1); END
-- +migrate StatementEnd
DROP TRIGGER `after_delete_items`;
-- +migrate StatementBegin
CREATE TRIGGER `after_delete_items` AFTER DELETE ON `items` FOR EACH ROW BEGIN DELETE FROM items_propagate where ID = OLD.ID ; END
-- +migrate StatementEnd
DROP TRIGGER `before_insert_items_ancestors`;
-- +migrate StatementBegin
CREATE TRIGGER `before_insert_items_ancestors` BEFORE INSERT ON `items_ancestors` FOR EACH ROW BEGIN IF (NEW.ID IS NULL OR NEW.ID = 0) THEN SET NEW.ID = FLOOR(RAND() * 1000000000) + FLOOR(RAND() * 1000000000) * 1000000000; END IF ; SELECT ROUND(UNIX_TIMESTAMP(CURTIME(2)) * 10) INTO @curVersion;SET NEW.iVersion = @curVersion; END
-- +migrate StatementEnd
DROP TRIGGER `after_insert_items_ancestors`;
-- +migrate StatementBegin
CREATE TRIGGER `after_insert_items_ancestors` AFTER INSERT ON `items_ancestors` FOR EACH ROW BEGIN INSERT INTO `history_items_ancestors` (`ID`,`iVersion`,`idItemAncestor`,`idItemChild`) VALUES (NEW.`ID`,@curVersion,NEW.`idItemAncestor`,NEW.`idItemChild`); END
-- +migrate StatementEnd
DROP TRIGGER `before_update_items_ancestors`;
-- +migrate StatementBegin
CREATE TRIGGER `before_update_items_ancestors` BEFORE UPDATE ON `items_ancestors` FOR EACH ROW BEGIN IF NEW.iVersion <> OLD.iVersion THEN SET @curVersion = NEW.iVersion; ELSE SELECT ROUND(UNIX_TIMESTAMP(CURTIME(2)) * 10) INTO @curVersion; END IF; IF NOT (OLD.`ID` = NEW.`ID` AND OLD.`idItemAncestor` <=> NEW.`idItemAncestor` AND OLD.`idItemChild` <=> NEW.`idItemChild`) THEN   SET NEW.iVersion = @curVersion;   UPDATE `history_items_ancestors` SET `iNextVersion` = @curVersion WHERE `ID` = OLD.`ID` AND `iNextVersion` IS NULL;   INSERT INTO `history_items_ancestors` (`ID`,`iVersion`,`idItemAncestor`,`idItemChild`)       VALUES (NEW.`ID`,@curVersion,NEW.`idItemAncestor`,NEW.`idItemChild`) ; END IF; END
-- +migrate StatementEnd
DROP TRIGGER `before_delete_items_ancestors`;
-- +migrate StatementBegin
CREATE TRIGGER `before_delete_items_ancestors` BEFORE DELETE ON `items_ancestors` FOR EACH ROW BEGIN SELECT ROUND(UNIX_TIMESTAMP(CURTIME(2)) * 10) INTO @curVersion; UPDATE `history_items_ancestors` SET `iNextVersion` = @curVersion WHERE `ID` = OLD.`ID` AND `iNextVersion` IS NULL; INSERT INTO `history_items_ancestors` (`ID`,`iVersion`,`idItemAncestor`,`idItemChild`, `bDeleted`) VALUES (OLD.`ID`,@curVersion,OLD.`idItemAncestor`,OLD.`idItemChild`, 1); END
-- +migrate StatementEnd
DROP TRIGGER `before_insert_items_items`;
-- +migrate StatementBegin
CREATE TRIGGER `before_insert_items_items` BEFORE INSERT ON `items_items` FOR EACH ROW BEGIN IF (NEW.ID IS NULL OR NEW.ID = 0) THEN SET NEW.ID = FLOOR(RAND() * 1000000000) + FLOOR(RAND() * 1000000000) * 1000000000; END IF ; SELECT ROUND(UNIX_TIMESTAMP(CURTIME(2)) * 10) INTO @curVersion;SET NEW.iVersion = @curVersion; INSERT IGNORE INTO `items_propagate` (ID, sAncestorsComputationState) VALUES (NEW.idItemChild, 'todo') ON DUPLICATE KEY UPDATE `sAncestorsComputationState` = 'todo' ; END
-- +migrate StatementEnd
DROP TRIGGER `after_insert_items_items`;
-- +migrate StatementBegin
CREATE TRIGGER `after_insert_items_items` AFTER INSERT ON `items_items` FOR EACH ROW BEGIN INSERT INTO `history_items_items` (`ID`,`iVersion`,`idItemParent`,`idItemChild`,`iChildOrder`,`sCategory`,`bAccessRestricted`,`bAlwaysVisible`,`iDifficulty`) VALUES (NEW.`ID`,@curVersion,NEW.`idItemParent`,NEW.`idItemChild`,NEW.`iChildOrder`,NEW.`sCategory`,NEW.`bAccessRestricted`,NEW.`bAlwaysVisible`,NEW.`iDifficulty`); INSERT IGNORE INTO `groups_items_propagate` SELECT `ID`, 'children' as `sPropagateAccess` FROM `groups_items` WHERE `groups_items`.`idItem` = NEW.`idItemParent` ON DUPLICATE KEY UPDATE sPropagateAccess='children' ; END
-- +migrate StatementEnd
DROP TRIGGER `before_update_items_items`;
-- +migrate StatementBegin
CREATE TRIGGER `before_update_items_items` BEFORE UPDATE ON `items_items` FOR EACH ROW BEGIN IF NEW.iVersion <> OLD.iVersion THEN SET @curVersion = NEW.iVersion; ELSE SELECT ROUND(UNIX_TIMESTAMP(CURTIME(2)) * 10) INTO @curVersion; END IF; IF NOT (OLD.`ID` = NEW.`ID` AND OLD.`idItemParent` <=> NEW.`idItemParent` AND OLD.`idItemChild` <=> NEW.`idItemChild` AND OLD.`iChildOrder` <=> NEW.`iChildOrder` AND OLD.`sCategory` <=> NEW.`sCategory` AND OLD.`bAccessRestricted` <=> NEW.`bAccessRestricted` AND OLD.`bAlwaysVisible` <=> NEW.`bAlwaysVisible` AND OLD.`iDifficulty` <=> NEW.`iDifficulty`) THEN   SET NEW.iVersion = @curVersion;   UPDATE `history_items_items` SET `iNextVersion` = @curVersion WHERE `ID` = OLD.`ID` AND `iNextVersion` IS NULL;   INSERT INTO `history_items_items` (`ID`,`iVersion`,`idItemParent`,`idItemChild`,`iChildOrder`,`sCategory`,`bAccessRestricted`,`bAlwaysVisible`,`iDifficulty`)       VALUES (NEW.`ID`,@curVersion,NEW.`idItemParent`,NEW.`idItemChild`,NEW.`iChildOrder`,NEW.`sCategory`,NEW.`bAccessRestricted`,NEW.`bAlwaysVisible`,NEW.`iDifficulty`) ; END IF; IF (OLD.idItemChild != NEW.idItemChild OR OLD.idItemParent != NEW.idItemParent) THEN INSERT IGNORE INTO `items_propagate` (`ID`, `sAncestorsComputationState`) VALUES (OLD.idItemChild, 'todo') ON DUPLICATE KEY UPDATE `sAncestorsComputationState` = 'todo'; INSERT IGNORE INTO `items_propagate` (`ID`, `sAncestorsComputationState`) VALUES (OLD.idItemParent, 'todo') ON DUPLICATE KEY UPDATE `sAncestorsComputationState` = 'todo'; INSERT IGNORE INTO `items_propagate` (`ID`, `sAncestorsComputationState`) (SELECT `items_ancestors`.`idItemChild`, 'todo' FROM `items_ancestors` WHERE `items_ancestors`.`idItemAncestor` = OLD.`idItemChild`) ON DUPLICATE KEY UPDATE `sAncestorsComputationState` = 'todo'; DELETE `items_ancestors` from `items_ancestors` WHERE `items_ancestors`.`idItemChild` = OLD.`idItemChild` and `items_ancestors`.`idItemAncestor` = OLD.`idItemParent`;DELETE `bridges` FROM `items_ancestors` `child_descendants` JOIN `items_ancestors` `parent_ancestors` JOIN `items_ancestors` `bridges` ON (`bridges`.`idItemAncestor` = `parent_ancestors`.`idItemAncestor` AND `bridges`.`idItemChild` = `child_descendants`.`idItemChild`) WHERE `parent_ancestors`.`idItemChild` = OLD.`idItemParent` AND `child_descendants`.`idItemAncestor` = OLD.`idItemChild`; DELETE `child_ancestors` FROM `items_ancestors` `child_ancestors` JOIN  `items_ancestors` `parent_ancestors` ON (`child_ancestors`.`idItemChild` = OLD.`idItemChild` AND `child_ancestors`.`idItemAncestor` = `parent_ancestors`.`idItemAncestor`) WHERE `parent_ancestors`.`idItemChild` = OLD.`idItemParent`; DELETE `parent_ancestors` FROM `items_ancestors` `parent_ancestors` JOIN  `items_ancestors` `child_ancestors` ON (`parent_ancestors`.`idItemAncestor` = OLD.`idItemParent` AND `child_ancestors`.`idItemChild` = `parent_ancestors`.`idItemChild`) WHERE `child_ancestors`.`idItemAncestor` = OLD.`idItemChild`  ; END IF; IF (OLD.idItemChild != NEW.idItemChild OR OLD.idItemParent != NEW.idItemParent) THEN INSERT IGNORE INTO `items_propagate` (ID, sAncestorsComputationState) VALUES (NEW.idItemChild, 'todo') ON DUPLICATE KEY UPDATE `sAncestorsComputationState` = 'todo'  ; END IF; END
-- +migrate StatementEnd
DROP TRIGGER `after_update_items_items`;
-- +migrate StatementBegin
CREATE TRIGGER `after_update_items_items` AFTER UPDATE ON `items_items` FOR EACH ROW BEGIN INSERT IGNORE INTO `groups_items_propagate` SELECT `ID`, 'children' as `sPropagateAccess` FROM `groups_items` WHERE `groups_items`.`idItem` = NEW.`idItemParent` OR `groups_items`.`idItem` = OLD.`idItemParent` ON DUPLICATE KEY UPDATE sPropagateAccess='children' ; END
-- +migrate StatementEnd
DROP TRIGGER `before_delete_items_items`;
-- +migrate StatementBegin
CREATE TRIGGER `before_delete_items_items` BEFORE DELETE ON `items_items` FOR EACH ROW BEGIN SELECT ROUND(UNIX_TIMESTAMP(CURTIME(2)) * 10) INTO @curVersion; UPDATE `history_items_items` SET `iNextVersion` = @curVersion WHERE `ID` = OLD.`ID` AND `iNextVersion` IS NULL; INSERT INTO `history_items_items` (`ID`,`iVersion`,`idItemParent`,`idItemChild`,`iChildOrder`,`sCategory`,`bAccessRestricted`,`bAlwaysVisible`,`iDifficulty`, `bDeleted`) VALUES (OLD.`ID`,@curVersion,OLD.`idItemParent`,OLD.`idItemChild`,OLD.`iChildOrder`,OLD.`sCategory`,OLD.`bAccessRestricted`,OLD.`bAlwaysVisible`,OLD.`iDifficulty`, 1); INSERT IGNORE INTO `items_propagate` (`ID`, `sAncestorsComputationState`) VALUES (OLD.idItemChild, 'todo') ON DUPLICATE KEY UPDATE `sAncestorsComputationState` = 'todo'; INSERT IGNORE INTO `items_propagate` (`ID`, `sAncestorsComputationState`) VALUES (OLD.idItemParent, 'todo') ON DUPLICATE KEY UPDATE `sAncestorsComputationState` = 'todo'; INSERT IGNORE INTO `items_propagate` (`ID`, `sAncestorsComputationState`) (SELECT `items_ancestors`.`idItemChild`, 'todo' FROM `items_ancestors` WHERE `items_ancestors`.`idItemAncestor` = OLD.`idItemChild`) ON DUPLICATE KEY UPDATE `sAncestorsComputationState` = 'todo'; DELETE `items_ancestors` from `items_ancestors` WHERE `items_ancestors`.`idItemChild` = OLD.`idItemChild` and `items_ancestors`.`idItemAncestor` = OLD.`idItemParent`;DELETE `bridges` FROM `items_ancestors` `child_descendants` JOIN `items_ancestors` `parent_ancestors` JOIN `items_ancestors` `bridges` ON (`bridges`.`idItemAncestor` = `parent_ancestors`.`idItemAncestor` AND `bridges`.`idItemChild` = `child_descendants`.`idItemChild`) WHERE `parent_ancestors`.`idItemChild` = OLD.`idItemParent` AND `child_descendants`.`idItemAncestor` = OLD.`idItemChild`; DELETE `child_ancestors` FROM `items_ancestors` `child_ancestors` JOIN  `items_ancestors` `parent_ancestors` ON (`child_ancestors`.`idItemChild` = OLD.`idItemChild` AND `child_ancestors`.`idItemAncestor` = `parent_ancestors`.`idItemAncestor`) WHERE `parent_ancestors`.`idItemChild` = OLD.`idItemParent`; DELETE `parent_ancestors` FROM `items_ancestors` `parent_ancestors` JOIN  `items_ancestors` `child_ancestors` ON (`parent_ancestors`.`idItemAncestor` = OLD.`idItemParent` AND `child_ancestors`.`idItemChild` = `parent_ancestors`.`idItemChild`) WHERE `child_ancestors`.`idItemAncestor` = OLD.`idItemChild` ; INSERT IGNORE INTO `groups_items_propagate` SELECT `ID`, 'children' as `sPropagateAccess` FROM `groups_items` WHERE `groups_items`.`idItem` = OLD.`idItemParent` ON DUPLICATE KEY UPDATE sPropagateAccess='children' ; END
-- +migrate StatementEnd
DROP TRIGGER `before_insert_items_strings`;
-- +migrate StatementBegin
CREATE TRIGGER `before_insert_items_strings` BEFORE INSERT ON `items_strings` FOR EACH ROW BEGIN IF (NEW.ID IS NULL OR NEW.ID = 0) THEN SET NEW.ID = FLOOR(RAND() * 1000000000) + FLOOR(RAND() * 1000000000) * 1000000000; END IF ; SELECT ROUND(UNIX_TIMESTAMP(CURTIME(2)) * 10) INTO @curVersion;SET NEW.iVersion = @curVersion; END
-- +migrate StatementEnd
DROP TRIGGER `after_insert_items_strings`;
-- +migrate StatementBegin
CREATE TRIGGER `after_insert_items_strings` AFTER INSERT ON `items_strings` FOR EACH ROW BEGIN INSERT INTO `history_items_strings` (`ID`,`iVersion`,`idItem`,`idLanguage`,`sTranslator`,`sTitle`,`sImageUrl`,`sSubtitle`,`sDescription`,`sEduComment`,`sRankingComment`) VALUES (NEW.`ID`,@curVersion,NEW.`idItem`,NEW.`idLanguage`,NEW.`sTranslator`,NEW.`sTitle`,NEW.`sImageUrl`,NEW.`sSubtitle`,NEW.`sDescription`,NEW.`sEduComment`,NEW.`sRankingComment`); END
-- +migrate StatementEnd
DROP TRIGGER `before_update_items_strings`;
-- +migrate StatementBegin
CREATE TRIGGER `before_update_items_strings` BEFORE UPDATE ON `items_strings` FOR EACH ROW BEGIN IF NEW.iVersion <> OLD.iVersion THEN SET @curVersion = NEW.iVersion; ELSE SELECT ROUND(UNIX_TIMESTAMP(CURTIME(2)) * 10) INTO @curVersion; END IF; IF NOT (OLD.`ID` = NEW.`ID` AND OLD.`idItem` <=> NEW.`idItem` AND OLD.`idLanguage` <=> NEW.`idLanguage` AND OLD.`sTranslator` <=> NEW.`sTranslator` AND OLD.`sTitle` <=> NEW.`sTitle` AND OLD.`sImageUrl` <=> NEW.`sImageUrl` AND OLD.`sSubtitle` <=> NEW.`sSubtitle` AND OLD.`sDescription` <=> NEW.`sDescription` AND OLD.`sEduComment` <=> NEW.`sEduComment` AND OLD.`sRankingComment` <=> NEW.`sRankingComment`) THEN   SET NEW.iVersion = @curVersion;   UPDATE `history_items_strings` SET `iNextVersion` = @curVersion WHERE `ID` = OLD.`ID` AND `iNextVersion` IS NULL;   INSERT INTO `history_items_strings` (`ID`,`iVersion`,`idItem`,`idLanguage`,`sTranslator`,`sTitle`,`sImageUrl`,`sSubtitle`,`sDescription`,`sEduComment`,`sRankingComment`)       VALUES (NEW.`ID`,@curVersion,NEW.`idItem`,NEW.`idLanguage`,NEW.`sTranslator`,NEW.`sTitle`,NEW.`sImageUrl`,NEW.`sSubtitle`,NEW.`sDescription`,NEW.`sEduComment`,NEW.`sRankingComment`) ; END IF; END
-- +migrate StatementEnd
DROP TRIGGER `before_delete_items_strings`;
-- +migrate StatementBegin
CREATE TRIGGER `before_delete_items_strings` BEFORE DELETE ON `items_strings` FOR EACH ROW BEGIN SELECT ROUND(UNIX_TIMESTAMP(CURTIME(2)) * 10) INTO @curVersion; UPDATE `history_items_strings` SET `iNextVersion` = @curVersion WHERE `ID` = OLD.`ID` AND `iNextVersion` IS NULL; INSERT INTO `history_items_strings` (`ID`,`iVersion`,`idItem`,`idLanguage`,`sTranslator`,`sTitle`,`sImageUrl`,`sSubtitle`,`sDescription`,`sEduComment`,`sRankingComment`, `bDeleted`) VALUES (OLD.`ID`,@curVersion,OLD.`idItem`,OLD.`idLanguage`,OLD.`sTranslator`,OLD.`sTitle`,OLD.`sImageUrl`,OLD.`sSubtitle`,OLD.`sDescription`,OLD.`sEduComment`,OLD.`sRankingComment`, 1); END
-- +migrate StatementEnd
DROP TRIGGER `before_insert_languages`;
-- +migrate StatementBegin
CREATE TRIGGER `before_insert_languages` BEFORE INSERT ON `languages` FOR EACH ROW BEGIN IF (NEW.ID IS NULL OR NEW.ID = 0) THEN SET NEW.ID = FLOOR(RAND() * 1000000000) + FLOOR(RAND() * 1000000000) * 1000000000; END IF ; SELECT ROUND(UNIX_TIMESTAMP(CURTIME(2)) * 10) INTO @curVersion;SET NEW.iVersion = @curVersion; END
-- +migrate StatementEnd
DROP TRIGGER `after_insert_languages`;
-- +migrate StatementBegin
CREATE TRIGGER `after_insert_languages` AFTER INSERT ON `languages` FOR EACH ROW BEGIN INSERT INTO `history_languages` (`ID`,`iVersion`,`sName`,`sCode`) VALUES (NEW.`ID`,@curVersion,NEW.`sName`,NEW.`sCode`); END
-- +migrate StatementEnd
DROP TRIGGER `before_update_languages`;
-- +migrate StatementBegin
CREATE TRIGGER `before_update_languages` BEFORE UPDATE ON `languages` FOR EACH ROW BEGIN IF NEW.iVersion <> OLD.iVersion THEN SET @curVersion = NEW.iVersion; ELSE SELECT ROUND(UNIX_TIMESTAMP(CURTIME(2)) * 10) INTO @curVersion; END IF; IF NOT (OLD.`ID` = NEW.`ID` AND OLD.`sName` <=> NEW.`sName` AND OLD.`sCode` <=> NEW.`sCode`) THEN   SET NEW.iVersion = @curVersion;   UPDATE `history_languages` SET `iNextVersion` = @curVersion WHERE `ID` = OLD.`ID` AND `iNextVersion` IS NULL;   INSERT INTO `history_languages` (`ID`,`iVersion`,`sName`,`sCode`)       VALUES (NEW.`ID`,@curVersion,NEW.`sName`,NEW.`sCode`) ; END IF; END
-- +migrate StatementEnd
DROP TRIGGER `before_delete_languages`;
-- +migrate StatementBegin
CREATE TRIGGER `before_delete_languages` BEFORE DELETE ON `languages` FOR EACH ROW BEGIN SELECT ROUND(UNIX_TIMESTAMP(CURTIME(2)) * 10) INTO @curVersion; UPDATE `history_languages` SET `iNextVersion` = @curVersion WHERE `ID` = OLD.`ID` AND `iNextVersion` IS NULL; INSERT INTO `history_languages` (`ID`,`iVersion`,`sName`,`sCode`, `bDeleted`) VALUES (OLD.`ID`,@curVersion,OLD.`sName`,OLD.`sCode`, 1); END
-- +migrate StatementEnd
DROP TRIGGER `before_insert_messages`;
-- +migrate StatementBegin
CREATE TRIGGER `before_insert_messages` BEFORE INSERT ON `messages` FOR EACH ROW BEGIN IF (NEW.ID IS NULL OR NEW.ID = 0) THEN SET NEW.ID = FLOOR(RAND() * 1000000000) + FLOOR(RAND() * 1000000000) * 1000000000; END IF ; SELECT ROUND(UNIX_TIMESTAMP(CURTIME(2)) * 10) INTO @curVersion;SET NEW.iVersion = @curVersion; END
-- +migrate StatementEnd
DROP TRIGGER `after_insert_messages`;
-- +migrate StatementBegin
CREATE TRIGGER `after_insert_messages` AFTER INSERT ON `messages` FOR EACH ROW BEGIN INSERT INTO `history_messages` (`ID`,`iVersion`,`idThread`,`idUser`,`sSubmissionDate`,`bPublished`,`sTitle`,`sBody`,`bTrainersOnly`,`bArchived`,`bPersistant`) VALUES (NEW.`ID`,@curVersion,NEW.`idThread`,NEW.`idUser`,NEW.`sSubmissionDate`,NEW.`bPublished`,NEW.`sTitle`,NEW.`sBody`,NEW.`bTrainersOnly`,NEW.`bArchived`,NEW.`bPersistant`); END
-- +migrate StatementEnd
DROP TRIGGER `before_update_messages`;
-- +migrate StatementBegin
CREATE TRIGGER `before_update_messages` BEFORE UPDATE ON `messages` FOR EACH ROW BEGIN IF NEW.iVersion <> OLD.iVersion THEN SET @curVersion = NEW.iVersion; ELSE SELECT ROUND(UNIX_TIMESTAMP(CURTIME(2)) * 10) INTO @curVersion; END IF; IF NOT (OLD.`ID` = NEW.`ID` AND OLD.`idThread` <=> NEW.`idThread` AND OLD.`idUser` <=> NEW.`idUser` AND OLD.`sSubmissionDate` <=> NEW.`sSubmissionDate` AND OLD.`bPublished` <=> NEW.`bPublished` AND OLD.`sTitle` <=> NEW.`sTitle` AND OLD.`sBody` <=> NEW.`sBody` AND OLD.`bTrainersOnly` <=> NEW.`bTrainersOnly` AND OLD.`bArchived` <=> NEW.`bArchived` AND OLD.`bPersistant` <=> NEW.`bPersistant`) THEN   SET NEW.iVersion = @curVersion;   UPDATE `history_messages` SET `iNextVersion` = @curVersion WHERE `ID` = OLD.`ID` AND `iNextVersion` IS NULL;   INSERT INTO `history_messages` (`ID`,`iVersion`,`idThread`,`idUser`,`sSubmissionDate`,`bPublished`,`sTitle`,`sBody`,`bTrainersOnly`,`bArchived`,`bPersistant`)       VALUES (NEW.`ID`,@curVersion,NEW.`idThread`,NEW.`idUser`,NEW.`sSubmissionDate`,NEW.`bPublished`,NEW.`sTitle`,NEW.`sBody`,NEW.`bTrainersOnly`,NEW.`bArchived`,NEW.`bPersistant`) ; END IF; END
-- +migrate StatementEnd
DROP TRIGGER `before_delete_messages`;
-- +migrate StatementBegin
CREATE TRIGGER `before_delete_messages` BEFORE DELETE ON `messages` FOR EACH ROW BEGIN SELECT ROUND(UNIX_TIMESTAMP(CURTIME(2)) * 10) INTO @curVersion; UPDATE `history_messages` SET `iNextVersion` = @curVersion WHERE `ID` = OLD.`ID` AND `iNextVersion` IS NULL; INSERT INTO `history_messages` (`ID`,`iVersion`,`idThread`,`idUser`,`sSubmissionDate`,`bPublished`,`sTitle`,`sBody`,`bTrainersOnly`,`bArchived`,`bPersistant`, `bDeleted`) VALUES (OLD.`ID`,@curVersion,OLD.`idThread`,OLD.`idUser`,OLD.`sSubmissionDate`,OLD.`bPublished`,OLD.`sTitle`,OLD.`sBody`,OLD.`bTrainersOnly`,OLD.`bArchived`,OLD.`bPersistant`, 1); END
-- +migrate StatementEnd
DROP TRIGGER `before_insert_threads`;
-- +migrate StatementBegin
CREATE TRIGGER `before_insert_threads` BEFORE INSERT ON `threads` FOR EACH ROW BEGIN IF (NEW.ID IS NULL OR NEW.ID = 0) THEN SET NEW.ID = FLOOR(RAND() * 1000000000) + FLOOR(RAND() * 1000000000) * 1000000000; END IF ; SELECT ROUND(UNIX_TIMESTAMP(CURTIME(2)) * 10) INTO @curVersion;SET NEW.iVersion = @curVersion; END
-- +migrate StatementEnd
DROP TRIGGER `after_insert_threads`;
-- +migrate StatementBegin
CREATE TRIGGER `after_insert_threads` AFTER INSERT ON `threads` FOR EACH ROW BEGIN INSERT INTO `history_threads` (`ID`,`iVersion`,`sType`,`idUserCreated`,`idItem`,`sTitle`,`bAdminHelpAsked`,`bHidden`,`sLastActivityDate`) VALUES (NEW.`ID`,@curVersion,NEW.`sType`,NEW.`idUserCreated`,NEW.`idItem`,NEW.`sTitle`,NEW.`bAdminHelpAsked`,NEW.`bHidden`,NEW.`sLastActivityDate`); END
-- +migrate StatementEnd
DROP TRIGGER `before_update_threads`;
-- +migrate StatementBegin
CREATE TRIGGER `before_update_threads` BEFORE UPDATE ON `threads` FOR EACH ROW BEGIN IF NEW.iVersion <> OLD.iVersion THEN SET @curVersion = NEW.iVersion; ELSE SELECT ROUND(UNIX_TIMESTAMP(CURTIME(2)) * 10) INTO @curVersion; END IF; IF NOT (OLD.`ID` = NEW.`ID` AND OLD.`sType` <=> NEW.`sType` AND OLD.`idUserCreated` <=> NEW.`idUserCreated` AND OLD.`idItem` <=> NEW.`idItem` AND OLD.`sTitle` <=> NEW.`sTitle` AND OLD.`bAdminHelpAsked` <=> NEW.`bAdminHelpAsked` AND OLD.`bHidden` <=> NEW.`bHidden` AND OLD.`sLastActivityDate` <=> NEW.`sLastActivityDate`) THEN   SET NEW.iVersion = @curVersion;   UPDATE `history_threads` SET `iNextVersion` = @curVersion WHERE `ID` = OLD.`ID` AND `iNextVersion` IS NULL;   INSERT INTO `history_threads` (`ID`,`iVersion`,`sType`,`idUserCreated`,`idItem`,`sTitle`,`bAdminHelpAsked`,`bHidden`,`sLastActivityDate`)       VALUES (NEW.`ID`,@curVersion,NEW.`sType`,NEW.`idUserCreated`,NEW.`idItem`,NEW.`sTitle`,NEW.`bAdminHelpAsked`,NEW.`bHidden`,NEW.`sLastActivityDate`) ; END IF; END
-- +migrate StatementEnd
DROP TRIGGER `before_delete_threads`;
-- +migrate StatementBegin
CREATE TRIGGER `before_delete_threads` BEFORE DELETE ON `threads` FOR EACH ROW BEGIN SELECT ROUND(UNIX_TIMESTAMP(CURTIME(2)) * 10) INTO @curVersion; UPDATE `history_threads` SET `iNextVersion` = @curVersion WHERE `ID` = OLD.`ID` AND `iNextVersion` IS NULL; INSERT INTO `history_threads` (`ID`,`iVersion`,`sType`,`idUserCreated`,`idItem`,`sTitle`,`bAdminHelpAsked`,`bHidden`,`sLastActivityDate`, `bDeleted`) VALUES (OLD.`ID`,@curVersion,OLD.`sType`,OLD.`idUserCreated`,OLD.`idItem`,OLD.`sTitle`,OLD.`bAdminHelpAsked`,OLD.`bHidden`,OLD.`sLastActivityDate`, 1); END
-- +migrate StatementEnd
DROP TRIGGER `before_insert_users`;
-- +migrate StatementBegin
CREATE TRIGGER `before_insert_users` BEFORE INSERT ON `users` FOR EACH ROW BEGIN IF (NEW.ID IS NULL OR NEW.ID = 0) THEN SET NEW.ID = FLOOR(RAND() * 1000000000) + FLOOR(RAND() * 1000000000) * 1000000000; END IF ; SELECT ROUND(UNIX_TIMESTAMP(CURTIME(2)) * 10) INTO @curVersion;SET NEW.iVersion = @curVersion; END
-- +migrate StatementEnd
DROP TRIGGER `after_insert_users`;
-- +migrate StatementBegin
CREATE TRIGGER `after_insert_users` AFTER INSERT ON `users` FOR EACH ROW BEGIN INSERT INTO `history_users` (`ID`,`iVersion`,`sLogin`,`sOpenIdIdentity`,`sPasswordMd5`,`sSalt`,`sRecover`,`sRegistrationDate`,`sEmail`,`bEmailVerified`,`sFirstName`,`sLastName`,`sCountryCode`,`sTimeZone`,`sBirthDate`,`iGraduationYear`,`iGrade`,`sSex`,`sStudentId`,`sAddress`,`sZipcode`,`sCity`,`sLandLineNumber`,`sCellPhoneNumber`,`sDefaultLanguage`,`bNotifyNews`,`sNotify`,`bPublicFirstName`,`bPublicLastName`,`sFreeText`,`sWebSite`,`bPhotoAutoload`,`sLangProg`,`sLastLoginDate`,`sLastActivityDate`,`sLastIP`,`bBasicEditorMode`,`nbSpacesForTab`,`iMemberState`,`idUserGodfather`,`iStepLevelInSite`,`bIsAdmin`,`bNoRanking`,`nbHelpGiven`,`idGroupSelf`,`idGroupOwned`,`idGroupAccess`,`sNotificationReadDate`,`loginModulePrefix`,`allowSubgroups`) VALUES (NEW.`ID`,@curVersion,NEW.`sLogin`,NEW.`sOpenIdIdentity`,NEW.`sPasswordMd5`,NEW.`sSalt`,NEW.`sRecover`,NEW.`sRegistrationDate`,NEW.`sEmail`,NEW.`bEmailVerified`,NEW.`sFirstName`,NEW.`sLastName`,NEW.`sCountryCode`,NEW.`sTimeZone`,NEW.`sBirthDate`,NEW.`iGraduationYear`,NEW.`iGrade`,NEW.`sSex`,NEW.`sStudentId`,NEW.`sAddress`,NEW.`sZipcode`,NEW.`sCity`,NEW.`sLandLineNumber`,NEW.`sCellPhoneNumber`,NEW.`sDefaultLanguage`,NEW.`bNotifyNews`,NEW.`sNotify`,NEW.`bPublicFirstName`,NEW.`bPublicLastName`,NEW.`sFreeText`,NEW.`sWebSite`,NEW.`bPhotoAutoload`,NEW.`sLangProg`,NEW.`sLastLoginDate`,NEW.`sLastActivityDate`,NEW.`sLastIP`,NEW.`bBasicEditorMode`,NEW.`nbSpacesForTab`,NEW.`iMemberState`,NEW.`idUserGodfather`,NEW.`iStepLevelInSite`,NEW.`bIsAdmin`,NEW.`bNoRanking`,NEW.`nbHelpGiven`,NEW.`idGroupSelf`,NEW.`idGroupOwned`,NEW.`idGroupAccess`,NEW.`sNotificationReadDate`,NEW.`loginModulePrefix`,NEW.`allowSubgroups`); END
-- +migrate StatementEnd
DROP TRIGGER `before_update_users`;
-- +migrate StatementBegin
CREATE TRIGGER `before_update_users` BEFORE UPDATE ON `users` FOR EACH ROW BEGIN IF NEW.iVersion <> OLD.iVersion THEN SET @curVersion = NEW.iVersion; ELSE SELECT ROUND(UNIX_TIMESTAMP(CURTIME(2)) * 10) INTO @curVersion; END IF; IF NOT (OLD.`ID` = NEW.`ID` AND OLD.`sLogin` <=> NEW.`sLogin` AND OLD.`sOpenIdIdentity` <=> NEW.`sOpenIdIdentity` AND OLD.`sPasswordMd5` <=> NEW.`sPasswordMd5` AND OLD.`sSalt` <=> NEW.`sSalt` AND OLD.`sRecover` <=> NEW.`sRecover` AND OLD.`sRegistrationDate` <=> NEW.`sRegistrationDate` AND OLD.`sEmail` <=> NEW.`sEmail` AND OLD.`bEmailVerified` <=> NEW.`bEmailVerified` AND OLD.`sFirstName` <=> NEW.`sFirstName` AND OLD.`sLastName` <=> NEW.`sLastName` AND OLD.`sCountryCode` <=> NEW.`sCountryCode` AND OLD.`sTimeZone` <=> NEW.`sTimeZone` AND OLD.`sBirthDate` <=> NEW.`sBirthDate` AND OLD.`iGraduationYear` <=> NEW.`iGraduationYear` AND OLD.`iGrade` <=> NEW.`iGrade` AND OLD.`sSex` <=> NEW.`sSex` AND OLD.`sStudentId` <=> NEW.`sStudentId` AND OLD.`sAddress` <=> NEW.`sAddress` AND OLD.`sZipcode` <=> NEW.`sZipcode` AND OLD.`sCity` <=> NEW.`sCity` AND OLD.`sLandLineNumber` <=> NEW.`sLandLineNumber` AND OLD.`sCellPhoneNumber` <=> NEW.`sCellPhoneNumber` AND OLD.`sDefaultLanguage` <=> NEW.`sDefaultLanguage` AND OLD.`bNotifyNews` <=> NEW.`bNotifyNews` AND OLD.`sNotify` <=> NEW.`sNotify` AND OLD.`bPublicFirstName` <=> NEW.`bPublicFirstName` AND OLD.`bPublicLastName` <=> NEW.`bPublicLastName` AND OLD.`sFreeText` <=> NEW.`sFreeText` AND OLD.`sWebSite` <=> NEW.`sWebSite` AND OLD.`bPhotoAutoload` <=> NEW.`bPhotoAutoload` AND OLD.`sLangProg` <=> NEW.`sLangProg` AND OLD.`sLastLoginDate` <=> NEW.`sLastLoginDate` AND OLD.`sLastActivityDate` <=> NEW.`sLastActivityDate` AND OLD.`sLastIP` <=> NEW.`sLastIP` AND OLD.`bBasicEditorMode` <=> NEW.`bBasicEditorMode` AND OLD.`nbSpacesForTab` <=> NEW.`nbSpacesForTab` AND OLD.`iMemberState` <=> NEW.`iMemberState` AND OLD.`idUserGodfather` <=> NEW.`idUserGodfather` AND OLD.`iStepLevelInSite` <=> NEW.`iStepLevelInSite` AND OLD.`bIsAdmin` <=> NEW.`bIsAdmin` AND OLD.`bNoRanking` <=> NEW.`bNoRanking` AND OLD.`nbHelpGiven` <=> NEW.`nbHelpGiven` AND OLD.`idGroupSelf` <=> NEW.`idGroupSelf` AND OLD.`idGroupOwned` <=> NEW.`idGroupOwned` AND OLD.`idGroupAccess` <=> NEW.`idGroupAccess` AND OLD.`sNotificationReadDate` <=> NEW.`sNotificationReadDate` AND OLD.`loginModulePrefix` <=> NEW.`loginModulePrefix` AND OLD.`allowSubgroups` <=> NEW.`allowSubgroups`) THEN   SET NEW.iVersion = @curVersion;   UPDATE `history_users` SET `iNextVersion` = @curVersion WHERE `ID` = OLD.`ID` AND `iNextVersion` IS NULL;   INSERT INTO `history_users` (`ID`,`iVersion`,`sLogin`,`sOpenIdIdentity`,`sPasswordMd5`,`sSalt`,`sRecover`,`sRegistrationDate`,`sEmail`,`bEmailVerified`,`sFirstName`,`sLastName`,`sCountryCode`,`sTimeZone`,`sBirthDate`,`iGraduationYear`,`iGrade`,`sSex`,`sStudentId`,`sAddress`,`sZipcode`,`sCity`,`sLandLineNumber`,`sCellPhoneNumber`,`sDefaultLanguage`,`bNotifyNews`,`sNotify`,`bPublicFirstName`,`bPublicLastName`,`sFreeText`,`sWebSite`,`bPhotoAutoload`,`sLangProg`,`sLastLoginDate`,`sLastActivityDate`,`sLastIP`,`bBasicEditorMode`,`nbSpacesForTab`,`iMemberState`,`idUserGodfather`,`iStepLevelInSite`,`bIsAdmin`,`bNoRanking`,`nbHelpGiven`,`idGroupSelf`,`idGroupOwned`,`idGroupAccess`,`sNotificationReadDate`,`loginModulePrefix`,`allowSubgroups`)    VALUES (NEW.`ID`,@curVersion,NEW.`sLogin`,NEW.`sOpenIdIdentity`,NEW.`sPasswordMd5`,NEW.`sSalt`,NEW.`sRecover`,NEW.`sRegistrationDate`,NEW.`sEmail`,NEW.`bEmailVerified`,NEW.`sFirstName`,NEW.`sLastName`,NEW.`sCountryCode`,NEW.`sTimeZone`,NEW.`sBirthDate`,NEW.`iGraduationYear`,NEW.`iGrade`,NEW.`sSex`,NEW.`sStudentId`,NEW.`sAddress`,NEW.`sZipcode`,NEW.`sCity`,NEW.`sLandLineNumber`,NEW.`sCellPhoneNumber`,NEW.`sDefaultLanguage`,NEW.`bNotifyNews`,NEW.`sNotify`,NEW.`bPublicFirstName`,NEW.`bPublicLastName`,NEW.`sFreeText`,NEW.`sWebSite`,NEW.`bPhotoAutoload`,NEW.`sLangProg`,NEW.`sLastLoginDate`,NEW.`sLastActivityDate`,NEW.`sLastIP`,NEW.`bBasicEditorMode`,NEW.`nbSpacesForTab`,NEW.`iMemberState`,NEW.`idUserGodfather`,NEW.`iStepLevelInSite`,NEW.`bIsAdmin`,NEW.`bNoRanking`,NEW.`nbHelpGiven`,NEW.`idGroupSelf`,NEW.`idGroupOwned`,NEW.`idGroupAccess`,NEW.`sNotificationReadDate`,NEW.`loginModulePrefix`,NEW.`allowSubgroups`) ; END IF; END
-- +migrate StatementEnd
DROP TRIGGER `before_delete_users`;
-- +migrate StatementBegin
CREATE TRIGGER `before_delete_users` BEFORE DELETE ON `users` FOR EACH ROW BEGIN SELECT ROUND(UNIX_TIMESTAMP(CURTIME(2)) * 10) INTO @curVersion; UPDATE `history_users` SET `iNextVersion` = @curVersion WHERE `ID` = OLD.`ID` AND `iNextVersion` IS NULL; INSERT INTO `history_users` (`ID`,`iVersion`,`sLogin`,`sOpenIdIdentity`,`sPasswordMd5`,`sSalt`,`sRecover`,`sRegistrationDate`,`sEmail`,`bEmailVerified`,`sFirstName`,`sLastName`,`sCountryCode`,`sTimeZone`,`sBirthDate`,`iGraduationYear`,`iGrade`,`sSex`,`sStudentId`,`sAddress`,`sZipcode`,`sCity`,`sLandLineNumber`,`sCellPhoneNumber`,`sDefaultLanguage`,`bNotifyNews`,`sNotify`,`bPublicFirstName`,`bPublicLastName`,`sFreeText`,`sWebSite`,`bPhotoAutoload`,`sLangProg`,`sLastLoginDate`,`sLastActivityDate`,`sLastIP`,`bBasicEditorMode`,`nbSpacesForTab`,`iMemberState`,`idUserGodfather`,`iStepLevelInSite`,`bIsAdmin`,`bNoRanking`,`nbHelpGiven`,`idGroupSelf`,`idGroupOwned`,`idGroupAccess`,`sNotificationReadDate`,`loginModulePrefix`,`allowSubgroups`, `bDeleted`) VALUES (OLD.`ID`,@curVersion,OLD.`sLogin`,OLD.`sOpenIdIdentity`,OLD.`sPasswordMd5`,OLD.`sSalt`,OLD.`sRecover`,OLD.`sRegistrationDate`,OLD.`sEmail`,OLD.`bEmailVerified`,OLD.`sFirstName`,OLD.`sLastName`,OLD.`sCountryCode`,OLD.`sTimeZone`,OLD.`sBirthDate`,OLD.`iGraduationYear`,OLD.`iGrade`,OLD.`sSex`,OLD.`sStudentId`,OLD.`sAddress`,OLD.`sZipcode`,OLD.`sCity`,OLD.`sLandLineNumber`,OLD.`sCellPhoneNumber`,OLD.`sDefaultLanguage`,OLD.`bNotifyNews`,OLD.`sNotify`,OLD.`bPublicFirstName`,OLD.`bPublicLastName`,OLD.`sFreeText`,OLD.`sWebSite`,OLD.`bPhotoAutoload`,OLD.`sLangProg`,OLD.`sLastLoginDate`,OLD.`sLastActivityDate`,OLD.`sLastIP`,OLD.`bBasicEditorMode`,OLD.`nbSpacesForTab`,OLD.`iMemberState`,OLD.`idUserGodfather`,OLD.`iStepLevelInSite`,OLD.`bIsAdmin`,OLD.`bNoRanking`,OLD.`nbHelpGiven`,OLD.`idGroupSelf`,OLD.`idGroupOwned`,OLD.`idGroupAccess`,OLD.`sNotificationReadDate`,OLD.`loginModulePrefix`,OLD.`allowSubgroups`, 1); END
-- +migrate StatementEnd
DROP TRIGGER `before_insert_users_items`;
-- +migrate StatementBegin
CREATE TRIGGER `before_insert_users_items` BEFORE INSERT ON `users_items` FOR EACH ROW BEGIN IF (NEW.ID IS NULL OR NEW.ID = 0) THEN SET NEW.ID = FLOOR(RAND() * 1000000000) + FLOOR(RAND() * 1000000000) * 1000000000; END IF ; SELECT ROUND(UNIX_TIMESTAMP(CURTIME(2)) * 10) INTO @curVersion;SET NEW.iVersion = @curVersion; END
-- +migrate StatementEnd
DROP TRIGGER `after_insert_users_items`;
-- +migrate StatementBegin
CREATE TRIGGER `after_insert_users_items` AFTER INSERT ON `users_items` FOR EACH ROW BEGIN INSERT INTO `history_users_items` (`ID`,`iVersion`,`idUser`,`idItem`,`idAttemptActive`,`iScore`,`iScoreComputed`,`iScoreReeval`,`iScoreDiffManual`,`sScoreDiffComment`,`nbSubmissionsAttempts`,`nbTasksTried`,`nbChildrenValidated`,`bValidated`,`bFinished`,`bKeyObtained`,`nbTasksWithHelp`,`sHintsRequested`,`nbHintsCached`,`nbCorrectionsRead`,`iPrecision`,`iAutonomy`,`sStartDate`,`sValidationDate`,`sBestAnswerDate`,`sLastAnswerDate`,`sThreadStartDate`,`sLastHintDate`,`sFinishDate`,`sLastActivityDate`,`sContestStartDate`,`bRanked`,`sAllLangProg`,`sState`,`sAnswer`) VALUES (NEW.`ID`,@curVersion,NEW.`idUser`,NEW.`idItem`,NEW.`idAttemptActive`,NEW.`iScore`,NEW.`iScoreComputed`,NEW.`iScoreReeval`,NEW.`iScoreDiffManual`,NEW.`sScoreDiffComment`,NEW.`nbSubmissionsAttempts`,NEW.`nbTasksTried`,NEW.`nbChildrenValidated`,NEW.`bValidated`,NEW.`bFinished`,NEW.`bKeyObtained`,NEW.`nbTasksWithHelp`,NEW.`sHintsRequested`,NEW.`nbHintsCached`,NEW.`nbCorrectionsRead`,NEW.`iPrecision`,NEW.`iAutonomy`,NEW.`sStartDate`,NEW.`sValidationDate`,NEW.`sBestAnswerDate`,NEW.`sLastAnswerDate`,NEW.`sThreadStartDate`,NEW.`sLastHintDate`,NEW.`sFinishDate`,NEW.`sLastActivityDate`,NEW.`sContestStartDate`,NEW.`bRanked`,NEW.`sAllLangProg`,NEW.`sState`,NEW.`sAnswer`); END
-- +migrate StatementEnd
DROP TRIGGER `before_update_users_items`;
-- +migrate StatementBegin
CREATE TRIGGER `before_update_users_items` BEFORE UPDATE ON `users_items` FOR EACH ROW BEGIN IF NEW.iVersion <> OLD.iVersion THEN SET @curVersion = NEW.iVersion; ELSE SELECT ROUND(UNIX_TIMESTAMP(CURTIME(2)) * 10) INTO @curVersion; END IF; IF NOT (OLD.`ID` = NEW.`ID` AND OLD.`idUser` <=> NEW.`idUser` AND OLD.`idItem` <=> NEW.`idItem` AND OLD.`idAttemptActive` <=> NEW.`idAttemptActive` AND OLD.`iScore` <=> NEW.`iScore` AND OLD.`iScoreComputed` <=> NEW.`iScoreComputed` AND OLD.`iScoreReeval` <=> NEW.`iScoreReeval` AND OLD.`iScoreDiffManual` <=> NEW.`iScoreDiffManual` AND OLD.`sScoreDiffComment` <=> NEW.`sScoreDiffComment` AND OLD.`nbTasksTried` <=> NEW.`nbTasksTried` AND OLD.`nbChildrenValidated` <=> NEW.`nbChildrenValidated` AND OLD.`bValidated` <=> NEW.`bValidated` AND OLD.`bFinished` <=> NEW.`bFinished` AND OLD.`bKeyObtained` <=> NEW.`bKeyObtained` AND OLD.`nbTasksWithHelp` <=> NEW.`nbTasksWithHelp` AND OLD.`sHintsRequested` <=> NEW.`sHintsRequested` AND OLD.`nbHintsCached` <=> NEW.`nbHintsCached` AND OLD.`nbCorrectionsRead` <=> NEW.`nbCorrectionsRead` AND OLD.`iPrecision` <=> NEW.`iPrecision` AND OLD.`iAutonomy` <=> NEW.`iAutonomy` AND OLD.`sStartDate` <=> NEW.`sStartDate` AND OLD.`sValidationDate` <=> NEW.`sValidationDate` AND OLD.`sBestAnswerDate` <=> NEW.`sBestAnswerDate` AND OLD.`sLastAnswerDate` <=> NEW.`sLastAnswerDate` AND OLD.`sThreadStartDate` <=> NEW.`sThreadStartDate` AND OLD.`sLastHintDate` <=> NEW.`sLastHintDate` AND OLD.`sFinishDate` <=> NEW.`sFinishDate` AND OLD.`sContestStartDate` <=> NEW.`sContestStartDate` AND OLD.`bRanked` <=> NEW.`bRanked` AND OLD.`sAllLangProg` <=> NEW.`sAllLangProg` AND OLD.`sState` <=> NEW.`sState` AND OLD.`sAnswer` <=> NEW.`sAnswer`) THEN   SET NEW.iVersion = @curVersion;   UPDATE `history_users_items` SET `iNextVersion` = @curVersion WHERE `ID` = OLD.`ID` AND `iNextVersion` IS NULL;   INSERT INTO `history_users_items` (`ID`,`iVersion`,`idUser`,`idItem`,`idAttemptActive`,`iScore`,`iScoreComputed`,`iScoreReeval`,`iScoreDiffManual`,`sScoreDiffComment`,`nbSubmissionsAttempts`,`nbTasksTried`,`nbChildrenValidated`,`bValidated`,`bFinished`,`bKeyObtained`,`nbTasksWithHelp`,`sHintsRequested`,`nbHintsCached`,`nbCorrectionsRead`,`iPrecision`,`iAutonomy`,`sStartDate`,`sValidationDate`,`sBestAnswerDate`,`sLastAnswerDate`,`sThreadStartDate`,`sLastHintDate`,`sFinishDate`,`sLastActivityDate`,`sContestStartDate`,`bRanked`,`sAllLangProg`,`sState`,`sAnswer`)       VALUES (NEW.`ID`,@curVersion,NEW.`idUser`,NEW.`idItem`,NEW.`idAttemptActive`,NEW.`iScore`,NEW.`iScoreComputed`,NEW.`iScoreReeval`,NEW.`iScoreDiffManual`,NEW.`sScoreDiffComment`,NEW.`nbSubmissionsAttempts`,NEW.`nbTasksTried`,NEW.`nbChildrenValidated`,NEW.`bValidated`,NEW.`bFinished`,NEW.`bKeyObtained`,NEW.`nbTasksWithHelp`,NEW.`sHintsRequested`,NEW.`nbHintsCached`,NEW.`nbCorrectionsRead`,NEW.`iPrecision`,NEW.`iAutonomy`,NEW.`sStartDate`,NEW.`sValidationDate`,NEW.`sBestAnswerDate`,NEW.`sLastAnswerDate`,NEW.`sThreadStartDate`,NEW.`sLastHintDate`,NEW.`sFinishDate`,NEW.`sLastActivityDate`,NEW.`sContestStartDate`,NEW.`bRanked`,NEW.`sAllLangProg`,NEW.`sState`,NEW.`sAnswer`) ; END IF; END
-- +migrate StatementEnd
DROP TRIGGER `before_delete_users_items`;
-- +migrate StatementBegin
CREATE TRIGGER `before_delete_users_items` BEFORE DELETE ON `users_items` FOR EACH ROW BEGIN SELECT ROUND(UNIX_TIMESTAMP(CURTIME(2)) * 10) INTO @curVersion; UPDATE `history_users_items` SET `iNextVersion` = @curVersion WHERE `ID` = OLD.`ID` AND `iNextVersion` IS NULL; INSERT INTO `history_users_items` (`ID`,`iVersion`,`idUser`,`idItem`,`idAttemptActive`,`iScore`,`iScoreComputed`,`iScoreReeval`,`iScoreDiffManual`,`sScoreDiffComment`,`nbSubmissionsAttempts`,`nbTasksTried`,`nbChildrenValidated`,`bValidated`,`bFinished`,`bKeyObtained`,`nbTasksWithHelp`,`sHintsRequested`,`nbHintsCached`,`nbCorrectionsRead`,`iPrecision`,`iAutonomy`,`sStartDate`,`sValidationDate`,`sBestAnswerDate`,`sLastAnswerDate`,`sThreadStartDate`,`sLastHintDate`,`sFinishDate`,`sLastActivityDate`,`sContestStartDate`,`bRanked`,`sAllLangProg`,`sState`,`sAnswer`, `bDeleted`) VALUES (OLD.`ID`,@curVersion,OLD.`idUser`,OLD.`idItem`,OLD.`idAttemptActive`,OLD.`iScore`,OLD.`iScoreComputed`,OLD.`iScoreReeval`,OLD.`iScoreDiffManual`,OLD.`sScoreDiffComment`,OLD.`nbSubmissionsAttempts`,OLD.`nbTasksTried`,OLD.`nbChildrenValidated`,OLD.`bValidated`,OLD.`bFinished`,OLD.`bKeyObtained`,OLD.`nbTasksWithHelp`,OLD.`sHintsRequested`,OLD.`nbHintsCached`,OLD.`nbCorrectionsRead`,OLD.`iPrecision`,OLD.`iAutonomy`,OLD.`sStartDate`,OLD.`sValidationDate`,OLD.`sBestAnswerDate`,OLD.`sLastAnswerDate`,OLD.`sThreadStartDate`,OLD.`sLastHintDate`,OLD.`sFinishDate`,OLD.`sLastActivityDate`,OLD.`sContestStartDate`,OLD.`bRanked`,OLD.`sAllLangProg`,OLD.`sState`,OLD.`sAnswer`, 1); END
-- +migrate StatementEnd
DROP TRIGGER `before_insert_users_threads`;
-- +migrate StatementBegin
CREATE TRIGGER `before_insert_users_threads` BEFORE INSERT ON `users_threads` FOR EACH ROW BEGIN IF (NEW.ID IS NULL OR NEW.ID = 0) THEN SET NEW.ID = FLOOR(RAND() * 1000000000) + FLOOR(RAND() * 1000000000) * 1000000000; END IF ; SELECT ROUND(UNIX_TIMESTAMP(CURTIME(2)) * 10) INTO @curVersion;SET NEW.iVersion = @curVersion; END
-- +migrate StatementEnd
DROP TRIGGER `after_insert_users_threads`;
-- +migrate StatementBegin
CREATE TRIGGER `after_insert_users_threads` AFTER INSERT ON `users_threads` FOR EACH ROW BEGIN INSERT INTO `history_users_threads` (`ID`,`iVersion`,`idUser`,`idThread`,`sLastReadDate`,`sLastWriteDate`,`bStarred`) VALUES (NEW.`ID`,@curVersion,NEW.`idUser`,NEW.`idThread`,NEW.`sLastReadDate`,NEW.`sLastWriteDate`,NEW.`bStarred`); END
-- +migrate StatementEnd
DROP TRIGGER `before_update_users_threads`;
-- +migrate StatementBegin
CREATE TRIGGER `before_update_users_threads` BEFORE UPDATE ON `users_threads` FOR EACH ROW BEGIN IF NEW.iVersion <> OLD.iVersion THEN SET @curVersion = NEW.iVersion; ELSE SELECT ROUND(UNIX_TIMESTAMP(CURTIME(2)) * 10) INTO @curVersion; END IF; IF NOT (OLD.`ID` = NEW.`ID` AND OLD.`idUser` <=> NEW.`idUser` AND OLD.`idThread` <=> NEW.`idThread` AND OLD.`sLastReadDate` <=> NEW.`sLastReadDate` AND OLD.`sLastWriteDate` <=> NEW.`sLastWriteDate` AND OLD.`bStarred` <=> NEW.`bStarred`) THEN   SET NEW.iVersion = @curVersion;   UPDATE `history_users_threads` SET `iNextVersion` = @curVersion WHERE `ID` = OLD.`ID` AND `iNextVersion` IS NULL;   INSERT INTO `history_users_threads` (`ID`,`iVersion`,`idUser`,`idThread`,`sLastReadDate`,`sLastWriteDate`,`bStarred`)       VALUES (NEW.`ID`,@curVersion,NEW.`idUser`,NEW.`idThread`,NEW.`sLastReadDate`,NEW.`sLastWriteDate`,NEW.`bStarred`) ; END IF; END
-- +migrate StatementEnd
DROP TRIGGER `before_delete_users_threads`;
-- +migrate StatementBegin
CREATE TRIGGER `before_delete_users_threads` BEFORE DELETE ON `users_threads` FOR EACH ROW BEGIN SELECT ROUND(UNIX_TIMESTAMP(CURTIME(2)) * 10) INTO @curVersion; UPDATE `history_users_threads` SET `iNextVersion` = @curVersion WHERE `ID` = OLD.`ID` AND `iNextVersion` IS NULL; INSERT INTO `history_users_threads` (`ID`,`iVersion`,`idUser`,`idThread`,`sLastReadDate`,`sLastWriteDate`,`bStarred`, `bDeleted`) VALUES (OLD.`ID`,@curVersion,OLD.`idUser`,OLD.`idThread`,OLD.`sLastReadDate`,OLD.`sLastWriteDate`,OLD.`bStarred`, 1); END
-- +migrate StatementEnd

ALTER ALGORITHM=UNDEFINED
    SQL SECURITY DEFINER
    VIEW `task_children_data_view` AS
SELECT
    `parent_users_items`.`ID` AS `idUserItem`,
    SUM(IF(`task_children`.`ID` IS NOT NULL AND `task_children`.`bValidated`, 1, 0)) AS `nbChildrenValidated`,
    SUM(IF(`task_children`.`ID` IS NOT NULL AND `task_children`.`bValidated`, 0, 1)) AS `nbChildrenNonValidated`,
    SUM(IF(`items_items`.`sCategory` = 'Validation' AND
           (ISNULL(`task_children`.`ID`) OR `task_children`.`bValidated` != 1), 1, 0)) AS `nbChildrenCategory`,
    MAX(`task_children`.`sValidationDate`) AS `maxValidationDate`,
    MAX(IF(`items_items`.`sCategory` = 'Validation', `task_children`.`sValidationDate`, NULL)) AS `maxValidationDateCategories`
FROM `users_items` AS `parent_users_items`
         JOIN `items_items` ON(
        `parent_users_items`.`idItem` = `items_items`.`idItemParent`
    )
         LEFT JOIN `users_items` AS `task_children` ON(
            `items_items`.`idItemChild` = `task_children`.`idItem` AND
            `task_children`.`idUser` = `parent_users_items`.`idUser`
    )
         JOIN `items` ON(
        `items`.`ID` = `items_items`.`idItemChild`
    )
WHERE `items`.`sType` <> 'Course' AND `items`.`bNoScore` = 0
GROUP BY `idUserItem`;
