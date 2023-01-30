Feature: Generate a read-only task token for an item from an answer
  Background:
    Given the database has the following table 'groups':
      | id  | name     | type  |
      | 101 | john     | User  |
      | 102 | team     | Team  |
      | 103 | manager  | User  |
      | 104 | Groupe A | Class |
    And the database has the following table 'users':
      | login | group_id |
      | john    | 101    |
      | manager | 103    |
    And the database has the following table 'groups_groups':
      | parent_group_id | child_group_id |
      | 102             | 101            |
      | 104             | 101            |
    And the database has the following table 'group_managers':
      | group_id | manager_id | can_watch_members |
      | 104      | 103        | 1                 |
    And the groups ancestors are computed
    And the database has the following table 'items':
      | id | url                                                                     | type | entry_participant_type | default_language_tag | text_id |
      | 10 | http://taskplatform.mblockelet.info/task.html?taskId=403449543672183936 | Task | User                   | fr                   | task10  |
      | 20 | http://taskplatform.mblockelet.info/task.html?taskId=403449543672183936 | Task | Team                   | fr                   | task20  |
      | 30 | http://taskplatform.mblockelet.info/task.html?taskId=403449543672183936 | Task | User                   | fr                   | task30  |
      | 40 | http://taskplatform.mblockelet.info/task.html?taskId=403449543672183936 | Task | User                   | fr                   | task40  |
      | 50 | http://taskplatform.mblockelet.info/task.html?taskId=403449543672183936 | Task | User                   | fr                   | task50  |
    And the database has the following table 'permissions_generated':
      | group_id | item_id | can_view_generated | can_watch_generated |
      | 101      | 10      | content            | none                |
      | 101      | 40      | solution           | none                |
      | 101      | 50      | content            | none                |
      | 102      | 20      | content            | none                |
      | 103      | 10      | content            | answer              |
      | 103      | 20      | content            | answer              |
      | 103      | 30      | content            | answer              |
      | 104      | 10      | content            | answer              |
      | 104      | 20      | content            | answer              |
      | 104      | 30      | content            | answer              |
    And the database has the following table 'attempts':
      | id | participant_id |
      | 0  | 101            |
      | 0  | 102            |
      | 0  | 103            |
      | 1  | 103            |
    And the database has the following table 'results':
      | attempt_id | participant_id | item_id | started_at          | validated_at        |
      | 0          | 101            | 10      | 2020-01-01 01:01:01 | null                |
      | 0          | 101            | 20      | 2020-01-01 01:01:01 | null                |
      | 0          | 101            | 30      | 2020-01-01 01:01:01 | null                |
      | 0          | 102            | 20      | 2020-01-01 01:01:01 | null                |
      | 0          | 103            | 10      | 2020-01-01 01:01:01 | null                |
      | 0          | 103            | 20      | 2020-01-01 01:01:01 | null                |
      | 0          | 103            | 30      | null                | null                |
      | 1          | 103            | 30      | 2020-01-01 01:01:01 | null                |
      | 0          | 101            | 40      | 2020-01-01 01:01:01 | null                |
      | 0          | 101            | 50      | 2020-01-01 01:01:01 | 2020-01-01 01:01:01 |
    And the database has the following table 'answers':
      | id | participant_id | attempt_id | item_id | author_id  | created_at          |
      | 1  | 101            | 0          | 10      | 101        | 2020-01-01 01:01:01 |
      | 2  | 102            | 0          | 20      | 101        | 2020-01-01 01:01:01 |
      | 3  | 101            | 0          | 30      | 101        | 2020-01-01 01:01:01 |
      | 4  | 101            | 0          | 40      | 101        | 2020-01-01 01:01:01 |
      | 5  | 101            | 0          | 50      | 101        | 2020-01-01 01:01:01 |
    And time is frozen

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
            "itemUrl": "http://taskplatform.mblockelet.info/task.html?taskId=403449543672183936",
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
            "idUser": "101",
            "idItemLocal": "20",
            "idItem": "task20",
            "itemUrl": "http://taskplatform.mblockelet.info/task.html?taskId=403449543672183936",
            "nbHintsGiven": "0",
            "randomSeed": "6688673066740915448",
            "platformName": "{{app().Config.GetString("token.platformName")}}",
            "sLogin": "john"
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
            "itemUrl": "http://taskplatform.mblockelet.info/task.html?taskId=403449543672183936",
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
            "itemUrl": "http://taskplatform.mblockelet.info/task.html?taskId=403449543672183936",
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

  Scenario: bAccessSolutions is true when item `can_view` >= 'solution'
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
            "idAttempt": "101/0",
            "idUser": "101",
            "idItemLocal": "40",
            "idItem": "task40",
            "itemUrl": "http://taskplatform.mblockelet.info/task.html?taskId=403449543672183936",
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

  Scenario: bAccessSolutions is true when item has been validated in the current attempt
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
            "idAttempt": "101/0",
            "idUser": "101",
            "idItemLocal": "50",
            "idItem": "task50",
            "itemUrl": "http://taskplatform.mblockelet.info/task.html?taskId=403449543672183936",
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
