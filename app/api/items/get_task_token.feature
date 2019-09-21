Feature: Get a task token with a refreshed active attempt for an item
  Background:
    Given the database has the following table 'users':
      | id | login | self_group_id |
      | 10 | john  | 101           |
      | 11 | jane  | 111           |
    And the database has the following table 'groups':
      | id  | team_item_id | type     |
      | 101 | null         | UserSelf |
      | 102 | 10           | Team     |
      | 111 | null         | UserSelf |
    And the database has the following table 'groups_groups':
      | parent_group_id | child_group_id | type               |
      | 102             | 101            | invitationAccepted |
    And the database has the following table 'groups_ancestors':
      | ancestor_group_id | child_group_id | is_self |
      | 101               | 101            | 1       |
      | 102               | 101            | 0       |
      | 102               | 102            | 1       |
      | 111               | 111            | 1       |
    And the database has the following table 'items':
      | id | url                                                                     | type    | has_attempts | hints_allowed | text_id | supported_lang_prog |
      | 10 | null                                                                    | Chapter | 0            | 0             | null    | null                |
      | 50 | http://taskplatform.mblockelet.info/task.html?taskId=403449543672183936 | Task    | 0            | 1             | task1   | null                |
      | 60 | http://taskplatform.mblockelet.info/task.html?taskId=403449543672183936 | Course  | 1            | 0             | null    | c,python            |
    And the database has the following table 'items_ancestors':
      | ancestor_item_id | child_item_id |
      | 10               | 60            |
    And the database has the following table 'groups_items':
      | group_id | item_id | cached_partial_access_date | cached_full_access_date | cached_access_solutions_date | creator_user_id |
      | 101      | 50      | 2017-05-29 06:38:38        | null                    | null                         | 10              |
      | 101      | 60      | 2017-05-29 06:38:38        | null                    | 2017-05-29 06:38:38          | 10              |
      | 111      | 50      | null                       | 2017-05-29 06:38:38     | null                         | 10              |
    And time is frozen

  Scenario: User is able to fetch an active attempt (no active attempt set)
    Given I am the user with id "11"
    And the database has the following table 'users_items':
      | user_id | item_id | active_attempt_id | score | best_answer_date | validation_date | start_date | hints_requested | hints_cached |
      | 11      | 50      | null              | 0     | null             | null            | null       | 1,2,3           | 3            |
    When I send a GET request to "/items/50/task-token"
    Then the response code should be 200
    And the response body decoded as "GetTaskTokenResponse" should be, in JSON:
      """
      {
        "task_token": {
          "date": "{{currentTimeInFormat("02-01-2006")}}",
          "bAccessSolutions": false,
          "bHintsAllowed": true,
          "bIsAdmin": false,
          "bReadAnswers": true,
          "bSubmissionPossible": true,
          "idAttempt": "8674665223082153551",
          "idUser": "11",
          "idItemLocal": "50",
          "idItem": "task1",
          "itemUrl": "http://taskplatform.mblockelet.info/task.html?taskId=403449543672183936",
          "nbHintsGiven": "0",
          "randomSeed": "8674665223082153551",
          "platformName": "{{app().TokenConfig.PlatformName}}"
        }
      }
      """
    And the table "users_items" should be:
      | user_id | item_id | score | tasks_tried | validated | key_obtained | ancestors_computation_state | ABS(TIMESTAMPDIFF(SECOND, last_activity_date, NOW())) < 3 | ABS(TIMESTAMPDIFF(SECOND, last_answer_date, NOW())) < 3 | ABS(TIMESTAMPDIFF(SECOND, best_answer_date, NOW())) < 3 | ABS(TIMESTAMPDIFF(SECOND, validation_date, NOW())) < 3 | ABS(TIMESTAMPDIFF(SECOND, start_date, NOW())) < 3 |
      | 11      | 50      | 0     | 0           | 0         | 0            | done                        | 1                                                         | null                                                    | null                                                    | null                                                   | 1                                                 |
    And the table "groups_attempts" should be:
      | id                  | group_id | item_id | score | tasks_tried | validated | key_obtained | ancestors_computation_state | ABS(TIMESTAMPDIFF(SECOND, last_activity_date, NOW())) < 3 | ABS(TIMESTAMPDIFF(SECOND, last_answer_date, NOW())) < 3 | ABS(TIMESTAMPDIFF(SECOND, best_answer_date, NOW())) < 3 | ABS(TIMESTAMPDIFF(SECOND, validation_date, NOW())) < 3 | ABS(TIMESTAMPDIFF(SECOND, start_date, NOW())) < 3 |
      | 8674665223082153551 | 111      | 50      | 0     | 0           | 0         | 0            | done                        | 1                                                         | null                                                    | null                                                    | null                                                   | 1                                                 |

  Scenario: User is able to fetch a task token (no active attempt set, only full access)
    Given I am the user with id "10"
    And the database has the following table 'users_items':
      | user_id | item_id | active_attempt_id | score | best_answer_date | validation_date | start_date |
      | 10      | 50      | null              | 0     | null             | null            | null       |
    When I send a GET request to "/items/50/task-token"
    Then the response code should be 200
    And the response body decoded as "GetTaskTokenResponse" should be, in JSON:
      """
      {
        "task_token": {
          "date": "{{currentTimeInFormat("02-01-2006")}}",
          "bAccessSolutions": false,
          "bHintsAllowed": true,
          "bIsAdmin": false,
          "bReadAnswers": true,
          "bSubmissionPossible": true,
          "idAttempt": "8674665223082153551",
          "idUser": "10",
          "idItem": "task1",
          "idItemLocal": "50",
          "itemUrl": "http://taskplatform.mblockelet.info/task.html?taskId=403449543672183936",
          "nbHintsGiven": "0",
          "randomSeed": "8674665223082153551",
          "platformName": "{{app().TokenConfig.PlatformName}}"
        }
      }
      """
    And the table "users_items" should be:
      | user_id | item_id | score | tasks_tried | validated | key_obtained | ancestors_computation_state | ABS(TIMESTAMPDIFF(SECOND, last_activity_date, NOW())) < 3 | ABS(TIMESTAMPDIFF(SECOND, last_answer_date, NOW())) < 3 | ABS(TIMESTAMPDIFF(SECOND, best_answer_date, NOW())) < 3 | ABS(TIMESTAMPDIFF(SECOND, validation_date, NOW())) < 3 | ABS(TIMESTAMPDIFF(SECOND, start_date, NOW())) < 3 |
      | 10      | 50      | 0     | 0           | 0         | 0            | done                        | 1                                                         | null                                                    | null                                                    | null                                                   | 1                                                 |
    And the table "groups_attempts" should be:
      | id                  | group_id | item_id | score | tasks_tried | validated | key_obtained | ancestors_computation_state | ABS(TIMESTAMPDIFF(SECOND, last_activity_date, NOW())) < 3 | ABS(TIMESTAMPDIFF(SECOND, last_answer_date, NOW())) < 3 | ABS(TIMESTAMPDIFF(SECOND, best_answer_date, NOW())) < 3 | ABS(TIMESTAMPDIFF(SECOND, validation_date, NOW())) < 3 | ABS(TIMESTAMPDIFF(SECOND, start_date, NOW())) < 3 |
      | 8674665223082153551 | 101      | 50      | 0     | 0           | 0         | 0            | done                        | 1                                                         | null                                                    | null                                                    | null                                                   | 1                                                 |

  Scenario: User is able to fetch a task token (no active attempt and item.has_attempts=1)
    Given I am the user with id "10"
    And the database has the following table 'users_items':
      | user_id | item_id | active_attempt_id | score | best_answer_date | validation_date | start_date |
      | 10      | 60      | null              | 0     | null             | null            | null       |
    When I send a GET request to "/items/60/task-token"
    Then the response code should be 200
    And the response body decoded as "GetTaskTokenResponse" should be, in JSON:
      """
      {
        "task_token": {
          "date": "{{currentTimeInFormat("02-01-2006")}}",
          "bAccessSolutions": true,
          "bHintsAllowed": false,
          "bIsAdmin": false,
          "bReadAnswers": true,
          "bSubmissionPossible": true,
          "idAttempt": "8674665223082153551",
          "idUser": "10",
          "idItemLocal": "60",
          "itemUrl": "http://taskplatform.mblockelet.info/task.html?taskId=403449543672183936",
          "nbHintsGiven": "0",
          "sSupportedLangProg": "c,python",
          "randomSeed": "8674665223082153551",
          "platformName": "{{app().TokenConfig.PlatformName}}"
        }
      }
      """
    And the table "users_items" should be:
      | user_id | item_id | score | tasks_tried | validated | key_obtained | ancestors_computation_state | ABS(TIMESTAMPDIFF(SECOND, last_activity_date, NOW())) < 3 | ABS(TIMESTAMPDIFF(SECOND, last_answer_date, NOW())) < 3 | ABS(TIMESTAMPDIFF(SECOND, best_answer_date, NOW())) < 3 | ABS(TIMESTAMPDIFF(SECOND, validation_date, NOW())) < 3 | ABS(TIMESTAMPDIFF(SECOND, start_date, NOW())) < 3 |
      | 10      | 60      | 0     | 0           | 0         | 0            | done                        | 1                                                         | null                                                    | null                                                    | null                                                   | 1                                                 |
    And the table "groups_attempts" should be:
      | id                  | group_id | item_id | score | tasks_tried | validated | key_obtained | ancestors_computation_state | ABS(TIMESTAMPDIFF(SECOND, last_activity_date, NOW())) < 3 | ABS(TIMESTAMPDIFF(SECOND, last_answer_date, NOW())) < 3 | ABS(TIMESTAMPDIFF(SECOND, best_answer_date, NOW())) < 3 | ABS(TIMESTAMPDIFF(SECOND, validation_date, NOW())) < 3 | ABS(TIMESTAMPDIFF(SECOND, start_date, NOW())) < 3 |
      | 8674665223082153551 | 102      | 60      | 0     | 0           | 0         | 0            | done                        | 1                                                         | null                                                    | null                                                    | null                                                   | 1                                                 |

  Scenario: User is able to fetch a task token (with active attempt set)
    Given I am the user with id "10"
    And the database has the following table 'users_items':
      | user_id | item_id | active_attempt_id | score | best_answer_date | validation_date | start_date |
      | 10      | 50      | 100               | 0     | null             | null            | null       |
    When I send a GET request to "/items/50/task-token"
    Then the response code should be 200
    And the response body decoded as "GetTaskTokenResponse" should be, in JSON:
      """
      {
        "task_token": {
          "date": "{{currentTimeInFormat("02-01-2006")}}",
          "bAccessSolutions": false,
          "bHintsAllowed": true,
          "bIsAdmin": false,
          "bReadAnswers": true,
          "bSubmissionPossible": true,
          "idAttempt": "100",
          "idUser": "10",
          "idItem": "task1",
          "idItemLocal": "50",
          "itemUrl": "http://taskplatform.mblockelet.info/task.html?taskId=403449543672183936",
          "nbHintsGiven": "0",
          "randomSeed": "100",
          "platformName": "{{app().TokenConfig.PlatformName}}"
        }
      }
      """
    And the table "users_items" should be:
      | user_id | item_id | score | tasks_tried | validated | key_obtained | ancestors_computation_state | ABS(TIMESTAMPDIFF(SECOND, last_activity_date, NOW())) < 3 | ABS(TIMESTAMPDIFF(SECOND, last_answer_date, NOW())) < 3 | ABS(TIMESTAMPDIFF(SECOND, best_answer_date, NOW())) < 3 | ABS(TIMESTAMPDIFF(SECOND, validation_date, NOW())) < 3 | ABS(TIMESTAMPDIFF(SECOND, start_date, NOW())) < 3 |
      | 10      | 50      | 0     | 0           | 0         | 0            | done                        | 1                                                         | null                                                    | null                                                    | null                                                   | 1                                                 |
    And the table "groups_attempts" should stay unchanged

  Scenario: User is able to fetch a task token (no active attempt set, but there are some in the DB)
    Given I am the user with id "10"
    And the database has the following table 'users_items':
      | user_id | item_id | active_attempt_id | score | best_answer_date | validation_date | start_date |
      | 10      | 50      | null              | 0     | null             | null            | null       |
    And the database has the following table 'groups_attempts':
      | id | group_id | item_id | order | last_activity_date  | start_date | score | best_answer_date | validation_date | hints_requested | hints_cached |
      | 1  | 101      | 50      | 0     | 2017-05-29 06:38:38 | null       | 0     | null             | null            | null            | 0            |
      | 2  | 101      | 50      | 1     | 2018-05-29 06:38:38 | null       | 0     | null             | null            | [1,2,3,4]       | 4            |
      | 3  | 102      | 50      | 0     | 2019-05-29 06:38:38 | null       | 0     | null             | null            | null            | 0            |
      | 4  | 101      | 51      | 0     | 2019-04-29 06:38:38 | null       | 0     | null             | null            | null            | 0            |
    When I send a GET request to "/items/50/task-token"
    Then the response code should be 200
    And the response body decoded as "GetTaskTokenResponse" should be, in JSON:
      """
      {
        "task_token": {
          "date": "{{currentTimeInFormat("02-01-2006")}}",
          "bAccessSolutions": false,
          "bHintsAllowed": true,
          "bIsAdmin": false,
          "bReadAnswers": true,
          "bSubmissionPossible": true,
          "idAttempt": "2",
          "idUser": "10",
          "idItemLocal": "50",
          "idItem": "task1",
          "itemUrl": "http://taskplatform.mblockelet.info/task.html?taskId=403449543672183936",
          "nbHintsGiven": "4",
          "sHintsRequested": "[1,2,3,4]",
          "randomSeed": "2",
          "platformName": "{{app().TokenConfig.PlatformName}}"
        }
      }
      """
    And the table "users_items" should be:
      | user_id | item_id | score | tasks_tried | validated | key_obtained | ancestors_computation_state | ABS(TIMESTAMPDIFF(SECOND, last_activity_date, NOW())) < 3 | ABS(TIMESTAMPDIFF(SECOND, last_answer_date, NOW())) < 3 | ABS(TIMESTAMPDIFF(SECOND, best_answer_date, NOW())) < 3 | ABS(TIMESTAMPDIFF(SECOND, validation_date, NOW())) < 3 | ABS(TIMESTAMPDIFF(SECOND, start_date, NOW())) < 3 |
      | 10      | 50      | 0     | 0           | 0         | 0            | done                        | 1                                                         | null                                                    | null                                                    | null                                                   | 1                                                 |
    And the table "groups_attempts" should stay unchanged but the row with id "2"
    And the table "groups_attempts" at id "2" should be:
      | id | group_id | item_id | score | tasks_tried | validated | key_obtained | ancestors_computation_state | ABS(TIMESTAMPDIFF(SECOND, last_activity_date, NOW())) < 3 | ABS(TIMESTAMPDIFF(SECOND, last_answer_date, NOW())) < 3 | ABS(TIMESTAMPDIFF(SECOND, best_answer_date, NOW())) < 3 | ABS(TIMESTAMPDIFF(SECOND, validation_date, NOW())) < 3 | ABS(TIMESTAMPDIFF(SECOND, start_date, NOW())) < 3 |
      | 2  | 101      | 50      | 0     | 0           | 0         | 0            | done                        | 1                                                         | null                                                    | null                                                    | null                                                   | 1                                                 |

  Scenario: User is able to fetch a task token (no active attempt set, but there are some in the DB and items.has_attempts=1)
    Given I am the user with id "10"
    And the database has the following table 'users_items':
      | user_id | item_id | active_attempt_id | score | best_answer_date | validation_date | start_date |
      | 10      | 60      | null              | 0     | null             | null            | null       |
    And the database has the following table 'groups_attempts':
      | id | group_id | item_id | order | last_activity_date  | start_date | score | best_answer_date | validation_date | hints_requested | hints_cached |
      | 1  | 102      | 60      | 0     | 2017-05-29 06:38:38 | null       | 0     | null             | null            | null            | 0            |
      | 2  | 102      | 60      | 1     | 2018-05-29 06:38:38 | null       | 0     | null             | null            | [1,2,3,4]       | 4            |
      | 3  | 101      | 60      | 0     | 2019-05-29 06:38:38 | null       | 0     | null             | null            | null            | 0            |
      | 4  | 102      | 61      | 0     | 2019-04-29 06:38:38 | null       | 0     | null             | null            | null            | 0            |
    When I send a GET request to "/items/60/task-token"
    Then the response code should be 200
    And the response body decoded as "GetTaskTokenResponse" should be, in JSON:
      """
      {
        "task_token": {
          "date": "{{currentTimeInFormat("02-01-2006")}}",
          "bAccessSolutions": true,
          "bHintsAllowed": false,
          "bIsAdmin": false,
          "bReadAnswers": true,
          "bSubmissionPossible": true,
          "idAttempt": "2",
          "idUser": "10",
          "idItemLocal": "60",
          "itemUrl": "http://taskplatform.mblockelet.info/task.html?taskId=403449543672183936",
          "nbHintsGiven": "4",
          "sHintsRequested": "[1,2,3,4]",
          "sSupportedLangProg": "c,python",
          "randomSeed": "2",
          "platformName": "{{app().TokenConfig.PlatformName}}"
        }
      }
      """
    And the table "users_items" should be:
      | user_id | item_id | score | tasks_tried | validated | key_obtained | ancestors_computation_state | ABS(TIMESTAMPDIFF(SECOND, last_activity_date, NOW())) < 3 | ABS(TIMESTAMPDIFF(SECOND, last_answer_date, NOW())) < 3 | ABS(TIMESTAMPDIFF(SECOND, best_answer_date, NOW())) < 3 | ABS(TIMESTAMPDIFF(SECOND, validation_date, NOW())) < 3 | ABS(TIMESTAMPDIFF(SECOND, start_date, NOW())) < 3 |
      | 10      | 60      | 0     | 0           | 0         | 0            | done                        | 1                                                         | null                                                    | null                                                    | null                                                   | 1                                                 |
    And the table "groups_attempts" should stay unchanged but the row with id "2"
    And the table "groups_attempts" at id "2" should be:
      | id | group_id | item_id | score | tasks_tried | validated | key_obtained | ancestors_computation_state | ABS(TIMESTAMPDIFF(SECOND, last_activity_date, NOW())) < 3 | ABS(TIMESTAMPDIFF(SECOND, last_answer_date, NOW())) < 3 | ABS(TIMESTAMPDIFF(SECOND, best_answer_date, NOW())) < 3 | ABS(TIMESTAMPDIFF(SECOND, validation_date, NOW())) < 3 | ABS(TIMESTAMPDIFF(SECOND, start_date, NOW())) < 3 |
      | 2  | 102      | 60      | 0     | 0           | 0         | 0            | done                        | 1                                                         | null                                                    | null                                                    | null                                                   | 1                                                 |

  Scenario: Keeps previous start_date values
    Given I am the user with id "10"
    And the database has the following table 'users_items':
      | user_id | item_id | active_attempt_id | score | best_answer_date | validation_date | start_date          |
      | 10      | 50      | null              | 0     | null             | null            | 2017-05-29 06:38:38 |
    And the database has the following table 'groups_attempts':
      | id | group_id | item_id | order | last_activity_date  | start_date          | score | best_answer_date | validation_date |
      | 2  | 101      | 50      | 0     | 2018-05-29 06:38:38 | 2017-05-29 06:38:38 | 0     | null             | null            |
    When I send a GET request to "/items/50/task-token"
    Then the response code should be 200
    And the response body decoded as "GetTaskTokenResponse" should be, in JSON:
      """
      {
        "task_token": {
          "date": "{{currentTimeInFormat("02-01-2006")}}",
          "bAccessSolutions": false,
          "bHintsAllowed": true,
          "bIsAdmin": false,
          "bReadAnswers": true,
          "bSubmissionPossible": true,
          "idAttempt": "2",
          "idUser": "10",
          "idItemLocal": "50",
          "idItem": "task1",
          "itemUrl": "http://taskplatform.mblockelet.info/task.html?taskId=403449543672183936",
          "nbHintsGiven": "0",
          "randomSeed": "2",
          "platformName": "{{app().TokenConfig.PlatformName}}"
        }
      }
      """
    And the table "users_items" should be:
      | user_id | item_id | score | tasks_tried | validated | key_obtained | ancestors_computation_state | ABS(TIMESTAMPDIFF(SECOND, last_activity_date, NOW())) < 3 | ABS(TIMESTAMPDIFF(SECOND, last_answer_date, NOW())) < 3 | ABS(TIMESTAMPDIFF(SECOND, best_answer_date, NOW())) < 3 | ABS(TIMESTAMPDIFF(SECOND, validation_date, NOW())) < 3 | start_date          |
      | 10      | 50      | 0     | 0           | 0         | 0            | done                        | 1                                                         | null                                                    | null                                                    | null                                                   | 2017-05-29 06:38:38 |
    And the table "groups_attempts" should stay unchanged but the row with id "2"
    And the table "groups_attempts" at id "2" should be:
      | id | group_id | item_id | score | tasks_tried | validated | key_obtained | ancestors_computation_state | ABS(TIMESTAMPDIFF(SECOND, last_activity_date, NOW())) < 3 | ABS(TIMESTAMPDIFF(SECOND, last_answer_date, NOW())) < 3 | ABS(TIMESTAMPDIFF(SECOND, best_answer_date, NOW())) < 3 | ABS(TIMESTAMPDIFF(SECOND, validation_date, NOW())) < 3 | start_date          |
      | 2  | 101      | 50      | 0     | 0           | 0         | 0            | done                        | 1                                                         | null                                                    | null                                                    | null                                                   | 2017-05-29 06:38:38 |

