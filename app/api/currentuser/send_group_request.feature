Feature: User sends a request to join a group
  Background:
    Given the database has the following table 'users':
      | ID | idGroupSelf | idGroupOwned |
      | 1  | 21          | 22           |
    And the database has the following table 'groups':
      | ID | bFreeAccess |
      | 11 | 1           |
      | 14 | 1           |
      | 21 | 0           |
      | 22 | 0           |
    And the database has the following table 'groups_ancestors':
      | idGroupAncestor | idGroupChild | bIsSelf |
      | 11              | 11           | 1       |
      | 14              | 14           | 1       |
      | 14              | 21           | 0       |
      | 21              | 21           | 1       |
      | 22              | 22           | 1       |
    And the database has the following table 'groups_groups':
      | ID | idGroupParent | idGroupChild | sType       | sStatusDate         |
      | 7  | 14            | 21           | requestSent | 2017-02-21 06:38:38 |

  Scenario: Successfully send a request
    Given I am the user with ID "1"
    When I send a POST request to "/current-user/group-requests/11"
    Then the response code should be 201
    And the response body should be, in JSON:
    """
    {
      "success": true,
      "message": "created",
      "data": {"changed": true}
    }
    """
    And the table "groups_groups" should be:
      | idGroupParent | idGroupChild | sType       | ABS(TIMESTAMPDIFF(SECOND, sStatusDate, NOW())) < 3 |
      | 11            | 21           | requestSent | 1                                                  |
      | 14            | 21           | requestSent | 0                                                  |
    And the table "groups_ancestors" should stay unchanged

  Scenario: Try to recreate a request that already exists
    Given I am the user with ID "1"
    When I send a POST request to "/current-user/group-requests/14"
    Then the response code should be 201
    And the response body should be, in JSON:
    """
    {
      "success": true,
      "message": "unchanged",
      "data": {"changed": false}
    }
    """
    And the table "groups_groups" should stay unchanged
    And the table "groups_ancestors" should stay unchanged

  Scenario: Automatically accepts the request if the user owns the group
    Given I am the user with ID "1"
    And the database table 'groups_groups' has also the following row:
      | ID | idGroupParent | idGroupChild | sType  | sStatusDate         |
      | 8  | 22            | 11           | direct | 2017-02-21 06:38:38 |
    And the database table 'groups_ancestors' has also the following row:
      | idGroupAncestor | idGroupChild | bIsSelf |
      | 22              | 11           | 0       |
    When I send a POST request to "/current-user/group-requests/11"
    Then the response code should be 201
    And the response body should be, in JSON:
    """
    {
      "success": true,
      "message": "created",
      "data": {"changed": true}
    }
    """
    And the table "groups_groups" should be:
      | idGroupParent | idGroupChild | sType           | ABS(TIMESTAMPDIFF(SECOND, sStatusDate, NOW())) < 3 |
      | 11            | 21           | requestAccepted | 1                                                  |
      | 14            | 21           | requestSent     | 0                                                  |
      | 22            | 11           | direct          | 0                                                  |
    And the table "groups_ancestors" should be:
      | idGroupAncestor | idGroupChild | bIsSelf |
      | 11              | 11           | 1       |
      | 11              | 21           | 0       |
      | 14              | 14           | 1       |
      | 14              | 21           | 0       |
      | 21              | 21           | 1       |
      | 22              | 11           | 0       |
      | 22              | 21           | 0       |
      | 22              | 22           | 1       |
