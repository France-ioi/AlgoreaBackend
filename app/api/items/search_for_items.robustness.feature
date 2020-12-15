Feature: Search for items - robustness
  Background:
    Given the database has the following users:
      | login | temp_user | group_id | first_name  | last_name | grade |
      | owner | 0         | 21       | Jean-Michel | Blanquer  | 3     |

  Scenario: Should fail if the search string is not present
    Given I am the user with id "21"
    When I send a GET request to "/items/search"
    Then the response code should be 400
    And the response error message should contain "Missing search"

  Scenario: Should fail if the search string is too small (search for "  中国  ")
    Given I am the user with id "21"
    When I send a GET request to "/items/search?search=%20%20%E4%B8%AD%E5%9B%BD%20%20"
    Then the response code should be 400
    And the response error message should contain "The search string should be at least 3 characters long"

  Scenario: Should fail if the user doesn't exist
    Given I am the user with id "404"
    When I send a GET request to "/items/search?search=abcdef"
    Then the response code should be 401
    And the response error message should contain "Invalid access token"

  Scenario: Invalid type in types_include
    Given I am the user with id "21"
    When I send a GET request to "/items/search?search=abc&types_include=Book"
    Then the response code should be 400
    And the response error message should contain "Wrong value in 'types_include': "Book""

  Scenario: Invalid type in types_exclude
    Given I am the user with id "21"
    When I send a GET request to "/items/search?search=abc&types_exclude=Drawing"
    Then the response code should be 400
    And the response error message should contain "Wrong value in 'types_exclude': "Drawing""
