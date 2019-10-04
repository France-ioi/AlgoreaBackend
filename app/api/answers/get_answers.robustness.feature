Feature: Get item answers - robustness
Background:
  Given the database has the following table 'users':
    | id | login | temp_user | self_group_id | owned_group_id |
    | 1  | jdoe  | 0         | 11            | 12             |

  Scenario: Should fail when neither user_id & item_id nor attempt_id is present
    Given I am the user with id "1"
    When I send a GET request to "/answers"
    Then the response code should be 400
    And the response error message should contain "Either user_id & item_id or attempt_id must be present"

  Scenario: Should fail when only user_id is present
    Given I am the user with id "1"
    When I send a GET request to "/answers?user_id=1"
    Then the response code should be 400
    And the response error message should contain "Either user_id & item_id or attempt_id must be present"

  Scenario: Should fail when only item_id is present
    Given I am the user with id "1"
    When I send a GET request to "/answers?item_id=1"
    Then the response code should be 400
    And the response error message should contain "Either user_id & item_id or attempt_id must be present"
