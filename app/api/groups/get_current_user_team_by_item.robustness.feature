Feature: Get current user's team for item (teamGetByItemID) - robustness
  Background:
    Given the database has the following table "groups":
      | id | type  |
      | 20 | Team  |
      | 21 | Class |
    And the database has the following users:
      | group_id | login  |
      | 11       | owner  |
      | 12       | user   |
      | 13       | jane   |
      | 14       | john   |
      | 15       | jack   |
      | 16       | james  |
      | 17       | jeremy |
      | 19       | jacob  |
    And the database has the following table "groups_groups":
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
