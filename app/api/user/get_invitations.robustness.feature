Feature: Get group invitations for the current user - robustness
  Background:
    Given the database has the following table 'users':
      | ID | sLogin | tempUser | idGroupSelf | idGroupOwned | sFirstName  | sLastName | iGrade |
      | 1  | owner  | 0        | 21          | 22           | Jean-Michel | Blanquer  | 3      |

  Scenario: User doesn't exist
    Given I am the user with ID "4"
    When I send a GET request to "/user/invitations"
    Then the response code should be 403
    And the response error message should contain "Insufficient access rights"

  Scenario: within_weeks is incorrect
    Given I am the user with ID "1"
    When I send a GET request to "/user/invitations?within_weeks=abc"
    Then the response code should be 400
    And the response error message should contain "Wrong value for within_weeks (should be int64)"

  Scenario: sort is incorrect
    Given I am the user with ID "1"
    When I send a GET request to "/user/invitations?sort=myname"
    Then the response code should be 400
    And the response error message should contain "Unallowed field in sorting parameters: "myname""

