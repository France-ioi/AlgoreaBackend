Feature: Join a group using a code (groupsJoinByCode)
  Background:
    Given the database has the following table 'users':
      | ID | idGroupSelf | idGroupOwned |
      | 1  | 21          | 22           |
    And the database has the following table 'groups':
      | ID | sType     | sCode      | sCodeEnd            | sCodeTimer | bFreeAccess |
      | 11 | Team      | 3456789abc | 2037-05-29 06:38:38 | 01:02:03   | true        |
      | 12 | Team      | abc3456789 | null                | 12:34:56   | true        |
      | 14 | Team      | cba9876543 | null                | null       | true        |
      | 21 | UserSelf  | null       | null                | null       | false       |
      | 22 | UserAdmin | null       | null                | null       | false       |
    And the database has the following table 'groups_ancestors':
      | idGroupAncestor | idGroupChild | bIsSelf |
      | 11              | 11           | 1       |
      | 12              | 12           | 1       |
      | 14              | 14           | 1       |
      | 21              | 21           | 1       |
      | 22              | 22           | 1       |
    And the database has the following table 'groups_groups':
      | ID | idGroupParent | idGroupChild | sType              | sStatusDate         |
      | 1  | 11            | 21           | invitationSent     | 2017-04-29 06:38:38 |
      | 7  | 14            | 21           | requestSent        | 2017-02-21 06:38:38 |

  Scenario: Successfully join an group
    Given I am the user with ID "1"
    When I send a POST request to "/current-user/group-memberships/by-code?code=3456789abc"
    Then the response code should be 201
    And the response body should be, in JSON:
    """
    {
      "success": true,
      "message": "created",
      "data": {"changed": true}
    }
    """
    And the table "groups" should stay unchanged
    And the table "groups_groups" should be:
      | idGroupParent | idGroupChild | sType        | (sStatusDate IS NOT NULL) AND (ABS(TIMESTAMPDIFF(SECOND, sStatusDate, NOW())) < 3) |
      | 11            | 21           | joinedByCode | 1                                                                                  |
      | 14            | 21           | requestSent  | 0                                                                                  |
    And the table "groups_ancestors" should be:
      | idGroupAncestor | idGroupChild | bIsSelf |
      | 11              | 11           | 1       |
      | 11              | 21           | 0       |
      | 12              | 12           | 1       |
      | 14              | 14           | 1       |
      | 21              | 21           | 1       |
      | 22              | 22           | 1       |

  Scenario: Updates the sCodeEnd
    Given I am the user with ID "1"
    When I send a POST request to "/current-user/group-memberships/by-code?code=abc3456789"
    Then the response code should be 201
    And the response body should be, in JSON:
    """
    {
      "success": true,
      "message": "created",
      "data": {"changed": true}
    }
    """
    And the table "groups" should stay unchanged but the row with ID "12"
    And the table "groups" at ID "12" should be:
      | ID | sType | sCode      | sCodeTimer | bFreeAccess | TIMESTAMPDIFF(SECOND, sCodeEnd, ADDTIME(NOW(), "12:34:56")) < 3 |
      | 12 | Team  | abc3456789 | 12:34:56   | true        | 1                                                               |
    And the table "groups_groups" should be:
      | idGroupParent | idGroupChild | sType          | (sStatusDate IS NOT NULL) AND (ABS(TIMESTAMPDIFF(SECOND, sStatusDate, NOW())) < 3) |
      | 11            | 21           | invitationSent | 0                                                                                  |
      | 12            | 21           | joinedByCode   | 1                                                                                  |
      | 14            | 21           | requestSent    | 0                                                                                  |
    And the table "groups_ancestors" should be:
      | idGroupAncestor | idGroupChild | bIsSelf |
      | 11              | 11           | 1       |
      | 12              | 12           | 1       |
      | 12              | 21           | 0       |
      | 14              | 14           | 1       |
      | 21              | 21           | 1       |
      | 22              | 22           | 1       |

  Scenario: Doesn't update the sCodeEnd if sCodeTimer is null
    Given I am the user with ID "1"
    When I send a POST request to "/current-user/group-memberships/by-code?code=cba9876543"
    Then the response code should be 201
    And the response body should be, in JSON:
    """
    {
      "success": true,
      "message": "created",
      "data": {"changed": true}
    }
    """
    And the table "groups" should stay unchanged
    And the table "groups_groups" should be:
      | idGroupParent | idGroupChild | sType          | (sStatusDate IS NOT NULL) AND (ABS(TIMESTAMPDIFF(SECOND, sStatusDate, NOW())) < 3) |
      | 11            | 21           | invitationSent | 0                                                                                  |
      | 14            | 21           | joinedByCode   | 1                                                                                  |
    And the table "groups_ancestors" should be:
      | idGroupAncestor | idGroupChild | bIsSelf |
      | 11              | 11           | 1       |
      | 12              | 12           | 1       |
      | 14              | 14           | 1       |
      | 14              | 21           | 0       |
      | 21              | 21           | 1       |
      | 22              | 22           | 1       |
