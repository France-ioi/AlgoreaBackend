Feature: Get current user's team for item (teamGetByItemID)
  Background:
    Given the database has the following table 'users':
      | ID | sLogin | idGroupSelf |
      | 2  | user   | 12          |
      | 3  | jane   | 13          |
      | 4  | john   | 14          |
    And the database has the following table 'groups':
      | ID | sType    | idTeamItem |
      | 12 | UserSelf | null       |
      | 13 | UserSelf | null       |
      | 14 | UserSelf | null       |
      | 20 | Team     | 100        |
      | 21 | Team     | 100        |
    And the database has the following table 'groups_groups':
      | idGroupParent | idGroupChild | sType              |
      | 20            | 12           | invitationAccepted |
      | 20            | 13           | requestAccepted    |
      | 20            | 14           | joinedByCode       |
      | 21            | 12           | invitationAccepted |
      | 21            | 13           | requestAccepted    |
      | 21            | 14           | joinedByCode       |

  Scenario: The user joined the team by invitation
    Given I am the user with ID "2"
    When I send a GET request to "/current-user/teams/by-item/100"
    Then the response code should be 200
    And the response body should be, in JSON:
    """
    {
      "group_id": "20"
    }
    """

  Scenario: The user joined the team by request
    Given I am the user with ID "3"
    When I send a GET request to "/current-user/teams/by-item/100"
    Then the response code should be 200
    And the response body should be, in JSON:
    """
    {
      "group_id": "20"
    }
    """

  Scenario: The user joined the team by code
    Given I am the user with ID "3"
    When I send a GET request to "/current-user/teams/by-item/100"
    Then the response code should be 200
    And the response body should be, in JSON:
    """
    {
      "group_id": "20"
    }
    """
