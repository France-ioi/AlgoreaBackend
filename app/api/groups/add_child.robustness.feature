Feature: Add a parent-child relation between two groups - robustness

  Background:
    Given the database has the following table 'users':
      | ID | sLogin  | idGroupSelf | idGroupOwned | sFirstName  | sLastName | allowSubgroups |
      | 1  | owner   | 21          | 22           | Jean-Michel | Blanquer  | 0              |
      | 2  | teacher | 23          | 24           | John        | Smith     | 1              |
      | 3  | student | 25          | 26           | Jane        | Doe       | 1              |
      | 4  | admin   | 27          | 28           | John        | Doe       | 1              |
    And the database has the following table 'groups':
      | ID | sName         | sType     |
      | 11 | Group A       | Class     |
      | 13 | Group B       | Class     |
      | 14 | UserAdmin     | UserAdmin |
      | 15 | Root          | Base      |
      | 16 | RootSelf      | Base      |
      | 17 | RootAdmin     | Base      |
      | 18 | UserSelf      | UserSelf  |
      | 19 | Team          | Team      |
      | 21 | owner         | UserSelf  |
      | 22 | owner-admin   | UserAdmin |
      | 23 | teacher       | UserSelf  |
      | 24 | teacher-admin | UserAdmin |
      | 25 | student       | UserSelf  |
      | 26 | student-admin | UserAdmin |
      | 27 | admin         | UserSelf  |
      | 28 | admin-admin   | UserAdmin |
    And the database has the following table 'groups_ancestors':
      | idGroupAncestor | idGroupChild | bIsSelf |
      | 11              | 11           | 1       |
      | 13              | 11           | 0       |
      | 13              | 13           | 1       |
      | 14              | 14           | 1       |
      | 15              | 15           | 1       |
      | 16              | 16           | 1       |
      | 17              | 17           | 1       |
      | 18              | 18           | 1       |
      | 19              | 19           | 1       |
      | 21              | 21           | 1       |
      | 22              | 11           | 0       |
      | 22              | 13           | 0       |
      | 22              | 22           | 1       |
      | 23              | 23           | 1       |
      | 24              | 13           | 0       |
      | 24              | 24           | 1       |
      | 25              | 25           | 1       |
      | 26              | 11           | 0       |
      | 26              | 26           | 1       |
      | 27              | 27           | 1       |
      | 28              | 11           | 0       |
      | 28              | 13           | 0       |
      | 28              | 14           | 0       |
      | 28              | 15           | 0       |
      | 28              | 16           | 0       |
      | 28              | 17           | 0       |
      | 28              | 18           | 0       |
      | 28              | 19           | 0       |
      | 28              | 28           | 1       |
    And the database has the following table 'groups_groups':
      | idGroupParent | idGroupChild | iChildOrder |
      | 13            | 11           | 1           |

  Scenario: Parent group ID is wrong
    Given I am the user with ID "1"
    When I send a POST request to "/groups/abc/relations/11"
    Then the response code should be 400
    And the response error message should contain "Wrong value for parent_group_id (should be int64)"
    And the table "groups_groups" should stay unchanged
    And the table "groups_ancestors" should stay unchanged

  Scenario: Child group ID is missing
    Given I am the user with ID "1"
    When I send a POST request to "/groups/13/relations/abc"
    Then the response code should be 400
    And the response error message should contain "Wrong value for child_group_id (should be int64)"
    And the table "groups_groups" should stay unchanged
    And the table "groups_ancestors" should stay unchanged

  Scenario: User is an owner of the two groups, but is not allowed to create subgroups
    Given I am the user with ID "1"
    When I send a POST request to "/groups/13/relations/11"
    Then the response code should be 403
    And the response error message should contain "Insufficient access rights"
    And the table "groups_groups" should stay unchanged
    And the table "groups_ancestors" should stay unchanged

  Scenario: User is an owner of the parent group, but is not an owner of the child group
    Given I am the user with ID "2"
    When I send a POST request to "/groups/13/relations/11"
    Then the response code should be 403
    And the response error message should contain "Insufficient access rights"
    And the table "groups_groups" should stay unchanged
    And the table "groups_ancestors" should stay unchanged

  Scenario: User is an owner of the child group, but is not an owner of the parent group
    Given I am the user with ID "3"
    When I send a POST request to "/groups/13/relations/11"
    Then the response code should be 403
    And the response error message should contain "Insufficient access rights"
    And the table "groups_groups" should stay unchanged
    And the table "groups_ancestors" should stay unchanged

  Scenario: User does not exist
    Given I am the user with ID "404"
    When I send a POST request to "/groups/13/relations/11"
    Then the response code should be 401
    And the response error message should contain "Invalid access token"
    And the table "groups_groups" should stay unchanged
    And the table "groups_ancestors" should stay unchanged

  Scenario: Child group is UserAdmin
    Given I am the user with ID "4"
    When I send a POST request to "/groups/13/relations/14"
    Then the response code should be 403
    And the response error message should contain "Insufficient access rights"
    And the table "groups_groups" should stay unchanged
    And the table "groups_ancestors" should stay unchanged

  Scenario: Child group is Root
    Given I am the user with ID "4"
    When I send a POST request to "/groups/13/relations/15"
    Then the response code should be 403
    And the response error message should contain "Insufficient access rights"
    And the table "groups_groups" should stay unchanged
    And the table "groups_ancestors" should stay unchanged

  Scenario: Child group is RootSelf
    Given I am the user with ID "4"
    When I send a POST request to "/groups/13/relations/16"
    Then the response code should be 403
    And the response error message should contain "Insufficient access rights"
    And the table "groups_groups" should stay unchanged
    And the table "groups_ancestors" should stay unchanged

  Scenario: Child group is RootAdmin
    Given I am the user with ID "4"
    When I send a POST request to "/groups/13/relations/17"
    Then the response code should be 403
    And the response error message should contain "Insufficient access rights"
    And the table "groups_groups" should stay unchanged
    And the table "groups_ancestors" should stay unchanged

  Scenario: Child group is UserSelf
    Given I am the user with ID "4"
    When I send a POST request to "/groups/13/relations/18"
    Then the response code should be 403
    And the response error message should contain "Insufficient access rights"
    And the table "groups_groups" should stay unchanged
    And the table "groups_ancestors" should stay unchanged

  Scenario: Parent group is UserSelf
    Given I am the user with ID "4"
    When I send a POST request to "/groups/18/relations/11"
    Then the response code should be 403
    And the response error message should contain "Insufficient access rights"
    And the table "groups_groups" should stay unchanged
    And the table "groups_ancestors" should stay unchanged

  Scenario: Parent group is Team
    Given I am the user with ID "4"
    When I send a POST request to "/groups/19/relations/11"
    Then the response code should be 403
    And the response error message should contain "Insufficient access rights"
    And the table "groups_groups" should stay unchanged
    And the table "groups_ancestors" should stay unchanged

  Scenario: A group cannot become an ancestor of itself
    Given I am the user with ID "4"
    When I send a POST request to "/groups/11/relations/13"
    Then the response code should be 403
    And the response error message should contain "A group cannot become an ancestor of itself"
    And the table "groups_groups" should stay unchanged
    And the table "groups_ancestors" should stay unchanged

  Scenario: Parent and child are the same
    Given I am the user with ID "4"
    When I send a POST request to "/groups/13/relations/13"
    Then the response code should be 400
    And the response error message should contain "A group cannot become its own parent"
    And the table "groups_groups" should stay unchanged
    And the table "groups_ancestors" should stay unchanged
