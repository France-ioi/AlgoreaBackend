Feature: Feature: Get user's answer by user_answer_id
  Background:
    Given the database has the following table 'users':
      | id | login | self_group_id | owned_group_id |
      | 1  | jdoe  | 11            | 12             |
    And the database has the following table 'items':
      | id  | has_attempts |
      | 200 | 0            |
      | 210 | 1            |
    And the database has the following table 'users_answers':
      | id  | user_id | item_id | attempt_id | type       | state   | answer   | lang_prog | submitted_at        | score | validated | graded_at           | grader_user_id |
      | 101 | 1       | 200     | 150        | Submission | Current | print(1) | python    | 2017-05-29 06:38:38 | 100   | true      | 2018-05-29 06:38:38 | 123            |

  Scenario: Wrong answer_id
    Given I am the user with id "1"
    When I send a GET request to "/answers/abc"
    Then the response code should be 400
    And the response error message should contain "Wrong value for answer_id (should be int64)"

  Scenario: User doesn't exist
    Given I am the user with id "404"
    When I send a GET request to "/answers/101"
    Then the response code should be 401
    And the response error message should contain "Invalid access token"

  Scenario: User doesn't have sufficient access rights to the answer
    Given I am the user with id "1"
    When I send a GET request to "/answers/101"
    Then the response code should be 403
    And the response error message should contain "Insufficient access rights"

  Scenario: No answers
    Given I am the user with id "1"
    When I send a GET request to "/answers/100"
    Then the response code should be 403
    And the response error message should contain "Insufficient access rights"
