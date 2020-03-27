Feature: Get group by name (contestGetGroupByName)
  Background:
    Given the database has the following table 'groups':
      | id | name    | type    |
      | 10 | Parent  | Club    |
      | 11 | Group A | Team    |
      | 13 | Group B | Team    |
      | 14 | Group B | Other   |
      | 15 | Team    | Team    |
      | 21 | owner   | User    |
      | 31 | john    | User    |
      | 41 | jane    | User    |
      | 50 | Group D | Class   |
    And the database has the following table 'users':
      | login | group_id |
      | owner | 21       |
      | john  | 31       |
      | jane  | 41       |
    And the database has the following table 'group_managers':
      | group_id | manager_id |
      | 11       | 21         |
      | 14       | 21         |
      | 31       | 21         |
    And the database has the following table 'groups_groups':
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
    And the database has the following table 'groups_ancestors':
      | ancestor_group_id | child_group_id |
      | 10                | 10             |
      | 10                | 11             |
      | 10                | 13             |
      | 10                | 15             |
      | 11                | 11             |
      | 11                | 13             |
      | 11                | 15             |
      | 13                | 13             |
      | 14                | 14             |
      | 15                | 15             |
      | 15                | 31             |
      | 15                | 41             |
      | 21                | 21             |
      | 50                | 50             |
      | 31                | 31             |
      | 32                | 32             |
      | 41                | 41             |
      | 42                | 42             |
    And the database has the following table 'items':
      | id | duration | entry_participant_type | default_language_tag |
      | 50 | 00:00:00 | null                   | fr                   |
      | 60 | 00:00:01 | Team                   | fr                   |
      | 10 | 00:00:02 | User                   | fr                   |
      | 70 | 00:00:03 | Team                   | fr                   |
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

  Scenario: Content access for group, solutions access for user, additional time from parent groups
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

  Scenario: Info access for group, full access for user
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
      "type": "User",
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
      "type": "Team",
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
      "type": "Team",
      "additional_time": 60,
      "total_additional_time": 3660
    }
    """
