Feature: Get current user's team for item (teamGetByItemID) - robustness
  Background:
    Given the database has the following table 'groups':
      | id | type  | team_item_id |
      | 11 | User  | null         |
      | 12 | User  | null         |
      | 13 | User  | null         |
      | 14 | User  | null         |
      | 15 | User  | null         |
      | 16 | User  | null         |
      | 17 | User  | null         |
      | 19 | User  | null         |
      | 20 | Team  | 100          |
      | 21 | Class | 100          |
    And the database has the following table 'users':
      | login  | group_id |
      | owner  | 11       |
      | user   | 12       |
      | jane   | 13       |
      | john   | 14       |
      | jack   | 15       |
      | james  | 16       |
      | jeremy | 17       |
      | jacob  | 19       |
    And the database has the following table 'groups_groups':
      | parent_group_id | child_group_id |
      | 21              | 19             |

  Scenario: Invalid item_id
    Given I am the user with id "19"
    When I send a GET request to "/current-user/teams/by-item/abc"
    Then the response code should be 400
    And the response error message should contain "Wrong value for item_id (should be int64)"

  Scenario: Not a team member
    Given I am the user with id "11"
    When I send a GET request to "/current-user/teams/by-item/100"
    Then the response code should be 404
    And the response error message should contain "No team for this item"

  Scenario: Wrong groups.type
    Given I am the user with id "19"
    When I send a GET request to "/current-user/teams/by-item/100"
    Then the response code should be 404
    And the response error message should contain "No team for this item"

  Scenario: No team for item
    Given I am the user with id "19"
    When I send a GET request to "/current-user/teams/by-item/101"
    Then the response code should be 404
    And the response error message should contain "No team for this item"
