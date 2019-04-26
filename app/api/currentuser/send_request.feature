Feature: User sends a request to join a group
  Background:
    Given the database has the following table 'users':
      | ID | idGroupSelf | idGroupOwned |
      | 1  | 21          | 22           |
    And the database has the following table 'groups':
      | ID |
      | 11 |
      | 14 |
      | 21 |
      | 22 |
    And the database has the following table 'groups_ancestors':
      | idGroupAncestor | idGroupChild | bIsSelf |
      | 11              | 11           | 1       |
      | 14              | 14           | 1       |
      | 14              | 21           | 0       |
      | 21              | 21           | 1       |
      | 22              | 22           | 1       |
    And the database has the following table 'groups_groups':
      | ID | idGroupParent | idGroupChild | sType       | sStatusDate          |
      | 7  | 14            | 21           | requestSent | 2017-02-21T06:38:38Z |

  Scenario: Successfully send a request
    Given I am the user with ID "1"
    When I send a POST request to "/current-user/requests/11"
    Then the response code should be 201
    And the response body should be, in JSON:
    """
    {
      "success": true,
      "message": "created"
    }
    """
    And the table "groups_groups" should be:
      | idGroupParent | idGroupChild | sType       | (sStatusDate IS NOT NULL) AND (NOW() - sStatusDate < 3) |
      | 11            | 21           | requestSent | 1                                                       |
      | 14            | 21           | requestSent | 0                                                       |
    And the table "groups_ancestors" should stay unchanged

  Scenario: Try to recreate a request that already exists
    Given I am the user with ID "1"
    When I send a POST request to "/current-user/requests/14"
    Then the response code should be 205
    And the response body should be, in JSON:
    """
    {
      "success": true,
      "message": "not changed"
    }
    """
    And the table "groups_groups" should stay unchanged
    And the table "groups_ancestors" should stay unchanged
