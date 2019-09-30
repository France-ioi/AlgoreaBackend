Feature: User rejects an invitation to join a group
  Background:
    Given the database has the following table 'users':
      | id | self_group_id | owned_group_id |
      | 1  | 21            | 22             |
    And the database has the following table 'groups':
      | id |
      | 11 |
      | 14 |
      | 21 |
      | 22 |
    And the database has the following table 'groups_ancestors':
      | ancestor_group_id | child_group_id | is_self |
      | 11                | 11             | 1       |
      | 14                | 14             | 1       |
      | 21                | 21             | 1       |
      | 22                | 22             | 1       |
    And the database has the following table 'groups_groups':
      | id | parent_group_id | child_group_id | type              | type_changed_at     |
      | 1  | 11              | 21             | invitationSent    | 2017-04-29 06:38:38 |
      | 7  | 14              | 21             | invitationRefused | 2017-02-21 06:38:38 |

  Scenario: Successfully reject an invitation
    Given I am the user with id "1"
    When I send a POST request to "/current-user/group-invitations/11/reject"
    Then the response code should be 200
    And the response body should be, in JSON:
    """
    {
      "success": true,
      "message": "updated",
      "data": {"changed": true}
    }
    """
    And the table "groups_groups" should stay unchanged but the row with id "1"
    And the table "groups_groups" at id "1" should be:
      | id | parent_group_id | child_group_id | type              | (type_changed_at IS NOT NULL) AND (ABS(TIMESTAMPDIFF(SECOND, type_changed_at, NOW())) < 3) |
      | 1  | 11              | 21             | invitationRefused | 1                                                                                          |
    And the table "groups_ancestors" should stay unchanged

  Scenario: Reject an already rejected invitation
    Given I am the user with id "1"
    When I send a POST request to "/current-user/group-invitations/14/reject"
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

