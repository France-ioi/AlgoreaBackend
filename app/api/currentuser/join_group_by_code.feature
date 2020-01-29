Feature: Join a group using a code (groupsJoinByCode)
  Background:
    Given the database has the following table 'groups':
      | id | type     | code       | code_expires_at     | code_lifetime | free_access | require_watch_approval |
      | 11 | Team     | 3456789abc | 2037-05-29 06:38:38 | 01:02:03      | true        | 0                      |
      | 12 | Team     | abc3456789 | null                | 12:34:56      | true        | 0                      |
      | 14 | Team     | cba9876543 | null                | null          | true        | 0                      |
      | 15 | Team     | 987654321a | null                | null          | true        | 1                      |
      | 21 | UserSelf | null       | null                | null          | false       | 0                      |
    And the database has the following table 'users':
      | group_id |
      | 21       |
    And the database has the following table 'groups_ancestors':
      | ancestor_group_id | child_group_id | is_self |
      | 11                | 11             | 1       |
      | 12                | 12             | 1       |
      | 14                | 14             | 1       |
      | 15                | 15             | 1       |
      | 21                | 21             | 1       |
    And the database has the following table 'group_pending_requests':
      | group_id | member_id | type         |
      | 11       | 21        | invitation   |
      | 14       | 21        | join_request |
    And the database has the following table 'items':
      | id | default_language_tag |
      | 20 | fr                   |
      | 30 | fr                   |
    And the database has the following table 'items_ancestors':
      | ancestor_item_id | child_item_id |
      | 20               | 30            |
    And the database has the following table 'items_items':
      | parent_item_id | child_item_id | child_order |
      | 20             | 30            | 1           |
    And the database has the following table 'permissions_generated':
      | group_id | item_id | can_view_generated |
      | 11       | 20      | content            |
      | 12       | 20      | content            |
      | 14       | 20      | solution           |
      | 15       | 20      | info               |
      | 21       | 30      | content            |
    And the database has the following table 'attempts':
      | group_id | item_id | order | result_propagation_state |
      | 21       | 30      | 1     | done                     |

  Scenario: Successfully join a group
    Given I am the user with id "21"
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
      | parent_group_id | child_group_id |
      | 11              | 21             |
    And the table "group_pending_requests" should be:
      | group_id | member_id | type         |
      | 14       | 21        | join_request |
    And the table "group_membership_changes" should be:
      | group_id | member_id | action         | initiator_id | ABS(TIMESTAMPDIFF(SECOND, at, NOW())) < 3 |
      | 11       | 21        | joined_by_code | 21           | 1                                         |
    And the table "groups_ancestors" should be:
      | ancestor_group_id | child_group_id | is_self |
      | 11                | 11             | 1       |
      | 11                | 21             | 0       |
      | 12                | 12             | 1       |
      | 14                | 14             | 1       |
      | 15                | 15             | 1       |
      | 21                | 21             | 1       |
    And the table "attempts" should be:
      | group_id | item_id | result_propagation_state |
      | 21       | 20      | done                     |
      | 21       | 30      | done                     |

  Scenario: Updates the code_expires_at
    Given I am the user with id "21"
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
      | parent_group_id | child_group_id |
      | 12              | 21             |
    And the table "group_pending_requests" should be:
      | group_id | member_id | type         |
      | 11       | 21        | invitation   |
      | 14       | 21        | join_request |
    And the table "group_membership_changes" should be:
      | group_id | member_id | action         | initiator_id | ABS(TIMESTAMPDIFF(SECOND, at, NOW())) < 3 |
      | 12       | 21        | joined_by_code | 21           | 1                                         |
    And the table "groups_ancestors" should be:
      | ancestor_group_id | child_group_id | is_self |
      | 11                | 11             | 1       |
      | 12                | 12             | 1       |
      | 12                | 21             | 0       |
      | 14                | 14             | 1       |
      | 15                | 15             | 1       |
      | 21                | 21             | 1       |
    And the table "attempts" should be:
      | group_id | item_id | result_propagation_state |
      | 21       | 20      | done                     |
      | 21       | 30      | done                     |

  Scenario: Doesn't update the code_expires_at if code_lifetime is null
    Given I am the user with id "21"
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
      | parent_group_id | child_group_id |
      | 14              | 21             |
    And the table "group_pending_requests" should be:
      | group_id | member_id | type       |
      | 11       | 21        | invitation |
    And the table "group_membership_changes" should be:
      | group_id | member_id | action         | initiator_id | ABS(TIMESTAMPDIFF(SECOND, at, NOW())) < 3 |
      | 14       | 21        | joined_by_code | 21           | 1                                         |
    And the table "groups_ancestors" should be:
      | ancestor_group_id | child_group_id | is_self |
      | 11                | 11             | 1       |
      | 12                | 12             | 1       |
      | 14                | 14             | 1       |
      | 14                | 21             | 0       |
      | 15                | 15             | 1       |
      | 21                | 21             | 1       |
    And the table "attempts" should be:
      | group_id | item_id | result_propagation_state |
      | 21       | 20      | done                     |
      | 21       | 30      | done                     |

  Scenario: Successfully join a group that requires approvals
    Given I am the user with id "21"
    When I send a POST request to "/current-user/group-memberships/by-code?code=987654321a&approvals=watch"
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
      | parent_group_id | child_group_id |
      | 15              | 21             |
    And the table "group_pending_requests" should be:
      | group_id | member_id | type         |
      | 11       | 21        | invitation   |
      | 14       | 21        | join_request |
    And the table "group_membership_changes" should be:
      | group_id | member_id | action         | initiator_id | ABS(TIMESTAMPDIFF(SECOND, at, NOW())) < 3 |
      | 15       | 21        | joined_by_code | 21           | 1                                         |
    And the table "groups_ancestors" should be:
      | ancestor_group_id | child_group_id | is_self |
      | 11                | 11             | 1       |
      | 12                | 12             | 1       |
      | 14                | 14             | 1       |
      | 15                | 15             | 1       |
      | 15                | 21             | 0       |
      | 21                | 21             | 1       |
    And the table "attempts" should be:
      | group_id | item_id | result_propagation_state |
      | 21       | 20      | done                     |
      | 21       | 30      | done                     |
