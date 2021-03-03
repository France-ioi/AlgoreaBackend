Feature: Find a group path - robustness
  Background:
    Given the database has the following table 'groups':
      | id | name                                     | type    | is_public |
      | 1  | Joined Base                              | Base    | false     |
      | 2  | Managed Base                             | Base    | false     |
      | 3  | Base                                     | Base    | false     |
      | 4  | Joined Class                             | Class   | false     |
      | 5  | School                                   | Club    | false     |
      | 6  | Joined Team                              | Team    | false     |
      | 7  | Joined By Ancestor Team                  | Class   | false     |
      | 8  | Ancestor Team                            | Team    | false     |
      | 9  | Managed Class                            | Class   | false     |
      | 10 | Managed By Ancestor Team                 | Class   | false     |
      | 11 | Ancestor Team                            | Team    | false     |
      | 12 | Managed Ancestor                         | Base    | false     |
      | 13 | Root With Managed Ancestor               | Friends | false     |
      | 14 | Root With Managed Descendant             | Other   | false     |
      | 15 | Managed Descendant                       | Team    | false     |
      | 16 | Joined By Ancestor                       | Class   | false     |
      | 17 | Intermediate Group                       | Class   | false     |
      | 18 | Ancestor                                 | Class   | false     |
      | 19 | Managed By Ancestor                      | Class   | false     |
      | 20 | Intermediate Group                       | Base    | false     |
      | 21 | Ancestor                                 | Base    | false     |
      | 22 | Root With Descendant Managed By Ancestor | Other   | false     |
      | 23 | Descendant Managed By Ancestor           | Class   | false     |
      | 24 | Intermediate Group                       | Base    | false     |
      | 25 | Ancestor                                 | Base    | false     |
      | 26 | Parent                                   | Class   | false     |
      | 27 | Public                                   | Base    | true      |
      | 28 | S1                                       | Club    | false     |
      | 29 | S2                                       | Club    | false     |
      | 30 | C1                                       | Class   | false     |
      | 31 | C2                                       | Class   | false     |
      | 32 | C3                                       | Class   | false     |
      | 33 | U1                                       | User    | false     |
      | 34 | U2                                       | User    | false     |
      | 35 | U3                                       | User    | false     |
      | 36 | U3                                       | User    | false     |
      | 38 | Public                                   | Friends | true      |
      | 41 | owner                                    | User    | false     |
      | 49 | jack                                     | User    | false     |
      | 50 | jane                                     | User    | false     |
      | 51 | john                                     | User    | false     |
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
      | 28       | 50         |
      | 30       | 51         |
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
      | 28              | 30             | 9999-12-31 23:59:59 |
      | 28              | 31             | 9999-12-31 23:59:59 |
      | 29              | 31             | 9999-12-31 23:59:59 |
      | 29              | 32             | 9999-12-31 23:59:59 |
      | 30              | 33             | 9999-12-31 23:59:59 |
      | 30              | 34             | 9999-12-31 23:59:59 |
      | 31              | 34             | 9999-12-31 23:59:59 |
      | 31              | 35             | 9999-12-31 23:59:59 |
      | 32              | 35             | 9999-12-31 23:59:59 |
      | 32              | 36             | 9999-12-31 23:59:59 |
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

  Scenario: Invalid group_id given
    Given I am the user with id "41"
    When I send a GET request to "/groups/1_1/path-from-root"
    Then the response code should be 400
    And the response error message should contain "Wrong value for group_id (should be int64)"

  Scenario Outline: A path should exist
    Given I am the user with id "41"
    When I send a GET request to "/groups/<group_id>/path-from-root"
    Then the response code should be 403
    And the response error message should contain "Insufficient access rights"
  Examples:
    | group_id |
    | 1        |
    | 10       |
    | 12       |
    | 50       |

  Scenario: Ancestors of managed users are not visible
    Given I am the user with id "51"
    When I send a GET request to "/groups/32/path-from-root"
    Then the response code should be 403
    And the response error message should contain "Insufficient access rights"
