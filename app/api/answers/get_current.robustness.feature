Feature: Get a current answer - robustness
  Background:
    Given the database has the following table 'groups':
      | id | name | type |
      | 11 | jdoe | User |
      | 13 | team | Team |
    And the database has the following table 'users':
      | login | group_id |
      | jdoe  | 11       |
    And the database has the following table 'items':
      | id  | entry_participant_type | default_language_tag |
      | 200 | User                   | fr                   |
      | 210 | Team                   | fr                   |
    And the database has the following table 'permissions_generated':
      | item_id | group_id | can_view_generated |
      | 200     | 11       | info               |
      | 210     | 11       | content            |
    And the database has the following table 'attempts':
      | id | participant_id |
      | 1  | 11             |
      | 1  | 13             |
    And the database has the following table 'results':
      | attempt_id | participant_id | item_id |
      | 1          | 11             | 200     |
      | 1          | 13             | 210     |
    And the database has the following table 'answers':
      | id  | author_id | participant_id | attempt_id | item_id | type       | state    | answer   | created_at          |
      | 101 | 11        | 11             | 1          | 200     | Current    | State101 | print(1) | 2017-05-29 06:38:38 |
      | 102 | 11        | 11             | 1          | 200     | Current    | State102 | print(1) | 2017-05-29 06:38:38 |
      | 103 | 11        | 11             | 1          | 210     | Submission | State103 | print(1) | 2017-05-29 06:38:38 |
      | 104 | 11        | 11             | 2          | 210     | Current    | State104 | print(1) | 2017-05-29 06:38:38 |
    And the database has the following table 'gradings':
      | answer_id | score | graded_at           |
      | 101       | 100   | 2018-05-29 06:38:38 |
      | 102       | 100   | 2019-05-29 06:38:38 |

  Scenario: Invalid item_id
    Given I am the user with id "11"
    When I send a GET request to "/items/1111111111111111111111111111/current-answer?attempt_id=1"
    Then the response code should be 400
    And the response error message should contain "Wrong value for item_id (should be int64)"

  Scenario: Invalid attempt_id
    Given I am the user with id "11"
    When I send a GET request to "/items/200/current-answer?attempt_id=abc"
    Then the response code should be 400
    And the response error message should contain "Wrong value for attempt_id (should be int64)"

  Scenario: Invalid as_team_id
    Given I am the user with id "11"
    When I send a GET request to "/items/200/current-answer?as_team_id=abc&attempt_id=1"
    Then the response code should be 400
    And the response error message should contain "Wrong value for as_team_id (should be int64)"

  Scenario: User doesn't have sufficient access rights to the item
    Given I am the user with id "11"
    When I send a GET request to "/items/200/current-answer?attempt_id=1"
    Then the response code should be 403
    And the response error message should contain "Insufficient access rights"

  Scenario: No current answer for the item and the attempt
    Given I am the user with id "11"
    When I send a GET request to "/items/210/current-answer?attempt_id=1"
    Then the response code should be 403
    And the response error message should contain "Insufficient access rights"

  Scenario: User is not a member of the team
    Given I am the user with id "11"
    When I send a GET request to "/items/210/current-answer?as_team_id=13&attempt_id=1"
    Then the response code should be 403
    And the response error message should contain "Can't use given as_team_id as a user's team"
