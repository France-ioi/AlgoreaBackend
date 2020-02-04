Feature: Get current user's team for item (teamGetByItemID)
  Background:
    Given the database has the following table 'groups':
      | id | type | team_item_id |
      | 12 | User | null         |
      | 13 | User | null         |
      | 14 | User | null         |
      | 20 | Team | 100          |
      | 21 | Team | 100          |
    And the database has the following table 'users':
      | login | group_id |
      | user  | 12       |
      | jane  | 13       |
      | john  | 14       |
    And the database has the following table 'groups_groups':
      | parent_group_id | child_group_id |
      | 20              | 12             |
      | 20              | 13             |
      | 20              | 14             |
      | 21              | 12             |
      | 21              | 13             |
      | 21              | 14             |

  Scenario: The user joined the team by invitation
    Given I am the user with id "12"
    When I send a GET request to "/current-user/teams/by-item/100"
    Then the response code should be 200
    And the response body should be, in JSON:
    """
    {
      "group_id": "20"
    }
    """

  Scenario: The user joined the team by request
    Given I am the user with id "13"
    When I send a GET request to "/current-user/teams/by-item/100"
    Then the response code should be 200
    And the response body should be, in JSON:
    """
    {
      "group_id": "20"
    }
    """

  Scenario: The user joined the team by code
    Given I am the user with id "13"
    When I send a GET request to "/current-user/teams/by-item/100"
    Then the response code should be 200
    And the response body should be, in JSON:
    """
    {
      "group_id": "20"
    }
    """
