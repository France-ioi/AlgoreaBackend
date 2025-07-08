Feature: Generate a read-only task token for an item from an answer
  Background:
    Given the database has the following table "groups":
      | id  | name     | type  |
      | 102 | team     | Team  |
      | 104 | Groupe A | Class |
    And the database has the following users:
      | group_id | login   |
      | 101      | john    |
      | 103      | manager |
      | 105      | jack    |
    And the database has the following table "groups_groups":
      | parent_group_id | child_group_id |
      | 102             | 101            |
      | 102             | 105            |
      | 104             | 101            |
    And the database has the following table "group_managers":
      | group_id | manager_id | can_watch_members |
      | 104      | 103        | 1                 |
    And the groups ancestors are computed
    And the database has the following table "items":
      | id | url                    | type | entry_participant_type | default_language_tag | text_id |
      | 10 | http://taskplatform/10 | Task | User                   | fr                   | task10  |
      | 20 | http://taskplatform/20 | Task | Team                   | fr                   | task20  |
      | 30 | http://taskplatform/30 | Task | User                   | fr                   | task30  |
      | 40 | http://taskplatform/40 | Task | User                   | fr                   | task40  |
      | 50 | http://taskplatform/50 | Task | User                   | fr                   | task50  |
      | 60 | http://taskplatform/60 | Task | User                   | fr                   | task60  |
    And the database has the following table "permissions_generated":
      | group_id | item_id | can_view_generated | can_watch_generated |
      | 101      | 10      | content            | none                |
      | 103      | 10      | content            | answer              |
      | 102      | 20      | content            | none                |
      | 104      | 20      | content            | answer              |
      | 103      | 30      | content            | answer              |
      | 104      | 30      | content            | answer              |
      | 101      | 40      | solution           | none                |
      | 102      | 40      | content            | none                |
      | 104      | 40      | content            | answer              |
      | 101      | 50      | content            | none                |
      | 105      | 50      | content            | answer              |
      | 102      | 60      | content            | none                |
      | 104      | 60      | content            | answer              |
    And the database has the following table "attempts":
      | id | participant_id |
      | 0  | 101            |
      | 0  | 102            |
      | 1  | 102            |
      | 0  | 103            |
      | 1  | 103            |
    And the database has the following table "results":
      | attempt_id | participant_id | item_id | started_at          | validated_at        | hints_requested | hints_cached |
      | 0          | 101            | 10      | 2020-01-01 01:01:01 | null                | null            | 0            |
      | 0          | 103            | 10      | 2020-01-01 01:01:01 | null                | null            | 0            |
      | 0          | 101            | 20      | 2020-01-01 01:01:01 | null                | null            | 0            |
      | 0          | 102            | 20      | 2020-01-01 01:01:01 | null                | null            | 0            |
      | 0          | 101            | 30      | 2020-01-01 01:01:01 | null                | null            | 0            |
      | 0          | 103            | 30      | null                | null                | null            | 0            |
      | 1          | 103            | 30      | 2020-01-01 01:01:01 | null                | null            | 0            |
      | 0          | 101            | 40      | 2020-01-01 01:01:01 | null                | null            | 0            |
      | 0          | 102            | 40      | 2020-01-01 01:01:01 | null                | null            | 0            |
      | 0          | 101            | 50      | 2020-01-01 01:01:01 | 2020-01-01 01:01:01 | null            | 0            |
      | 0          | 102            | 50      | 2020-01-01 01:01:01 | null                | null            | 0            |
      | 0          | 101            | 60      | 2020-01-01 01:01:01 | null                | null            | 0            |
      | 1          | 102            | 60      | 2020-01-01 01:01:01 | null                | [1,2,3,4]       | 4            |
    And the database has the following table "answers":
      | id | participant_id | attempt_id | item_id | author_id | created_at          |
      | 1  | 101            | 0          | 10      | 101       | 2020-01-01 01:01:01 |
      | 2  | 102            | 0          | 20      | 105       | 2020-01-01 01:01:01 |
      | 3  | 101            | 0          | 30      | 101       | 2020-01-01 01:01:01 |
      | 4  | 102            | 0          | 40      | 105       | 2020-01-01 01:01:01 |
      | 5  | 102            | 0          | 50      | 105       | 2020-01-01 01:01:01 |
      | 6  | 102            | 1          | 60      | 105       | 2020-01-01 01:01:01 |
    And the server time is frozen

  Scenario: User is able to fetch a task token when participant is the current user
    Given I am the user with id "101"
    When I send a POST request to "/answers/1/generate-task-token"
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
            "idAttempt": "101/0",
            "idUser": "101",
            "idItemLocal": "10",
            "idItem": "task10",
            "itemUrl": "http://taskplatform/10",
            "nbHintsGiven": "0",
            "randomSeed": "2147886519731235493",
            "platformName": "{{app().Config.GetString("token.platformName")}}",
            "sLogin": "john"
          }
        },
        "message": "created",
        "success": true
      }
      """

  Scenario: User is able to fetch a task token when participant is a team which the current user is member of
    Given I am the user with id "101"
    When I send a POST request to "/answers/2/generate-task-token"
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
            "idAttempt": "102/0",
            "idUser": "105",
            "idItemLocal": "20",
            "idItem": "task20",
            "itemUrl": "http://taskplatform/20",
            "nbHintsGiven": "0",
            "randomSeed": "6688673066740915448",
            "platformName": "{{app().Config.GetString("token.platformName")}}",
            "sLogin": "jack"
          }
        },
        "message": "created",
        "success": true
      }
      """

  Scenario: User should see the hints infos of the answer's result
    Given I am the user with id "101"
    When I send a POST request to "/answers/6/generate-task-token"
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
            "idAttempt": "102/1",
            "idUser": "105",
            "idItemLocal": "60",
            "idItem": "task60",
            "itemUrl": "http://taskplatform/60",
            "nbHintsGiven": "4",
            "sHintsRequested": "[1,2,3,4]",
            "randomSeed": "17292903417420170135",
            "platformName": "{{app().Config.GetString("token.platformName")}}",
            "sLogin": "jack"
          }
        },
        "message": "created",
        "success": true
      }
      """

  Scenario: User is able to fetch a task token when participant is watched
    Given I am the user with id "103"
    When I send a POST request to "/answers/1/generate-task-token"
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
            "idAttempt": "101/0",
            "idUser": "101",
            "idItemLocal": "10",
            "idItem": "task10",
            "itemUrl": "http://taskplatform/10",
            "nbHintsGiven": "0",
            "randomSeed": "2147886519731235493",
            "platformName": "{{app().Config.GetString("token.platformName")}}",
            "sLogin": "john"
          }
        },
        "message": "created",
        "success": true
      }
      """

  Scenario: User is able to fetch a task token when participant is watched when having a started result on an other attempt
    Given I am the user with id "103"
    When I send a POST request to "/answers/3/generate-task-token"
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
            "idAttempt": "101/0",
            "idUser": "101",
            "idItemLocal": "30",
            "idItem": "task30",
            "itemUrl": "http://taskplatform/30",
            "nbHintsGiven": "0",
            "randomSeed": "2147886519731235493",
            "platformName": "{{app().Config.GetString("token.platformName")}}",
            "sLogin": "john"
          }
        },
        "message": "created",
        "success": true
      }
      """

  Scenario: bAccessSolutions is true when the requester has `can_view` >= 'solution' on the item
    Given I am the user with id "101"
    When I send a POST request to "/answers/4/generate-task-token"
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
            "idAttempt": "102/0",
            "idUser": "105",
            "idItemLocal": "40",
            "idItem": "task40",
            "itemUrl": "http://taskplatform/40",
            "nbHintsGiven": "0",
            "randomSeed": "6688673066740915448",
            "platformName": "{{app().Config.GetString("token.platformName")}}",
            "sLogin": "jack"
          }
        },
        "message": "created",
        "success": true
      }
      """

  Scenario: bAccessSolutions is true when the requester has validated the item
    Given I am the user with id "101"
    When I send a POST request to "/answers/5/generate-task-token"
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
            "idAttempt": "102/0",
            "idUser": "105",
            "idItemLocal": "50",
            "idItem": "task50",
            "itemUrl": "http://taskplatform/50",
            "nbHintsGiven": "0",
            "randomSeed": "6688673066740915448",
            "platformName": "{{app().Config.GetString("token.platformName")}}",
            "sLogin": "jack"
          }
        },
        "message": "created",
        "success": true
      }
      """
