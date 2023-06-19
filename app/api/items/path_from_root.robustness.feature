Feature: Find an item path - robustness
  Background:
    Given the database has the following table 'groups':
      | id  | type  | root_activity_id | root_skill_id |
      | 101 | User  | 70               | null          |
      | 102 | Team  | null             | null          |
      | 103 | Class | 50               | 90            |
      | 104 | Team  | 50               | 90            |
    And the database has the following table 'users':
      | login | group_id |
      | john  | 101      |
    And the database has the following table 'groups_groups':
      | parent_group_id | child_group_id |
      | 103             | 101            |
      | 104             | 101            |
    And the groups ancestors are computed
    And the database has the following table 'items':
      | id | url                                                                     | type  | allows_multiple_attempts | default_language_tag | requires_explicit_entry |
      | 50 | http://taskplatform.mblockelet.info/task.html?taskId=403449543672183936 | Task  | 0                        | fr                   | false                   |
      | 60 | http://taskplatform.mblockelet.info/task.html?taskId=403449543672183936 | Task  | 1                        | fr                   | false                   |
      | 70 | http://taskplatform.mblockelet.info/task.html?taskId=403449543672183936 | Task  | 1                        | fr                   | true                    |
      | 71 | http://taskplatform.mblockelet.info/task.html?taskId=403449543672183936 | Task  | 1                        | fr                   | false                   |
      | 90 | http://taskplatform.mblockelet.info/task.html?taskId=403449543672183936 | Skill | 1                        | fr                   | false                   |
    And the database has the following table 'items_items':
      | parent_item_id | child_item_id | child_order |
      | 70             | 71            | 1           |
    And the database has the following table 'permissions_generated':
      | group_id | item_id | can_view_generated |
      | 101      | 50      | none               |
      | 101      | 60      | content            |
      | 101      | 70      | content            |
      | 101      | 71      | content            |
      | 101      | 90      | content            |
    And the database has the following table 'attempts':
      | participant_id | id |
      | 101            | 0  |

  Scenario: Invalid item id
    Given I am the user with id "101"
    When I send a GET request to "/items/11111111111111111111111111111/path-from-root?attempt_id=0"
    Then the response code should be 400
    And the response error message should contain "Wrong value for item_id (should be int64)"

  Scenario: Invalid as_team_id
    Given I am the user with id "101"
    When I send a GET request to "/items/50/path-from-root?as_team_id=abc"
    Then the response code should be 400
    And the response error message should contain "Wrong value for as_team_id (should be int64)"

  Scenario: User not found
    Given I am the user with id "404"
    When I send a GET request to "/items/50/path-from-root"
    Then the response code should be 401
    And the response error message should contain "Invalid access token"

  Scenario: No access to the item (no item)
    Given I am the user with id "101"
    When I send a GET request to "/items/404/path-from-root"
    Then the response code should be 403
    And the response error message should contain "Insufficient access rights"

  Scenario: No access to the item
    Given I am the user with id "101"
    When I send a GET request to "/items/50/path-from-root"
    Then the response code should be 403
    And the response error message should contain "Insufficient access rights"

  Scenario: No access to the item (as a team)
    Given I am the user with id "101"
    When I send a GET request to "/items/50/path-from-root?as_team_id=104"
    Then the response code should be 403
    And the response error message should contain "Insufficient access rights"

  Scenario: User is not a team member
    Given I am the user with id "101"
    When I send a GET request to "/items/60/path-from-root?as_team_id=102"
    Then the response code should be 403
    And the response error message should contain "Can't use given as_team_id as a user's team"

  Scenario: as_team_id is not a team
    Given I am the user with id "101"
    When I send a GET request to "/items/60/path-from-root?as_team_id=103"
    Then the response code should be 403
    And the response error message should contain "Can't use given as_team_id as a user's team"

  Scenario: No path
    Given I am the user with id "101"
    When I send a GET request to "/items/71/path-from-root"
    Then the response code should be 403
    And the response error message should contain "Insufficient access rights"
