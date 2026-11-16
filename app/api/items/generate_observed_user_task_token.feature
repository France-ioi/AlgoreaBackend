Feature: Generate a read-only task token for an observed user's attempt
  Background:
    Given the database has the following table "groups":
      | id  | name  | type  |
      | 102 | team  | Team  |
      | 104 | class | Class |
    And the database has the following users:
      | group_id | login    |
      | 101      | john     |
      | 103      | manager  |
      | 105      | stranger |
    And the database has the following table "groups_groups":
      | parent_group_id | child_group_id |
      | 104             | 101            |
      | 104             | 102            |
    And the database has the following table "group_managers":
      | group_id | manager_id | can_watch_members |
      | 104      | 103        | 1                 |
    And the groups ancestors are computed
    And the database has the following table "items":
      | id | url                    | type    | default_language_tag | text_id | hints_allowed | supported_lang_prog |
      | 50 | http://taskplatform/50 | Task    | fr                   | task50  | 1             | c,python            |
      | 60 | http://taskplatform/60 | Task    | fr                   | task60  | 0             | null                |
      | 70 | http://taskplatform/70 | Task    | fr                   | task70  | 0             | null                |
      | 80 | http://taskplatform/80 | Task    | fr                   | task80  | 0             | null                |
      | 90 | http://taskplatform/90 | Chapter | fr                   | null    | 0             | null                |
    And the database has the following table "permissions_generated":
      | group_id | item_id | can_view_generated | can_watch_generated |
      | 103      | 50      | content            | answer              |
      | 103      | 60      | content            | result              |
      | 103      | 70      | solution           | answer              |
      | 103      | 80      | content            | answer              |
      | 103      | 90      | content            | answer              |
      | 101      | 50      | content            | answer              |
      | 105      | 50      | content            | answer              |
    And the database has the following table "attempts":
      | id | participant_id | allows_submissions_until |
      | 0  | 103            | 9999-12-31 23:59:59      |
      | 1  | 101            | 2019-05-30 11:00:00      |
      | 1  | 102            | 9999-12-31 23:59:59      |
      | 2  | 101            | 9999-12-31 23:59:59      |
    And the database has the following table "results":
      | attempt_id | participant_id | item_id | latest_activity_at  | started_at          | hints_requested | hints_cached | validated_at        |
      | 1          | 101            | 50      | 2018-05-29 06:38:38 | 2018-05-29 05:00:00 | [1,2,3,4]       | 4            | null                |
      | 1          | 101            | 60      | 2018-05-29 06:38:38 | 2018-05-29 05:00:00 | null            | 0            | null                |
      | 1          | 101            | 70      | 2018-05-29 06:38:38 | 2018-05-29 05:00:00 | null            | 0            | null                |
      | 1          | 101            | 80      | 2018-05-29 06:38:38 | 2018-05-29 05:00:00 | null            | 0            | null                |
      | 1          | 101            | 90      | 2018-05-29 06:38:38 | 2018-05-29 05:00:00 | null            | 0            | null                |
      | 1          | 102            | 50      | 2018-05-29 06:38:38 | 2018-05-29 05:00:00 | null            | 0            | null                |
      | 2          | 101            | 50      | 2018-05-29 06:38:38 | null                | null            | 0            | null                |
      | 0          | 103            | 80      | 2018-05-29 06:38:38 | 2018-05-29 05:00:00 | null            | 0            | 2018-05-30 05:00:00 |
    And the server time is frozen

  Scenario: An observer who may watch the user's answers gets a read-only token for that attempt
    Given I am the user with id "103"
    When I send a POST request to "/items/50/participant/101/attempts/1/generate-task-token"
    Then the response code should be 200
    And the response body decoded as "GenerateTaskTokenResponse" should be, in JSON:
      """
      {
        "data": {
          "task_token": {
            "date": "{{currentTimeInFormat("02-01-2006")}}",
            "bAccessSolutions": false,
            "bHintsAllowed": false,
            "bIsAdmin": false,
            "bReadAnswers": true,
            "bSubmissionPossible": false,
            "idAttempt": "101/1",
            "idUser": "101",
            "idItemLocal": "50",
            "idItem": "task50",
            "itemUrl": "http://taskplatform/50",
            "nbHintsGiven": "4",
            "sHintsRequested": "[1,2,3,4]",
            "sSupportedLangProg": "c,python",
            "randomSeed": "12601247502642542026",
            "platformName": "{{app().Config.GetString("token.platformName")}}",
            "sLogin": "john"
          }
        },
        "message": "created",
        "success": true
      }
      """
    And the table "results" should remain unchanged
    And the table "results_propagate" should be empty
    And the table "results_propagate_internal" should be empty

  Scenario: bAccessSolutions is true when the caller has can_view >= solution
    Given I am the user with id "103"
    When I send a POST request to "/items/70/participant/101/attempts/1/generate-task-token"
    Then the response code should be 200
    And the response body decoded as "GenerateTaskTokenResponse" should be, in JSON:
      """
      {
        "data": {
          "task_token": {
            "date": "{{currentTimeInFormat("02-01-2006")}}",
            "bAccessSolutions": true,
            "bHintsAllowed": false,
            "bIsAdmin": false,
            "bReadAnswers": true,
            "bSubmissionPossible": false,
            "idAttempt": "101/1",
            "idUser": "101",
            "idItemLocal": "70",
            "idItem": "task70",
            "itemUrl": "http://taskplatform/70",
            "nbHintsGiven": "0",
            "randomSeed": "12601247502642542026",
            "platformName": "{{app().Config.GetString("token.platformName")}}",
            "sLogin": "john"
          }
        },
        "message": "created",
        "success": true
      }
      """

  Scenario: bAccessSolutions is true when the caller has validated the item
    Given I am the user with id "103"
    When I send a POST request to "/items/80/participant/101/attempts/1/generate-task-token"
    Then the response code should be 200
    And the response body decoded as "GenerateTaskTokenResponse" should be, in JSON:
      """
      {
        "data": {
          "task_token": {
            "date": "{{currentTimeInFormat("02-01-2006")}}",
            "bAccessSolutions": true,
            "bHintsAllowed": false,
            "bIsAdmin": false,
            "bReadAnswers": true,
            "bSubmissionPossible": false,
            "idAttempt": "101/1",
            "idUser": "101",
            "idItemLocal": "80",
            "idItem": "task80",
            "itemUrl": "http://taskplatform/80",
            "nbHintsGiven": "0",
            "randomSeed": "12601247502642542026",
            "platformName": "{{app().Config.GetString("token.platformName")}}",
            "sLogin": "john"
          }
        },
        "message": "created",
        "success": true
      }
      """
