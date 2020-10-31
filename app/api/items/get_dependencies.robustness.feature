Feature: Get item dependencies - robustness
Background:
  Given the database has the following table 'groups':
    | id | name    | text_id | grade | type  |
    | 11 | jdoe    |         | -2    | User  |
    | 13 | Group B |         | -2    | Class |
    | 14 | Team    |         | -2    | Team  |
    | 15 | Team2   |         | -2    | Team  |
  And the database has the following table 'users':
    | login | temp_user | group_id |
    | jdoe  | 0         | 11       |
  And the database has the following table 'groups_groups':
    | parent_group_id | child_group_id |
    | 13              | 11             |
    | 15              | 11             |
  And the groups ancestors are computed
  And the database has the following table 'items':
    | id  | type    | no_score | default_language_tag |
    | 190 | Chapter | false    | fr                   |
    | 200 | Chapter | false    | fr                   |
    | 210 | Task    | false    | fr                   |
  And the database has the following table 'items_items':
    | parent_item_id | child_item_id | child_order |
    | 210            | 200           | 1           |
  And the database has the following table 'permissions_generated':
    | group_id | item_id | can_view_generated       |
    | 13       | 190     | none                     |
    | 13       | 200     | content_with_descendants |
    | 13       | 210     | content_with_descendants |
    | 15       | 200     | none                     |
    | 15       | 210     | content_with_descendants |
  And the database has the following table 'items_strings':
    | item_id | language_tag | title      |
    | 200     | en           | Category 1 |

  Scenario: Invalid item_id
    Given I am the user with id "11"
    When I send a GET request to "/items/abc/dependencies"
    Then the response code should be 400
    And the response error message should contain "Wrong value for item_id (should be int64)"

  Scenario: Should fail when as_team_id is invalid
    Given I am the user with id "11"
    When I send a GET request to "/items/200/dependencies?as_team_id=abc"
    Then the response code should be 400
    And the response error message should contain "Wrong value for as_team_id (should be int64)"

  Scenario: Should fail when as_team_id is not a team
    Given I am the user with id "11"
    When I send a GET request to "/items/200/dependencies?as_team_id=13"
    Then the response code should be 403
    And the response error message should contain "Can't use given as_team_id as a user's team"

  Scenario: Should fail when the current user is not a member of as_team_id
    Given I am the user with id "11"
    When I send a GET request to "/items/200/dependencies?as_team_id=14"
    Then the response code should be 403
    And the response error message should contain "Can't use given as_team_id as a user's team"

  Scenario: Should fail when the user doesn't exist
    Given I am the user with id "404"
    When I send a GET request to "/items/190/dependencies?attempt_id=0"
    Then the response code should be 401
    And the response error message should contain "Invalid access token"

  Scenario: Should fail when the user doesn't have info access to the root item
    Given I am the user with id "11"
    When I send a GET request to "/items/190/dependencies"
    Then the response code should be 403
    And the response error message should contain "Insufficient access rights"

  Scenario: Should fail when the root item doesn't exist
    Given I am the user with id "11"
    When I send a GET request to "/items/404/dependencies"
    Then the response code should be 403
    And the response error message should contain "Insufficient access rights"

  Scenario: Should fail when the team doesn't have content access to the root item
    Given I am the user with id "11"
    When I send a GET request to "/items/200/dependencies?as_team_id=15"
    Then the response code should be 403
    And the response error message should contain "Insufficient access rights"

  Scenario: Should fail when watched_group_id is invalid
    Given I am the user with id "11"
    When I send a GET request to "/items/210/dependencies?as_team_id=15&watched_group_id=abc"
    Then the response code should be 400
    And the response error message should contain "Wrong value for watched_group_id (should be int64)"

  Scenario: Should fail when the current user doesn't have 'can_watch_members' permission on watched_group_id
    Given I am the user with id "11"
    And the database has the following table 'group_managers':
      | group_id | manager_id | can_watch_members |
      | 13       | 11         | false             |
    When I send a GET request to "/items/210/dependencies?as_team_id=15&watched_group_id=13"
    Then the response code should be 403
    And the response error message should contain "No rights to watch for watched_group_id"
