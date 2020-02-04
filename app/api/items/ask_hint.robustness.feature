Feature: Ask for a hint - robustness
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
      | id | regexp                                            | public_key                |
      | 10 | http://taskplatform.mblockelet.info/task.html\?.* | {{taskPlatformPublicKey}} |
    And the database has the following table 'items':
      | id | platform_id | url                                                                     | read_only | default_language_tag |
      | 50 | 10          | http://taskplatform.mblockelet.info/task.html?taskId=403449543672183936 | 1         | fr                   |
      | 10 | 10          | http://taskplatform.mblockelet.info/task.html?taskId=403449543672183936 | 0         | fr                   |
    And the database has the following table 'items_items':
      | parent_item_id | child_item_id | child_order |
      | 10             | 50            | 0           |
    And the database has the following table 'items_ancestors':
      | ancestor_item_id | child_item_id |
      | 10               | 50            |
    And the database has the following table 'permissions_generated':
      | group_id | item_id | can_view_generated |
      | 101      | 10      | content            |
      | 101      | 50      | content            |
    And the database has the following table 'attempts':
      | id  | group_id | item_id | hints_requested        | order |
      | 100 | 101      | 50      | [0,  1, "hint" , null] | 1     |
      | 200 | 101      | 10      | null                   | 1     |
    And time is frozen

  Scenario: Wrong JSON in request
    Given I am the user with id "101"
    When I send a POST request to "/items/ask-hint" with the following body:
      """
      []
      """
    Then the response code should be 400
    And the response error message should contain "Json: cannot unmarshal array into Go value of type items.askHintRequestWrapper"
    And the table "attempts" should stay unchanged

  Scenario: User not found
    Given I am the user with id "404"
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
    And the table "attempts" should stay unchanged

  Scenario: idUser in task_token doesn't match the user's id
    Given I am the user with id "101"
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
        "idUser": "101",
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
    And the table "attempts" should stay unchanged

  Scenario: itemUrls of task_token and hint_requested don't match
    Given I am the user with id "101"
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
    And the following token "hintRequestToken" signed by the task platform is distributed:
      """
      {
        "idUser": "101",
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
    And the table "attempts" should stay unchanged

  Scenario: idUser in hint_requested doesn't match the user's id
    Given I am the user with id "101"
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
    And the table "attempts" should stay unchanged

  Scenario: idAttempt in hint_requested & task_token don't match
    Given I am the user with id "101"
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
    And the following token "hintRequestToken" signed by the task platform is distributed:
      """
      {
        "idUser": "101",
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
    And the table "attempts" should stay unchanged

  Scenario: idItemLocal in hint_requested & task_token don't match
    Given I am the user with id "101"
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
    And the following token "hintRequestToken" signed by the task platform is distributed:
      """
      {
        "idUser": "101",
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
    And the table "attempts" should stay unchanged

  Scenario: No submission rights
    Given I am the user with id "101"
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
    And the following token "hintRequestToken" signed by the task platform is distributed:
      """
      {
        "idUser": "101",
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
    And the table "attempts" should stay unchanged

  Scenario: idAttempt not found
    Given I am the user with id "101"
    And the following token "priorUserTaskToken" signed by the app is distributed:
      """
      {
        "idUser": "101",
        "idItemLocal": "10",
        "idAttempt": "101",
        "itemURL": "http://taskplatform.mblockelet.info/task.html?taskId=403449543672183936",
        "platformName": "{{app().TokenConfig.PlatformName}}"
      }
      """
    And the following token "hintRequestToken" signed by the task platform is distributed:
      """
      {
        "idUser": "101",
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
    And the table "attempts" should stay unchanged

  Scenario: missing askedHint
    Given I am the user with id "101"
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
    And the following token "hintRequestToken" signed by the task platform is distributed:
      """
      {
        "idUser": "101",
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
    And the table "attempts" should stay unchanged
