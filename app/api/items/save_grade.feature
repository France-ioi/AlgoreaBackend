Feature: Save grading result
  Background:
    Given the database has the following users:
      | login | group_id |
      | john  | 101      |
    And the database has the following table 'groups_ancestors':
      | ancestor_group_id | child_group_id | is_self |
      | 101               | 101            | 1       |
    And the database has the following table 'groups_groups':
      | id | parent_group_id | child_group_id |
      | 15 | 22              | 13             |
    And the database has the following table 'platforms':
      | id | uses_tokens | regexp                                             | public_key                |
      | 10 | 1           | http://taskplatform.mblockelet.info/task.html\?.*  | {{taskPlatformPublicKey}} |
      | 20 | 0           | http://taskplatform1.mblockelet.info/task.html\?.* |                           |
    And the database has the following table 'items':
      | id | platform_id | url                                                                     | validation_type |
      | 50 | 10          | http://taskplatform.mblockelet.info/task.html?taskId=403449543672183936 | All             |
      | 60 | 10          | http://taskplatform.mblockelet.info/task.html?taskId=403449543672183937 | All             |
      | 10 | null        | null                                                                    | AllButOne       |
      | 70 | 20          | http://taskplatform1.mblockelet.info/task.html?taskId=4034495436721839  | All             |
    And the database has the following table 'item_unlocking_rules':
      | unlocking_item_id | unlocked_item_id | score |
      | 60                | 50               | 98    |
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
    And the database has the following table 'users_answers':
      | id  | user_id | item_id | submitted_at        |
      | 123 | 101     | 50      | 2017-05-29 06:38:38 |
      | 124 | 101     | 60      | 2017-05-29 06:38:38 |
      | 125 | 101     | 70      | 2017-05-29 06:38:38 |
    And time is frozen

  Scenario: User is able to save the grading result with a high score and attempt_id
    Given I am the user with id "101"
    And the database has the following table 'groups_attempts':
      | id  | group_id | item_id | hints_requested        | order |
      | 100 | 101      | 50      | [0,  1, "hint" , null] | 1     |
      | 101 | 101      | 60      | [0,  1, "hint" , null] | 2     |
    And the database has the following table 'users_items':
      | user_id | item_id | active_attempt_id |
      | 101     | 50      | 100               |
      | 101     | 60      | 101               |
    And the following token "priorUserTaskToken" signed by the app is distributed:
      """
      {
        "idUser": "101",
        "idItemLocal": "50",
        "idAttempt": "100",
        "randomSeed": "456",
        "itemURL": "http://taskplatform.mblockelet.info/task.html?taskId=403449543672183936",
        "platformName": "{{app().TokenConfig.PlatformName}}"
      }
      """
    And the following token "scoreToken" signed by the task platform is distributed:
      """
      {
        "idUser": "101",
        "idItemLocal": "50",
        "idAttempt": "100",
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
            "idAttempt": "100",
            "randomSeed": "456",
            "itemUrl": "http://taskplatform.mblockelet.info/task.html?taskId=403449543672183936",
            "platformName": "{{app().TokenConfig.PlatformName}}"
          },
          "has_unlocked_items": true,
          "validated": true
        },
        "message": "created",
        "success": true
      }
      """
    And the table "users_answers" should be:
      | id  | user_id | item_id | score | validated | ABS(TIMESTAMPDIFF(SECOND, graded_at, NOW())) < 3 |
      | 123 | 101     | 50      | 100   | 1         | 1                                                |
      | 124 | 101     | 60      | null  | null      | null                                             |
      | 125 | 101     | 70      | null  | null      | null                                             |
    And the table "users_items" should be:
      | user_id | item_id |
      | 101     | 50      |
      | 101     | 60      |
    And the table "groups_attempts" should be:
      | id  | score | tasks_tried | validated | ancestors_computation_state | ABS(TIMESTAMPDIFF(SECOND, latest_activity_at, NOW())) < 3 | ABS(TIMESTAMPDIFF(SECOND, latest_answer_at, NOW())) < 3 | ABS(TIMESTAMPDIFF(SECOND, best_answer_at, NOW())) < 3 | ABS(TIMESTAMPDIFF(SECOND, validated_at, NOW())) < 3 |
      | 100 | 100   | 1           | 1         | done                        | 1                                                         | 1                                                       | 1                                                     | 1                                                   |
      | 101 | 0     | 0           | 0         | done                        | null                                                      | null                                                    | null                                                  | null                                                |

  Scenario: User is able to save the grading result with a low score and idAttempt
    Given I am the user with id "101"
    And the database has the following table 'groups_attempts':
      | id  | group_id | item_id | hints_requested        | order |
      | 100 | 101      | 50      | [0,  1, "hint" , null] | 1     |
      | 101 | 101      | 60      | [0,  1, "hint" , null] | 2     |
    And the database has the following table 'users_items':
      | user_id | item_id | active_attempt_id |
      | 101     | 50      | 100               |
      | 101     | 60      | 101               |
    And the following token "priorUserTaskToken" signed by the app is distributed:
      """
      {
        "idUser": "101",
        "idItemLocal": "50",
        "idAttempt": "100",
        "itemURL": "http://taskplatform.mblockelet.info/task.html?taskId=403449543672183936",
        "platformName": "{{app().TokenConfig.PlatformName}}"
      }
      """
    And the following token "scoreToken" signed by the task platform is distributed:
      """
      {
        "idUser": "101",
        "idItemLocal": "50",
        "idAttempt": "100",
        "itemUrl": "http://taskplatform.mblockelet.info/task.html?taskId=403449543672183936",
        "score": "99",
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
            "idAttempt": "100",
            "randomSeed": "",
            "itemUrl": "http://taskplatform.mblockelet.info/task.html?taskId=403449543672183936",
            "platformName": "{{app().TokenConfig.PlatformName}}"
          },
          "has_unlocked_items": false,
          "validated": false
        },
        "message": "created",
        "success": true
      }
      """
    And the table "users_answers" should be:
      | id  | user_id | item_id | score | validated | ABS(TIMESTAMPDIFF(SECOND, graded_at, NOW())) < 3 |
      | 123 | 101     | 50      | 99    | 0         | 1                                                |
      | 124 | 101     | 60      | null  | null      | null                                             |
      | 125 | 101     | 70      | null  | null      | null                                             |
    And the table "users_items" should be:
      | user_id | item_id |
      | 101     | 50      |
      | 101     | 60      |
    And the table "groups_attempts" should be:
      | id  | score | tasks_tried | validated | ancestors_computation_state | ABS(TIMESTAMPDIFF(SECOND, latest_activity_at, NOW())) < 3 | ABS(TIMESTAMPDIFF(SECOND, latest_answer_at, NOW())) < 3 | ABS(TIMESTAMPDIFF(SECOND, best_answer_at, NOW())) < 3 | ABS(TIMESTAMPDIFF(SECOND, validated_at, NOW())) < 3 |
      | 100 | 99    | 1           | 0         | done                        | 1                                                         | 1                                                       | 1                                                     | null                                                |
      | 101 | 0     | 0           | 0         | done                        | null                                                      | null                                                    | null                                                  | null                                                |

  Scenario: User is able to save the grading result with a low score, but still obtaining a key (with idAttempt)
    Given I am the user with id "101"
    And the database has the following table 'groups_attempts':
      | id  | group_id | item_id | best_answer_at      | order |
      | 100 | 101      | 50      | 2017-05-29 06:38:38 | 1     |
      | 101 | 101      | 60      | 2017-05-29 06:38:38 | 2     |
    And the database has the following table 'users_items':
      | user_id | item_id | active_attempt_id |
      | 101     | 50      | 100               |
      | 101     | 60      | 101               |
    And the following token "priorUserTaskToken" signed by the app is distributed:
      """
      {
        "idUser": "101",
        "idItemLocal": "60",
        "idAttempt": "100",
        "itemURL": "http://taskplatform.mblockelet.info/task.html?taskId=403449543672183937",
        "platformName": "{{app().TokenConfig.PlatformName}}"
      }
      """
    And the following token "scoreToken" signed by the task platform is distributed:
      """
      {
        "idUser": "101",
        "idItemLocal": "60",
        "idAttempt": "100",
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
            "idAttempt": "100",
            "randomSeed": "",
            "itemUrl": "http://taskplatform.mblockelet.info/task.html?taskId=403449543672183937",
            "platformName": "{{app().TokenConfig.PlatformName}}"
          },
          "has_unlocked_items": true,
          "validated": false
        },
        "message": "created",
        "success": true
      }
      """
    And the table "users_answers" should be:
      | id  | user_id | item_id | score | validated | ABS(TIMESTAMPDIFF(SECOND, graded_at, NOW())) < 3 |
      | 123 | 101     | 50      | null  | null      | null                                             |
      | 124 | 101     | 60      | 99    | 0         | 1                                                |
      | 125 | 101     | 70      | null  | null      | null                                             |
    And the table "users_items" should be:
      | user_id | item_id |
      | 101     | 50      |
      | 101     | 60      |
    And the table "groups_attempts" should be:
      | id  | score | tasks_tried | validated | ancestors_computation_state | ABS(TIMESTAMPDIFF(SECOND, latest_activity_at, NOW())) < 3 | ABS(TIMESTAMPDIFF(SECOND, latest_answer_at, NOW())) < 3 | ABS(TIMESTAMPDIFF(SECOND, best_answer_at, NOW())) < 3 | ABS(TIMESTAMPDIFF(SECOND, validated_at, NOW())) < 3 |
      | 100 | 99    | 1           | 0         | done                        | 1                                                         | 1                                                       | 1                                                     | null                                                |
      | 101 | 0     | 0           | 0         | done                        | null                                                      | null                                                    | 0                                                     | null                                                |

  Scenario: Should keep previous score if it is greater
    Given I am the user with id "101"
    And the database has the following table 'groups_attempts':
      | id  | group_id | item_id | score | best_answer_at      | order |
      | 100 | 101      | 50      | 20    | 2018-05-29 06:38:38 | 1     |
      | 101 | 101      | 60      | 20    | 2018-05-29 06:38:38 | 2     |
    And the database has the following table 'users_items':
      | user_id | item_id | active_attempt_id |
      | 101     | 50      | 100               |
      | 101     | 60      | 101               |
    And the following token "priorUserTaskToken" signed by the app is distributed:
      """
      {
        "idUser": "101",
        "idItemLocal": "60",
        "idAttempt": "100",
        "itemURL": "http://taskplatform.mblockelet.info/task.html?taskId=403449543672183937",
        "platformName": "{{app().TokenConfig.PlatformName}}"
      }
      """
    And the following token "scoreToken" signed by the task platform is distributed:
      """
      {
        "idUser": "101",
        "idItemLocal": "60",
        "idAttempt": "100",
        "itemUrl": "http://taskplatform.mblockelet.info/task.html?taskId=403449543672183937",
        "score": "5",
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
            "idAttempt": "100",
            "itemUrl": "http://taskplatform.mblockelet.info/task.html?taskId=403449543672183937",
            "randomSeed": "",
            "platformName": "{{app().TokenConfig.PlatformName}}"
          },
          "has_unlocked_items": false,
          "validated": false
        },
        "message": "created",
        "success": true
      }
      """
    And the table "users_answers" should be:
      | id  | user_id | item_id | score | validated | ABS(TIMESTAMPDIFF(SECOND, graded_at, NOW())) < 3 |
      | 123 | 101     | 50      | null  | null      | null                                             |
      | 124 | 101     | 60      | 5     | 0         | 1                                                |
      | 125 | 101     | 70      | null  | null      | null                                             |
    And the table "users_items" should be:
      | user_id | item_id |
      | 101     | 50      |
      | 101     | 60      |
    And the table "groups_attempts" should stay unchanged

  Scenario: Should keep previous sValidationDate if it is earlier
    Given I am the user with id "101"
    And the database has the following table 'groups_attempts':
      | id  | group_id | item_id | validated_at        | order |
      | 100 | 101      | 50      | 2018-05-29 06:38:38 | 1     |
      | 101 | 101      | 60      | 2018-05-29 06:38:38 | 2     |
    And the database has the following table 'users_items':
      | user_id | item_id | active_attempt_id |
      | 101     | 50      | 100               |
      | 101     | 60      | 101               |
    And the following token "priorUserTaskToken" signed by the app is distributed:
      """
      {
        "idUser": "101",
        "idItemLocal": "60",
        "idAttempt": "100",
        "itemURL": "http://taskplatform.mblockelet.info/task.html?taskId=403449543672183937",
        "platformName": "{{app().TokenConfig.PlatformName}}"
      }
      """
    And the following token "scoreToken" signed by the task platform is distributed:
      """
      {
        "idUser": "101",
        "idItemLocal": "60",
        "idAttempt": "100",
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
            "idAttempt": "100",
            "itemUrl": "http://taskplatform.mblockelet.info/task.html?taskId=403449543672183937",
            "randomSeed": "",
            "platformName": "{{app().TokenConfig.PlatformName}}"
          },
          "has_unlocked_items": true,
          "validated": true
        },
        "message": "created",
        "success": true
      }
      """
    And the table "users_answers" should be:
      | id  | user_id | item_id | score | validated | ABS(TIMESTAMPDIFF(SECOND, graded_at, NOW())) < 3 |
      | 123 | 101     | 50      | null  | null      | null                                             |
      | 124 | 101     | 60      | 100   | 1         | 1                                                |
      | 125 | 101     | 70      | null  | null      | null                                             |
    And the table "users_items" should be:
      | user_id | item_id |
      | 101     | 50      |
      | 101     | 60      |
    And the table "groups_attempts" should stay unchanged

  Scenario: Should set bAccessSolutions=1 if the task has been validated
    Given I am the user with id "101"
    And the following token "priorUserTaskToken" signed by the app is distributed:
      """
      {
        "idUser": "101",
        "idItemLocal": "50",
        "idAttempt": "100",
        "itemURL": "http://taskplatform.mblockelet.info/task.html?taskId=403449543672183936",
        "bAccessSolutions": false,
        "platformName": "{{app().TokenConfig.PlatformName}}"
      }
      """
    And the following token "scoreToken" signed by the task platform is distributed:
      """
      {
        "idUser": "101",
        "idItemLocal": "50",
        "idAttempt": "100",
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
            "idAttempt": "100",
            "itemUrl": "http://taskplatform.mblockelet.info/task.html?taskId=403449543672183936",
            "randomSeed": "",
            "bAccessSolutions": true,
            "platformName": "{{app().TokenConfig.PlatformName}}"
          },
          "has_unlocked_items": true,
          "validated": true
        },
        "message": "created",
        "success": true
      }
      """

  Scenario: Platform doesn't support tokens
    Given I am the user with id "101"
    And the following token "priorUserTaskToken" signed by the app is distributed:
      """
      {
        "idUser": "101",
        "idItemLocal": "70",
        "idAttempt": "100",
        "itemURL": "http://taskplatform1.mblockelet.info/task.html?taskId=4034495436721839",
        "platformName": "{{app().TokenConfig.PlatformName}}"
      }
      """
    And the following token "answerToken" signed by the app is distributed:
      """
      {
        "idUser": "101",
        "idItemLocal": "70",
        "idAttempt": "100",
        "itemURL": "http://taskplatform1.mblockelet.info/task.html?taskId=4034495436721839",
        "idUserAnswer": "125",
        "platformName": "{{app().TokenConfig.PlatformName}}"
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
            "idAttempt": "100",
            "itemUrl": "http://taskplatform1.mblockelet.info/task.html?taskId=4034495436721839",
            "randomSeed": "",
            "platformName": "{{app().TokenConfig.PlatformName}}"
          },
          "has_unlocked_items": true,
          "validated": true
        },
        "message": "created",
        "success": true
      }
      """
