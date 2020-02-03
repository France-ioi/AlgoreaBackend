Feature: User leaves a group - robustness
  Background:
    Given the database has the following table 'groups':
      | id | require_lock_membership_approval_until |
      | 11 | null                                   |
      | 14 | null                                   |
      | 15 | 2037-04-29                             |
      | 21 | null                                   |
      | 31 | null                                   |
    And the database has the following table 'users':
      | group_id | login |
      | 21       | john  |
      | 31       | jane  |
    And the database has the following table 'groups_ancestors':
      | ancestor_group_id | child_group_id |
      | 11                | 11             |
      | 14                | 14             |
      | 14                | 21             |
      | 15                | 15             |
      | 15                | 31             |
      | 21                | 21             |
      | 31                | 31             |
    And the database has the following table 'groups_groups':
      | id | parent_group_id | child_group_id | lock_membership_approved_at |
      | 2  | 14              | 21             | null                        |
      | 3  | 15              | 31             | 2019-05-30 11:00:00         |
    And the database has the following table 'group_pending_requests':
      | group_id | member_id | type         |
      | 11       | 21        | join_request |

  Scenario: User tries to leave a group (s)he is not a member of
    Given I am the user with id "21"
    When I send a DELETE request to "/current-user/group-memberships/11"
    Then the response code should be 404
    And the response body should be, in JSON:
    """
    {
      "success": false,
      "message": "Not Found",
      "error_text": "No such relation"
    }
    """
    And the table "groups_groups" should stay unchanged
    And the table "groups_ancestors" should stay unchanged

  Scenario: Fails when the group id is wrong
    Given I am the user with id "21"
    When I send a DELETE request to "/current-user/group-memberships/abc"
    Then the response code should be 400
    And the response error message should contain "Wrong value for group_id (should be int64)"
    And the table "groups_groups" should stay unchanged
    And the table "groups_ancestors" should stay unchanged

  Scenario: Fails if the user doesn't exist
    Given I am the user with id "404"
    When I send a DELETE request to "/current-user/group-memberships/14"
    Then the response code should be 401
    And the response error message should contain "Invalid access token"

  Scenario: Fails if require_lock_membership_approval_until = NOW() + 1
    Given the DB time now is "2037-04-28 23:59:59"
    And I am the user with id "31"
    When I send a DELETE request to "/current-user/group-memberships/15"
    Then the response code should be 403
    And the response error message should contain "User deletion is locked for this group"

  Scenario: Fails if lock_user_deletion_until > NOW()
    Given I am the user with id "31"
    When I send a DELETE request to "/current-user/group-memberships/15"
    Then the response code should be 403
    And the response error message should contain "User deletion is locked for this group"
