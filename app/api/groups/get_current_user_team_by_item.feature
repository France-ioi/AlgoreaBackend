Feature: Get current user's team for item (teamGetByItemID)
  Background:
    Given the database has the following table "groups":
      | id | type |
      | 20 | Team |
      | 21 | Team |
    And the database has the following users:
      | group_id | login |
      | 12       | user  |
      | 13       | jane  |
      | 14       | john  |
    And the database has the following table "groups_groups":
      | parent_group_id | child_group_id |
      | 20              | 12             |
      | 20              | 13             |
      | 20              | 14             |
      | 21              | 12             |
      | 21              | 13             |
      | 21              | 14             |
    And the database has the following table "attempts":
      | participant_id | id | root_item_id |
      | 20             | 1  | 100          |
      | 21             | 1  | 100          |

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
