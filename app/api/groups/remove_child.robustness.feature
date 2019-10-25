Feature: Remove a direct parent-child relation between two groups - robustness

  Background:
    Given the database has the following table 'groups':
      | id | name          | type      |
      | 11 | Group A       | Class     |
      | 13 | Group B       | Class     |
      | 14 | Group C       | Class     |
      | 15 | Team          | Team      |
      | 21 | owner         | UserSelf  |
      | 22 | owner-admin   | UserAdmin |
      | 23 | teacher       | UserSelf  |
      | 24 | teacher-admin | UserAdmin |
      | 51 | UserAdmin     | UserAdmin |
      | 52 | Root          | Base      |
      | 53 | RootSelf      | Base      |
      | 54 | RootAdmin     | Base      |
      | 55 | UserSelf      | UserSelf  |
    And the database has the following table 'users':
      | login   | group_id | owned_group_id | first_name  | last_name |
      | owner   | 21       | 22             | Jean-Michel | Blanquer  |
      | teacher | 23       | 24             | John        | Smith     |
    And the database has the following table 'groups_groups':
      | parent_group_id | child_group_id | type               |
      | 13              | 11             | direct             |
      | 13              | 14             | requestSent        |
      | 13              | 55             | invitationAccepted |
      | 15              | 55             | requestAccepted    |
      | 22              | 11             | direct             |
      | 22              | 13             | direct             |
      | 22              | 14             | direct             |
      | 22              | 51             | direct             |
      | 22              | 52             | direct             |
      | 22              | 53             | direct             |
      | 22              | 54             | direct             |
      | 22              | 55             | direct             |
      | 24              | 11             | direct             |
      | 55              | 14             | direct             |
    And the database has the following table 'groups_ancestors':
      | ancestor_group_id | child_group_id | is_self |
      | 11                | 11             | 1       |
      | 13                | 11             | 0       |
      | 13                | 13             | 1       |
      | 13                | 55             | 0       |
      | 14                | 14             | 1       |
      | 15                | 15             | 1       |
      | 15                | 55             | 0       |
      | 21                | 21             | 1       |
      | 22                | 11             | 0       |
      | 22                | 13             | 0       |
      | 22                | 14             | 0       |
      | 22                | 15             | 0       |
      | 22                | 22             | 1       |
      | 22                | 51             | 0       |
      | 22                | 52             | 0       |
      | 22                | 53             | 0       |
      | 22                | 54             | 0       |
      | 22                | 55             | 0       |
      | 24                | 11             | 0       |
      | 24                | 24             | 1       |
      | 51                | 51             | 1       |
      | 52                | 52             | 1       |
      | 53                | 53             | 1       |
      | 54                | 54             | 1       |
      | 55                | 14             | 0       |
      | 55                | 55             | 1       |

  Scenario: User tries to delete a relation making a child group an orphan
    Given I am the user with group_id "21"
    When I send a DELETE request to "/groups/22/relations/13"
    Then the response code should be 422
    And the response error message should contain "Group 13 would become an orphan: confirm that you want to delete it"
    And the table "groups_groups" should stay unchanged
    And the table "groups_ancestors" should stay unchanged
    And the table "groups" should stay unchanged

  Scenario: Parent group id is wrong
    Given I am the user with group_id "21"
    When I send a DELETE request to "/groups/abc/relations/11"
    Then the response code should be 400
    And the response error message should contain "Wrong value for parent_group_id (should be int64)"
    And the table "groups_groups" should stay unchanged
    And the table "groups_ancestors" should stay unchanged

  Scenario: Child group id is missing
    Given I am the user with group_id "21"
    When I send a DELETE request to "/groups/13/relations/abc"
    Then the response code should be 400
    And the response error message should contain "Wrong value for child_group_id (should be int64)"
    And the table "groups_groups" should stay unchanged
    And the table "groups_ancestors" should stay unchanged

  Scenario: delete_orphans is wrong
    Given I am the user with group_id "21"
    When I send a DELETE request to "/groups/13/relations/11?delete_orphans=abc"
    Then the response code should be 400
    And the response error message should contain "Wrong value for delete_orphans (should have a boolean value (0 or 1))"
    And the table "groups_groups" should stay unchanged
    And the table "groups_ancestors" should stay unchanged

  Scenario: User is an owner of the child group, but is not an owner of the parent group
    Given I am the user with group_id "23"
    When I send a DELETE request to "/groups/13/relations/11"
    Then the response code should be 403
    And the response error message should contain "Insufficient access rights"
    And the table "groups_groups" should stay unchanged
    And the table "groups_ancestors" should stay unchanged

  Scenario: User does not exist
    Given I am the user with group_id "404"
    When I send a DELETE request to "/groups/13/relations/11"
    Then the response code should be 401
    And the response error message should contain "Invalid access token"
    And the table "groups_groups" should stay unchanged
    And the table "groups_ancestors" should stay unchanged

  Scenario: Child group is UserAdmin
    Given I am the user with group_id "21"
    When I send a DELETE request to "/groups/13/relations/51"
    Then the response code should be 403
    And the response error message should contain "Insufficient access rights"
    And the table "groups_groups" should stay unchanged
    And the table "groups_ancestors" should stay unchanged

  Scenario: Child group is Root
    Given I am the user with group_id "21"
    When I send a DELETE request to "/groups/13/relations/52"
    Then the response code should be 403
    And the response error message should contain "Insufficient access rights"
    And the table "groups_groups" should stay unchanged
    And the table "groups_ancestors" should stay unchanged

  Scenario: Child group is RootSelf
    Given I am the user with group_id "21"
    When I send a DELETE request to "/groups/13/relations/53"
    Then the response code should be 403
    And the response error message should contain "Insufficient access rights"
    And the table "groups_groups" should stay unchanged
    And the table "groups_ancestors" should stay unchanged

  Scenario: Child group is RootAdmin
    Given I am the user with group_id "21"
    When I send a DELETE request to "/groups/13/relations/54"
    Then the response code should be 403
    And the response error message should contain "Insufficient access rights"
    And the table "groups_groups" should stay unchanged
    And the table "groups_ancestors" should stay unchanged

  Scenario: Child group is UserSelf
    Given I am the user with group_id "21"
    When I send a DELETE request to "/groups/13/relations/55"
    Then the response code should be 403
    And the response error message should contain "Insufficient access rights"
    And the table "groups_groups" should stay unchanged
    And the table "groups_ancestors" should stay unchanged

  Scenario: Parent group is UserSelf
    Given I am the user with group_id "21"
    When I send a DELETE request to "/groups/55/relations/14"
    Then the response code should be 403
    And the response error message should contain "Insufficient access rights"
    And the table "groups_groups" should stay unchanged
    And the table "groups_ancestors" should stay unchanged

  Scenario: Parent group is Team
    Given I am the user with group_id "21"
    When I send a DELETE request to "/groups/55/relations/15"
    Then the response code should be 403
    And the response error message should contain "Insufficient access rights"
    And the table "groups_groups" should stay unchanged
    And the table "groups_ancestors" should stay unchanged

  Scenario: Parent and child are the same
    Given I am the user with group_id "21"
    When I send a DELETE request to "/groups/13/relations/13"
    Then the response code should be 403
    And the response error message should contain "Insufficient access rights"
    And the table "groups_groups" should stay unchanged
    And the table "groups_ancestors" should stay unchanged

  Scenario: Relation doesn't exist
    Given I am the user with group_id "21"
    When I send a DELETE request to "/groups/14/relations/11"
    Then the response code should be 403
    And the response error message should contain "Insufficient access rights"
    And the table "groups_groups" should stay unchanged
    And the table "groups_ancestors" should stay unchanged

  Scenario: Relation is not direct
    Given I am the user with group_id "21"
    When I send a DELETE request to "/groups/13/relations/14"
    Then the response code should be 403
    And the response error message should contain "Insufficient access rights"
    And the table "groups_groups" should stay unchanged
    And the table "groups_ancestors" should stay unchanged
