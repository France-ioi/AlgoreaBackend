Feature: Publish a result to LTI - robustness
  Background:
    Given the database has the following table "groups":
      | id  | type  |
      | 21  | User  |
      | 31  | User  |
      | 99  | Class |
      | 100 | Team  |
    And the database has the following table "groups_groups":
      | parent_group_id | child_group_id |
      | 99              | 21             |
      | 100             | 21             |
      | 100             | 31             |
    And the groups ancestors are computed
    And the database has the following table "items":
      | id  | default_language_tag |
      | 123 | fr                   |
      | 124 | fr                   |
    And the database has the following table "permissions_generated":
      | group_id | item_id | can_view_generated |
      | 21       | 123     | content            |
      | 99       | 124     | info               |
    And the database has the following table "attempts":
      | id  | participant_id |
      | 0   | 21             |
      | 1   | 21             |
    And the database has the following table "results":
      | attempt_id | participant_id | item_id | score_computed |
      | 0          | 21             | 123     | 12.3           |
      | 1          | 21             | 123     | 15.6           |
      | 1          | 21             | 124     | 20.1           |
      | 1          | 31             | 123     | 9.5            |
    And the database has the following table "users":
      | temp_user | login | group_id | login_id |
      | 0         | john  | 21       | 1234567  |
      | 1         | jane  | 31       | null     |

  Scenario: Invalid item_id
    Given I am the user with id "21"
    When I send a POST request to "/items/abc/attempts/1/publish"
    Then the response code should be 400
    And the response error message should contain "Wrong value for item_id (should be int64)"

  Scenario: The current user has no login_id
    Given I am the user with id "31"
    When I send a POST request to "/items/123/attempts/1/publish"
    Then the response code should be 403
    And the response error message should contain "Insufficient access rights"

  Scenario: The current user has no content permission on the item
    Given I am the user with id "21"
    When I send a POST request to "/items/124/attempts/1/publish"
    Then the response code should be 403
    And the response error message should contain "Insufficient access rights"

  Scenario: Invalid attempt_id
    Given I am the user with id "21"
    When I send a POST request to "/items/123/attempts/abc/publish"
    Then the response code should be 400
    And the response error message should contain "Wrong value for attempt_id (should be int64)"

  Scenario: Invalid as_team_id
    Given I am the user with id "21"
    When I send a POST request to "/items/123/attempts/1/publish?as_team_id=abc"
    Then the response code should be 400
    And the response error message should contain "Wrong value for as_team_id (should be int64)"

  Scenario: as_team_id is not a user's team
    Given I am the user with id "21"
    When I send a POST request to "/items/123/attempts/1/publish?as_team_id=99"
    Then the response code should be 403
    And the response error message should contain "Can't use given as_team_id as a user's team"

  Scenario: as_team_id is a user's team, but team work is not supported
    Given I am the user with id "21"
    When I send a POST request to "/items/123/attempts/1/publish?as_team_id=100"
    Then the response code should be 400
    And the response error message should contain "The service doesn't support 'as_team_id'"
