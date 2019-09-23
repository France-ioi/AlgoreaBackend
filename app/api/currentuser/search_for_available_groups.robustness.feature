Feature: Search for groups available to the current user - robustness
  Background:
    Given the database has the following table 'users':
      | id | login | temp_user | self_group_id | owned_group_id | first_name  | last_name | grade |
      | 1  | owner | 0         | 21            | 22             | Jean-Michel | Blanquer  | 3     |

  Scenario: Should fail if the search string is not present
    Given I am the user with id "1"
    When I send a GET request to "/current-user/available-groups"
    Then the response code should be 400
    And the response error message should contain "Missing search"

  Scenario: Should fail if the search string is too small (search for "  中国  ")
    Given I am the user with id "1"
    When I send a GET request to "/current-user/available-groups?search=%20%20%E4%B8%AD%E5%9B%BD%20%20"
    Then the response code should be 400
    And the response error message should contain "The search string should be at least 3 characters long"

  Scenario: Should fail if the user doesn't exist
    Given I am the user with id "2"
    When I send a GET request to "/current-user/available-groups?search=abcdef"
    Then the response code should be 401
    And the response error message should contain "Invalid access token"

  Scenario: sort is incorrect
    Given I am the user with id "1"
    When I send a GET request to "/current-user/available-groups?search=abcdef&sort=myname"
    Then the response code should be 400
    And the response error message should contain "Unallowed field in sorting parameters: "myname""

