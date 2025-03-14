Feature: Get additional times for a group of users/teams on an item with duration (itemListMembersAdditionalTime)
  Background:
    Given the database has the following table "groups":
      | id | name        | type    |
      | 10 | Parent      | Club    |
      | 11 | Group A     | Friends |
      | 13 | Group B     | Team    |
      | 14 | Group B     | Other   |
      | 15 | Team        | Team    |
      | 16 | Team for 70 | Team    |
      | 17 | Team wo/acc | Team    |
      | 21 | owner       | User    |
      | 31 | john        | User    |
      | 41 | jane        | User    |
      | 51 | jack        | User    |
      | 61 | jack        | User    |
    And the database has the following users:
      | group_id | login |
      | 21       | owner |
      | 31       | john  |
      | 41       | jane  |
      | 51       | jack  |
      | 61       | paul  |
    And the database has the following table "group_managers":
      | group_id | manager_id | can_grant_group_access | can_watch_members |
      | 11       | 21         | true                   | true              |
      | 13       | 21         | false                  | false             |
      | 14       | 21         | false                  | false             |
      | 15       | 21         | false                  | false             |
      | 16       | 21         | false                  | false             |
      | 17       | 21         | false                  | false             |
      | 31       | 21         | false                  | false             |
      | 41       | 21         | false                  | false             |
    And the database has the following table "groups_groups":
      | parent_group_id | child_group_id |
      | 10              | 11             |
      | 10              | 13             |
      | 11              | 13             |
      | 11              | 14             |
      | 11              | 15             |
      | 11              | 16             |
      | 11              | 17             |
      | 11              | 51             |
      | 11              | 61             |
      | 14              | 31             |
      | 14              | 41             |
      | 15              | 31             |
      | 15              | 41             |
    And the groups ancestors are computed
    And the database has the following table "items":
      | id | duration | entry_participant_type | default_language_tag |
      | 50 | 00:00:00 | User                   | fr                   |
      | 60 | 00:00:01 | Team                   | fr                   |
      | 10 | 00:00:02 | User                   | fr                   |
      | 70 | 00:00:03 | Team                   | fr                   |
    And the database has the following table "items_ancestors":
      | ancestor_item_id | child_item_id |
      | 10               | 60            |
      | 10               | 70            |
      | 60               | 70            |
    And the database has the following table "permissions_granted":
      | group_id | item_id | source_group_id | can_enter_from      | can_enter_until     |
      | 51       | 50      | 51              | 9999-12-31 23:59:58 | 9999-12-31 23:59:59 |
      | 61       | 50      | 61              | 9999-12-31 23:59:59 | 9999-12-31 23:59:59 |
    And the database has the following table "permissions_generated":
      | group_id | item_id | can_view_generated       | can_grant_view_generated | can_watch_generated |
      | 10       | 50      | none                     | none                     | none                |
      | 11       | 50      | none                     | none                     | none                |
      | 11       | 70      | content_with_descendants | none                     | none                |
      | 13       | 50      | content                  | none                     | none                |
      | 13       | 60      | info                     | none                     | none                |
      | 14       | 50      | content_with_descendants | none                     | none                |
      | 15       | 60      | info                     | none                     | none                |
      | 16       | 60      | info                     | none                     | none                |
      | 21       | 50      | content                  | enter                    | result              |
      | 21       | 60      | content_with_descendants | content                  | answer              |
      | 21       | 70      | content_with_descendants | none                     | answer              |
      | 31       | 70      | content_with_descendants | none                     | none                |
      | 41       | 50      | info                     | none                     | none                |
      | 41       | 70      | content                  | none                     | none                |
    And the database has the following table "group_item_additional_times":
      | group_id | item_id | additional_time |
      | 10       | 50      | 01:00:00        |
      | 11       | 50      | 00:01:00        |
      | 13       | 50      | 00:00:01        |
      | 13       | 60      | 00:00:30        |
      | 15       | 60      | 00:00:45        |
      | 16       | 60      | 00:00:45        |
      | 21       | 50      | 00:01:00        |
      | 21       | 60      | 00:01:00        |
      | 21       | 70      | 00:01:00        |
      | 31       | 50      | 00:01:00        |
      | 31       | 70      | 00:01:00        |
      | 41       | 70      | 00:01:00        |
    And the database has the following table "attempts":
      | participant_id | id | root_item_id |
      | 13             | 1  | 60           |
      | 15             | 1  | 60           |
      | 31             | 1  | 50           |
      | 41             | 1  | 50           |
      | 51             | 1  | 50           |
      | 61             | 1  | 50           |

  Scenario: Non-team item
    Given I am the user with id "21"
    When I send a GET request to "/items/50/groups/11/members/additional-times"
    Then the response code should be 200
    And the response body should be, in JSON:
    """
    [
      {
        "group_id": "51",
        "name": "jack",
        "type": "User",
        "additional_time": 0,
        "total_additional_time": 3660
      },
      {
        "group_id": "41",
        "name": "jane",
        "type": "User",
        "additional_time": 0,
        "total_additional_time": 3660
      },
      {
        "group_id": "31",
        "name": "john",
        "type": "User",
        "additional_time": 60,
        "total_additional_time": 3720
      }
    ]
    """

  Scenario: Team-only item
    Given I am the user with id "21"
    When I send a GET request to "/items/60/groups/11/members/additional-times"
    Then the response code should be 200
    And the response body should be, in JSON:
    """
    [
      {
        "group_id": "13",
        "name": "Group B",
        "type": "Team",
        "additional_time": 30,
        "total_additional_time": 30
      },
      {
        "group_id": "15",
        "name": "Team",
        "type": "Team",
        "additional_time": 45,
        "total_additional_time": 45
      }
    ]
    """

  Scenario: Team-only item (only the first row)
    Given I am the user with id "21"
    When I send a GET request to "/items/60/groups/11/members/additional-times?limit=1"
    Then the response code should be 200
    And the response body should be, in JSON:
    """
    [
      {
        "group_id": "13",
        "name": "Group B",
        "type": "Team",
        "additional_time": 30,
        "total_additional_time": 30
      }
    ]
    """

  Scenario: Non-team item (only the first row, inverse order)
    Given I am the user with id "21"
    When I send a GET request to "/items/50/groups/11/members/additional-times?limit=1&sort=-name,id"
    Then the response code should be 200
    And the response body should be, in JSON:
    """
    [
      {
        "group_id": "31",
        "name": "john",
        "type": "User",
        "additional_time": 60,
        "total_additional_time": 3720
      }
    ]
    """

  Scenario: Team-only item (start from the second row)
    Given I am the user with id "21"
    When I send a GET request to "/items/60/groups/11/members/additional-times?from.id=13"
    Then the response code should be 200
    And the response body should be, in JSON:
    """
    [
      {
        "group_id": "15",
        "name": "Team",
        "type": "Team",
        "additional_time": 45,
        "total_additional_time": 45
      }
    ]
    """

