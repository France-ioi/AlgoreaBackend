Feature: User leaves a group
  Background:
    Given the database has the following table 'users':
      | id | group_self_id | group_owned_id |
      | 1  | 21            | 22             |
    And the database has the following table 'groups':
      | id | lock_user_deletion_date |
      | 11 | 2019-08-20              |
      | 14 | null                    |
      | 21 | null                    |
      | 22 | null                    |
    And the database has the following table 'groups_ancestors':
      | id | group_ancestor_id | group_child_id | is_self |
      | 1  | 11                | 11             | 1       |
      | 2  | 11                | 21             | 0       |
      | 3  | 14                | 14             | 1       |
      | 4  | 21                | 21             | 1       |
      | 5  | 22                | 22             | 1       |
    And the database has the following table 'groups_groups':
      | id | group_parent_id | group_child_id | type               | status_date         |
      | 1  | 11              | 21             | invitationAccepted | 2017-04-29 06:38:38 |
      | 7  | 14              | 21             | left               | 2017-02-21 06:38:38 |

  Scenario: Successfully leave a group
    Given I am the user with id "1"
    When I send a DELETE request to "/current-user/group-memberships/11"
    Then the response code should be 200
    And the response body should be, in JSON:
    """
    {
      "success": true,
      "message": "deleted",
      "data": {"changed": true}
    }
    """
    And the table "groups_groups" should stay unchanged but the row with id "1"
    And the table "groups_groups" at id "1" should be:
      | id | group_parent_id | group_child_id | type | (status_date IS NOT NULL) AND (ABS(TIMESTAMPDIFF(SECOND, status_date, NOW())) < 3) |
      | 1  | 11              | 21             | left | 1                                                                                  |
    And the table "groups_ancestors" should stay unchanged but the row with id "2"
    And the table "groups_ancestors" should not contain id "2"

  Scenario: Leave a group that already have been left
    Given I am the user with id "1"
    When I send a DELETE request to "/current-user/group-memberships/14"
    Then the response code should be 200
    And the response body should be, in JSON:
    """
    {
      "success": true,
      "message": "unchanged",
      "data": {"changed": false}
    }
    """
    And the table "groups_groups" should stay unchanged
    And the table "groups_ancestors" should stay unchanged

  Scenario: Successfully leave a group (lock_user_deletion_date = NOW())
    Given I am the user with id "1"
    And the DB time now is "2019-08-20 00:00:00"
    When I send a DELETE request to "/current-user/group-memberships/11"
    Then the response code should be 200
    And the response body should be, in JSON:
    """
    {
      "success": true,
      "message": "deleted",
      "data": {"changed": true}
    }
    """
    And the table "groups_groups" should stay unchanged but the row with id "1"
    And the table "groups_groups" at id "1" should be:
      | id | group_parent_id | group_child_id | type | (status_date IS NOT NULL) AND (ABS(TIMESTAMPDIFF(SECOND, status_date, NOW())) < 3) |
      | 1  | 11              | 21             | left | 1                                                                                  |
    And the table "groups_ancestors" should stay unchanged but the row with id "2"
    And the table "groups_ancestors" should not contain id "2"

