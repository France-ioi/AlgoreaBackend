Feature: Get the contests that the user has administration rights on (contestAdminList) - robustness
  Background:
    Given the database has the following users:
      | login      | group_id | owned_group_id | default_language |
      | possesseur | 21       | 22             | fr               |
    And the database has the following table 'groups_ancestors':
      | ancestor_group_id | child_group_id | is_self |
      | 21                | 21             | 1       |
      | 22                | 22             | 1       |

  Scenario: Wrong sort
    Given I am the user with id "21"
    When I send a GET request to "/contests/administered?sort=name"
    Then the response code should be 400
    And the response error message should contain "Unallowed field in sorting parameters: "name""
