Feature: Save grading result - security
  Background:
    Given the database has the following user:
      | login | group_id | default_language |
      | john  | 101      | en               |
    And the database has the following table "platforms":
      | id | regexp                                             | priority | public_key                |
      | 10 | http://taskplatform.mblockelet.info/task.html\?.*  | 2        | {{taskPlatformPublicKey}} |
      | 20 | http://taskplatform1.mblockelet.info/task.html\?.* | 1        | null                      |
    And the database has the following table "items":
      | id | platform_id | url                                                                     | validation_type | default_language_tag |
      | 50 | 10          | http://taskplatform.mblockelet.info/task.html?taskId=403449543672183936 | All             | fr                   |
      | 60 | 10          | http://taskplatform.mblockelet.info/task.html?taskId=403449543672183937 | All             | fr                   |
      | 10 | null        | null                                                                    | AllButOne       | fr                   |
      | 70 | 20          | http://taskplatform1.mblockelet.info/task.html?taskId=4034495436721839  | All             | fr                   |
    And the database has the following table "items_strings":
      | item_id | language_tag | title        |
      | 50      | fr           | Chapitre A   |
      | 50      | en           | Chapter A    |
    And the database has the following table "item_dependencies":
      | item_id | dependent_item_id | score |
      | 60      | 50                | 98    |
    And the database has the following table "items_items":
      | parent_item_id | child_item_id | child_order |
      | 10             | 50            | 0           |
      | 10             | 60            | 1           |
    And the database has the following table "items_ancestors":
      | ancestor_item_id | child_item_id |
      | 10               | 50            |
      | 10               | 60            |
    And the database has the following table "permissions_generated":
      | group_id | item_id | can_view_generated |
      | 101      | 50      | content            |
      | 101      | 60      | content            |
      | 101      | 70      | content            |
      | 201      | 50      | content            |
      | 201      | 60      | content            |
      | 201      | 70      | content            |
    And the server time is frozen

  Scenario: answer_token should be ignored if the score_token is present and the platform has a public key
    Given the database has the following table "attempts":
      | id | participant_id |
      | 0  | 101            |
      | 1  | 101            |
    And the database has the following table "results":
      | attempt_id | participant_id | item_id | latest_activity_at  | hints_requested        |
      | 0          | 101            | 10      | 2019-05-30 11:00:00 | null                   |
      | 0          | 101            | 50      | 2019-05-30 11:00:00 | [0,  1, "hint" , null] |
      | 1          | 101            | 60      | 2019-05-29 11:00:00 | [0,  1, "hint" , null] |
    And the database has the following table "answers":
      | id  | author_id | participant_id | attempt_id | item_id | created_at          |
      | 123 | 101       | 101            | 0          | 50      | 2017-05-29 06:38:38 |
      | 124 | 101       | 101            | 0          | 60      | 2017-05-29 06:38:38 |
    And "scoreToken" is a token signed by the task platform with the following payload:
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
    And "answerToken" is a token signed by the app with the following payload:
      """
      {
        "idUser": "102",
        "idItemLocal": "60",
        "idAttempt": "101/1",
        "itemURL": "http://taskplatform.mblockelet.info/task.html?taskId=403449543672183936",
        "idUserAnswer": "125",
        "platformName": "{{app().Config.GetString("token.platformName")}}"
      }
      """
    When I send a POST request to "/items/save-grade" with the following body:
      """
      {
        "score_token": "{{scoreToken}}",
        "answer_token": "{{answerToken}}"
      }
      """
    Then the response code should be 201
    And the response body should be, in JSON:
      """
      {
        "data": {
          "validated": true,
          "unlocked_items": []
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
      | attempt_id | participant_id | item_id | score_computed | tasks_tried | validated | latest_activity_at  | latest_submission_at | score_obtained_at   | validated_at        |
      | 0          | 101            | 10      | 50             | 1           | 1         | 2019-05-30 11:00:00 | null                 | null                | 2017-05-29 06:38:38 |
      | 0          | 101            | 50      | 100            | 1           | 1         | 2019-05-30 11:00:00 | null                 | 2017-05-29 06:38:38 | 2017-05-29 06:38:38 |
      | 1          | 101            | 60      | 0              | 0           | 0         | 2019-05-29 11:00:00 | null                 | null                | null                |
    And the table "results_propagate" should be empty
    And the table "results_propagate_sync" should be empty

  Scenario: score should be ignored if the score_token is present and the platform has a public key
    Given the database has the following table "attempts":
      | id | participant_id |
      | 0  | 101            |
      | 1  | 101            |
    And the database has the following table "results":
      | attempt_id | participant_id | item_id | latest_activity_at  | hints_requested        |
      | 0          | 101            | 10      | 2019-05-30 11:00:00 | null                   |
      | 0          | 101            | 50      | 2019-05-30 11:00:00 | [0,  1, "hint" , null] |
      | 1          | 101            | 60      | 2019-05-29 11:00:00 | [0,  1, "hint" , null] |
    And the database has the following table "answers":
      | id  | author_id | participant_id | attempt_id | item_id | created_at          |
      | 123 | 101       | 101            | 0          | 50      | 2017-05-29 06:38:38 |
      | 124 | 101       | 101            | 0          | 60      | 2017-05-29 06:38:38 |
    And "scoreToken" is a token signed by the task platform with the following payload:
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
        "score_token": "{{scoreToken}}",
        "score": 101.0
      }
      """
    Then the response code should be 201
    And the response body should be, in JSON:
      """
      {
        "data": {
          "validated": true,
          "unlocked_items": []
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
      | attempt_id | participant_id | item_id | score_computed | tasks_tried | validated | latest_activity_at  | latest_submission_at | score_obtained_at   | validated_at        |
      | 0          | 101            | 10      | 50             | 1           | 1         | 2019-05-30 11:00:00 | null                 | null                | 2017-05-29 06:38:38 |
      | 0          | 101            | 50      | 100            | 1           | 1         | 2019-05-30 11:00:00 | null                 | 2017-05-29 06:38:38 | 2017-05-29 06:38:38 |
      | 1          | 101            | 60      | 0              | 0           | 0         | 2019-05-29 11:00:00 | null                 | null                | null                |
    And the table "results_propagate" should be empty
    And the table "results_propagate_sync" should be empty

  Scenario: Should fail when score_token is expired
    Given the server time now is "2020-01-01T00:00:00Z"
    And "scoreToken" is a token signed by the task platform with the following payload:
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
    Then the server time now is "2020-01-03T00:00:00Z"
    When I send a POST request to "/items/save-grade" with the following body:
      """
      {
        "score_token": "{{scoreToken}}"
      }
      """
    Then the response code should be 400
    And the response error message should contain "Invalid score_token: the token has expired"
    And the table "gradings" should stay unchanged
    And the table "results" should stay unchanged
    And the table "results_propagate" should be empty
    And the table "results_propagate_sync" should be empty

  Scenario: Should fail on getting a falsified score_token
    Given "scoreToken" is a falsified token signed by the task platform with the following payload:
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
        "score_token": "{{scoreToken}}"
      }
      """
    Then the response code should be 400
    And the response error message should contain "Invalid score_token: invalid token: crypto/rsa: verification error"
    And the table "gradings" should stay unchanged
    And the table "results" should stay unchanged
    And the table "results_propagate" should be empty
    And the table "results_propagate_sync" should be empty

  Scenario: Should fail when the platform doesn't use tokens and answer_token is expired
    Given the server time now is "2020-01-01T00:00:00Z"
    And "answerToken" is a token signed by the app with the following payload:
      """
      {
        "idUser": "101",
        "idItemLocal": "70",
        "itemURL": "http://taskplatform1.mblockelet.info/task.html?taskId=4034495436721839",
        "idUserAnswer": "123",
        "idAttempt": "101/0",
        "platformName": "{{app().Config.GetString("token.platformName")}}"
      }
      """
    Then the server time now is "2020-01-03T00:00:00Z"
    When I send a POST request to "/items/save-grade" with the following body:
      """
      {
        "score": 100.0,
        "answer_token": "{{answerToken}}"
      }
      """
    Then the response code should be 400
    And the response error message should contain "Invalid answer_token: the token has expired"
    And the table "gradings" should stay unchanged
    And the table "results" should stay unchanged
    And the table "results_propagate" should be empty
    And the table "results_propagate_sync" should be empty

  Scenario: Should fail when the platform doesn't use tokens and answer_token is falsified
    Given "answerToken" is a falsified token signed by the app with the following payload:
      """
      {
        "idUser": "101",
        "idItemLocal": "70",
        "itemURL": "http://taskplatform1.mblockelet.info/task.html?taskId=4034495436721839",
        "idUserAnswer": "123",
        "platformName": "{{app().Config.GetString("token.platformName")}}"
      }
      """
    When I send a POST request to "/items/save-grade" with the following body:
      """
      {
        "score": 100.0,
        "answer_token": "{{answerToken}}"
      }
      """
    Then the response code should be 400
    And the response error message should contain "Invalid answer_token: invalid token: crypto/rsa: verification error"
    And the table "gradings" should stay unchanged
    And the table "results" should stay unchanged
    And the table "results_propagate" should be empty
    And the table "results_propagate_sync" should be empty

  Scenario: Should fail when the platform uses tokens, but the score_token is not given
    Given the database table "attempts" also has the following row:
      | id | participant_id |
      | 1  | 101            |
    And the database table "answers" also has the following row:
      | id  | author_id | participant_id | attempt_id | item_id | created_at          |
      | 125 | 101       | 101            | 0          | 50      | 2017-05-29 06:38:38 |
    And "answerToken" is a token signed by the app with the following payload:
      """
      {
        "idUser": "101",
        "idItemLocal": "50",
        "idAttempt": "101/0",
        "itemURL": "http://taskplatform.mblockelet.info/task.html?taskId=403449543672183936",
        "idUserAnswer": "125",
        "platformName": "{{app().Config.GetString("token.platformName")}}"
      }
      """
    When I send a POST request to "/items/save-grade" with the following body:
      """
      {
        "score": 100.0,
        "answer_token": "{{answerToken}}"
      }
      """
    Then the response code should be 400
    And the response error message should contain "The platform has a public key, but the score_token is not given"
    And the table "results" should stay unchanged
    And the table "gradings" should stay unchanged
    And the table "results_propagate" should be empty
    And the table "results_propagate_sync" should be empty

  Scenario: Should fail when the platform doesn't use tokens, but the score_token is given
    Given the database has the following table "attempts":
      | id | participant_id |
      | 1  | 101            |
    And the database has the following table "results":
      | attempt_id | participant_id | item_id | validated_at        |
      | 1          | 101            | 70      | 2018-05-29 06:38:38 |
    And the database has the following table "answers":
      | id  | author_id | participant_id | attempt_id | item_id | created_at          |
      | 125 | 101       | 101            | 100        | 70      | 2017-05-29 06:38:38 |
    And "answerToken" is a token signed by the app with the following payload:
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
    And "scoreToken" is a token signed by the task platform with the following payload:
      """
      {
        "idUser": "101",
        "idItemLocal": "70",
        "idAttempt": "101/1",
        "itemURL": "http://taskplatform1.mblockelet.info/task.html?taskId=4034495436721839",
        "score": "99",
        "idUserAnswer": "125"
      }
      """
    When I send a POST request to "/items/save-grade" with the following body:
      """
      {
        "score_token": "{{scoreToken}}",
        "score": 100.0,
        "answer_token": "{{answerToken}}"
      }
      """
    Then the response code should be 400
    And the response error message should contain "The platform does not have a public key, but the score_token is given"
    And the table "results" should stay unchanged
    And the table "gradings" should stay unchanged
    And the table "results_propagate" should be empty
    And the table "results_propagate_sync" should be empty
