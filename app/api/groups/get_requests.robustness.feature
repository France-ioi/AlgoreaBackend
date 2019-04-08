Feature: Get requests for group_id - robustness
  Background:
    Given the database has the following table 'users':
      | ID | sLogin | tempUser | idGroupSelf | idGroupOwned | sFirstName  | sLastName | iGrade |
      | 1  | owner  | 0        | 21          | 22           | Jean-Michel | Blanquer  | 3      |
      | 2  | user   | 0        | 11          | 12           | John        | Doe       | 1      |
      | 3  | jane   | 0        | 31          | 32           | Jane        | Doe       | 2      |
    And the database has the following table 'groups_ancestors':
      | ID | idGroupAncestor | idGroupChild | bIsSelf | iVersion |
      | 75 | 22              | 13           | 0       | 0        |
      | 76 | 13              | 11           | 0       | 0        |
      | 77 | 22              | 11           | 0       | 0        |
      | 78 | 21              | 21           | 1       | 0        |

  Scenario: User is not an admin of the group
    Given I am the user with ID "2"
    When I send a GET request to "/groups/13/requests"
    Then the response code should be 403
    And the response error message should contain "Insufficient access rights"

  Scenario: Group ID is incorrect
    Given I am the user with ID "1"
    When I send a GET request to "/groups/abc/requests"
    Then the response code should be 400
    And the response error message should contain "Wrong value for group_id (should be int64)"

  Scenario: old_rejections_weeks is incorrect
    Given I am the user with ID "1"
    When I send a GET request to "/groups/13/requests?old_rejections_weeks=abc"
    Then the response code should be 400
    And the response error message should contain "Wrong value for old_rejections_weeks (should be int64)"

  Scenario: sort is incorrect
    Given I am the user with ID "1"
    When I send a GET request to "/groups/13/requests?sort=myname"
    Then the response code should be 400
    And the response error message should contain "Unallowed field in sorting parameters: "myname""

