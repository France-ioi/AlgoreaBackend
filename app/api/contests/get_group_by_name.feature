Feature: Get group by name (contestGetGroupByName)
  Background:
    Given the database has the following table 'users':
      | ID | sLogin | idGroupSelf | idGroupOwned |
      | 1  | owner  | 21          | 22           |
    And the database has the following table 'groups':
      | ID | sName      |
      | 10 | Parent     |
      | 11 | Group A    |
      | 13 | Group B    |
      | 14 | Group B    |
    And the database has the following table 'groups_ancestors':
      | idGroupAncestor | idGroupChild | bIsSelf |
      | 10              | 10           | 1       |
      | 10              | 11           | 0       |
      | 10              | 13           | 0       |
      | 11              | 11           | 1       |
      | 11              | 13           | 0       |
      | 13              | 13           | 1       |
      | 14              | 14           | 1       |
      | 21              | 21           | 1       |
      | 22              | 13           | 0       |
      | 22              | 14           | 0       |
      | 22              | 22           | 1       |
    And the database has the following table 'items':
      | ID | sDuration |
      | 50 | 00:00:00  |
      | 60 | 00:00:01  |
      | 10 | 00:00:02  |
      | 70 | 00:00:03  |
    And the database has the following table 'groups_items':
      | idGroup | idItem | sCachedPartialAccessDate | sCachedGrayedAccessDate | sCachedFullAccessDate | sCachedAccessSolutionsDate | sAdditionalTime     |
      | 10      | 50     | null                     | null                    | null                  | null                       | 0000-00-00T01:00:00 |
      | 11      | 50     | null                     | null                    | null                  | null                       | 0000-00-00T00:01:00 |
      | 13      | 50     | 2017-05-29T06:38:38Z     | null                    | null                  | null                       | 0000-00-00T00:00:01 |
      | 11      | 60     | null                     | null                    | null                  | null                       | null                |
      | 13      | 60     | null                     | 2017-05-29T06:38:38Z    | null                  | null                       | 0000-00-00T00:00:30 |
      | 11      | 70     | null                     | null                    | 2017-05-29T06:38:38Z  | null                       | null                |
      | 21      | 50     | null                     | null                    | null                  | 2018-05-29T06:38:38Z       | 0000-00-00T00:01:00 |
      | 21      | 60     | null                     | null                    | 2018-05-29T06:38:38Z  | null                       | 0000-00-00T00:01:00 |
      | 21      | 70     | null                     | null                    | 2018-05-29T06:38:38Z  | null                       | 0000-00-00T00:01:00 |

  Scenario: Partial access for group, solutions access for user, additional time from parent groups
    Given I am the user with ID "1"
    When I send a GET request to "/contests/50/group-by-name?name=Group%20B"
    Then the response code should be 200
    And the response body should be, in JSON:
    """
    {
      "group_id": "13",
      "additional_time": 1,
      "total_additional_time": 3661
    }
    """

  Scenario: Grayed access for group, full access for user
    Given I am the user with ID "1"
    When I send a GET request to "/contests/60/group-by-name?name=Group%20B"
    Then the response code should be 200
    And the response body should be, in JSON:
    """
    {
      "group_id": "13",
      "additional_time": 30,
      "total_additional_time": 30
    }
    """

  Scenario: Full access for group, full access for user, additional time is null
    Given I am the user with ID "1"
    When I send a GET request to "/contests/70/group-by-name?name=Group%20B"
    Then the response code should be 200
    And the response body should be, in JSON:
    """
    {
      "group_id": "13",
      "additional_time": 0,
      "total_additional_time": 0
    }
    """

  Scenario: Should ignore case
    Given I am the user with ID "1"
    When I send a GET request to "/contests/50/group-by-name?name=group%20b"
    Then the response code should be 200
    And the response body should be, in JSON:
    """
    {
      "group_id": "13",
      "additional_time": 1,
      "total_additional_time": 3661
    }
    """
