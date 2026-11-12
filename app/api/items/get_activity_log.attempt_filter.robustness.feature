Feature: Activity log attempt filter - robustness

  Background:
    Given the database has the following users:
      | login | group_id |
      | user  | 11       |
    And the database has the following table "languages":
      | tag |
      | en  |
    And the database has the following table "items":
      | id  | type    | no_score | default_language_tag |
      | 500 | Task    | false    | en                   |
      | 600 | Chapter | false    | en                   |
      | 700 | Skill   | false    | en                   |
    And the database has the following table "permissions_generated":
      | group_id | item_id | can_view_generated |
      | 11       | 500     | content            |
      | 11       | 600     | content            |
      | 11       | 700     | content            |
    And the groups ancestors are computed

  Scenario: Invalid attempt_id value
    Given I am the user with id "11"
    When I send a GET request to "/items/500/log?attempt_id=abc"
    Then the response code should be 400
    And the response error message should contain "Wrong value for attempt_id (should be int64)"

  Scenario: Attempt filter rejected on a chapter
    Given I am the user with id "11"
    When I send a GET request to "/items/600/log?attempt_id=1"
    Then the response code should be 400
    And the response error message should contain "Attempt_id is only allowed when the item is a task"

  Scenario: Attempt filter rejected on a skill
    Given I am the user with id "11"
    When I send a GET request to "/items/700/log?attempt_id=1"
    Then the response code should be 400
    And the response error message should contain "Attempt_id is only allowed when the item is a task"
