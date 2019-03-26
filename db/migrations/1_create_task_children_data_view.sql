-- +migrate Up
DROP VIEW IF EXISTS `task_children_data_view`;
SET @saved_cs_client          = @@character_set_client;
SET @saved_cs_results         = @@character_set_results;
SET @saved_col_connection     = @@collation_connection;
SET character_set_client      = utf8mb4;
SET character_set_results     = utf8mb4;
SET collation_connection      = utf8mb4_general_ci;
CREATE ALGORITHM=UNDEFINED
  DEFINER=`algorea`@`%` SQL SECURITY DEFINER
  VIEW `task_children_data_view` AS
  SELECT
    `parent_users_items`.`ID` AS `idUserItem`,
    SUM(IF(`task_children`.`ID` IS NOT NULL AND `task_children`.`bValidated`, 1, 0)) AS `nbChildrenValidated`,
    SUM(IF(`task_children`.`ID` IS NOT NULL AND `task_children`.`bValidated`, 0, 1)) AS `nbChildrenNonValidated`,
    SUM(IF(`items_items`.`sCategory` = 'Validation' AND
      (ISNULL(`task_children`.`ID`) OR `task_children`.`bValidated` != 1), 1, 0)) AS `nbChildrenCategory`,
    MAX(`task_children`.`sValidationDate`) AS `maxValidationDate`,
    MAX(IF(`items_items`.`sCategory` = 'Validation', `task_children`.`sValidationDate`, NULL)) AS `maxValidationDateCategories`
  FROM `users_items` `parent_users_items`
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
SET character_set_client      = @saved_cs_client;
SET character_set_results     = @saved_cs_results;
SET collation_connection      = @saved_col_connection;

-- +migrate Down
DROP VIEW IF EXISTS `task_children_data_view`;
