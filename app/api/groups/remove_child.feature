Feature: Remove a direct parent-child relation between two groups

  Background:
    Given the database has the following table 'users':
      | ID | sLogin | idGroupSelf | idGroupOwned | sFirstName  | sLastName | allowSubgroups |
      | 1  | owner  | 21          | 22           | Jean-Michel | Blanquer  | 1              |
    And the database has the following table 'groups':
      | ID | sName   | sType     |
      | 11 | Group A | Class     |
      | 13 | Group B | Class     |
      | 14 | Group C | Class     |
      | 21 | Self    | UserSelf  |
      | 22 | Owned   | UserAdmin |

    And the database has the following table 'groups_groups':
      | idGroupParent | idGroupChild | sType  |
      | 13            | 11           | direct |
      | 22            | 11           | direct |
      | 22            | 13           | direct |
      | 22            | 14           | direct |
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

  Scenario: User deletes a relation
    Given I am the user with ID "1"
    When I send a POST request to "/groups/13/remove_child/11"
    Then the response code should be 200
    And the response body should be, in JSON:
    """
    {
      "success": true,
      "message": "deleted"
    }
    """
    And the table "groups_groups" should be:
      | idGroupParent | idGroupChild | sType  | sRole  |
      | 22            | 11           | direct | member |
      | 22            | 13           | direct | member |
      | 22            | 14           | direct | member |
    And the table "groups_ancestors" should be:
      | idGroupAncestor | idGroupChild | bIsSelf |
      | 11              | 11           | 1       |
      | 13              | 13           | 1       |
      | 14              | 14           | 1       |
      | 21              | 21           | 1       |
      | 22              | 11           | 0       |
      | 22              | 13           | 0       |
      | 22              | 14           | 0       |
      | 22              | 22           | 1       |
    And the table "groups" should stay unchanged

  Scenario: User deletes a relation and an orphaned child group
    Given I am the user with ID "1"
    When I send a POST request to "/groups/22/remove_child/13?delete_orphans=1"
    Then the response code should be 200
    And the response body should be, in JSON:
    """
    {
      "success": true,
      "message": "deleted"
    }
    """
    And the table "groups_groups" should be:
      | idGroupParent | idGroupChild | sType  | sRole  |
      | 22            | 11           | direct | member |
      | 22            | 14           | direct | member |
    And the table "groups_ancestors" should be:
      | idGroupAncestor | idGroupChild | bIsSelf |
      | 11              | 11           | 1       |
      | 14              | 14           | 1       |
      | 21              | 21           | 1       |
      | 22              | 11           | 0       |
      | 22              | 14           | 0       |
      | 22              | 22           | 1       |
    And the table "groups" should stay unchanged but the row with ID "13"
    And the table "groups" at ID "13" should be:
      | ID |
