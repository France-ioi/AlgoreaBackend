Feature: Save grading result - robustness
  Background:
    Given the database has the following users:
      | login | group_id |
      | john  | 101      |
    And the database has the following table 'groups_ancestors':
      | ancestor_group_id | child_group_id |
      | 101               | 101            |
    And the database has the following table 'groups_groups':
      | parent_group_id | child_group_id |
      | 22              | 13             |
    And the database has the following table 'platforms':
      | id | regexp                                             | priority | public_key                |
      | 10 | http://taskplatform.mblockelet.info/task.html\?.*  | 2        | {{taskPlatformPublicKey}} |
      | 20 | http://taskplatform1.mblockelet.info/task.html\?.* | 1        | null                      |
    And the database has the following table 'items':
      | id | platform_id | url                                                                     | read_only | validation_type | default_language_tag |
      | 50 | 10          | http://taskplatform.mblockelet.info/task.html?taskId=403449543672183936 | 1         | All             | fr                   |
      | 70 | 20          | http://taskplatform1.mblockelet.info/task.html?taskId=4034495436721839  | 0         | All             | fr                   |
      | 80 | 10          | http://taskplatform.mblockelet.info/task.html?taskId=403449543672183937 | 0         | All             | fr                   |
      | 10 | null        | null                                                                    | 0         | AllButOne       | fr                   |
    And the database has the following table 'items_items':
      | parent_item_id | child_item_id | child_order |
      | 10             | 50            | 0           |
    And the database has the following table 'items_ancestors':
      | ancestor_item_id | child_item_id |
      | 10               | 50            |
    And the database has the following table 'permissions_generated':
      | group_id | item_id | can_view_generated |
      | 101      | 50      | content            |
      | 101      | 70      | content            |
      | 101      | 80      | content            |
    And the database has the following table 'attempts':
      | id | participant_id |
      | 0  | 101            |
    And the database has the following table 'results':
      | attempt_id | participant_id | item_id | hints_requested        |
      | 0          | 101            | 50      | [0,  1, "hint" , null] |
    And the database has the following table 'answers':
      | id  | author_id | participant_id | attempt_id | item_id | created_at          |
      | 123 | 101       | 101            | 100        | 50      | 2017-05-29 06:38:38 |
    And time is frozen

  Scenario: Wrong JSON in request
    Given I am the user with id "101"
    When I send a POST request to "/items/save-grade" with the following body:
      """
      []
      """
    Then the response code should be 400
    And the response error message should contain "Json: cannot unmarshal array into Go value of type items.saveGradeRequest"
    And the table "answers" should stay unchanged
    And the table "attempts" should stay unchanged

  Scenario: User not found
    Given I am the user with id "404"
    And the following token "priorUserTaskToken" signed by the app is distributed:
      """
      {
        "idUser": "404",
        "idItemLocal": "50",
        "idAttempt": "101/0",
        "itemURL": "http://taskplatform.mblockelet.info/task.html?taskId=403449543672183936",
        "platformName": "{{app().TokenConfig.PlatformName}}"
      }
      """
    And the following token "scoreToken" signed by the task platform is distributed:
      """
      {
        "idUser": "404",
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
    Then the response code should be 401
    And the response error message should contain "Invalid access token"
    And the table "answers" should stay unchanged
    And the table "attempts" should stay unchanged

  Scenario: idUser in task_token doesn't match the user's id
    Given I am the user with id "101"
    And the following token "priorUserTaskToken" signed by the app is distributed:
      """
      {
        "idUser": "20",
        "idItemLocal": "50",
        "idAttempt": "101/0",
        "itemURL": "http://taskplatform.mblockelet.info/task.html?taskId=403449543672183936",
        "platformName": "{{app().TokenConfig.PlatformName}}"
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
    Then the response code should be 400
    And the response error message should contain "Token in task_token doesn't correspond to user session: got idUser=20, expected 10"
    And the table "answers" should stay unchanged
    And the table "attempts" should stay unchanged

  Scenario: idUser in score_token doesn't match the user's id
    Given I am the user with id "101"
    And the following token "priorUserTaskToken" signed by the app is distributed:
      """
      {
        "idUser": "101",
        "idItemLocal": "50",
        "idAttempt": "101/0",
        "itemURL": "http://taskplatform.mblockelet.info/task.html?taskId=403449543672183936",
        "platformName": "{{app().TokenConfig.PlatformName}}"
      }
      """
    And the following token "scoreToken" signed by the task platform is distributed:
      """
      {
        "idUser": "20",
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
    Then the response code should be 400
    And the response error message should contain "Token in score_token doesn't correspond to user session: got idUser=20, expected 10"
    And the table "answers" should stay unchanged
    And the table "attempts" should stay unchanged

  Scenario: idAttempt in score_token and task_token don't match
    Given I am the user with id "101"
    And the following token "priorUserTaskToken" signed by the app is distributed:
      """
      {
        "idUser": "101",
        "idItemLocal": "50",
        "idAttempt": "101/0",
        "itemURL": "http://taskplatform.mblockelet.info/task.html?taskId=403449543672183936",
        "platformName": "{{app().TokenConfig.PlatformName}}"
      }
      """
    And the following token "scoreToken" signed by the task platform is distributed:
      """
      {
        "idUser": "101",
        "idItemLocal": "50",
        "itemUrl": "http://taskplatform.mblockelet.info/task.html?taskId=403449543672183936",
        "idAttempt": "101/1",
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
    Then the response code should be 400
    And the response error message should contain "Wrong idAttempt in score_token"
    And the table "answers" should stay unchanged
    And the table "attempts" should stay unchanged

  Scenario: idItemLocal in score_token and task_token don't match
    Given I am the user with id "101"
    And the following token "priorUserTaskToken" signed by the app is distributed:
      """
      {
        "idUser": "101",
        "idItemLocal": "50",
        "idAttempt": "101/0",
        "itemURL": "http://taskplatform.mblockelet.info/task.html?taskId=403449543672183936",
        "platformName": "{{app().TokenConfig.PlatformName}}"
      }
      """
    And the following token "scoreToken" signed by the task platform is distributed:
      """
      {
        "idUser": "101",
        "idItemLocal": "51",
        "itemUrl": "http://taskplatform.mblockelet.info/task.html?taskId=403449543672183936",
        "idAttempt": "101/0",
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
    Then the response code should be 400
    And the response error message should contain "Wrong idItemLocal in score_token"
    And the table "answers" should stay unchanged
    And the table "attempts" should stay unchanged

  Scenario: itemUrl of score_token doesn't match itemUrl of task_token
    Given I am the user with id "101"
    And the following token "priorUserTaskToken" signed by the app is distributed:
      """
      {
        "idUser": "101",
        "idItemLocal": "50",
        "idAttempt": "101/0",
        "itemURL": "http://taskplatform.mblockelet.info/task.html?taskId=403449543672183936",
        "platformName": "{{app().TokenConfig.PlatformName}}"
      }
      """
    And the following token "scoreToken" signed by the task platform is distributed:
      """
      {
        "idUser": "101",
        "idItemLocal": "50",
        "idAttempt": "101/0",
        "itemUrl": "http://taskplatform.mblockelet.info/task.html?taskId=403449543672183937",
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
    Then the response code should be 400
    And the response error message should contain "Wrong itemUrl in score_token"
    And the table "answers" should stay unchanged
    And the table "attempts" should stay unchanged

  Scenario: Missing task_token
    Given I am the user with id "101"
    And the following token "scoreToken" signed by the task platform is distributed:
      """
      {
        "idUser": "101",
        "idItemLocal": "50",
        "idAttempt": "101/0",
        "itemUrl": "http://taskplatform.mblockelet.info/task.html?taskId=403449543672183937",
        "score": "100",
        "idUserAnswer": "123"
      }
      """
    When I send a POST request to "/items/save-grade" with the following body:
      """
      {
        "score_token": "{{scoreToken}}"
      }
      """
    Then the response code should be 400
    And the response error message should contain "Missing task_token"
    And the table "answers" should stay unchanged
    And the table "attempts" should stay unchanged

  Scenario: Invalid task_token
    Given I am the user with id "101"
    And the following token "scoreToken" signed by the task platform is distributed:
      """
      {
        "idUser": "101",
        "idItemLocal": "50",
        "idAttempt": "101/0",
        "itemUrl": "http://taskplatform.mblockelet.info/task.html?taskId=403449543672183937",
        "score": "100",
        "idUserAnswer": "123"
      }
      """
    When I send a POST request to "/items/save-grade" with the following body:
      """
      {
        "task_token": "abcdef",
        "score_token": "{{scoreToken}}"
      }
      """
    Then the response code should be 400
    And the response error message should contain "Invalid task_token"
    And the table "answers" should stay unchanged
    And the table "attempts" should stay unchanged

  Scenario: Invalid score_token
    Given I am the user with id "101"
    And the following token "priorUserTaskToken" signed by the app is distributed:
      """
      {
        "idUser": "101",
        "idItemLocal": "50",
        "idAttempt": "101/0",
        "itemURL": "http://taskplatform.mblockelet.info/task.html?taskId=403449543672183936",
        "platformName": "{{app().TokenConfig.PlatformName}}"
      }
      """
    When I send a POST request to "/items/save-grade" with the following body:
      """
      {
        "task_token": "{{priorUserTaskToken}}",
        "score_token": "abcdef"
      }
      """
    Then the response code should be 400
    And the response error message should contain "Invalid score_token"
    And the table "answers" should stay unchanged
    And the table "attempts" should stay unchanged

  Scenario: Platform doesn't use tokens and answer_token is missing
    Given I am the user with id "101"
    And the following token "priorUserTaskToken" signed by the app is distributed:
      """
      {
        "idUser": "101",
        "idItemLocal": "70",
        "idAttempt": "101/0",
        "itemURL": "http://taskplatform1.mblockelet.info/task.html?taskId=4034495436721839",
        "platformName": "{{app().TokenConfig.PlatformName}}"
      }
      """
    When I send a POST request to "/items/save-grade" with the following body:
      """
      {
        "task_token": "{{priorUserTaskToken}}",
        "score": 100.0
      }
      """
    Then the response code should be 400
    And the response error message should contain "Missing answer_token"
    And the table "answers" should stay unchanged
    And the table "attempts" should stay unchanged

  Scenario: Platform doesn't use tokens and answer_token is invalid
    Given I am the user with id "101"
    And the following token "priorUserTaskToken" signed by the app is distributed:
      """
      {
        "idUser": "101",
        "idItemLocal": "70",
        "idAttempt": "101/0",
        "itemURL": "http://taskplatform1.mblockelet.info/task.html?taskId=4034495436721839",
        "platformName": "{{app().TokenConfig.PlatformName}}"
      }
      """
    When I send a POST request to "/items/save-grade" with the following body:
      """
      {
        "task_token": "{{priorUserTaskToken}}",
        "score": 100.0,
        "answer_token": "abc"
      }
      """
    Then the response code should be 400
    And the response error message should contain "Invalid answer_token"
    And the table "answers" should stay unchanged
    And the table "attempts" should stay unchanged

  Scenario: Platform doesn't use tokens and idUser in answer_token is wrong
    Given I am the user with id "101"
    And the following token "priorUserTaskToken" signed by the app is distributed:
      """
      {
        "idUser": "101",
        "idItemLocal": "70",
        "idAttempt": "101/0",
        "itemURL": "http://taskplatform1.mblockelet.info/task.html?taskId=4034495436721839",
        "platformName": "{{app().TokenConfig.PlatformName}}"
      }
      """
    And the following token "answerToken" signed by the app is distributed:
      """
      {
        "idUser": "20",
        "idItemLocal": "70",
        "idAttempt": "101/0",
        "itemURL": "http://taskplatform1.mblockelet.info/task.html?taskId=4034495436721839",
        "idUserAnswer": "123",
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
    Then the response code should be 400
    And the response error message should contain "Wrong idUser in answer_token"
    And the table "answers" should stay unchanged
    And the table "attempts" should stay unchanged

  Scenario: Platform doesn't use tokens and idItemLocal in answer_token is wrong
    Given I am the user with id "101"
    And the following token "priorUserTaskToken" signed by the app is distributed:
      """
      {
        "idUser": "101",
        "idItemLocal": "70",
        "idAttempt": "101/0",
        "itemURL": "http://taskplatform1.mblockelet.info/task.html?taskId=4034495436721839",
        "platformName": "{{app().TokenConfig.PlatformName}}"
      }
      """
    And the following token "answerToken" signed by the app is distributed:
      """
      {
        "idUser": "101",
        "idItemLocal": "60",
        "idAttempt": "101/0",
        "itemURL": "http://taskplatform1.mblockelet.info/task.html?taskId=4034495436721839",
        "idUserAnswer": "123",
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
    Then the response code should be 400
    And the response error message should contain "Wrong idItemLocal in answer_token"
    And the table "answers" should stay unchanged
    And the table "attempts" should stay unchanged

  Scenario: Platform doesn't use tokens and itemUrl in answer_token is wrong
    Given I am the user with id "101"
    And the following token "priorUserTaskToken" signed by the app is distributed:
      """
      {
        "idUser": "101",
        "idItemLocal": "70",
        "idAttempt": "101/0",
        "itemURL": "http://taskplatform1.mblockelet.info/task.html?taskId=4034495436721839",
        "platformName": "{{app().TokenConfig.PlatformName}}"
      }
      """
    And the following token "answerToken" signed by the app is distributed:
      """
      {
        "idUser": "101",
        "idItemLocal": "70",
        "idAttempt": "101/0",
        "itemURL": "http://taskplatform1.mblockelet.info/task.html?taskId=403449543672183",
        "idUserAnswer": "123",
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
    Then the response code should be 400
    And the response error message should contain "Wrong itemUrl in answer_token"
    And the table "answers" should stay unchanged
    And the table "attempts" should stay unchanged

  Scenario: Platform doesn't use tokens and idAttempt in answer_token is wrong (should not be null)
    Given I am the user with id "101"
    And the following token "priorUserTaskToken" signed by the app is distributed:
      """
      {
        "idUser": "101",
        "idItemLocal": "70",
        "itemURL": "http://taskplatform1.mblockelet.info/task.html?taskId=4034495436721839",
        "idAttempt": "101/0",
        "platformName": "{{app().TokenConfig.PlatformName}}"
      }
      """
    And the following token "answerToken" signed by the app is distributed:
      """
      {
        "idUser": "101",
        "idItemLocal": "70",
        "itemURL": "http://taskplatform1.mblockelet.info/task.html?taskId=4034495436721839",
        "idUserAnswer": "123",
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
    Then the response code should be 400
    And the response error message should contain "Wrong idAttempt in answer_token"
    And the table "answers" should stay unchanged
    And the table "attempts" should stay unchanged

  Scenario: Platform doesn't use tokens and idAttempt in answer_token is wrong (should be equal)
    Given I am the user with id "101"
    And the following token "priorUserTaskToken" signed by the app is distributed:
      """
      {
        "idUser": "101",
        "idItemLocal": "70",
        "itemURL": "http://taskplatform1.mblockelet.info/task.html?taskId=4034495436721839",
        "idAttempt": "101/0",
        "platformName": "{{app().TokenConfig.PlatformName}}"
      }
      """
    And the following token "answerToken" signed by the app is distributed:
      """
      {
        "idUser": "101",
        "idItemLocal": "70",
        "itemURL": "http://taskplatform1.mblockelet.info/task.html?taskId=4034495436721839",
        "idAttempt": "110/0",
        "idUserAnswer": "123",
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
    Then the response code should be 400
    And the response error message should contain "Wrong idAttempt in answer_token"
    And the table "answers" should stay unchanged
    And the table "attempts" should stay unchanged

  Scenario: Platform doesn't use tokens and score is missing
    Given I am the user with id "101"
    And the following token "priorUserTaskToken" signed by the app is distributed:
      """
      {
        "idUser": "101",
        "idItemLocal": "70",
        "idAttempt": "101/0",
        "itemURL": "http://taskplatform1.mblockelet.info/task.html?taskId=4034495436721839",
        "platformName": "{{app().TokenConfig.PlatformName}}"
      }
      """
    And the following token "answerToken" signed by the app is distributed:
      """
      {
        "idUser": "101",
        "idItemLocal": "70",
        "idAttempt": "101/0",
        "itemURL": "http://taskplatform1.mblockelet.info/task.html?taskId=4034495436721839",
        "idUserAnswer": "123",
        "platformName": "{{app().TokenConfig.PlatformName}}"
      }
      """
    When I send a POST request to "/items/save-grade" with the following body:
      """
      {
        "task_token": "{{priorUserTaskToken}}",
        "answer_token": "{{answerToken}}"
      }
      """
    Then the response code should be 400
    And the response error message should contain "Missing score"
    And the table "answers" should stay unchanged
    And the table "attempts" should stay unchanged

  Scenario: Platform doesn't use tokens and idUserAnswer in answer_token is invalid
    Given I am the user with id "101"
    And the following token "priorUserTaskToken" signed by the app is distributed:
      """
      {
        "idUser": "101",
        "idItemLocal": "70",
        "idAttempt": "101/0",
        "itemURL": "http://taskplatform1.mblockelet.info/task.html?taskId=4034495436721839",
        "platformName": "{{app().TokenConfig.PlatformName}}"
      }
      """
    And the following token "answerToken" signed by the app is distributed:
      """
      {
        "idUser": "101",
        "idItemLocal": "70",
        "idAttempt": "101/0",
        "itemURL": "http://taskplatform1.mblockelet.info/task.html?taskId=4034495436721839",
        "idUserAnswer": "abc",
        "platformName": "{{app().TokenConfig.PlatformName}}"
      }
      """
    When I send a POST request to "/items/save-grade" with the following body:
      """
      {
        "task_token": "{{priorUserTaskToken}}",
        "answer_token": "{{answerToken}}",
        "score": 99.0
      }
      """
    Then the response code should be 400
    And the response error message should contain "Invalid idUserAnswer in answer_token"
    And the table "answers" should stay unchanged
    And the table "attempts" should stay unchanged

  Scenario: The answer has been already graded
    Given I am the user with id "101"
    And the database table 'attempts' has also the following row:
      | id | participant_id |
      | 1  | 101            |
    And the database table 'results' has also the following row:
      | attempt_id | participant_id | item_id | validated_at        |
      | 1          | 101            | 80      | 2018-05-29 06:38:38 |
    And the database has the following table 'answers':
      | id  | author_id | participant_id | attempt_id | item_id | created_at          |
      | 124 | 101       | 101            | 105        | 80      | 2017-05-29 06:38:38 |
    And the database has the following table 'gradings':
      | answer_id | score | graded_at           |
      | 124       | 0     | 2017-05-29 06:38:38 |
    And the following token "priorUserTaskToken" signed by the app is distributed:
      """
      {
        "idUser": "101",
        "idItemLocal": "80",
        "idAttempt": "101/1",
        "itemURL": "http://taskplatform.mblockelet.info/task.html?taskId=403449543672183937",
        "bAccessSolutions": false,
        "platformName": "{{app().TokenConfig.PlatformName}}"
      }
      """
    And the following token "scoreToken" signed by the task platform is distributed:
      """
      {
        "idUser": "101",
        "idItemLocal": "80",
        "idAttempt": "101/1",
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
    Then the response code should be 403
    And the response error message should contain "The answer has been already graded or is not found"
    And logs should contain:
    """
    A user tries to replay a score token with a different score value ({"idAttempt":"101/1","idItem":"80","idUser":"101","idUserAnswer":"124","newScore":100,"oldScore":0})
    """
    And the table "answers" should stay unchanged
    And the table "attempts" should stay unchanged

  Scenario: The answer is not found
    Given I am the user with id "101"
    And the following token "priorUserTaskToken" signed by the app is distributed:
      """
      {
        "idUser": "101",
        "idItemLocal": "80",
        "idAttempt": "101/0",
        "itemURL": "http://taskplatform.mblockelet.info/task.html?taskId=403449543672183937",
        "bAccessSolutions": false,
        "platformName": "{{app().TokenConfig.PlatformName}}"
      }
      """
    And the following token "scoreToken" signed by the task platform is distributed:
      """
      {
        "idUser": "101",
        "idItemLocal": "80",
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
    Then the response code should be 403
    And the response error message should contain "The answer has been already graded or is not found"
    And the table "answers" should stay unchanged
    And the table "attempts" should stay unchanged
