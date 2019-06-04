Feature: Feature: Get user's answer by user_answer_id
  Given the database has the following table 'users':
    | ID | sLogin | idGroupSelf | idGroupOwned |
    | 1  | jdoe   | 11          | 12           |
  Scenario: Wrong answer_id
    Given I am the user with ID "1"
    When I send a GET request to "/answers/abc"
    Then the response code should be 400
    And the response error message should contain "Wrong value for answer_id (should be int64)"

  Scenario: User doesnt have sufficient access rights to the answer
    Given I am the user with ID "404"
    When I send a GET request to "/answers/1"
    Then the response code should be 403
    And the response error message should contain "Insufficient access rights"
