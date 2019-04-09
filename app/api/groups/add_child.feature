Feature: Add a parent-child relation between two groups

  Background:
    Given the database has the following table 'users':
      | ID | sLogin | tempUser | idGroupSelf | idGroupOwned | sFirstName  | sLastName | allowSubgroups |
      | 1  | owner  | 0        | 21          | 22           | Jean-Michel | Blanquer  | 1              |
    And the database has the following table 'groups':
      | ID | sName   | sType     |
      | 11 | Group A | Class     |
      | 13 | Group B | Class     |
      | 14 | Group C | UserAdmin |
      | 21 | Self    | UserSelf  |
      | 22 | Owned   | UserAdmin |
    And the database has the following table 'groups_groups':
      | idGroupParent | idGroupChild | sType  | iChildOrder |
      | 22            | 11           | direct | 1           |
      | 22            | 13           | direct | 1           |
    And the database has the following table 'groups_ancestors':
      | idGroupAncestor | idGroupChild | bIsSelf |
      | 11              | 11           | 1       |
      | 13              | 13           | 1       |
      | 14              | 14           | 1       |
      | 21              | 21           | 1       |
      | 22              | 11           | 0       |
      | 22              | 13           | 0       |
      | 22              | 22           | 1       |

  Scenario: User is an owner of the two groups and is allowed to create sub-groups
    Given I am the user with ID "1"
    When I send a POST request to "/groups/13/add_child/11"
    Then the response code should be 201
    And the response body should be, in JSON:
    """
    {
      "success": true,
      "message": "created"
    }
    """
    And the table "groups_groups" should be:
      | idGroupParent | idGroupChild | iChildOrder | sType  | sRole  |
      | 13            | 11           | 1           | direct | member |
      | 22            | 11           | 1           | direct | member |
      | 22            | 13           | 1           | direct | member |
    And the table "groups_ancestors" should be:
      | idGroupAncestor | idGroupChild | bIsSelf |
      | 11              | 11           | 1       |
      | 13              | 11           | 0       |
      | 13              | 13           | 1       |
      | 14              | 14           | 1       |
      | 21              | 21           | 1       |
      | 22              | 11           | 0       |
      | 22              | 13           | 0       |
      | 22              | 22           | 1       |
