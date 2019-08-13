Feature: Get group by name (contestGetGroupByName)
  Background:
    Given the database has the following table 'users':
      | ID | sLogin | idGroupSelf | idGroupOwned |
      | 1  | owner  | 21          | 22           |
    And the database has the following table 'groups':
      | ID | sName      |
      | 11 | Group A    |
      | 13 | Group B    |
      | 14 | Group B    |
    And the database has the following table 'groups_ancestors':
      | idGroupAncestor | idGroupChild | bIsSelf |
      | 11              | 11           | 1       |
      | 11              | 13           | 0       |
      | 13              | 13           | 1       |
      | 14              | 14           | 1       |
      | 21              | 21           | 1       |
      | 22              | 13           | 0       |
      | 22              | 14           | 0       |
      | 22              | 22           | 1       |
    And the database has the following table 'items':
      | ID |
      | 50 |
      | 60 |
      | 10 |
      | 70 |
    And the database has the following table 'groups_items':
      | idGroup | idItem | sCachedPartialAccessDate | sCachedGrayedAccessDate | sCachedFullAccessDate |
      | 13      | 50     | 2017-05-29T06:38:38Z     | null                    | null                  |
      | 13      | 60     | null                     | 2017-05-29T06:38:38Z    | null                  |
      | 11      | 70     | null                     | null                    | 2017-05-29T06:38:38Z  |

  Scenario: Partial access
    Given I am the user with ID "1"
    When I send a GET request to "/contests/50/group-by-name?name=Group%20B"
    Then the response code should be 200
    And the response body should be, in JSON:
    """
    {
      "group_id": "13"
    }
    """

  Scenario: Grayed access
    Given I am the user with ID "1"
    When I send a GET request to "/contests/60/group-by-name?name=Group%20B"
    Then the response code should be 200
    And the response body should be, in JSON:
    """
    {
      "group_id": "13"
    }
    """

  Scenario: Full access
    Given I am the user with ID "1"
    When I send a GET request to "/contests/70/group-by-name?name=Group%20B"
    Then the response code should be 200
    And the response body should be, in JSON:
    """
    {
      "group_id": "13"
    }
    """
