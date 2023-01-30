Feature: Get activity log - robustness
  Background:
    Given the database has the following users:
      | login   | temp_user | group_id | first_name  | last_name |
      | someone | 0         | 21       | Bill        | Clinton   |
      | user    | 0         | 11       | John        | Doe       |
      | owner   | 0         | 23       | Jean-Michel | Blanquer  |
    And the database has the following table 'groups':
      | id | type  |
      | 13 | Class |
      | 30 | Team  |
    And the database has the following table 'groups_groups':
      | parent_group_id | child_group_id |
      | 30              | 23             |
    And the database has the following table 'group_managers':
      | group_id | manager_id | can_watch_members |
      | 13       | 21         | false             |
      | 13       | 23         | true              |
    And the groups ancestors are computed
    And the database has the following table 'attempts':
      | id | participant_id |
      | 0  | 11             |
      | 1  | 11             |
    And the database has the following table 'results':
      | attempt_id | item_id | participant_id |
      | 0          | 200     | 11             |
      | 1          | 200     | 11             |
    And the database has the following table 'answers':
      | id | author_id | participant_id | attempt_id | item_id | type       | state  | created_at          |
      | 1  | 11        | 11             | 0          | 200     | Submission | State1 | 2017-05-29 06:38:38 |
      | 2  | 11        | 11             | 1          | 200     | Submission | State2 | 2017-05-29 06:38:38 |
    And the database has the following table 'gradings':
      | answer_id | graded_at           | score |
      | 1         | 2017-05-29 06:38:38 | 100   |
      | 2         | 2017-05-29 06:38:38 | 100   |
    And the database has the following table 'items':
      | id  | type    | no_score | default_language_tag |
      | 200 | Chapter | false    | fr                   |
    And the database has the following table 'permissions_generated':
      | group_id | item_id | can_view_generated       |
      | 21       | 200     | content_with_descendants |
      | 23       | 200     | none                     |
    And the database has the following table 'items_ancestors':
      | ancestor_item_id | child_item_id |
      | 200              | 200           |

  Scenario: Wrong as_team_id
    Given I am the user with id "23"
    When I send a GET request to "/items/200/log?as_team_id=abc"
    Then the response code should be 400
    And the response error message should contain "Wrong value for as_team_id (should be int64)"

  Scenario: Wrong watched_group_id
    Given I am the user with id "23"
    When I send a GET request to "/items/200/log?watched_group_id=abc"
    Then the response code should be 400
    And the response error message should contain "Wrong value for watched_group_id (should be int64)"

  Scenario: Both as_team_id and watched_group_id are given
    Given I am the user with id "23"
    When I send a GET request to "/items/200/log?watched_group_id=13&as_team_id=30"
    Then the response code should be 400
    And the response error message should contain "Only one of as_team_id and watched_group_id can be given"

  Scenario: Wrong ancestor item
    Given I am the user with id "23"
    When I send a GET request to "/items/abc/log?watched_group_id=13"
    Then the response code should be 400
    And the response error message should contain "Wrong value for ancestor_item_id (should be int64)"

  Scenario: Should fail when user cannot watch group members of watched_group_id
    Given I am the user with id "21"
    When I send a GET request to "/items/200/log?watched_group_id=13"
    Then the response code should be 403
    And the response error message should contain "No rights to watch for watched_group_id"

  Scenario: Should fail when user doesn't exist
    Given I am the user with id "404"
    When I send a GET request to "/items/200/log"
    Then the response code should be 401
    And the response error message should contain "Invalid access token"

  Scenario: Should return empty array when user is an admin of the group, but has no access rights to the item
    Given I am the user with id "23"
    When I send a GET request to "/items/200/log?watched_group_id=13"
    Then the response code should be 200
    And the response body should be, in JSON:
    """
    []
    """

  Scenario: Should fail when some of from.* parameters are missing
    Given I am the user with id "23"
    When I send a GET request to "/items/200/log?watched_group_id=13&from.answer_id=1"
    Then the response code should be 400
    And the response error message should contain "All 'from' parameters (from.activity_type, from.answer_id, from.attempt_id, from.item_id, from.participant_id) or none of them must be present"

  Scenario: Should fail when from.activity_type is invalid
    Given I am the user with id "23"
    When I send a GET request to "/items/200/log?watched_group_id=13&from.activity_type=grading"
    Then the response code should be 400
    And the response error message should contain "Wrong value for from.activity_type (should be one of (result_started, submission, result_validated, saved_answer, current_answer))"

  Scenario: Wrong as_team_id (without item_id)
    Given I am the user with id "23"
    When I send a GET request to "/items/log?as_team_id=abc"
    Then the response code should be 400
    And the response error message should contain "Wrong value for as_team_id (should be int64)"

  Scenario: Wrong watched_group_id (without item_id)
    Given I am the user with id "23"
    When I send a GET request to "/items/log?watched_group_id=abc"
    Then the response code should be 400
    And the response error message should contain "Wrong value for watched_group_id (should be int64)"

  Scenario: Both as_team_id and watched_group_id are given (without item_id)
    Given I am the user with id "23"
    When I send a GET request to "/items/200/log?watched_group_id=13&as_team_id=30"
    Then the response code should be 400
    And the response error message should contain "Only one of as_team_id and watched_group_id can be given"

  Scenario: Should fail when user cannot watch group members of watched_group_id (without item_id)
    Given I am the user with id "21"
    When I send a GET request to "/items/log?watched_group_id=13"
    Then the response code should be 403
    And the response error message should contain "No rights to watch for watched_group_id"

  Scenario: Should fail when user doesn't exist (without item_id)
    Given I am the user with id "404"
    When I send a GET request to "/items/log"
    Then the response code should be 401
    And the response error message should contain "Invalid access token"

  Scenario: Should return empty array when user is an admin of the group, but has no visible items (without item_id)
    Given I am the user with id "23"
    When I send a GET request to "/items/log?watched_group_id=13"
    Then the response code should be 200
    And the response body should be, in JSON:
    """
    []
    """

  Scenario: Should fail when some of from.* parameters are missing (without item_id)
    Given I am the user with id "23"
    When I send a GET request to "/items/log?watched_group_id=13&from.answer_id=1"
    Then the response code should be 400
    And the response error message should contain "All 'from' parameters (from.activity_type, from.answer_id, from.attempt_id, from.item_id, from.participant_id) or none of them must be present"

  Scenario: Should fail when from.activity_type is invalid (without item_id)
    Given I am the user with id "23"
    When I send a GET request to "/items/log?watched_group_id=13&from.activity_type=grading"
    Then the response code should be 400
    And the response error message should contain "Wrong value for from.activity_type (should be one of (result_started, submission, result_validated, saved_answer, current_answer))"
