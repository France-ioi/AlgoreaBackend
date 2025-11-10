Feature: Get activity log for a thread - can_watch_answer field
  Background:
    Given the database has the following users:
      | login | group_id | default_language | profile                                                |
      | owner | 21       | fr               | {"first_name": "Jean-Michel", "last_name": "Blanquer"} |
      | user  | 11       | en               | {"first_name": "John", "last_name": "Doe"}             |
      | jane  | 31       | en               | {"first_name": "Jane", "last_name": "Doe"}             |
      | paul  | 41       | en               | {"first_name": "Paul", "last_name": "Smith"}           |
    And the database has the following table "groups":
      | id | type  | name         |
      | 13 | Class | Our Class    |
      | 20 | Other | Some Group   |
      | 30 | Team  | Our Team     |
      | 40 | Club  | Our Club     |
      | 50 | Club  | Team Club    |
      | 60 | Team  | Another Team |
    And the database has the following table "groups_groups":
      | parent_group_id | child_group_id | personal_info_view_approved_at |
      | 13              | 11             | 2019-05-30 11:00:00            |
      | 13              | 41             | 2019-05-30 11:00:00            |
      | 20              | 21             | null                           |
      | 30              | 21             | null                           |
      | 40              | 21             | null                           |
      | 50              | 30             | null                           |
      | 60              | 31             | null                           |
    And the groups ancestors are computed
    And the database has the following table "attempts":
      | id | participant_id |
      | 0  | 11             |
      | 1  | 11             |
    And the database has the following table "results":
      | attempt_id | item_id | participant_id | started_at          | validated_at        | latest_submission_at |
      | 0          | 200     | 30             | 2017-05-29 06:38:00 | 2017-05-30 12:00:00 | 2020-05-29 06:38:38  |
    And the database has the following table "answers":
      | id | author_id | participant_id | attempt_id | item_id | type       | state   | created_at          |
      | 11 | 11        | 11             | 0          | 200     | Submission | State11 | 2017-05-29 06:38:38 |
    And the database has the following table "items":
      | id  | type    | no_score | default_language_tag |
      | 200 | Task    | false    | fr                   |
    And the database has the following table "items_strings":
      | item_id | language_tag | title      | image_url                  | subtitle     | description   | edu_comment    |
      | 200     | en           | Task 1     | http://example.com/my0.jpg | Subtitle 0   | Description 0 | Some comment   |
      | 200     | fr           | Tache 1    | http://example.com/mf0.jpg | Sous-titre 0 | texte 0       | Un commentaire |
    And the database has the following table "languages":
      | tag |
      | fr  |

  Scenario Outline: An ancestor of the user can watch submissions from the participant group
    Given I am the user with id "31"
    And I can view content of the item 200
    And I have the watch permission set to "<can_watch>" on the item 200
    And I am a member of the group @ParentGroup
    And @ParentGroup is a member of the group @GrandParentGroup
    And the group @GrandParentGroup is a manager of the group <managed_group> and can watch for submissions from the group and its descendants
    And I am a member of the group @Helper
    And I have a validated result on the item 200
    And there is a thread with "item_id=200,participant_id=11,helper_group_id=@Helper,status=closed,latest_update_at={{relativeTimeDBMs("-1h")}}"
    When I send a GET request to "/items/200/participant/11/thread/log"
    Then the response code should be 200
    And the response body should be, in JSON:
    """
    [
      {
        "activity_type": "submission",
        "answer_id": "11",
        "at": "2017-05-29T06:38:38Z",
        "attempt_id": "0",
        "can_watch_answer": <can_watch_answer>,
        "from_answer_id": "11",
        "item": {
          "id": "200",
          "string": {
            "title": "Task 1"
          },
          "type": "Task"
        },
        "participant": {
          "id": "11",
          "name": "user",
          "type": "User"
        },
        "user": {
          "id": "11",
          "login": "user"
        }
      }
    ]
    """
    Examples:
      | can_watch         | managed_group | can_watch_answer |
      | result            | 11            | false            |
      | answer            | 11            | true             |
      | answer_with_grant | 11            | true             |
      | result            | 31            | false            |
      | answer            | 31            | false            |
      | answer_with_grant | 31            | false            |

  Scenario: The user is the participant themselves
    Given I am the user with id "11"
    And I can view content of the item 200
    And there is a thread with "item_id=200,participant_id=11,helper_group_id=@Helper,status=closed,latest_update_at={{relativeTimeDBMs("-1h")}}"
    When I send a GET request to "/items/200/participant/11/thread/log"
    Then the response code should be 200
    And the response body should be, in JSON:
    """
    [
      {
        "activity_type": "submission",
        "answer_id": "11",
        "at": "2017-05-29T06:38:38Z",
        "attempt_id": "0",
        "can_watch_answer": true,
        "from_answer_id": "11",
        "item": {
          "id": "200",
          "string": {
            "title": "Task 1"
          },
          "type": "Task"
        },
        "participant": {
          "id": "11",
          "name": "user",
          "type": "User"
        },
        "user": {
          "id": "11",
          "login": "user",
          "first_name": "John",
          "last_name": "Doe"
        }
      }
    ]
    """

  Scenario: The participant is a team and the user is a team member of that team
    Given I am the user with id "21"
    And I can view content of the item 200
    And there is a thread with "item_id=200,participant_id=30,helper_group_id=@Helper,status=closed,latest_update_at={{relativeTimeDBMs("-1h")}}"
    When I send a GET request to "/items/200/participant/30/thread/log"
    Then the response code should be 200
    And the response body should be, in JSON:
    """
    [
      {
        "activity_type": "result_validated",
        "at": "2017-05-30T12:00:00Z",
        "attempt_id": "0",
        "can_watch_answer": true,
        "from_answer_id": "-1",
        "item": {
          "id": "200",
          "string": {
            "title": "Tache 1"
          },
          "type": "Task"
        },
        "participant": {
          "id": "30",
          "name": "Our Team",
          "type": "Team"
        }
      },
      {
        "activity_type": "result_started",
        "at": "2017-05-29T06:38:00Z",
        "attempt_id": "0",
        "can_watch_answer": true,
        "from_answer_id": "-1",
        "item": {
          "id": "200",
          "string": {
            "title": "Tache 1"
          },
          "type": "Task"
        },
        "participant": {
          "id": "30",
          "name": "Our Team",
          "type": "Team"
        }
      }
    ]
    """

  Scenario: The participant is a team and the user is a team member of another team
    Given I am the user with id "31"
    And I can view content of the item 200
    And I have the watch permission set to "answer" on the item 200
    And there is a thread with "item_id=200,participant_id=30,helper_group_id=@Helper,status=closed,latest_update_at={{relativeTimeDBMs("-1h")}}"
    When I send a GET request to "/items/200/participant/30/thread/log"
    Then the response code should be 200
    And the response body should be, in JSON:
    """
    [
      {
        "activity_type": "result_validated",
        "at": "2017-05-30T12:00:00Z",
        "attempt_id": "0",
        "can_watch_answer": false,
        "from_answer_id": "-1",
        "item": {
          "id": "200",
          "string": {
            "title": "Task 1"
          },
          "type": "Task"
        },
        "participant": {
          "id": "30",
          "name": "Our Team",
          "type": "Team"
        }
      },
      {
        "activity_type": "result_started",
        "at": "2017-05-29T06:38:00Z",
        "attempt_id": "0",
        "can_watch_answer": false,
        "from_answer_id": "-1",
        "item": {
          "id": "200",
          "string": {
            "title": "Task 1"
          },
          "type": "Task"
        },
        "participant": {
          "id": "30",
          "name": "Our Team",
          "type": "Team"
        }
      }
    ]
    """
