Feature: List user descendants of the group (groupUserDescendantView) - robustness
  Background:
    Given the database has the following table 'users':
      | id | login | group_self_id | group_owned_id |
      | 1  | owner | 21            | 22             |
      | 2  | user  | 11            | 12             |
    And the database has the following table 'groups_ancestors':
      | group_ancestor_id | group_child_id | is_self |
      | 11                | 11             | 1       |
      | 12                | 12             | 1       |
      | 13                | 13             | 1       |
      | 21                | 21             | 1       |
      | 22                | 13             | 0       |
      | 22                | 22             | 1       |

  Scenario: User is not an admin of the group
    Given I am the user with id "2"
    When I send a GET request to "/groups/13/user-descendants"
    Then the response code should be 403
    And the response error message should contain "Insufficient access rights"

  Scenario: Group id is incorrect
    Given I am the user with id "1"
    When I send a GET request to "/groups/abc/user-descendants"
    Then the response code should be 400
    And the response error message should contain "Wrong value for group_id (should be int64)"

  Scenario: User not found
    Given I am the user with id "404"
    When I send a GET request to "/groups/13/user-descendants"
    Then the response code should be 401
    And the response error message should contain "Invalid access token"

  Scenario: sort is incorrect
    Given I am the user with id "1"
    When I send a GET request to "/groups/13/user-descendants?sort=myname"
    Then the response code should be 400
    And the response error message should contain "Unallowed field in sorting parameters: "myname""

