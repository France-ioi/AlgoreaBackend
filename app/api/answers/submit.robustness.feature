Feature: Submit a new answer - robustness
  Background:
    Given the database has the following users:
      | login | group_id |
      | john  | 101      |
    And the database has the following table 'groups_groups':
      | parent_group_id | child_group_id |
      | 22              | 13             |
    And the groups ancestors are computed
    And the database has the following table 'items':
      | id | read_only | default_language_tag |
      | 50 | 1         | fr                   |
      | 60 | 0         | fr                   |
    And the database has the following table 'permissions_generated':
      | group_id | item_id | can_view_generated |
      | 101      | 50      | content            |
      | 101      | 60      | content            |
    And the database has the following table 'attempts':
      | participant_id | id | allows_submissions_until |
      | 101            | 1  | 2019-05-30 11:00:00      |
    And the database has the following table 'results':
      | participant_id | attempt_id | item_id |
      | 101            | 1          | 60      |

  Scenario: Wrong JSON in request
    Given I am the user with id "101"
    When I send a POST request to "/answers" with the following body:
      """
      []
      """
    Then the response code should be 400
    And the response error message should contain "Json: cannot unmarshal array into Go value of type answers.submitRequestWrapper"
    And the table "answers" should stay unchanged

  Scenario: No task_token
    Given I am the user with id "101"
    When I send a POST request to "/answers" with the following body:
      """
      {
        "answer": "print 1"
      }
      """
    Then the response code should be 400
    And the response error message should contain "Missing task_token"
    And the table "answers" should stay unchanged

  Scenario: Wrong task_token
    Given I am the user with id "101"
    When I send a POST request to "/answers" with the following body:
      """
      {
        "task_token": "ADSFADQER.ASFDAS.ASDFSDA",
        "answer": "print 1"
      }
      """
    Then the response code should be 400
    And the response error message should contain "Invalid task_token: illegal base64 data at input byte 8"
    And the table "answers" should stay unchanged

  Scenario: Missing answer
    Given I am the user with id "101"
    And "userTaskToken" is a token signed by the app with the following payload:
      """
      {
        "idUser": "101",
        "idAttempt": "100/2",
        "idItemLocal": "50",
        "platformName": "{{app().Config.GetString("token.platformName")}}"
      }
      """
    When I send a POST request to "/answers" with the following body:
      """
      {
        "task_token": "{{userTaskToken}}"
      }
      """
    Then the response code should be 400
    And the response error message should contain "Missing answer"
    And the table "answers" should stay unchanged

  Scenario: Wrong idUser
    Given I am the user with id "101"
    And "userTaskToken" is a token signed by the app with the following payload:
      """
      {
        "idUser": "",
        "idAttempt": "100",
        "idItemLocal": "50",
        "platformName": "{{app().Config.GetString("token.platformName")}}"
      }
      """
    When I send a POST request to "/answers" with the following body:
      """
      {
        "task_token": "{{userTaskToken}}",
        "answer": "print(1)"
      }
      """
    Then the response code should be 400
    And the response error message should contain "Invalid task_token: wrong idUser"
    And the table "answers" should stay unchanged

  Scenario: Wrong idItemLocal
    Given I am the user with id "101"
    And "userTaskToken" is a token signed by the app with the following payload:
      """
      {
        "idUser": "101",
        "idAttempt": "100",
        "platformName": "{{app().Config.GetString("token.platformName")}}"
      }
      """
    When I send a POST request to "/answers" with the following body:
      """
      {
        "task_token": "{{userTaskToken}}",
        "answer": "print(1)"
      }
      """
    Then the response code should be 400
    And the response error message should contain "Invalid task_token: wrong idItemLocal"
    And the table "answers" should stay unchanged

  Scenario: Wrong idAttempt
    Given I am the user with id "101"
    And "userTaskToken" is a token signed by the app with the following payload:
      """
      {
        "idUser": "101",
        "idItemLocal": "50",
        "idAttempt": "abc",
        "platformName": "{{app().Config.GetString("token.platformName")}}"
      }
      """
    When I send a POST request to "/answers" with the following body:
      """
      {
        "task_token": "{{userTaskToken}}",
        "answer": "print(1)"
      }
      """
    Then the response code should be 400
    And the response error message should contain "Invalid task_token: wrong idAttempt"
    And the table "answers" should stay unchanged

  Scenario: idUser doesn't match the user's group id
    Given I am the user with id "101"
    And "userTaskToken" is a token signed by the app with the following payload:
      """
      {
        "idUser": "20",
        "idItemLocal": "50",
        "idAttempt": "100/1",
        "platformName": "{{app().Config.GetString("token.platformName")}}"
      }
      """
    When I send a POST request to "/answers" with the following body:
      """
      {
        "task_token": "{{userTaskToken}}",
        "answer": "print(1)"
      }
      """
    Then the response code should be 400
    And the response error message should contain "Token doesn't correspond to user session: got idUser=20, expected 101"
    And the table "answers" should stay unchanged

  Scenario: User not found
    Given I am the user with id "404"
    And "userTaskToken" is a token signed by the app with the following payload:
      """
      {
        "idUser": "404",
        "idItemLocal": "50",
        "idAttempt": "100/1",
        "platformName": "{{app().Config.GetString("token.platformName")}}"
      }
      """
    When I send a POST request to "/answers" with the following body:
      """
      {
        "task_token": "{{userTaskToken}}",
        "answer": "print(1)"
      }
      """
    Then the response code should be 401
    And the response error message should contain "Invalid access token"
    And the table "answers" should stay unchanged

  Scenario: No submission rights
    Given I am the user with id "101"
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
        "answer": "print(1)"
      }
      """
    Then the response code should be 403
    And the response error message should contain "Item is read-only"
    And the table "answers" should stay unchanged

  Scenario: The attempt is expired (doesn't allow submissions anymore)
    Given I am the user with id "101"
    And "userTaskToken" is a token signed by the app with the following payload:
      """
      {
        "idUser": "101",
        "idItemLocal": "60",
        "idAttempt": "101/1",
        "platformName": "{{app().Config.GetString("token.platformName")}}"
      }
      """
    When I send a POST request to "/answers" with the following body:
      """
      {
        "task_token": "{{userTaskToken}}",
        "answer": "print(1)"
      }
      """
    Then the response code should be 403
    And the response error message should contain "No active attempt found"
    And the table "answers" should stay unchanged
