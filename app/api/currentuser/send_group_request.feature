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
    And the database has the following table 'groups_groups':
      | ID | idGroupParent | idGroupChild | sType       | sStatusDate          |
      | 7  | 14            | 21           | requestSent | 2017-02-21T06:38:38Z |

  Scenario: Successfully send a request
    Given I am the user with ID "1"
    When I send a POST request to "/current-user/group-requests/11"
    Then the response code should be 201
    And the response body should be, in JSON:
    """
    {
      "success": true,
      "message": "created"
    }
    """
    And the table "groups_groups" should be:
      | idGroupParent | idGroupChild | sType       | ABS(TIMESTAMPDIFF(SECOND, sStatusDate, NOW())) < 3 |
      | 11            | 21           | requestSent | 1                                                  |
      | 14            | 21           | requestSent | 0                                                  |

  Scenario: Try to recreate a request that already exists
    Given I am the user with ID "1"
    When I send a POST request to "/current-user/group-requests/14"
    Then the response code should be 205
    And the response body should be, in JSON:
    """
    {
      "success": true,
      "message": "not changed"
    }
    """
    And the table "groups_groups" should stay unchanged
