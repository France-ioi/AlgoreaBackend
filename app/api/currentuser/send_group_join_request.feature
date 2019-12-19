Feature: User sends a request to join a group
  Background:
    Given the database has the following table 'groups':
      | id | free_access | require_personal_info_access_approval | require_lock_membership_approval_until | require_watch_approval |
      | 11 | 1           | edit                                  | 9999-12-31 23:59:59                    | 1                      |
      | 14 | 1           | none                                  | null                                   | 0                      |
      | 21 | 0           | none                                  | null                                   | 0                      |
    And the database has the following table 'users':
      | group_id |
      | 21       |
    And the database has the following table 'groups_ancestors':
      | ancestor_group_id | child_group_id | is_self |
      | 11                | 11             | 1       |
      | 14                | 14             | 1       |
      | 14                | 21             | 0       |
      | 21                | 21             | 1       |
    And the database has the following table 'group_pending_requests':
      | group_id | member_id | type         | at                  |
      | 14       | 21        | join_request | 2019-05-30 11:00:00 |

  Scenario: Successfully send a request
    Given I am the user with id "21"
    When I send a POST request to "/current-user/group-requests/11?approvals=personal_info_view,lock_membership,watch"
    Then the response code should be 201
    And the response body should be, in JSON:
    """
    {
      "success": true,
      "message": "created",
      "data": {"changed": true}
    }
    """
    And the table "group_pending_requests" should be:
      | group_id | member_id | type         | personal_info_view_approved | lock_membership_approved | watch_approved | ABS(TIMESTAMPDIFF(SECOND, at, NOW())) < 3 |
      | 11       | 21        | join_request | 1                           | 1                        | 1              | 1                                         |
      | 14       | 21        | join_request | 0                           | 0                        | 0              | 0                                         |
    And the table "group_membership_changes" should be:
      | group_id | member_id | action               | initiator_id | ABS(TIMESTAMPDIFF(SECOND, at, NOW())) < 3 |
      | 11       | 21        | join_request_created | 21           | 1                                         |
    And the table "groups_ancestors" should stay unchanged

  Scenario: Try to recreate a request that already exists
    Given I am the user with id "21"
    When I send a POST request to "/current-user/group-requests/14"
    Then the response code should be 201
    And the response body should be, in JSON:
    """
    {
      "success": true,
      "message": "unchanged",
      "data": {"changed": false}
    }
    """
    And the table "group_pending_requests" should stay unchanged
    And the table "groups_ancestors" should stay unchanged

  Scenario: Automatically accepts the request if the user owns the group
    Given I am the user with id "21"
    And the database has the following table 'group_managers':
      | group_id | manager_id |
      | 11       | 21         |
    When I send a POST request to "/current-user/group-requests/11"
    Then the response code should be 201
    And the response body should be, in JSON:
    """
    {
      "success": true,
      "message": "created",
      "data": {"changed": true}
    }
    """
    And the table "groups_groups" should be:
      | parent_group_id | child_group_id |
      | 11              | 21             |
    And the table "group_pending_requests" should be:
      | group_id | member_id | type         | ABS(TIMESTAMPDIFF(SECOND, at, NOW())) < 3 |
      | 14       | 21        | join_request | 0                                         |
    And the table "group_membership_changes" should be:
      | group_id | member_id | action                | initiator_id | ABS(TIMESTAMPDIFF(SECOND, at, NOW())) < 3 |
      | 11       | 21        | join_request_accepted | 21           | 1                                         |
    And the table "groups_ancestors" should be:
      | ancestor_group_id | child_group_id | is_self |
      | 11                | 11             | 1       |
      | 11                | 21             | 0       |
      | 14                | 14             | 1       |
      | 14                | 21             | 0       |
      | 21                | 21             | 1       |
