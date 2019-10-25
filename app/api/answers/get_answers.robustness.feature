Feature: Get item answers - robustness
Background:
  Given the database has the following table 'groups':
    | id | name       | type      |
    | 1  | jdoe       | UserSelf  |
    | 2  | jdoe-admin | UserAdmin |
  And the database has the following table 'users':
    | login | temp_user | group_id | owned_group_id |
    | jdoe  | 0         | 1        | 2              |

  Scenario: Should fail when neither user_group_id & item_id nor attempt_id is present
    Given I am the user with group_id "1"
    When I send a GET request to "/answers"
    Then the response code should be 400
    And the response error message should contain "Either user_group_id & item_id or attempt_id must be present"

  Scenario: Should fail when only user_group_id is present
    Given I am the user with group_id "1"
    When I send a GET request to "/answers?user_group_id=1"
    Then the response code should be 400
    And the response error message should contain "Either user_group_id & item_id or attempt_id must be present"

  Scenario: Should fail when only item_id is present
    Given I am the user with group_id "1"
    When I send a GET request to "/answers?item_id=1"
    Then the response code should be 400
    And the response error message should contain "Either user_group_id & item_id or attempt_id must be present"
