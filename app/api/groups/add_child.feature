Feature: Add a parent-child relation between two groups

  Background:
    Given the database has the following table 'users':
      | ID | sLogin | tempUser | idGroupSelf | idGroupOwned | allowSubgroups |
      | 1  | owner  | 0        | 21          | 22           | 1              |
    And the database has the following table 'groups':
      | ID | sName       | sType     | idTeamItem |
      | 11 | Group A     | Class     | null       |
      | 13 | Group B     | Class     | null       |
      | 14 | Group C     | Class     | null       |
      | 15 | Team        | Team      | null       |
      | 16 | Team        | Team      | 100        |
      | 17 | Team        | Team      | 110        |
      | 21 | owner       | UserSelf  | null       |
      | 22 | owner-admin | UserAdmin | null       |
    And the database has the following table 'groups_groups':
      | ID | idGroupParent | idGroupChild | sType           | iChildOrder | sRole  |
      | 1  | 17            | 21           | requestAccepted | 1           | member |
      | 2  | 22            | 11           | direct          | 1           | owner  |
      | 3  | 22            | 13           | direct          | 2           | owner  |
      | 4  | 22            | 14           | direct          | 3           | owner  |
      | 5  | 22            | 15           | direct          | 4           | owner  |
      | 6  | 22            | 16           | direct          | 5           | owner  |
      | 7  | 22            | 21           | direct          | 6           | owner  |
    And the database has the following table 'groups_ancestors':
      | idGroupAncestor | idGroupChild | bIsSelf |
      | 11              | 11           | 1       |
      | 13              | 13           | 1       |
      | 14              | 14           | 1       |
      | 15              | 15           | 1       |
      | 16              | 16           | 1       |
      | 17              | 17           | 1       |
      | 17              | 21           | 0       |
      | 21              | 21           | 1       |
      | 22              | 11           | 0       |
      | 22              | 13           | 0       |
      | 22              | 14           | 0       |
      | 22              | 15           | 0       |
      | 22              | 16           | 0       |
      | 22              | 21           | 0       |
      | 22              | 22           | 1       |

  Scenario: User is an owner of the two groups and is allowed to create sub-groups
    Given I am the user with ID "1"
    When I send a POST request to "/groups/13/relations/11"
    Then the response code should be 201
    And the response body should be, in JSON:
    """
    {
      "success": true,
      "message": "created"
    }
    """
    And the table "groups_groups" should stay unchanged but the row with ID "5577006791947779410"
    And the table "groups_groups" at ID "5577006791947779410" should be:
      | idGroupParent | idGroupChild | iChildOrder | sType  | sRole  |
      | 13            | 11           | 1           | direct | member |
    And the table "groups_ancestors" should be:
      | idGroupAncestor | idGroupChild | bIsSelf |
      | 11              | 11           | 1       |
      | 13              | 11           | 0       |
      | 13              | 13           | 1       |
      | 14              | 14           | 1       |
      | 15              | 15           | 1       |
      | 16              | 16           | 1       |
      | 17              | 17           | 1       |
      | 17              | 21           | 0       |
      | 21              | 21           | 1       |
      | 22              | 11           | 0       |
      | 22              | 13           | 0       |
      | 22              | 14           | 0       |
      | 22              | 15           | 0       |
      | 22              | 16           | 0       |
      | 22              | 21           | 0       |
      | 22              | 22           | 1       |
    When I send a POST request to "/groups/13/relations/14"
    Then the response code should be 201
    And the response body should be, in JSON:
    """
    {
      "success": true,
      "message": "created"
    }
    """
    And the table "groups_groups" should stay unchanged but the rows with ID "5577006791947779410,8674665223082153551"
    And the table "groups_groups" at IDs "5577006791947779410,8674665223082153551" should be:
      | idGroupParent | idGroupChild | iChildOrder | sType  | sRole  |
      | 13            | 11           | 1           | direct | member |
      | 13            | 14           | 2           | direct | member |
    And the table "groups_ancestors" should be:
      | idGroupAncestor | idGroupChild | bIsSelf |
      | 11              | 11           | 1       |
      | 13              | 11           | 0       |
      | 13              | 13           | 1       |
      | 13              | 14           | 0       |
      | 14              | 14           | 1       |
      | 15              | 15           | 1       |
      | 16              | 16           | 1       |
      | 17              | 17           | 1       |
      | 17              | 21           | 0       |
      | 21              | 21           | 1       |
      | 22              | 11           | 0       |
      | 22              | 13           | 0       |
      | 22              | 14           | 0       |
      | 22              | 15           | 0       |
      | 22              | 16           | 0       |
      | 22              | 21           | 0       |
      | 22              | 22           | 1       |
    When I send a POST request to "/groups/13/relations/11"
    Then the response code should be 201
    And the response body should be, in JSON:
    """
    {
      "success": true,
      "message": "created"
    }
    """
    And the table "groups_groups" should stay unchanged but the rows with ID "6129484611666145821,8674665223082153551"
    And the table "groups_groups" at IDs "6129484611666145821,8674665223082153551" should be:
      | idGroupParent | idGroupChild | iChildOrder | sType  | sRole  |
      | 13            | 11           | 3           | direct | member |
      | 13            | 14           | 2           | direct | member |

  Scenario: Add a user into a team
    Given I am the user with ID "1"
    When I send a POST request to "/groups/15/relations/21"
    Then the response code should be 201
    And the response body should be, in JSON:
    """
    {
      "success": true,
      "message": "created"
    }
    """
    And the table "groups_groups" should stay unchanged but the row with ID "5577006791947779410"
    And the table "groups_groups" at ID "5577006791947779410" should be:
      | idGroupParent | idGroupChild | iChildOrder | sType  | sRole  |
      | 15            | 21           | 1           | direct | member |
    And the table "groups_ancestors" should be:
      | idGroupAncestor | idGroupChild | bIsSelf |
      | 11              | 11           | 1       |
      | 13              | 13           | 1       |
      | 14              | 14           | 1       |
      | 15              | 15           | 1       |
      | 15              | 21           | 0       |
      | 16              | 16           | 1       |
      | 17              | 17           | 1       |
      | 17              | 21           | 0       |
      | 21              | 21           | 1       |
      | 22              | 11           | 0       |
      | 22              | 13           | 0       |
      | 22              | 14           | 0       |
      | 22              | 15           | 0       |
      | 22              | 16           | 0       |
      | 22              | 21           | 0       |
      | 22              | 22           | 1       |
    When I send a POST request to "/groups/15/relations/21"
    Then the response code should be 201
    And the response body should be, in JSON:
    """
    {
      "success": true,
      "message": "created"
    }
    """
    And the table "groups_groups" should stay unchanged but the row with ID "8674665223082153551"
    And the table "groups_groups" at ID "8674665223082153551" should be:
      | idGroupParent | idGroupChild | iChildOrder | sType  | sRole  |
      | 15            | 21           | 1           | direct | member |

  Scenario Outline: Add a user into a team with idTeamItem set
    Given the database table 'groups' has also the following rows:
      | ID | sName        | sType | idTeamItem |
      | 31 | Another team | Team  | 100        |
    And the database table 'groups_groups' has also the following row:
      | ID | idGroupParent | idGroupChild | sType  |
      | 8  | 31            | 18           | <type> |
    And I am the user with ID "1"
    When I send a POST request to "/groups/16/relations/21"
    Then the response code should be 201
    And the response body should be, in JSON:
    """
    {
      "success": true,
      "message": "created"
    }
    """
    And the table "groups_groups" should stay unchanged but the row with ID "5577006791947779410"
    And the table "groups_groups" at ID "5577006791947779410" should be:
      | idGroupParent | idGroupChild | iChildOrder | sType  | sRole  |
      | 16            | 21           | 1           | direct | member |
    And the table "groups_ancestors" should be:
      | idGroupAncestor | idGroupChild | bIsSelf |
      | 11              | 11           | 1       |
      | 13              | 13           | 1       |
      | 14              | 14           | 1       |
      | 15              | 15           | 1       |
      | 16              | 16           | 1       |
      | 16              | 21           | 0       |
      | 17              | 17           | 1       |
      | 17              | 21           | 0       |
      | 18              | 18           | 1       |
      | 21              | 21           | 1       |
      | 22              | 11           | 0       |
      | 22              | 13           | 0       |
      | 22              | 14           | 0       |
      | 22              | 15           | 0       |
      | 22              | 16           | 0       |
      | 22              | 21           | 0       |
      | 22              | 22           | 1       |
      | 31              | 31           | 1       |
    When I send a POST request to "/groups/16/relations/21"
    Then the response code should be 201
    And the response body should be, in JSON:
    """
    {
      "success": true,
      "message": "created"
    }
    """
    And the table "groups_groups" should stay unchanged but the row with ID "8674665223082153551"
    And the table "groups_groups" at ID "8674665223082153551" should be:
      | idGroupParent | idGroupChild | iChildOrder | sType  | sRole  |
      | 16            | 21           | 1           | direct | member |
  Examples:
    | type              |
    | invitationSent    |
    | requestSent       |
    | invitationRefused |
    | requestRefused    |
    | removed           |
    | left              |
