Feature: Get a task token with a refreshed attempt for an item - robustness
  Background:
    Given the database has the following table 'groups':
      | id  | team_item_id | type     |
      | 101 | null         | UserSelf |
      | 102 | 60           | Team     |
      | 103 | 60           | Class    |
      | 104 | 50           | Team     |
    And the database has the following table 'users':
      | login | group_id |
      | john  | 101      |
    And the database has the following table 'groups_groups':
      | parent_group_id | child_group_id |
      | 103             | 101            |
      | 104             | 101            |
    And the database has the following table 'groups_ancestors':
      | ancestor_group_id | child_group_id | is_self |
      | 101               | 101            | 1       |
      | 102               | 102            | 1       |
      | 103               | 101            | 0       |
      | 103               | 103            | 1       |
      | 104               | 101            | 0       |
      | 104               | 104            | 1       |
    And the database has the following table 'items':
      | id | url                                                                     | type     | has_attempts |
      | 50 | http://taskplatform.mblockelet.info/task.html?taskId=403449543672183936 | Task     | 0            |
      | 60 | http://taskplatform.mblockelet.info/task.html?taskId=403449543672183936 | Course   | 1            |
      | 70 | http://taskplatform.mblockelet.info/task.html?taskId=403449543672183936 | Root     | 1            |
      | 80 | http://taskplatform.mblockelet.info/task.html?taskId=403449543672183936 | Category | 1            |
      | 90 | http://taskplatform.mblockelet.info/task.html?taskId=403449543672183936 | Chapter  | 1            |
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
      | id | group_id | item_id | order | latest_activity_at  | started_at          | score_computed | score_obtained_at | validated_at |
      | 2  | 101      | 50      | 0     | 2018-05-29 06:38:38 | 2017-05-29 06:38:38 | 0              | null              | null         |
      | 3  | 101      | 70      | 0     | 2018-05-29 06:38:38 | 2017-05-29 06:38:38 | 0              | null              | null         |
      | 4  | 101      | 80      | 0     | 2018-05-29 06:38:38 | 2017-05-29 06:38:38 | 0              | null              | null         |
      | 5  | 101      | 90      | 0     | 2018-05-29 06:38:38 | 2017-05-29 06:38:38 | 0              | null              | null         |
      | 6  | 102      | 60      | 0     | 2018-05-29 06:38:38 | 2017-05-29 06:38:38 | 0              | null              | null         |
      | 7  | 103      | 60      | 0     | 2018-05-29 06:38:38 | 2017-05-29 06:38:38 | 0              | null              | null         |
      | 8  | 104      | 60      | 0     | 2018-05-29 06:38:38 | 2017-05-29 06:38:38 | 0              | null              | null         |
    And time is frozen

  Scenario: Invalid attempt_id
    Given I am the user with id "101"
    When I send a GET request to "/attempts/abc/task-token"
    Then the response code should be 400
    And the response error message should contain "Wrong value for attempt_id (should be int64)"
    And the table "attempts" should stay unchanged

  Scenario: User not found
    Given I am the user with id "404"
    When I send a GET request to "/attempts/2/task-token"
    Then the response code should be 401
    And the response error message should contain "Invalid access token"
    And the table "attempts" should stay unchanged

  Scenario: No attempt
    Given I am the user with id "101"
    When I send a GET request to "/attempts/404/task-token"
    Then the response code should be 403
    And the response error message should contain "Insufficient access rights"
    And the table "attempts" should stay unchanged

  Scenario: No access to the item (info access)
    Given I am the user with id "101"
    When I send a GET request to "/attempts/2/task-token"
    Then the response code should be 403
    And the response error message should contain "Insufficient access rights"
    And the table "attempts" should stay unchanged

  Scenario: No access to the item (type='Root')
    Given I am the user with id "101"
    When I send a GET request to "/attempts/3/task-token"
    Then the response code should be 403
    And the response error message should contain "Insufficient access rights"
    And the table "attempts" should stay unchanged

  Scenario: No access to the item (type='Category')
    Given I am the user with id "101"
    When I send a GET request to "/attempts/4/task-token"
    Then the response code should be 403
    And the response error message should contain "Insufficient access rights"
    And the table "attempts" should stay unchanged

  Scenario: No access to the item (type='Chapter')
    Given I am the user with id "101"
    When I send a GET request to "/attempts/5/task-token"
    Then the response code should be 403
    And the response error message should contain "Insufficient access rights"
    And the table "attempts" should stay unchanged

  Scenario: User is not a team member
    Given I am the user with id "101"
    When I send a GET request to "/attempts/6/task-token"
    Then the response code should be 403
    And the response error message should contain "Insufficient access rights"
    And the table "attempts" should stay unchanged

  Scenario: Attempt group is not a team
    Given I am the user with id "101"
    When I send a GET request to "/attempts/7/task-token"
    Then the response code should be 403
    And the response error message should contain "Insufficient access rights"
    And the table "attempts" should stay unchanged

  Scenario: as_team_id is a team for a different item
    Given I am the user with id "101"
    When I send a GET request to "/attempts/8/task-token"
    Then the response code should be 403
    And the response error message should contain "Insufficient access rights"
    And the table "attempts" should stay unchanged
