Feature: Get current user's team for item (teamGetByItemID) - robustness
  Background:
    Given the database has the following table 'users':
      | id | login  | group_self_id |
      | 1  | owner  | 11            |
      | 2  | user   | 12            |
      | 3  | jane   | 13            |
      | 4  | john   | 14            |
      | 5  | jack   | 15            |
      | 6  | james  | 16            |
      | 7  | jeremy | 17            |
      | 9  | jacob  | 19            |
    And the database has the following table 'groups':
      | id | type     | team_item_id |
      | 11 | UserSelf | null         |
      | 12 | UserSelf | null         |
      | 13 | UserSelf | null         |
      | 14 | UserSelf | null         |
      | 15 | UserSelf | null         |
      | 16 | UserSelf | null         |
      | 17 | UserSelf | null         |
      | 19 | UserSelf | null         |
      | 20 | Team     | 100          |
      | 21 | Class    | 100          |
    And the database has the following table 'groups_groups':
      | group_parent_id | group_child_id | type              |
      | 20              | 12             | invitationSent    |
      | 20              | 13             | requestSent       |
      | 20              | 14             | invitationRefused |
      | 20              | 15             | requestRefused    |
      | 20              | 16             | removed           |
      | 20              | 17             | left              |
      | 21              | 19             | joinedByCode      |

  Scenario: Invalid item_id
    Given I am the user with id "9"
    When I send a GET request to "/current-user/teams/by-item/abc"
    Then the response code should be 400
    And the response error message should contain "Wrong value for item_id (should be int64)"

  Scenario Outline: Wrong groups_groups.type
    Given I am the user with id "<user_id>"
    When I send a GET request to "/current-user/teams/by-item/100"
    Then the response code should be 404
    And the response error message should contain "No team for this item"
    Examples:
      | user_id |
      | 1       |
      | 2       |
      | 3       |
      | 4       |
      | 5       |
      | 6       |
      | 7       |

  Scenario: Wrong groups.type
    Given I am the user with id "9"
    When I send a GET request to "/current-user/teams/by-item/100"
    Then the response code should be 404
    And the response error message should contain "No team for this item"

  Scenario: No team for item
    Given I am the user with id "9"
    When I send a GET request to "/current-user/teams/by-item/101"
    Then the response code should be 404
    And the response error message should contain "No team for this item"
