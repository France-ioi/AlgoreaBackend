Feature: Get group by name (contestGetGroupByName) - robustness
  Background:
    Given the database has the following table 'users':
      | ID | sLogin | idGroupSelf | idGroupOwned |
      | 1  | owner  | 21          | 22           |
    And the database has the following table 'groups':
      | ID | sName      |
      | 12 | Group A    |
      | 13 | Group B    |
    And the database has the following table 'groups_ancestors':
      | idGroupAncestor | idGroupChild | bIsSelf |
      | 12              | 12           | 1       |
      | 13              | 13           | 1       |
      | 21              | 21           | 1       |
      | 22              | 13           | 0       |
      | 22              | 22           | 1       |
    And the database has the following table 'items':
      | ID | sDuration |
      | 50 | 00:00:00  |
      | 60 | null      |
      | 10 | 00:00:02  |
      | 70 | 00:00:03  |
    And the database has the following table 'groups_items':
      | idGroup | idItem | sCachedPartialAccessDate | sCachedGrayedAccessDate | sCachedFullAccessDate | sCachedAccessSolutionsDate | idUserCreated |
      | 13      | 50     | 2017-05-29 06:38:38      | null                    | null                  | null                       | 1             |
      | 13      | 60     | null                     | 2017-05-29 06:38:38     | null                  | null                       | 1             |
      | 13      | 70     | null                     | null                    | 2017-05-29 06:38:38   | null                       | 1             |
      | 21      | 50     | null                     | null                    | null                  | null                       | 1             |
      | 21      | 60     | null                     | null                    | 2018-05-29 06:38:38   | null                       | 1             |
      | 21      | 70     | null                     | null                    | 2018-05-29 06:38:38   | null                       | 1             |

  Scenario: Wrong item_id
    Given I am the user with ID "1"
    When I send a GET request to "/contests/abc/groups/by-name?name=Group%20B"
    Then the response code should be 400
    And the response error message should contain "Wrong value for item_id (should be int64)"

  Scenario: name is missing
    Given I am the user with ID "1"
    When I send a GET request to "/contests/50/groups/by-name"
    Then the response code should be 400
    And the response error message should contain "Missing name"

  Scenario: No such item
    Given I am the user with ID "1"
    When I send a GET request to "/contests/90/groups/by-name?name=Group%20B"
    Then the response code should be 403
    And the response error message should contain "Insufficient access rights"

  Scenario: No access to the item
    Given I am the user with ID "1"
    When I send a GET request to "/contests/10/groups/by-name?name=Group%20B"
    Then the response code should be 403
    And the response error message should contain "Insufficient access rights"

  Scenario: The item is not a timed contest
    Given I am the user with ID "1"
    When I send a GET request to "/contests/60/groups/by-name?name=Group%20B"
    Then the response code should be 403
    And the response error message should contain "Insufficient access rights"

  Scenario: The user is not a contest admin
    Given I am the user with ID "1"
    When I send a GET request to "/contests/50/groups/by-name?name=Group%20B"
    Then the response code should be 403
    And the response error message should contain "Insufficient access rights"

  Scenario: The group is not owned by the user
    Given I am the user with ID "1"
    When I send a GET request to "/contests/70/groups/by-name?name=Group%20A"
    Then the response code should be 403
    And the response error message should contain "Insufficient access rights"

  Scenario: No such group (space)
    Given I am the user with ID "1"
    When I send a GET request to "/contests/70/groups/by-name?name=Group%20B%20"
    Then the response code should be 403
    And the response error message should contain "Insufficient access rights"
