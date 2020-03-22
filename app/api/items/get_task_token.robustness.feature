Feature: Get a task token with a refreshed attempt for an item - robustness
  Background:
    Given the database has the following table 'groups':
      | id  | team_item_id | type  |
      | 101 | null         | User  |
      | 102 | 60           | Team  |
      | 103 | 60           | Class |
      | 104 | 50           | Team  |
    And the database has the following table 'users':
      | login | group_id |
      | john  | 101      |
    And the database has the following table 'groups_groups':
      | parent_group_id | child_group_id |
      | 103             | 101            |
      | 104             | 101            |
    And the database has the following table 'groups_ancestors':
      | ancestor_group_id | child_group_id |
      | 101               | 101            |
      | 102               | 102            |
      | 103               | 101            |
      | 103               | 103            |
      | 104               | 101            |
      | 104               | 104            |
    And the database has the following table 'items':
      | id | url                                                                     | type    | entry_participant_type | default_language_tag |
      | 50 | http://taskplatform.mblockelet.info/task.html?taskId=403449543672183936 | Task    | null                   | fr                   |
      | 60 | http://taskplatform.mblockelet.info/task.html?taskId=403449543672183936 | Course  | Team                   | fr                   |
      | 90 | http://taskplatform.mblockelet.info/task.html?taskId=403449543672183936 | Chapter | Team                   | fr                   |
    And the database has the following table 'permissions_generated':
      | group_id | item_id | can_view_generated |
      | 101      | 50      | info               |
      | 101      | 60      | content            |
      | 101      | 70      | content            |
      | 101      | 80      | content            |
      | 101      | 90      | content            |
      | 102      | 60      | content            |
      | 103      | 60      | content            |
      | 104      | 60      | content            |
    And the database has the following table 'attempts':
      | id | participant_id | created_at          |
      | 0  | 101            | 2017-05-29 05:38:38 |
      | 0  | 102            | 2017-05-29 05:38:38 |
      | 0  | 103            | 2017-05-29 05:38:38 |
      | 0  | 104            | 2017-05-29 05:38:38 |
    And the database has the following table 'results':
      | attempt_id | participant_id | item_id | latest_activity_at  | started_at          | score_computed | score_obtained_at | validated_at |
      | 0          | 101            | 50      | 2018-05-29 06:38:38 | 2017-05-29 06:38:38 | 0              | null              | null         |
      | 0          | 101            | 90      | 2018-05-29 06:38:38 | 2017-05-29 06:38:38 | 0              | null              | null         |
      | 0          | 102            | 60      | 2018-05-29 06:38:38 | 2017-05-29 06:38:38 | 0              | null              | null         |
      | 0          | 103            | 60      | 2018-05-29 06:38:38 | 2017-05-29 06:38:38 | 0              | null              | null         |
      | 0          | 104            | 60      | 2018-05-29 06:38:38 | 2017-05-29 06:38:38 | 0              | null              | null         |
    And time is frozen

  Scenario: Invalid attempt_id
    Given I am the user with id "101"
    When I send a GET request to "/items/50/attempts/abc/task-token"
    Then the response code should be 400
    And the response error message should contain "Wrong value for attempt_id (should be int64)"
    And the table "attempts" should stay unchanged

  Scenario: Invalid item_id
    Given I am the user with id "101"
    When I send a GET request to "/items/abc/attempts/0/task-token"
    Then the response code should be 400
    And the response error message should contain "Wrong value for item_id (should be int64)"
    And the table "attempts" should stay unchanged

  Scenario: Invalid as_team_id
    Given I am the user with id "101"
    When I send a GET request to "/items/50/attempts/0/task-token?as_team_id=abc"
    Then the response code should be 400
    And the response error message should contain "Wrong value for as_team_id (should be int64)"
    And the table "attempts" should stay unchanged

  Scenario: User not found
    Given I am the user with id "404"
    When I send a GET request to "/items/50/attempts/0/task-token"
    Then the response code should be 401
    And the response error message should contain "Invalid access token"
    And the table "attempts" should stay unchanged

  Scenario: No attempt
    Given I am the user with id "101"
    When I send a GET request to "/items/50/attempts/404/task-token"
    Then the response code should be 403
    And the response error message should contain "Insufficient access rights"
    And the table "attempts" should stay unchanged

  Scenario: No access to the item (info access)
    Given I am the user with id "101"
    When I send a GET request to "/items/50/attempts/0/task-token"
    Then the response code should be 403
    And the response error message should contain "Insufficient access rights"
    And the table "attempts" should stay unchanged

  Scenario: No access to the item (type='Chapter')
    Given I am the user with id "101"
    When I send a GET request to "/items/90/attempts/0/task-token"
    Then the response code should be 403
    And the response error message should contain "Insufficient access rights"
    And the table "attempts" should stay unchanged

  Scenario: User is not a team member
    Given I am the user with id "101"
    When I send a GET request to "/items/60/attempts/0/task-token?as_team_id=102"
    Then the response code should be 403
    And the response error message should contain "Insufficient access rights"
    And the table "attempts" should stay unchanged

  Scenario: Attempt group is not a team
    Given I am the user with id "101"
    When I send a GET request to "/items/60/attempts/0/task-token?as_team_id=103"
    Then the response code should be 403
    And the response error message should contain "Insufficient access rights"
    And the table "attempts" should stay unchanged

  Scenario: as_team_id is a team for a different item
    Given I am the user with id "101"
    When I send a GET request to "/items/60/attempts/0/task-token?as_team_id=104"
    Then the response code should be 403
    And the response error message should contain "Insufficient access rights"
    And the table "attempts" should stay unchanged

  Scenario: No result in the DB
    Given I am the user with id "101"
    When I send a GET request to "/items/60/attempts/0/task-token"
    Then the response code should be 403
    And the response error message should contain "Insufficient access rights"
    And the table "attempts" should stay unchanged
