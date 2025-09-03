Feature: Get an answer by id
  Background:
    Given the database has the following table "items":
      | id     | default_language_tag |
      | @Item1 | fr                   |
      | @Item2 | fr                   |
    And the time now is "2024-10-07T20:13:14Z"

  Scenario: User has can_view>=content on the item and the answers.participant_id = authenticated user's self group
    Given I am @User
    And I can view content of the item @Item1
    And the database has the following table "answers":
      | id  | author_id | participant_id | attempt_id | item_id | type       | state  | answer   | created_at          |
      | 104 | @Author   | @User          | 2          | @Item1  | Submission | State1 | print(3) | 2017-05-29 06:38:39 |
    And the database has the following table "gradings":
      | answer_id | score | graded_at           |
      | 104       | 95    | 2017-05-29 06:38:40 |
    When I send a GET request to "/answers/104"
    Then the response code should be 200
    And the response body should be, in JSON:
    """
    {
      "id": "104",
      "attempt_id": "2",
      "participant_id": "@User",
      "score": 95,
      "answer": "print(3)",
      "state": "State1",
      "created_at": "2017-05-29T06:38:39Z",
      "type": "Submission",
      "item_id": "@Item1",
      "author_id": "@Author",
      "graded_at": "2017-05-29T06:38:40Z"
    }
    """

  Scenario: User has can_view>=content on the item (via an ancestor group) and the answers.participant_id = authenticated user's self group
    Given I am @User
    And I am a member of the group @ParentGroup
    And the group @ParentGroup can view content of the item @Item1
    And the database has the following table "answers":
      | id  | author_id | participant_id | attempt_id | item_id | type       | state  | answer   | created_at          |
      | 104 | @Author   | @User          | 2          | @Item1  | Submission | State1 | print(3) | 2017-05-29 06:38:39 |
    When I send a GET request to "/answers/104"
    Then the response code should be 200
    And the response body should be, in JSON:
    """
    {
      "id": "104",
      "attempt_id": "2",
      "participant_id": "@User",
      "score": null,
      "answer": "print(3)",
      "state": "State1",
      "created_at": "2017-05-29T06:38:39Z",
      "type": "Submission",
      "item_id": "@Item1",
      "author_id": "@Author",
      "graded_at": null
    }
    """

  Scenario: User has can_view>=content on the item and the user is a team member of attempts.participant_id
    Given I am @User
    And I can view content of the item @Item2
    And there is a team @Team
    And I am a member of the group @Team
    And the database has the following table "answers":
      | id  | author_id | participant_id | attempt_id | item_id | type       | state  | answer   | created_at          |
      | 104 | @Author   | @Team          | 2          | @Item2  | Submission | State1 | print(3) | 2017-05-29 06:38:39 |
    When I send a GET request to "/answers/104"
    Then the response code should be 200
    And the response body should be, in JSON:
    """
    {
      "id": "104",
      "attempt_id": "2",
      "participant_id": "@Team",
      "score": null,
      "answer": "print(3)",
      "state": "State1",
      "created_at": "2017-05-29T06:38:39Z",
      "type": "Submission",
      "item_id": "@Item2",
      "author_id": "@Author",
      "graded_at": null
    }
    """

  Scenario: One of the user's teams has can_view>=content on the item and the user is a team member of attempts.participant_id
    Given I am @User
    And there is a team @Team1
    And there is a team @Team2
    And I am a member of the group @Team1
    And I am a member of the group @Team2
    And the group @Team1 can view content of the item @Item2
    And the database has the following table "answers":
      | id  | author_id | participant_id | attempt_id | item_id | type       | state  | answer   | created_at          |
      | 104 | @Author   | @Team2         | 2          | @Item2  | Submission | State1 | print(3) | 2017-05-29 06:38:39 |
    When I send a GET request to "/answers/104"
    Then the response code should be 200
    And the response body should be, in JSON:
    """
    {
      "id": "104",
      "attempt_id": "2",
      "participant_id": "@Team2",
      "score": null,
      "answer": "print(3)",
      "state": "State1",
      "created_at": "2017-05-29T06:38:39Z",
      "type": "Submission",
      "item_id": "@Item2",
      "author_id": "@Author",
      "graded_at": null
    }
    """

  Scenario: One of the user's teams has can_view>=content (via an ancestor) on the item and the user is a team member of attempts.participant_id
    Given I am @User
    And there is a team @Team1
    And there is a team @Team2
    And I am a member of the group @Team1
    And I am a member of the group @Team2
    And the group @Team1 is a child of the group @TeamParent
    And the group @TeamParent can view content of the item @Item2
    And the database has the following table "answers":
      | id  | author_id | participant_id | attempt_id | item_id | type       | state  | answer   | created_at          |
      | 104 | @Author   | @Team2         | 2          | @Item2  | Submission | State1 | print(3) | 2017-05-29 06:38:39 |
    When I send a GET request to "/answers/104"
    Then the response code should be 200
    And the response body should be, in JSON:
    """
    {
      "id": "104",
      "attempt_id": "2",
      "participant_id": "@Team2",
      "score": null,
      "answer": "print(3)",
      "state": "State1",
      "created_at": "2017-05-29T06:38:39Z",
      "type": "Submission",
      "item_id": "@Item2",
      "author_id": "@Author",
      "graded_at": null
    }
    """

  Scenario: User has can_watch>=answer on the item and can_watch_members on the participant
    Given I am @User
    And I have the watch permission set to "answer" on the item @Item2
    And I am a manager of the group @Participant and can watch for submissions from the group and its descendants
    And the database has the following table "answers":
      | id  | author_id | participant_id | attempt_id | item_id | type       | state  | answer   | created_at          |
      | 104 | @Author   | @Participant   | 2          | @Item2  | Submission | State1 | print(3) | 2017-05-29 06:38:39 |
    When I send a GET request to "/answers/104"
    Then the response code should be 200
    And the response body should be, in JSON:
    """
    {
      "id": "104",
      "attempt_id": "2",
      "participant_id": "@Participant",
      "score": null,
      "answer": "print(3)",
      "state": "State1",
      "created_at": "2017-05-29T06:38:39Z",
      "type": "Submission",
      "item_id": "@Item2",
      "author_id": "@Author",
      "graded_at": null
    }
    """

  Scenario: User has can_watch>=answer (via an ancestor) on the item and can_watch_members (via an ancestor) on the participant
    Given I am @User
    And I am a member of the group @ChildGroupAbleToWatch
    And the group @ChildGroupAbleToWatch is a child of the group @GroupAbleToWatch
    And the group @GroupAbleToWatch has the watch permission set to "answer" on the item @Item2
    And I am a member of the group @ManagerGroup
    And @Participant is a member of the group @ManagedGroup
    And the group @ManagerGroup is a manager of the group @ManagedGroup and can watch for submissions from the group and its descendants
    And the database has the following table "answers":
      | id  | author_id | participant_id | attempt_id | item_id | type       | state  | answer   | created_at          |
      | 104 | @Author   | @Participant   | 2          | @Item2  | Submission | State1 | print(3) | 2017-05-29 06:38:39 |
    When I send a GET request to "/answers/104"
    Then the response code should be 200
    And the response body should be, in JSON:
    """
    {
      "id": "104",
      "attempt_id": "2",
      "participant_id": "@Participant",
      "score": null,
      "answer": "print(3)",
      "state": "State1",
      "created_at": "2017-05-29T06:38:39Z",
      "type": "Submission",
      "item_id": "@Item2",
      "author_id": "@Author",
      "graded_at": null
    }
    """

  Scenario: User has can_watch>=result on the item and can_watch_members on the participant, and has a validated result on the item
    Given I am @User
    And I have the watch permission set to "result" on the item @Item2
    And I am a manager of the group @Participant and can watch for submissions from the group and its descendants
    And I have a validated result on the item @Item2
    And the database has the following table "answers":
      | id  | author_id | participant_id | attempt_id | item_id | type       | state  | answer   | created_at          |
      | 104 | @Author   | @Participant   | 2          | @Item2  | Submission | State1 | print(3) | 2017-05-29 06:38:39 |
    When I send a GET request to "/answers/104"
    Then the response code should be 200
    And the response body should be, in JSON:
    """
    {
      "id": "104",
      "attempt_id": "2",
      "participant_id": "@Participant",
      "score": null,
      "answer": "print(3)",
      "state": "State1",
      "created_at": "2017-05-29T06:38:39Z",
      "type": "Submission",
      "item_id": "@Item2",
      "author_id": "@Author",
      "graded_at": null
    }
    """

  Scenario: User has can_watch>=result on the item (via an ancestor group) and can_watch_members on the participant, and has a validated result on the item
    Given I am @User
    And I am a member of the group @ParentGroup
    And the group @ParentGroup has the watch permission set to "result" on the item @Item2
    And I am a manager of the group @Participant and can watch for submissions from the group and its descendants
    And I have a validated result on the item @Item2
    And the database has the following table "answers":
      | id  | author_id | participant_id | attempt_id | item_id | type       | state  | answer   | created_at          |
      | 104 | @Author   | @Participant   | 2          | @Item2  | Submission | State1 | print(3) | 2017-05-29 06:38:39 |
    When I send a GET request to "/answers/104"
    Then the response code should be 200
    And the response body should be, in JSON:
    """
    {
      "id": "104",
      "attempt_id": "2",
      "participant_id": "@Participant",
      "score": null,
      "answer": "print(3)",
      "state": "State1",
      "created_at": "2017-05-29T06:38:39Z",
      "type": "Submission",
      "item_id": "@Item2",
      "author_id": "@Author",
      "graded_at": null
    }
    """

  Scenario: User has can_watch>=result on the item and can_watch_members on the participant, and one of the user's teams has a validated result on the item
    Given I am @User
    And I have the watch permission set to "result" on the item @Item2
    And I am a manager of the group @Participant and can watch for submissions from the group and its descendants
    And there is a team @Team1
    And there is a team @Team2
    And I am a member of the group @Team1
    And I am a member of the group @Team2
    And the group @Team2 has a validated result on the item @Item2
    And the database has the following table "answers":
      | id  | author_id | participant_id | attempt_id | item_id | type       | state  | answer   | created_at          |
      | 104 | @Author   | @Participant   | 2          | @Item2  | Submission | State1 | print(3) | 2017-05-29 06:38:39 |
    When I send a GET request to "/answers/104"
    Then the response code should be 200
    And the response body should be, in JSON:
    """
    {
      "id": "104",
      "attempt_id": "2",
      "participant_id": "@Participant",
      "score": null,
      "answer": "print(3)",
      "state": "State1",
      "created_at": "2017-05-29T06:38:39Z",
      "type": "Submission",
      "item_id": "@Item2",
      "author_id": "@Author",
      "graded_at": null
    }
    """

  Scenario: User has can_watch>=answer on the item and the thread exists
    Given I am @User
    And I have the watch permission set to "answer" on the item @Item2
    And I am a member of the group @Helper
    And there is a thread with "item_id=@Item2,participant_id=@Participant,helper_group_id=@Author,status=closed,latest_update_at={{relativeTimeDB("-1000h")}}"
    And the database has the following table "answers":
      | id  | author_id | participant_id | attempt_id | item_id | type       | state  | answer   | created_at          |
      | 104 | @Author   | @Participant   | 2          | @Item2  | Submission | State1 | print(3) | 2017-05-29 06:38:39 |
    When I send a GET request to "/answers/104"
    Then the response code should be 200
    And the response body should be, in JSON:
    """
    {
      "id": "104",
      "attempt_id": "2",
      "participant_id": "@Participant",
      "score": null,
      "answer": "print(3)",
      "state": "State1",
      "created_at": "2017-05-29T06:38:39Z",
      "type": "Submission",
      "item_id": "@Item2",
      "author_id": "@Author",
      "graded_at": null
    }
    """

  Scenario: User has can_watch>=answer (via an ancestor) on the item and the thread exists
    Given I am @User
    And I am a member of the group @ChildGroupAbleToWatch
    And the group @ChildGroupAbleToWatch is a child of the group @GroupAbleToWatch
    And the group @GroupAbleToWatch has the watch permission set to "answer" on the item @Item2
    And there is a thread with "item_id=@Item2,participant_id=@Participant,helper_group_id=@Author,status=closed,latest_update_at={{relativeTimeDB("-1000h")}}"
    And the database has the following table "answers":
      | id  | author_id | participant_id | attempt_id | item_id | type       | state  | answer   | created_at          |
      | 104 | @Author   | @Participant   | 2          | @Item2  | Submission | State1 | print(3) | 2017-05-29 06:38:39 |
    When I send a GET request to "/answers/104"
    Then the response code should be 200
    And the response body should be, in JSON:
    """
    {
      "id": "104",
      "attempt_id": "2",
      "participant_id": "@Participant",
      "score": null,
      "answer": "print(3)",
      "state": "State1",
      "created_at": "2017-05-29T06:38:39Z",
      "type": "Submission",
      "item_id": "@Item2",
      "author_id": "@Author",
      "graded_at": null
    }
    """

  Scenario Outline: User has can_watch>=result on the item and is a thread reader, and has a validated result on the item
    Given I am @User
    And I have the watch permission set to "result" on the item @Item2
    And I have a validated result on the item @Item2
    And I am a member of the group @Helper
    And there is a thread with "item_id=@Item2,participant_id=@Participant,helper_group_id=@Helper,status=<thread_status>,latest_update_at=<thread_latest_update_at>"
    And the database has the following table "answers":
      | id  | author_id | participant_id | attempt_id | item_id | type       | state  | answer   | created_at          |
      | 104 | @Author   | @Participant   | 2          | @Item2  | Submission | State1 | print(3) | 2017-05-29 06:38:39 |
    When I send a GET request to "/answers/104"
    Then the response code should be 200
    And the response body should be, in JSON:
    """
    {
      "id": "104",
      "attempt_id": "2",
      "participant_id": "@Participant",
      "score": null,
      "answer": "print(3)",
      "state": "State1",
      "created_at": "2017-05-29T06:38:39Z",
      "type": "Submission",
      "item_id": "@Item2",
      "author_id": "@Author",
      "graded_at": null
    }
    """
  Examples:
    | thread_status           | thread_latest_update_at        |
    | waiting_for_participant | 2020-05-30 12:00:00            |
    | waiting_for_trainer     | 2020-05-30 12:00:00            |
    | closed                  | {{relativeTimeDB("-335h59m")}} |

  Scenario Outline: User has can_watch>=result on the item (via an ancestor group) and is a thread reader, and has a validated result on the item
    Given I am @User
    And I am a member of the group @ParentGroup
    And the group @ParentGroup has the watch permission set to "result" on the item @Item2
    And I am a member of the group @Helper
    And there is a thread with "item_id=@Item2,participant_id=@Participant,helper_group_id=@Helper,status=<thread_status>,latest_update_at=<thread_latest_update_at>"
    And I have a validated result on the item @Item2
    And the database has the following table "answers":
      | id  | author_id | participant_id | attempt_id | item_id | type       | state  | answer   | created_at          |
      | 104 | @Author   | @Participant   | 2          | @Item2  | Submission | State1 | print(3) | 2017-05-29 06:38:39 |
    When I send a GET request to "/answers/104"
    Then the response code should be 200
    And the response body should be, in JSON:
    """
    {
      "id": "104",
      "attempt_id": "2",
      "participant_id": "@Participant",
      "score": null,
      "answer": "print(3)",
      "state": "State1",
      "created_at": "2017-05-29T06:38:39Z",
      "type": "Submission",
      "item_id": "@Item2",
      "author_id": "@Author",
      "graded_at": null
    }
    """
  Examples:
    | thread_status           | thread_latest_update_at        |
    | waiting_for_participant | 2020-05-30 12:00:00            |
    | waiting_for_trainer     | 2020-05-30 12:00:00            |
    | closed                  | {{relativeTimeDB("-335h59m")}} |

  Scenario Outline: User has can_watch>=result on the item and is a thread reader, and one of the user's teams has a validated result on the item
    Given I am @User
    And I have the watch permission set to "result" on the item @Item2
    And I am a member of the group @Helper
    And there is a thread with "item_id=@Item2,participant_id=@Participant,helper_group_id=@Helper,status=<thread_status>,latest_update_at=<thread_latest_update_at>"
    And there is a team @Team1
    And there is a team @Team2
    And I am a member of the group @Team1
    And I am a member of the group @Team2
    And the group @Team2 has a validated result on the item @Item2
    And the database has the following table "answers":
      | id  | author_id | participant_id | attempt_id | item_id | type       | state  | answer   | created_at          |
      | 104 | @Author   | @Participant   | 2          | @Item2  | Submission | State1 | print(3) | 2017-05-29 06:38:39 |
    When I send a GET request to "/answers/104"
    Then the response code should be 200
    And the response body should be, in JSON:
    """
    {
      "id": "104",
      "attempt_id": "2",
      "participant_id": "@Participant",
      "score": null,
      "answer": "print(3)",
      "state": "State1",
      "created_at": "2017-05-29T06:38:39Z",
      "type": "Submission",
      "item_id": "@Item2",
      "author_id": "@Author",
      "graded_at": null
    }
    """
  Examples:
    | thread_status           | thread_latest_update_at        |
    | waiting_for_participant | 2020-05-30 12:00:00            |
    | waiting_for_trainer     | 2020-05-30 12:00:00            |
    | closed                  | {{relativeTimeDB("-335h59m")}} |
