Feature: Accept group requests - robustness
  Background:
    Given the database has the following table 'groups':
      | id  | type  | frozen_membership |
      | 11  | User  | false             |
      | 12  | User  | false             |
      | 13  | Club  | false             |
      | 14  | Class | false             |
      | 21  | User  | false             |
      | 31  | Class | true              |
      | 111 | User  | false             |
      | 121 | User  | false             |
      | 122 | User  | false             |
      | 123 | User  | false             |
      | 131 | User  | false             |
      | 141 | User  | false             |
    And the database has the following table 'users':
      | login | group_id | first_name  | last_name | grade |
      | owner | 21       | Jean-Michel | Blanquer  | 3     |
      | user  | 11       | John        | Doe       | 1     |
      | jane  | 12       | Jane        | Doe       | 1     |
    And the database has the following table 'group_managers':
      | group_id | manager_id | can_manage  |
      | 12       | 12         | memberships |
      | 13       | 21         | memberships |
      | 13       | 12         | none        |
      | 31       | 21         | memberships |
    And the database has the following table 'groups_groups':
      | parent_group_id | child_group_id |
      | 13              | 111            |
      | 13              | 121            |
      | 13              | 123            |
    And the groups ancestors are computed
    And the database has the following table 'group_pending_requests':
      | group_id | member_id | type         |
      | 13       | 21        | invitation   |
      | 13       | 31        | join_request |
      | 13       | 141       | join_request |
      | 14       | 11        | invitation   |
      | 14       | 21        | join_request |
      | 31       | 141       | join_request |

  Scenario: Fails when the user is not a manager of the parent group
    Given I am the user with id "11"
    When I send a POST request to "/groups/13/join-requests/accept?group_ids=31,141,21,11,13"
    Then the response code should be 403
    And the response error message should contain "Insufficient access rights"
    And the table "groups_groups" should stay unchanged
    And the table "group_pending_requests" should stay unchanged
    And the table "group_membership_changes" should be empty
    And the table "groups_ancestors" should stay unchanged

  Scenario: Fails when the user is a manager of the parent group, but doesn't have enough rights to manage memberships
    Given I am the user with id "12"
    When I send a POST request to "/groups/13/join-requests/accept?group_ids=31,141,21,11,13"
    Then the response code should be 403
    And the response error message should contain "Insufficient access rights"
    And the table "groups_groups" should stay unchanged
    And the table "group_pending_requests" should stay unchanged
    And the table "group_membership_changes" should be empty
    And the table "groups_ancestors" should stay unchanged

  Scenario: Fails when the user has enough rights to manage memberships, but the group is a user
    Given I am the user with id "12"
    When I send a POST request to "/groups/12/join-requests/accept?group_ids=31,141,21,11,13"
    Then the response code should be 403
    And the response error message should contain "Insufficient access rights"
    And the table "groups_groups" should stay unchanged
    And the table "group_pending_requests" should stay unchanged
    And the table "group_membership_changes" should be empty
    And the table "groups_ancestors" should stay unchanged

  Scenario: Fails when the user doesn't exist
    Given I am the user with id "404"
    When I send a POST request to "/groups/13/join-requests/accept?group_ids=31,141,21,11,13"
    Then the response code should be 401
    And the response error message should contain "Invalid access token"
    And the table "groups_groups" should stay unchanged
    And the table "group_pending_requests" should stay unchanged
    And the table "group_membership_changes" should be empty
    And the table "groups_ancestors" should stay unchanged

  Scenario: Fails when the parent group id is wrong
    Given I am the user with id "21"
    When I send a POST request to "/groups/abc/join-requests/accept?group_ids=31,141,21,11,13"
    Then the response code should be 400
    And the response error message should contain "Wrong value for parent_group_id (should be int64)"
    And the table "groups_groups" should stay unchanged
    And the table "group_pending_requests" should stay unchanged
    And the table "group_membership_changes" should be empty
    And the table "groups_ancestors" should stay unchanged

  Scenario: Fails when group_ids is wrong
    Given I am the user with id "21"
    When I send a POST request to "/groups/13/join-requests/accept?group_ids=31,abc,11,13"
    Then the response code should be 400
    And the response error message should contain "Unable to parse one of the integers given as query args (value: 'abc', param: 'group_ids')"
    And the table "groups_groups" should stay unchanged
    And the table "group_pending_requests" should stay unchanged
    And the table "group_membership_changes" should be empty
    And the table "groups_ancestors" should stay unchanged

  Scenario: Fails when the group membership is frozen
    Given I am the user with id "21"
    When I send a POST request to "/groups/31/join-requests/accept?group_ids=141"
    Then the response code should be 403
    And the response error message should contain "Group membership is frozen"
    And the table "groups_groups" should stay unchanged
    And the table "group_pending_requests" should stay unchanged
    And the table "group_membership_changes" should be empty
    And the table "groups_ancestors" should stay unchanged
