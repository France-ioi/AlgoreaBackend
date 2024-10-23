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
      | 2        | 30      | none                     | answer              |
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

  Scenario: Should return all fields when the thread exists
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

  Scenario: The current-user is not the thread participant but has "can_watch >= answer" permission on the item
    Given I am the user with id "2"
    When I send a GET request to "/items/30/participant/1/thread"
    Then the response code should be 200
    And the response at $.participant_id should be "1"
    And the response at $.item_id should be "30"
    And the response at $.status should be "waiting_for_trainer"

  Scenario: The current-user is descendant of the thread helper group and the thread is open and user has validated the item
    Given I am the user with id "4"
    When I send a GET request to "/items/40/participant/3/thread"
    Then the response code should be 200
    And the response at $.participant_id should be "3"
    And the response at $.item_id should be "40"
    And the response at $.status should be "waiting_for_participant"

  Scenario: >
      The current-user is descendant of the thread helper group
      and the thread is closed for less than 2 weeks
      and user has validated the item
    Given I am the user with id "4"
    When I send a GET request to "/items/50/participant/3/thread"
    Then the response code should be 200
    And the response at $.participant_id should be "3"
    And the response at $.item_id should be "50"
    And the response at $.status should be "closed"
