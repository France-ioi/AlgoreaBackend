Feature: Get time-limited items that the user has administration rights on (itemTimeLimitedAdminList) - robustness
  Background:
    Given the database has the following users:
      | login      | group_id | default_language |
      | possesseur | 21       | fr               |
    And the groups ancestors are computed

  Scenario: Wrong sort
    Given I am the user with id "21"
    When I send a GET request to "/items/time-limited/administered?sort=name"
    Then the response code should be 400
    And the response error message should contain "Unallowed field in sorting parameters: "name""
