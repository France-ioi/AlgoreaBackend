Feature: Accept requests to leave a group - robustness
  Background:
    Given the database has the following table 'groups':
      | id  | type    |
      | 11  | Class   |
      | 13  | Team    |
      | 14  | Friends |
      | 21  | User    |
      | 31  | User    |
      | 111 | User    |
      | 121 | User    |
      | 122 | User    |
      | 123 | User    |
      | 131 | User    |
      | 141 | User    |
      | 151 | User    |
      | 161 | User    |
      | 444 | Team    |
    And the database has the following table 'users':
      | login | group_id | first_name  | last_name | grade |
      | owner | 21       | Jean-Michel | Blanquer  | 3     |
    And the database has the following table 'groups_ancestors':
      | ancestor_group_id | child_group_id |
      | 11                | 11             |
      | 13                | 13             |
      | 13                | 111            |
      | 13                | 121            |
      | 13                | 123            |
      | 13                | 151            |
      | 14                | 14             |
      | 21                | 21             |
      | 31                | 31             |
      | 111               | 111            |
      | 121               | 121            |
      | 122               | 122            |
      | 123               | 123            |
      | 151               | 151            |
      | 161               | 161            |
    And the database has the following table 'groups_groups':
      | parent_group_id | child_group_id |
      | 13              | 31             |
      | 13              | 111            |
      | 13              | 121            |
      | 13              | 123            |
      | 13              | 141            |
      | 13              | 151            |
    And the database has the following table 'group_pending_requests':
      | group_id | member_id | type          |
      | 13       | 21        | invitation    |
      | 13       | 31        | leave_request |
      | 13       | 141       | leave_request |
      | 13       | 161       | join_request  |
      | 14       | 11        | invitation    |
      | 14       | 21        | join_request  |

  Scenario: Fails when the user is not a manager of the parent group
    Given I am the user with id "21"
    When I send a POST request to "/groups/13/leave-requests/reject?group_ids=31,141,21,11,13"
    Then the response code should be 403
    And the response error message should contain "Insufficient access rights"
    And the table "groups_groups" should stay unchanged
    And the table "group_pending_requests" should stay unchanged
    And the table "group_membership_changes" should be empty
    And the table "groups_ancestors" should stay unchanged

  Scenario: Fails when the user is a manager of the parent group, but doesn't have enough rights to manage memberships
    Given I am the user with id "21"
    And the database has the following table 'group_managers':
      | group_id | manager_id | can_manage |
      | 13       | 21         | none       |
    When I send a POST request to "/groups/13/leave-requests/reject?group_ids=31,141,21,11,13"
    Then the response code should be 403
    And the response error message should contain "Insufficient access rights"
    And the table "groups_groups" should stay unchanged
    And the table "group_pending_requests" should stay unchanged
    And the table "group_membership_changes" should be empty
    And the table "groups_ancestors" should stay unchanged

  Scenario: Fails when the user has enough rights to manage memberships, but the group is a user group
    Given I am the user with id "21"
    And the database has the following table 'group_managers':
      | group_id | manager_id | can_manage  |
      | 21       | 21         | memberships |
    When I send a POST request to "/groups/21/leave-requests/reject?group_ids=31,141,21,11,13"
    Then the response code should be 403
    And the response error message should contain "Insufficient access rights"
    And the table "groups_groups" should stay unchanged
    And the table "group_pending_requests" should stay unchanged
    And the table "group_membership_changes" should be empty
    And the table "groups_ancestors" should stay unchanged

  Scenario: Fails when the parent group id is wrong
    Given I am the user with id "21"
    When I send a POST request to "/groups/abc/leave-requests/reject?group_ids=31,141,21,11,13"
    Then the response code should be 400
    And the response error message should contain "Wrong value for parent_group_id (should be int64)"
    And the table "groups_groups" should stay unchanged
    And the table "group_pending_requests" should stay unchanged
    And the table "group_membership_changes" should be empty
    And the table "groups_ancestors" should stay unchanged

  Scenario: Fails when group_ids is wrong
    Given I am the user with id "21"
    When I send a POST request to "/groups/13/leave-requests/reject?group_ids=31,abc,11,13"
    Then the response code should be 400
    And the response error message should contain "Unable to parse one of the integers given as query args (value: 'abc', param: 'group_ids')"
    And the table "groups_groups" should stay unchanged
    And the table "group_pending_requests" should stay unchanged
    And the table "group_membership_changes" should be empty
    And the table "groups_ancestors" should stay unchanged
