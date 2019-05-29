Feature: Save grading result
  Background:
    Given the database has the following table 'users':
      | ID  | sLogin | idGroupSelf |
      | 10  | john   | 101         |
    And the database has the following table 'groups':
      | ID  |
      | 101 |
    And the database has the following table 'groups_ancestors':
      | idGroupAncestor | idGroupChild | bIsSelf |
      | 101             | 101          | 1       |
    And the database has the following table 'groups_groups':
      | ID | idGroupParent | idGroupChild | sType              | sStatusDate |
      | 15 | 22            | 13           | direct             | null        |
    And the database has the following table 'platforms':
      | ID | bUsesTokens | sRegexp                                            | sPublicKey                |
      | 10 | 1           | http://taskplatform.mblockelet.info/task.html\?.*  | {{taskPlatformPublicKey}} |
      | 20 | 0           | http://taskplatform1.mblockelet.info/task.html\?.* |                           |
    And the database has the following table 'items':
      | ID | idPlatform | sUrl                                                                    | idItemUnlocked | iScoreMinUnlock | sValidationType |
      | 50 | 10         | http://taskplatform.mblockelet.info/task.html?taskId=403449543672183936 |                |                 |                 |
      | 60 | 10         | http://taskplatform.mblockelet.info/task.html?taskId=403449543672183937 | 50             | 98              |                 |
      | 10 | null       | null                                                                    |                |                 | AllButOne       |
      | 70 | 20         | http://taskplatform1.mblockelet.info/task.html?taskId=4034495436721839  |                |                 |                 |
    And the database has the following table 'items_items':
      | idItemParent | idItemChild |
      | 10           | 50          |
      | 10           | 60          |
    And the database has the following table 'items_ancestors':
      | idItemAncestor | idItemChild |
      | 10             | 50          |
      | 10             | 60          |
    And the database has the following table 'groups_items':
      | idGroup | idItem | sCachedPartialAccessDate |
      | 101     | 50     | 2017-05-29T06:38:38Z     |
      | 101     | 60     | 2017-05-29T06:38:38Z     |
      | 101     | 70     | 2017-05-29T06:38:38Z     |
    And the database has the following table 'users_items':
      | idUser | idItem | idAttemptActive | iScore | sBestAnswerDate      | sValidationDate      |
      | 10     | 10     | null            | 0      |                      | null                 |
      | 10     | 50     | 100             | 0      |                      | null                 |
      | 10     | 60     | 101             | 10     | 2017-05-29T06:38:38Z | 2019-03-29T06:38:38Z |
    And the database has the following table 'users_answers':
      | ID  | idUser | idItem |
      | 123 | 10     | 50     |
      | 124 | 10     | 60     |
      | 125 | 10     | 70     |
    And time is frozen

  Scenario: User is able to save the grading result with a high score and idAttempt
    Given I am the user with ID "10"
    And the database has the following table 'groups_attempts':
      | ID  | idGroup | idItem | sHintsRequested        |
      | 100 | 101     | 50     | [0,  1, "hint" , null] |
    And the following token "priorUserTaskToken" signed by the app is distributed:
      """
      {
        "idUser": "10",
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
        "idUser": "10",
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
            "idUser": "10",
            "idItemLocal": "50",
            "idAttempt": "100",
            "randomSeed": "456",
            "itemUrl": "http://taskplatform.mblockelet.info/task.html?taskId=403449543672183936",
            "platformName": "{{app().TokenConfig.PlatformName}}"
          },
          "key_obtained": true,
          "validated": true
        },
        "message": "created",
        "success": true
      }
      """
    And the table "users_answers" should be:
      | ID  | idUser | idItem | iScore | bValidated | ABS(sGradingDate - NOW()) < 3 |
      | 123 | 10     | 50     | 100    | 1          | 1                             |
      | 124 | 10     | 60     | null   | null       | null                          |
      | 125 | 10     | 70     | null   | null       | null                          |
    And the table "users_items" should be:
      | idUser | idItem | iScore | nbTasksTried | bValidated | bKeyObtained | sAncestorsComputationState | ABS(sLastActivityDate - NOW()) < 3 | ABS(sLastAnswerDate - NOW()) < 3 | ABS(sBestAnswerDate - NOW()) < 3 | ABS(sValidationDate - NOW()) < 3 |
      | 10     | 10     | 0      | 1            | 1          | 0            | done                       | 1                                  | null                             | 0                                | 1                                |
      | 10     | 50     | 100    | 1            | 1          | 1            | done                       | 1                                  | 1                                | 1                                | 1                                |
      | 10     | 60     | 10     | 0            | 0          | 0            | done                       | null                               | null                             | 0                                | 0                                |
    And the table "groups_attempts" should be:
      | ID  | iScore | nbTasksTried | bValidated | bKeyObtained | sAncestorsComputationState | ABS(sLastActivityDate - NOW()) < 3 | ABS(sLastAnswerDate - NOW()) < 3 | ABS(sBestAnswerDate - NOW()) < 3 | ABS(sValidationDate - NOW()) < 3 |
      | 100 | 100    | 1            | 1          | 1            | done                       | 1                                  | 1                                | 1                                | 1                                |

  Scenario: User is able to save the grading result with a low score and idAttempt
    Given I am the user with ID "10"
    And the database has the following table 'groups_attempts':
      | ID  | idGroup | idItem | sHintsRequested        |
      | 100 | 101     | 50     | [0,  1, "hint" , null] |
    And the following token "priorUserTaskToken" signed by the app is distributed:
      """
      {
        "idUser": "10",
        "idItemLocal": "50",
        "idAttempt": "100",
        "itemURL": "http://taskplatform.mblockelet.info/task.html?taskId=403449543672183936",
        "platformName": "{{app().TokenConfig.PlatformName}}"
      }
      """
    And the following token "scoreToken" signed by the task platform is distributed:
      """
      {
        "idUser": "10",
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
            "idUser": "10",
            "idItemLocal": "50",
            "idAttempt": "100",
            "randomSeed": "",
            "itemUrl": "http://taskplatform.mblockelet.info/task.html?taskId=403449543672183936",
            "platformName": "{{app().TokenConfig.PlatformName}}"
          },
          "key_obtained": false,
          "validated": false
        },
        "message": "created",
        "success": true
      }
      """
    And the table "users_answers" should be:
      | ID  | idUser | idItem | iScore | bValidated | ABS(sGradingDate - NOW()) < 3 |
      | 123 | 10     | 50     | 99     | 0          | 1                             |
      | 124 | 10     | 60     | null   | null       | null                          |
      | 125 | 10     | 70     | null   | null       | null                          |
    And the table "users_items" should be:
      | idUser | idItem | iScore | nbTasksTried | bValidated | bKeyObtained | sAncestorsComputationState | ABS(sLastActivityDate - NOW()) < 3 | ABS(sLastAnswerDate - NOW()) < 3 | ABS(sBestAnswerDate - NOW()) < 3 | ABS(sValidationDate - NOW()) < 3 |
      | 10     | 10     | 0      | 1            | 0          | 0            | done                       | 1                                  | null                             | 0                                | 0                                |
      | 10     | 50     | 99     | 1            | 0          | 0            | done                       | 1                                  | 1                                | 1                                | null                             |
      | 10     | 60     | 10     | 0            | 0          | 0            | done                       | null                               | null                             | 0                                | 0                                |
    And the table "groups_attempts" should be:
      | ID  | iScore | nbTasksTried | bValidated | bKeyObtained | sAncestorsComputationState | ABS(sLastActivityDate - NOW()) < 3 | ABS(sLastAnswerDate - NOW()) < 3 | ABS(sBestAnswerDate - NOW()) < 3 | ABS(sValidationDate - NOW()) < 3 |
      | 100 | 99     | 1            | 0          | 0            | done                       | 1                                  | 1                                | 1                                | null                             |

  Scenario: User is able to save the grading result with a low score, but still obtaining a key (with idAttempt)
    Given I am the user with ID "10"
    And the database has the following table 'groups_attempts':
      | ID  | idGroup | idItem | sBestAnswerDate      |
      | 100 | 101     | 60     | 2017-05-29T06:38:38Z |
    And the following token "priorUserTaskToken" signed by the app is distributed:
      """
      {
        "idUser": "10",
        "idItemLocal": "60",
        "idAttempt": "100",
        "itemURL": "http://taskplatform.mblockelet.info/task.html?taskId=403449543672183937",
        "platformName": "{{app().TokenConfig.PlatformName}}"
      }
      """
    And the following token "scoreToken" signed by the task platform is distributed:
      """
      {
        "idUser": "10",
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
            "idUser": "10",
            "idItemLocal": "60",
            "idAttempt": "100",
            "randomSeed": "",
            "itemUrl": "http://taskplatform.mblockelet.info/task.html?taskId=403449543672183937",
            "platformName": "{{app().TokenConfig.PlatformName}}"
          },
          "key_obtained": true,
          "validated": false
        },
        "message": "created",
        "success": true
      }
      """
    And the table "users_answers" should be:
      | ID  | idUser | idItem | iScore | bValidated | ABS(sGradingDate - NOW()) < 3 |
      | 123 | 10     | 50     | null   | null       | null                          |
      | 124 | 10     | 60     | 99     | 0          | 1                             |
      | 125 | 10     | 70     | null   | null       | null                          |
    And the table "users_items" should be:
      | idUser | idItem | iScore | nbTasksTried | bValidated | bKeyObtained | sAncestorsComputationState | ABS(sLastActivityDate - NOW()) < 3 | ABS(sLastAnswerDate - NOW()) < 3 | ABS(sBestAnswerDate - NOW()) < 3 | ABS(sValidationDate - NOW()) < 3 |
      | 10     | 10     | 0      | 1            | 0          | 0            | done                       | 1                                  | null                             | 0                                | 0                                |
      | 10     | 50     | 0      | 0            | 0          | 0            | done                       | null                               | null                             | 0                                | null                             |
      | 10     | 60     | 99     | 1            | 0          | 1            | done                       | 1                                  | 1                                | 1                                | 0                                |
    And the table "groups_attempts" should be:
      | ID  | iScore | nbTasksTried | bValidated | bKeyObtained | sAncestorsComputationState | ABS(sLastActivityDate - NOW()) < 3 | ABS(sLastAnswerDate - NOW()) < 3 | ABS(sBestAnswerDate - NOW()) < 3 | ABS(sValidationDate - NOW()) < 3 |
      | 100 | 99     | 1            | 0          | 1            | done                       | 1                                  | 1                                | 1                                | null                             |


  Scenario: Should keep previous score if it is greater
    Given I am the user with ID "10"
    And the database has the following table 'groups_attempts':
      | ID  | idGroup | idItem | iScore | sBestAnswerDate      |
      | 100 | 101     | 60     | 20     | 2018-05-29T06:38:38Z |
    And the following token "priorUserTaskToken" signed by the app is distributed:
      """
      {
        "idUser": "10",
        "idItemLocal": "60",
        "idAttempt": "100",
        "itemURL": "http://taskplatform.mblockelet.info/task.html?taskId=403449543672183937",
        "platformName": "{{app().TokenConfig.PlatformName}}"
      }
      """
    And the following token "scoreToken" signed by the task platform is distributed:
      """
      {
        "idUser": "10",
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
            "idUser": "10",
            "idItemLocal": "60",
            "idAttempt": "100",
            "itemUrl": "http://taskplatform.mblockelet.info/task.html?taskId=403449543672183937",
            "randomSeed": "",
            "platformName": "{{app().TokenConfig.PlatformName}}"
          },
          "key_obtained": false,
          "validated": false
        },
        "message": "created",
        "success": true
      }
      """
    And the table "users_answers" should be:
      | ID  | idUser | idItem | iScore | bValidated | ABS(sGradingDate - NOW()) < 3 |
      | 123 | 10     | 50     | null   | null       | null                          |
      | 124 | 10     | 60     | 5      | 0          | 1                             |
      | 125 | 10     | 70     | null   | null       | null                          |
    And the table "users_items" should be:
      | idUser | idItem | iScore | nbTasksTried | bValidated | bKeyObtained | sAncestorsComputationState | ABS(sLastActivityDate - NOW()) < 3 | ABS(sLastAnswerDate - NOW()) < 3 | ABS(sBestAnswerDate - NOW()) < 3 | ABS(sValidationDate - NOW()) < 3 |
      | 10     | 10     | 0      | 1            | 0          | 0            | done                       | 1                                  | null                             | 0                                | 0                                |
      | 10     | 50     | 0      | 0            | 0          | 0            | done                       | null                               | null                             | 0                                | null                             |
      | 10     | 60     | 10     | 1            | 0          | 0            | done                       | 1                                  | 1                                | 0                                | 0                                |
    And the table "groups_attempts" should stay unchanged

  Scenario: Should keep previous sValidationDate if it is earlier
    Given I am the user with ID "10"
    And the database has the following table 'groups_attempts':
      | ID  | idGroup | idItem | sValidationDate      |
      | 100 | 101     | 60     | 2018-05-29T06:38:38Z |
    And the following token "priorUserTaskToken" signed by the app is distributed:
      """
      {
        "idUser": "10",
        "idItemLocal": "60",
        "idAttempt": "100",
        "itemURL": "http://taskplatform.mblockelet.info/task.html?taskId=403449543672183937",
        "platformName": "{{app().TokenConfig.PlatformName}}"
      }
      """
    And the following token "scoreToken" signed by the task platform is distributed:
      """
      {
        "idUser": "10",
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
            "idUser": "10",
            "idItemLocal": "60",
            "idAttempt": "100",
            "itemUrl": "http://taskplatform.mblockelet.info/task.html?taskId=403449543672183937",
            "randomSeed": "",
            "platformName": "{{app().TokenConfig.PlatformName}}"
          },
          "key_obtained": true,
          "validated": true
        },
        "message": "created",
        "success": true
      }
      """
    And the table "users_answers" should be:
      | ID  | idUser | idItem | iScore | bValidated | ABS(sGradingDate - NOW()) < 3 |
      | 123 | 10     | 50     | null   | null       | null                          |
      | 124 | 10     | 60     | 100    | 1          | 1                             |
      | 125 | 10     | 70     | null   | null       | null                          |
    And the table "users_items" should be:
      | idUser | idItem | iScore | nbTasksTried | bValidated | bKeyObtained | sAncestorsComputationState | ABS(sLastActivityDate - NOW()) < 3 | ABS(sLastAnswerDate - NOW()) < 3 | ABS(sBestAnswerDate - NOW()) < 3 | ABS(sValidationDate - NOW()) < 3 |
      | 10     | 10     | 0      | 1            | 1          | 0            | done                       | 1                                  | null                             | 0                                | 0                                |
      | 10     | 50     | 0      | 0            | 0          | 0            | done                       | null                               | null                             | 0                                | null                             |
      | 10     | 60     | 100    | 1            | 1          | 1            | done                       | 1                                  | 1                                | 1                                | 0                                |
    And the table "groups_attempts" should stay unchanged

  Scenario: Should set bAccessSolutions=1 if the task has been validated
    Given I am the user with ID "10"
    And the following token "priorUserTaskToken" signed by the app is distributed:
      """
      {
        "idUser": "10",
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
        "idUser": "10",
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
            "idUser": "10",
            "idItemLocal": "50",
            "idAttempt": "100",
            "itemUrl": "http://taskplatform.mblockelet.info/task.html?taskId=403449543672183936",
            "randomSeed": "",
            "bAccessSolutions": true,
            "platformName": "{{app().TokenConfig.PlatformName}}"
          },
          "key_obtained": true,
          "validated": true
        },
        "message": "created",
        "success": true
      }
      """

  Scenario: Platform doesn't support tokens
    Given I am the user with ID "10"
    And the following token "priorUserTaskToken" signed by the app is distributed:
      """
      {
        "idUser": "10",
        "idItemLocal": "70",
        "idAttempt": "100",
        "itemURL": "http://taskplatform1.mblockelet.info/task.html?taskId=4034495436721839",
        "platformName": "{{app().TokenConfig.PlatformName}}"
      }
      """
    And the following token "answerToken" signed by the app is distributed:
      """
      {
        "idUser": "10",
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
            "idUser": "10",
            "idItemLocal": "70",
            "idAttempt": "100",
            "itemUrl": "http://taskplatform1.mblockelet.info/task.html?taskId=4034495436721839",
            "randomSeed": "",
            "platformName": "{{app().TokenConfig.PlatformName}}"
          },
          "key_obtained": true,
          "validated": true
        },
        "message": "created",
        "success": true
      }
      """
