Feature: Get breadcrumbs (groupBreadcrumbsView)
  Background:
    Given the database has the following table 'groups':
      | id | name                                | type    | is_public |
      | 1  | Joined Base                         | Base    | false     |
      | 2  | Managed Base                        | Base    | false     |
      | 3  | Base                                | Base    | false     |
      | 4  | Joined Class                        | Class   | false     |
      | 5  | School                              | Club    | false     |
      | 6  | Joined Team                         | Team    | false     |
      | 7  | Joined By Ancestor Team             | Class   | false     |
      | 8  | Ancestor Team                       | Team    | false     |
      | 9  | Managed Class                       | Class   | false     |
      | 10 | Managed By Ancestor Team            | Class   | false     |
      | 11 | Ancestor Team                       | Team    | false     |
      | 12 | Managed Ancestor                    | Base    | false     |
      | 13 | With Managed Ancestor               | Friends | false     |
      | 14 | With Managed Descendant             | Other   | false     |
      | 15 | Managed Descendant                  | Team    | false     |
      | 16 | Joined By Ancestor                  | Class   | false     |
      | 17 | Intermediate Group                  | Class   | false     |
      | 18 | Ancestor                            | Class   | false     |
      | 19 | Managed By Ancestor                 | Class   | false     |
      | 20 | Intermediate Group                  | Base    | false     |
      | 21 | Ancestor                            | Base    | false     |
      | 22 | With Descendant Managed By Ancestor | Other   | false     |
      | 23 | Descendant Managed By Ancestor      | Class   | false     |
      | 24 | Intermediate Group                  | Base    | false     |
      | 25 | Ancestor                            | Base    | false     |
      | 26 | Parent                              | Class   | false     |
      | 27 | Public                              | Base    | true      |
      | 41 | user                                | User    | false     |
      | 49 | User                                | User    | false     |
    And the database has the following table 'users':
      | login | group_id | first_name  | last_name |
      | owner | 41       | Jean-Michel | Blanquer  |
      | jack  | 49       | Jack        | Smith     |
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
      | 26              | 4              | 9999-12-31 23:59:59 |
      | 26              | 5              | 9999-12-31 23:59:59 |
      | 26              | 7              | 9999-12-31 23:59:59 |
      | 26              | 9              | 9999-12-31 23:59:59 |
      | 26              | 10             | 9999-12-31 23:59:59 |
      | 26              | 11             | 9999-12-31 23:59:59 |
      | 26              | 13             | 9999-12-31 23:59:59 |
      | 26              | 14             | 9999-12-31 23:59:59 |
      | 26              | 16             | 9999-12-31 23:59:59 |
      | 26              | 19             | 9999-12-31 23:59:59 |
      | 26              | 22             | 9999-12-31 23:59:59 |
      | 26              | 25             | 2010-01-01 00:00:00 |
      | 26              | 27             | 9999-12-31 23:59:59 |
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

  Scenario: Get breadcrumbs
    Given I am the user with id "41"
    When I send a GET request to "/groups/9/11/7/16/4/19/27/22/13/14/breadcrumbs"
    Then the response code should be 200
    And the response body should be, in JSON:
    """
    [
      {"id": "9", "name": "Managed Class", "type": "Class"},
      {"id": "11", "name": "Ancestor Team", "type": "Team"},
      {"id": "7", "name": "Joined By Ancestor Team", "type": "Class"},
      {"id": "16", "name": "Joined By Ancestor", "type": "Class"},
      {"id": "4", "name": "Joined Class", "type": "Class"},
      {"id": "19", "name": "Managed By Ancestor", "type": "Class"},
      {"id": "27", "name": "Public", "type": "Base"},
      {"id": "22", "name": "With Descendant Managed By Ancestor", "type": "Other"},
      {"id": "13", "name": "With Managed Ancestor", "type": "Friends"},
      {"id": "14", "name": "With Managed Descendant", "type": "Other"}
    ]
    """

  Scenario: Displays members of a managed team
    Given I am the user with id "41"
    When I send a GET request to "/groups/6/41/breadcrumbs"
    Then the response code should be 200
    And the response body should be, in JSON:
    """
    [
      {"id": "6", "name": "Joined Team", "type": "Team"},
      {"id": "41", "name": "user", "type": "User"}
    ]
    """

  Scenario: Get breadcrumbs for a public group
    Given I am the user with id "41"
    When I send a GET request to "/groups/27/breadcrumbs"
    Then the response code should be 200
    And the response body should be, in JSON:
    """
    [
      {"id": "27", "name": "Public", "type": "Base"}
    ]
    """

  Scenario: Get breadcrumbs for an ancestor of a joined group
    Given I am the user with id "41"
    When I send a GET request to "/groups/16/18/breadcrumbs"
    Then the response code should be 200
    And the response body should be, in JSON:
    """
    [
      {"id": "16", "name": "Joined By Ancestor", "type": "Class"},
      {"id": "18", "name": "Ancestor", "type": "Class"}
    ]
    """

  Scenario: Get breadcrumbs for an ancestor of a joined team
    Given I am the user with id "41"
    When I send a GET request to "/groups/5/6/breadcrumbs"
    Then the response code should be 200
    And the response body should be, in JSON:
    """
    [
      {"id": "5", "name": "School", "type": "Club"},
      {"id": "6", "name": "Joined Team", "type": "Team"}
    ]
    """

  Scenario: Get breadcrumbs for an ancestor of a managed group
    Given I am the user with id "41"
    When I send a GET request to "/groups/14/15/breadcrumbs"
    Then the response code should be 200
    And the response body should be, in JSON:
    """
    [
      {"id": "14", "name": "With Managed Descendant", "type": "Other"},
      {"id": "15", "name": "Managed Descendant", "type": "Team"}
    ]
    """

  Scenario: Get breadcrumbs for a descendant of a managed group
    Given I am the user with id "41"
    When I send a GET request to "/groups/13/breadcrumbs"
    Then the response code should be 200
    And the response body should be, in JSON:
    """
    [
      {"id": "13", "name": "With Managed Ancestor", "type": "Friends"}
    ]
    """
