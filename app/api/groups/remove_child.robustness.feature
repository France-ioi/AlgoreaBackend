Feature: Remove a direct parent-child relation between two groups - robustness

  Background:
    Given the database has the following table 'users':
      | ID | sLogin  | idGroupSelf | idGroupOwned | sFirstName  | sLastName |
      | 1  | owner   | 21          | 22           | Jean-Michel | Blanquer  |
      | 2  | teacher | 23          | 24           | John        | Smith     |
    And the database has the following table 'groups':
      | ID | sName     | sType     |
      | 11 | Group A   | Class     |
      | 13 | Group B   | Class     |
      | 14 | Group C   | Class     |
      | 21 | Self      | UserSelf  |
      | 22 | Owned     | UserAdmin |
      | 51 | UserAdmin | UserAdmin |
      | 52 | Root      | Root      |
      | 53 | RootSelf  | RootSelf  |
      | 54 | RootAdmin | RootAdmin |
      | 55 | UserSelf  | UserSelf  |
    And the database has the following table 'groups_groups':
      | idGroupParent | idGroupChild | sType       |
      | 13            | 11           | direct      |
      | 13            | 14           | requestSent |
      | 22            | 11           | direct      |
      | 22            | 13           | direct      |
      | 22            | 14           | direct      |
      | 22            | 51           | direct      |
      | 22            | 52           | direct      |
      | 22            | 53           | direct      |
      | 22            | 54           | direct      |
      | 22            | 55           | direct      |
      | 24            | 11           | direct      |
      | 55            | 14           | direct      |
    And the database has the following table 'groups_ancestors':
      | idGroupAncestor | idGroupChild | bIsSelf |
      | 11              | 11           | 1       |
      | 13              | 11           | 0       |
      | 13              | 13           | 1       |
      | 14              | 14           | 1       |
      | 21              | 21           | 1       |
      | 22              | 11           | 0       |
      | 22              | 13           | 0       |
      | 22              | 14           | 0       |
      | 22              | 22           | 1       |
      | 22              | 51           | 0       |
      | 22              | 52           | 0       |
      | 22              | 53           | 0       |
      | 22              | 54           | 0       |
      | 22              | 55           | 0       |
      | 24              | 11           | 0       |
      | 24              | 24           | 1       |
      | 51              | 51           | 1       |
      | 52              | 52           | 1       |
      | 53              | 53           | 1       |
      | 54              | 54           | 1       |
      | 55              | 14           | 0       |
      | 55              | 55           | 1       |

  Scenario: User tries to delete a relation making a child group an orphan
    Given I am the user with ID "1"
    When I send a POST request to "/groups/22/remove_child/13"
    Then the response code should be 400
    And the response error message should contain "Group 13 would become an orphan: confirm that you want to delete it"
    And the table "groups_groups" should stay unchanged
    And the table "groups_ancestors" should stay unchanged
    And the table "groups" should stay unchanged

  Scenario: Parent group ID is wrong
    Given I am the user with ID "1"
    When I send a POST request to "/groups/abc/remove_child/11"
    Then the response code should be 400
    And the response error message should contain "Wrong value for parent_group_id (should be int64)"
    And the table "groups_groups" should stay unchanged
    And the table "groups_ancestors" should stay unchanged

  Scenario: Child group ID is missing
    Given I am the user with ID "1"
    When I send a POST request to "/groups/13/remove_child/abc"
    Then the response code should be 400
    And the response error message should contain "Wrong value for child_group_id (should be int64)"
    And the table "groups_groups" should stay unchanged
    And the table "groups_ancestors" should stay unchanged

  Scenario: delete_orphans is wrong
    Given I am the user with ID "1"
    When I send a POST request to "/groups/13/remove_child/11?delete_orphans=abc"
    Then the response code should be 400
    And the response error message should contain "Wrong value for delete_orphans (should have a boolean value (0 or 1))"
    And the table "groups_groups" should stay unchanged
    And the table "groups_ancestors" should stay unchanged

  Scenario: User is an owner of the child group, but is not an owner of the parent group
    Given I am the user with ID "2"
    When I send a POST request to "/groups/13/remove_child/11"
    Then the response code should be 403
    And the response error message should contain "Insufficient access rights"
    And the table "groups_groups" should stay unchanged
    And the table "groups_ancestors" should stay unchanged

  Scenario: User does not exist
    Given I am the user with ID "404"
    When I send a POST request to "/groups/13/remove_child/11"
    Then the response code should be 403
    And the response error message should contain "Insufficient access rights"
    And the table "groups_groups" should stay unchanged
    And the table "groups_ancestors" should stay unchanged

  Scenario: Child group is UserAdmin
    Given I am the user with ID "1"
    When I send a POST request to "/groups/13/remove_child/51"
    Then the response code should be 403
    And the response error message should contain "Insufficient access rights"
    And the table "groups_groups" should stay unchanged
    And the table "groups_ancestors" should stay unchanged

  Scenario: Child group is Root
    Given I am the user with ID "1"
    When I send a POST request to "/groups/13/remove_child/52"
    Then the response code should be 403
    And the response error message should contain "Insufficient access rights"
    And the table "groups_groups" should stay unchanged
    And the table "groups_ancestors" should stay unchanged

  Scenario: Child group is RootSelf
    Given I am the user with ID "1"
    When I send a POST request to "/groups/13/remove_child/53"
    Then the response code should be 403
    And the response error message should contain "Insufficient access rights"
    And the table "groups_groups" should stay unchanged
    And the table "groups_ancestors" should stay unchanged

  Scenario: Child group is RootAdmin
    Given I am the user with ID "1"
    When I send a POST request to "/groups/13/remove_child/54"
    Then the response code should be 403
    And the response error message should contain "Insufficient access rights"
    And the table "groups_groups" should stay unchanged
    And the table "groups_ancestors" should stay unchanged

  Scenario: Parent group is UserSelf
    Given I am the user with ID "1"
    When I send a POST request to "/groups/55/remove_child/14"
    Then the response code should be 403
    And the response error message should contain "Insufficient access rights"
    And the table "groups_groups" should stay unchanged
    And the table "groups_ancestors" should stay unchanged

  Scenario: Parent and child are the same
    Given I am the user with ID "1"
    When I send a POST request to "/groups/13/remove_child/13"
    Then the response code should be 403
    And the response error message should contain "Insufficient access rights"
    And the table "groups_groups" should stay unchanged
    And the table "groups_ancestors" should stay unchanged

  Scenario: Relation doesn't exist
    Given I am the user with ID "1"
    When I send a POST request to "/groups/14/remove_child/11"
    Then the response code should be 403
    And the response error message should contain "Insufficient access rights"
    And the table "groups_groups" should stay unchanged
    And the table "groups_ancestors" should stay unchanged

  Scenario: Relation is not direct
    Given I am the user with ID "1"
    When I send a POST request to "/groups/13/remove_child/14"
    Then the response code should be 403
    And the response error message should contain "Insufficient access rights"
    And the table "groups_groups" should stay unchanged
    And the table "groups_ancestors" should stay unchanged

