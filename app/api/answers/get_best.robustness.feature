Feature: Get the best answer - robustness
  Background:
    Given the database has the following table 'groups':
      | id | name    | type |
      | 11 | jdoe    | User |
      | 12 | manager | User |
      | 13 | team    | Team |
    And the database has the following table 'users':
      | login   | group_id |
      | jdoe    | 11       |
      | manager | 12       |
    And the database has the following table 'items':
      | id  | entry_participant_type | default_language_tag |
      | 200 | User                   | fr                   |
      | 210 | Team                   | fr                   |
    And the database has the following table 'groups_groups':
      | parent_group_id | child_group_id |
      | 13              | 11             |
      | 13              | 12             |
    And the groups ancestors are computed
    And the database has the following table 'group_managers':
      | group_id | manager_id | can_watch_members |
      | 13       | 11         | false             |
      | 13       | 12         | true              |
    And the database has the following table 'permissions_generated':
      | item_id | group_id | can_view_generated | can_watch_generated |
      | 200     | 11       | info               | none                |
      | 210     | 11       | content            | answer              |
      | 210     | 12       | content            | result              |
      | 210     | 13       | content            | answer              |
    And the database has the following table 'attempts':
      | id | participant_id |
      | 1  | 11             |
      | 2  | 13             |
    And the database has the following table 'results':
      | attempt_id | participant_id | item_id |
      | 1          | 11             | 200     |
      | 2          | 13             | 210     |
    And the database has the following table 'answers':
      | id  | author_id | participant_id | attempt_id | item_id | type          | state    | answer   | created_at          |
      | 101 | 11        | 11             | 1          | 200     | Submission    | State101 | print(1) | 2020-01-01 06:00:00 |
      | 102 | 13        | 13             | 2          | 210     | Submission    | State102 | print(3) | 2020-01-01 06:00:00 |
    And the database has the following table 'gradings':
      | answer_id | score | graded_at           |
      | 101       | 100   | 2020-01-01 06:00:00 |
      | 102       | 100   | 2020-01-01 06:00:00 |

  Scenario: Invalid item_id
    Given I am the user with id "11"
    When I send a GET request to "/items/1111111111111111111111111111/best-answer"
    Then the response code should be 400
    And the response error message should contain "Wrong value for item_id (should be int64)"

  Scenario: Non-existent item_id
    Given I am the user with id "11"
    When I send a GET request to "/items/404/best-answer"
    Then the response code should be 403
    And the response error message should contain "Insufficient access rights"

  Scenario: Invalid watched_group_id
    Given I am the user with id "11"
    When I send a GET request to "/items/200/best-answer?watched_group_id=abc"
    Then the response code should be 400
    And the response error message should contain "Wrong value for watched_group_id (should be int64)"

  Scenario: Non-existent watched_group_id
    Given I am the user with id "11"
    When I send a GET request to "/items/200/best-answer?watched_group_id=404"
    Then the response code should be 403
    And the response error message should contain "No rights to watch for watched_group_id"

  Scenario: User doesn't have sufficient access rights to the item
    Given I am the user with id "11"
    When I send a GET request to "/items/200/best-answer"
    Then the response code should be 403
    And the response error message should contain "Insufficient access rights"

  Scenario: No answer for the item
    Given I am the user with id "11"
    When I send a GET request to "/items/210/best-answer"
    Then the response code should be 403
    And the response error message should contain "Insufficient access rights"

  Scenario: The user is not allowed to watch the participant
    Given I am the user with id "11"
    When I send a GET request to "/items/210/best-answer?watched_group_id=13"
    Then the response code should be 403
    And the response error message should contain "No rights to watch for watched_group_id"

  Scenario: The user is not allowed to watch "answer" of the item
    Given I am the user with id "12"
    When I send a GET request to "/items/210/best-answer?watched_group_id=13"
    Then the response code should be 403
    And the response error message should contain "Insufficient access rights"
