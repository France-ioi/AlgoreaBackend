Feature: Submit a new answer
  Background:
    Given the database has the following users:
      | login | group_id |
      | john  | 101      |
    And the database has the following table 'groups':
      | id  | name | type |
      | 201 | team | Team |
    And the database has the following table 'groups_groups':
      | parent_group_id | child_group_id |
      | 22              | 13             |
      | 201             | 101            |
    And the groups ancestors are computed
    And the database has the following table 'items':
      | id | default_language_tag |
      | 10 | fr                   |
      | 50 | fr                   |
    And the database has the following table 'items_items':
      | parent_item_id | child_item_id | child_order |
      | 10             | 50            | 1           |
    And the database has the following table 'items_ancestors':
      | ancestor_item_id | child_item_id |
      | 10               | 50            |
    And the database has the following table 'permissions_generated':
      | group_id | item_id | can_view_generated |
      | 101      | 50      | content            |
      | 201      | 50      | content            |
    And the database has the following table 'attempts':
      | id | participant_id |
      | 1  | 101            |
      | 1  | 201            |
    And the database has the following table 'results':
      | attempt_id | participant_id | item_id | hints_requested                 | hints_cached | submissions | latest_activity_at  | started_at          |
      | 1          | 101            | 50      | [{"rotorIndex":0,"cellRank":0}] | 12           | 2           | 2019-05-30 11:00:00 | 2019-05-30 11:00:00 |
      | 1          | 101            | 10      | null                            | 0            | 0           | 2019-05-30 11:00:00 | 2019-05-30 11:00:00 |
      | 1          | 201            | 50      | [{"rotorIndex":0,"cellRank":0}] | 2            | 1           | 2020-01-01 00:00:00 | 2020-01-01 00:00:00 |

  Scenario: User is able to submit a new answer
    Given time is frozen
    And "userTaskToken" is a token signed by the app with the following payload:
      """
      {
        "idUser": "101",
        "idItemLocal": "50",
        "idAttempt": "101/1",
        "platformName": "{{app().Config.GetString("token.platformName")}}"
      }
      """
    When I send a POST request to "/answers" with the following body:
      """
      {
        "task_token": "{{userTaskToken}}",
        "answer": "print 1"
      }
      """
    Then the response code should be 201
    And the response body decoded as "AnswersSubmitResponse" should be, in JSON:
      """
      {
        "data": {
          "answer_token": {
            "date": "{{currentTimeInFormat("02-01-2006")}}",
            "idUser": "101",
            "idItem": null,
            "idAttempt": "101/1",
            "itemUrl": "",
            "idItemLocal": "50",
            "platformName": "algrorea_backend",
            "randomSeed": "",
            "sHintsRequested": "[{\"rotorIndex\":0,\"cellRank\":0}]",
            "nbHintsGiven": "12",
            "sAnswer": "print 1",
            "idUserAnswer": "5577006791947779410"
          }
        },
        "message": "created",
        "success": true
      }
      """
    And the table "answers" should be:
      | author_id | participant_id | attempt_id | item_id | type       | answer  | ABS(TIMESTAMPDIFF(SECOND, created_at, NOW())) < 3 |
      | 101       | 101            | 1          | 50      | Submission | print 1 | 1                                                 |
    And the table "results" should be:
      | attempt_id | participant_id | item_id | hints_requested                 | hints_cached | submissions | ABS(TIMESTAMPDIFF(SECOND, latest_activity_at, NOW())) < 3 | ABS(TIMESTAMPDIFF(SECOND, latest_submission_at, NOW())) < 3 |
      | 1          | 101            | 10      | null                            | 0            | 0           | 1                                                         | null                                                        |
      | 1          | 101            | 50      | [{"rotorIndex":0,"cellRank":0}] | 12           | 3           | 1                                                         | 1                                                           |
      | 1          | 201            | 50      | [{"rotorIndex":0,"cellRank":0}] | 2            | 1           | 0                                                         | null                                                           |
    And the table "results_propagate" should be empty

  Scenario: User is able to submit a new answer for his team (participant_id is the first integer in the idAttempt in the task token)
    Given time is frozen
    And "userTaskToken" is a token signed by the app with the following payload:
      """
      {
        "idUser": "101",
        "idItemLocal": "50",
        "idAttempt": "201/1",
        "platformName": "{{app().Config.GetString("token.platformName")}}"
      }
      """
    When I send a POST request to "/answers" with the following body:
      """
      {
        "task_token": "{{userTaskToken}}",
        "answer": "print 1"
      }
      """
    Then the response code should be 201
    And the response body decoded as "AnswersSubmitResponse" should be, in JSON:
      """
      {
        "data": {
          "answer_token": {
            "date": "{{currentTimeInFormat("02-01-2006")}}",
            "idUser": "101",
            "idItem": null,
            "idAttempt": "201/1",
            "itemUrl": "",
            "idItemLocal": "50",
            "platformName": "algrorea_backend",
            "randomSeed": "",
            "sHintsRequested": "[{\"rotorIndex\":0,\"cellRank\":0}]",
            "nbHintsGiven": "2",
            "sAnswer": "print 1",
            "idUserAnswer": "5577006791947779410"
          }
        },
        "message": "created",
        "success": true
      }
      """
    And the table "answers" should be:
      | author_id | participant_id | attempt_id | item_id | type       | answer  | ABS(TIMESTAMPDIFF(SECOND, created_at, NOW())) < 3 |
      | 101       | 201            | 1          | 50      | Submission | print 1 | 1                                                 |
    And the table "results" should be:
      | attempt_id | participant_id | item_id | hints_requested                 | hints_cached | submissions | ABS(TIMESTAMPDIFF(SECOND, latest_activity_at, NOW())) < 3 | ABS(TIMESTAMPDIFF(SECOND, latest_submission_at, NOW())) < 3 |
      | 1          | 101            | 10      | null                            | 0            | 0           | 0                                                         | null                                                        |
      | 1          | 101            | 50      | [{"rotorIndex":0,"cellRank":0}] | 12           | 2           | 0                                                         | null                                                        |
      | 1          | 201            | 50      | [{"rotorIndex":0,"cellRank":0}] | 2            | 2           | 1                                                         | 1                                                           |
  And the table "results_propagate" should be empty

  Scenario: User is able to submit a new answer (with all fields filled in the token)
    Given time is frozen
    And "userTaskToken" is a token signed by the app with the following payload:
      """
      {
        "idItem": "50",
        "idUser": "101",
        "idItemLocal": "50",
        "idAttempt": "101/1",
        "itemUrl": "http://taskplatform.mblockelet.info/task.html?taskId=403449543672183936",
        "idItemLocal": "50",
        "randomSeed": "100",
        "platformName": "{{app().Config.GetString("token.platformName")}}"
      }
      """
    When I send a POST request to "/answers" with the following body:
      """
      {
        "task_token": "{{userTaskToken}}",
        "answer": "print(2)"
      }
      """
    Then the response code should be 201
    And the response body decoded as "AnswersSubmitResponse" should be, in JSON:
      """
      {
        "data": {
          "answer_token": {
            "date": "{{currentTimeInFormat("02-01-2006")}}",
            "idUser": "101",
            "idItem": "50",
            "idAttempt": "101/1",
            "itemUrl": "http://taskplatform.mblockelet.info/task.html?taskId=403449543672183936",
            "idItemLocal": "50",
            "platformName": "algrorea_backend",
            "randomSeed": "100",
            "sHintsRequested": "[{\"rotorIndex\":0,\"cellRank\":0}]",
            "nbHintsGiven": "12",
            "sAnswer": "print(2)",
            "idUserAnswer": "5577006791947779410"
          }
        },
        "message": "created",
        "success": true
      }
      """
    And the table "answers" should be:
      | author_id | participant_id | attempt_id | item_id | type       | answer   | ABS(TIMESTAMPDIFF(SECOND, created_at, NOW())) < 3 |
      | 101       | 101            | 1          | 50      | Submission | print(2) | 1                                                 |
    And the table "results" should be:
      | attempt_id | participant_id | item_id | hints_requested                 | hints_cached | submissions | ABS(TIMESTAMPDIFF(SECOND, latest_activity_at, NOW())) < 3 | ABS(TIMESTAMPDIFF(SECOND, latest_submission_at, NOW())) < 3 |
      | 1          | 101            | 10      | null                            | 0            | 0           | 1                                                         | null                                                        |
      | 1          | 101            | 50      | [{"rotorIndex":0,"cellRank":0}] | 12           | 3           | 1                                                         | 1                                                           |
      | 1          | 201            | 50      | [{"rotorIndex":0,"cellRank":0}] | 2            | 1           | 0                                                         | null                                                           |
    And the table "results_propagate" should be empty
