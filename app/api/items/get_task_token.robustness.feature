Feature: Get a task token with a refreshed active attempt for an item - robustness
  Background:
    Given the database has the following table 'users':
      | id | login | self_group_id |
      | 10 | john  | 101           |
    And the database has the following table 'groups':
      | id  | team_item_id | type     |
      | 101 | null         | UserSelf |
      | 102 | 60           | Team     |
    And the database has the following table 'groups_ancestors':
      | ancestor_group_id | child_group_id | is_self |
      | 101               | 101            | 1       |
      | 102               | 102            | 1       |
    And the database has the following table 'items':
      | id | url                                                                     | type     | has_attempts |
      | 50 | http://taskplatform.mblockelet.info/task.html?taskId=403449543672183936 | Task     | 0            |
      | 60 | http://taskplatform.mblockelet.info/task.html?taskId=403449543672183936 | Course   | 1            |
      | 70 | http://taskplatform.mblockelet.info/task.html?taskId=403449543672183936 | Root     | 1            |
      | 80 | http://taskplatform.mblockelet.info/task.html?taskId=403449543672183936 | Category | 1            |
      | 90 | http://taskplatform.mblockelet.info/task.html?taskId=403449543672183936 | Chapter  | 1            |
    And the database has the following table 'groups_items':
      | group_id | item_id | cached_partial_access_since | creator_user_id |
      | 101      | 50      | 2017-05-29 06:38:38         | 10              |
      | 101      | 60      | 2017-05-29 06:38:38         | 10              |
      | 101      | 70      | 2017-05-29 06:38:38         | 10              |
      | 101      | 80      | 2017-05-29 06:38:38         | 10              |
      | 101      | 90      | 2017-05-29 06:38:38         | 10              |
    And time is frozen

  Scenario: Invalid item_id
    Given I am the user with id "10"
    When I send a GET request to "/items/abc/task-token"
    Then the response code should be 400
    And the response error message should contain "Wrong value for item_id (should be int64)"
    And the table "users_answers" should stay unchanged
    And the table "users_items" should stay unchanged
    And the table "groups_attempts" should stay unchanged

  Scenario: User not found
    Given I am the user with id "404"
    When I send a GET request to "/items/50/task-token"
    Then the response code should be 401
    And the response error message should contain "Invalid access token"
    And the table "users_answers" should stay unchanged
    And the table "users_items" should stay unchanged
    And the table "groups_attempts" should stay unchanged

  Scenario: No access to the item (no item)
    Given I am the user with id "10"
    When I send a GET request to "/items/404/task-token"
    Then the response code should be 403
    And the response error message should contain "Insufficient access rights"
    And the table "users_answers" should stay unchanged
    And the table "users_items" should stay unchanged
    And the table "groups_attempts" should stay unchanged

  Scenario: No access to the item (type='Root')
    Given I am the user with id "10"
    When I send a GET request to "/items/70/task-token"
    Then the response code should be 403
    And the response error message should contain "Insufficient access rights"
    And the table "users_answers" should stay unchanged
    And the table "users_items" should stay unchanged
    And the table "groups_attempts" should stay unchanged

  Scenario: No access to the item (type='Category')
    Given I am the user with id "10"
    When I send a GET request to "/items/80/task-token"
    Then the response code should be 403
    And the response error message should contain "Insufficient access rights"
    And the table "users_answers" should stay unchanged
    And the table "users_items" should stay unchanged
    And the table "groups_attempts" should stay unchanged

  Scenario: No access to the item (type='Chapter')
    Given I am the user with id "10"
    When I send a GET request to "/items/90/task-token"
    Then the response code should be 403
    And the response error message should contain "Insufficient access rights"
    And the table "users_answers" should stay unchanged
    And the table "users_items" should stay unchanged
    And the table "groups_attempts" should stay unchanged

  Scenario: User is not a team member
    Given I am the user with id "10"
    When I send a GET request to "/items/60/task-token"
    Then the response code should be 403
    And the response error message should contain "No team found for the user"
    And the table "users_answers" should stay unchanged
    And the table "users_items" should stay unchanged
    And the table "groups_attempts" should stay unchanged
