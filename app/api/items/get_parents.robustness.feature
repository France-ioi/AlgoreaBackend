Feature: Get item parents - robustness
Background:
  Given the database has the following table "groups":
    | id | name    | type  |
    | 13 | Group B | Class |
    | 14 | Team    | Team  |
    | 15 | Team2   | Team  |
  And the database has the following user:
    | group_id | login |
    | 11       | jdoe  |
  And the database has the following table "groups_groups":
    | parent_group_id | child_group_id |
    | 13              | 11             |
    | 15              | 11             |
  And the groups ancestors are computed
  And the database has the following table "items":
    | id  | type    | no_score | default_language_tag |
    | 190 | Chapter | false    | fr                   |
    | 200 | Chapter | false    | fr                   |
    | 210 | Task    | false    | fr                   |
  And the database has the following table "items_items":
    | parent_item_id | child_item_id | child_order |
    | 210            | 200           | 1           |
  And the database has the following table "permissions_generated":
    | group_id | item_id | can_view_generated       |
    | 13       | 190     | none                     |
    | 13       | 200     | content_with_descendants |
    | 13       | 210     | content_with_descendants |
    | 15       | 200     | none                     |
    | 15       | 210     | content_with_descendants |
  And the database has the following table "items_strings":
    | item_id | language_tag | title      |
    | 200     | en           | Category 1 |
  And the database has the following table "attempts":
    | id | participant_id | created_at          | root_item_id | parent_attempt_id | ended_at            |
    | 0  | 15             | 2019-01-30 08:26:41 | null         | null              | null                |
    | 1  | 15             | 2019-01-30 08:26:41 | 200          | 0                 | null                |
    | 2  | 15             | 2019-01-30 08:26:41 | 230          | 0                 | 2019-01-30 09:26:48 |
    | 3  | 15             | 2019-01-30 08:26:41 | 210          | 0                 | 2019-01-30 09:26:48 |
  And the database has the following table "results":
    | attempt_id | participant_id | item_id | score_computed | submissions | started_at          | validated_at        | latest_activity_at  |
    | 0          | 11             | 190     | 91             | 11          | 2019-01-30 09:26:42 | null                | 2019-01-30 09:36:41 |
    | 0          | 15             | 200     | 91             | 11          | 2019-01-30 09:26:42 | null                | 2019-01-30 09:36:41 |
    | 1          | 15             | 200     | 92             | 12          | 2019-01-30 09:26:42 | 2019-01-31 09:26:42 | 2019-01-30 09:36:42 |
    | 1          | 15             | 210     | 92             | 12          | null                | 2019-01-31 09:26:42 | 2019-01-30 09:36:42 |
    | 2          | 15             | 200     | 92             | 12          | null                | 2019-01-31 09:26:42 | 2019-01-30 09:36:42 |
    | 2          | 15             | 210     | 92             | 12          | 2019-01-30 09:26:42 | 2019-01-31 09:26:42 | 2019-01-30 09:36:42 |
    | 3          | 15             | 200     | 92             | 12          | 2019-01-30 09:26:42 | 2019-01-31 09:26:42 | 2019-01-30 09:36:42 |
    | 3          | 15             | 210     | 92             | 12          | 2019-01-30 09:26:42 | 2019-01-31 09:26:42 | 2019-01-30 09:36:42 |

  Scenario: Invalid item_id
    Given I am the user with id "11"
    When I send a GET request to "/items/abc/parents?attempt_id=0"
    Then the response code should be 400
    And the response error message should contain "Wrong value for item_id (should be int64)"

  Scenario: Should fail when attempt_id is invalid
    Given I am the user with id "11"
    When I send a GET request to "/items/200/parents?as_team_id=15&attempt_id=abc"
    Then the response code should be 400
    And the response error message should contain "Wrong value for attempt_id (should be int64)"

  Scenario: Should fail when as_team_id is invalid
    Given I am the user with id "11"
    When I send a GET request to "/items/200/parents?as_team_id=abc&attempt_id=0"
    Then the response code should be 400
    And the response error message should contain "Wrong value for as_team_id (should be int64)"

  Scenario: Should fail when as_team_id is not a team
    Given I am the user with id "11"
    When I send a GET request to "/items/200/parents?as_team_id=13&attempt_id=0"
    Then the response code should be 403
    And the response error message should contain "Can't use given as_team_id as a user's team"

  Scenario: Should fail when the current user is not a member of as_team_id
    Given I am the user with id "11"
    When I send a GET request to "/items/200/parents?as_team_id=14&attempt_id=0"
    Then the response code should be 403
    And the response error message should contain "Can't use given as_team_id as a user's team"

  Scenario: Should fail when the user doesn't exist
    Given I am the user with id "404"
    When I send a GET request to "/items/190/parents?attempt_id=0"
    Then the response code should be 401
    And the response error message should contain "Invalid access token"

  Scenario: Should fail when the user doesn't have info access to the root item
    Given I am the user with id "11"
    When I send a GET request to "/items/190/parents?attempt_id=0"
    Then the response code should be 403
    And the response error message should contain "Insufficient access rights"

  Scenario: Should fail when the root item doesn't exist
    Given I am the user with id "11"
    When I send a GET request to "/items/404/parents?attempt_id=0"
    Then the response code should be 403
    And the response error message should contain "Insufficient access rights"

  Scenario: Should fail when the team doesn't have content access to the root item
    Given I am the user with id "11"
    When I send a GET request to "/items/200/parents?as_team_id=15&attempt_id=0"
    Then the response code should be 403
    And the response error message should contain "Insufficient access rights"

  Scenario: Should fail when attempt_id is not given
    Given I am the user with id "11"
    When I send a GET request to "/items/200/parents?as_team_id=15"
    Then the response code should be 400
    And the response error message should contain "Missing attempt_id"

  Scenario: Should fail when watched_group_id is invalid
    Given I am the user with id "11"
    When I send a GET request to "/items/210/parents?as_team_id=15&attempt_id=2&watched_group_id=abc"
    Then the response code should be 400
    And the response error message should contain "Wrong value for watched_group_id (should be int64)"

  Scenario: Should fail when the current user doesn't have 'can_watch_members' permission on watched_group_id
    Given I am the user with id "11"
    And the database has the following table "group_managers":
      | group_id | manager_id | can_watch_members |
      | 13       | 11         | false             |
    When I send a GET request to "/items/210/parents?as_team_id=15&attempt_id=2&watched_group_id=13"
    Then the response code should be 403
    And the response error message should contain "No rights to watch for watched_group_id"
