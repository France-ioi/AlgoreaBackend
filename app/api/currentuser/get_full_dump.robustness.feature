Feature: Export the current user's data - robustness
  Scenario: Unauthorized
    When I send a GET request to "/current-user/full-dump"
    Then the response code should be 401
    And the response error message should contain "No access token provided"
    And the response header "Content-Type" should be "application/json; charset=utf-8"
    And the response header "Content-Disposition" should be "[NULL]"

  Scenario: No such user
    Given I am the user with id "1"
    When I send a GET request to "/current-user/full-dump"
    Then the response code should be 401
    And the response error message should contain "Invalid access token"
    And the response header "Content-Type" should be "application/json; charset=utf-8"
    And the response header "Content-Disposition" should be "[NULL]"
