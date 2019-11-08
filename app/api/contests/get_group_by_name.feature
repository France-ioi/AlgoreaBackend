Feature: Get group by name (contestGetGroupByName)
  Background:
    Given the database has the following table 'groups':
      | id | name        | type      | team_item_id |
      | 10 | Parent      | Club      | null         |
      | 11 | Group A     | Friends   | null         |
      | 13 | Group B     | Team      | 60           |
      | 14 | Group B     | Other     | null         |
      | 15 | Team        | Team      | 60           |
      | 21 | owner       | UserSelf  | null         |
      | 22 | owner-admin | UserAdmin | null         |
      | 31 | john        | UserSelf  | null         |
      | 32 | john-admin  | UserAdmin | null         |
      | 41 | jane        | UserSelf  | null         |
      | 42 | jane-admin  | UserAdmin | null         |
    And the database has the following table 'users':
      | login | group_id | owned_group_id |
      | owner | 21       | 22             |
      | john  | 31       | 32             |
      | jane  | 41       | 42             |
    And the database has the following table 'groups_groups':
      | parent_group_id | child_group_id | type               |
      | 10              | 11             | direct             |
      | 10              | 13             | direct             |
      | 11              | 13             | direct             |
      | 11              | 15             | direct             |
      | 14              | 14             | direct             |
      | 15              | 31             | invitationAccepted |
      | 15              | 41             | requestAccepted    |
      | 22              | 13             | direct             |
      | 22              | 14             | direct             |
      | 22              | 15             | direct             |
      | 22              | 31             | direct             |
    And the database has the following table 'groups_ancestors':
      | ancestor_group_id | child_group_id | is_self |
      | 10                | 10             | 1       |
      | 10                | 11             | 0       |
      | 10                | 13             | 0       |
      | 10                | 15             | 0       |
      | 11                | 11             | 1       |
      | 11                | 13             | 0       |
      | 11                | 15             | 0       |
      | 13                | 13             | 1       |
      | 14                | 14             | 1       |
      | 15                | 15             | 1       |
      | 15                | 31             | 0       |
      | 15                | 41             | 0       |
      | 21                | 21             | 1       |
      | 22                | 11             | 0       |
      | 22                | 13             | 0       |
      | 22                | 14             | 0       |
      | 22                | 15             | 0       |
      | 22                | 22             | 1       |
      | 22                | 31             | 0       |
      | 31                | 31             | 1       |
      | 32                | 32             | 1       |
      | 41                | 41             | 1       |
      | 42                | 42             | 1       |
    And the database has the following table 'items':
      | id | duration | has_attempts |
      | 50 | 00:00:00 | 0            |
      | 60 | 00:00:01 | 1            |
      | 10 | 00:00:02 | 0            |
      | 70 | 00:00:03 | 1            |
    And the database has the following table 'items_ancestors':
      | ancestor_item_id | child_item_id |
      | 60               | 70            |
    And the database has the following table 'permissions_generated':
      | group_id | item_id | can_view_generated       |
      | 10       | 50      | content                  |
      | 11       | 50      | none                     |
      | 11       | 60      | info                     |
      | 11       | 70      | content_with_descendants |
      | 13       | 50      | content                  |
      | 13       | 60      | info                     |
      | 15       | 60      | info                     |
      | 21       | 50      | solution                 |
      | 21       | 60      | content_with_descendants |
      | 21       | 70      | content_with_descendants |
      | 31       | 50      | content_with_descendants |
      | 31       | 70      | content_with_descendants |
      | 41       | 70      | content                  |
    And the database has the following table 'groups_contest_items':
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
      | 41       | 70      | 00:01:00        |

  Scenario: Partial access for group, solutions access for user, additional time from parent groups
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

  Scenario: Grayed access for group, full access for user
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

  Scenario: Full access for group, full access for user, additional time is null
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

  Scenario: Group is a user group (non-team contest)
    Given I am the user with id "21"
    When I send a GET request to "/contests/50/groups/by-name?name=john"
    Then the response code should be 200
    And the response body should be, in JSON:
    """
    {
      "group_id": "31",
      "name": "john",
      "type": "UserSelf",
      "additional_time": 60,
      "total_additional_time": 60
    }
    """

  Scenario: Group is a user group (team contest) [through invitationAccepted]
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

  Scenario: Group is a user group (team contest) [through requestAccepted]
    Given I am the user with id "21"
    When I send a GET request to "/contests/60/groups/by-name?name=jane"
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

  Scenario: Group is an ancestor group (team contest)
    Given I am the user with id "21"
    When I send a GET request to "/contests/60/groups/by-name?name=Group%20A"
    Then the response code should be 200
    And the response body should be, in JSON:
    """
    {
      "group_id": "11",
      "name": "Group A",
      "type": "Friends",
      "additional_time": 0,
      "total_additional_time": 0
    }
    """

  Scenario: Group is an ancestor group (non-team contest)
    Given I am the user with id "21"
    When I send a GET request to "/contests/50/groups/by-name?name=Group%20A"
    Then the response code should be 200
    And the response body should be, in JSON:
    """
    {
      "group_id": "11",
      "name": "Group A",
      "type": "Friends",
      "additional_time": 60,
      "total_additional_time": 3660
    }
    """
