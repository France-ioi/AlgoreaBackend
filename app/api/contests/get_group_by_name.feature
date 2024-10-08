Feature: Get group by name (contestGetGroupByName)
  Background:
    Given the database has the following table "groups":
      | id | name    | type  |
      | 6  | Group B | Team  |
      | 7  | Group B | Team  |
      | 10 | Parent  | Club  |
      | 11 | Group A | Class |
      | 13 | Group B | Team  |
      | 14 | Group B | Other |
      | 15 | Team    | Team  |
      | 16 | Team A  | Team  |
      | 21 | owner   | User  |
      | 31 | john    | User  |
      | 41 | jane    | User  |
      | 50 | Group D | Class |
    And the database has the following table "users":
      | login | group_id |
      | owner | 21       |
      | john  | 31       |
      | jane  | 41       |
    And the database has the following table "group_managers":
      | group_id | manager_id | can_grant_group_access | can_watch_members |
      | 6        | 21         | false                  | true              |
      | 7        | 21         | true                   | false             |
      | 11       | 21         | true                   | true              |
      | 14       | 21         | true                   | true              |
      | 16       | 21         | true                   | true              |
      | 31       | 21         | true                   | true              |
      | 41       | 21         | true                   | true              |
    And the database has the following table "groups_groups":
      | parent_group_id | child_group_id |
      | 10              | 11             |
      | 10              | 13             |
      | 11              | 13             |
      | 11              | 15             |
      | 14              | 14             |
      | 15              | 31             |
      | 15              | 41             |
      | 50              | 13             |
      | 50              | 14             |
      | 50              | 15             |
      | 50              | 31             |
    And the groups ancestors are computed
    And the database has the following table "items":
      | id | duration | entry_participant_type | default_language_tag |
      | 50 | 00:00:00 | Team                   | fr                   |
      | 60 | 00:00:01 | Team                   | fr                   |
      | 10 | 00:00:02 | User                   | fr                   |
      | 70 | 00:00:03 | Team                   | fr                   |
    And the database has the following table "items_ancestors":
      | ancestor_item_id | child_item_id |
      | 60               | 70            |
    And the database has the following table "permissions_granted":
      | group_id | item_id | source_group_id | can_enter_from      | can_enter_until     |
      | 16       | 70      | 16              | 9999-12-31 23:59:58 | 9999-12-31 23:59:59 |
    And the database has the following table "permissions_generated":
      | group_id | item_id | can_view_generated       | can_grant_view_generated | can_watch_generated |
      | 10       | 50      | content                  | enter                    | result              |
      | 11       | 50      | none                     | none                     | none                |
      | 11       | 60      | info                     | none                     | none                |
      | 11       | 70      | content_with_descendants | none                     | none                |
      | 6        | 50      | info                     | none                     | none                |
      | 7        | 50      | info                     | none                     | none                |
      | 13       | 50      | content                  | none                     | none                |
      | 13       | 60      | info                     | none                     | none                |
      | 15       | 60      | info                     | none                     | none                |
      | 21       | 10      | content_with_descendants | enter                    | result              |
      | 21       | 50      | content                  | enter                    | result              |
      | 21       | 60      | content_with_descendants | content                  | answer              |
      | 21       | 70      | content_with_descendants | enter                    | result              |
      | 31       | 50      | content_with_descendants | none                     | none                |
      | 31       | 70      | content_with_descendants | none                     | none                |
      | 41       | 10      | content                  | none                     | none                |
      | 41       | 70      | content                  | none                     | none                |
    And the database has the following table "groups_contest_items":
      | group_id | item_id | additional_time |
      | 10       | 50      | 01:00:00        |
      | 11       | 50      | 00:01:00        |
      | 13       | 50      | 00:00:01        |
      | 13       | 60      | 00:00:30        |
      | 15       | 60      | 00:00:45        |
      | 21       | 50      | 00:01:00        |
      | 21       | 60      | 00:01:00        |
      | 21       | 70      | 00:01:00        |
      | 31       | 50      | 00:01:00        |
      | 31       | 70      | 00:01:00        |
      | 41       | 10      | 00:02:00        |
      | 41       | 70      | 00:01:00        |

  Scenario: Additional time from parent groups
    Given I am the user with id "21"
    When I send a GET request to "/contests/50/groups/by-name?name=Group%20B"
    Then the response code should be 200
    And the response body should be, in JSON:
    """
    {
      "group_id": "13",
      "name": "Group B",
      "type": "Team",
      "additional_time": 1,
      "total_additional_time": 3661
    }
    """

  Scenario: Additional time for the group itself
    Given I am the user with id "21"
    When I send a GET request to "/contests/60/groups/by-name?name=Group%20B"
    Then the response code should be 200
    And the response body should be, in JSON:
    """
    {
      "group_id": "13",
      "name": "Group B",
      "type": "Team",
      "additional_time": 30,
      "total_additional_time": 30
    }
    """

  Scenario: Additional time is null
    Given I am the user with id "21"
    When I send a GET request to "/contests/70/groups/by-name?name=Group%20B"
    Then the response code should be 200
    And the response body should be, in JSON:
    """
    {
      "group_id": "13",
      "name": "Group B",
      "type": "Team",
      "additional_time": 0,
      "total_additional_time": 0
    }
    """

  Scenario: Group cannot view the item, but can enter it
    Given I am the user with id "21"
    When I send a GET request to "/contests/70/groups/by-name?name=Team%20A"
    Then the response code should be 200
    And the response body should be, in JSON:
    """
    {
      "group_id": "16",
      "name": "Team A",
      "type": "Team",
      "additional_time": 0,
      "total_additional_time": 0
    }
    """

  Scenario: Should ignore case
    Given I am the user with id "21"
    When I send a GET request to "/contests/50/groups/by-name?name=group%20b"
    Then the response code should be 200
    And the response body should be, in JSON:
    """
    {
      "group_id": "13",
      "name": "Group B",
      "type": "Team",
      "additional_time": 1,
      "total_additional_time": 3661
    }
    """

  Scenario: Group is a user (non-team contest)
    Given I am the user with id "21"
    When I send a GET request to "/contests/10/groups/by-name?name=jane"
    Then the response code should be 200
    And the response body should be, in JSON:
    """
    {
      "group_id": "41",
      "name": "jane",
      "type": "User",
      "additional_time": 120,
      "total_additional_time": 120
    }
    """

  Scenario: Group is a user (user-only contest)
    Given I am the user with id "21"
    When I send a GET request to "/contests/10/groups/by-name?name=jane"
    Then the response code should be 200
    And the response body should be, in JSON:
    """
    {
      "group_id": "41",
      "name": "jane",
      "type": "User",
      "additional_time": 120,
      "total_additional_time": 120
    }
    """

  Scenario: Group is a user (team contest)
    Given I am the user with id "21"
    When I send a GET request to "/contests/60/groups/by-name?name=john"
    Then the response code should be 200
    And the response body should be, in JSON:
    """
    {
      "group_id": "15",
      "name": "Team",
      "type": "Team",
      "additional_time": 45,
      "total_additional_time": 45
    }
    """
