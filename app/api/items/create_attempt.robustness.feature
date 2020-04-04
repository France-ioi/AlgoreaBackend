Feature: Create an attempt for an item - robustness
  Background:
    Given the database has the following table 'groups':
      | id  | type  |
      | 101 | User  |
      | 102 | Team  |
      | 103 | Class |
      | 104 | Team  |
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
      | id | url                                                                     | type   | allows_multiple_attempts | default_language_tag |
      | 50 | http://taskplatform.mblockelet.info/task.html?taskId=403449543672183936 | Task   | 0                        | fr                   |
      | 60 | http://taskplatform.mblockelet.info/task.html?taskId=403449543672183936 | Course | 1                        | fr                   |
      | 90 | http://taskplatform.mblockelet.info/task.html?taskId=403449543672183936 | Skill  | 1                        | fr                   |
    And the database has the following table 'permissions_generated':
      | group_id | item_id | can_view_generated |
      | 101      | 50      | info               |
      | 101      | 60      | content            |
      | 101      | 90      | content            |

  Scenario: Invalid item_id
    Given I am the user with id "101"
    When I send a POST request to "/items/abc/attempts"
    Then the response code should be 400
    And the response error message should contain "Wrong value for item_id (should be int64)"
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
    And the table "attempts" should stay unchanged

  Scenario: No access to the item (no item)
    Given I am the user with id "101"
    When I send a POST request to "/items/404/attempts"
    Then the response code should be 403
    And the response error message should contain "Insufficient access rights"
    And the table "attempts" should stay unchanged

  Scenario: No access to the item (info access)
    Given I am the user with id "101"
    When I send a POST request to "/items/50/attempts"
    Then the response code should be 403
    And the response error message should contain "Insufficient access rights"
    And the table "attempts" should stay unchanged

  Scenario: No access to the item (as a team)
    Given I am the user with id "101"
    When I send a POST request to "/items/50/attempts?as_team_id=104"
    Then the response code should be 403
    And the response error message should contain "Insufficient access rights"
    And the table "attempts" should stay unchanged

  Scenario: No access to the item (type='Chapter')
    Given I am the user with id "101"
    When I send a POST request to "/items/90/attempts"
    Then the response code should be 403
    And the response error message should contain "Insufficient access rights"
    And the table "attempts" should stay unchanged

  Scenario: User is not a team member
    Given I am the user with id "101"
    When I send a POST request to "/items/60/attempts?as_team_id=102"
    Then the response code should be 403
    And the response error message should contain "Can't use given as_team_id as a user's team"
    And the table "attempts" should stay unchanged

  Scenario: as_team_id is not a team
    Given I am the user with id "101"
    When I send a POST request to "/items/60/attempts?as_team_id=103"
    Then the response code should be 403
    And the response error message should contain "Can't use given as_team_id as a user's team"
    And the table "attempts" should stay unchanged

  Scenario: There is an attempt for the (group, item) pair already, but items.allows_multiple_attempts = 0
    Given I am the user with id "101"
    And the database table 'permissions_generated' has also the following row:
      | group_id | item_id | can_view_generated |
      | 104      | 50      | content            |
    And the database has the following table 'attempts':
      | id | participant_id |
      | 0  | 104            |
    And the database has the following table 'results':
      | attempt_id | participant_id | item_id |
      | 0          | 104            | 50      |
    When I send a POST request to "/items/50/attempts?as_team_id=104"
    Then the response code should be 422
    And the response error message should contain "The item doesn't allow multiple attempts"
    And the table "attempts" should stay unchanged
