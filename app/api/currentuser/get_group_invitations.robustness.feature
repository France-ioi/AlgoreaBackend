Feature: Get group invitations for the current user - robustness
  Background:
    Given the database has the following users:
      | login | temp_user | group_id | first_name  | last_name | grade |
      | owner | 0         | 21       | Jean-Michel | Blanquer  | 3     |

  Scenario: User doesn't exist
    Given I am the user with id "404"
    When I send a GET request to "/current-user/group-invitations"
    Then the response code should be 401
    And the response error message should contain "Invalid access token"

  Scenario: within_weeks is incorrect
    Given I am the user with id "21"
    When I send a GET request to "/current-user/group-invitations?within_weeks=abc"
    Then the response code should be 400
    And the response error message should contain "Wrong value for within_weeks (should be int64)"

  Scenario: sort is incorrect
    Given I am the user with id "21"
    When I send a GET request to "/current-user/group-invitations?sort=myname"
    Then the response code should be 400
    And the response error message should contain "Unallowed field in sorting parameters: "myname""

