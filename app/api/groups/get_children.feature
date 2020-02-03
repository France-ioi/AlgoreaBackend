Feature: Get group children (groupChildrenView)
  Background:
    Given the database has the following table 'groups':
      | id | name          | grade | type     | opened | free_access | code       |
      | 11 | Group A       | -3    | Class    | true   | true        | ybqybxnlyo |
      | 13 | Group B       | -2    | Class    | true   | true        | ybabbxnlyo |
      | 21 | user          | -2    | UserSelf | true   | false       | null       |
      | 23 | Our Class     | -3    | Class    | true   | false       | null       |
      | 24 | Root          | -2    | Base     | true   | false       | 3456789abc |
      | 25 | Our Team      | -1    | Team     | true   | false       | 456789abcd |
      | 26 | Our Club      | 0     | Club     | true   | false       | null       |
      | 27 | Our Friends   | 0     | Friends  | true   | false       | 56789abcde |
      | 28 | Other         | 0     | Other    | true   | false       | null       |
      | 29 | UserSelf      | 0     | UserSelf | true   | false       | null       |
      | 30 | RootSelf      | 0     | Base     | true   | false       | null       |
      | 42 | Their Class   | -3    | Class    | true   | false       | null       |
      | 43 | Other Root    | -2    | Base     | true   | false       | 3567894abc |
      | 44 | Other Team    | -1    | Team     | true   | false       | 678934abcd |
      | 45 | Their Club    | 0     | Club     | true   | false       | null       |
      | 46 | Their Friends | 0     | Friends  | true   | false       | 98765abcde |
      | 47 | Other         | 0     | Other    | true   | false       | null       |
      | 51 | john          | 0     | UserSelf | false  | false       | null       |
      | 53 | jane          | 0     | UserSelf | false  | false       | null       |
      | 90 | Sub-Class     | 0     | Team     | false  | false       | null       |
    And the database has the following table 'users':
      | login | group_id | first_name  | last_name |
      | owner | 21       | Jean-Michel | Blanquer  |
      | john  | 51       | John        | Doe       |
      | jane  | 53       | Jane        | Doe       |
    And the database has the following table 'group_managers':
      | group_id | manager_id |
      | 13       | 21         |
    And the database has the following table 'groups_ancestors':
      | ancestor_group_id | child_group_id |
      | 11                | 11             |
      | 13                | 11             |
      | 13                | 13             |
      | 13                | 21             |
      | 13                | 23             |
      | 13                | 24             |
      | 13                | 25             |
      | 13                | 26             |
      | 13                | 27             |
      | 13                | 28             |
      | 13                | 29             |
      | 13                | 30             |
      | 13                | 51             |
      | 13                | 53             |
      | 13                | 90             |
      | 21                | 21             |
      | 23                | 23             |
      | 23                | 51             |
      | 23                | 53             |
      | 23                | 90             |
      | 24                | 24             |
      | 25                | 25             |
      | 25                | 53             |
      | 26                | 26             |
      | 27                | 27             |
      | 27                | 53             |
      | 28                | 28             |
      | 29                | 29             |
      | 30                | 30             |
      | 42                | 42             |
      | 43                | 43             |
      | 44                | 44             |
      | 45                | 45             |
      | 46                | 46             |
      | 47                | 47             |
      | 51                | 51             |
      | 53                | 53             |
      | 90                | 51             |
      | 90                | 53             |
      | 90                | 90             |
    And the database has the following table 'groups_groups':
      | parent_group_id | child_group_id |
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

  Scenario: User is a manager of the parent group, rows are sorted by name by default, UserSelf is skipped
    Given I am the user with id "21"
    When I send a GET request to "/groups/13/children?types_exclude=UserSelf"
    Then the response code should be 200
    And the response body should be, in JSON:
    """
    [
      {"id": "28", "name": "Other", "type": "Other", "free_access": false, "grade": 0, "opened": true, "code": null, "user_count": 0},
      {"id": "23", "name": "Our Class", "type": "Class", "free_access": false, "grade": -3, "opened": true, "code": null, "user_count": 2},
      {"id": "26", "name": "Our Club", "type": "Club", "free_access": false, "grade": 0, "opened": true, "code": null, "user_count": 0},
      {"id": "27", "name": "Our Friends", "type": "Friends", "free_access": false, "grade": 0, "opened": true, "code": "56789abcde", "user_count": 1},
      {"id": "25", "name": "Our Team", "type": "Team", "free_access": false, "grade": -1, "opened": true, "code": "456789abcd", "user_count": 1},
      {"id": "24", "name": "Root", "type": "Base", "free_access": false, "grade": -2, "opened": true, "code": "3456789abc", "user_count": 0},
      {"id": "30", "name": "RootSelf", "type": "Base", "free_access": false, "grade": 0, "opened": true, "code": null, "user_count": 0}
    ]
    """

  Scenario: User is a manager of the parent group, rows are sorted by name by default, all the types are by default
    Given I am the user with id "21"
    When I send a GET request to "/groups/13/children"
    Then the response code should be 200
    And the response body should be, in JSON:
    """
    [
      {"id": "28", "name": "Other", "type": "Other", "free_access": false, "grade": 0, "opened": true, "code": null, "user_count": 0},
      {"id": "23", "name": "Our Class", "type": "Class", "free_access": false, "grade": -3, "opened": true, "code": null, "user_count": 2},
      {"id": "26", "name": "Our Club", "type": "Club", "free_access": false, "grade": 0, "opened": true, "code": null, "user_count": 0},
      {"id": "27", "name": "Our Friends", "type": "Friends", "free_access": false, "grade": 0, "opened": true, "code": "56789abcde", "user_count": 1},
      {"id": "25", "name": "Our Team", "type": "Team", "free_access": false, "grade": -1, "opened": true, "code": "456789abcd", "user_count": 1},
      {"id": "24", "name": "Root", "type": "Base", "free_access": false, "grade": -2, "opened": true, "code": "3456789abc", "user_count": 0},
      {"id": "30", "name": "RootSelf", "type": "Base", "free_access": false, "grade": 0, "opened": true, "code": null, "user_count": 0},
      {"id": "21", "name": "user", "type": "UserSelf", "free_access": false, "grade": -2, "opened": true, "code": null, "user_count": 0},
      {"id": "29", "name": "UserSelf", "type": "UserSelf", "free_access": false, "grade": 0, "opened": true, "code": null, "user_count": 0}
    ]
    """

  Scenario: User is a manager of the parent group, rows are sorted by name by default, all the types are included explicitly
    Given I am the user with id "21"
    When I send a GET request to "/groups/13/children?types_include=Base,Class,Team,Club,Friends,Other,UserSelf"
    Then the response code should be 200
    And the response body should be, in JSON:
    """
    [
      {"id": "28", "name": "Other", "type": "Other", "free_access": false, "grade": 0, "opened": true, "code": null, "user_count": 0},
      {"id": "23", "name": "Our Class", "type": "Class", "free_access": false, "grade": -3, "opened": true, "code": null, "user_count": 2},
      {"id": "26", "name": "Our Club", "type": "Club", "free_access": false, "grade": 0, "opened": true, "code": null, "user_count": 0},
      {"id": "27", "name": "Our Friends", "type": "Friends", "free_access": false, "grade": 0, "opened": true, "code": "56789abcde", "user_count": 1},
      {"id": "25", "name": "Our Team", "type": "Team", "free_access": false, "grade": -1, "opened": true, "code": "456789abcd", "user_count": 1},
      {"id": "24", "name": "Root", "type": "Base", "free_access": false, "grade": -2, "opened": true, "code": "3456789abc", "user_count": 0},
      {"id": "30", "name": "RootSelf", "type": "Base", "free_access": false, "grade": 0, "opened": true, "code": null, "user_count": 0},
      {"id": "21", "name": "user", "type": "UserSelf", "free_access": false, "grade": -2, "opened": true, "code": null, "user_count": 0},
      {"id": "29", "name": "UserSelf", "type": "UserSelf", "free_access": false, "grade": 0, "opened": true, "code": null, "user_count": 0}
    ]
    """

  Scenario: User is a manager of the parent group, rows are sorted by name by default, some types are excluded
    Given I am the user with id "21"
    When I send a GET request to "/groups/13/children?types_exclude=Class,Team,Club,Friends"
    Then the response code should be 200
    And the response body should be, in JSON:
    """
    [
      {"id": "28", "name": "Other", "type": "Other", "free_access": false, "grade": 0, "opened": true, "code": null, "user_count": 0},
      {"id": "24", "name": "Root", "type": "Base", "free_access": false, "grade": -2, "opened": true, "code": "3456789abc", "user_count": 0},
      {"id": "30", "name": "RootSelf", "type": "Base", "free_access": false, "grade": 0, "opened": true, "code": null, "user_count": 0},
      {"id": "21", "name": "user", "type": "UserSelf", "free_access": false, "grade": -2, "opened": true, "code": null, "user_count": 0},
      {"id": "29", "name": "UserSelf", "type": "UserSelf", "free_access": false, "grade": 0, "opened": true, "code": null, "user_count": 0}
    ]
    """

  Scenario: User is a manager of the parent group, rows are sorted by grade, UserSelf is skipped
    Given I am the user with id "21"
    When I send a GET request to "/groups/13/children?sort=grade,id&types_exclude=UserSelf"
    Then the response code should be 200
    And the response body should be, in JSON:
    """
    [
      {"id": "23", "name": "Our Class", "type": "Class", "free_access": false, "grade": -3, "opened": true, "code": null, "user_count": 2},
      {"id": "24", "name": "Root", "type": "Base", "free_access": false, "grade": -2, "opened": true, "code": "3456789abc", "user_count": 0},
      {"id": "25", "name": "Our Team", "type": "Team", "free_access": false, "grade": -1, "opened": true, "code": "456789abcd", "user_count": 1},
      {"id": "26", "name": "Our Club", "type": "Club", "free_access": false, "grade": 0, "opened": true, "code": null, "user_count": 0},
      {"id": "27", "name": "Our Friends", "type": "Friends", "free_access": false, "grade": 0, "opened": true, "code": "56789abcde", "user_count": 1},
      {"id": "28", "name": "Other", "type": "Other", "free_access": false, "grade": 0, "opened": true, "code": null, "user_count": 0},
      {"id": "30", "name": "RootSelf", "type": "Base", "free_access": false, "grade": 0, "opened": true, "code": null, "user_count": 0}
    ]
    """

  Scenario: User is a manager of the parent group, rows are sorted by type, UserSelf is skipped
    Given I am the user with id "21"
    When I send a GET request to "/groups/13/children?sort=type,id&types_exclude=UserSelf"
    Then the response code should be 200
    And the response body should be, in JSON:
    """
    [
      {"id": "23", "name": "Our Class", "type": "Class", "free_access": false, "grade": -3, "opened": true, "code": null, "user_count": 2},
      {"id": "25", "name": "Our Team", "type": "Team", "free_access": false, "grade": -1, "opened": true, "code": "456789abcd", "user_count": 1},
      {"id": "26", "name": "Our Club", "type": "Club", "free_access": false, "grade": 0, "opened": true, "code": null, "user_count": 0},
      {"id": "27", "name": "Our Friends", "type": "Friends", "free_access": false, "grade": 0, "opened": true, "code": "56789abcde", "user_count": 1},
      {"id": "28", "name": "Other", "type": "Other", "free_access": false, "grade": 0, "opened": true, "code": null, "user_count": 0},
      {"id": "24", "name": "Root", "type": "Base", "free_access": false, "grade": -2, "opened": true, "code": "3456789abc", "user_count": 0},
      {"id": "30", "name": "RootSelf", "type": "Base", "free_access": false, "grade": 0, "opened": true, "code": null, "user_count": 0}
    ]
    """

  Scenario: User is a manager of the parent group, rows are sorted by name by default, limit applied
    Given I am the user with id "21"
    When I send a GET request to "/groups/13/children?limit=1"
    Then the response code should be 200
    And the response body should be, in JSON:
    """
    [
      {"id": "28", "name": "Other", "type": "Other", "free_access": false, "grade": 0, "opened": true, "code": null, "user_count": 0}
    ]
    """

  Scenario: User is a manager of the parent group, paging applied, UserSelf is skipped
    Given I am the user with id "21"
    When I send a GET request to "/groups/13/children?from.name=Root&from.id=24&types_exclude=UserSelf"
    Then the response code should be 200
    And the response body should be, in JSON:
    """
    [
      {"id": "30", "name": "RootSelf", "type": "Base", "free_access": false, "grade": 0, "opened": true, "code": null, "user_count": 0}
    ]
    """
