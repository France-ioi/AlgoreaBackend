Feature: Generate a login state - robustness
  Scenario: Authorization header is present
    Given the "Authorization" request header is "Bearer 1234567890"
    When I send a POST request to "/auth/login"
    Then the response code should be 400
    And the response error message should contain "'Authorization' header should not be present"
    And the table "login_states" should stay unchanged
