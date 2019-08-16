Feature: Get the contests that the user has administration rights on (contestAdminList)
  Background:
    Given the database has the following table 'users':
      | ID | sLogin         | idGroupSelf | idGroupOwned | sDefaultLanguage |
      | 1  | possesseur     | 21          | 22           | fr               |
      | 2  | owner          | 31          | 32           | en               |
      | 3  | administrateur | 41          | 42           | fr               |
      | 4  | admin          | 51          | 52           | en               |
      | 5  | guest          | 61          | 62           | en               |
    And the database has the following table 'languages':
      | ID | sCode |
      | 1  | en    |
      | 2  | fr    |
    And the database has the following table 'groups_ancestors':
      | idGroupAncestor | idGroupChild | bIsSelf |
      | 21              | 21           | 1       |
      | 22              | 13           | 0       |
      | 22              | 14           | 0       |
      | 22              | 22           | 1       |
      | 31              | 31           | 1       |
      | 32              | 32           | 1       |
      | 41              | 41           | 1       |
      | 42              | 42           | 1       |
      | 51              | 51           | 1       |
      | 52              | 52           | 1       |
      | 61              | 61           | 1       |
      | 62              | 62           | 1       |
    And the database has the following table 'items':
      | ID | sDuration | idDefaultLanguage | bHasAttempts |
      | 50 | 00:00:00  | 2                 | 0            |
      | 60 | 00:00:01  | 1                 | 1            |
      | 10 | 00:00:02  | 1                 | 0            |
      | 70 | 00:00:03  | 2                 | 0            |
    And the database has the following table 'items_items':
      | idItemParent | idItemChild |
      | 10           | 60          |
      | 10           | 70          |
      | 60           | 70          |
    And the database has the following table 'items_strings':
      | idItem | idLanguage | sTitle     |
      | 10     | 1          | Chapter    |
      | 10     | 2          | Chapitre   |
      | 60     | 1          | Contest    |
      | 70     | 1          | Contest 2  |
      | 70     | 2          | Concours 2 |
    And the database has the following table 'groups_items':
      | idGroup | idItem | sCachedPartialAccessDate | sCachedGrayedAccessDate | sCachedFullAccessDate | sCachedAccessSolutionsDate |
      | 21      | 50     | null                     | null                    | null                  | 2018-05-29T06:38:38Z       |
      | 21      | 60     | null                     | null                    | 2018-05-29T06:38:38Z  | null                       |
      | 21      | 70     | null                     | null                    | 2018-05-29T06:38:38Z  | null                       |
      | 31      | 50     | null                     | null                    | null                  | 2018-05-29T06:38:38Z       |
      | 31      | 60     | null                     | null                    | 2018-05-29T06:38:38Z  | null                       |
      | 31      | 70     | null                     | null                    | 2018-05-29T06:38:38Z  | null                       |
      | 41      | 10     | 2018-05-29T06:38:38Z     | null                    | null                  | null                       |
      | 41      | 50     | null                     | null                    | null                  | 2018-05-29T06:38:38Z       |
      | 41      | 60     | null                     | 2018-05-29T06:38:38Z    | null                  | 2018-05-29T06:38:38Z       |
      | 41      | 70     | null                     | null                    | 2018-05-29T06:38:38Z  | null                       |
      | 51      | 10     | null                     | 2018-05-29T06:38:38Z    | null                  | null                       |
      | 51      | 50     | null                     | null                    | null                  | 2018-05-29T06:38:38Z       |
      | 51      | 60     | null                     | null                    | 2018-05-29T06:38:38Z  | 2018-05-29T06:38:38Z       |
      | 51      | 70     | null                     | null                    | 2018-05-29T06:38:38Z  | null                       |

  Scenario: User's default language is French (most parents are invisible)
    Given I am the user with ID "1"
    When I send a GET request to "/contests/administered"
    Then the response code should be 200
    And the response body should be, in JSON:
    """
    [
      {"id": "50", "team_only_contest": false, "parents": [], "title": null},
      {"id": "70", "team_only_contest": false, "parents": [{"title": "Contest"}], "title": "Concours 2"},
      {"id": "60", "team_only_contest": true, "parents": [], "title": "Contest"}
    ]
    """

  Scenario: User's default language is English  (most parents are invisible)
    Given I am the user with ID "2"
    When I send a GET request to "/contests/administered"
    Then the response code should be 200
    And the response body should be, in JSON:
    """
    [
      {"id": "50", "team_only_contest": false, "parents": [], "title": null},
      {"id": "60", "team_only_contest": true, "parents": [], "title": "Contest"},
      {"id": "70", "team_only_contest": false, "parents": [{"title": "Contest"}], "title": "Contest 2"}
    ]
    """

  Scenario: User's default language is French (parents are visible)
    Given I am the user with ID "3"
    When I send a GET request to "/contests/administered"
    Then the response code should be 200
    And the response body should be, in JSON:
    """
    [
      {"id": "50", "team_only_contest": false, "parents": [], "title": null},
      {"id": "70", "team_only_contest": false, "parents": [{"title": "Chapitre"}, {"title": "Contest"}], "title": "Concours 2"},
      {"id": "60", "team_only_contest": true, "parents": [{"title": "Chapitre"}], "title": "Contest"}
    ]
    """

  Scenario: User's default language is English  (parents are visible)
    Given I am the user with ID "4"
    When I send a GET request to "/contests/administered"
    Then the response code should be 200
    And the response body should be, in JSON:
    """
    [
      {"id": "50", "team_only_contest": false, "parents": [], "title": null},
      {"id": "60", "team_only_contest": true, "parents": [{"title": "Chapter"}], "title": "Contest"},
      {"id": "70", "team_only_contest": false, "parents": [{"title": "Chapter"}, {"title": "Contest"}], "title": "Contest 2"}
    ]
    """

  Scenario: Empty result
    Given I am the user with ID "5"
    When I send a GET request to "/contests/administered"
    Then the response code should be 200
    And the response body should be, in JSON:
    """
    [
    ]
    """

  Scenario: User's default language is English  (parents are visible), limit=1
    Given I am the user with ID "4"
    When I send a GET request to "/contests/administered?limit=1"
    Then the response code should be 200
    And the response body should be, in JSON:
    """
    [
      {"id": "50", "team_only_contest": false, "parents": [], "title": null}
    ]
    """

  Scenario: User's default language is English  (parents are visible), start from the second row, limit=1
    Given I am the user with ID "4"
    When I send a GET request to "/contests/administered?from.title&from.id=50&limit=1"
    Then the response code should be 200
    And the response body should be, in JSON:
    """
    [
      {"id": "60", "team_only_contest": true, "parents": [{"title": "Chapter"}], "title": "Contest"}
    ]
    """

  Scenario: User's default language is English  (parents are visible), inverse order
    Given I am the user with ID "4"
    When I send a GET request to "/contests/administered?sort=-title,id"
    Then the response code should be 200
    And the response body should be, in JSON:
    """
    [
      {"id": "70", "team_only_contest": false, "parents": [{"title": "Chapter"}, {"title": "Contest"}], "title": "Contest 2"},
      {"id": "60", "team_only_contest": true, "parents": [{"title": "Chapter"}], "title": "Contest"},
      {"id": "50", "team_only_contest": false, "parents": [], "title": null}
    ]
    """
