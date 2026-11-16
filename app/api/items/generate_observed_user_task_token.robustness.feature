Feature: Generate a read-only task token for an observed user's attempt - robustness
  Background:
    Given the database has the following table "groups":
      | id  | name  | type  |
      | 102 | team  | Team  |
      | 104 | class | Class |
    And the database has the following users:
      | group_id | login    |
      | 101      | john     |
      | 103      | manager  |
      | 105      | stranger |
    And the database has the following table "groups_groups":
      | parent_group_id | child_group_id |
      | 104             | 101            |
      | 104             | 102            |
    And the database has the following table "group_managers":
      | group_id | manager_id | can_watch_members |
      | 104      | 103        | 1                 |
    And the groups ancestors are computed
    And the database has the following table "items":
      | id | url                    | type    | default_language_tag | text_id | hints_allowed | supported_lang_prog |
      | 50 | http://taskplatform/50 | Task    | fr                   | task50  | 1             | c,python            |
      | 60 | http://taskplatform/60 | Task    | fr                   | task60  | 0             | null                |
      | 90 | http://taskplatform/90 | Chapter | fr                   | null    | 0             | null                |
    And the database has the following table "permissions_generated":
      | group_id | item_id | can_view_generated | can_watch_generated |
      | 103      | 50      | content            | answer              |
      | 103      | 60      | content            | result              |
      | 103      | 90      | content            | answer              |
      | 101      | 50      | content            | answer              |
      | 105      | 50      | content            | answer              |
    And the database has the following table "attempts":
      | id | participant_id | allows_submissions_until |
      | 1  | 101            | 2019-05-30 11:00:00      |
      | 1  | 102            | 9999-12-31 23:59:59      |
      | 2  | 101            | 9999-12-31 23:59:59      |
    And the database has the following table "results":
      | attempt_id | participant_id | item_id | latest_activity_at  | started_at          | hints_requested | hints_cached | validated_at |
      | 1          | 101            | 50      | 2018-05-29 06:38:38 | 2018-05-29 05:00:00 | [1,2,3,4]       | 4            | null         |
      | 1          | 101            | 60      | 2018-05-29 06:38:38 | 2018-05-29 05:00:00 | null            | 0            | null         |
      | 1          | 101            | 90      | 2018-05-29 06:38:38 | 2018-05-29 05:00:00 | null            | 0            | null         |
      | 1          | 102            | 50      | 2018-05-29 06:38:38 | 2018-05-29 05:00:00 | null            | 0            | null         |
      | 2          | 101            | 50      | 2018-05-29 06:38:38 | null                | null            | 0            | null         |
    And the server time is frozen

  Scenario: Caller who may only watch results is refused
    Given I am the user with id "103"
    When I send a POST request to "/items/60/participant/101/attempts/1/generate-task-token"
    Then the response code should be 403
    And the response error message should contain "Insufficient access rights"
    And the table "results" should remain unchanged

  Scenario: Caller who cannot watch the participant is refused
    Given I am the user with id "105"
    When I send a POST request to "/items/50/participant/101/attempts/1/generate-task-token"
    Then the response code should be 403
    And the response error message should contain "Insufficient access rights"
    And the table "results" should remain unchanged

  Scenario: The participant cannot use this service for their own attempt
    Given I am the user with id "101"
    When I send a POST request to "/items/50/participant/101/attempts/1/generate-task-token"
    Then the response code should be 403
    And the response error message should contain "Insufficient access rights"
    And the table "results" should remain unchanged

  Scenario: A team participant is refused
    Given I am the user with id "103"
    When I send a POST request to "/items/50/participant/102/attempts/1/generate-task-token"
    Then the response code should be 403
    And the response error message should contain "Insufficient access rights"
    And the table "results" should remain unchanged

  Scenario: The item is not a task
    Given I am the user with id "103"
    When I send a POST request to "/items/90/participant/101/attempts/1/generate-task-token"
    Then the response code should be 403
    And the response error message should contain "Insufficient access rights"
    And the table "results" should remain unchanged

  Scenario: The attempt is unknown
    Given I am the user with id "103"
    When I send a POST request to "/items/50/participant/101/attempts/404/generate-task-token"
    Then the response code should be 403
    And the response error message should contain "Insufficient access rights"
    And the table "results" should remain unchanged

  Scenario: The result is not started
    Given I am the user with id "103"
    When I send a POST request to "/items/50/participant/101/attempts/2/generate-task-token"
    Then the response code should be 403
    And the response error message should contain "Insufficient access rights"
    And the table "results" should remain unchanged

  Scenario: Invalid item_id
    Given I am the user with id "103"
    When I send a POST request to "/items/abc/participant/101/attempts/1/generate-task-token"
    Then the response code should be 400
    And the response error message should contain "Wrong value for item_id (should be int64)"

  Scenario: Invalid participant_id
    Given I am the user with id "103"
    When I send a POST request to "/items/50/participant/abc/attempts/1/generate-task-token"
    Then the response code should be 400
    And the response error message should contain "Wrong value for participant_id (should be int64)"

  Scenario: Invalid attempt_id
    Given I am the user with id "103"
    When I send a POST request to "/items/50/participant/101/attempts/abc/generate-task-token"
    Then the response code should be 400
    And the response error message should contain "Wrong value for attempt_id (should be int64)"

  Scenario: User not found
    Given I am the user with id "404"
    When I send a POST request to "/items/50/participant/101/attempts/1/generate-task-token"
    Then the response code should be 401
    And the response error message should contain "Invalid access token"
