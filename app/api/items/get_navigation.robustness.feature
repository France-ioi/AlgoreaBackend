Feature: Get item navigation - robustness
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
    | 200            | 210           | 1           |
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
    | 0          | 15             | 200     | 91             | 11          | null                | null                | 2019-01-30 09:36:41 |
    | 1          | 15             | 200     | 92             | 12          | 2019-01-30 09:26:42 | 2019-01-31 09:26:42 | 2019-01-30 09:36:42 |
    | 1          | 15             | 210     | 92             | 12          | null                | 2019-01-31 09:26:42 | 2019-01-30 09:36:42 |
    | 2          | 15             | 200     | 92             | 12          | null                | 2019-01-31 09:26:42 | 2019-01-30 09:36:42 |
    | 2          | 15             | 210     | 92             | 12          | 2019-01-30 09:26:42 | 2019-01-31 09:26:42 | 2019-01-30 09:36:42 |
    | 3          | 15             | 200     | 92             | 12          | 2019-01-30 09:26:42 | 2019-01-31 09:26:42 | 2019-01-30 09:36:42 |
    | 3          | 15             | 210     | 92             | 12          | 2019-01-30 09:26:42 | 2019-01-31 09:26:42 | 2019-01-30 09:36:42 |

  Scenario: Should fail when the user doesn't have access to the root item
    Given I am the user with id "11"
    When I send a GET request to "/items/190/navigation?attempt_id=0"
    Then the response code should be 403
    And the response error message should contain "Insufficient access rights on given item id"

  Scenario: Should fail when the user doesn't exist
    Given I am the user with id "404"
    When I send a GET request to "/items/190/navigation?attempt_id=0"
    Then the response code should be 401
    And the response error message should contain "Invalid access token"

  Scenario: Should fail when the root item doesn't exist
    Given I am the user with id "11"
    When I send a GET request to "/items/404/navigation?attempt_id=0"
    Then the response code should be 403
    And the response error message should contain "Insufficient access rights on given item id"

  Scenario: Invalid item_id
    Given I am the user with id "11"
    When I send a GET request to "/items/abc/navigation?attempt_id=0"
    Then the response code should be 400
    And the response error message should contain "Wrong value for item_id (should be int64)"

  Scenario: Should fail when as_team_id is invalid
    Given I am the user with id "11"
    When I send a GET request to "/items/200/navigation?as_team_id=abc&attempt_id=0"
    Then the response code should be 400
    And the response error message should contain "Wrong value for as_team_id (should be int64)"

  Scenario: Should fail when as_team_id is not a team
    Given I am the user with id "11"
    When I send a GET request to "/items/200/navigation?as_team_id=13&attempt_id=0"
    Then the response code should be 403
    And the response error message should contain "Can't use given as_team_id as a user's team"

  Scenario: Should fail when the current user is not a member of as_team_id
    Given I am the user with id "11"
    When I send a GET request to "/items/200/navigation?as_team_id=14&attempt_id=0"
    Then the response code should be 403
    And the response error message should contain "Can't use given as_team_id as a user's team"

  Scenario: Should fail when the team doesn't have access to the root item
    Given I am the user with id "11"
    When I send a GET request to "/items/200/navigation?as_team_id=15&attempt_id=1"
    Then the response code should be 403
    And the response error message should contain "Insufficient access rights on given item id"

  Scenario: Should fail when both attempt_id and child_attempt_id are given
    Given I am the user with id "11"
    When I send a GET request to "/items/200/navigation?as_team_id=15&attempt_id=0&child_attempt_id=0"
    Then the response code should be 400
    And the response error message should contain "Only one of attempt_id and child_attempt_id can be given"

  Scenario: Should fail when attempt_id is invalid
    Given I am the user with id "11"
    When I send a GET request to "/items/200/navigation?as_team_id=15&attempt_id=abc"
    Then the response code should be 400
    And the response error message should contain "Wrong value for attempt_id (should be int64)"

  Scenario: Should fail when child_attempt_id is invalid
    Given I am the user with id "11"
    When I send a GET request to "/items/200/navigation?as_team_id=15&child_attempt_id=abc"
    Then the response code should be 400
    And the response error message should contain "Wrong value for child_attempt_id (should be int64)"

  Scenario: Should fail when neither attempt_id nor child_attempt_id is given
    Given I am the user with id "11"
    When I send a GET request to "/items/200/navigation?as_team_id=15"
    Then the response code should be 400
    And the response error message should contain "One of attempt_id and child_attempt_id should be given"

  Scenario: Should fail when there is no started result for a child item and child_attempt_id
    Given I am the user with id "11"
    When I send a GET request to "/items/210/navigation?as_team_id=15&child_attempt_id=1"
    Then the response code should be 403
    And the response error message should contain "Insufficient access rights"

  Scenario: Should fail when there is a started result for a child item, but there is no related started result for the parent item
    Given I am the user with id "11"
    When I send a GET request to "/items/210/navigation?as_team_id=15&child_attempt_id=2"
    Then the response code should be 403
    And the response error message should contain "Insufficient access rights"

  Scenario: Should fail when there is a started result for a child item, but there is no related started result for the parent item (because of root_item_id)
    Given I am the user with id "11"
    When I send a GET request to "/items/210/navigation?as_team_id=15&child_attempt_id=3"
    Then the response code should be 403
    And the response error message should contain "Insufficient access rights"

  Scenario: Should fail when there is no started result for the item and attempt_id
    Given I am the user with id "11"
    When I send a GET request to "/items/210/navigation?as_team_id=15&attempt_id=0"
    Then the response code should be 403
    And the response error message should contain "Insufficient access rights"

  Scenario: Should fail when watched_group_id is invalid
    Given I am the user with id "11"
    When I send a GET request to "/items/210/navigation?as_team_id=15&attempt_id=1&watched_group_id=abc"
    Then the response code should be 400
    And the response error message should contain "Wrong value for watched_group_id (should be int64)"

  Scenario: Should fail when the current user doesn't have 'can_watch_members' permission on watched_group_id
    Given I am the user with id "11"
    And the database has the following table "group_managers":
      | group_id | manager_id | can_watch_members |
      | 13       | 11         | false             |
    When I send a GET request to "/items/210/navigation?as_team_id=15&attempt_id=0&watched_group_id=13"
    Then the response code should be 403
    And the response error message should contain "No rights to watch for watched_group_id"
