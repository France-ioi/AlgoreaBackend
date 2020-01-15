Feature: Get a task token with a refreshed attempt for an item
  Background:
    Given the database has the following table 'groups':
      | id  | team_item_id | type     |
      | 101 | null         | UserSelf |
      | 102 | 10           | Team     |
      | 111 | null         | UserSelf |
    And the database has the following table 'users':
      | login | group_id |
      | john  | 101      |
      | jane  | 111      |
    And the database has the following table 'groups_groups':
      | parent_group_id | child_group_id |
      | 102             | 101            |
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
    And the database has the following table 'permissions_generated':
      | group_id | item_id | can_view_generated       |
      | 101      | 50      | content                  |
      | 101      | 60      | solution                 |
      | 102      | 60      | solution                 |
      | 111      | 50      | content_with_descendants |
    And time is frozen

  Scenario: User is able to fetch a task token
    Given I am the user with id "101"
    And the database has the following table 'attempts':
      | id | group_id | item_id | order | latest_activity_at  | started_at | score_computed | score_obtained_at | validated_at | hints_requested | hints_cached |
      | 1  | 101      | 50      | 0     | 2017-05-29 06:38:38 | null       | 0              | null              | null         | null            | 0            |
      | 2  | 101      | 50      | 1     | 2018-05-29 06:38:38 | null       | 0              | null              | null         | [1,2,3,4]       | 4            |
      | 3  | 102      | 50      | 0     | 2019-05-29 06:38:38 | null       | 0              | null              | null         | null            | 0            |
      | 4  | 101      | 51      | 0     | 2019-04-29 06:38:38 | null       | 0              | null              | null         | null            | 0            |
    When I send a GET request to "/attempts/2/task-token"
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
          "idUser": "101",
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
    And the table "attempts" should stay unchanged but the row with id "2"
    And the table "attempts" at id "2" should be:
      | id | group_id | item_id | score_computed | tasks_tried | result_propagation_state | ABS(TIMESTAMPDIFF(SECOND, latest_activity_at, NOW())) < 3 | ABS(TIMESTAMPDIFF(SECOND, latest_answer_at, NOW())) < 3 | ABS(TIMESTAMPDIFF(SECOND, score_obtained_at, NOW())) < 3 | ABS(TIMESTAMPDIFF(SECOND, validated_at, NOW())) < 3 | ABS(TIMESTAMPDIFF(SECOND, started_at, NOW())) < 3 |
      | 2  | 101      | 50      | 0              | 0           | done                     | 1                                                         | null                                                    | null                                                     | null                                                | 1                                                 |

  Scenario: User is able to fetch a task token as a team
    Given I am the user with id "101"
    And the database has the following table 'attempts':
      | id | group_id | item_id | order | latest_activity_at  | started_at | score_computed | score_obtained_at | validated_at | hints_requested | hints_cached |
      | 1  | 102      | 60      | 0     | 2017-05-29 06:38:38 | null       | 0              | null              | null         | null            | 0            |
      | 2  | 102      | 60      | 1     | 2018-05-29 06:38:38 | null       | 0              | null              | null         | [1,2,3,4]       | 4            |
      | 3  | 101      | 60      | 0     | 2019-05-29 06:38:38 | null       | 0              | null              | null         | null            | 0            |
      | 4  | 102      | 61      | 0     | 2019-04-29 06:38:38 | null       | 0              | null              | null         | null            | 0            |
    When I send a GET request to "/attempts/2/task-token"
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
          "idUser": "101",
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
    And the table "attempts" should stay unchanged but the row with id "2"
    And the table "attempts" at id "2" should be:
      | id | group_id | item_id | score_computed | tasks_tried | result_propagation_state | ABS(TIMESTAMPDIFF(SECOND, latest_activity_at, NOW())) < 3 | ABS(TIMESTAMPDIFF(SECOND, latest_answer_at, NOW())) < 3 | ABS(TIMESTAMPDIFF(SECOND, score_obtained_at, NOW())) < 3 | ABS(TIMESTAMPDIFF(SECOND, validated_at, NOW())) < 3 | ABS(TIMESTAMPDIFF(SECOND, started_at, NOW())) < 3 |
      | 2  | 102      | 60      | 0              | 0           | done                     | 1                                                         | null                                                    | null                                                     | null                                                | 1                                                 |

  Scenario: Keeps previous started_at values
    Given I am the user with id "101"
    And the database has the following table 'attempts':
      | id | group_id | item_id | order | latest_activity_at  | started_at          | score_computed | score_obtained_at | validated_at |
      | 2  | 101      | 50      | 0     | 2018-05-29 06:38:38 | 2017-05-29 06:38:38 | 0              | null              | null         |
    When I send a GET request to "/attempts/2/task-token"
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
          "idUser": "101",
          "idItemLocal": "50",
          "idItem": "task1",
          "itemUrl": "http://taskplatform.mblockelet.info/task.html?taskId=403449543672183936",
          "nbHintsGiven": "0",
          "randomSeed": "2",
          "platformName": "{{app().TokenConfig.PlatformName}}"
        }
      }
      """
    And the table "attempts" should stay unchanged but the row with id "2"
    And the table "attempts" at id "2" should be:
      | id | group_id | item_id | score_computed | tasks_tried | result_propagation_state | ABS(TIMESTAMPDIFF(SECOND, latest_activity_at, NOW())) < 3 | ABS(TIMESTAMPDIFF(SECOND, latest_answer_at, NOW())) < 3 | ABS(TIMESTAMPDIFF(SECOND, score_obtained_at, NOW())) < 3 | ABS(TIMESTAMPDIFF(SECOND, validated_at, NOW())) < 3 | started_at          |
      | 2  | 101      | 50      | 0              | 0           | done                     | 1                                                         | null                                                    | null                                                     | null                                                | 2017-05-29 06:38:38 |
