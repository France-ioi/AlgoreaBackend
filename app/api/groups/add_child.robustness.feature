Feature: Add a parent-child relation between two groups - robustness

  Background:
    Given the database has the following table 'users':
      | ID | sLogin  | idGroupSelf | idGroupOwned | sFirstName  | sLastName | allowSubgroups |
      | 1  | owner   | 21          | 22           | Jean-Michel | Blanquer  | 0              |
      | 2  | teacher | 23          | 24           | John        | Smith     | 1              |
      | 3  | student | 25          | 26           | Jane        | Doe       | 1              |
      | 4  | admin   | 27          | 28           | John        | Doe       | 1              |
    And the database has the following table 'groups':
      | ID | sName     | sType     |
      | 11 | Group A   | Class     |
      | 13 | Group B   | Class     |
      | 14 | UserAdmin | UserAdmin |
      | 15 | Root      | Base      |
      | 16 | RootSelf  | Base      |
      | 17 | RootAdmin | Base      |
      | 18 | UserSelf  | UserSelf  |
      | 19 | Team      | Team      |
    And the database has the following table 'groups_ancestors':
      | idGroupAncestor | idGroupChild | bIsSelf |
      | 13              | 11           | 1       |
      | 21              | 21           | 1       |
      | 22              | 11           | 0       |
      | 22              | 13           | 0       |
      | 24              | 13           | 0       |
      | 26              | 11           | 0       |
      | 28              | 11           | 0       |
      | 28              | 13           | 0       |
      | 28              | 14           | 0       |
      | 28              | 15           | 0       |
      | 28              | 16           | 0       |
      | 28              | 17           | 0       |
      | 28              | 18           | 0       |
      | 28              | 19           | 0       |
    And the database has the following table 'groups_groups':
      | idGroupParent | idGroupChild | sType  |
      | 13            | 11           | direct |
      | 22            | 11           | direct |
      | 22            | 13           | direct |
      | 24            | 13           | direct |
      | 26            | 11           | direct |
      | 28            | 11           | direct |
      | 28            | 13           | direct |
      | 28            | 14           | direct |
      | 28            | 15           | direct |
      | 28            | 16           | direct |
      | 28            | 17           | direct |
      | 28            | 18           | direct |
      | 28            | 19           | direct |

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

  Scenario: Parent group is UserSelf
    Given I am the user with ID "4"
    When I send a POST request to "/groups/18/relations/11"
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

  Scenario: Parent is a team and child is not a user self group
    Given I am the user with ID "4"
    When I send a POST request to "/groups/19/relations/13"
    Then the response code should be 403
    And the response error message should contain "Only users can be team members"
    And the table "groups_groups" should stay unchanged
    And the table "groups_ancestors" should stay unchanged

  Scenario Outline: Parent is a team and child is a member of another team for the same item
    Given the database table 'groups' has also the following rows:
      | ID | sName        | sType | idTeamItem |
      | 30 | Team         | Team  | 100        |
      | 31 | Another team | Team  | 100        |
    And the database table 'groups_groups' has also the following row:
      | idGroupParent | idGroupChild | sType  |
      | 28            | 30           | direct |
      | 31            | 18           | <type> |
    And the database table 'groups_ancestors' has also the following rows:
      | idGroupAncestor | idGroupChild | bIsSelf |
      | 28              | 30           | 0       |
      | 30              | 30           | 1       |
      | 31              | 18           | 0       |
      | 31              | 31           | 1       |
    And I am the user with ID "4"
    When I send a POST request to "/groups/30/relations/18"
    Then the response code should be 403
    And the response error message should contain "The user is a member of another team with the same item"
    And the table "groups_groups" should stay unchanged
    And the table "groups_ancestors" should stay unchanged
  Examples:
    | type               |
    | direct             |
    | invitationAccepted |
    | requestAccepted    |
    | joinedByCode       |
