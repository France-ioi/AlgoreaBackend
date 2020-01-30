Feature: Get item for tree navigation - robustness
Background:
  Given the database has the following table 'groups':
    | id | name    | text_id | grade | type     |
    | 11 | jdoe    |         | -2    | UserSelf |
    | 13 | Group B |         | -2    | Class    |
    | 14 | Team    |         | -2    | Team     |
    | 15 | Team2   |         | -2    | Team     |
  And the database has the following table 'users':
    | login | temp_user | group_id |
    | jdoe  | 0         | 11       |
  And the database has the following table 'groups_groups':
    | parent_group_id | child_group_id |
    | 13              | 11             |
    | 15              | 11             |
  And the database has the following table 'groups_ancestors':
    | ancestor_group_id | child_group_id | is_self |
    | 11                | 11             | 1       |
    | 13                | 11             | 0       |
    | 13                | 13             | 1       |
    | 15                | 11             | 0       |
    | 15                | 15             | 1       |
  And the database has the following table 'items':
    | id  | type    | teams_editable | no_score | default_language_tag |
    | 190 | Chapter | false          | false    | fr                   |
    | 200 | Chapter | false          | false    | fr                   |
  And the database has the following table 'permissions_generated':
    | group_id | item_id | can_view_generated       |
    | 13       | 190     | none                     |
    | 13       | 200     | content_with_descendants |
  And the database has the following table 'items_strings':
    | item_id | language_tag | title      |
    | 200     | en           | Category 1 |

  Scenario: Should fail when the user doesn't have access to the root item
    Given I am the user with id "11"
    When I send a GET request to "/items/190/as-nav-tree"
    Then the response code should be 403
    And the response error message should contain "Insufficient access rights on given item id"

  Scenario: Should fail when the user doesn't exist
    Given I am the user with id "404"
    When I send a GET request to "/items/190/as-nav-tree"
    Then the response code should be 401
    And the response error message should contain "Invalid access token"

  Scenario: Should fail when the root item doesn't exist
    Given I am the user with id "11"
    When I send a GET request to "/items/404/as-nav-tree"
    Then the response code should be 403
    And the response error message should contain "Insufficient access rights on given item id"

  Scenario: Invalid item_id
    Given I am the user with id "11"
    When I send a GET request to "/items/abc/as-nav-tree"
    Then the response code should be 400
    And the response error message should contain "Wrong value for item_id (should be int64)"

  Scenario: Should fail when as_team_id is invalid
    Given I am the user with id "11"
    When I send a GET request to "/items/200/as-nav-tree?as_team_id=abc"
    Then the response code should be 400
    And the response error message should contain "Wrong value for as_team_id (should be int64)"

  Scenario: Should fail when as_team_id is not a team
    Given I am the user with id "11"
    When I send a GET request to "/items/200/as-nav-tree?as_team_id=13"
    Then the response code should be 403
    And the response error message should contain "Can't use given as_team_id as a user's team"

  Scenario: Should fail when the current user is not a member of as_team_id
    Given I am the user with id "11"
    When I send a GET request to "/items/200/as-nav-tree?as_team_id=14"
    Then the response code should be 403
    And the response error message should contain "Can't use given as_team_id as a user's team"

  Scenario: Should fail when the team doesn't have access to the root item
    Given I am the user with id "11"
    When I send a GET request to "/items/200/as-nav-tree?as_team_id=15"
    Then the response code should be 403
    And the response error message should contain "Insufficient access rights on given item id"
