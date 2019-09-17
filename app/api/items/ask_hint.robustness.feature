Feature: Ask for a hint - robustness
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
      | ID | bUsesTokens | sRegexp                                           | sPublicKey                |
      | 10 | 1           | http://taskplatform.mblockelet.info/task.html\?.* | {{taskPlatformPublicKey}} |
    And the database has the following table 'items':
      | ID | idPlatform | sUrl                                                                    | bReadOnly |
      | 50 | 10         | http://taskplatform.mblockelet.info/task.html?taskId=403449543672183936 | 1         |
      | 10 | 10         | http://taskplatform.mblockelet.info/task.html?taskId=403449543672183936 | 0         |
    And the database has the following table 'items_items':
      | idItemParent | idItemChild | iChildOrder |
      | 10           | 50          | 0           |
    And the database has the following table 'items_ancestors':
      | idItemAncestor | idItemChild |
      | 10             | 50          |
    And the database has the following table 'groups_items':
      | idGroup | idItem | sCachedPartialAccessDate | idUserCreated |
      | 101     | 10     | 2017-05-29 06:38:38      | 10            |
      | 101     | 50     | 2017-05-29 06:38:38      | 10            |
    And the database has the following table 'users_items':
      | idUser | idItem | sHintsRequested                 | nbHintsCached | nbSubmissionsAttempts | idAttemptActive |
      | 10     | 10     | null                            | 0             | 0                     | null            |
      | 10     | 50     | [{"rotorIndex":0,"cellRank":0}] | 12            | 2                     | 100             |
    And the database has the following table 'groups_attempts':
      | ID  | idGroup | idItem | sHintsRequested        | iOrder |
      | 100 | 101     | 50     | [0,  1, "hint" , null] | 0      |
    And time is frozen

  Scenario: Wrong JSON in request
    Given I am the user with ID "10"
    When I send a POST request to "/items/ask-hint" with the following body:
      """
      []
      """
    Then the response code should be 400
    And the response error message should contain "Json: cannot unmarshal array into Go value of type items.askHintRequestWrapper"
    And the table "users_items" should stay unchanged
    And the table "groups_attempts" should stay unchanged

  Scenario: User not found
    Given I am the user with ID "404"
    And the following token "priorUserTaskToken" signed by the app is distributed:
      """
      {
        "idUser": "404",
        "idItemLocal": "50",
        "idAttempt": "100",
        "itemURL": "http://taskplatform.mblockelet.info/task.html?taskId=403449543672183936",
        "platformName": "{{app().TokenConfig.PlatformName}}"
      }
      """
    And the following token "hintRequestToken" signed by the task platform is distributed:
      """
      {
        "idUser": "404",
        "idItemLocal": "50",
        "idAttempt": "100",
        "itemUrl": "http://taskplatform.mblockelet.info/task.html?taskId=403449543672183936",
        "askedHint": {"rotorIndex":1}
      }
      """
    When I send a POST request to "/items/ask-hint" with the following body:
      """
      {
        "task_token": "{{priorUserTaskToken}}",
        "hint_requested": "{{hintRequestToken}}"
      }
      """
    Then the response code should be 401
    And the response error message should contain "Invalid access token"
    And the table "users_items" should stay unchanged
    And the table "groups_attempts" should stay unchanged

  Scenario: idUser in task_token doesn't match the user's ID
    Given I am the user with ID "10"
    And the following token "priorUserTaskToken" signed by the app is distributed:
      """
      {
        "idUser": "20",
        "idItemLocal": "50",
        "idAttempt": "100",
        "itemURL": "http://taskplatform.mblockelet.info/task.html?taskId=403449543672183936",
        "platformName": "{{app().TokenConfig.PlatformName}}"
      }
      """
    And the following token "hintRequestToken" signed by the task platform is distributed:
      """
      {
        "idUser": "10",
        "idItemLocal": "50",
        "idAttempt": "100",
        "itemUrl": "http://taskplatform.mblockelet.info/task.html?taskId=403449543672183936",
        "askedHint": {"rotorIndex":1}
      }
      """
    When I send a POST request to "/items/ask-hint" with the following body:
      """
      {
        "task_token": "{{priorUserTaskToken}}",
        "hint_requested": "{{hintRequestToken}}"
      }
      """
    Then the response code should be 400
    And the response error message should contain "Token in task_token doesn't correspond to user session: got idUser=20, expected 10"
    And the table "users_items" should stay unchanged
    And the table "groups_attempts" should stay unchanged

  Scenario: itemUrls of task_token and hint_requested don't match
    Given I am the user with ID "10"
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
    And the following token "hintRequestToken" signed by the task platform is distributed:
      """
      {
        "idUser": "10",
        "idItemLocal": "50",
        "idAttempt": "100",
        "itemUrl": "http://taskplatform.mblockelet.info/task.html?taskId=555555555555555555",
        "askedHint": {"rotorIndex":1}
      }
      """
    When I send a POST request to "/items/ask-hint" with the following body:
      """
      {
        "task_token": "{{priorUserTaskToken}}",
        "hint_requested": "{{hintRequestToken}}"
      }
      """
    Then the response code should be 400
    And the response error message should contain "Wrong itemUrl in hint_requested token"
    And the table "users_items" should stay unchanged
    And the table "groups_attempts" should stay unchanged

  Scenario: idUser in hint_requested doesn't match the user's ID
    Given I am the user with ID "10"
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
    And the following token "hintRequestToken" signed by the task platform is distributed:
      """
      {
        "idUser": "20",
        "idItemLocal": "50",
        "idAttempt": "100",
        "itemURL": "http://taskplatform.mblockelet.info/task.html?taskId=403449543672183936",
        "askedHint": {"rotorIndex":1}
      }
      """
    When I send a POST request to "/items/ask-hint" with the following body:
      """
      {
        "task_token": "{{priorUserTaskToken}}",
        "hint_requested": "{{hintRequestToken}}"
      }
      """
    Then the response code should be 400
    And the response error message should contain "Token in hint_requested doesn't correspond to user session: got idUser=20, expected 10"
    And the table "users_items" should stay unchanged
    And the table "groups_attempts" should stay unchanged

  Scenario: idAttempt in hint_requested & task_token don't match
    Given I am the user with ID "10"
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
    And the following token "hintRequestToken" signed by the task platform is distributed:
      """
      {
        "idUser": "10",
        "idItemLocal": "50",
        "idAttempt": "101",
        "itemURL": "http://taskplatform.mblockelet.info/task.html?taskId=403449543672183936",
        "askedHint": {"rotorIndex":1}
      }
      """
    When I send a POST request to "/items/ask-hint" with the following body:
      """
      {
        "task_token": "{{priorUserTaskToken}}",
        "hint_requested": "{{hintRequestToken}}"
      }
      """
    Then the response code should be 400
    And the response error message should contain "Wrong idAttempt in hint_requested token"
    And the table "users_items" should stay unchanged
    And the table "groups_attempts" should stay unchanged

  Scenario: idItemLocal in hint_requested & task_token don't match
    Given I am the user with ID "10"
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
    And the following token "hintRequestToken" signed by the task platform is distributed:
      """
      {
        "idUser": "10",
        "idAttempt": "100",
        "idItemLocal": "51",
        "itemURL": "http://taskplatform.mblockelet.info/task.html?taskId=403449543672183936",
        "askedHint": {"rotorIndex":1}
      }
      """
    When I send a POST request to "/items/ask-hint" with the following body:
      """
      {
        "task_token": "{{priorUserTaskToken}}",
        "hint_requested": "{{hintRequestToken}}"
      }
      """
    Then the response code should be 400
    And the response error message should contain "Wrong idItemLocal in hint_requested token"
    And the table "users_items" should stay unchanged
    And the table "groups_attempts" should stay unchanged

  Scenario: No submission rights
    Given I am the user with ID "10"
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
    And the following token "hintRequestToken" signed by the task platform is distributed:
      """
      {
        "idUser": "10",
        "idItemLocal": "50",
        "idAttempt": "100",
        "itemURL": "http://taskplatform.mblockelet.info/task.html?taskId=403449543672183936",
        "askedHint": {"rotorIndex":1}
      }
      """
    When I send a POST request to "/items/ask-hint" with the following body:
      """
      {
        "task_token": "{{priorUserTaskToken}}",
        "hint_requested": "{{hintRequestToken}}"
      }
      """
    Then the response code should be 403
    And the response error message should contain "Item is read-only"
    And the table "users_items" should stay unchanged
    And the table "groups_attempts" should stay unchanged

  Scenario: idAttempt not found
    Given I am the user with ID "10"
    And the following token "priorUserTaskToken" signed by the app is distributed:
      """
      {
        "idUser": "10",
        "idItemLocal": "10",
        "idAttempt": "101",
        "itemURL": "http://taskplatform.mblockelet.info/task.html?taskId=403449543672183936",
        "platformName": "{{app().TokenConfig.PlatformName}}"
      }
      """
    And the following token "hintRequestToken" signed by the task platform is distributed:
      """
      {
        "idUser": "10",
        "idItemLocal": "10",
        "idAttempt": "101",
        "itemURL": "http://taskplatform.mblockelet.info/task.html?taskId=403449543672183936",
        "askedHint": {"rotorIndex":1}
      }
      """
    When I send a POST request to "/items/ask-hint" with the following body:
      """
      {
        "task_token": "{{priorUserTaskToken}}",
        "hint_requested": "{{hintRequestToken}}"
      }
      """
    Then the response code should be 404
    And the response error message should contain "Can't find previously requested hints info"
    And the table "users_items" should stay unchanged
    And the table "groups_attempts" should stay unchanged

  Scenario: missing askedHint
    Given I am the user with ID "10"
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
    And the following token "hintRequestToken" signed by the task platform is distributed:
      """
      {
        "idUser": "10",
        "idItemLocal": "50",
        "idAttempt": "100",
        "itemURL": "http://taskplatform.mblockelet.info/task.html?taskId=403449543672183936"
      }
      """
    When I send a POST request to "/items/ask-hint" with the following body:
      """
      {
        "task_token": "{{priorUserTaskToken}}",
        "hint_requested": "{{hintRequestToken}}"
      }
      """
    Then the response code should be 400
    And the response error message should contain "Asked hint should not be empty"
    And the table "users_items" should stay unchanged
    And the table "groups_attempts" should stay unchanged
