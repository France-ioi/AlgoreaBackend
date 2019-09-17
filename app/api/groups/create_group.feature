Feature: Create a group (groupCreate)

  Background:
    Given the database has the following table 'users':
      | ID | sLogin | tempUser | idGroupSelf | idGroupOwned | sFirstName  | sLastName | allowSubgroups |
      | 1  | owner  | 0        | 21          | 22           | Jean-Michel | Blanquer  | 1              |
    And the database has the following table 'groups':
      | ID | sName       | sType     |
      | 21 | owner       | UserSelf  |
      | 22 | owner-admin | UserAdmin |
    And the database has the following table 'groups_ancestors':
      | idGroupAncestor | idGroupChild | bIsSelf |
      | 21              | 21           | 1       |
      | 22              | 22           | 1       |
    And the database has the following table 'groups_items':
      | idGroup | idItem | sCachedFullAccessDate | sCachedPartialAccessDate | sCachedGrayedAccessDate | idUserCreated |
      | 21      | 10     | 2019-07-16 21:28:47   | null                     | null                    | 1             |
      | 21      | 11     | null                  | 2019-07-16 21:28:47      | null                    | 1             |
      | 21      | 12     | null                  | null                     | 2019-07-16 21:28:47     | 1             |

  Scenario Outline: Create a group
    Given I am the user with ID "1"
    When I send a POST request to "/groups" with the following body:
    """
    {
      "name": "some name",
      "type": "<group_type>"
      <item_spec>
    }
    """
    Then the response code should be 201
    And the response body should be, in JSON:
    """
    {
      "success": true,
      "message": "created",
      "data": {"id":"5577006791947779410"}
    }
    """
    And the table "groups" should stay unchanged but the row with ID "5577006791947779410"
    And the table "groups" at ID "5577006791947779410" should be:
      | ID                  | sName     | sType        | idTeamItem     | TIMESTAMPDIFF(SECOND, NOW(), sDateCreated) < 3 |
      | 5577006791947779410 | some name | <group_type> | <want_item_id> | true                                           |
    And the table "groups_groups" should be:
      | idGroupParent       | idGroupChild        | iChildOrder | sType  | sRole  |
      | 22                  | 5577006791947779410 | 1           | direct | owner  |
    And the table "groups_ancestors" should be:
      | idGroupAncestor     | idGroupChild        | bIsSelf |
      | 21                  | 21                  | 1       |
      | 22                  | 22                  | 1       |
      | 22                  | 5577006791947779410 | 0       |
      | 5577006791947779410 | 5577006791947779410 | 1       |
  Examples:
    | group_type | item_spec         | want_item_id |
    | Class      |                   | null         |
    | Team       |                   | null         |
    | Team       | , "item_id": "10" | 10           | # full access
    | Team       | , "item_id": "11" | 11           | # partial access
    | Team       | , "item_id": "12" | 12           | # grayed access
    | Club       |                   | null         |
    | Friends    |                   | null         |
    | Other      |                   | null         |
