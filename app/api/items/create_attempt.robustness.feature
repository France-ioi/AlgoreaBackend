Feature: Create an attempt for an item - robustness
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

  Scenario: Invalid item_id
    Given I am the user with id "101"
    When I send a POST request to "/items/abc/attempts"
    Then the response code should be 400
    And the response error message should contain "Wrong value for item_id (should be int64)"
    And the table "users_items" should stay unchanged
    And the table "attempts" should stay unchanged

  Scenario: Invalid as_team_id
    Given I am the user with id "101"
    When I send a POST request to "/items/50/attempts?as_team_id=abc"
    Then the response code should be 400
    And the response error message should contain "Wrong value for as_team_id (should be int64)"

  Scenario: User not found
    Given I am the user with id "404"
    When I send a POST request to "/items/50/attempts"
    Then the response code should be 401
    And the response error message should contain "Invalid access token"
    And the table "users_items" should stay unchanged
    And the table "attempts" should stay unchanged

  Scenario: No access to the item (no item)
    Given I am the user with id "101"
    When I send a POST request to "/items/404/attempts"
    Then the response code should be 403
    And the response error message should contain "Insufficient access rights"
    And the table "users_items" should stay unchanged
    And the table "attempts" should stay unchanged

  Scenario: No access to the item (info access)
    Given I am the user with id "101"
    When I send a POST request to "/items/50/attempts"
    Then the response code should be 403
    And the response error message should contain "Insufficient access rights"
    And the table "users_items" should stay unchanged
    And the table "attempts" should stay unchanged

  Scenario: No access to the item (as a team)
    Given I am the user with id "101"
    When I send a POST request to "/items/50/attempts?as_team_id=104"
    Then the response code should be 403
    And the response error message should contain "Insufficient access rights"
    And the table "users_items" should stay unchanged
    And the table "attempts" should stay unchanged

  Scenario: No access to the item (type='Root')
    Given I am the user with id "101"
    When I send a POST request to "/items/70/attempts"
    Then the response code should be 403
    And the response error message should contain "Insufficient access rights"
    And the table "users_items" should stay unchanged
    And the table "attempts" should stay unchanged

  Scenario: No access to the item (type='Category')
    Given I am the user with id "101"
    When I send a POST request to "/items/80/attempts"
    Then the response code should be 403
    And the response error message should contain "Insufficient access rights"
    And the table "users_items" should stay unchanged
    And the table "attempts" should stay unchanged

  Scenario: No access to the item (type='Chapter')
    Given I am the user with id "101"
    When I send a POST request to "/items/90/attempts"
    Then the response code should be 403
    And the response error message should contain "Insufficient access rights"
    And the table "users_items" should stay unchanged
    And the table "attempts" should stay unchanged

  Scenario: User is not a team member
    Given I am the user with id "101"
    When I send a POST request to "/items/60/attempts?as_team_id=102"
    Then the response code should be 403
    And the response error message should contain "Can't use given as_team_id as a user's team for the item"
    And the table "users_items" should stay unchanged
    And the table "attempts" should stay unchanged

  Scenario: as_team_id is not a team
    Given I am the user with id "101"
    When I send a POST request to "/items/60/attempts?as_team_id=103"
    Then the response code should be 403
    And the response error message should contain "Can't use given as_team_id as a user's team for the item"
    And the table "users_items" should stay unchanged
    And the table "attempts" should stay unchanged

  Scenario: as_team_id is a team for a different item
    Given I am the user with id "101"
    When I send a POST request to "/items/60/attempts?as_team_id=104"
    Then the response code should be 403
    And the response error message should contain "Can't use given as_team_id as a user's team for the item"
    And the table "users_items" should stay unchanged
    And the table "attempts" should stay unchanged

  Scenario: There is an attempt for the (group, item) pair already, but items.has_attempts = 0
    Given I am the user with id "101"
    And the database table 'permissions_generated' has also the following row:
      | group_id | item_id | can_view_generated |
      | 104      | 50      | content            |
    And the database has the following table 'attempts':
      | group_id | item_id | order |
      | 104      | 50      | 1     |
    When I send a POST request to "/items/50/attempts?as_team_id=104"
    Then the response code should be 422
    And the response error message should contain "The item doesn't allow multiple attempts"
    And the table "users_items" should stay unchanged
    And the table "attempts" should stay unchanged
