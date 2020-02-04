Feature: Add a parent-child relation between two groups - robustness

  Background:
    Given the database has the following table 'groups':
      | id | name     | type  |
      | 11 | Group A  | Class |
      | 13 | Group B  | Class |
      | 15 | Root     | Base  |
      | 16 | RootSelf | Base  |
      | 18 | User     | User  |
      | 19 | Team     | Team  |
      | 21 | owner    | User  |
      | 25 | student  | User  |
      | 27 | admin    | User  |
      | 77 | Group C  | Class |
      | 78 | Group D  | Class |
      | 79 | Group E  | Class |
    And the database has the following table 'users':
      | login   | group_id | first_name  | last_name | allow_subgroups |
      | owner   | 21       | Jean-Michel | Blanquer  | 0               |
      | student | 25       | Jane        | Doe       | 1               |
      | admin   | 27       | John        | Doe       | 1               |
    And the database has the following table 'group_managers':
      | group_id | manager_id | can_manage            |
      | 11       | 21         | memberships_and_group |
      | 13       | 21         | memberships           |
      | 78       | 21         | memberships           |
      | 79       | 21         | none                  |
      | 11       | 25         | memberships           |
      | 11       | 27         | memberships_and_group |
      | 13       | 27         | memberships_and_group |
      | 15       | 27         | memberships_and_group |
      | 16       | 27         | memberships_and_group |
      | 18       | 27         | memberships_and_group |
      | 19       | 27         | memberships_and_group |
    And the database has the following table 'groups_ancestors':
      | ancestor_group_id | child_group_id |
      | 11                | 11             |
      | 13                | 11             |
      | 13                | 13             |
      | 15                | 15             |
      | 16                | 16             |
      | 18                | 18             |
      | 19                | 19             |
      | 21                | 21             |
      | 25                | 25             |
      | 27                | 27             |
    And the database has the following table 'groups_groups':
      | parent_group_id | child_group_id | child_order |
      | 13              | 11             | 1           |

  Scenario: Parent group id is wrong
    Given I am the user with id "21"
    When I send a POST request to "/groups/abc/relations/11"
    Then the response code should be 400
    And the response error message should contain "Wrong value for parent_group_id (should be int64)"
    And the table "groups_groups" should stay unchanged
    And the table "groups_ancestors" should stay unchanged

  Scenario: Child group id is missing
    Given I am the user with id "21"
    When I send a POST request to "/groups/13/relations/abc"
    Then the response code should be 400
    And the response error message should contain "Wrong value for child_group_id (should be int64)"
    And the table "groups_groups" should stay unchanged
    And the table "groups_ancestors" should stay unchanged

  Scenario: User is a manager of the two groups, but is not allowed to create subgroups
    Given I am the user with id "21"
    When I send a POST request to "/groups/13/relations/11"
    Then the response code should be 403
    And the response error message should contain "Insufficient access rights"
    And the table "groups_groups" should stay unchanged
    And the table "groups_ancestors" should stay unchanged

  Scenario: User is a manager of the parent group, but is not a manager of the child group
    Given I am the user with id "21"
    When I send a POST request to "/groups/13/relations/77"
    Then the response code should be 403
    And the response error message should contain "Insufficient access rights"
    And the table "groups_groups" should stay unchanged
    And the table "groups_ancestors" should stay unchanged

  Scenario: User is a manager of the child group, but is not a manager of the parent group
    Given I am the user with id "25"
    When I send a POST request to "/groups/13/relations/11"
    Then the response code should be 403
    And the response error message should contain "Insufficient access rights"
    And the table "groups_groups" should stay unchanged
    And the table "groups_ancestors" should stay unchanged

  Scenario: User is a manager of the two groups, but doesn't have enough rights on the child group
    Given I am the user with id "21"
    When I send a POST request to "/groups/13/relations/78"
    Then the response code should be 403
    And the response error message should contain "Insufficient access rights"
    And the table "groups_groups" should stay unchanged
    And the table "groups_ancestors" should stay unchanged

  Scenario: User is a manager of the two groups, but doesn't have enough rights on the parent group
    Given I am the user with id "21"
    When I send a POST request to "/groups/79/relations/11"
    Then the response code should be 403
    And the response error message should contain "Insufficient access rights"
    And the table "groups_groups" should stay unchanged
    And the table "groups_ancestors" should stay unchanged

  Scenario: User does not exist
    Given I am the user with id "404"
    When I send a POST request to "/groups/13/relations/11"
    Then the response code should be 401
    And the response error message should contain "Invalid access token"
    And the table "groups_groups" should stay unchanged
    And the table "groups_ancestors" should stay unchanged

  Scenario: Child group is Root
    Given I am the user with id "27"
    When I send a POST request to "/groups/13/relations/15"
    Then the response code should be 403
    And the response error message should contain "Insufficient access rights"
    And the table "groups_groups" should stay unchanged
    And the table "groups_ancestors" should stay unchanged

  Scenario: Child group is RootSelf
    Given I am the user with id "27"
    When I send a POST request to "/groups/13/relations/16"
    Then the response code should be 403
    And the response error message should contain "Insufficient access rights"
    And the table "groups_groups" should stay unchanged
    And the table "groups_ancestors" should stay unchanged

  Scenario: Child group is User
    Given I am the user with id "27"
    When I send a POST request to "/groups/13/relations/18"
    Then the response code should be 403
    And the response error message should contain "Insufficient access rights"
    And the table "groups_groups" should stay unchanged
    And the table "groups_ancestors" should stay unchanged

  Scenario: Parent group is User
    Given I am the user with id "27"
    When I send a POST request to "/groups/18/relations/11"
    Then the response code should be 403
    And the response error message should contain "Insufficient access rights"
    And the table "groups_groups" should stay unchanged
    And the table "groups_ancestors" should stay unchanged

  Scenario: Parent group is Team
    Given I am the user with id "27"
    When I send a POST request to "/groups/19/relations/11"
    Then the response code should be 403
    And the response error message should contain "Insufficient access rights"
    And the table "groups_groups" should stay unchanged
    And the table "groups_ancestors" should stay unchanged

  Scenario: A group cannot become an ancestor of itself
    Given I am the user with id "27"
    When I send a POST request to "/groups/11/relations/13"
    Then the response code should be 403
    And the response error message should contain "A group cannot become an ancestor of itself"
    And the table "groups_groups" should stay unchanged
    And the table "groups_ancestors" should stay unchanged

  Scenario: Parent and child are the same
    Given I am the user with id "27"
    When I send a POST request to "/groups/13/relations/13"
    Then the response code should be 400
    And the response error message should contain "A group cannot become its own parent"
    And the table "groups_groups" should stay unchanged
    And the table "groups_ancestors" should stay unchanged
