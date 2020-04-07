Feature: Get requests for group_id - robustness
  Background:
    Given the database has the following users:
      | login | temp_user | group_id | first_name  | last_name | grade |
      | owner | 0         | 21       | Jean-Michel | Blanquer  | 3     |
      | user  | 0         | 11       | John        | Doe       | 1     |
      | jane  | 0         | 31       | Jane        | Doe       | 2     |
    And the database has the following table 'groups':
      | id |
      | 13 |
    And the database has the following table 'group_managers':
      | group_id | manager_id | can_manage  |
      | 13       | 21         | memberships |
      | 13       | 31         | none        |
      | 21       | 31         | memberships |
    And the database has the following table 'groups_ancestors':
      | ancestor_group_id | child_group_id |
      | 11                | 11             |
      | 13                | 11             |
      | 13                | 13             |
      | 21                | 21             |
      | 31                | 31             |

  Scenario: User is not a manager of the group
    Given I am the user with id "11"
    When I send a GET request to "/groups/13/requests"
    Then the response code should be 403
    And the response error message should contain "Insufficient access rights"

  Scenario: User is a manager of the group, but doesn't have enough permissions on it
    Given I am the user with id "31"
    When I send a GET request to "/groups/13/requests"
    Then the response code should be 403
    And the response error message should contain "Insufficient access rights"

  Scenario: User has enough permissions on the group, but the group is a user
    Given I am the user with id "31"
    When I send a GET request to "/groups/21/requests"
    Then the response code should be 403
    And the response error message should contain "Insufficient access rights"

  Scenario: User doesn't exist
    Given I am the user with id "404"
    When I send a GET request to "/groups/13/requests"
    Then the response code should be 401
    And the response error message should contain "Invalid access token"

  Scenario: Group id is incorrect
    Given I am the user with id "21"
    When I send a GET request to "/groups/abc/requests"
    Then the response code should be 400
    And the response error message should contain "Wrong value for group_id (should be int64)"

  Scenario: rejections_within_weeks is incorrect
    Given I am the user with id "21"
    When I send a GET request to "/groups/13/requests?rejections_within_weeks=abc"
    Then the response code should be 400
    And the response error message should contain "Wrong value for rejections_within_weeks (should be int64)"

  Scenario: sort is incorrect
    Given I am the user with id "21"
    When I send a GET request to "/groups/13/requests?sort=myname"
    Then the response code should be 400
    And the response error message should contain "Unallowed field in sorting parameters: "myname""

