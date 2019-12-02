Feature: User leaves a group
  Background:
    Given the database has the following table 'groups':
      | id | lock_user_deletion_until |
      | 11 | 2019-08-20               |
      | 14 | null                     |
      | 21 | null                     |
    And the database has the following table 'users':
      | group_id |
      | 21       |
    And the database has the following table 'groups_ancestors':
      | id | ancestor_group_id | child_group_id | is_self |
      | 1  | 11                | 11             | 1       |
      | 2  | 11                | 21             | 0       |
      | 3  | 14                | 14             | 1       |
      | 4  | 21                | 21             | 1       |
    And the database has the following table 'groups_groups':
      | id | parent_group_id | child_group_id |
      | 1  | 11              | 21             |

  Scenario: Successfully leave a group
    Given I am the user with id "21"
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
    And the table "groups_groups" should be empty
    And the table "group_membership_changes" should be:
      | group_id | member_id | action | initiator_id | ABS(TIMESTAMPDIFF(SECOND, at, NOW())) < 3 |
      | 11       | 21        | left   | 21           | 1                                         |
    And the table "groups_ancestors" should stay unchanged but the row with id "2"
    And the table "groups_ancestors" should not contain id "2"

  Scenario: Successfully leave a group (lock_user_deletion_until = NOW())
    Given I am the user with id "21"
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
    And the table "groups_groups" should be empty
    And the table "group_membership_changes" should be:
      | group_id | member_id | action | initiator_id | ABS(TIMESTAMPDIFF(SECOND, "2019-08-20 00:00:00", NOW())) < 3 |
      | 11       | 21        | left   | 21           | 1                                                            |
    And the table "groups_ancestors" should stay unchanged but the row with id "2"
    And the table "groups_ancestors" should not contain id "2"

