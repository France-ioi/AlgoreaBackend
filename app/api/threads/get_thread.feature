Feature: Get thread
  Background:
    Given the database has the following table "groups":
      | id | name       | type  |
      | 10 | Group      | Class |
      | 20 | Help Group | Class |
    And the database has the following users:
      | group_id | login   |
      | 1        | john    |
      | 2        | manager |
      | 3        | jack    |
      | 4        | helper  |
    And the database has the following table "groups_groups":
      | parent_group_id | child_group_id |
      | 20              | 4              |
    And the groups ancestors are computed
    And the database has the following table "items":
      | id | default_language_tag |
      | 10 | en                   |
      | 20 | en                   |
      | 21 | en                   |
      | 22 | en                   |
      | 30 | en                   |
      | 40 | en                   |
      | 50 | en                   |
    And the database has the following table "results":
      | attempt_id | participant_id | item_id | validated_at        |
      | 0          | 4              | 40      | 2020-01-01 00:00:00 |
      | 0          | 4              | 50      | 2020-01-01 00:00:00 |
    And the database has the following table "permissions_generated":
      | group_id | item_id | can_view_generated       | can_watch_generated |
      | 1        | 10      | content                  | none                |
      | 1        | 21      | content_with_descendants | none                |
      | 2        | 30      | content                  | answer              |
      | 4        | 40      | content                  | result              |
      | 20       | 50      | content                  | result              |
    And the database has the following table "threads":
      | item_id | participant_id | status                  | helper_group_id | latest_update_at    |
      | 10      | 2              | closed                  | 2               | 2019-01-01 00:00:00 |
      | 20      | 1              | closed                  | 10              | 2019-01-01 00:00:00 |
      | 21      | 1              | waiting_for_trainer     | 10              | 2019-01-01 00:00:00 |
      | 22      | 1              | closed                  | 10              | 2019-01-01 00:00:00 |
      | 30      | 1              | waiting_for_trainer     | 1               | 2019-01-01 00:00:00 |
      | 40      | 3              | waiting_for_participant | 20              | 2019-01-01 00:00:00 |
      | 50      | 3              | closed                  | 20              | 2020-01-10 00:00:00 |
    And the time now is "2020-01-20T00:00:00Z"

  Scenario: Should return all fields when the thread exists (also checks that token.can_write=true when the current user is the participant)
    Given I am the user with id "1"
    When I send a GET request to "/items/21/participant/1/thread"
    Then the response code should be 200
    And "threadToken" is a token signed by the app with the following payload:
      """
        {
          "can_watch": false,
          "can_write": true,
          "date": "20-01-2020",
          "exp": "1579485600",
          "is_mine": true,
          "item_id": "21",
          "participant_id": "1",
          "user_id": "1"
        }
      """
    And the response body should be, in JSON:
      """
        {
          "participant_id": "1",
          "item_id": "21",
          "status": "waiting_for_trainer",
          "token": "{{threadToken}}"
        }
      """

  Scenario: Status should be not_started if the thread doesn't exists
    Given I am the user with id "1"
    When I send a GET request to "/items/10/participant/1/thread"
    Then the response code should be 200
    And the response at $.status should be "not_started"

  Scenario: The current user is not the thread's participant but has can_watch>=answer permission on the item
    Given I am the user with id "2"
    When I send a GET request to "/items/30/participant/1/thread"
    Then the response code should be 200
    And the response at $.participant_id should be "1"
    And the response at $.item_id should be "30"
    And the response at $.status should be "waiting_for_trainer"

  Scenario: >
    The current user is a descendant of the thread's helper group
      and can watch for results and the thread is open
      and the current user has a validated result on the item
    Given I am the user with id "4"
    When I send a GET request to "/items/40/participant/3/thread"
    Then the response code should be 200
    And the response at $.participant_id should be "3"
    And the response at $.item_id should be "40"
    And the response at $.status should be "waiting_for_participant"

  Scenario: >
      The current user is a descendant of the thread's helper group
      and the current user can watch for results
      and the thread is closed for less than 2 weeks
      and the current user has a validated result on the item
    Given I am the user with id "4"
    When I send a GET request to "/items/50/participant/3/thread"
    Then the response code should be 200
    And the response at $.participant_id should be "3"
    And the response at $.item_id should be "50"
    And the response at $.status should be "closed"

  Scenario: >
    token.can_watch=true and token.can_write=true
      when the user has can_watch>=answer permission on the item
      and the user can watch for the participant
    Given I am @User
    And the group @User is a child of the group @ParentGroup
    And I can view content of the item 50
    And the group @ParentGroup can view content of the item 50
    And the group @ParentGroup has the watch permission set to "answer" on the item 50
    And the group @ParentGroup is a manager of the group @Participant and can watch for submissions from the group and its descendants
    And I am a manager of the group @Participant and can watch for submissions from the group and its descendants
    And the database table "threads" also has the following row:
      | item_id | participant_id | status                  | helper_group_id | latest_update_at    |
      | 50      | @Participant   | waiting_for_trainer     | 2               | 2020-01-10 00:00:00 |
    When I send a GET request to "/items/50/participant/@Participant/thread"
    Then the response code should be 200
    And "threadToken" is a token signed by the app with the following payload:
      """
        {
          "can_watch": true,
          "can_write": true,
          "date": "20-01-2020",
          "exp": "1579485600",
          "is_mine": false,
          "item_id": "50",
          "participant_id": "@Participant",
          "user_id": "@User"
        }
      """
    And the response body should be, in JSON:
      """
        {
          "participant_id": "@Participant",
          "item_id": "50",
          "status": "waiting_for_trainer",
          "token": "{{threadToken}}"
        }
      """

  Scenario: >
    token.can_watch=false and token.can_write=true
      when the user has can_watch=result permission and
      a validated result on the item and
      the user can watch for the participant and
      the user is a descendant of the thread's helper group
    Given I am @User
    And the group @User is a child of the group @ParentGroup
    And I can view content of the item 50
    And the group @ParentGroup can view content of the item 50
    And the group @ParentGroup has the watch permission set to "result" on the item 50
    And I have a validated result on the item 50
    And the group @ParentGroup is a manager of the group @Participant and can watch for submissions from the group and its descendants
    And I am a manager of the group @Participant and can watch for submissions from the group and its descendants
    And the database table "threads" also has the following row:
      | item_id | participant_id | status                  | helper_group_id | latest_update_at    |
      | 50      | @Participant   | waiting_for_trainer     | @ParentGroup    | 2020-01-10 00:00:00 |
    When I send a GET request to "/items/50/participant/@Participant/thread"
    Then the response code should be 200
    And "threadToken" is a token signed by the app with the following payload:
      """
        {
          "can_watch": false,
          "can_write": true,
          "date": "20-01-2020",
          "exp": "1579485600",
          "is_mine": false,
          "item_id": "50",
          "participant_id": "@Participant",
          "user_id": "@User"
        }
      """
    And the response body should be, in JSON:
      """
        {
          "participant_id": "@Participant",
          "item_id": "50",
          "status": "waiting_for_trainer",
          "token": "{{threadToken}}"
        }
      """

  Scenario: token.can_watch=false and token.can_write=false when the user has can_watch>=answer permission on the item, but the user cannot watch for the participant
    Given I am @User
    And the group @User is a child of the group @ParentGroup
    And I can view content of the item 50
    And the group @ParentGroup can view content of the item 50
    And the group @ParentGroup has the watch permission set to "answer" on the item 50
    And the database table "threads" also has the following row:
      | item_id | participant_id | status                  | helper_group_id | latest_update_at    |
      | 50      | @Participant   | waiting_for_trainer     | 2               | 2020-01-10 00:00:00 |
    When I send a GET request to "/items/50/participant/@Participant/thread"
    Then the response code should be 200
    And "threadToken" is a token signed by the app with the following payload:
      """
        {
          "can_watch": false,
          "can_write": false,
          "date": "20-01-2020",
          "exp": "1579485600",
          "is_mine": false,
          "item_id": "50",
          "participant_id": "@Participant",
          "user_id": "@User"
        }
      """
    And the response body should be, in JSON:
      """
        {
          "participant_id": "@Participant",
          "item_id": "50",
          "status": "waiting_for_trainer",
          "token": "{{threadToken}}"
        }
      """

  Scenario: token.can_write=true when the user has can_watch>=answer permission on the item and the user is a descendant of the thread's helper group
    Given I am @User
    And the group @User is a child of the group @ParentGroup
    And I can view content of the item 50
    And the group @ParentGroup can view content of the item 50
    And the group @ParentGroup has the watch permission set to "answer" on the item 50
    And the database table "threads" also has the following row:
      | item_id | participant_id | status                  | helper_group_id | latest_update_at    |
      | 50      | @Participant   | waiting_for_trainer     | @ParentGroup    | 2020-01-10 00:00:00 |
    When I send a GET request to "/items/50/participant/@Participant/thread"
    Then the response code should be 200
    And "threadToken" is a token signed by the app with the following payload:
      """
        {
          "can_watch": false,
          "can_write": true,
          "date": "20-01-2020",
          "exp": "1579485600",
          "is_mine": false,
          "item_id": "50",
          "participant_id": "@Participant",
          "user_id": "@User"
        }
      """
    And the response body should be, in JSON:
      """
        {
          "participant_id": "@Participant",
          "item_id": "50",
          "status": "waiting_for_trainer",
          "token": "{{threadToken}}"
        }
      """
