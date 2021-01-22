Feature: Get group children (groupChildrenView)
  Background:
    Given the database has the following table 'groups':
      | id | name          | grade | type    | is_open | is_public | code       |
      | 11 | Group A       | -3    | Class   | true    | true      | ybqybxnlyo |
      | 13 | Group B       | -2    | Class   | true    | true      | ybabbxnlyo |
      | 21 | user          | -2    | User    | true    | false     | null       |
      | 23 | Our Class     | -3    | Class   | true    | false     | null       |
      | 24 | Root          | -2    | Base    | true    | false     | 3456789abc |
      | 25 | Our Team      | -1    | Team    | true    | false     | 456789abcd |
      | 26 | Our Club      | 0     | Club    | true    | false     | null       |
      | 27 | Our Friends   | 0     | Friends | true    | false     | 56789abcde |
      | 28 | Other         | 0     | Other   | true    | false     | null       |
      | 29 | User          | 0     | User    | true    | false     | null       |
      | 30 | AllUsers      | 0     | Base    | true    | false     | null       |
      | 42 | Their Class   | -3    | Class   | true    | false     | null       |
      | 43 | Other Root    | -2    | Base    | true    | false     | 3567894abc |
      | 44 | Other Team    | -1    | Team    | true    | false     | 678934abcd |
      | 45 | Their Club    | 0     | Club    | true    | false     | null       |
      | 46 | Their Friends | 0     | Friends | true    | false     | 98765abcde |
      | 47 | Other         | 0     | Other   | true    | false     | null       |
      | 51 | john          | 0     | User    | false   | false     | null       |
      | 53 | jane          | 0     | User    | false   | false     | null       |
      | 90 | Sub-Class     | 0     | Team    | false   | false     | null       |
    And the database has the following table 'users':
      | login | group_id | first_name  | last_name |
      | owner | 21       | Jean-Michel | Blanquer  |
      | jack  | 29       | Jack        | Smith     |
      | john  | 51       | John        | Doe       |
      | jane  | 53       | Jane        | Doe       |
    And the database has the following table 'group_managers':
      | group_id | manager_id | can_manage            | can_grant_group_access | can_watch_members |
      | 13       | 11         | memberships           | true                   | false             |
      | 13       | 21         | none                  | false                  | false             |
      | 23       | 21         | none                  | true                   | true              |
      | 28       | 21         | memberships_and_group | false                  | true              |
      | 30       | 21         | memberships           | true                   | false             |
    And the database has the following table 'groups_groups':
      | parent_group_id | child_group_id |
      | 11              | 29             |
      | 13              | 21             |
      | 13              | 23             |
      | 13              | 24             |
      | 13              | 25             |
      | 13              | 26             |
      | 13              | 27             |
      | 13              | 28             |
      | 13              | 29             |
      | 13              | 30             |
      | 23              | 51             |
      | 23              | 90             |
      | 25              | 53             |
      | 27              | 53             |
      | 90              | 51             |
    And the groups ancestors are computed

  Scenario: User is a manager of the parent group, rows are sorted by name by default, User is skipped
    Given I am the user with id "21"
    When I send a GET request to "/groups/13/children?types_exclude=User"
    Then the response code should be 200
    And the response body should be, in JSON:
    """
    [
      {"id": "30", "name": "AllUsers", "type": "Base", "is_public": false, "grade": 0, "is_open": true, "code": null, "user_count": 0, "can_manage": "memberships", "can_grant_group_access": true, "can_watch_members": false},
      {"id": "28", "name": "Other", "type": "Other", "is_public": false, "grade": 0, "is_open": true, "code": null, "user_count": 0, "can_manage": "memberships_and_group", "can_grant_group_access": false, "can_watch_members": true},
      {"id": "23", "name": "Our Class", "type": "Class", "is_public": false, "grade": -3, "is_open": true, "code": null, "user_count": 1, "can_manage": "none", "can_grant_group_access": true, "can_watch_members": true},
      {"id": "26", "name": "Our Club", "type": "Club", "is_public": false, "grade": 0, "is_open": true, "code": null, "user_count": 0, "can_manage": "none", "can_grant_group_access": false, "can_watch_members": false},
      {"id": "27", "name": "Our Friends", "type": "Friends", "is_public": false, "grade": 0, "is_open": true, "code": "56789abcde", "user_count": 1, "can_manage": "none", "can_grant_group_access": false, "can_watch_members": false},
      {"id": "25", "name": "Our Team", "type": "Team", "is_public": false, "grade": -1, "is_open": true, "code": "456789abcd", "user_count": 1, "can_manage": "none", "can_grant_group_access": false, "can_watch_members": false},
      {"id": "24", "name": "Root", "type": "Base", "is_public": false, "grade": -2, "is_open": true, "code": "3456789abc", "user_count": 0, "can_manage": "none", "can_grant_group_access": false, "can_watch_members": false}
    ]
    """

  Scenario: User is a manager of the parent group, rows are sorted by name by default, all the types are by default
    Given I am the user with id "21"
    When I send a GET request to "/groups/13/children"
    Then the response code should be 200
    And the response body should be, in JSON:
    """
    [
      {"id": "30", "name": "AllUsers", "type": "Base", "is_public": false, "grade": 0, "is_open": true, "code": null, "user_count": 0, "can_manage": "memberships", "can_grant_group_access": true, "can_watch_members": false},
      {"id": "28", "name": "Other", "type": "Other", "is_public": false, "grade": 0, "is_open": true, "code": null, "user_count": 0, "can_manage": "memberships_and_group", "can_grant_group_access": false, "can_watch_members": true},
      {"id": "23", "name": "Our Class", "type": "Class", "is_public": false, "grade": -3, "is_open": true, "code": null, "user_count": 1, "can_manage": "none", "can_grant_group_access": true, "can_watch_members": true},
      {"id": "26", "name": "Our Club", "type": "Club", "is_public": false, "grade": 0, "is_open": true, "code": null, "user_count": 0, "can_manage": "none", "can_grant_group_access": false, "can_watch_members": false},
      {"id": "27", "name": "Our Friends", "type": "Friends", "is_public": false, "grade": 0, "is_open": true, "code": "56789abcde", "user_count": 1, "can_manage": "none", "can_grant_group_access": false, "can_watch_members": false},
      {"id": "25", "name": "Our Team", "type": "Team", "is_public": false, "grade": -1, "is_open": true, "code": "456789abcd", "user_count": 1, "can_manage": "none", "can_grant_group_access": false, "can_watch_members": false},
      {"id": "24", "name": "Root", "type": "Base", "is_public": false, "grade": -2, "is_open": true, "code": "3456789abc", "user_count": 0, "can_manage": "none", "can_grant_group_access": false, "can_watch_members": false},
      {"id": "21", "name": "user", "type": "User", "is_public": false, "grade": -2, "is_open": true, "code": null, "user_count": 0, "can_manage": "none", "can_grant_group_access": false, "can_watch_members": false},
      {"id": "29", "name": "User", "type": "User", "is_public": false, "grade": 0, "is_open": true, "code": null, "user_count": 0, "can_manage": "none", "can_grant_group_access": false, "can_watch_members": false}
    ]
    """

  Scenario: User is a manager of the parent group, rows are sorted by name by default, all the types are included explicitly
    Given I am the user with id "21"
    When I send a GET request to "/groups/13/children?types_include=Base,Class,Team,Club,Friends,Other,User"
    Then the response code should be 200
    And the response body should be, in JSON:
    """
    [
      {"id": "30", "name": "AllUsers", "type": "Base", "is_public": false, "grade": 0, "is_open": true, "code": null, "user_count": 0, "can_manage": "memberships", "can_grant_group_access": true, "can_watch_members": false},
      {"id": "28", "name": "Other", "type": "Other", "is_public": false, "grade": 0, "is_open": true, "code": null, "user_count": 0, "can_manage": "memberships_and_group", "can_grant_group_access": false, "can_watch_members": true},
      {"id": "23", "name": "Our Class", "type": "Class", "is_public": false, "grade": -3, "is_open": true, "code": null, "user_count": 1, "can_manage": "none", "can_grant_group_access": true, "can_watch_members": true},
      {"id": "26", "name": "Our Club", "type": "Club", "is_public": false, "grade": 0, "is_open": true, "code": null, "user_count": 0, "can_manage": "none", "can_grant_group_access": false, "can_watch_members": false},
      {"id": "27", "name": "Our Friends", "type": "Friends", "is_public": false, "grade": 0, "is_open": true, "code": "56789abcde", "user_count": 1, "can_manage": "none", "can_grant_group_access": false, "can_watch_members": false},
      {"id": "25", "name": "Our Team", "type": "Team", "is_public": false, "grade": -1, "is_open": true, "code": "456789abcd", "user_count": 1, "can_manage": "none", "can_grant_group_access": false, "can_watch_members": false},
      {"id": "24", "name": "Root", "type": "Base", "is_public": false, "grade": -2, "is_open": true, "code": "3456789abc", "user_count": 0, "can_manage": "none", "can_grant_group_access": false, "can_watch_members": false},
      {"id": "21", "name": "user", "type": "User", "is_public": false, "grade": -2, "is_open": true, "code": null, "user_count": 0, "can_manage": "none", "can_grant_group_access": false, "can_watch_members": false},
      {"id": "29", "name": "User", "type": "User", "is_public": false, "grade": 0, "is_open": true, "code": null, "user_count": 0, "can_manage": "none", "can_grant_group_access": false, "can_watch_members": false}
    ]
    """

  Scenario: User is a manager of the parent group, rows are sorted by name by default, some types are excluded
    Given I am the user with id "21"
    When I send a GET request to "/groups/13/children?types_exclude=Class,Team,Club,Friends"
    Then the response code should be 200
    And the response body should be, in JSON:
    """
    [
      {"id": "30", "name": "AllUsers", "type": "Base", "is_public": false, "grade": 0, "is_open": true, "code": null, "user_count": 0, "can_manage": "memberships", "can_grant_group_access": true, "can_watch_members": false},
      {"id": "28", "name": "Other", "type": "Other", "is_public": false, "grade": 0, "is_open": true, "code": null, "user_count": 0, "can_manage": "memberships_and_group", "can_grant_group_access": false, "can_watch_members": true},
      {"id": "24", "name": "Root", "type": "Base", "is_public": false, "grade": -2, "is_open": true, "code": "3456789abc", "user_count": 0, "can_manage": "none", "can_grant_group_access": false, "can_watch_members": false},
      {"id": "21", "name": "user", "type": "User", "is_public": false, "grade": -2, "is_open": true, "code": null, "user_count": 0, "can_manage": "none", "can_grant_group_access": false, "can_watch_members": false},
      {"id": "29", "name": "User", "type": "User", "is_public": false, "grade": 0, "is_open": true, "code": null, "user_count": 0, "can_manage": "none", "can_grant_group_access": false, "can_watch_members": false}
    ]
    """

  Scenario: User is a manager of the parent group, rows are sorted by grade, User is skipped
    Given I am the user with id "21"
    When I send a GET request to "/groups/13/children?sort=grade,id&types_exclude=User"
    Then the response code should be 200
    And the response body should be, in JSON:
    """
    [
      {"id": "23", "name": "Our Class", "type": "Class", "is_public": false, "grade": -3, "is_open": true, "code": null, "user_count": 1, "can_manage": "none", "can_grant_group_access": true, "can_watch_members": true},
      {"id": "24", "name": "Root", "type": "Base", "is_public": false, "grade": -2, "is_open": true, "code": "3456789abc", "user_count": 0, "can_manage": "none", "can_grant_group_access": false, "can_watch_members": false},
      {"id": "25", "name": "Our Team", "type": "Team", "is_public": false, "grade": -1, "is_open": true, "code": "456789abcd", "user_count": 1, "can_manage": "none", "can_grant_group_access": false, "can_watch_members": false},
      {"id": "26", "name": "Our Club", "type": "Club", "is_public": false, "grade": 0, "is_open": true, "code": null, "user_count": 0, "can_manage": "none", "can_grant_group_access": false, "can_watch_members": false},
      {"id": "27", "name": "Our Friends", "type": "Friends", "is_public": false, "grade": 0, "is_open": true, "code": "56789abcde", "user_count": 1, "can_manage": "none", "can_grant_group_access": false, "can_watch_members": false},
      {"id": "28", "name": "Other", "type": "Other", "is_public": false, "grade": 0, "is_open": true, "code": null, "user_count": 0, "can_manage": "memberships_and_group", "can_grant_group_access": false, "can_watch_members": true},
      {"id": "30", "name": "AllUsers", "type": "Base", "is_public": false, "grade": 0, "is_open": true, "code": null, "user_count": 0, "can_manage": "memberships", "can_grant_group_access": true, "can_watch_members": false}
    ]
    """

  Scenario: User is a manager of the parent group, rows are sorted by type, User is skipped
    Given I am the user with id "21"
    When I send a GET request to "/groups/13/children?sort=type,id&types_exclude=User"
    Then the response code should be 200
    And the response body should be, in JSON:
    """
    [
      {"id": "23", "name": "Our Class", "type": "Class", "is_public": false, "grade": -3, "is_open": true, "code": null, "user_count": 1, "can_manage": "none", "can_grant_group_access": true, "can_watch_members": true},
      {"id": "25", "name": "Our Team", "type": "Team", "is_public": false, "grade": -1, "is_open": true, "code": "456789abcd", "user_count": 1, "can_manage": "none", "can_grant_group_access": false, "can_watch_members": false},
      {"id": "26", "name": "Our Club", "type": "Club", "is_public": false, "grade": 0, "is_open": true, "code": null, "user_count": 0, "can_manage": "none", "can_grant_group_access": false, "can_watch_members": false},
      {"id": "27", "name": "Our Friends", "type": "Friends", "is_public": false, "grade": 0, "is_open": true, "code": "56789abcde", "user_count": 1, "can_manage": "none", "can_grant_group_access": false, "can_watch_members": false},
      {"id": "28", "name": "Other", "type": "Other", "is_public": false, "grade": 0, "is_open": true, "code": null, "user_count": 0, "can_manage": "memberships_and_group", "can_grant_group_access": false, "can_watch_members": true},
      {"id": "24", "name": "Root", "type": "Base", "is_public": false, "grade": -2, "is_open": true, "code": "3456789abc", "user_count": 0, "can_manage": "none", "can_grant_group_access": false, "can_watch_members": false},
      {"id": "30", "name": "AllUsers", "type": "Base", "is_public": false, "grade": 0, "is_open": true, "code": null, "user_count": 0, "can_manage": "memberships", "can_grant_group_access": true, "can_watch_members": false}
    ]
    """

  Scenario: User is a manager of the parent group, rows are sorted by name by default, limit applied
    Given I am the user with id "21"
    When I send a GET request to "/groups/13/children?limit=1"
    Then the response code should be 200
    And the response body should be, in JSON:
    """
    [
      {"id": "30", "name": "AllUsers", "type": "Base", "is_public": false, "grade": 0, "is_open": true, "code": null, "user_count": 0, "can_manage": "memberships", "can_grant_group_access": true, "can_watch_members": false}
    ]
    """

  Scenario: User is a manager of the parent group, paging applied, User is skipped
    Given I am the user with id "21"
    When I send a GET request to "/groups/13/children?from.name=Our%20Team&from.id=25&types_exclude=User"
    Then the response code should be 200
    And the response body should be, in JSON:
    """
    [
      {"id": "24", "name": "Root", "type": "Base", "is_public": false, "grade": -2, "is_open": true, "code": "3456789abc", "user_count": 0, "can_manage": "none", "can_grant_group_access": false, "can_watch_members": false}
    ]
    """

  Scenario: User's ancestor is a manager of the parent group
    Given I am the user with id "29"
    When I send a GET request to "/groups/13/children"
    Then the response code should be 200
    And the response body should be, in JSON:
    """
    [
      {"id": "30", "name": "AllUsers", "type": "Base", "is_public": false, "grade": 0, "is_open": true, "code": null, "user_count": 0, "can_manage": "memberships", "can_grant_group_access": true, "can_watch_members": false},
      {"id": "28", "name": "Other", "type": "Other", "is_public": false, "grade": 0, "is_open": true, "code": null, "user_count": 0, "can_manage": "memberships", "can_grant_group_access": true, "can_watch_members": false},
      {"id": "23", "name": "Our Class", "type": "Class", "is_public": false, "grade": -3, "is_open": true, "code": null, "user_count": 1, "can_manage": "memberships", "can_grant_group_access": true, "can_watch_members": false},
      {"id": "26", "name": "Our Club", "type": "Club", "is_public": false, "grade": 0, "is_open": true, "code": null, "user_count": 0, "can_manage": "memberships", "can_grant_group_access": true, "can_watch_members": false},
      {"id": "27", "name": "Our Friends", "type": "Friends", "is_public": false, "grade": 0, "is_open": true, "code": "56789abcde", "user_count": 1, "can_manage": "memberships", "can_grant_group_access": true, "can_watch_members": false},
      {"id": "25", "name": "Our Team", "type": "Team", "is_public": false, "grade": -1, "is_open": true, "code": "456789abcd", "user_count": 1, "can_manage": "memberships", "can_grant_group_access": true, "can_watch_members": false},
      {"id": "24", "name": "Root", "type": "Base", "is_public": false, "grade": -2, "is_open": true, "code": "3456789abc", "user_count": 0, "can_manage": "memberships", "can_grant_group_access": true, "can_watch_members": false},
      {"id": "21", "name": "user", "type": "User", "is_public": false, "grade": -2, "is_open": true, "code": null, "user_count": 0, "can_manage": "memberships", "can_grant_group_access": true, "can_watch_members": false},
      {"id": "29", "name": "User", "type": "User", "is_public": false, "grade": 0, "is_open": true, "code": null, "user_count": 0, "can_manage": "memberships", "can_grant_group_access": true, "can_watch_members": false}
    ]
    """
