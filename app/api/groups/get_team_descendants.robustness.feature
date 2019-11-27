Feature: List team descendants of the group (groupTeamDescendantView) - robustness
  Background:
    Given the database has the following users:
      | login | group_id | owned_group_id |
      | owner | 21       | 22             |
      | user  | 11       | 12             |
    And the database has the following table 'groups':
      | id |
      | 13 |
    And the database has the following table 'group_managers':
      | group_id | manager_id |
      | 13       | 21         |
    And the database has the following table 'groups_ancestors':
      | ancestor_group_id | child_group_id | is_self |
      | 11                | 11             | 1       |
      | 12                | 12             | 1       |
      | 13                | 13             | 1       |
      | 21                | 21             | 1       |
      | 22                | 22             | 1       |

  Scenario: User is not a manager of the group
    Given I am the user with id "11"
    When I send a GET request to "/groups/13/team-descendants"
    Then the response code should be 403
    And the response error message should contain "Insufficient access rights"

  Scenario: Group id is incorrect
    Given I am the user with id "21"
    When I send a GET request to "/groups/abc/team-descendants"
    Then the response code should be 400
    And the response error message should contain "Wrong value for group_id (should be int64)"

  Scenario: User not found
    Given I am the user with id "404"
    When I send a GET request to "/groups/13/team-descendants"
    Then the response code should be 401
    And the response error message should contain "Invalid access token"

  Scenario: sort is incorrect
    Given I am the user with id "21"
    When I send a GET request to "/groups/13/team-descendants?sort=myname"
    Then the response code should be 400
    And the response error message should contain "Unallowed field in sorting parameters: "myname""

