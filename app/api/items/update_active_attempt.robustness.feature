Feature: Update active attempt for an item - robustness
  Background:
    Given the database has the following table 'groups':
      | id  | team_item_id | type     |
      | 101 | null         | UserSelf |
      | 102 | 10           | Team     |
      | 103 | 10           | Team     |
      | 109 | 10           | Team     |
      | 111 | null         | UserSelf |
      | 121 | null         | UserSelf |
    And the database has the following table 'users':
      | login | group_id |
      | john  | 101      |
      | jane  | 111      |
      | guest | 121      |
    And the database has the following table 'groups_groups':
      | parent_group_id | child_group_id |
      | 102             | 101            |
      | 109             | 101            |
    And the database has the following table 'groups_ancestors':
      | ancestor_group_id | child_group_id | is_self |
      | 101               | 101            | 1       |
      | 102               | 101            | 0       |
      | 102               | 102            | 1       |
      | 111               | 111            | 1       |
      | 121               | 121            | 1       |
    And the database has the following table 'items':
      | id | url                                                                     | type    |
      | 10 | null                                                                    | Chapter |
      | 50 | http://taskplatform.mblockelet.info/task.html?taskId=403449543672183936 | Task    |
      | 60 | http://taskplatform.mblockelet.info/task.html?taskId=403449543672183936 | Course  |
    And the database has the following table 'items_ancestors':
      | ancestor_item_id | child_item_id |
      | 10               | 60            |
    And the database has the following table 'permissions_generated':
      | group_id | item_id | can_view_generated       |
      | 101      | 50      | content                  |
      | 101      | 60      | content                  |
      | 111      | 50      | content_with_descendants |
      | 121      | 50      | info                     |

  Scenario: Invalid attempt_id
    Given I am the user with id "101"
    When I send a PUT request to "/attempts/abc/active"
    Then the response code should be 400
    And the response error message should contain "Wrong value for attempt_id (should be int64)"
    And the table "users_items" should stay unchanged
    And the table "attempts" should stay unchanged

  Scenario: User not found
    Given I am the user with id "404"
    When I send a PUT request to "/attempts/100/active"
    Then the response code should be 401
    And the response error message should contain "Invalid access token"
    And the table "users_items" should stay unchanged
    And the table "attempts" should stay unchanged

  Scenario: User doesn't have access to the item
    Given I am the user with id "121"
    And the database has the following table 'attempts':
      | id  | group_id | item_id | order |
      | 100 | 121      | 50      | 1     |
      | 101 | 121      | 50      | 2     |
    And the database has the following table 'users_items':
      | user_id | item_id | active_attempt_id |
      | 121     | 50      | 101               |
    When I send a PUT request to "/attempts/100/active"
    Then the response code should be 403
    And the response error message should contain "Insufficient access rights"
    And the table "users_items" should stay unchanged
    And the table "attempts" should stay unchanged

  Scenario: User has only info access to the item
    Given I am the user with id "121"
    And the database has the following table 'attempts':
      | id  | group_id | item_id | order |
      | 100 | 121      | 50      | 1     |
      | 101 | 121      | 50      | 2     |
    And the database has the following table 'users_items':
      | user_id | item_id | active_attempt_id |
      | 121     | 50      | 101               |
    When I send a PUT request to "/attempts/100/active"
    Then the response code should be 403
    And the response error message should contain "Insufficient access rights"
    And the table "users_items" should stay unchanged
    And the table "attempts" should stay unchanged

  Scenario: No attempts
    Given I am the user with id "101"
    When I send a PUT request to "/attempts/100/active"
    Then the response code should be 403
    And the response error message should contain "Insufficient access rights"
    And the table "users_items" should stay unchanged
    And the table "attempts" should stay unchanged

  Scenario: Wrong item in attempts
    Given I am the user with id "101"
    And the database has the following table 'attempts':
      | id  | group_id | item_id | order |
      | 100 | 101      | 51      | 1     |
      | 101 | 101      | 50      | 2     |
    And the database has the following table 'users_items':
      | user_id | item_id | active_attempt_id |
      | 101     | 50      | 101               |
    When I send a PUT request to "/attempts/100/active"
    Then the response code should be 403
    And the response error message should contain "Insufficient access rights"
    And the table "users_items" should stay unchanged
    And the table "attempts" should stay unchanged

  Scenario: User is not a member of the team
    Given I am the user with id "101"
    And the database has the following table 'attempts':
      | id  | group_id | item_id | order |
      | 100 | 103      | 60      | 1     |
      | 200 | 102      | 60      | 2     |
    And the database has the following table 'users_items':
      | user_id | item_id | active_attempt_id |
      | 101     | 60      | 200               |
    When I send a PUT request to "/attempts/100/active"
    Then the response code should be 403
    And the response error message should contain "Insufficient access rights"
    And the table "users_items" should stay unchanged
    And the table "attempts" should stay unchanged
