Feature: Get user info the current user - robustness
  Scenario: Should fail if the user doesn't exist
    Given I am the user with ID "1"
    When I send a GET request to "/current-user"
    Then the response code should be 403
    And the response error message should contain "Insufficient access rights"
