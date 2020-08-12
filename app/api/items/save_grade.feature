Feature: Save grading result
  Background:
    Given the database has the following users:
      | login | group_id |
      | john  | 101      |
    And the database has the following table 'groups_groups':
      | parent_group_id | child_group_id |
      | 22              | 13             |
    And the groups ancestors are computed
    And the database has the following table 'platforms':
      | id | regexp                                             | priority | public_key                |
      | 10 | http://taskplatform.mblockelet.info/task.html\?.*  | 2        | {{taskPlatformPublicKey}} |
      | 20 | http://taskplatform1.mblockelet.info/task.html\?.* | 1        | null                      |
    And the database has the following table 'items':
      | id | platform_id | url                                                                     | validation_type | default_language_tag |
      | 50 | 10          | http://taskplatform.mblockelet.info/task.html?taskId=403449543672183936 | All             | fr                   |
      | 60 | 10          | http://taskplatform.mblockelet.info/task.html?taskId=403449543672183937 | All             | fr                   |
      | 10 | null        | null                                                                    | AllButOne       | fr                   |
      | 70 | 20          | http://taskplatform1.mblockelet.info/task.html?taskId=4034495436721839  | All             | fr                   |
    And the database has the following table 'item_dependencies':
      | item_id | dependent_item_id | score |
      | 60      | 50                | 98    |
    And the database has the following table 'items_items':
      | parent_item_id | child_item_id | child_order |
      | 10             | 50            | 0           |
      | 10             | 60            | 1           |
    And the database has the following table 'items_ancestors':
      | ancestor_item_id | child_item_id |
      | 10               | 50            |
      | 10               | 60            |
    And the database has the following table 'permissions_generated':
      | group_id | item_id | can_view_generated |
      | 101      | 50      | content            |
      | 101      | 60      | content            |
      | 101      | 70      | content            |
    And time is frozen

  Scenario: User is able to save the grading result with a high score and attempt_id
    Given I am the user with id "101"
    And the database has the following table 'attempts':
      | id | participant_id |
      | 0  | 101            |
      | 1  | 101            |
    And the database has the following table 'results':
      | attempt_id | participant_id | item_id | latest_activity_at  | hints_requested        | result_propagation_state |
      | 0          | 101            | 10      | 2019-05-30 11:00:00 | null                   | done                     |
      | 0          | 101            | 50      | 2019-05-30 11:00:00 | [0,  1, "hint" , null] | done                     |
      | 1          | 101            | 60      | 2019-05-29 11:00:00 | [0,  1, "hint" , null] | done                     |
    And the database has the following table 'answers':
      | id  | author_id | participant_id | attempt_id | item_id | created_at          |
      | 123 | 101       | 101            | 0          | 50      | 2017-05-29 06:38:38 |
      | 124 | 101       | 101            | 0          | 60      | 2017-05-29 06:38:38 |
    And the following token "priorUserTaskToken" signed by the app is distributed:
      """
      {
        "idUser": "101",
        "idItemLocal": "50",
        "idAttempt": "101/0",
        "randomSeed": "456",
        "itemURL": "http://taskplatform.mblockelet.info/task.html?taskId=403449543672183936",
        "platformName": "{{app().Config.GetString("token.platformName")}}"
      }
      """
    And the following token "scoreToken" signed by the task platform is distributed:
      """
      {
        "idUser": "101",
        "idItemLocal": "50",
        "idAttempt": "101/0",
        "itemUrl": "http://taskplatform.mblockelet.info/task.html?taskId=403449543672183936",
        "score": "100",
        "idUserAnswer": "123"
      }
      """
    When I send a POST request to "/items/save-grade" with the following body:
      """
      {
        "task_token": "{{priorUserTaskToken}}",
        "score_token": "{{scoreToken}}"
      }
      """
    Then the response code should be 201
    And the response body decoded as "SaveGradeResponse" should be, in JSON:
      """
      {
        "data": {
          "task_token": {
            "date": "{{currentTimeInFormat("02-01-2006")}}",
            "idUser": "101",
            "idItemLocal": "50",
            "idAttempt": "101/0",
            "randomSeed": "456",
            "itemUrl": "http://taskplatform.mblockelet.info/task.html?taskId=403449543672183936",
            "platformName": "{{app().Config.GetString("token.platformName")}}"
          },
          "validated": true
        },
        "message": "created",
        "success": true
      }
      """
    And the table "answers" should stay unchanged
    And the table "gradings" should be:
      | answer_id | score | ABS(TIMESTAMPDIFF(SECOND, graded_at, NOW())) < 3 |
      | 123       | 100   | 1                                                |
    And the table "attempts" should stay unchanged
    And the table "results" should be:
      | attempt_id | participant_id | item_id | score_computed | tasks_tried | validated | result_propagation_state | latest_activity_at  | latest_submission_at | score_obtained_at   | validated_at        |
      | 0          | 101            | 10      | 50             | 1           | 1         | done                     | 2019-05-30 11:00:00 | null                 | null                | 2017-05-29 06:38:38 |
      | 0          | 101            | 50      | 100            | 1           | 1         | done                     | 2019-05-30 11:00:00 | null                 | 2017-05-29 06:38:38 | 2017-05-29 06:38:38 |
      | 1          | 101            | 60      | 0              | 0           | 0         | done                     | 2019-05-29 11:00:00 | null                 | null                | null                |

  Scenario Outline: User is able to save the grading result with a low score and idAttempt
    Given I am the user with id "101"
    And the database has the following table 'attempts':
      | id | participant_id |
      | 0  | 101            |
      | 1  | 101            |
    And the database has the following table 'results':
      | attempt_id | participant_id | item_id | hints_requested        | latest_activity_at  | score_edit_rule   | score_edit_value   | result_propagation_state |
      | 0          | 101            | 10      | null                   | 2019-05-30 11:00:00 | null              | null               | done                     |
      | 0          | 101            | 50      | [0,  1, "hint" , null] | 2019-05-30 11:00:00 | <score_edit_rule> | <score_edit_value> | done                     |
      | 1          | 101            | 60      | [0,  1, "hint" , null] | 2019-05-29 11:00:00 | null              | null               | done                     |
    And the database has the following table 'answers':
      | id  | author_id | participant_id | attempt_id | item_id | created_at          |
      | 123 | 101       | 101            | 0          | 50      | 2017-05-29 06:38:38 |
      | 124 | 101       | 101            | 1          | 60      | 2017-05-29 06:38:38 |
    And the following token "priorUserTaskToken" signed by the app is distributed:
      """
      {
        "idUser": "101",
        "idItemLocal": "50",
        "idAttempt": "101/0",
        "itemURL": "http://taskplatform.mblockelet.info/task.html?taskId=403449543672183936",
        "platformName": "{{app().Config.GetString("token.platformName")}}"
      }
      """
    And the following token "scoreToken" signed by the task platform is distributed:
      """
      {
        "idUser": "101",
        "idItemLocal": "50",
        "idAttempt": "101/0",
        "itemUrl": "http://taskplatform.mblockelet.info/task.html?taskId=403449543672183936",
        "score": "<score>",
        "idUserAnswer": "123"
      }
      """
    When I send a POST request to "/items/save-grade" with the following body:
      """
      {
        "task_token": "{{priorUserTaskToken}}",
        "score_token": "{{scoreToken}}"
      }
      """
    Then the response code should be 201
    And the response body decoded as "SaveGradeResponse" should be, in JSON:
      """
      {
        "data": {
          "task_token": {
            "date": "{{currentTimeInFormat("02-01-2006")}}",
            "idUser": "101",
            "idItemLocal": "50",
            "idAttempt": "101/0",
            "randomSeed": "",
            "itemUrl": "http://taskplatform.mblockelet.info/task.html?taskId=403449543672183936",
            "platformName": "{{app().Config.GetString("token.platformName")}}"
          },
          "validated": false
        },
        "message": "created",
        "success": true
      }
      """
    And the table "answers" should stay unchanged
    And the table "gradings" should be:
      | answer_id | score   | ABS(TIMESTAMPDIFF(SECOND, graded_at, NOW())) < 3 |
      | 123       | <score> | 1                                                |
    And the table "attempts" should stay unchanged
    And the table "results" should be:
      | attempt_id | participant_id | item_id | score_computed   | tasks_tried | validated | result_propagation_state | latest_activity_at  | latest_submission_at | score_obtained_at   | validated_at |
      | 0          | 101            | 10      | <parent_score>   | 1           | 0         | done                     | 2019-05-30 11:00:00 | null                 | null                | null         |
      | 0          | 101            | 50      | <score_computed> | 1           | 0         | done                     | 2019-05-30 11:00:00 | null                 | 2017-05-29 06:38:38 | null         |
      | 1          | 101            | 60      | 0                | 0           | 0         | done                     | 2019-05-29 11:00:00 | null                 | null                | null         |
  Examples:
    | score | score_edit_rule | score_edit_value | score_computed | parent_score |
    | 99    | null            | null             | 99             | 49.5         |
    | 89    | set             | 10               | 10             | 5            |
    | 89    | set             | -10              | 0              | 0            |
    | 79    | diff            | -10              | 69             | 34.5         |
    | 79    | diff            | -80              | 0              | 0            |
    | 79    | diff            | 80               | 100            | 50           |

  Scenario: User is able to save the grading result with a low score, but still obtaining a key (with idAttempt)
    Given I am the user with id "101"
    And the database has the following table 'attempts':
      | id | participant_id |
      | 0  | 101            |
      | 1  | 101            |
    And the database has the following table 'results':
      | attempt_id | participant_id | item_id | score_obtained_at   | latest_activity_at  | result_propagation_state |
      | 0          | 101            | 50      | 2017-04-29 06:38:38 | 2019-05-30 11:00:00 | done                     |
      | 1          | 101            | 60      | 2017-05-29 06:38:38 | 2019-05-29 11:00:00 | done                     |
    And the database has the following table 'answers':
      | id  | author_id | participant_id | attempt_id | item_id | created_at          |
      | 123 | 101       | 101            | 0          | 50      | 2017-05-29 06:38:38 |
      | 124 | 101       | 101            | 1          | 60      | 2017-05-29 06:38:38 |
    And the following token "priorUserTaskToken" signed by the app is distributed:
      """
      {
        "idUser": "101",
        "idItemLocal": "60",
        "idAttempt": "101/1",
        "itemURL": "http://taskplatform.mblockelet.info/task.html?taskId=403449543672183937",
        "platformName": "{{app().Config.GetString("token.platformName")}}"
      }
      """
    And the following token "scoreToken" signed by the task platform is distributed:
      """
      {
        "idUser": "101",
        "idItemLocal": "60",
        "idAttempt": "101/1",
        "itemUrl": "http://taskplatform.mblockelet.info/task.html?taskId=403449543672183937",
        "score": "99",
        "idUserAnswer": "124"
      }
      """
    When I send a POST request to "/items/save-grade" with the following body:
      """
      {
        "task_token": "{{priorUserTaskToken}}",
        "score_token": "{{scoreToken}}"
      }
      """
    Then the response code should be 201
    And the response body decoded as "SaveGradeResponse" should be, in JSON:
      """
      {
        "data": {
          "task_token": {
            "date": "{{currentTimeInFormat("02-01-2006")}}",
            "idUser": "101",
            "idItemLocal": "60",
            "idAttempt": "101/1",
            "randomSeed": "",
            "itemUrl": "http://taskplatform.mblockelet.info/task.html?taskId=403449543672183937",
            "platformName": "{{app().Config.GetString("token.platformName")}}"
          },
          "validated": false
        },
        "message": "created",
        "success": true
      }
      """
    And the table "answers" should stay unchanged
    And the table "gradings" should be:
      | answer_id | score | ABS(TIMESTAMPDIFF(SECOND, graded_at, NOW())) < 3 |
      | 124       | 99    | 1                                                |
    And the table "attempts" should stay unchanged
    And the table "results" should be:
      | participant_id | attempt_id | item_id | score_computed | tasks_tried | validated | result_propagation_state | latest_activity_at  | latest_submission_at | score_obtained_at   | validated_at |
      | 101            | 0          | 50      | 0              | 0           | 0         | done                     | 2019-05-30 11:00:00 | null                 | 2017-04-29 06:38:38 | null         |
      | 101            | 1          | 60      | 99             | 1           | 0         | done                     | 2019-05-29 11:00:00 | null                 | 2017-05-29 06:38:38 | null         |

  Scenario Outline: Should keep previous score if it is greater
    Given I am the user with id "101"
    And the database has the following table 'answers':
      | id  | author_id | participant_id | attempt_id | item_id | created_at          |
      | 123 | 101       | 101            | 0          | 10      | 2018-05-29 06:38:38 |
      | 124 | 101       | 101            | 0          | 50      | 2018-05-29 06:38:38 |
      | 125 | 101       | 101            | 0          | 10      | 2018-05-29 06:38:38 |
    And the database has the following table 'gradings':
      | answer_id | score | graded_at           |
      | 123       | 5     | 2018-05-29 06:38:38 |
      | 125       | 20    | 2018-05-29 06:38:38 |
    And the database has the following table 'attempts':
      | id | participant_id |
      | 0  | 101            |
    And the database has the following table 'results':
      | participant_id | attempt_id | item_id | score_computed | score_obtained_at   | score_edit_rule   | score_edit_value   | result_propagation_state |
      | 101            | 0          | 10      | 20             | 2018-05-29 06:38:38 | null              | null               | done                     |
      | 101            | 0          | 50      | 20             | 2018-05-29 06:38:38 | <score_edit_rule> | <score_edit_value> | done                     |
      | 101            | 0          | 60      | 20             | 2018-05-29 06:38:38 | null              | null               | done                     |
    And the following token "priorUserTaskToken" signed by the app is distributed:
      """
      {
        "idUser": "101",
        "idItemLocal": "60",
        "idAttempt": "101/0",
        "itemURL": "http://taskplatform.mblockelet.info/task.html?taskId=403449543672183937",
        "platformName": "{{app().Config.GetString("token.platformName")}}"
      }
      """
    And the following token "scoreToken" signed by the task platform is distributed:
      """
      {
        "idUser": "101",
        "idItemLocal": "60",
        "idAttempt": "101/0",
        "itemUrl": "http://taskplatform.mblockelet.info/task.html?taskId=403449543672183937",
        "score": "<score>",
        "idUserAnswer": "124"
      }
      """
    When I send a POST request to "/items/save-grade" with the following body:
      """
      {
        "task_token": "{{priorUserTaskToken}}",
        "score_token": "{{scoreToken}}"
      }
      """
    Then the response code should be 201
    And the response body decoded as "SaveGradeResponse" should be, in JSON:
      """
      {
        "data": {
          "task_token": {
            "date": "{{currentTimeInFormat("02-01-2006")}}",
            "idUser": "101",
            "idItemLocal": "60",
            "idAttempt": "101/0",
            "itemUrl": "http://taskplatform.mblockelet.info/task.html?taskId=403449543672183937",
            "randomSeed": "",
            "platformName": "{{app().Config.GetString("token.platformName")}}"
          },
          "validated": false
        },
        "message": "created",
        "success": true
      }
      """
    And the table "answers" should stay unchanged
    And the table "gradings" should be:
      | answer_id | score   | ABS(TIMESTAMPDIFF(SECOND, graded_at, NOW())) < 3 |
      | 123       | 5       | 0                                                |
      | 124       | <score> | 1                                                |
      | 125       | 20      | 0                                                |
    And the table "attempts" should stay unchanged
    And the table "results" should stay unchanged
    Examples:
      | score | score_edit_rule | score_edit_value |
      | 19    | null            | null             |
      | 19    | set             | 10               |
      | 19    | set             | -10              |
      | 20    | diff            | -1               |
      | 15    | diff            | -80              |

  Scenario: Should keep previous validated_at if it is earlier
    Given I am the user with id "101"
    And the database has the following table 'attempts':
      | id | participant_id |
      | 0  | 101            |
    And the database has the following table 'results':
      | attempt_id | participant_id | item_id | validated_at        | result_propagation_state |
      | 0          | 101            | 10      | 2016-05-29 06:38:37 | done                     |
      | 0          | 101            | 50      | 2016-05-29 06:38:37 | done                     |
      | 0          | 101            | 60      | 2015-05-29 06:38:37 | done                     |
    And the database has the following table 'answers':
      | id  | author_id | participant_id | attempt_id | item_id | created_at          |
      | 123 | 101       | 101            | 0          | 50      | 2017-05-29 06:38:38 |
      | 124 | 101       | 101            | 0          | 60      | 2017-05-29 06:38:38 |
    And the following token "priorUserTaskToken" signed by the app is distributed:
      """
      {
        "idUser": "101",
        "idItemLocal": "60",
        "idAttempt": "101/0",
        "itemURL": "http://taskplatform.mblockelet.info/task.html?taskId=403449543672183937",
        "platformName": "{{app().Config.GetString("token.platformName")}}"
      }
      """
    And the following token "scoreToken" signed by the task platform is distributed:
      """
      {
        "idUser": "101",
        "idItemLocal": "60",
        "idAttempt": "101/0",
        "itemUrl": "http://taskplatform.mblockelet.info/task.html?taskId=403449543672183937",
        "score": "100",
        "idUserAnswer": "124"
      }
      """
    When I send a POST request to "/items/save-grade" with the following body:
      """
      {
        "task_token": "{{priorUserTaskToken}}",
        "score_token": "{{scoreToken}}"
      }
      """
    Then the response code should be 201
    And the response body decoded as "SaveGradeResponse" should be, in JSON:
      """
      {
        "data": {
          "task_token": {
            "date": "{{currentTimeInFormat("02-01-2006")}}",
            "idUser": "101",
            "idItemLocal": "60",
            "idAttempt": "101/0",
            "itemUrl": "http://taskplatform.mblockelet.info/task.html?taskId=403449543672183937",
            "randomSeed": "",
            "platformName": "{{app().Config.GetString("token.platformName")}}"
          },
          "validated": true
        },
        "message": "created",
        "success": true
      }
      """
    And the table "answers" should stay unchanged
    And the table "gradings" should be:
      | answer_id | score | ABS(TIMESTAMPDIFF(SECOND, graded_at, NOW())) < 3 |
      | 124       | 100   | 1                                                |
    And the table "attempts" should stay unchanged
    And the table "results" should stay unchanged

  Scenario: Should set bAccessSolutions=1 if the task has been validated
    Given I am the user with id "101"
    And the database has the following table 'attempts':
      | id | participant_id |
      | 0  | 101            |
    And the database has the following table 'results':
      | attempt_id | participant_id | item_id | validated_at        | result_propagation_state |
      | 0          | 101            | 50      | 2018-05-29 06:38:38 | done                     |
    And the database has the following table 'answers':
      | id  | author_id | participant_id | attempt_id | item_id | created_at          |
      | 123 | 101       | 101            | 100        | 50      | 2017-05-29 06:38:38 |
    And the following token "priorUserTaskToken" signed by the app is distributed:
      """
      {
        "idUser": "101",
        "idItemLocal": "50",
        "idAttempt": "101/0",
        "itemURL": "http://taskplatform.mblockelet.info/task.html?taskId=403449543672183936",
        "bAccessSolutions": false,
        "platformName": "{{app().Config.GetString("token.platformName")}}"
      }
      """
    And the following token "scoreToken" signed by the task platform is distributed:
      """
      {
        "idUser": "101",
        "idItemLocal": "50",
        "idAttempt": "101/0",
        "itemUrl": "http://taskplatform.mblockelet.info/task.html?taskId=403449543672183936",
        "score": "100",
        "idUserAnswer": "123"
      }
      """
    When I send a POST request to "/items/save-grade" with the following body:
      """
      {
        "task_token": "{{priorUserTaskToken}}",
        "score_token": "{{scoreToken}}"
      }
      """
    Then the response code should be 201
    And the response body decoded as "SaveGradeResponse" should be, in JSON:
      """
      {
        "data": {
          "task_token": {
            "date": "{{currentTimeInFormat("02-01-2006")}}",
            "idUser": "101",
            "idItemLocal": "50",
            "idAttempt": "101/0",
            "itemUrl": "http://taskplatform.mblockelet.info/task.html?taskId=403449543672183936",
            "randomSeed": "",
            "bAccessSolutions": true,
            "platformName": "{{app().Config.GetString("token.platformName")}}"
          },
          "validated": true
        },
        "message": "created",
        "success": true
      }
      """

  Scenario: Platform doesn't support tokens
    Given I am the user with id "101"
    And the database has the following table 'attempts':
      | id | participant_id |
      | 1  | 101            |
    And the database has the following table 'results':
      | attempt_id | participant_id | item_id | validated_at        | result_propagation_state |
      | 1          | 101            | 70      | 2018-05-29 06:38:38 | done                     |
    And the database has the following table 'answers':
      | id  | author_id | participant_id | attempt_id | item_id | created_at          |
      | 125 | 101       | 101            | 100        | 70      | 2017-05-29 06:38:38 |
    And the following token "priorUserTaskToken" signed by the app is distributed:
      """
      {
        "idUser": "101",
        "idItemLocal": "70",
        "idAttempt": "101/1",
        "itemURL": "http://taskplatform1.mblockelet.info/task.html?taskId=4034495436721839",
        "platformName": "{{app().Config.GetString("token.platformName")}}"
      }
      """
    And the following token "answerToken" signed by the app is distributed:
      """
      {
        "idUser": "101",
        "idItemLocal": "70",
        "idAttempt": "101/1",
        "itemURL": "http://taskplatform1.mblockelet.info/task.html?taskId=4034495436721839",
        "idUserAnswer": "125",
        "platformName": "{{app().Config.GetString("token.platformName")}}"
      }
      """
    When I send a POST request to "/items/save-grade" with the following body:
      """
      {
        "task_token": "{{priorUserTaskToken}}",
        "score": 100.0,
        "answer_token": "{{answerToken}}"
      }
      """
    Then the response code should be 201
    And the response body decoded as "SaveGradeResponse" should be, in JSON:
      """
      {
        "data": {
          "task_token": {
            "date": "{{currentTimeInFormat("02-01-2006")}}",
            "idUser": "101",
            "idItemLocal": "70",
            "idAttempt": "101/1",
            "itemUrl": "http://taskplatform1.mblockelet.info/task.html?taskId=4034495436721839",
            "randomSeed": "",
            "platformName": "{{app().Config.GetString("token.platformName")}}"
          },
          "validated": true
        },
        "message": "created",
        "success": true
      }
      """
