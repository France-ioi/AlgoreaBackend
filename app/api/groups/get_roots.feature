Feature: Get root groups (groupRootsView)
  Background:
    Given the database has the following table 'groups':
      | id | name                                     | type    |
      | 1  | Joined Base                              | Base    |
      | 2  | Managed Base                             | Base    |
      | 3  | Base                                     | Base    |
      | 4  | Joined Class                             | Class   |
      | 5  | School                                   | Club    |
      | 6  | Joined Team                              | Team    |
      | 7  | Joined By Ancestor Team                  | Class   |
      | 8  | Ancestor Team                            | Team    |
      | 9  | Managed Class                            | Class   |
      | 10 | Managed By Ancestor Team                 | Class   |
      | 11 | Ancestor Team                            | Team    |
      | 12 | Managed Ancestor                         | Base    |
      | 13 | Root With Managed Ancestor               | Friends |
      | 14 | Root With Managed Descendant             | Other   |
      | 15 | Managed Descendant                       | Team    |
      | 16 | Joined By Ancestor                       | Class   |
      | 17 | Intermediate Group                       | Class   |
      | 18 | Ancestor                                 | Class   |
      | 19 | Managed By Ancestor                      | Class   |
      | 20 | Intermediate Group                       | Base    |
      | 21 | Ancestor                                 | Base    |
      | 22 | Root With Descendant Managed By Ancestor | Other   |
      | 23 | Descendant Managed By Ancestor           | Class   |
      | 24 | Intermediate Group                       | Base    |
      | 25 | Ancestor                                 | Base    |
      | 26 | S1                                       | Club    |
      | 27 | S2                                       | Club    |
      | 28 | C1                                       | Class   |
      | 29 | C2                                       | Class   |
      | 30 | C3                                       | Class   |
      | 31 | U1                                       | User    |
      | 32 | U2                                       | User    |
      | 33 | U3                                       | User    |
      | 34 | U3                                       | User    |
      | 41 | owner                                    | User    |
      | 49 | jack                                     | User    |
      | 50 | jane                                     | User    |
      | 51 | john                                     | User    |
    And the database has the following table 'users':
      | login | group_id | first_name  | last_name |
      | owner | 41       | Jean-Michel | Blanquer  |
      | jack  | 49       | Jack        | Smith     |
      | jane  | 50       | Jane        | Doe       |
      | john  | 51       | John        | Doe       |
    And the database has the following table 'group_managers':
      | group_id | manager_id |
      | 2        | 41         |
      | 4        | 41         |
      | 9        | 41         |
      | 10       | 11         |
      | 12       | 41         |
      | 15       | 41         |
      | 19       | 21         |
      | 23       | 25         |
      | 26       | 50         |
      | 28       | 51         |
    And the database has the following table 'groups_groups':
      | parent_group_id | child_group_id | expires_at          |
      | 1               | 41             | 9999-12-31 23:59:59 |
      | 3               | 4              | 9999-12-31 23:59:59 |
      | 3               | 5              | 9999-12-31 23:59:59 |
      | 4               | 41             | 9999-12-31 23:59:59 |
      | 5               | 6              | 9999-12-31 23:59:59 |
      | 6               | 41             | 9999-12-31 23:59:59 |
      | 7               | 8              | 9999-12-31 23:59:59 |
      | 8               | 41             | 9999-12-31 23:59:59 |
      | 11              | 41             | 9999-12-31 23:59:59 |
      | 12              | 13             | 9999-12-31 23:59:59 |
      | 14              | 15             | 9999-12-31 23:59:59 |
      | 16              | 18             | 9999-12-31 23:59:59 |
      | 17              | 41             | 9999-12-31 23:59:59 |
      | 18              | 17             | 9999-12-31 23:59:59 |
      | 20              | 41             | 9999-12-31 23:59:59 |
      | 21              | 20             | 9999-12-31 23:59:59 |
      | 22              | 23             | 9999-12-31 23:59:59 |
      | 24              | 41             | 9999-12-31 23:59:59 |
      | 25              | 24             | 9999-12-31 23:59:59 |
      | 26              | 28             | 9999-12-31 23:59:59 |
      | 26              | 29             | 9999-12-31 23:59:59 |
      | 27              | 29             | 9999-12-31 23:59:59 |
      | 27              | 30             | 9999-12-31 23:59:59 |
      | 28              | 31             | 9999-12-31 23:59:59 |
      | 28              | 32             | 9999-12-31 23:59:59 |
      | 29              | 32             | 9999-12-31 23:59:59 |
      | 29              | 33             | 9999-12-31 23:59:59 |
      | 30              | 33             | 9999-12-31 23:59:59 |
      | 30              | 34             | 9999-12-31 23:59:59 |
      | 5               | 41             | 2010-01-01 00:00:00 |
      | 7               | 41             | 2010-01-01 00:00:00 |
      | 9               | 41             | 2010-01-01 00:00:00 |
      | 10              | 41             | 2010-01-01 00:00:00 |
      | 12              | 41             | 2010-01-01 00:00:00 |
      | 13              | 41             | 2010-01-01 00:00:00 |
      | 14              | 41             | 2010-01-01 00:00:00 |
      | 15              | 41             | 2010-01-01 00:00:00 |
      | 16              | 41             | 2010-01-01 00:00:00 |
      | 18              | 41             | 2010-01-01 00:00:00 |
      | 19              | 41             | 2010-01-01 00:00:00 |
      | 21              | 41             | 2010-01-01 00:00:00 |
      | 22              | 41             | 2010-01-01 00:00:00 |
      | 23              | 41             | 2010-01-01 00:00:00 |
      | 25              | 41             | 2010-01-01 00:00:00 |
    And the groups ancestors are computed
    And the database has the following table 'groups_ancestors':
      | ancestor_group_id | child_group_id | expires_at          |
      | 5                 | 41             | 2010-01-01 00:00:00 |
      | 7                 | 41             | 2010-01-01 00:00:00 |
      | 9                 | 41             | 2010-01-01 00:00:00 |
      | 10                | 41             | 2010-01-01 00:00:00 |
      | 12                | 41             | 2010-01-01 00:00:00 |
      | 13                | 41             | 2010-01-01 00:00:00 |
      | 14                | 41             | 2010-01-01 00:00:00 |
      | 15                | 41             | 2010-01-01 00:00:00 |
      | 19                | 41             | 2010-01-01 00:00:00 |
      | 22                | 41             | 2010-01-01 00:00:00 |
      | 23                | 41             | 2010-01-01 00:00:00 |

  Scenario: Get root groups
    Given I am the user with id "41"
    When I send a GET request to "/groups/roots"
    Then the response code should be 200
    And the response body should be, in JSON:
    """
    [
      {
        "id": "11",
        "name": "Ancestor Team",
        "type": "Team",
        "current_user_membership": "direct",
        "current_user_managership": "none"
      },
      {
        "id": "16",
        "name": "Joined By Ancestor",
        "type": "Class",
        "current_user_membership": "descendant",
        "current_user_managership": "descendant"
      },
      {
        "id": "7",
        "name": "Joined By Ancestor Team",
        "type": "Class",
        "current_user_membership": "descendant",
        "current_user_managership": "none"
      },
      {
        "id": "4",
        "name": "Joined Class",
        "type": "Class",
        "current_user_membership": "direct",
        "current_user_managership": "direct"
      },
      {
        "id": "19",
        "name": "Managed By Ancestor",
        "type": "Class",
        "current_user_membership": "none",
        "current_user_managership": "direct"
      },
      {
        "id": "9",
        "name": "Managed Class",
        "type": "Class",
        "current_user_membership": "none",
        "current_user_managership": "direct"
      },
      {
        "id": "22",
        "name": "Root With Descendant Managed By Ancestor",
        "type": "Other",
        "current_user_membership": "none",
        "current_user_managership": "descendant"
      },
      {
        "id": "13",
        "name": "Root With Managed Ancestor",
        "type": "Friends",
        "current_user_membership": "none",
        "current_user_managership": "ancestor"
      },
      {
        "id": "14",
        "name": "Root With Managed Descendant",
        "type": "Other",
        "current_user_membership": "none",
        "current_user_managership": "descendant"
      },
      {
        "id": "5",
        "name": "School",
        "type": "Club",
        "current_user_membership": "descendant",
        "current_user_managership": "none"
      }
    ]
    """

  Scenario: The user himself is not a root group
    Given I am the user with id "49"
    When I send a GET request to "/groups/roots"
    Then the response code should be 200
    And the response body should be, in JSON:
    """
    [
    ]
    """

  Scenario: Ancestors of managed groups are shown
    Given I am the user with id "50"
    When I send a GET request to "/groups/roots"
    Then the response code should be 200
    And the response body should be, in JSON:
    """
    [
      {
        "current_user_managership": "direct",
        "current_user_membership": "none",
        "id": "26",
        "name": "S1",
        "type": "Club"
      },
      {
        "current_user_managership": "descendant",
        "current_user_membership": "none",
        "id": "27",
        "name": "S2",
        "type": "Club"
      }
    ]
    """

  Scenario: Ancestors of managed users are not shown
    Given I am the user with id "51"
    When I send a GET request to "/groups/roots"
    Then the response code should be 200
    And the response body should be, in JSON:
    """
    [
      {
        "current_user_managership": "descendant",
        "current_user_membership": "none",
        "id": "26",
        "name": "S1",
        "type": "Club"
      }
    ]
    """
