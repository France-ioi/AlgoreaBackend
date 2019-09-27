Feature: Submit a new answer
  Background:
    Given the database has the following table 'users':
      | id  | login | self_group_id |
      | 10  | john  | 101           |
    And the database has the following table 'groups':
      | id  |
      | 101 |
    And the database has the following table 'groups_ancestors':
      | ancestor_group_id | child_group_id | is_self |
      | 101               | 101            | 1       |
    And the database has the following table 'groups_groups':
      | id | parent_group_id | child_group_id | type   | status_changed_at |
      | 15 | 22              | 13             | direct | null              |
    And the database has the following table 'items':
      | id |
      | 50 |
    And the database has the following table 'groups_items':
      | group_id | item_id | cached_partial_access_since | creator_user_id |
      | 101      | 50      | 2017-05-29 06:38:38         | 10              |
    And the database has the following table 'users_items':
      | user_id | item_id | hints_requested                 | hints_cached | submissions_attempts |
      | 10      | 50      | [{"rotorIndex":0,"cellRank":0}] | 12           | 2                    |
    And the database has the following table 'groups_attempts':
      | id  | group_id | item_id | hints_requested                 | hints_cached | submissions_attempts | order |
      | 100 | 101      | 50      | [{"rotorIndex":0,"cellRank":0}] | 12           | 2                    | 0     |

  Scenario: User is able to submit a new answer
    Given I am the user with id "10"
    And time is frozen
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
        "answer": "print 1"
      }
      """
    Then the response code should be 201
    And the response body decoded as "AnswersSubmitResponse" should be, in JSON:
      """
      {
        "data": {
          "answer_token": {
            "date": "{{currentTimeInFormat("02-01-2006")}}",
            "idUser": "10",
            "idItem": null,
            "idAttempt": "100",
            "itemUrl": "",
            "idItemLocal": "50",
            "platformName": "algrorea_backend",
            "randomSeed": "",
            "sHintsRequested": "[{\"rotorIndex\":0,\"cellRank\":0}]",
            "nbHintsGiven": "12",
            "sAnswer": "print 1",
            "idUserAnswer": "8674665223082153551"
          }
        },
        "message": "created",
        "success": true
      }
      """
    And the table "users_items" should be:
      | user_id | item_id | submissions_attempts | ABS(TIMESTAMPDIFF(SECOND, last_activity_at, NOW())) < 3 |
      | 10      | 50      | 3                    | 1                                                       |
    And the table "users_answers" should be:
      | user_id | item_id | attempt_id | type       | answer  | ABS(TIMESTAMPDIFF(SECOND, submitted_at, NOW())) < 3 |
      | 10      | 50      | 100        | Submission | print 1 | 1                                                   |
    And the table "groups_attempts" should be:
      | id  | group_id | item_id | hints_requested                 | hints_cached | submissions_attempts | ABS(TIMESTAMPDIFF(SECOND, last_activity_at, NOW())) < 3 |
      | 100 | 101      | 50      | [{"rotorIndex":0,"cellRank":0}] | 12           | 3                    | 1                                                       |

  Scenario: User is able to submit a new answer (with all fields filled in the token)
    Given I am the user with id "10"
    And time is frozen
    And the following token "userTaskToken" signed by the app is distributed:
      """
      {
        "idItem": "50",
        "idUser": "10",
        "idItemLocal": "50",
        "idAttempt": "100",
        "itemUrl": "http://taskplatform.mblockelet.info/task.html?taskId=403449543672183936",
        "idItemLocal": "50",
        "randomSeed": "100",
        "platformName": "{{app().TokenConfig.PlatformName}}"
      }
      """
    When I send a POST request to "/answers" with the following body:
      """
      {
        "task_token": "{{userTaskToken}}",
        "answer": "print(2)"
      }
      """
    Then the response code should be 201
    And the response body decoded as "AnswersSubmitResponse" should be, in JSON:
      """
      {
        "data": {
          "answer_token": {
            "date": "{{currentTimeInFormat("02-01-2006")}}",
            "idUser": "10",
            "idItem": "50",
            "idAttempt": "100",
            "itemUrl": "http://taskplatform.mblockelet.info/task.html?taskId=403449543672183936",
            "idItemLocal": "50",
            "platformName": "algrorea_backend",
            "randomSeed": "100",
            "sHintsRequested": "[{\"rotorIndex\":0,\"cellRank\":0}]",
            "nbHintsGiven": "12",
            "sAnswer": "print(2)",
            "idUserAnswer": "8674665223082153551"
          }
        },
        "message": "created",
        "success": true
      }
      """
    And the table "users_items" should be:
      | user_id | item_id | submissions_attempts | ABS(TIMESTAMPDIFF(SECOND, last_activity_at, NOW())) < 3 |
      | 10      | 50      | 3                    | 1                                                       |
    And the table "users_answers" should be:
      | user_id | item_id | attempt_id | type       | answer   | ABS(TIMESTAMPDIFF(SECOND, submitted_at, NOW())) < 3 |
      | 10      | 50      | 100        | Submission | print(2) | 1                                                   |
    And the table "groups_attempts" should be:
      | id  | group_id | item_id | hints_requested                 | hints_cached | submissions_attempts | ABS(TIMESTAMPDIFF(SECOND, last_activity_at, NOW())) < 3 |
      | 100 | 101      | 50      | [{"rotorIndex":0,"cellRank":0}] | 12           | 3                    | 1                                                       |
