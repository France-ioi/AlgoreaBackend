Feature: Get requests for group_id - robustness
  Background:
    Given the database has the following table 'users':
      | id | login | temp_user | self_group_id | owned_group_id | first_name  | last_name | grade |
      | 1  | owner | 0         | 21            | 22             | Jean-Michel | Blanquer  | 3     |
      | 2  | user  | 0         | 11            | 12             | John        | Doe       | 1     |
      | 3  | jane  | 0         | 31            | 32             | Jane        | Doe       | 2     |
    And the database has the following table 'groups_ancestors':
      | id | ancestor_group_id | child_group_id | is_self |
      | 75 | 22                | 13             | 0       |
      | 76 | 13                | 11             | 0       |
      | 77 | 22                | 11             | 0       |
      | 78 | 21                | 21             | 1       |

  Scenario: User is not an admin of the group
    Given I am the user with id "2"
    When I send a GET request to "/groups/13/requests"
    Then the response code should be 403
    And the response error message should contain "Insufficient access rights"

  Scenario: User doesn't exist
    Given I am the user with id "404"
    When I send a GET request to "/groups/13/requests"
    Then the response code should be 401
    And the response error message should contain "Invalid access token"

  Scenario: Group id is incorrect
    Given I am the user with id "1"
    When I send a GET request to "/groups/abc/requests"
    Then the response code should be 400
    And the response error message should contain "Wrong value for group_id (should be int64)"

  Scenario: rejections_within_weeks is incorrect
    Given I am the user with id "1"
    When I send a GET request to "/groups/13/requests?rejections_within_weeks=abc"
    Then the response code should be 400
    And the response error message should contain "Wrong value for rejections_within_weeks (should be int64)"

  Scenario: sort is incorrect
    Given I am the user with id "1"
    When I send a GET request to "/groups/13/requests?sort=myname"
    Then the response code should be 400
    And the response error message should contain "Unallowed field in sorting parameters: "myname""

