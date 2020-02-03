Feature: User withdraws a request to join a group - robustness
  Background:
    Given the database has the following table 'groups':
      | id |
      | 11 |
      | 14 |
      | 21 |
      | 22 |
    And the database has the following table 'users':
      | group_id | login |
      | 21       | john  |
      | 22       | jane  |
    And the database has the following table 'groups_ancestors':
      | ancestor_group_id | child_group_id |
      | 11                | 11             |
      | 14                | 14             |
      | 21                | 21             |
    And the database has the following table 'group_pending_requests':
      | group_id | member_id | type         | at                  |
      | 14       | 21        | join_request | 2019-05-30 11:00:00 |

  Scenario: User tries to withdraw a non-existing join request
    Given I am the user with id "21"
    When I send a DELETE request to "/current-user/group-requests/11"
    Then the response code should be 404
    And the response error message should contain "No such relation"
    And the table "groups_groups" should be empty
    And the table "group_pending_requests" should stay unchanged
    And the table "group_membership_changes" should be empty
    And the table "groups_ancestors" should stay unchanged

  Scenario: Fails when the group id is wrong
    Given I am the user with id "21"
    When I send a DELETE request to "/current-user/group-requests/abc"
    Then the response code should be 400
    And the response error message should contain "Wrong value for group_id (should be int64)"
    And the table "groups_groups" should be empty
    And the table "group_pending_requests" should stay unchanged
    And the table "group_membership_changes" should be empty
    And the table "groups_ancestors" should stay unchanged
