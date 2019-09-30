Feature: Ask for a hint
  Background:
    Given the database has the following table 'users':
      | id | login | self_group_id |
      | 10 | john  | 101           |
    And the database has the following table 'groups':
      | id  |
      | 101 |
    And the database has the following table 'groups_ancestors':
      | ancestor_group_id | child_group_id | is_self |
      | 101               | 101            | 1       |
    And the database has the following table 'groups_groups':
      | id | parent_group_id | child_group_id | type   | type_changed_at |
      | 15 | 22              | 13             | direct | null            |
    And the database has the following table 'platforms':
      | id | uses_tokens | regexp                                            | public_key                |
      | 10 | 1           | http://taskplatform.mblockelet.info/task.html\?.* | {{taskPlatformPublicKey}} |
    And the database has the following table 'items':
      | id | platform_id | url                                                                     |
      | 50 | 10          | http://taskplatform.mblockelet.info/task.html?taskId=403449543672183936 |
      | 10 | null        | null                                                                    |
    And the database has the following table 'items_items':
      | parent_item_id | child_item_id | child_order |
      | 10             | 50            | 0           |
    And the database has the following table 'items_ancestors':
      | ancestor_item_id | child_item_id |
      | 10               | 50            |
    And the database has the following table 'groups_items':
      | group_id | item_id | cached_partial_access_since | creator_user_id |
      | 101      | 50      | 2017-05-29 06:38:38         | 10              |
    And the database has the following table 'users_items':
      | user_id | item_id | hints_requested    | hints_cached | submissions_attempts | active_attempt_id |
      | 10      | 50      | [{"rotorIndex":0}] | 1            | 2                    | 100               |
      | 10      | 10      | null               | 0            | 0                    | null              |
    And time is frozen

  Scenario: User is able to ask for a hint
    Given I am the user with id "10"
    And the database has the following table 'groups_attempts':
      | id  | group_id | item_id | hints_requested        | hints_cached | order |
      | 100 | 101      | 50      | [0,  1, "hint" , null] | 4            | 0     |
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
    Then the response code should be 201
    And the response body decoded as "AskHintResponse" should be, in JSON:
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
            "platformName": "{{app().TokenConfig.PlatformName}}",
            "sHintsRequested": "[0,1,\"hint\",null,{\"rotorIndex\":1}]",
            "nbHintsGiven": "5"
          }
        },
        "message": "created",
        "success": true
      }
      """
    And the table "users_items" should be:
      | user_id | item_id | tasks_with_help | hints_cached | hints_requested    | ancestors_computation_state | ABS(TIMESTAMPDIFF(SECOND, latest_activity_at, NOW())) < 3 | ABS(TIMESTAMPDIFF(SECOND, latest_hint_at, NOW())) < 3 |
      | 10      | 10      | 1               | 0            | null               | done                        | 1                                                         | null                                                  |
      | 10      | 50      | 1               | 1            | [{"rotorIndex":0}] | done                        | 1                                                         | 1                                                     |
    And the table "groups_attempts" should be:
      | id  | group_id | item_id | tasks_with_help | hints_cached | hints_requested                    | ancestors_computation_state | ABS(TIMESTAMPDIFF(SECOND, latest_activity_at, NOW())) < 3 | ABS(TIMESTAMPDIFF(SECOND, latest_hint_at, NOW())) < 3 |
      | 100 | 101      | 50      | 1               | 5            | [0,1,"hint",null,{"rotorIndex":1}] | done                        | 1                                                         | 1                                                     |

  Scenario: User is able to ask for a hint with a minimal hint token
    Given I am the user with id "10"
    And the database has the following table 'groups_attempts':
      | id  | group_id | item_id | hints_requested        | order |
      | 100 | 101      | 50      | [0,  1, "hint" , null] | 0     |
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
    Then the response code should be 201
    And the response body decoded as "AskHintResponse" should be, in JSON:
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
            "platformName": "{{app().TokenConfig.PlatformName}}",
            "sHintsRequested": "[0,1,\"hint\",null,{\"rotorIndex\":1}]",
            "nbHintsGiven": "5"
          }
        },
        "message": "created",
        "success": true
      }
      """
    And the table "users_items" should be:
      | user_id | item_id | tasks_with_help | hints_cached | hints_requested    | ancestors_computation_state | ABS(TIMESTAMPDIFF(SECOND, latest_activity_at, NOW())) < 3 | ABS(TIMESTAMPDIFF(SECOND, latest_hint_at, NOW())) < 3 |
      | 10      | 10      | 1               | 0            | null               | done                        | 1                                                         | null                                                  |
      | 10      | 50      | 1               | 1            | [{"rotorIndex":0}] | done                        | 1                                                         | 1                                                     |
    And the table "groups_attempts" should be:
      | id  | group_id | item_id | tasks_with_help | hints_cached | hints_requested                    | ancestors_computation_state | ABS(TIMESTAMPDIFF(SECOND, latest_activity_at, NOW())) < 3 | ABS(TIMESTAMPDIFF(SECOND, latest_hint_at, NOW())) < 3 |
      | 100 | 101      | 50      | 1               | 5            | [0,1,"hint",null,{"rotorIndex":1}] | done                        | 1                                                         | 1                                                     |

  Scenario: User is able to ask for an already given hint
    Given I am the user with id "10"
    And the database has the following table 'groups_attempts':
      | id  | group_id | item_id | hints_requested        | order |
      | 100 | 101      | 50      | [0,  1, "hint" , null] | 0     |
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
        "itemUrl": "http://taskplatform.mblockelet.info/task.html?taskId=403449543672183936",
        "askedHint": "hint"
      }
      """
    When I send a POST request to "/items/ask-hint" with the following body:
      """
      {
        "task_token": "{{priorUserTaskToken}}",
        "hint_requested": "{{hintRequestToken}}"
      }
      """
    Then the response code should be 201
    And the response body decoded as "AskHintResponse" should be, in JSON:
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
            "platformName": "{{app().TokenConfig.PlatformName}}",
            "sHintsRequested": "[0,1,\"hint\",null]",
            "nbHintsGiven": "4"
          }
        },
        "message": "created",
        "success": true
      }
      """
    And the table "users_items" should be:
      | user_id | item_id | tasks_with_help | hints_cached | hints_requested    | ancestors_computation_state | ABS(TIMESTAMPDIFF(SECOND, latest_activity_at, NOW())) < 3 | ABS(TIMESTAMPDIFF(SECOND, latest_hint_at, NOW())) < 3 |
      | 10      | 10      | 1               | 0            | null               | done                        | 1                                                         | null                                                  |
      | 10      | 50      | 1               | 1            | [{"rotorIndex":0}] | done                        | 1                                                         | 1                                                     |
    And the table "groups_attempts" should be:
      | id  | group_id | item_id | tasks_with_help | hints_cached | hints_requested   | ancestors_computation_state | ABS(TIMESTAMPDIFF(SECOND, latest_activity_at, NOW())) < 3 | ABS(TIMESTAMPDIFF(SECOND, latest_hint_at, NOW())) < 3 |
      | 100 | 101      | 50      | 1               | 4            | [0,1,"hint",null] | done                        | 1                                                         | 1                                                     |

  Scenario: Can't parse hints_requested
    Given I am the user with id "10"
    And the database has the following table 'groups_attempts':
      | id  | group_id | item_id | hints_requested | order |
      | 100 | 101      | 50      | not an array    | 0     |
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
    Then the response code should be 201
    And the response body decoded as "AskHintResponse" should be, in JSON:
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
            "platformName": "{{app().TokenConfig.PlatformName}}",
            "sHintsRequested": "[{\"rotorIndex\":1}]",
            "nbHintsGiven": "1"
          }
        },
        "message": "created",
        "success": true
      }
      """
    And the table "users_items" should be:
      | user_id | item_id | tasks_with_help | hints_cached | hints_requested    | ancestors_computation_state | ABS(TIMESTAMPDIFF(SECOND, latest_activity_at, NOW())) < 3 | ABS(TIMESTAMPDIFF(SECOND, latest_hint_at, NOW())) < 3 |
      | 10      | 10      | 1               | 0            | null               | done                        | 1                                                         | null                                                  |
      | 10      | 50      | 1               | 1            | [{"rotorIndex":0}] | done                        | 1                                                         | 1                                                     |
    And the table "groups_attempts" should be:
      | id  | group_id | item_id | tasks_with_help | hints_cached | hints_requested    | ancestors_computation_state | ABS(TIMESTAMPDIFF(SECOND, latest_activity_at, NOW())) < 3 | ABS(TIMESTAMPDIFF(SECOND, latest_hint_at, NOW())) < 3 |
      | 100 | 101      | 50      | 1               | 1            | [{"rotorIndex":1}] | done                        | 1                                                         | 1                                                     |
    And logs should contain:
      """
      Unable to parse hints_requested ({"idAttempt":100,"idItemLocal":50,"idUser":10}) having value "not an array": invalid character 'o' in literal null (expecting 'u')
      """

