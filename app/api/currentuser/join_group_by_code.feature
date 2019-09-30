Feature: Join a group using a code (groupsJoinByCode)
  Background:
    Given the database has the following table 'users':
      | id | self_group_id | owned_group_id |
      | 1  | 21            | 22             |
    And the database has the following table 'groups':
      | id | type      | code       | code_expires_at     | code_lifetime | free_access |
      | 11 | Team      | 3456789abc | 2037-05-29 06:38:38 | 01:02:03      | true        |
      | 12 | Team      | abc3456789 | null                | 12:34:56      | true        |
      | 14 | Team      | cba9876543 | null                | null          | true        |
      | 21 | UserSelf  | null       | null                | null          | false       |
      | 22 | UserAdmin | null       | null                | null          | false       |
    And the database has the following table 'groups_ancestors':
      | ancestor_group_id | child_group_id | is_self |
      | 11                | 11             | 1       |
      | 12                | 12             | 1       |
      | 14                | 14             | 1       |
      | 21                | 21             | 1       |
      | 22                | 22             | 1       |
    And the database has the following table 'groups_groups':
      | id | parent_group_id | child_group_id | type           | type_changed_at     |
      | 1  | 11              | 21             | invitationSent | 2017-04-29 06:38:38 |
      | 7  | 14              | 21             | requestSent    | 2017-02-21 06:38:38 |

  Scenario: Successfully join an group
    Given I am the user with id "1"
    When I send a POST request to "/current-user/group-memberships/by-code?code=3456789abc"
    Then the response code should be 201
    And the response body should be, in JSON:
    """
    {
      "success": true,
      "message": "created",
      "data": {"changed": true}
    }
    """
    And the table "groups" should stay unchanged
    And the table "groups_groups" should be:
      | parent_group_id | child_group_id | type         | (type_changed_at IS NOT NULL) AND (ABS(TIMESTAMPDIFF(SECOND, type_changed_at, NOW())) < 3) |
      | 11              | 21             | joinedByCode | 1                                                                                          |
      | 14              | 21             | requestSent  | 0                                                                                          |
    And the table "groups_ancestors" should be:
      | ancestor_group_id | child_group_id | is_self |
      | 11                | 11             | 1       |
      | 11                | 21             | 0       |
      | 12                | 12             | 1       |
      | 14                | 14             | 1       |
      | 21                | 21             | 1       |
      | 22                | 22             | 1       |

  Scenario: Updates the code_expires_at
    Given I am the user with id "1"
    When I send a POST request to "/current-user/group-memberships/by-code?code=abc3456789"
    Then the response code should be 201
    And the response body should be, in JSON:
    """
    {
      "success": true,
      "message": "created",
      "data": {"changed": true}
    }
    """
    And the table "groups" should stay unchanged but the row with id "12"
    And the table "groups" at id "12" should be:
      | id | type | code       | code_lifetime | free_access | TIMESTAMPDIFF(SECOND, code_expires_at, ADDTIME(NOW(), "12:34:56")) < 3 |
      | 12 | Team | abc3456789 | 12:34:56      | true        | 1                                                                      |
    And the table "groups_groups" should be:
      | parent_group_id | child_group_id | type           | (type_changed_at IS NOT NULL) AND (ABS(TIMESTAMPDIFF(SECOND, type_changed_at, NOW())) < 3) |
      | 11              | 21             | invitationSent | 0                                                                                          |
      | 12              | 21             | joinedByCode   | 1                                                                                          |
      | 14              | 21             | requestSent    | 0                                                                                          |
    And the table "groups_ancestors" should be:
      | ancestor_group_id | child_group_id | is_self |
      | 11                | 11             | 1       |
      | 12                | 12             | 1       |
      | 12                | 21             | 0       |
      | 14                | 14             | 1       |
      | 21                | 21             | 1       |
      | 22                | 22             | 1       |

  Scenario: Doesn't update the code_expires_at if code_lifetime is null
    Given I am the user with id "1"
    When I send a POST request to "/current-user/group-memberships/by-code?code=cba9876543"
    Then the response code should be 201
    And the response body should be, in JSON:
    """
    {
      "success": true,
      "message": "created",
      "data": {"changed": true}
    }
    """
    And the table "groups" should stay unchanged
    And the table "groups_groups" should be:
      | parent_group_id | child_group_id | type           | (type_changed_at IS NOT NULL) AND (ABS(TIMESTAMPDIFF(SECOND, type_changed_at, NOW())) < 3) |
      | 11              | 21             | invitationSent | 0                                                                                          |
      | 14              | 21             | joinedByCode   | 1                                                                                          |
    And the table "groups_ancestors" should be:
      | ancestor_group_id | child_group_id | is_self |
      | 11                | 11             | 1       |
      | 12                | 12             | 1       |
      | 14                | 14             | 1       |
      | 14                | 21             | 0       |
      | 21                | 21             | 1       |
      | 22                | 22             | 1       |
