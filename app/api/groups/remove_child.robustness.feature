Feature: Remove a direct parent-child relation between two groups - robustness

  Background:
    Given the database has the following table 'groups':
      | id | name     | type  |
      | 11 | Group A  | Class |
      | 13 | Group B  | Class |
      | 14 | Group C  | Class |
      | 15 | Team     | Team  |
      | 21 | owner    | User  |
      | 22 | Group    | Class |
      | 23 | teacher  | User  |
      | 52 | Root     | Base  |
      | 53 | RootSelf | Base  |
      | 55 | User     | User  |
    And the database has the following table 'users':
      | login   | group_id | first_name  | last_name |
      | owner   | 21       | Jean-Michel | Blanquer  |
      | teacher | 23       | John        | Smith     |
    And the database has the following table 'group_managers':
      | group_id | manager_id | can_manage  |
      | 13       | 21         | memberships |
      | 14       | 21         | memberships |
      | 15       | 21         | memberships |
      | 22       | 21         | memberships |
      | 52       | 21         | none        |
      | 53       | 21         | none        |
      | 55       | 21         | memberships |
      | 11       | 23         | none        |
      | 55       | 23         | none        |
    And the database has the following table 'groups_groups':
      | parent_group_id | child_group_id |
      | 11              | 55             |
      | 13              | 11             |
      | 13              | 55             |
      | 15              | 55             |
      | 22              | 13             |
      | 55              | 14             |
    And the groups ancestors are computed

  Scenario: User tries to delete a relation making a child group an orphan
    Given I am the user with id "21"
    When I send a DELETE request to "/groups/22/relations/13"
    Then the response code should be 422
    And the response error message should contain "Group 13 would become an orphan: confirm that you want to delete it"
    And the table "groups_groups" should stay unchanged
    And the table "groups_ancestors" should stay unchanged
    And the table "groups" should stay unchanged

  Scenario: Parent group id is wrong
    Given I am the user with id "21"
    When I send a DELETE request to "/groups/abc/relations/11"
    Then the response code should be 400
    And the response error message should contain "Wrong value for parent_group_id (should be int64)"
    And the table "groups_groups" should stay unchanged
    And the table "groups_ancestors" should stay unchanged

  Scenario: Child group id is missing
    Given I am the user with id "21"
    When I send a DELETE request to "/groups/13/relations/abc"
    Then the response code should be 400
    And the response error message should contain "Wrong value for child_group_id (should be int64)"
    And the table "groups_groups" should stay unchanged
    And the table "groups_ancestors" should stay unchanged

  Scenario: delete_orphans is wrong
    Given I am the user with id "21"
    When I send a DELETE request to "/groups/13/relations/11?delete_orphans=abc"
    Then the response code should be 400
    And the response error message should contain "Wrong value for delete_orphans (should have a boolean value (0 or 1))"
    And the table "groups_groups" should stay unchanged
    And the table "groups_ancestors" should stay unchanged

  Scenario: User is a manager of the child group, but is not a manager of the parent group
    Given I am the user with id "23"
    When I send a DELETE request to "/groups/13/relations/11"
    Then the response code should be 403
    And the response error message should contain "Insufficient access rights"
    And the table "groups_groups" should stay unchanged
    And the table "groups_ancestors" should stay unchanged

  Scenario: User is a manager of the two groups, but doesn't have enough rights on the parent group
    Given I am the user with id "23"
    When I send a DELETE request to "/groups/11/relations/11"
    Then the response code should be 403
    And the response error message should contain "Insufficient access rights"
    And the table "groups_groups" should stay unchanged
    And the table "groups_ancestors" should stay unchanged

  Scenario: User does not exist
    Given I am the user with id "404"
    When I send a DELETE request to "/groups/13/relations/11"
    Then the response code should be 401
    And the response error message should contain "Invalid access token"
    And the table "groups_groups" should stay unchanged
    And the table "groups_ancestors" should stay unchanged

  Scenario: Child group is Root
    Given I am the user with id "21"
    When I send a DELETE request to "/groups/13/relations/52"
    Then the response code should be 403
    And the response error message should contain "Insufficient access rights"
    And the table "groups_groups" should stay unchanged
    And the table "groups_ancestors" should stay unchanged

  Scenario: Child group is RootSelf
    Given I am the user with id "21"
    When I send a DELETE request to "/groups/13/relations/53"
    Then the response code should be 403
    And the response error message should contain "Insufficient access rights"
    And the table "groups_groups" should stay unchanged
    And the table "groups_ancestors" should stay unchanged

  Scenario: Child group is User
    Given I am the user with id "21"
    When I send a DELETE request to "/groups/13/relations/55"
    Then the response code should be 403
    And the response error message should contain "Insufficient access rights"
    And the table "groups_groups" should stay unchanged
    And the table "groups_ancestors" should stay unchanged

  Scenario: Parent group is User
    Given I am the user with id "21"
    When I send a DELETE request to "/groups/55/relations/14"
    Then the response code should be 403
    And the response error message should contain "Insufficient access rights"
    And the table "groups_groups" should stay unchanged
    And the table "groups_ancestors" should stay unchanged

  Scenario: Parent group is Team
    Given I am the user with id "21"
    When I send a DELETE request to "/groups/15/relations/55"
    Then the response code should be 403
    And the response error message should contain "Insufficient access rights"
    And the table "groups_groups" should stay unchanged
    And the table "groups_ancestors" should stay unchanged

  Scenario: Parent and child are the same
    Given I am the user with id "21"
    When I send a DELETE request to "/groups/13/relations/13"
    Then the response code should be 403
    And the response error message should contain "Insufficient access rights"
    And the table "groups_groups" should stay unchanged
    And the table "groups_ancestors" should stay unchanged

  Scenario: Relation doesn't exist
    Given I am the user with id "21"
    When I send a DELETE request to "/groups/14/relations/11"
    Then the response code should be 403
    And the response error message should contain "Insufficient access rights"
    And the table "groups_groups" should stay unchanged
    And the table "groups_ancestors" should stay unchanged
