Feature: Get group memberships history for the current user - robustness
  Background:
    Given the database has the following users:
      | login | temp_user | group_id | first_name  | last_name | grade |
      | owner | 0         | 21       | Jean-Michel | Blanquer  | 3     |

  Scenario: User doesn't exist
    Given I am the user with id "404"
    When I send a GET request to "/current-user/group-memberships-history"
    Then the response code should be 401
    And the response error message should contain "Invalid access token"

  Scenario: sort is incorrect
    Given I am the user with id "21"
    When I send a GET request to "/current-user/group-memberships-history?sort=myname"
    Then the response code should be 400
    And the response error message should contain "Unallowed field in sorting parameters: "myname""

