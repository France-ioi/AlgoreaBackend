Feature: Set additional time in the contest for the group (contestSetAdditionalTime)
  Background:
    Given the database has the following table 'users':
      | ID | sLogin | idGroupSelf | idGroupOwned |
      | 1  | owner  | 21          | 22           |
      | 2  | john   | 31          | 32           |
    And the database has the following table 'groups':
      | ID | sName       | sType     |
      | 10 | Parent      | Club      |
      | 11 | Group A     | Class     |
      | 13 | Group B     | Other     |
      | 14 | Group B     | Friends   |
      | 21 | owner       | UserSelf  |
      | 22 | owner-admin | UserAdmin |
      | 31 | john        | UserSelf  |
      | 32 | john-admin  | UserAdmin |
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
      | 22              | 31           | 0       |
      | 22              | 22           | 1       |
      | 31              | 31           | 1       |
      | 32              | 32           | 1       |
    And the database has the following table 'items':
      | ID | sDuration |
      | 50 | 00:00:00  |
      | 60 | 00:00:01  |
      | 10 | 00:00:02  |
      | 70 | 00:00:03  |
    And the database has the following table 'groups_items':
      | ID | idGroup | idItem | sCachedPartialAccessDate | sCachedGrayedAccessDate | sCachedFullAccessDate | sCachedAccessSolutionsDate | sAdditionalTime | idUserCreated |
      | 1  | 10      | 50     | null                     | null                    | null                  | null                       | 01:00:00        | 3             |
      | 2  | 11      | 50     | null                     | null                    | null                  | null                       | 00:01:00        | 3             |
      | 3  | 13      | 50     | 2017-05-29 06:38:38      | null                    | null                  | null                       | 00:00:01        | 3             |
      | 4  | 11      | 60     | null                     | null                    | null                  | null                       | null            | 3             |
      | 5  | 13      | 60     | null                     | 2017-05-29 06:38:38     | null                  | null                       | 00:00:30        | 3             |
      | 6  | 11      | 70     | null                     | null                    | 2017-05-29 06:38:38   | null                       | null            | 3             |
      | 7  | 21      | 50     | null                     | null                    | null                  | 2018-05-29 06:38:38        | 00:01:00        | 3             |
      | 8  | 21      | 60     | null                     | null                    | 2018-05-29 06:38:38   | null                       | 00:01:00        | 3             |
      | 9  | 21      | 70     | null                     | null                    | 2018-05-29 06:38:38   | null                       | 00:01:00        | 3             |

  Scenario: Updates an existing row
    Given I am the user with ID "1"
    When I send a PUT request to "/contests/50/groups/13/additional-times?seconds=3020399"
    Then the response code should be 200
    And the response should be "updated"
    And the table "groups_items" should stay unchanged but the row with ID "3"
    And the table "groups_items" at ID "3" should be:
      | ID | idGroup | idItem | sCachedPartialAccessDate | sCachedGrayedAccessDate | sCachedFullAccessDate | sCachedAccessSolutionsDate | sAdditionalTime | idUserCreated |
      | 3  | 13      | 50     | 2017-05-29 06:38:38      | null                    | null                  | null                       | 838:59:59       | 3             |

  Scenario: Creates a new row
    Given I am the user with ID "1"
    When I send a PUT request to "/contests/70/groups/13/additional-times?seconds=-3020399"
    Then the response code should be 200
    And the response should be "updated"
    And the table "groups_items" should stay unchanged but the row with ID "5577006791947779410"
    And the table "groups_items" at ID "5577006791947779410" should be:
      | ID                  | idGroup | idItem | sCachedPartialAccessDate | sCachedGrayedAccessDate | sCachedFullAccessDate | sCachedAccessSolutionsDate | sAdditionalTime | idUserCreated |
      | 5577006791947779410 | 13      | 70     | null                     | null                    | null                  | null                       | -838:59:59      | 1             |

  Scenario: Doesn't create a new row when seconds=0
    Given I am the user with ID "1"
    When I send a PUT request to "/contests/70/groups/13/additional-times?seconds=0"
    Then the response code should be 200
    And the response should be "updated"
    And the table "groups_items" should stay unchanged

  Scenario: Creates a new row for a user group
    Given I am the user with ID "1"
    When I send a PUT request to "/contests/70/groups/31/additional-times?seconds=-3020399"
    Then the response code should be 200
    And the response should be "updated"
    And the table "groups_items" should stay unchanged but the row with ID "5577006791947779410"
    And the table "groups_items" at ID "5577006791947779410" should be:
      | ID                  | idGroup | idItem | sCachedPartialAccessDate | sCachedGrayedAccessDate | sCachedFullAccessDate | sCachedAccessSolutionsDate | sAdditionalTime | idUserCreated |
      | 5577006791947779410 | 31      | 70     | null                     | null                    | null                  | null                       | -838:59:59      | 1             |
