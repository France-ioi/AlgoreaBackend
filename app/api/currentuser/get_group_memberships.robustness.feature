Feature: Get group memberships for the current user - robustness
  Background:
    Given the database has the following table 'users':
      | ID | sLogin | tempUser | idGroupSelf | idGroupOwned | sFirstName  | sLastName | iGrade |
      | 1  | owner  | 0        | 21          | 22           | Jean-Michel | Blanquer  | 3      |

  Scenario: User doesn't exist
    Given I am the user with ID "4"
    When I send a GET request to "/current-user/group-memberships"
    Then the response code should be 401
    And the response error message should contain "Invalid access token"

  Scenario: sort is incorrect
    Given I am the user with ID "1"
    When I send a GET request to "/current-user/group-memberships?sort=myname"
    Then the response code should be 400
    And the response error message should contain "Unallowed field in sorting parameters: "myname""

