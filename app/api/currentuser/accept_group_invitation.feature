Feature: User accepts an invitation to join a group
  Background:
    Given the database has the following table 'groups':
      | id |
      | 11 |
      | 14 |
      | 21 |
      | 22 |
    Given the database has the following table 'users':
      | group_id | owned_group_id |
      | 21       | 22             |
    And the database has the following table 'groups_ancestors':
      | ancestor_group_id | child_group_id | is_self |
      | 11                | 11             | 1       |
      | 14                | 14             | 1       |
      | 14                | 21             | 0       |
      | 21                | 21             | 1       |
      | 22                | 22             | 1       |
    And the database has the following table 'groups_groups':
      | id | parent_group_id | child_group_id |
      | 7  | 14              | 21             |
    And the database has the following table 'group_pending_requests':
      | group_id | member_id | type       |
      | 11       | 21        | invitation |

  Scenario: Successfully accept an invitation
    Given I am the user with id "21"
    When I send a POST request to "/current-user/group-invitations/11/accept"
    Then the response code should be 200
    And the response body should be, in JSON:
    """
    {
      "success": true,
      "message": "updated",
      "data": {"changed": true}
    }
    """
    And the table "groups_groups" should stay unchanged but the row with parent_group_id "11"
    And the table "groups_groups" at parent_group_id "11" should be:
      | parent_group_id | child_group_id |
      | 11              | 21             |
    And the table "groups_ancestors" should be:
      | ancestor_group_id | child_group_id | is_self |
      | 11                | 11             | 1       |
      | 11                | 21             | 0       |
      | 14                | 14             | 1       |
      | 14                | 21             | 0       |
      | 21                | 21             | 1       |
      | 22                | 22             | 1       |
