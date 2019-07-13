Feature: Submit a new answer - robustness
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
    And the database has the following table 'items':
      | ID | bReadOnly |
      | 50 | 1         |
    And the database has the following table 'groups_items':
      | idGroup | idItem | sCachedPartialAccessDate |
      | 101     | 50     | 2017-05-29T06:38:38Z     |
    And the database has the following table 'users_items':
      | idUser | idItem | sHintsRequested                 | nbHintsCached |
      | 10     | 50     | [{"rotorIndex":0,"cellRank":0}] | 12            |

  Scenario: Wrong JSON in request
    Given I am the user with ID "10"
    When I send a POST request to "/answers" with the following body:
      """
      []
      """
    Then the response code should be 400
    And the response error message should contain "Json: cannot unmarshal array into Go value of type answers.submitRequestWrapper"
    And the table "users_items" should stay unchanged
    And the table "users_answers" should stay unchanged

  Scenario: No task_token
    Given I am the user with ID "10"
    When I send a POST request to "/answers" with the following body:
      """
      {
        "answer": "print 1"
      }
      """
    Then the response code should be 400
    And the response error message should contain "Missing task_token"
    And the table "users_items" should stay unchanged
    And the table "users_answers" should stay unchanged

  Scenario: Wrong task_token
    Given I am the user with ID "10"
    When I send a POST request to "/answers" with the following body:
      """
      {
        "task_token": "ADSFADQER.ASFDAS.ASDFSDA",
        "answer": "print 1"
      }
      """
    Then the response code should be 400
    And the response error message should contain "Invalid task_token: illegal base64 data at input byte 8"
    And the table "users_items" should stay unchanged
    And the table "users_answers" should stay unchanged

  Scenario: Missing answer
    Given I am the user with ID "10"
    And the following token "userTaskToken" signed by the app is distributed:
      """
      {
        "idUser": "10",
        "idAttempt": "100",
        "idItemLocal": "50",
        "platformName": "{{app().TokenConfig.PlatformName}}"
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
    And the table "users_items" should stay unchanged
    And the table "users_answers" should stay unchanged

  Scenario: Wrong idUser
    Given I am the user with ID "10"
    And the following token "userTaskToken" signed by the app is distributed:
      """
      {
        "idUser": "",
        "idAttempt": "100",
        "idItemLocal": "50",
        "platformName": "{{app().TokenConfig.PlatformName}}"
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
    And the table "users_items" should stay unchanged
    And the table "users_answers" should stay unchanged

  Scenario: Wrong idItemLocal
    Given I am the user with ID "10"
    And the following token "userTaskToken" signed by the app is distributed:
      """
      {
        "idUser": "10",
        "idAttempt": "100",
        "platformName": "{{app().TokenConfig.PlatformName}}"
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
    And the table "users_items" should stay unchanged
    And the table "users_answers" should stay unchanged

  Scenario: Wrong idAttempt
    Given I am the user with ID "10"
    And the following token "userTaskToken" signed by the app is distributed:
      """
      {
        "idUser": "10",
        "idItemLocal": "50",
        "idAttempt": "abc",
        "platformName": "{{app().TokenConfig.PlatformName}}"
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
    And the table "users_items" should stay unchanged
    And the table "users_answers" should stay unchanged

  Scenario: idUser doesn't match the user's ID
    Given I am the user with ID "10"
    And the following token "userTaskToken" signed by the app is distributed:
      """
      {
        "idUser": "20",
        "idItemLocal": "50",
        "idAttempt": "100",
        "platformName": "{{app().TokenConfig.PlatformName}}"
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
    And the response error message should contain "Token doesn't correspond to user session: got idUser=20, expected 10"
    And the table "users_items" should stay unchanged
    And the table "users_answers" should stay unchanged

  Scenario: User not found
    Given I am the user with ID "404"
    And the following token "userTaskToken" signed by the app is distributed:
      """
      {
        "idUser": "404",
        "idItemLocal": "50",
        "idAttempt": "100",
        "platformName": "{{app().TokenConfig.PlatformName}}"
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
    And the table "users_items" should stay unchanged
    And the table "users_answers" should stay unchanged

  Scenario: No submission rights
    Given I am the user with ID "10"
    And the following token "userTaskToken" signed by the app is distributed:
      """
      {
        "idUser": "10",
        "idItemLocal": "50",
        "idAttempt": "100",
        "platformName": "{{app().TokenConfig.PlatformName}}"
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
    And the table "users_items" should stay unchanged
    And the table "users_answers" should stay unchanged
