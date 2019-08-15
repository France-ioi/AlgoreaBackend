Feature: Set additional time in the contest for the group (contestSetAdditionalTime) - robustness
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
      | idGroup | idItem | sCachedPartialAccessDate | sCachedGrayedAccessDate | sCachedFullAccessDate | sCachedAccessSolutionsDate | sAdditionalTime |
      | 13      | 50     | 2017-05-29T06:38:38Z     | null                    | null                  | null                       | 01:00:00        |
      | 13      | 60     | null                     | 2017-05-29T06:38:38Z    | null                  | null                       | 01:01:00        |
      | 13      | 70     | null                     | null                    | 2017-05-29T06:38:38Z  | null                       | null            |
      | 21      | 50     | null                     | null                    | null                  | null                       | null            |
      | 21      | 60     | null                     | null                    | 2018-05-29T06:38:38Z  | null                       | null            |
      | 21      | 70     | null                     | null                    | 2018-05-29T06:38:38Z  | null                       | null            |

  Scenario: Wrong item_id
    Given I am the user with ID "1"
    When I send a PUT request to "/contests/abc/additional-time?group_id=13&time=0"
    Then the response code should be 400
    And the response error message should contain "Wrong value for item_id (should be int64)"

  Scenario: Wrong group_id
    Given I am the user with ID "1"
    When I send a PUT request to "/contests/50/additional-time?group_id=abc&time=0"
    Then the response code should be 400
    And the response error message should contain "Wrong value for group_id (should be int64)"

  Scenario: Wrong time
    Given I am the user with ID "1"
    When I send a PUT request to "/contests/50/additional-time?group_id=13&time=abc"
    Then the response code should be 400
    And the response error message should contain "Wrong value for time (should be int64)"

  Scenario: Time is too big
    Given I am the user with ID "1"
    When I send a PUT request to "/contests/50/additional-time?group_id=13&time=3020400"
    Then the response code should be 400
    And the response error message should contain "'time' should be between -3020399 and 3020399"

  Scenario: Time is too small
    Given I am the user with ID "1"
    When I send a PUT request to "/contests/50/additional-time?group_id=13&time=-3020400"
    Then the response code should be 400
    And the response error message should contain "'time' should be between -3020399 and 3020399"

  Scenario: No such item
    Given I am the user with ID "1"
    When I send a PUT request to "/contests/90/additional-time?group_id=13&time=0"
    Then the response code should be 403
    And the response error message should contain "Insufficient access rights"

  Scenario: No access to the item
    Given I am the user with ID "1"
    When I send a PUT request to "/contests/10/additional-time?group_id=13&time=0"
    Then the response code should be 403
    And the response error message should contain "Insufficient access rights"

  Scenario: The item is not a timed contest
    Given I am the user with ID "1"
    When I send a PUT request to "/contests/60/additional-time?group_id=13&time=0"
    Then the response code should be 403
    And the response error message should contain "Insufficient access rights"

  Scenario: The user is not a contest admin
    Given I am the user with ID "1"
    When I send a PUT request to "/contests/50/additional-time?group_id=13&time=0"
    Then the response code should be 403
    And the response error message should contain "Insufficient access rights"

  Scenario: The group is not owned by the user
    Given I am the user with ID "1"
    When I send a PUT request to "/contests/70/additional-time?group_id=12&time=0"
    Then the response code should be 403
    And the response error message should contain "Insufficient access rights"

  Scenario: No such group
    Given I am the user with ID "1"
    When I send a PUT request to "/contests/70/additional-time?group_id=404&time=0"
    Then the response code should be 403
    And the response error message should contain "Insufficient access rights"
