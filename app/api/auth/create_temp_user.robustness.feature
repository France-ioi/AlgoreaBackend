Feature: Create a temporary user - robustness
  Scenario: Authorization header is present
    Given the "Authorization" request header is "Bearer 1234567890"
    When I send a POST request to "/auth/temp-user"
    Then the response code should be 400
    And the response error message should contain "'Authorization' header should not be present"
    And the table "users" should stay unchanged
    And the table "groups" should stay unchanged
    And the table "groups_groups" should stay unchanged
    And the table "groups_ancestors" should stay unchanged
    And the table "sessions" should stay unchanged
